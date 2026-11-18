package topic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
)

const replySelect = `SELECT r.id, r.topic_id, r.parent_id, r.content,
    u.id, u.username, u.display_name, u.avatar_url,
    r.like_count, r.is_deleted, r.created_at, r.edited_at
    FROM replies r JOIN users u ON u.id = r.author_id`

func scanReply(row scanner) (*Reply, error) {
	var (
		rp       Reply
		parent   sql.NullString
		avatar   sql.NullString
		editedAt sql.NullTime
	)
	err := row.Scan(
		&rp.ID, &rp.TopicID, &parent, &rp.Content,
		&rp.Author.ID, &rp.Author.Username, &rp.Author.DisplayName, &avatar,
		&rp.LikeCount, &rp.IsDeleted, &rp.CreatedAt, &editedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if parent.Valid {
		rp.ParentID = &parent.String
	}
	if avatar.Valid {
		rp.Author.AvatarURL = &avatar.String
	}
	if editedAt.Valid {
		t := editedAt.Time
		rp.EditedAt = &t
	}
	if rp.IsDeleted {
		rp.Content = ""
	}
	return &rp, nil
}

// CreateReply inserts a reply and bumps the topic reply count and activity.
func (r *Repository) CreateReply(ctx context.Context, rp *Reply, authorID string) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	ins := r.store.Rebind(`INSERT INTO replies (id, topic_id, author_id, parent_id, content, like_count, is_deleted, created_at, updated_at, edited_at)
        VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, NULL)`)
	if _, err := tx.ExecContext(ctx, ins, rp.ID, rp.TopicID, authorID, rp.ParentID, rp.Content, false, now, now); err != nil {
		return err
	}
	upd := r.store.Rebind("UPDATE topics SET reply_count = reply_count + 1, last_activity_at = ? WHERE id = ?")
	if _, err := tx.ExecContext(ctx, upd, now, rp.TopicID); err != nil {
		return err
	}
	return tx.Commit()
}

// GetReplyByID fetches a single reply.
func (r *Repository) GetReplyByID(ctx context.Context, id string) (*Reply, error) {
	q := r.store.Rebind(replySelect + " WHERE r.id = ?")
	return scanReply(r.store.DB.QueryRowContext(ctx, q, id))
}

// ReplyMeta returns the author and topic ids of a reply.
func (r *Repository) ReplyMeta(ctx context.Context, id string) (authorID, topicID string, err error) {
	q := r.store.Rebind("SELECT author_id, topic_id FROM replies WHERE id = ?")
	err = r.store.DB.QueryRowContext(ctx, q, id).Scan(&authorID, &topicID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", apperr.ErrNotFound
	}
	return authorID, topicID, err
}

// ListReplies returns a paginated list of replies for a topic.
func (r *Repository) ListReplies(ctx context.Context, topicID string, limit, offset int) ([]Reply, error) {
	q := r.store.Rebind(replySelect + " WHERE r.topic_id = ? ORDER BY r.created_at ASC LIMIT ? OFFSET ?")
	rows, err := r.store.DB.QueryContext(ctx, q, topicID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Reply
	for rows.Next() {
		rp, err := scanReply(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rp)
	}
	return out, rows.Err()
}

// CountReplies counts replies in a topic.
func (r *Repository) CountReplies(ctx context.Context, topicID string) (int64, error) {
	q := r.store.Rebind("SELECT COUNT(*) FROM replies WHERE topic_id = ?")
	var n int64
	err := r.store.DB.QueryRowContext(ctx, q, topicID).Scan(&n)
	return n, err
}

// UpdateReply edits a reply's content and stamps edited_at.
func (r *Repository) UpdateReply(ctx context.Context, id, content string) error {
	now := time.Now().UTC()
	q := r.store.Rebind("UPDATE replies SET content = ?, updated_at = ?, edited_at = ? WHERE id = ?")
	res, err := r.store.DB.ExecContext(ctx, q, content, now, now, id)
	return affected(res, err)
}

// SoftDeleteReply marks a reply deleted and decrements the topic reply count.
func (r *Repository) SoftDeleteReply(ctx context.Context, id, topicID string) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	del := r.store.Rebind("UPDATE replies SET is_deleted = ?, updated_at = ? WHERE id = ? AND is_deleted = ?")
	res, err := tx.ExecContext(ctx, del, true, time.Now().UTC(), id, false)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.ErrNotFound
	}
	dec := r.store.Rebind("UPDATE topics SET reply_count = reply_count - 1 WHERE id = ? AND reply_count > 0")
	if _, err := tx.ExecContext(ctx, dec, topicID); err != nil {
		return err
	}
	return tx.Commit()
}

// AddEdit records a content edit history entry.
func (r *Repository) AddEdit(ctx context.Context, entityType, entityID, editorID, previous string) error {
	q := r.store.Rebind("INSERT INTO content_edits (id, entity_type, entity_id, editor_id, previous_content, created_at) VALUES (?, ?, ?, ?, ?, ?)")
	_, err := r.store.DB.ExecContext(ctx, q, id.New(), entityType, entityID, editorID, previous, time.Now().UTC())
	return err
}

// ToggleLike adds or removes a like, updating the target counter atomically.
// It returns the new liked state and like count.
func (r *Repository) ToggleLike(ctx context.Context, userID, targetType, targetID string) (bool, int, error) {
	table := "topics"
	if targetType == TargetReply {
		table = "replies"
	}
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	check := r.store.Rebind("SELECT id FROM likes WHERE user_id = ? AND target_type = ? AND target_id = ?")
	var likeID string
	scanErr := tx.QueryRowContext(ctx, check, userID, targetType, targetID).Scan(&likeID)

	liked := false
	if errors.Is(scanErr, sql.ErrNoRows) {
		ins := r.store.Rebind("INSERT INTO likes (id, user_id, target_type, target_id, created_at) VALUES (?, ?, ?, ?, ?)")
		if _, err := tx.ExecContext(ctx, ins, id.New(), userID, targetType, targetID, time.Now().UTC()); err != nil {
			return false, 0, err
		}
		if _, err := tx.ExecContext(ctx, r.store.Rebind("UPDATE "+table+" SET like_count = like_count + 1 WHERE id = ?"), targetID); err != nil {
			return false, 0, err
		}
		liked = true
	} else if scanErr != nil {
		return false, 0, scanErr
	} else {
		if _, err := tx.ExecContext(ctx, r.store.Rebind("DELETE FROM likes WHERE id = ?"), likeID); err != nil {
			return false, 0, err
		}
		if _, err := tx.ExecContext(ctx, r.store.Rebind("UPDATE "+table+" SET like_count = like_count - 1 WHERE id = ? AND like_count > 0"), targetID); err != nil {
			return false, 0, err
		}
	}

	var count int
	if err := tx.QueryRowContext(ctx, r.store.Rebind("SELECT like_count FROM "+table+" WHERE id = ?"), targetID).Scan(&count); err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return liked, count, nil
}

// ToggleFollowTopic follows or unfollows a topic, updating the follower count.
func (r *Repository) ToggleFollowTopic(ctx context.Context, userID, topicID string) (bool, int, error) {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	check := r.store.Rebind("SELECT id FROM follows WHERE user_id = ? AND target_type = ? AND target_id = ?")
	var followID string
	scanErr := tx.QueryRowContext(ctx, check, userID, FollowTopic, topicID).Scan(&followID)

	following := false
	if errors.Is(scanErr, sql.ErrNoRows) {
		ins := r.store.Rebind("INSERT INTO follows (id, user_id, target_type, target_id, created_at) VALUES (?, ?, ?, ?, ?)")
		if _, err := tx.ExecContext(ctx, ins, id.New(), userID, FollowTopic, topicID, time.Now().UTC()); err != nil {
			return false, 0, err
		}
		if _, err := tx.ExecContext(ctx, r.store.Rebind("UPDATE topics SET follower_count = follower_count + 1 WHERE id = ?"), topicID); err != nil {
			return false, 0, err
		}
		following = true
	} else if scanErr != nil {
		return false, 0, scanErr
	} else {
		if _, err := tx.ExecContext(ctx, r.store.Rebind("DELETE FROM follows WHERE id = ?"), followID); err != nil {
			return false, 0, err
		}
		if _, err := tx.ExecContext(ctx, r.store.Rebind("UPDATE topics SET follower_count = follower_count - 1 WHERE id = ? AND follower_count > 0"), topicID); err != nil {
			return false, 0, err
		}
	}

	var count int
	if err := tx.QueryRowContext(ctx, r.store.Rebind("SELECT follower_count FROM topics WHERE id = ?"), topicID).Scan(&count); err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return following, count, nil
}

// IsLiked reports whether a user liked a target.
func (r *Repository) IsLiked(ctx context.Context, userID, targetType, targetID string) (bool, error) {
	q := r.store.Rebind("SELECT 1 FROM likes WHERE user_id = ? AND target_type = ? AND target_id = ? LIMIT 1")
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, userID, targetType, targetID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// IsFollowingTopic reports whether a user follows a topic.
func (r *Repository) IsFollowingTopic(ctx context.Context, userID, topicID string) (bool, error) {
	q := r.store.Rebind("SELECT 1 FROM follows WHERE user_id = ? AND target_type = ? AND target_id = ? LIMIT 1")
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, userID, FollowTopic, topicID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// LikedSet returns which of targetIDs the user has liked, in a single query.
func (r *Repository) LikedSet(ctx context.Context, userID, targetType string, targetIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(targetIDs))
	if userID == "" || len(targetIDs) == 0 {
		return out, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(targetIDs)), ", ")
	q := r.store.Rebind("SELECT target_id FROM likes WHERE user_id = ? AND target_type = ? AND target_id IN (" + placeholders + ")")
	args := make([]any, 0, len(targetIDs)+2)
	args = append(args, userID, targetType)
	for _, id := range targetIDs {
		args = append(args, id)
	}
	rows, err := r.store.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
