package topic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// Repository persists topics, replies, likes and follows.
type Repository struct {
	store *database.Store
}

// NewRepository builds a topic Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

const topicSelect = `SELECT t.id, t.category_id, c.slug, t.title, t.slug, t.content,
    u.id, u.username, u.display_name, u.avatar_url,
    t.is_pinned, t.is_locked, t.is_featured, t.status,
    t.view_count, t.reply_count, t.like_count, t.follower_count,
    t.created_at, t.updated_at, t.last_activity_at
    FROM topics t
    JOIN users u ON u.id = t.author_id
    JOIN categories c ON c.id = t.category_id`

type scanner interface {
	Scan(dest ...any) error
}

func scanTopic(row scanner) (*Topic, error) {
	var (
		t      Topic
		avatar sql.NullString
	)
	err := row.Scan(
		&t.ID, &t.CategoryID, &t.CategorySlug, &t.Title, &t.Slug, &t.Content,
		&t.Author.ID, &t.Author.Username, &t.Author.DisplayName, &avatar,
		&t.IsPinned, &t.IsLocked, &t.IsFeatured, &t.Status,
		&t.ViewCount, &t.ReplyCount, &t.LikeCount, &t.FollowerCount,
		&t.CreatedAt, &t.UpdatedAt, &t.LastActivity,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if avatar.Valid {
		t.Author.AvatarURL = &avatar.String
	}
	return &t, nil
}

// CreateTopic inserts a topic.
func (r *Repository) CreateTopic(ctx context.Context, t *Topic, authorID string) error {
	q := r.store.Rebind(`INSERT INTO topics
        (id, category_id, author_id, title, slug, content, is_pinned, is_locked, is_featured, status,
         view_count, reply_count, like_count, follower_count, created_at, updated_at, last_activity_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q,
		t.ID, t.CategoryID, authorID, t.Title, t.Slug, t.Content,
		t.IsPinned, t.IsLocked, t.IsFeatured, t.Status,
		t.CreatedAt, t.UpdatedAt, t.LastActivity)
	return err
}

// GetTopicByID fetches a topic with author and category slug.
func (r *Repository) GetTopicByID(ctx context.Context, id string) (*Topic, error) {
	q := r.store.Rebind(topicSelect + " WHERE t.id = ?")
	return scanTopic(r.store.DB.QueryRowContext(ctx, q, id))
}

// TopicAuthor returns the author id of a topic.
func (r *Repository) TopicAuthor(ctx context.Context, id string) (string, error) {
	q := r.store.Rebind("SELECT author_id FROM topics WHERE id = ?")
	var author string
	err := r.store.DB.QueryRowContext(ctx, q, id).Scan(&author)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperr.ErrNotFound
	}
	return author, err
}

// ListTopics returns a paginated list, optionally filtered by category id.
func (r *Repository) ListTopics(ctx context.Context, categoryID string, limit, offset int) ([]Topic, error) {
	query := topicSelect
	args := []any{}
	if categoryID != "" {
		query += " WHERE t.category_id = ?"
		args = append(args, categoryID)
	}
	query += " ORDER BY t.is_pinned DESC, t.last_activity_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Topic{}
	for rows.Next() {
		t, err := scanTopic(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

// CountTopics counts topics, optionally by category.
func (r *Repository) CountTopics(ctx context.Context, categoryID string) (int64, error) {
	query := "SELECT COUNT(*) FROM topics"
	args := []any{}
	if categoryID != "" {
		query += " WHERE category_id = ?"
		args = append(args, categoryID)
	}
	var n int64
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(query), args...).Scan(&n)
	return n, err
}

// IncrementView bumps the view counter.
func (r *Repository) IncrementView(ctx context.Context, id string) error {
	q := r.store.Rebind("UPDATE topics SET view_count = view_count + 1 WHERE id = ?")
	_, err := r.store.DB.ExecContext(ctx, q, id)
	return err
}

// UpdateTopic edits title, slug and content.
func (r *Repository) UpdateTopic(ctx context.Context, id, title, slug, content string) error {
	q := r.store.Rebind("UPDATE topics SET title = ?, slug = ?, content = ?, updated_at = ? WHERE id = ?")
	res, err := r.store.DB.ExecContext(ctx, q, title, slug, content, time.Now().UTC(), id)
	return affected(res, err)
}

// SetFlag sets a boolean flag (is_pinned, is_locked, is_featured) on a topic.
func (r *Repository) SetFlag(ctx context.Context, id, column string, value bool) error {
	q := r.store.Rebind("UPDATE topics SET " + column + " = ?, updated_at = ? WHERE id = ?")
	res, err := r.store.DB.ExecContext(ctx, q, value, time.Now().UTC(), id)
	return affected(res, err)
}

// DeleteTopic removes a topic (cascades to replies and content_edits via app).
func (r *Repository) DeleteTopic(ctx context.Context, id string) error {
	q := r.store.Rebind("DELETE FROM topics WHERE id = ?")
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

// CategoryIDBySlug resolves a category slug to its id.
func (r *Repository) CategoryIDBySlug(ctx context.Context, slug string) (string, error) {
	q := r.store.Rebind("SELECT id FROM categories WHERE slug = ?")
	var id string
	err := r.store.DB.QueryRowContext(ctx, q, slug).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperr.ErrNotFound
	}
	return id, err
}

// CategoryInfo returns a category's name and parent id.
func (r *Repository) CategoryInfo(ctx context.Context, id string) (string, *string, error) {
	q := r.store.Rebind("SELECT name, parent_id FROM categories WHERE id = ?")
	var (
		name   string
		parent sql.NullString
	)
	err := r.store.DB.QueryRowContext(ctx, q, id).Scan(&name, &parent)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, apperr.ErrNotFound
	}
	if err != nil {
		return "", nil, err
	}
	if parent.Valid {
		return name, &parent.String, nil
	}
	return name, nil, nil
}

// CategoryExists reports whether a category id exists.
func (r *Repository) CategoryExists(ctx context.Context, id string) (bool, error) {
	q := r.store.Rebind("SELECT 1 FROM categories WHERE id = ? LIMIT 1")
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
