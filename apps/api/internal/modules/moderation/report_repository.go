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

// The target is resolved in the same query (no N+1): one LEFT JOIN branch per
// target type, only one of which matches.
const reportSelect = `SELECT r.id, r.target_type, r.target_id, r.reason, r.details, r.status,
    r.resolution_note, r.created_at, r.updated_at,
    rep.id, rep.username, rep.display_name,
    h.id, h.username, h.display_name,
    tu.username, tu.display_name,
    tt.title, tt.content, tta.username,
    rr.content, rr.topic_id, rr.is_deleted, rrt.title, rra.username
    FROM reports r
    LEFT JOIN users rep ON rep.id = r.reporter_id
    LEFT JOIN users h ON h.id = r.handled_by
    LEFT JOIN users tu ON r.target_type = 'user' AND tu.id = r.target_id
    LEFT JOIN topics tt ON r.target_type = 'topic' AND tt.id = r.target_id
    LEFT JOIN users tta ON tta.id = tt.author_id
    LEFT JOIN replies rr ON r.target_type = 'reply' AND rr.id = r.target_id
    LEFT JOIN topics rrt ON rrt.id = rr.topic_id
    LEFT JOIN users rra ON rra.id = rr.author_id`

const excerptLen = 280

func excerpt(s string) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= excerptLen {
		return string(runes)
	}
	return string(runes[:excerptLen]) + "…"
}

func scanReport(row scanner) (*Report, error) {
	var (
		rp                                    Report
		details, note                         sql.NullString
		reporter, handler                     nullRef
		tuUser, tuDisplay                     sql.NullString
		ttTitle, ttContent, ttAuthor          sql.NullString
		rrContent, rrTopic, rrTitle, rrAuthor sql.NullString
		rrDeleted                             sql.NullBool
	)
	err := row.Scan(&rp.ID, &rp.Target.Type, &rp.Target.ID, &rp.Reason, &details, &rp.Status,
		&note, &rp.CreatedAt, &rp.UpdatedAt,
		&reporter.id, &reporter.username, &reporter.display,
		&handler.id, &handler.username, &handler.display,
		&tuUser, &tuDisplay,
		&ttTitle, &ttContent, &ttAuthor,
		&rrContent, &rrTopic, &rrDeleted, &rrTitle, &rrAuthor)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rp.Details = strPtr(details)
	rp.ResolutionNote = strPtr(note)
	rp.Reporter = reporter.ref()
	rp.HandledBy = handler.ref()

	t := &rp.Target
	switch t.Type {
	case TargetUser:
		if tuUser.Valid {
			t.Exists = true
			t.Title = tuDisplay.String + " (@" + tuUser.String + ")"
			t.Link = "/u/" + tuUser.String
			t.Author = &tuUser.String
		}
	case TargetTopic:
		if ttTitle.Valid {
			t.Exists = true
			t.Title = ttTitle.String
			t.Excerpt = excerpt(ttContent.String)
			t.Link = "/t/" + t.ID
			t.Author = strPtr(ttAuthor)
		}
	case TargetReply:
		if rrTopic.Valid {
			t.Exists = !rrDeleted.Bool
			t.Title = "Resposta em “" + rrTitle.String + "”"
			if t.Exists {
				t.Excerpt = excerpt(rrContent.String)
			}
			t.Link = "/t/" + rrTopic.String + "#reply-" + t.ID
			t.Author = strPtr(rrAuthor)
		}
	}
	return &rp, nil
}

// TargetExists reports whether the reported entity exists.
func (r *Repository) TargetExists(ctx context.Context, targetType, targetID string) (bool, error) {
	switch targetType {
	case TargetUser:
		return r.exists(ctx, "SELECT 1 FROM users WHERE id = ?", targetID)
	case TargetTopic:
		return r.exists(ctx, "SELECT 1 FROM topics WHERE id = ?", targetID)
	case TargetReply:
		return r.exists(ctx, "SELECT 1 FROM replies WHERE id = ? AND is_deleted = FALSE", targetID)
	}
	return false, nil
}

// HasOpenReport reports whether the reporter already has an open report on a target.
func (r *Repository) HasOpenReport(ctx context.Context, reporterID, targetType, targetID string) (bool, error) {
	q := r.store.Rebind(`SELECT 1 FROM reports WHERE reporter_id = ? AND target_type = ? AND target_id = ?
        AND status IN ('pending', 'reviewing') LIMIT 1`)
	var one int
	err := r.store.DB.QueryRowContext(ctx, q, reporterID, targetType, targetID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// CreateReport inserts a report.
func (r *Repository) CreateReport(ctx context.Context, id, reporterID string, in CreateReportInput) error {
	now := time.Now().UTC()
	var details any
	if d := strings.TrimSpace(in.Details); d != "" {
		details = d
	}
	q := r.store.Rebind(`INSERT INTO reports (id, reporter_id, target_type, target_id, reason, details, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := r.store.DB.ExecContext(ctx, q, id, reporterID, in.TargetType, in.TargetID, in.Reason, details, StatusPending, now, now)
	return err
}

// GetReport fetches a report by id.
func (r *Repository) GetReport(ctx context.Context, id string) (*Report, error) {
	q := r.store.Rebind(reportSelect + " WHERE r.id = ?")
	return scanReport(r.store.DB.QueryRowContext(ctx, q, id))
}

// UpdateReportStatus changes the status, handler and note of a report.
func (r *Repository) UpdateReportStatus(ctx context.Context, id, status, handledBy, note string) error {
	var noteVal any
	if note != "" {
		noteVal = note
	}
	// An empty note keeps the existing one (moving a report between statuses
	// must not erase what a moderator already wrote).
	if noteVal == nil {
		q := r.store.Rebind("UPDATE reports SET status = ?, handled_by = ?, updated_at = ? WHERE id = ?")
		_, err := r.store.DB.ExecContext(ctx, q, status, handledBy, time.Now().UTC(), id)
		return err
	}
	q := r.store.Rebind("UPDATE reports SET status = ?, handled_by = ?, resolution_note = ?, updated_at = ? WHERE id = ?")
	_, err := r.store.DB.ExecContext(ctx, q, status, handledBy, noteVal, time.Now().UTC(), id)
	return err
}

// ListReports returns a filtered page of reports (oldest first for open
// statuses, newest first otherwise) and the total.
func (r *Repository) ListReports(ctx context.Context, f ReportFilter) ([]Report, int64, error) {
	var (
		conds []string
		args  []any
	)
	if f.Status != "" {
		conds = append(conds, "r.status = ?")
		args = append(args, f.Status)
	}
	if f.TargetType != "" {
		conds = append(conds, "r.target_type = ?")
		args = append(args, f.TargetType)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		p := database.ContainsPattern(q)
		conds = append(conds, "(LOWER(r.reason) LIKE ? OR LOWER(COALESCE(r.details, '')) LIKE ? OR LOWER(COALESCE(rep.username, '')) LIKE ?)")
		args = append(args, p, p, p)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	var total int64
	countQ := "SELECT COUNT(*) FROM reports r LEFT JOIN users rep ON rep.id = r.reporter_id" + where
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(countQ), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	order := " ORDER BY r.created_at DESC"
	if f.Order == "asc" || (f.Order == "" && (f.Status == StatusPending || f.Status == StatusReviewing)) {
		order = " ORDER BY r.created_at ASC"
	}
	q := r.store.Rebind(reportSelect + where + order + " LIMIT ? OFFSET ?")
	rows, err := r.store.DB.QueryContext(ctx, q, append(args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Report{}
	for rows.Next() {
		rp, err := scanReport(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *rp)
	}
	return out, total, rows.Err()
}

// ReportCounts returns the number of reports per status.
func (r *Repository) ReportCounts(ctx context.Context) (map[string]int64, error) {
	rows, err := r.store.DB.QueryContext(ctx, "SELECT status, COUNT(*) FROM reports GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{StatusPending: 0, StatusReviewing: 0, StatusResolved: 0, StatusRejected: 0}
	for rows.Next() {
		var (
			s string
			n int64
		)
		if err := rows.Scan(&s, &n); err != nil {
			return nil, err
		}
		out[s] = n
	}
	return out, rows.Err()
}
