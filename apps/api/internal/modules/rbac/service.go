package rbac

import (
	"context"

	"github.com/forum-framework/api/internal/modules/audit"
)

// Service exposes RBAC business logic.
type Service struct {
	repo  *Repository
	audit *audit.Service
}

// NewService builds an rbac Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// PermissionsForUser returns the active permission slugs for a user.
func (s *Service) PermissionsForUser(ctx context.Context, userID string) ([]string, error) {
	return s.repo.PermissionSlugsForUser(ctx, userID)
}

// RolesForUser returns the role slugs assigned to a user.
func (s *Service) RolesForUser(ctx context.Context, userID string) ([]string, error) {
	return s.repo.RoleSlugsForUser(ctx, userID)
}

// HasPermission reports whether a user holds a given permission slug.
func (s *Service) HasPermission(ctx context.Context, userID, slug string) (bool, error) {
	slugs, err := s.repo.PermissionSlugsForUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, s := range slugs {
		if s == slug {
			return true, nil
		}
	}
	return false, nil
}

// AssignRoleBySlug assigns a role (by slug) to a user.
func (s *Service) AssignRoleBySlug(ctx context.Context, userID, roleSlug string) error {
	role, err := s.repo.GetRoleBySlug(ctx, roleSlug)
	if err != nil {
		return err
	}
	return s.repo.AssignRoleToUser(ctx, userID, role.ID)
}

// allPermissions is the default catalog seeded on startup. Owners receive all.
var allPermissions = []struct{ Slug, Name string }{
	{PermUsersManage, "Gerenciar usuários"},
	{PermUsersBan, "Banir usuários"},
	{PermUsersMute, "Mutar usuários"},
	{PermRolesManage, "Gerenciar cargos"},
	{PermPermissionsManage, "Gerenciar permissões"},
	{PermCategoriesManage, "Gerenciar categorias"},
	{PermTopicsEdit, "Editar tópicos"},
	{PermTopicsDelete, "Excluir tópicos"},
	{PermTopicsLock, "Bloquear tópicos"},
	{PermTopicsPin, "Fixar tópicos"},
	{PermRepliesEdit, "Editar respostas"},
	{PermRepliesDelete, "Excluir respostas"},
	{PermReportsManage, "Gerenciar denúncias"},
	{PermWarningsApply, "Aplicar advertências"},
	{PermSettingsManage, "Gerenciar configurações"},
	{PermStorageManage, "Gerenciar armazenamento"},
	{PermLogsView, "Ver logs"},
	{PermMetricsView, "Ver métricas"},
}

var adminPermissions = []string{
	PermUsersManage, PermUsersBan, PermUsersMute, PermCategoriesManage,
	PermTopicsEdit, PermTopicsDelete, PermTopicsLock, PermTopicsPin,
	PermRepliesEdit, PermRepliesDelete, PermReportsManage, PermWarningsApply,
	PermMetricsView, PermLogsView,
}

var moderatorPermissions = []string{
	PermTopicsEdit, PermTopicsDelete, PermTopicsLock,
	PermRepliesEdit, PermRepliesDelete,
	PermUsersBan, PermUsersMute, PermReportsManage, PermWarningsApply,
}

// Seed creates the default roles and permissions if missing (idempotent).
// Owner always receives the full permission set, including newly added ones.
func (s *Service) Seed(ctx context.Context) error {
	permIDs := make(map[string]string, len(allPermissions))
	for _, p := range allPermissions {
		pid, err := s.repo.EnsurePermission(ctx, p.Slug, p.Name)
		if err != nil {
			return err
		}
		permIDs[p.Slug] = pid
	}

	ownerID, err := s.repo.EnsureRole(ctx, RoleOwner, "Dono", 100, true)
	if err != nil {
		return err
	}
	adminID, err := s.repo.EnsureRole(ctx, RoleAdmin, "Administrador", 50, false)
	if err != nil {
		return err
	}
	modID, err := s.repo.EnsureRole(ctx, RoleModerator, "Moderador", 20, false)
	if err != nil {
		return err
	}

	// Owner: every permission.
	for _, pid := range permIDs {
		if err := s.repo.AssignPermissionToRole(ctx, ownerID, pid); err != nil {
			return err
		}
	}
	if err := s.assign(ctx, adminID, permIDs, adminPermissions); err != nil {
		return err
	}
	if err := s.assign(ctx, modID, permIDs, moderatorPermissions); err != nil {
		return err
	}
	return nil
}

func (s *Service) assign(ctx context.Context, roleID string, permIDs map[string]string, slugs []string) error {
	for _, slug := range slugs {
		pid, ok := permIDs[slug]
		if !ok {
			continue
		}
		if err := s.repo.AssignPermissionToRole(ctx, roleID, pid); err != nil {
			return err
		}
	}
	return nil
}
