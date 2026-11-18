package auth

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const maxBody = 1 << 20 // 1 MiB

// Handler exposes authentication endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds the auth Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Register handles POST /auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	res, err := h.svc.Register(r.Context(), in, httpx.UserAgent(r), httpx.ClientIP(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, res)
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	res, err := h.svc.Login(r.Context(), in, httpx.UserAgent(r), httpx.ClientIP(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, res)
}

// Refresh handles POST /auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	pair, err := h.svc.Refresh(r.Context(), in.RefreshToken, httpx.UserAgent(r), httpx.ClientIP(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, pair)
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var in refreshRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	if err := h.svc.Logout(r.Context(), in.RefreshToken); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"loggedOut": true})
}
