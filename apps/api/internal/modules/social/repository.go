package social

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
)

// Repository persists follows of users and categories.
type Repository struct {
	store *database.Store
}

// NewRepository builds a social Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

// UserIDByUsername resolves an active user's id.
func (r *Repository) UserIDByUsername(ctx context.Context, username string) (string, error) {
	return r.scalar(ctx, "SELECT id FROM users WHERE username = ? AND status = ?", username, "active")
}

// CategoryIDBySlug resolves an active category's id.
func (r *Repository) CategoryIDBySlug(ctx context.Context, slug string) (string, error) {
	return r.scalar(ctx, "SELECT id FROM categories WHERE slug = ? AND is_active = ?", slug, true)
}

func (r *Repository) scalar(ctx context.Context, query string, args ...any) (string, error) {
	var v string
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(query), args...).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperr.ErrNotFound
	}
	return v, err
}

// Toggle follows or unfollows a target and returns the new state. The unique
// constraint (user_id, target_type, target_id) guards against duplicates.
func (r *Repository) Toggle(ctx context.Context, userID, targetType, targetID string) (bool, error) {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var followID string
	check := r.store.Rebind("SELECT id FROM follows WHERE user_id = ? AND target_type = ? AND target_id = ?")
	scanErr := tx.QueryRowContext(ctx, check, userID, targetType, targetID).Scan(&followID)
	following := false
	switch {
	case errors.Is(scanErr, sql.ErrNoRows):
		ins := r.store.Rebind("INSERT INTO follows (id, user_id, target_type, target_id, created_at) VALUES (?, ?, ?, ?, ?)")
		if _, err := tx.ExecContext(ctx, ins, id.New(), userID, targetType, targetID, time.Now().UTC()); err != nil {
			return false, err
		}
		following = true
	case scanErr != nil:
		return false, scanErr
	default:
		if _, err := tx.ExecContext(ctx, r.store.Rebind("DELETE FROM follows WHERE id = ?"), followID); err != nil {
			return false, err
		}
	}
	return following, tx.Commit()
}

// IsFollowing reports whether userID follows the target.
func (r *Repository) IsFollowing(ctx context.Context, userID, targetType, targetID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	q := r.store.Rebind("SELECT 1 FROM follows WHERE user_id = ? AND target_type = ? AND target_id = ? LIMIT 1")
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, userID, targetType, targetID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) count(ctx context.Context, query string, args ...any) (int64, error) {
	var n int64
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(query), args...).Scan(&n)
	return n, err
}

// CountFollowers counts who follows a target.
func (r *Repository) CountFollowers(ctx context.Context, targetType, targetID string) (int64, error) {
	return r.count(ctx, "SELECT COUNT(*) FROM follows WHERE target_type = ? AND target_id = ?", targetType, targetID)
}

// CountFollowing counts the targets of a type that userID follows.
func (r *Repository) CountFollowing(ctx context.Context, userID, targetType string) (int64, error) {
	return r.count(ctx, "SELECT COUNT(*) FROM follows WHERE user_id = ? AND target_type = ?", userID, targetType)
}

// CountTopicsByAuthor counts published topics by a user.
func (r *Repository) CountTopicsByAuthor(ctx context.Context, userID string) (int64, error) {
	return r.count(ctx, "SELECT COUNT(*) FROM topics WHERE author_id = ?", userID)
}

// CountRepliesByAuthor counts non-deleted replies by a user.
func (r *Repository) CountRepliesByAuthor(ctx context.Context, userID string) (int64, error) {
	return r.count(ctx, "SELECT COUNT(*) FROM replies WHERE author_id = ? AND is_deleted = ?", userID, false)
}

// CountTopicsInCategory counts topics in a category and its direct subcategories.
func (r *Repository) CountTopicsInCategory(ctx context.Context, categoryID string) (int64, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM topics WHERE category_id = ?
        OR category_id IN (SELECT id FROM categories WHERE parent_id = ?)`, categoryID, categoryID)
}

const summaryColumns = "u.id, u.username, u.display_name, u.avatar_url, u.bio, u.reputation"

// Followers lists users following userID, most recent first.
func (r *Repository) Followers(ctx context.Context, userID string, limit, offset int) ([]UserSummary, error) {
	return r.summaries(ctx, `SELECT `+summaryColumns+` FROM follows f
        JOIN users u ON u.id = f.user_id
        WHERE f.target_type = ? AND f.target_id = ?
        ORDER BY f.created_at DESC LIMIT ? OFFSET ?`, TargetUser, userID, limit, offset)
}

// Following lists users that userID follows, most recent first.
func (r *Repository) Following(ctx context.Context, userID string, limit, offset int) ([]UserSummary, error) {
	return r.summaries(ctx, `SELECT `+summaryColumns+` FROM follows f
        JOIN users u ON u.id = f.target_id
        WHERE f.user_id = ? AND f.target_type = ?
        ORDER BY f.created_at DESC LIMIT ? OFFSET ?`, userID, TargetUser, limit, offset)
}

func (r *Repository) summaries(ctx context.Context, query string, args ...any) ([]UserSummary, error) {
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UserSummary{}
	for rows.Next() {
		var (
			u           UserSummary
			avatar, bio sql.NullString
		)
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &avatar, &bio, &u.Reputation); err != nil {
			return nil, err
		}
		if avatar.Valid {
			u.AvatarURL = &avatar.String
		}
		if bio.Valid {
			u.Bio = &bio.String
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
