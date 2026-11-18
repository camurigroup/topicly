package search

import (
	"net/http"
	"strconv"

	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

// Handler exposes the global search endpoint.
type Handler struct {
	svc *Service
}

// NewHandler builds the search Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Search handles GET /search?q=&type=&category=&author=&from=&to=
// &minReplies=&minLikes=&sort=&page=&limit=.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pg := httpx.Paginate(r, 20, 50)
	atoi := func(key string) int {
		n, _ := strconv.Atoi(q.Get(key))
		return n
	}
	res, err := h.svc.Search(r.Context(), Params{
		Query:      q.Get("q"),
		Type:       q.Get("type"),
		Category:   q.Get("category"),
		Author:     q.Get("author"),
		From:       q.Get("from"),
		To:         q.Get("to"),
		MinReplies: atoi("minReplies"),
		MinLikes:   atoi("minLikes"),
		Sort:       q.Get("sort"),
		Limit:      pg.Limit,
		Offset:     pg.Offset,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if res.Type == TypeAll {
		response.OK(w, res)
		return
	}
	response.Paginated(w, res, response.Meta{Total: sectionTotal(res), Page: pg.Page, Limit: pg.Limit})
}

func sectionTotal(res *Result) int64 {
	switch {
	case res.Topics != nil:
		return res.Topics.Total
	case res.Replies != nil:
		return res.Replies.Total
	case res.Users != nil:
		return res.Users.Total
	case res.Categories != nil:
		return res.Categories.Total
	}
	return 0
}
