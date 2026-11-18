package topic

import (
	"context"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/database"
)

// Sort orders accepted by searches and discovery.
const (
	SortRelevance = "relevance"
	SortRecent    = "recent"
	SortPopular   = "popular"
	SortReplies   = "replies"
	SortLikes     = "likes"
	SortActivity  = "activity"
)

// popularityScore weighs engagement for "popular" orderings.
const popularityScore = "(t.like_count * 3 + t.reply_count * 2 + t.view_count / 10)"

// SearchFilter narrows topic and reply searches. Zero values mean "no filter".
type SearchFilter struct {
	Terms      []string // lowercased words; each must match
	CategoryID string   // matches the category and its direct subcategories
	AuthorID   string
	From, To   *time.Time // created_at in [From, To)
	MinReplies int
	MinLikes   int
	Sort       string
	Limit      int
	Offset     int
}

// ReplyHit is a reply search result with its topic context.
type ReplyHit struct {
	Reply
	TopicTitle   string `json:"topicTitle"`
	CategorySlug string `json:"categorySlug"`
}

type where struct {
	clauses []string
	args    []any
}

func (w *where) add(clause string, args ...any) {
	w.clauses = append(w.clauses, clause)
	w.args = append(w.args, args...)
}

func (w *where) sql() string {
	if len(w.clauses) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.clauses, " AND ")
}

// publicTopics restricts to published topics in active categories.
func publicTopics(w *where) {
	w.add("t.status = ?", "published")
	w.add("c.is_active = ?", true)
}

func commonFilters(w *where, f SearchFilter, createdCol string) {
	if f.CategoryID != "" {
		w.add("(t.category_id = ? OR c.parent_id = ?)", f.CategoryID, f.CategoryID)
	}
	if f.From != nil {
		w.add(createdCol+" >= ?", *f.From)
	}
	if f.To != nil {
		w.add(createdCol+" < ?", *f.To)
	}
}

// SearchTopics returns matching topics and the total count.
func (r *Repository) SearchTopics(ctx context.Context, f SearchFilter) ([]Topic, int64, error) {
	w := &where{}
	publicTopics(w)
	for _, term := range f.Terms {
		p := database.ContainsPattern(term)
		w.add("(LOWER(t.title) LIKE ? OR LOWER(t.content) LIKE ?)", p, p)
	}
	commonFilters(w, f, "t.created_at")
	if f.AuthorID != "" {
		w.add("t.author_id = ?", f.AuthorID)
	}
	if f.MinReplies > 0 {
		w.add("t.reply_count >= ?", f.MinReplies)
	}
	if f.MinLikes > 0 {
		w.add("t.like_count >= ?", f.MinLikes)
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM topics t JOIN categories c ON c.id = t.category_id" + w.sql()
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(countQ), w.args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []Topic{}, 0, nil
	}

	order, orderArgs := topicOrder(f)
	args := append(append([]any{}, w.args...), orderArgs...)
	args = append(args, f.Limit, f.Offset)
	items, err := r.queryTopics(ctx, topicSelect+w.sql()+" ORDER BY "+order+" LIMIT ? OFFSET ?", args...)
	return items, total, err
}

func topicOrder(f SearchFilter) (string, []any) {
	switch f.Sort {
	case SortPopular:
		return popularityScore + " DESC, t.created_at DESC", nil
	case SortReplies:
		return "t.reply_count DESC, t.created_at DESC", nil
	case SortLikes:
		return "t.like_count DESC, t.created_at DESC", nil
	case SortActivity:
		return "t.last_activity_at DESC", nil
	case SortRelevance:
		if len(f.Terms) > 0 {
			// Title hits on the first term rank above content-only hits.
			return "CASE WHEN LOWER(t.title) LIKE ? THEN 0 ELSE 1 END, t.created_at DESC",
				[]any{database.ContainsPattern(f.Terms[0])}
		}
	}
	return "t.created_at DESC", nil
}

func (r *Repository) queryTopics(ctx context.Context, query string, args ...any) ([]Topic, error) {
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

// SearchReplies returns matching non-deleted replies and the total count.
func (r *Repository) SearchReplies(ctx context.Context, f SearchFilter) ([]ReplyHit, int64, error) {
	w := &where{}
	publicTopics(w)
	w.add("r.is_deleted = ?", false)
	for _, term := range f.Terms {
		w.add("LOWER(r.content) LIKE ?", database.ContainsPattern(term))
	}
	commonFilters(w, f, "r.created_at")
	if f.AuthorID != "" {
		w.add("r.author_id = ?", f.AuthorID)
	}
	if f.MinLikes > 0 {
		w.add("r.like_count >= ?", f.MinLikes)
	}

	const from = ` FROM replies r
        JOIN users u ON u.id = r.author_id
        JOIN topics t ON t.id = r.topic_id
        JOIN categories c ON c.id = t.category_id`

	var total int64
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind("SELECT COUNT(*)"+from+w.sql()), w.args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []ReplyHit{}, 0, nil
	}

	order := "r.created_at DESC"
	if f.Sort == SortLikes || f.Sort == SortPopular {
		order = "r.like_count DESC, r.created_at DESC"
	}
	q := `SELECT r.id, r.topic_id, r.parent_id, r.content,
        u.id, u.username, u.display_name, u.avatar_url,
        r.like_count, r.is_deleted, r.created_at, r.edited_at, t.title, c.slug` +
		from + w.sql() + " ORDER BY " + order + " LIMIT ? OFFSET ?"
	args := append(append([]any{}, w.args...), f.Limit, f.Offset)

	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []ReplyHit{}
	for rows.Next() {
		var hit ReplyHit
		rp, err := scanReply(rowWithExtras{rows, []any{&hit.TopicTitle, &hit.CategorySlug}})
		if err != nil {
			return nil, 0, err
		}
		hit.Reply = *rp
		out = append(out, hit)
	}
	return out, total, rows.Err()
}

// rowWithExtras lets scanReply read extra trailing columns.
type rowWithExtras struct {
	row    scanner
	extras []any
}

func (r rowWithExtras) Scan(dest ...any) error {
	return r.row.Scan(append(dest, r.extras...)...)
}

// Discovery sections.
const (
	DiscoverRecent   = "recent"
	DiscoverPopular  = "popular"
	DiscoverTrending = "trending"
)

// trendingWindow is how far back engagement counts toward "trending".
const trendingWindow = 7 * 24 * time.Hour

// Discover lists public topics for a discovery section.
func (r *Repository) Discover(ctx context.Context, section string, limit, offset int) ([]Topic, int64, error) {
	w := &where{}
	publicTopics(w)
	var (
		order     string
		orderArgs []any
	)
	switch section {
	case DiscoverPopular:
		order = popularityScore + " DESC, t.created_at DESC"
	case DiscoverTrending:
		since := time.Now().UTC().Add(-trendingWindow)
		w.add("t.last_activity_at >= ?", since)
		// Engagement inside the window: recent likes weigh 3, recent replies 2.
		order = `((SELECT COUNT(*) FROM likes l WHERE l.target_type = 'topic' AND l.target_id = t.id AND l.created_at >= ?) * 3
            + (SELECT COUNT(*) FROM replies rr WHERE rr.topic_id = t.id AND rr.is_deleted = ? AND rr.created_at >= ?) * 2) DESC,
            t.last_activity_at DESC`
		orderArgs = []any{since, false, since}
	default:
		order = "t.created_at DESC"
	}

	var total int64
	countQ := "SELECT COUNT(*) FROM topics t JOIN categories c ON c.id = t.category_id" + w.sql()
	if err := r.store.DB.QueryRowContext(ctx, r.store.Rebind(countQ), w.args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []Topic{}, 0, nil
	}
	args := append(append([]any{}, w.args...), orderArgs...)
	args = append(args, limit, offset)
	items, err := r.queryTopics(ctx, topicSelect+w.sql()+" ORDER BY "+order+" LIMIT ? OFFSET ?", args...)
	return items, total, err
}

// Featured lists public topics flagged as featured, most recent activity first.
func (r *Repository) Featured(ctx context.Context, limit int) ([]Topic, error) {
	w := &where{}
	publicTopics(w)
	w.add("t.is_featured = ?", true)
	args := append(append([]any{}, w.args...), limit)
	return r.queryTopics(ctx, topicSelect+w.sql()+" ORDER BY t.last_activity_at DESC LIMIT ?", args...)
}
