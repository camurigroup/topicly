package rbac

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/shared/apperr"
)

const roleColumns = "id, slug, name, rank_order, is_system, created_at, updated_at"

func scanRole(row interface{ Scan(...any) error }) (*Role, error) {
	var role Role
	err := row.Scan(&role.ID, &role.Slug, &role.Name, &role.RankOrder, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles returns every role ordered by hierarchy (highest first).
func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.store.DB.QueryContext(ctx, "SELECT "+roleColumns+" FROM roles ORDER BY rank_order DESC, name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Role{}
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *role)
	}
	return out, rows.Err()
}

// GetRoleByID fetches a role by id.
func (r *Repository) GetRoleByID(ctx context.Context, id string) (*Role, error) {
	q := r.store.Rebind("SELECT " + roleColumns + " FROM roles WHERE id = ?")
	return scanRole(r.store.DB.QueryRowContext(ctx, q, id))
}

// CreateRole inserts a custom role.
func (r *Repository) CreateRole(ctx context.Context, role *Role) error {
	q := r.store.Rebind("INSERT INTO roles (" + roleColumns + ") VALUES (?, ?, ?, ?, ?, ?, ?)")
	_, err := r.store.DB.ExecContext(ctx, q, role.ID, role.Slug, role.Name, role.RankOrder, role.IsSystem, role.CreatedAt, role.UpdatedAt)
	return err
}

// UpdateRole changes a role's name and rank.
func (r *Repository) UpdateRole(ctx context.Context, id, name string, rank int) error {
	q := r.store.Rebind("UPDATE roles SET name = ?, rank_order = ?, updated_at = ? WHERE id = ?")
	_, err := r.store.DB.ExecContext(ctx, q, name, rank, time.Now().UTC(), id)
	return err
}

// DeleteRole removes a role (assignments cascade).
func (r *Repository) DeleteRole(ctx context.Context, id string) error {
	_, err := r.store.DB.ExecContext(ctx, r.store.Rebind("DELETE FROM roles WHERE id = ?"), id)
	return err
}

// RolePermissions maps role id -> permission slugs (including inactive ones).
func (r *Repository) RolePermissions(ctx context.Context) (map[string][]string, error) {
	rows, err := r.store.DB.QueryContext(ctx, `SELECT rp.role_id, p.slug FROM role_permissions rp
        JOIN permissions p ON p.id = rp.permission_id ORDER BY p.slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var roleID, slug string
		if err := rows.Scan(&roleID, &slug); err != nil {
			return nil, err
		}
		out[roleID] = append(out[roleID], slug)
	}
	return out, rows.Err()
}

// RoleMemberCounts maps role id -> number of users holding it.
func (r *Repository) RoleMemberCounts(ctx context.Context) (map[string]int, error) {
	rows, err := r.store.DB.QueryContext(ctx, "SELECT role_id, COUNT(*) FROM user_roles GROUP BY role_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}

// SetRolePermissions replaces the permission set of a role atomically.
// Unknown slugs yield a validation error.
func (r *Repository) SetRolePermissions(ctx context.Context, roleID string, slugs []string) error {
	ids, err := r.permissionIDs(ctx, slugs)
	if err != nil {
		return err
	}
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, r.store.Rebind("DELETE FROM role_permissions WHERE role_id = ?"), roleID); err != nil {
		return err
	}
	ins := r.store.Rebind("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)")
	for _, pid := range ids {
		if _, err := tx.ExecContext(ctx, ins, roleID, pid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) permissionIDs(ctx context.Context, slugs []string) ([]string, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(slugs)), ", ")
	args := make([]any, len(slugs))
	for i, s := range slugs {
		args[i] = s
	}
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind("SELECT id FROM permissions WHERE slug IN ("+placeholders+")"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) != len(slugs) {
		return nil, invalid("permissão inexistente na lista")
	}
	return ids, nil
}

const permissionColumns = "id, slug, name, description, is_active, created_at"

func scanPermission(row interface{ Scan(...any) error }) (*Permission, error) {
	var (
		p    Permission
		desc sql.NullString
	)
	err := row.Scan(&p.ID, &p.Slug, &p.Name, &desc, &p.IsActive, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	return &p, nil
}

// ListPermissions returns the permission catalog.
func (r *Repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.store.DB.QueryContext(ctx, "SELECT "+permissionColumns+" FROM permissions ORDER BY slug")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Permission{}
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// GetPermissionByID fetches a permission by id.
func (r *Repository) GetPermissionByID(ctx context.Context, id string) (*Permission, error) {
	q := r.store.Rebind("SELECT " + permissionColumns + " FROM permissions WHERE id = ?")
	return scanPermission(r.store.DB.QueryRowContext(ctx, q, id))
}

// CreatePermission inserts a custom permission.
func (r *Repository) CreatePermission(ctx context.Context, p *Permission) error {
	q := r.store.Rebind("INSERT INTO permissions (" + permissionColumns + ") VALUES (?, ?, ?, ?, ?, ?)")
	_, err := r.store.DB.ExecContext(ctx, q, p.ID, p.Slug, p.Name, p.Description, p.IsActive, p.CreatedAt)
	return err
}

// SetPermissionActive toggles a permission globally.
func (r *Repository) SetPermissionActive(ctx context.Context, id string, active bool) error {
	_, err := r.store.DB.ExecContext(ctx, r.store.Rebind("UPDATE permissions SET is_active = ? WHERE id = ?"), active, id)
	return err
}

// DeletePermission removes a permission (role links cascade).
func (r *Repository) DeletePermission(ctx context.Context, id string) error {
	_, err := r.store.DB.ExecContext(ctx, r.store.Rebind("DELETE FROM permissions WHERE id = ?"), id)
	return err
}

// RolesOfUser returns the full roles assigned to a user (highest first).
func (r *Repository) RolesOfUser(ctx context.Context, userID string) ([]Role, error) {
	q := r.store.Rebind(`SELECT r.id, r.slug, r.name, r.rank_order, r.is_system, r.created_at, r.updated_at
        FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = ? ORDER BY r.rank_order DESC`)
	rows, err := r.store.DB.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Role{}
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *role)
	}
	return out, rows.Err()
}

// SetUserRoles replaces the roles of a user atomically.
func (r *Repository) SetUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, r.store.Rebind("DELETE FROM user_roles WHERE user_id = ?"), userID); err != nil {
		return err
	}
	ins := r.store.Rebind("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)")
	for _, id := range roleIDs {
		if _, err := tx.ExecContext(ctx, ins, userID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
