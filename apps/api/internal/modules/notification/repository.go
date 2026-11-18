package notification

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
)

// insertBatch bounds rows per multi-row INSERT (keeps placeholder count low).
const insertBatch = 200

// Repository persists notifications and resolves notification recipients.
type Repository struct {
	store *database.Store
}

// NewRepository builds a notification Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

const notificationSelect = `SELECT n.id, n.type, n.entity_type, n.entity_id, n.message, n.link, n.is_read, n.created_at,
    a.id, a.username, a.display_name, a.avatar_url
    FROM notifications n
    LEFT JOIN users a ON a.id = n.actor_id`

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(row scanner) (*Notification, error) {
	var (
		n                                     Notification
		entityType, entityID, link            sql.NullString
		actorID, actorUser, actorName, avatar sql.NullString
	)
	if err := row.Scan(&n.ID, &n.Type, &entityType, &entityID, &n.Message, &link, &n.IsRead, &n.CreatedAt,
		&actorID, &actorUser, &actorName, &avatar); err != nil {
		return nil, err
	}
	if entityType.Valid {
		n.EntityType = &entityType.String
	}
	if entityID.Valid {
		n.EntityID = &entityID.String
	}
	if link.Valid {
		n.Link = &link.String
	}
	if actorID.Valid {
		n.Actor = &Actor{ID: actorID.String, Username: actorUser.String, DisplayName: actorName.String}
		if avatar.Valid {
			n.Actor.AvatarURL = &avatar.String
		}
	}
	return &n, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// InsertMany stores one notification per recipient inside a single transaction
// and returns them (without actor details) in recipient order.
func (r *Repository) InsertMany(ctx context.Context, recipients []string, ev Event) ([]Notification, error) {
	if len(recipients) == 0 {
		return nil, nil
	}
	now := time.Now().UTC()
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	out := make([]Notification, 0, len(recipients))
	const cols = "(id, user_id, actor_id, type, entity_type, entity_id, message, link, is_read, created_at)"
	for start := 0; start < len(recipients); start += insertBatch {
		end := min(start+insertBatch, len(recipients))
		var sb strings.Builder
		sb.WriteString("INSERT INTO notifications " + cols + " VALUES ")
		args := make([]any, 0, (end-start)*10)
		for i, uid := range recipients[start:end] {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			nid := id.New()
			args = append(args, nid, uid, nullable(ev.ActorID), ev.Type, nullable(ev.EntityType),
				nullable(ev.EntityID), ev.Message, nullable(ev.Link), false, now)
			n := Notification{ID: nid, Type: ev.Type, Message: ev.Message, CreatedAt: now}
			if ev.EntityType != "" {
				n.EntityType = &ev.EntityType
			}
			if ev.EntityID != "" {
				n.EntityID = &ev.EntityID
			}
			if ev.Link != "" {
				n.Link = &ev.Link
			}
			out = append(out, n)
		}
		if _, err := tx.ExecContext(ctx, r.store.Rebind(sb.String()), args...); err != nil {
			return nil, err
		}
	}
	return out, tx.Commit()
}

// ActorSummary returns the public summary of a user, or nil if absent.
func (r *Repository) ActorSummary(ctx context.Context, userID string) (*Actor, error) {
	if userID == "" {
		return nil, nil
	}
	q := r.store.Rebind("SELECT id, username, display_name, avatar_url FROM users WHERE id = ?")
	var (
		a      Actor
		avatar sql.NullString
	)
	err := r.store.DB.QueryRowContext(ctx, q, userID).Scan(&a.ID, &a.Username, &a.DisplayName, &avatar)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if avatar.Valid {
		a.AvatarURL = &avatar.String
	}
	return &a, nil
}

// List returns a page of a user's notifications, newest first.
func (r *Repository) List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]Notification, error) {
	query := notificationSelect + " WHERE n.user_id = ?"
	args := []any{userID}
	if unreadOnly {
		query += " AND n.is_read = ?"
		args = append(args, false)
	}
	query += " ORDER BY n.created_at DESC, n.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Notification{}
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}

// Count counts a user's notifications (optionally only unread ones).
func (r *Repository) Count(ctx context.Context, userID string, unreadOnly bool) (int64, error) {
	query := "SELECT COUNT(*) FROM notifications WHERE user_id = ?"
	args := []any{userID}
	if unreadOnly {
		query += " AND is_read = ?"
		args = append(args, false)
	}
	var n int64
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(query), args...).Scan(&n)
	return n, err
}

// MarkRead marks one of the user's notifications as read.
func (r *Repository) MarkRead(ctx context.Context, userID, notificationID string) error {
	q := r.store.Rebind("UPDATE notifications SET is_read = ? WHERE id = ? AND user_id = ?")
	res, err := r.store.DB.ExecContext(ctx, q, true, notificationID, userID)
	return affected(res, err)
}

// MarkAllRead marks every unread notification of the user as read.
func (r *Repository) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	q := r.store.Rebind("UPDATE notifications SET is_read = ? WHERE user_id = ? AND is_read = ?")
	res, err := r.store.DB.ExecContext(ctx, q, true, userID, false)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Delete removes one of the user's notifications.
func (r *Repository) Delete(ctx context.Context, userID, notificationID string) error {
	q := r.store.Rebind("DELETE FROM notifications WHERE id = ? AND user_id = ?")
	res, err := r.store.DB.ExecContext(ctx, q, notificationID, userID)
	return affected(res, err)
}

// FollowerIDs lists the users following a target (user, category or topic).
func (r *Repository) FollowerIDs(ctx context.Context, targetType, targetID string) ([]string, error) {
	q := r.store.Rebind("SELECT user_id FROM follows WHERE target_type = ? AND target_id = ?")
	return r.ids(ctx, q, targetType, targetID)
}

// UserIDsByUsernames resolves usernames (case-insensitive) to active user ids.
func (r *Repository) UserIDsByUsernames(ctx context.Context, usernames []string) ([]string, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(usernames)), ", ")
	q := r.store.Rebind("SELECT id FROM users WHERE status = ? AND LOWER(username) IN (" + placeholders + ")")
	args := make([]any, 0, len(usernames)+1)
	args = append(args, "active")
	for _, u := range usernames {
		args = append(args, strings.ToLower(u))
	}
	return r.ids(ctx, q, args...)
}

// ActiveUserIDs lists every active user (used for admin broadcasts).
func (r *Repository) ActiveUserIDs(ctx context.Context) ([]string, error) {
	return r.ids(ctx, r.store.Rebind("SELECT id FROM users WHERE status = ?"), "active")
}

func (r *Repository) ids(ctx context.Context, query string, args ...any) ([]string, error) {
	rows, err := r.store.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
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
