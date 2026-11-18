package moderation

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// Repository persists sanctions, warnings and reports.
type Repository struct {
	store *database.Store
}

// NewRepository builds a moderation Repository.
func NewRepository(store *database.Store) *Repository { return &Repository{store: store} }

type scanner interface{ Scan(...any) error }

// nullRef scans an optional joined user.
type nullRef struct{ id, username, display sql.NullString }

func (n nullRef) ref() *UserRef {
	if !n.id.Valid {
		return nil
	}
	return &UserRef{ID: n.id.String, Username: n.username.String, DisplayName: n.display.String}
}

func timePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func strPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	v := s.String
	return &v
}

// ---- Users ----

// UserByUsername resolves a user reference by username.
func (r *Repository) UserByUsername(ctx context.Context, username string) (*UserRef, error) {
	q := r.store.Rebind("SELECT id, username, display_name, avatar_url FROM users WHERE username = ?")
	var (
		u      UserRef
		avatar sql.NullString
	)
	err := r.store.DB.QueryRowContext(ctx, q, username).Scan(&u.ID, &u.Username, &u.DisplayName, &avatar)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.AvatarURL = strPtr(avatar)
	return &u, nil
}

// UserExists reports whether a user id exists.
func (r *Repository) UserExists(ctx context.Context, id string) (bool, error) {
	return r.exists(ctx, "SELECT 1 FROM users WHERE id = ?", id)
}

func (r *Repository) exists(ctx context.Context, q, arg string) (bool, error) {
	var one int
	err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(q), arg).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// ---- Sanctions ----

const sanctionSelect = `SELECT s.id, s.type, s.reason, s.expires_at, s.revoked_at, s.revoke_reason, s.created_at,
    u.id, u.username, u.display_name,
    i.id, i.username, i.display_name,
    v.id, v.username, v.display_name
    FROM sanctions s
    JOIN users u ON u.id = s.user_id
    LEFT JOIN users i ON i.id = s.issued_by
    LEFT JOIN users v ON v.id = s.revoked_by`

func scanSanction(row scanner, now time.Time) (*Sanction, error) {
	var (
		s                Sanction
		expires, revoked sql.NullTime
		revokeReason     sql.NullString
		issuer, revoker  nullRef
	)
	err := row.Scan(&s.ID, &s.Type, &s.Reason, &expires, &revoked, &revokeReason, &s.CreatedAt,
		&s.User.ID, &s.User.Username, &s.User.DisplayName,
		&issuer.id, &issuer.username, &issuer.display,
		&revoker.id, &revoker.username, &revoker.display)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.ExpiresAt = timePtr(expires)
	s.RevokedAt = timePtr(revoked)
	s.RevokeReason = strPtr(revokeReason)
	s.IssuedBy = issuer.ref()
	s.RevokedBy = revoker.ref()
	s.Active = s.ActiveAt(now)
	return &s, nil
}

// CreateSanction inserts a ban or mute.
func (r *Repository) CreateSanction(ctx context.Context, id, userID, typ, reason, issuedBy string, expiresAt *time.Time) error {
	q := r.store.Rebind(`INSERT INTO sanctions (id, user_id, type, reason, issued_by, expires_at, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q, id, userID, typ, reason, issuedBy, expiresAt, time.Now().UTC())
	return err
}

// GetSanction fetches a sanction by id.
func (r *Repository) GetSanction(ctx context.Context, id string) (*Sanction, error) {
	q := r.store.Rebind(sanctionSelect + " WHERE s.id = ?")
	return scanSanction(r.store.DB.QueryRowContext(ctx, q, id), time.Now().UTC())
}

// ActiveSanction returns the active sanction of a type for a user, or nil.
func (r *Repository) ActiveSanction(ctx context.Context, userID, typ string) (*Sanction, error) {
	now := time.Now().UTC()
	q := r.store.Rebind(sanctionSelect + ` WHERE s.user_id = ? AND s.type = ? AND s.revoked_at IS NULL
        AND (s.expires_at IS NULL OR s.expires_at > ?) ORDER BY s.created_at DESC LIMIT 1`)
	s, err := scanSanction(r.store.DB.QueryRowContext(ctx, q, userID, typ, now), now)
	if errors.Is(err, apperr.ErrNotFound) {
		return nil, nil
	}
	return s, err
}

// RevokeActive lifts every active sanction of a type for a user.
func (r *Repository) RevokeActive(ctx context.Context, userID, typ, revokedBy, reason string) (int64, error) {
	now := time.Now().UTC()
	q := r.store.Rebind(`UPDATE sanctions SET revoked_at = ?, revoked_by = ?, revoke_reason = ?
        WHERE user_id = ? AND type = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > ?)`)
	res, err := r.store.DB.ExecContext(ctx, q, now, revokedBy, reason, userID, typ, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// RevokeSanction lifts one sanction if still active.
func (r *Repository) RevokeSanction(ctx context.Context, id, revokedBy, reason string) error {
	now := time.Now().UTC()
	q := r.store.Rebind(`UPDATE sanctions SET revoked_at = ?, revoked_by = ?, revoke_reason = ?
        WHERE id = ? AND revoked_at IS NULL`)
	res, err := r.store.DB.ExecContext(ctx, q, now, revokedBy, reason, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &ruleError{msg: "a sanção já foi revogada"}
	}
	return nil
}

// ListSanctions returns a filtered page of sanctions, newest first.
func (r *Repository) ListSanctions(ctx context.Context, f SanctionFilter) ([]Sanction, int64, error) {
	now := time.Now().UTC()
	var (
		conds []string
		args  []any
	)
	if f.Type != "" {
		conds = append(conds, "s.type = ?")
		args = append(args, f.Type)
	}
	if f.UserID != "" {
		conds = append(conds, "s.user_id = ?")
		args = append(args, f.UserID)
	}
	if f.ActiveOnly {
		conds = append(conds, "s.revoked_at IS NULL AND (s.expires_at IS NULL OR s.expires_at > ?)")
		args = append(args, now)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	var total int64
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind("SELECT COUNT(*) FROM sanctions s"+where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := r.store.Rebind(sanctionSelect + where + " ORDER BY s.created_at DESC LIMIT ? OFFSET ?")
	rows, err := r.store.DB.QueryContext(ctx, q, append(args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Sanction{}
	for rows.Next() {
		s, err := scanSanction(rows, now)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *s)
	}
	return out, total, rows.Err()
}

// ---- Warnings ----

// CreateWarning inserts a warning.
func (r *Repository) CreateWarning(ctx context.Context, id, userID, reason, issuedBy string) error {
	q := r.store.Rebind("INSERT INTO warnings (id, user_id, reason, issued_by, created_at) VALUES (?, ?, ?, ?, ?)")
	_, err := r.store.DB.ExecContext(ctx, q, id, userID, reason, issuedBy, time.Now().UTC())
	return err
}

// WarningsOf lists the warnings of a user, newest first.
func (r *Repository) WarningsOf(ctx context.Context, userID string, limit int) ([]Warning, error) {
	q := r.store.Rebind(`SELECT w.id, w.reason, w.created_at, u.id, u.username, u.display_name,
        i.id, i.username, i.display_name
        FROM warnings w JOIN users u ON u.id = w.user_id LEFT JOIN users i ON i.id = w.issued_by
        WHERE w.user_id = ? ORDER BY w.created_at DESC LIMIT ?`)
	rows, err := r.store.DB.QueryContext(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Warning{}
	for rows.Next() {
		var (
			w      Warning
			issuer nullRef
		)
		if err := rows.Scan(&w.ID, &w.Reason, &w.CreatedAt, &w.User.ID, &w.User.Username, &w.User.DisplayName,
			&issuer.id, &issuer.username, &issuer.display); err != nil {
			return nil, err
		}
		w.IssuedBy = issuer.ref()
		out = append(out, w)
	}
	return out, rows.Err()
}
