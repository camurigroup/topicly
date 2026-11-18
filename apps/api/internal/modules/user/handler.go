package user

import (
	"net/http"

	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const maxBody = 1 << 20

// Handler exposes user and profile endpoints.
type Handler struct {
	svc  *Service
	rbac *rbac.Service
}

// NewHandler builds the user Handler.
func NewHandler(svc *Service, rbacSvc *rbac.Service) *Handler {
	return &Handler{svc: svc, rbac: rbacSvc}
}

// PublicProfile handles GET /users/{username}.
func (h *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	profile, err := h.svc.PublicProfile(r.Context(), username)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, profile)
}

// Me handles GET /me and returns the account plus roles and permissions.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	account, err := h.svc.Account(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	roles, err := h.rbac.RolesForUser(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	perms, err := h.rbac.PermissionsForUser(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"user": account, "roles": roles, "permissions": perms})
}

// UpdateProfile handles PATCH /me.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	var in UpdateProfileInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	account, err := h.svc.UpdateProfile(r.Context(), principal.UserID, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, account)
}

// ChangePassword handles PUT /me/password.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	principal, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	var in ChangePasswordInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	if err := h.svc.ChangePassword(r.Context(), principal.UserID, in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"updated": true})
}

// ChangeEmail handles PUT /me/email.
func (h *Handler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	principal, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return
	}
	var in ChangeEmailInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	account, err := h.svc.ChangeEmail(r.Context(), principal.UserID, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, account)
}
