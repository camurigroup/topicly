package rbac

import (
	"context"
	"net/http"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

const maxBody = 1 << 16

// UserLookup resolves a username to a user id (implemented outside rbac to
// avoid an import cycle with the user module).
type UserLookup func(ctx context.Context, username string) (string, error)

// AdminHandler exposes role and permission management endpoints.
type AdminHandler struct {
	svc    *Service
	lookup UserLookup
}

// NewAdminHandler builds the AdminHandler.
func NewAdminHandler(svc *Service, lookup UserLookup) *AdminHandler {
	return &AdminHandler{svc: svc, lookup: lookup}
}

func principalID(w http.ResponseWriter, r *http.Request) (string, bool) {
	p, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
		return "", false
	}
	return p.UserID, true
}

// ListRoles handles GET /admin/roles.
func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListRoles(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, roles)
}

// CreateRole handles POST /admin/roles.
func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	var in CreateRoleInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	role, err := h.svc.CreateRole(r.Context(), uid, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, role)
}

// UpdateRole handles PATCH /admin/roles/{id}.
func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	var in UpdateRoleInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	role, err := h.svc.UpdateRole(r.Context(), uid, r.PathValue("id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, role)
}

type permissionsRequest struct {
	Permissions []string `json:"permissions"`
}

// SetRolePermissions handles PUT /admin/roles/{id}/permissions.
func (h *AdminHandler) SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	var in permissionsRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	perms, err := h.svc.SetRolePermissions(r.Context(), uid, r.PathValue("id"), in.Permissions)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"permissions": perms})
}

// DeleteRole handles DELETE /admin/roles/{id}.
func (h *AdminHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteRole(r.Context(), uid, r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}

// ListPermissions handles GET /admin/permissions.
func (h *AdminHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.svc.ListPermissions(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, perms)
}

// CreatePermission handles POST /admin/permissions.
func (h *AdminHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	var in CreatePermissionInput
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	p, err := h.svc.CreatePermission(r.Context(), uid, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.Created(w, p)
}

type activeRequest struct {
	IsActive bool `json:"isActive"`
}

// SetPermissionActive handles PATCH /admin/permissions/{id}.
func (h *AdminHandler) SetPermissionActive(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	var in activeRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	p, err := h.svc.SetPermissionActive(r.Context(), uid, r.PathValue("id"), in.IsActive)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, p)
}

// DeletePermission handles DELETE /admin/permissions/{id}.
func (h *AdminHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeletePermission(r.Context(), uid, r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}

type rolesRequest struct {
	Roles []string `json:"roles"`
}

// UserRoles handles GET /admin/users/{username}/roles.
func (h *AdminHandler) UserRoles(w http.ResponseWriter, r *http.Request) {
	targetID, err := h.lookup(r.Context(), r.PathValue("username"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	roles, err := h.svc.UserRoles(r.Context(), targetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, roles)
}

// SetUserRoles handles PUT /admin/users/{username}/roles.
func (h *AdminHandler) SetUserRoles(w http.ResponseWriter, r *http.Request) {
	uid, ok := principalID(w, r)
	if !ok {
		return
	}
	var in rolesRequest
	if err := httpx.DecodeJSON(r, &in, maxBody); err != nil {
		httpx.WriteValidation(w, "corpo da requisição inválido")
		return
	}
	targetID, err := h.lookup(r.Context(), r.PathValue("username"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	roles, err := h.svc.SetUserRoles(r.Context(), uid, targetID, in.Roles)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, roles)
}
