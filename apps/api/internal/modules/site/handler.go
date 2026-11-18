package site

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

// maxBody fits every setting at its maximum length (legal texts included).
const maxBody = 1 << 18

// Handler exposes the home page and site settings endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds the site Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Home handles GET /home.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	home, err := h.svc.Home(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, home)
}

// Settings handles GET /site/settings (public values).
func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	values, err := h.svc.Settings(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, values)
}

// AdminSettings handles GET /admin/site/settings: definitions plus values.
func (h *Handler) AdminSettings(w http.ResponseWriter, r *http.Request) {
	values, err := h.svc.Settings(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"definitions": Definitions, "values": values})
}

// UpdateSettings handles PUT /admin/site/settings with a {key: value} map.
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var in map[string]string
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	values, err := h.svc.UpdateSettings(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, values)
}
