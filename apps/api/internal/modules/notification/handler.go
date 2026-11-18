package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const (
	maxBody   = 1 << 16
	heartbeat = 25 * time.Second
)

// Handler exposes notification endpoints, including the live SSE stream.
type Handler struct {
	svc *Service
}

// NewHandler builds the notification Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func principal(w http.ResponseWriter, r *http.Request) (string, bool) {
	p, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return "", false
	}
	return p.UserID, true
}

// List handles GET /notifications?unread=true&page=&limit=.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	p := httpx.Paginate(r, 20, 50)
	unreadOnly := r.URL.Query().Get("unread") == "true"
	items, total, err := h.svc.List(r.Context(), uid, unreadOnly, p.Limit, p.Offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Paginated(w, items, response.Meta{Total: total, Page: p.Page, Limit: p.Limit})
}

// UnreadCount handles GET /notifications/unread-count.
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	n, err := h.svc.Unread(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"unread": n})
}

// MarkRead handles PUT /notifications/{id}/read.
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	if err := h.svc.MarkRead(r.Context(), uid, r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"updated": true})
}

// MarkAllRead handles PUT /notifications/read-all.
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	n, err := h.svc.MarkAllRead(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"updated": n})
}

// Delete handles DELETE /notifications/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), uid, r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}

type broadcastRequest struct {
	Message string `json:"message"`
	Link    string `json:"link"`
}

// Broadcast handles POST /admin/notifications (permission-gated).
func (h *Handler) Broadcast(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	var in broadcastRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	n, err := h.svc.Broadcast(r.Context(), uid, in.Message, in.Link)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, map[string]any{"sent": n})
}

// Stream handles GET /notifications/stream as Server-Sent Events. The client
// authenticates with the usual bearer header (fetch-based SSE), so the access
// token never travels in the URL.
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	uid, ok := principal(w, r)
	if !ok {
		return
	}
	rc := http.NewResponseController(w)
	// Streams outlive the server's WriteTimeout; lift it for this connection.
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		response.Error(w, http.StatusInternalServerError, "stream_unsupported", "streaming indisponível")
		return
	}

	events, unsubscribe := h.svc.Subscribe(uid)
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	unread, _ := h.svc.Unread(r.Context(), uid)
	if err := writeEvent(w, StreamEvent{Kind: "sync", Unread: unread}); err != nil || rc.Flush() != nil {
		return
	}

	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-events:
			if err := writeEvent(w, ev); err != nil || rc.Flush() != nil {
				return
			}
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil || rc.Flush() != nil {
				return
			}
		}
	}
}

func writeEvent(w http.ResponseWriter, ev StreamEvent) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Kind, payload)
	return err
}
