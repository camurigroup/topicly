package search

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/modules/category"
)

// UserHit is a user search result.
type UserHit struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Bio         *string `json:"bio"`
	Reputation  int     `json:"reputation"`
}

// Repository runs the user and category searches and resolves filter values.
type Repository struct {
	store *database.Store
}

// NewRepository builds a search Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

// UserIDByUsername resolves a username (case-insensitive); "" when unknown.
func (r *Repository) UserIDByUsername(ctx context.Context, username string) (string, error) {
	return r.optionalID(ctx, "SELECT id FROM users WHERE LOWER(username) = ?", strings.ToLower(username))
}

// CategoryIDBySlug resolves an active category slug; "" when unknown.
func (r *Repository) CategoryIDBySlug(ctx context.Context, slug string) (string, error) {
	return r.optionalID(ctx, "SELECT id FROM categories WHERE slug = ? AND is_active = ?", slug, true)
}

func (r *Repository) optionalID(ctx context.Context, query string, args ...any) (string, error) {
	var v string
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(query), args...).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// SearchUsers matches active users by username or display name.
func (r *Repository) SearchUsers(ctx context.Context, terms []string, sort string, limit, offset int) ([]UserHit, int64, error) {
	clauses := []string{"status = ?"}
	args := []any{"active"}
	for _, t := range terms {
		p := database.ContainsPattern(t)
		clauses = append(clauses, "(LOWER(username) LIKE ? OR LOWER(display_name) LIKE ?)")
		args = append(args, p, p)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")

	var total int64
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind("SELECT COUNT(*) FROM users"+where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []UserHit{}, 0, nil
	}
	order := "reputation DESC, created_at ASC"
	if sort == "recent" {
		order = "created_at DESC"
	}
	q := "SELECT id, username, display_name, avatar_url, bio, reputation FROM users" + where +
		" ORDER BY " + order + " LIMIT ? OFFSET ?"
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []UserHit{}
	for rows.Next() {
		var (
			u           UserHit
			avatar, bio sql.NullString
		)
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &avatar, &bio, &u.Reputation); err != nil {
			return nil, 0, err
		}
		if avatar.Valid {
			u.AvatarURL = &avatar.String
		}
		if bio.Valid {
			u.Bio = &bio.String
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

// SearchCategories matches active categories by name or description.
func (r *Repository) SearchCategories(ctx context.Context, terms []string, limit, offset int) ([]category.Category, int64, error) {
	clauses := []string{"is_active = ?"}
	args := []any{true}
	for _, t := range terms {
		p := database.ContainsPattern(t)
		clauses = append(clauses, "(LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ?)")
		args = append(args, p, p)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")

	var total int64
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind("SELECT COUNT(*) FROM categories"+where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []category.Category{}, 0, nil
	}
	q := `SELECT id, parent_id, name, slug, description, icon, cover_url, position, is_active, created_at, updated_at
        FROM categories` + where + " ORDER BY position ASC, name ASC LIMIT ? OFFSET ?"
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []category.Category{}
	for rows.Next() {
		var (
			c                            category.Category
			parent, desc, icon, coverURL sql.NullString
		)
		if err := rows.Scan(&c.ID, &parent, &c.Name, &c.Slug, &desc, &icon, &coverURL,
			&c.Position, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		c.ParentID = nullStr(parent)
		c.Description = nullStr(desc)
		c.Icon = nullStr(icon)
		c.CoverURL = nullStr(coverURL)
		out = append(out, c)
	}
	return out, total, rows.Err()
}

func nullStr(v sql.NullString) *string {
	if v.Valid {
		return &v.String
	}
	return nil
}
