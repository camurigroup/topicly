package rbac

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/modules/audit"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// Audit action names recorded by the RBAC admin service.
const (
	ActionRoleCreate       = "role.create"
	ActionRoleUpdate       = "role.update"
	ActionRoleDelete       = "role.delete"
	ActionRolePermissions  = "role.permissions"
	ActionPermissionCreate = "permission.create"
	ActionPermissionToggle = "permission.toggle"
	ActionPermissionDelete = "permission.delete"
	ActionUserRolesChange  = "user.roles"
)

// corePermissions can never be deactivated or deleted: doing so would lock
// everyone (including the Owner) out of access management.
var corePermissions = map[string]bool{PermRolesManage: true, PermPermissionsManage: true}

func isBuiltinPermission(slug string) bool {
	for _, p := range allPermissions {
		if p.Slug == slug {
			return true
		}
	}
	return false
}

// RoleDetail is a role with its permissions and member count.
type RoleDetail struct {
	Role
	Permissions []string `json:"permissions"`
	MemberCount int      `json:"memberCount"`
}

// PermissionDetail is a permission flagged as built-in/core.
type PermissionDetail struct {
	Permission
	IsBuiltin bool `json:"isBuiltin"`
	IsCore    bool `json:"isCore"`
}

// CreateRoleInput is the payload to create a role.
type CreateRoleInput struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	RankOrder   int      `json:"rankOrder"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleInput is the payload to rename or re-rank a role.
type UpdateRoleInput struct {
	Name      string `json:"name"`
	RankOrder int    `json:"rankOrder"`
}

// CreatePermissionInput is the payload to create a custom permission.
type CreatePermissionInput struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// SetAuditor attaches the audit recorder (optional).
func (s *Service) SetAuditor(a *audit.Service) { s.audit = a }

// ResolveActor loads the hierarchy rank, ownership and permissions of a user.
func (s *Service) ResolveActor(ctx context.Context, userID string) (Actor, error) {
	roles, err := s.repo.RolesOfUser(ctx, userID)
	if err != nil {
		return Actor{}, err
	}
	perms, err := s.repo.PermissionSlugsForUser(ctx, userID)
	if err != nil {
		return Actor{}, err
	}
	a := Actor{ID: userID, Perms: make(map[string]bool, len(perms))}
	for _, p := range perms {
		a.Perms[p] = true
	}
	for _, r := range roles {
		if r.Slug == RoleOwner {
			a.IsOwner = true
		}
		if r.RankOrder > a.Rank {
			a.Rank = r.RankOrder
		}
	}
	return a, nil
}

// ListRoles returns every role with its permissions and member counts.
func (s *Service) ListRoles(ctx context.Context) ([]RoleDetail, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	perms, err := s.repo.RolePermissions(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.RoleMemberCounts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RoleDetail, len(roles))
	for i, r := range roles {
		p := perms[r.ID]
		if p == nil {
			p = []string{}
		}
		out[i] = RoleDetail{Role: r, Permissions: p, MemberCount: counts[r.ID]}
	}
	return out, nil
}

func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if l := len([]rune(name)); l < 2 || l > 80 {
		return "", invalid("o nome deve ter entre 2 e 80 caracteres")
	}
	return name, nil
}

// CreateRole creates a custom role below the actor in the hierarchy, with
// only permissions the actor itself holds.
func (s *Service) CreateRole(ctx context.Context, actorID string, in CreateRoleInput) (*Role, error) {
	actor, err := s.ResolveActor(ctx, actorID)
	if err != nil {
		return nil, err
	}
	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	if err := ValidRoleSlug(slug); err != nil {
		return nil, err
	}
	name, err := cleanName(in.Name)
	if err != nil {
		return nil, err
	}
	if err := actor.CanUseRank(in.RankOrder); err != nil {
		return nil, err
	}
	perms := dedupe(in.Permissions)
	if err := actor.CanGrant(perms); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetRoleBySlug(ctx, slug); err == nil {
		return nil, &policyError{msg: "já existe um cargo com este slug", kind: apperr.ErrConflict}
	}
	now := time.Now().UTC()
	role := &Role{ID: newID(), Slug: slug, Name: name, RankOrder: in.RankOrder, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	if err := s.repo.SetRolePermissions(ctx, role.ID, perms); err != nil {
		_ = s.repo.DeleteRole(ctx, role.ID)
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionRoleCreate, "role", role.ID, map[string]any{
		"slug": slug, "name": name, "rankOrder": in.RankOrder, "permissions": perms,
	})
	return role, nil
}

func (s *Service) manageableRole(ctx context.Context, actorID, roleID string) (Actor, *Role, error) {
	actor, err := s.ResolveActor(ctx, actorID)
	if err != nil {
		return Actor{}, nil, err
	}
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return Actor{}, nil, err
	}
	if err := actor.CanManageRole(*role); err != nil {
		return Actor{}, nil, err
	}
	return actor, role, nil
}

// UpdateRole renames or re-ranks a role the actor outranks.
func (s *Service) UpdateRole(ctx context.Context, actorID, roleID string, in UpdateRoleInput) (*Role, error) {
	actor, role, err := s.manageableRole(ctx, actorID, roleID)
	if err != nil {
		return nil, err
	}
	name, err := cleanName(in.Name)
	if err != nil {
		return nil, err
	}
	if err := actor.CanUseRank(in.RankOrder); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateRole(ctx, roleID, name, in.RankOrder); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionRoleUpdate, "role", roleID, map[string]any{
		"slug": role.Slug, "name": map[string]any{"from": role.Name, "to": name},
		"rankOrder": map[string]any{"from": role.RankOrder, "to": in.RankOrder},
	})
	return s.repo.GetRoleByID(ctx, roleID)
}

// SetRolePermissions replaces a role's permissions. Newly added permissions
// must be held by the actor. The change is effective on the next request.
func (s *Service) SetRolePermissions(ctx context.Context, actorID, roleID string, permissions []string) ([]string, error) {
	actor, role, err := s.manageableRole(ctx, actorID, roleID)
	if err != nil {
		return nil, err
	}
	current, err := s.repo.RolePermissions(ctx)
	if err != nil {
		return nil, err
	}
	next := dedupe(permissions)
	added, removed := Diff(current[roleID], next)
	if err := actor.CanGrant(added); err != nil {
		return nil, err
	}
	if err := s.repo.SetRolePermissions(ctx, roleID, next); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionRolePermissions, "role", roleID, map[string]any{
		"slug": role.Slug, "added": orEmpty(added), "removed": orEmpty(removed),
	})
	return next, nil
}

// DeleteRole removes a non-system role the actor outranks.
func (s *Service) DeleteRole(ctx context.Context, actorID, roleID string) error {
	_, role, err := s.manageableRole(ctx, actorID, roleID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return forbidden("cargos de sistema não podem ser excluídos")
	}
	if err := s.repo.DeleteRole(ctx, roleID); err != nil {
		return err
	}
	s.audit.Record(ctx, actorID, ActionRoleDelete, "role", roleID, map[string]any{"slug": role.Slug, "name": role.Name})
	return nil
}

// ListPermissions returns the catalog flagged with built-in/core markers.
func (s *Service) ListPermissions(ctx context.Context) ([]PermissionDetail, error) {
	perms, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PermissionDetail, len(perms))
	for i, p := range perms {
		out[i] = PermissionDetail{Permission: p, IsBuiltin: isBuiltinPermission(p.Slug), IsCore: corePermissions[p.Slug]}
	}
	return out, nil
}

var permSlugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,78}[a-z0-9]$`)

// CreatePermission adds a custom permission to the catalog.
func (s *Service) CreatePermission(ctx context.Context, actorID string, in CreatePermissionInput) (*Permission, error) {
	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	if !permSlugRe.MatchString(slug) {
		return nil, invalid("slug inválido (letras minúsculas, números, pontos e hífens)")
	}
	name, err := cleanName(in.Name)
	if err != nil {
		return nil, err
	}
	var desc *string
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if len([]rune(d)) > 255 {
			return nil, invalid("a descrição deve ter no máximo 255 caracteres")
		}
		if d != "" {
			desc = &d
		}
	}
	if ids, _ := s.repo.permissionIDs(ctx, []string{slug}); len(ids) > 0 {
		return nil, &policyError{msg: "já existe uma permissão com este slug", kind: apperr.ErrConflict}
	}
	p := &Permission{ID: newID(), Slug: slug, Name: name, Description: desc, IsActive: true, CreatedAt: time.Now().UTC()}
	if err := s.repo.CreatePermission(ctx, p); err != nil {
		return nil, err
	}
	// The Owner always holds every permission.
	if owner, err := s.repo.GetRoleBySlug(ctx, RoleOwner); err == nil {
		_ = s.repo.AssignPermissionToRole(ctx, owner.ID, p.ID)
	}
	s.audit.Record(ctx, actorID, ActionPermissionCreate, "permission", p.ID, map[string]any{"slug": slug, "name": name})
	return p, nil
}

// SetPermissionActive activates or deactivates a permission globally.
func (s *Service) SetPermissionActive(ctx context.Context, actorID, permID string, active bool) (*Permission, error) {
	p, err := s.repo.GetPermissionByID(ctx, permID)
	if err != nil {
		return nil, err
	}
	if !active && corePermissions[p.Slug] {
		return nil, forbidden("esta permissão é essencial e não pode ser desativada")
	}
	if err := s.repo.SetPermissionActive(ctx, permID, active); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionPermissionToggle, "permission", permID, map[string]any{"slug": p.Slug, "active": active})
	p.IsActive = active
	return p, nil
}

// DeletePermission removes a custom (non built-in) permission.
func (s *Service) DeletePermission(ctx context.Context, actorID, permID string) error {
	p, err := s.repo.GetPermissionByID(ctx, permID)
	if err != nil {
		return err
	}
	if isBuiltinPermission(p.Slug) {
		return forbidden("permissões nativas não podem ser excluídas (desative-as)")
	}
	if err := s.repo.DeletePermission(ctx, permID); err != nil {
		return err
	}
	s.audit.Record(ctx, actorID, ActionPermissionDelete, "permission", permID, map[string]any{"slug": p.Slug})
	return nil
}

// UserRoles returns the roles of a user.
func (s *Service) UserRoles(ctx context.Context, userID string) ([]Role, error) {
	return s.repo.RolesOfUser(ctx, userID)
}

// SetUserRoles replaces the roles of target. The actor must outrank the
// target and every role added or removed; the Owner role is never touched.
func (s *Service) SetUserRoles(ctx context.Context, actorID, targetID string, roleSlugs []string) ([]Role, error) {
	actor, err := s.ResolveActor(ctx, actorID)
	if err != nil {
		return nil, err
	}
	target, err := s.ResolveActor(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if err := actor.CanActOnUser(targetID, target.Rank, target.IsOwner); err != nil {
		return nil, err
	}
	currentRoles, err := s.repo.RolesOfUser(ctx, targetID)
	if err != nil {
		return nil, err
	}
	current := make([]string, len(currentRoles))
	for i, r := range currentRoles {
		current[i] = r.Slug
	}
	next := dedupe(roleSlugs)
	added, removed := Diff(current, next)

	ids := make([]string, 0, len(next))
	for _, slug := range next {
		role, err := s.repo.GetRoleBySlug(ctx, slug)
		if err != nil {
			return nil, invalid("cargo inexistente: " + slug)
		}
		ids = append(ids, role.ID)
	}
	for _, slug := range append(append([]string{}, added...), removed...) {
		role, err := s.repo.GetRoleBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if err := actor.CanManageRole(*role); err != nil {
			return nil, err
		}
	}
	if err := s.repo.SetUserRoles(ctx, targetID, ids); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionUserRolesChange, "user", targetID, map[string]any{
		"added": orEmpty(added), "removed": orEmpty(removed),
	})
	return s.repo.RolesOfUser(ctx, targetID)
}

func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
