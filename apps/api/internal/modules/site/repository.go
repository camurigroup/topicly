package site

import (
	"context"
	"database/sql"
	"time"

	"github.com/forum-framework/api/internal/database"
)

// Repository persists site settings and runs the home page aggregates.
type Repository struct {
	store *database.Store
}

// NewRepository builds a site Repository.
func NewRepository(store *database.Store) *Repository {
	return &Repository{store: store}
}

// Values returns every stored setting.
func (r *Repository) Values(ctx context.Context) (map[string]string, error) {
	rows, err := r.store.DB.QueryContext(ctx, "SELECT setting_key, value FROM site_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// InsertMissing stores defaults for keys that have no value yet (idempotent).
func (r *Repository) InsertMissing(ctx context.Context, defaults map[string]string) error {
	existing, err := r.Values(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	q := r.store.Rebind("INSERT INTO site_settings (setting_key, value, updated_at) VALUES (?, ?, ?)")
	for k, v := range defaults {
		if _, ok := existing[k]; ok {
			continue
		}
		if _, err := r.store.DB.ExecContext(ctx, q, k, v, now); err != nil {
			return err
		}
	}
	return nil
}

// Save upserts several settings atomically.
func (r *Repository) Save(ctx context.Context, values map[string]string) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	upd := r.store.Rebind("UPDATE site_settings SET value = ?, updated_at = ? WHERE setting_key = ?")
	ins := r.store.Rebind("INSERT INTO site_settings (setting_key, value, updated_at) VALUES (?, ?, ?)")
	for k, v := range values {
		res, err := tx.ExecContext(ctx, upd, v, now, k)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			if _, err := tx.ExecContext(ctx, ins, k, v, now); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// Stats are the public community counters.
type Stats struct {
	Users      int64 `json:"users"`
	Topics     int64 `json:"topics"`
	Replies    int64 `json:"replies"`
	Categories int64 `json:"categories"`
	Likes      int64 `json:"likes"`
}

// Stats counts users, published topics, replies, categories and likes in one round trip.
func (r *Repository) Stats(ctx context.Context) (*Stats, error) {
	q := r.store.Rebind(`SELECT
        (SELECT COUNT(*) FROM users WHERE status = ?),
        (SELECT COUNT(*) FROM topics WHERE status = ?),
        (SELECT COUNT(*) FROM replies WHERE is_deleted = ?),
        (SELECT COUNT(*) FROM categories WHERE is_active = ?),
        (SELECT COUNT(*) FROM likes)`)
	var s Stats
	err := r.store.DB.QueryRowContext(ctx, q, "active", "published", false, true).
		Scan(&s.Users, &s.Topics, &s.Replies, &s.Categories, &s.Likes)
	return &s, err
}

// FeaturedUser is a highlighted community member.
type FeaturedUser struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Reputation  int     `json:"reputation"`
	Followers   int64   `json:"followers"`
	Topics      int64   `json:"topics"`
}

// FeaturedUsers ranks active users by reputation, followers and topics.
func (r *Repository) FeaturedUsers(ctx context.Context, limit int) ([]FeaturedUser, error) {
	q := r.store.Rebind(`SELECT u.id, u.username, u.display_name, u.avatar_url, u.reputation,
            (SELECT COUNT(*) FROM follows f WHERE f.target_type = ? AND f.target_id = u.id) AS followers,
            (SELECT COUNT(*) FROM topics t WHERE t.author_id = u.id AND t.status = ?) AS topic_count
        FROM users u
        WHERE u.status = ?
        ORDER BY u.reputation DESC, followers DESC, topic_count DESC, u.created_at ASC
        LIMIT ?`)
	rows, err := r.store.DB.QueryContext(ctx, q, "user", "published", "active", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []FeaturedUser{}
	for rows.Next() {
		var (
			u      FeaturedUser
			avatar sql.NullString
		)
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &avatar, &u.Reputation, &u.Followers, &u.Topics); err != nil {
			return nil, err
		}
		if avatar.Valid {
			u.AvatarURL = &avatar.String
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// Activity is one entry of the "latest activity" feed.
type Activity struct {
	Kind       string    `json:"kind"` // "topic" | "reply"
	ID         string    `json:"id"`
	TopicID    string    `json:"topicId"`
	TopicTitle string    `json:"topicTitle"`
	Actor      ActorRef  `json:"actor"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ActorRef is the author of an activity entry.
type ActorRef struct {
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}

// RecentActivity merges the latest public topics and replies, newest first.
func (r *Repository) RecentActivity(ctx context.Context, limit int) ([]Activity, error) {
	topicsQ := r.store.Rebind(`SELECT 'topic', t.id, t.id, t.title, u.username, u.display_name, u.avatar_url, t.created_at
        FROM topics t JOIN users u ON u.id = t.author_id JOIN categories c ON c.id = t.category_id
        WHERE t.status = ? AND c.is_active = ?
        ORDER BY t.created_at DESC LIMIT ?`)
	repliesQ := r.store.Rebind(`SELECT 'reply', r.id, t.id, t.title, u.username, u.display_name, u.avatar_url, r.created_at
        FROM replies r JOIN topics t ON t.id = r.topic_id JOIN users u ON u.id = r.author_id
        JOIN categories c ON c.id = t.category_id
        WHERE r.is_deleted = ? AND t.status = ? AND c.is_active = ?
        ORDER BY r.created_at DESC LIMIT ?`)

	topics, err := r.activity(ctx, topicsQ, "published", true, limit)
	if err != nil {
		return nil, err
	}
	replies, err := r.activity(ctx, repliesQ, false, "published", true, limit)
	if err != nil {
		return nil, err
	}
	return mergeActivity(topics, replies, limit), nil
}

func (r *Repository) activity(ctx context.Context, query string, args ...any) ([]Activity, error) {
	rows, err := r.store.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Activity
	for rows.Next() {
		var (
			a      Activity
			avatar sql.NullString
		)
		if err := rows.Scan(&a.Kind, &a.ID, &a.TopicID, &a.TopicTitle,
			&a.Actor.Username, &a.Actor.DisplayName, &avatar, &a.CreatedAt); err != nil {
			return nil, err
		}
		if avatar.Valid {
			a.Actor.AvatarURL = &avatar.String
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// mergeActivity merges two newest-first lists into one, capped at limit.
func mergeActivity(a, b []Activity, limit int) []Activity {
	out := make([]Activity, 0, min(len(a)+len(b), limit))
	i, j := 0, 0
	for len(out) < limit && (i < len(a) || j < len(b)) {
		if j >= len(b) || (i < len(a) && !a[i].CreatedAt.Before(b[j].CreatedAt)) {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	return out
}
