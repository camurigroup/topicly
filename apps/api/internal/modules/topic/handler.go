package topic

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const maxBody = 1 << 20

// Handler exposes topic, reply, like and follow endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds the topic Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func viewerID(r *http.Request) string {
	if p, ok := authctx.From(r.Context()); ok {
		return p.UserID
	}
	return ""
}

func requirePrincipal(w http.ResponseWriter, r *http.Request) (string, bool) {
	p, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return "", false
	}
	return p.UserID, true
}

// CreateTopic handles POST /topics.
func (h *Handler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var in CreateTopicInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	t, err := h.svc.CreateTopic(r.Context(), uid, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, t)
}

// ListTopics handles GET /topics?category=&page=&limit=.
func (h *Handler) ListTopics(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 50)
	category := r.URL.Query().Get("category")
	topics, total, err := h.svc.ListTopics(r.Context(), category, p.Limit, p.Offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, topics, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// GetTopic handles GET /topics/{id}.
func (h *Handler) GetTopic(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.GetTopic(r.Context(), r.PathValue("id"), viewerID(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, t)
}

// UpdateTopic handles PATCH /topics/{id}.
func (h *Handler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var in UpdateTopicInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	t, err := h.svc.UpdateTopic(r.Context(), uid, r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, t)
}

// DeleteTopic handles DELETE /topics/{id}.
func (h *Handler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteTopic(r.Context(), uid, r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}

type flagRequest struct {
	Value bool `json:"value"`
}

func (h *Handler) setFlag(w http.ResponseWriter, r *http.Request, column string) {
	var in flagRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	if err := h.svc.SetFlag(r.Context(), r.PathValue("id"), column, in.Value); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"updated": true})
}

// Pin handles PUT /topics/{id}/pin (permission-gated).
func (h *Handler) Pin(w http.ResponseWriter, r *http.Request) { h.setFlag(w, r, "is_pinned") }

// Lock handles PUT /topics/{id}/lock (permission-gated).
func (h *Handler) Lock(w http.ResponseWriter, r *http.Request) { h.setFlag(w, r, "is_locked") }

// Feature handles PUT /topics/{id}/feature (permission-gated).
func (h *Handler) Feature(w http.ResponseWriter, r *http.Request) { h.setFlag(w, r, "is_featured") }

// CreateReply handles POST /topics/{id}/replies.
func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var in CreateReplyInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	rp, err := h.svc.CreateReply(r.Context(), uid, r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, rp)
}

// ListReplies handles GET /topics/{id}/replies.
func (h *Handler) ListReplies(w http.ResponseWriter, r *http.Request) {
	p := httpx.Paginate(r, 20, 100)
	replies, total, err := h.svc.ListReplies(r.Context(), r.PathValue("id"), viewerID(r), p.Limit, p.Offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, replies, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// UpdateReply handles PATCH /replies/{id}.
func (h *Handler) UpdateReply(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var in UpdateReplyInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	rp, err := h.svc.UpdateReply(r.Context(), uid, r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, rp)
}

// DeleteReply handles DELETE /replies/{id}.
func (h *Handler) DeleteReply(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteReply(r.Context(), uid, r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}

func writeToggle(w http.ResponseWriter, active bool, count int, activeKey string) {
	response.OK(w, map[string]any{activeKey: active, "count": count})
}

// LikeTopic handles POST /topics/{id}/like.
func (h *Handler) LikeTopic(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	liked, count, err := h.svc.ToggleTopicLike(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	writeToggle(w, liked, count, "liked")
}

// FollowTopic handles POST /topics/{id}/follow.
func (h *Handler) FollowTopic(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	following, count, err := h.svc.ToggleTopicFollow(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	writeToggle(w, following, count, "following")
}

// LikeReply handles POST /replies/{id}/like.
func (h *Handler) LikeReply(w http.ResponseWriter, r *http.Request) {
	uid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	liked, count, err := h.svc.ToggleReplyLike(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	writeToggle(w, liked, count, "liked")
}
