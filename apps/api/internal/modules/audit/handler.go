package audit

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

// Handler exposes the audit log.
type Handler struct {
	svc *Service
}

// NewHandler builds the audit Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List handles GET /admin/audit?action=&actor=&target=&page=&limit=.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 30, 100)
	q := r.URL.Query()
	entries, total, err := h.svc.List(r.Context(), Filter{
		Action:  q.Get("action"),
		ActorID: q.Get("actor"),
		Target:  q.Get("target"),
		Q:       q.Get("q"),
		Asc:     q.Get("order") == "asc",
		Limit:   p.Limit,
		Offset:  p.Offset,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, entries, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// Actions handles GET /admin/audit/actions.
func (h *Handler) Actions(w http.ResponseWriter, r *http.Request) {
	actions, err := h.svc.Actions(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, actions)
}
