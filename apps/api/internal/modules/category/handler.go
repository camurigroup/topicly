package category

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const maxBody = 1 << 20

// Handler exposes category endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds the category Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List handles GET /categories (public: active only).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context(), true)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, cats)
}

// ListAll handles GET /admin/categories (includes inactive).
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context(), false)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, cats)
}

// GetBySlug handles GET /categories/{slug}.
func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	cat, err := h.svc.GetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, cat)
}

// Create handles POST /categories (permission-gated).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	cat, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, cat)
}

// Update handles PATCH /categories/{id} (permission-gated).
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var in UpdateInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	cat, err := h.svc.Update(r.Context(), r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, cat)
}

// Delete handles DELETE /categories/{id} (permission-gated).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}
