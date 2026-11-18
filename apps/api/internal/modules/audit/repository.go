package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/id"
)

// Repository persists audit log entries.
type Repository struct {
	store *database.Store
}

// NewRepository builds an audit Repository.
func NewRepository(store *database.Store) *Repository { return &Repository{store: store} }

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Insert writes one entry.
func (r *Repository) Insert(ctx context.Context, actorID, action, targetType, targetID string, details map[string]any, ip string) error {
	var payload any
	if len(details) > 0 {
		b, err := json.Marshal(details)
		if err != nil {
			return err
		}
		payload = string(b)
	}
	q := r.store.Rebind(`INSERT INTO audit_logs (id, actor_id, action, target_type, target_id, details, ip, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q, id.New(), nullable(actorID), action,
		nullable(targetType), nullable(targetID), payload, nullable(ip), time.Now().UTC())
	return err
}

func (f Filter) where() (string, []any) {
	var (
		conds []string
		args  []any
	)
	if f.Action != "" {
		conds = append(conds, "a.action = ?")
		args = append(args, f.Action)
	}
	if f.ActorID != "" {
		conds = append(conds, "a.actor_id = ?")
		args = append(args, f.ActorID)
	}
	if f.Target != "" {
		conds = append(conds, "a.target_id = ?")
		args = append(args, f.Target)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		p := database.ContainsPattern(q)
		conds = append(conds, "(LOWER(u.username) LIKE ? OR LOWER(a.action) LIKE ?)")
		args = append(args, p, p)
	}
	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

// List returns a page of entries (newest first) and the total count.
func (r *Repository) List(ctx context.Context, f Filter) ([]Entry, int64, error) {
	where, args := f.where()
	dir := "DESC"
	if f.Asc {
		dir = "ASC"
	}

	var total int64
	cq := r.store.Rebind("SELECT COUNT(*) FROM audit_logs a LEFT JOIN users u ON u.id = a.actor_id" + where)
	if err := r.store.DB.QueryRowContext(ctx, cq, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := r.store.Rebind(`SELECT a.id, a.action, a.target_type, a.target_id, a.details, a.ip, a.created_at,
        u.id, u.username, u.display_name
        FROM audit_logs a LEFT JOIN users u ON u.id = a.actor_id` + where +
		" ORDER BY a.created_at " + dir + " LIMIT ? OFFSET ?")
	rows, err := r.store.DB.QueryContext(ctx, q, append(args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Entry{}
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func scanEntry(rows *sql.Rows) (Entry, error) {
	var (
		e                       Entry
		tType, tID, details, ip sql.NullString
		uid, uname, udisplay    sql.NullString
	)
	if err := rows.Scan(&e.ID, &e.Action, &tType, &tID, &details, &ip, &e.CreatedAt, &uid, &uname, &udisplay); err != nil {
		return e, err
	}
	if tType.Valid {
		e.TargetType = &tType.String
	}
	if tID.Valid {
		e.TargetID = &tID.String
	}
	if ip.Valid {
		e.IP = &ip.String
	}
	if details.Valid && details.String != "" {
		_ = json.Unmarshal([]byte(details.String), &e.Details)
	}
	if uid.Valid {
		e.Actor = &Actor{ID: uid.String, Username: uname.String, DisplayName: udisplay.String}
	}
	return e, nil
}

// Actions returns the distinct recorded action names (for filters).
func (r *Repository) Actions(ctx context.Context) ([]string, error) {
	rows, err := r.store.DB.QueryContext(ctx, "SELECT DISTINCT action FROM audit_logs ORDER BY action")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
