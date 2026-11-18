package rbac

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// Repository persists roles, permissions and their assignments.
type Repository struct {
	store *database.Store
}

// NewRepository builds an rbac Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

// GetRoleBySlug fetches a role by its slug.
func (r *Repository) GetRoleBySlug(ctx context.Context, slug string) (*Role, error) {
	q := r.store.Rebind("SELECT id, slug, name, rank_order, is_system, created_at, updated_at FROM roles WHERE slug = ?")
	var role Role
	err := r.store.DB.QueryRowContext(ctx, q, slug).Scan(
		&role.ID, &role.Slug, &role.Name, &role.RankOrder, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// EnsureRole inserts a role if its slug is new, returning the existing or new id.
func (r *Repository) EnsureRole(ctx context.Context, slug, name string, rank int, isSystem bool) (string, error) {
	if existing, err := r.GetRoleBySlug(ctx, slug); err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, apperr.ErrNotFound) {
		return "", err
	}
	id := newID()
	now := time.Now().UTC()
	q := r.store.Rebind("INSERT INTO roles (id, slug, name, rank_order, is_system, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if _, err := r.store.DB.ExecContext(ctx, q, id, slug, name, rank, isSystem, now, now); err != nil {
		return "", err
	}
	return id, nil
}

// EnsurePermission inserts a permission if its slug is new, returning the id.
func (r *Repository) EnsurePermission(ctx context.Context, slug, name string) (string, error) {
	q := r.store.Rebind("SELECT id FROM permissions WHERE slug = ?")
	var id string
	err := r.store.DB.QueryRowContext(ctx, q, slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	id = newID()
	ins := r.store.Rebind("INSERT INTO permissions (id, slug, name, is_active, created_at) VALUES (?, ?, ?, ?, ?)")
	if _, err := r.store.DB.ExecContext(ctx, ins, id, slug, name, true, time.Now().UTC()); err != nil {
		return "", err
	}
	return id, nil
}

// AssignPermissionToRole links a permission to a role (idempotent).
func (r *Repository) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	check := r.store.Rebind("SELECT 1 FROM role_permissions WHERE role_id = ? AND permission_id = ?")
	var one int
	err := r.store.DB.QueryRowContext(ctx, check, roleID, permissionID).Scan(&one)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	ins := r.store.Rebind("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)")
	_, err = r.store.DB.ExecContext(ctx, ins, roleID, permissionID)
	return err
}

// AssignRoleToUser links a role to a user (idempotent).
func (r *Repository) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	check := r.store.Rebind("SELECT 1 FROM user_roles WHERE user_id = ? AND role_id = ?")
	var one int
	err := r.store.DB.QueryRowContext(ctx, check, userID, roleID).Scan(&one)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	ins := r.store.Rebind("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)")
	_, err = r.store.DB.ExecContext(ctx, ins, userID, roleID)
	return err
}

// PermissionSlugsForUser returns active permission slugs granted to a user.
func (r *Repository) PermissionSlugsForUser(ctx context.Context, userID string) ([]string, error) {
	q := r.store.Rebind(`SELECT DISTINCT p.slug
        FROM user_roles ur
        JOIN role_permissions rp ON rp.role_id = ur.role_id
        JOIN permissions p ON p.id = rp.permission_id
        WHERE ur.user_id = ? AND p.is_active = ?`)
	rows, err := r.store.DB.QueryContext(ctx, q, userID, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		slugs = append(slugs, s)
	}
	return slugs, rows.Err()
}

// RoleSlugsForUser returns the role slugs assigned to a user.
func (r *Repository) RoleSlugsForUser(ctx context.Context, userID string) ([]string, error) {
	q := r.store.Rebind(`SELECT r.slug FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = ?`)
	rows, err := r.store.DB.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		slugs = append(slugs, s)
	}
	return slugs, rows.Err()
}
