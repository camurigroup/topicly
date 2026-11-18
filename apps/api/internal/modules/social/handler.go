package social

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

// Handler exposes follow endpoints for users and categories.
type Handler struct {
	svc *Service
}

// NewHandler builds the social Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func viewerID(r *http.Request) string {
	if p, ok := authctx.From(r.Context()); ok {
		return p.UserID
	}
	return ""
}

// FollowUser handles POST /users/{username}/follow (toggle).
func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	p, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	res, err := h.svc.ToggleUser(r.Context(), p.UserID, p.Username, r.PathValue("username"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, res)
}

// FollowCategory handles POST /categories/{slug}/follow (toggle).
func (h *Handler) FollowCategory(w http.ResponseWriter, r *http.Request) {
	uid := viewerID(r)
	if uid == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	res, err := h.svc.ToggleCategory(r.Context(), uid, r.PathValue("slug"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, res)
}

// UserStats handles GET /users/{username}/stats.
func (h *Handler) UserStats(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.UserStats(r.Context(), r.PathValue("username"), viewerID(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, st)
}

// CategoryStats handles GET /categories/{slug}/stats.
func (h *Handler) CategoryStats(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.CategoryStats(r.Context(), r.PathValue("slug"), viewerID(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, st)
}

// Followers handles GET /users/{username}/followers.
func (h *Handler) Followers(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 50)
	items, total, err := h.svc.Followers(r.Context(), r.PathValue("username"), p.Limit, p.Offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, items, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// Following handles GET /users/{username}/following.
func (h *Handler) Following(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 50)
	items, total, err := h.svc.Following(r.Context(), r.PathValue("username"), p.Limit, p.Offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, items, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}
