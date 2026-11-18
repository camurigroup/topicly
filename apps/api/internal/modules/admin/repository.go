package admin

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/database"
)

// Repository runs the dashboard aggregates and admin listings.
type Repository struct {
	store *database.Store
}

// NewRepository builds an admin Repository.
func NewRepository(store *database.Store) *Repository { return &Repository{store: store} }

const activeSanction = "s.revoked_at IS NULL AND (s.expires_at IS NULL OR s.expires_at > ?)"

func (r *Repository) count(ctx context.Context, q string, args ...any) (int64, error) {
	var n int64
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(q), args...).Scan(&n)
	return n, err
}

// Totals computes every dashboard counter.
func (r *Repository) Totals(ctx context.Context, now time.Time) (*Totals, error) {
	t := &Totals{}
	queries := []struct {
		dst  *int64
		q    string
		args []any
	}{
		{&t.Users, "SELECT COUNT(*) FROM users", nil},
		{&t.ActiveUsers, "SELECT COUNT(*) FROM users WHERE last_active_at >= ?", []any{now.AddDate(0, 0, -7)}},
		{&t.NewUsers, "SELECT COUNT(*) FROM users WHERE created_at >= ?", []any{now.AddDate(0, 0, -30)}},
		{&t.Categories, "SELECT COUNT(*) FROM categories WHERE parent_id IS NULL", nil},
		{&t.Subcategories, "SELECT COUNT(*) FROM categories WHERE parent_id IS NOT NULL", nil},
		{&t.Topics, "SELECT COUNT(*) FROM topics", nil},
		{&t.Replies, "SELECT COUNT(*) FROM replies WHERE is_deleted = ?", []any{false}},
		{&t.Likes, "SELECT COUNT(*) FROM likes", nil},
		{&t.ReportsPending, "SELECT COUNT(*) FROM reports WHERE status = 'pending'", nil},
		{&t.Reports, "SELECT COUNT(*) FROM reports", nil},
		{&t.Banned, "SELECT COUNT(DISTINCT s.user_id) FROM sanctions s WHERE s.type = 'ban' AND " + activeSanction, []any{now}},
		{&t.Muted, "SELECT COUNT(DISTINCT s.user_id) FROM sanctions s WHERE s.type = 'mute' AND " + activeSanction, []any{now}},
		{&t.Warnings, "SELECT COUNT(*) FROM warnings", nil},
	}
	for _, q := range queries {
		n, err := r.count(ctx, q.q, q.args...)
		if err != nil {
			return nil, err
		}
		*q.dst = n
	}
	return t, nil
}

// DailyCounts returns "YYYY-MM-DD" (UTC) → rows created that day since `since`.
func (r *Repository) DailyCounts(ctx context.Context, metric string, since time.Time) (map[string]int64, error) {
	table, ok := seriesTables[metric]
	if !ok {
		return nil, seriesError("métrica inválida: " + metric)
	}
	day := "DATE(created_at)"
	if r.store.Dialect == "postgres" {
		day = "CAST(created_at AT TIME ZONE 'UTC' AS DATE)"
	}
	q := r.store.Rebind("SELECT " + day + ", COUNT(*) FROM " + table + " WHERE created_at >= ? GROUP BY " + day)
	rows, err := r.store.DB.QueryContext(ctx, q, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var (
			d time.Time
			n int64
		)
		if err := rows.Scan(&d, &n); err != nil {
			return nil, err
		}
		out[d.Format(dateLayout)] = n
	}
	return out, rows.Err()
}

type where struct {
	conds []string
	args  []any
}

func (w *where) add(cond string, args ...any) {
	w.conds = append(w.conds, cond)
	w.args = append(w.args, args...)
}

func (w *where) sql() string {
	if len(w.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.conds, " AND ")
}

func boolInt(n int) bool { return n != 0 }

// ListUsers returns a filtered, sorted page of users with roles and sanction flags.
func (r *Repository) ListUsers(ctx context.Context, f UserFilter, now time.Time) ([]UserRow, int64, error) {
	order, err := OrderBy(map[string]string{
		"created": "u.created_at", "active": "COALESCE(u.last_active_at, u.created_at)",
		"username": "u.username", "reputation": "u.reputation", "topics": "topic_count",
	}, "created", f.Sort, f.Order)
	if err != nil {
		return nil, 0, err
	}
	banned := "EXISTS (SELECT 1 FROM sanctions s WHERE s.user_id = u.id AND s.type = 'ban' AND " + activeSanction + ")"
	muted := "EXISTS (SELECT 1 FROM sanctions s WHERE s.user_id = u.id AND s.type = 'mute' AND " + activeSanction + ")"

	var w where
	if q := strings.TrimSpace(f.Q); q != "" {
		p := database.ContainsPattern(q)
		w.add("(LOWER(u.username) LIKE ? OR LOWER(u.display_name) LIKE ? OR LOWER(u.email) LIKE ?)", p, p, p)
	}
	switch f.Role {
	case "":
	case "member":
		w.add("NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)")
	default:
		w.add("EXISTS (SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE ur.user_id = u.id AND ro.slug = ?)", f.Role)
	}
	switch f.Status {
	case "banned":
		w.add(banned, now)
	case "muted":
		w.add(muted, now)
	case "active":
		w.add("NOT "+banned, now)
	}

	total, err := r.count(ctx, "SELECT COUNT(*) FROM users u"+w.sql(), w.args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT u.id, u.username, u.display_name, u.email, u.avatar_url, u.reputation, u.created_at, u.last_active_at,
        (SELECT COUNT(*) FROM topics t WHERE t.author_id = u.id) AS topic_count,
        (SELECT COUNT(*) FROM replies rp WHERE rp.author_id = u.id AND rp.is_deleted = ?) AS reply_count,
        CASE WHEN ` + banned + ` THEN 1 ELSE 0 END,
        CASE WHEN ` + muted + ` THEN 1 ELSE 0 END
        FROM users u` + w.sql() + " ORDER BY " + order + ", u.id LIMIT ? OFFSET ?"
	args := append([]any{false, now, now}, w.args...)
	args = append(args, f.Limit, f.Offset)
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []UserRow{}
	for rows.Next() {
		var (
			u              UserRow
			avatar         sql.NullString
			last           sql.NullTime
			isBan, isMuted int
		)
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email, &avatar, &u.Reputation, &u.CreatedAt, &last,
			&u.Topics, &u.Replies, &isBan, &isMuted); err != nil {
			return nil, 0, err
		}
		if avatar.Valid {
			u.AvatarURL = &avatar.String
		}
		if last.Valid {
			t := last.Time
			u.LastActiveAt = &t
		}
		u.IsBanned, u.IsMuted = boolInt(isBan), boolInt(isMuted)
		u.Roles = []string{}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, r.attachRoles(ctx, out)
}

// attachRoles loads the role slugs of a page of users in one query.
func (r *Repository) attachRoles(ctx context.Context, users []UserRow) error {
	if len(users) == 0 {
		return nil
	}
	idx := make(map[string]int, len(users))
	args := make([]any, len(users))
	for i, u := range users {
		idx[u.ID] = i
		args[i] = u.ID
	}
	ph := strings.TrimSuffix(strings.Repeat("?, ", len(users)), ", ")
	q := r.store.Rebind(`SELECT ur.user_id, ro.slug FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id
        WHERE ur.user_id IN (` + ph + `) ORDER BY ro.rank_order DESC`)
	rows, err := r.store.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var uid, slug string
		if err := rows.Scan(&uid, &slug); err != nil {
			return err
		}
		users[idx[uid]].Roles = append(users[idx[uid]].Roles, slug)
	}
	return rows.Err()
}

// ListTopics returns a filtered, sorted page of topics.
func (r *Repository) ListTopics(ctx context.Context, f TopicFilter) ([]TopicRow, int64, error) {
	order, err := OrderBy(map[string]string{
		"created": "t.created_at", "activity": "t.last_activity_at", "replies": "t.reply_count",
		"likes": "t.like_count", "views": "t.view_count", "title": "t.title",
	}, "created", f.Sort, f.Order)
	if err != nil {
		return nil, 0, err
	}
	var w where
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add("LOWER(t.title) LIKE ?", database.ContainsPattern(q))
	}
	if f.Category != "" {
		w.add("c.slug = ?", f.Category)
	}
	switch f.Flag {
	case "pinned":
		w.add("t.is_pinned = ?", true)
	case "locked":
		w.add("t.is_locked = ?", true)
	case "featured":
		w.add("t.is_featured = ?", true)
	}
	from := " FROM topics t JOIN categories c ON c.id = t.category_id JOIN users u ON u.id = t.author_id"
	total, err := r.count(ctx, "SELECT COUNT(*)"+from+w.sql(), w.args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT t.id, t.title, c.slug, c.name, u.username, t.is_pinned, t.is_locked, t.is_featured,
        t.view_count, t.reply_count, t.like_count, t.created_at, t.last_activity_at` + from + w.sql() +
		" ORDER BY " + order + ", t.id LIMIT ? OFFSET ?"
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), append(w.args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []TopicRow{}
	for rows.Next() {
		var t TopicRow
		if err := rows.Scan(&t.ID, &t.Title, &t.CategorySlug, &t.CategoryName, &t.Author, &t.IsPinned, &t.IsLocked,
			&t.IsFeatured, &t.Views, &t.Replies, &t.Likes, &t.CreatedAt, &t.LastActivity); err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

const replyExcerpt = 200

// ListReplies returns a filtered, sorted page of replies.
func (r *Repository) ListReplies(ctx context.Context, f ReplyFilter) ([]ReplyRow, int64, error) {
	order, err := OrderBy(map[string]string{"created": "rp.created_at", "likes": "rp.like_count"}, "created", f.Sort, f.Order)
	if err != nil {
		return nil, 0, err
	}
	var w where
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add("LOWER(rp.content) LIKE ?", database.ContainsPattern(q))
	}
	switch f.State {
	case "visible":
		w.add("rp.is_deleted = ?", false)
	case "deleted":
		w.add("rp.is_deleted = ?", true)
	}
	from := " FROM replies rp JOIN topics t ON t.id = rp.topic_id JOIN users u ON u.id = rp.author_id"
	total, err := r.count(ctx, "SELECT COUNT(*)"+from+w.sql(), w.args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT rp.id, rp.content, rp.topic_id, t.title, u.username, rp.like_count, rp.is_deleted, rp.created_at` +
		from + w.sql() + " ORDER BY " + order + ", rp.id LIMIT ? OFFSET ?"
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), append(w.args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []ReplyRow{}
	for rows.Next() {
		var (
			rp      ReplyRow
			content string
		)
		if err := rows.Scan(&rp.ID, &content, &rp.TopicID, &rp.TopicTitle, &rp.Author, &rp.Likes, &rp.IsDeleted, &rp.CreatedAt); err != nil {
			return nil, 0, err
		}
		if runes := []rune(content); len(runes) > replyExcerpt {
			content = string(runes[:replyExcerpt]) + "…"
		}
		rp.Excerpt = content
		out = append(out, rp)
	}
	return out, total, rows.Err()
}

// ListWarnings returns a filtered page of warnings.
func (r *Repository) ListWarnings(ctx context.Context, f WarningFilter) ([]WarningRow, int64, error) {
	order, _ := OrderBy(map[string]string{"created": "w.created_at"}, "created", "", f.Order)
	var w where
	if q := strings.TrimSpace(f.Q); q != "" {
		p := database.ContainsPattern(q)
		w.add("(LOWER(u.username) LIKE ? OR LOWER(w.reason) LIKE ?)", p, p)
	}
	from := " FROM warnings w JOIN users u ON u.id = w.user_id LEFT JOIN users i ON i.id = w.issued_by"
	total, err := r.count(ctx, "SELECT COUNT(*)"+from+w.sql(), w.args...)
	if err != nil {
		return nil, 0, err
	}
	q := "SELECT w.id, u.username, u.display_name, w.reason, i.username, w.created_at" + from + w.sql() +
		" ORDER BY " + order + ", w.id LIMIT ? OFFSET ?"
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), append(w.args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []WarningRow{}
	for rows.Next() {
		var (
			row    WarningRow
			issuer sql.NullString
		)
		if err := rows.Scan(&row.ID, &row.Username, &row.Display, &row.Reason, &issuer, &row.CreatedAt); err != nil {
			return nil, 0, err
		}
		if issuer.Valid {
			row.IssuedBy = &issuer.String
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

// MediaCounts returns how many users have an avatar and a banner.
func (r *Repository) MediaCounts(ctx context.Context) (avatars, banners int64, err error) {
	if avatars, err = r.count(ctx, "SELECT COUNT(*) FROM users WHERE avatar_url IS NOT NULL AND avatar_url <> ''"); err != nil {
		return 0, 0, err
	}
	banners, err = r.count(ctx, "SELECT COUNT(*) FROM users WHERE banner_url IS NOT NULL AND banner_url <> ''")
	return avatars, banners, err
}
