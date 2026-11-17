package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/forum-framework/api/internal/shared/response"
)

// Handler serves liveness and readiness endpoints.
type Handler struct {
	db *sql.DB
}

// New builds a health Handler.
func New(db *sql.DB) *Handler { return &Handler{db: db} }

// Live reports that the process is up (no dependency checks).
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]any{"status": "ok"})
}

// Ping is a trivial authenticated-group smoke endpoint under /api/v1.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]any{"message": "pong"})
}

// Ready reports readiness, including database connectivity.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		response.Error(w, http.StatusServiceUnavailable, "db_unavailable", "database not configured")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		response.Error(w, http.StatusServiceUnavailable, "db_unavailable", "database not reachable")
		return
	}
	response.OK(w, map[string]any{"status": "ready", "database": "up"})
}
