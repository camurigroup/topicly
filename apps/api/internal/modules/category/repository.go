package category

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
)

const columns = "id, parent_id, name, slug, description, icon, cover_url, position, is_active, created_at, updated_at"

// Repository persists categories.
type Repository struct {
	store *database.Store
}

// NewRepository builds a category Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

type scanner interface {
	Scan(dest ...any) error
}

func scan(row scanner) (*Category, error) {
	var (
		c      Category
		parent sql.NullString
		desc   sql.NullString
		icon   sql.NullString
		cover  sql.NullString
	)
	err := row.Scan(&c.ID, &parent, &c.Name, &c.Slug, &desc, &icon, &cover, &c.Position, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if parent.Valid {
		c.ParentID = &parent.String
	}
	if desc.Valid {
		c.Description = &desc.String
	}
	if icon.Valid {
		c.Icon = &icon.String
	}
	if cover.Valid {
		c.CoverURL = &cover.String
	}
	return &c, nil
}

// Create inserts a category.
func (r *Repository) Create(ctx context.Context, c *Category) error {
	q := r.store.Rebind(`INSERT INTO categories (` + columns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q,
		c.ID, c.ParentID, c.Name, c.Slug, c.Description, c.Icon, c.CoverURL, c.Position, c.IsActive, c.CreatedAt, c.UpdatedAt)
	return err
}

// GetByID fetches a category by id.
func (r *Repository) GetByID(ctx context.Context, id string) (*Category, error) {
	q := r.store.Rebind("SELECT " + columns + " FROM categories WHERE id = ?")
	return scan(r.store.DB.QueryRowContext(ctx, q, id))
}

// GetBySlug fetches a category by slug.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	q := r.store.Rebind("SELECT " + columns + " FROM categories WHERE slug = ?")
	return scan(r.store.DB.QueryRowContext(ctx, q, slug))
}

// SlugExists reports whether a slug is already taken.
func (r *Repository) SlugExists(ctx context.Context, slug string) (bool, error) {
	q := r.store.Rebind("SELECT 1 FROM categories WHERE slug = ? LIMIT 1")
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, slug).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// List returns categories, optionally only active ones, ordered by position.
func (r *Repository) List(ctx context.Context, activeOnly bool) ([]Category, error) {
	query := "SELECT " + columns + " FROM categories"
	if activeOnly {
		query += " WHERE is_active = ?"
	}
	query += " ORDER BY position ASC, name ASC"
	query = r.store.Rebind(query)

	var rows *sql.Rows
	var err error
	if activeOnly {
		rows, err = r.store.DB.QueryContext(ctx, query, true)
	} else {
		rows, err = r.store.DB.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

// Update persists editable category fields.
func (r *Repository) Update(ctx context.Context, c *Category) error {
	q := r.store.Rebind(`UPDATE categories SET name = ?, description = ?, icon = ?, cover_url = ?, position = ?, is_active = ?, updated_at = ? WHERE id = ?`)
	res, err := r.store.DB.ExecContext(ctx, q, c.Name, c.Description, c.Icon, c.CoverURL, c.Position, c.IsActive, time.Now().UTC(), c.ID)
	return affected(res, err)
}

// Delete removes a category (cascades to children and topics).
func (r *Repository) Delete(ctx context.Context, id string) error {
	q := r.store.Rebind("DELETE FROM categories WHERE id = ?")
	res, err := r.store.DB.ExecContext(ctx, q, id)
	return affected(res, err)
}

func affected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.ErrNotFound
	}
	return nil
}
