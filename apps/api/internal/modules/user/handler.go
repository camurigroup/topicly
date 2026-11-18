package user

import (
	"context"
	"net/http"

	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/id"
	"github.com/forum-framework/api/internal/shared/response"
	"github.com/forum-framework/api/internal/storage"
)

const maxBody = 1 << 20

// Handler exposes user and profile endpoints.
type Handler struct {
	svc     *Service
	rbac    *rbac.Service
	storage storage.Provider
}

// NewHandler builds the user Handler.
func NewHandler(svc *Service, rbacSvc *rbac.Service, store storage.Provider) *Handler {
	return &Handler{svc: svc, rbac: rbacSvc, storage: store}
}

func principalOrUnauthorized(w http.ResponseWriter, r *http.Request) (authctx.Principal, bool) {
	principal, ok := authctx.From(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
	}
	return principal, ok
}

// PublicProfile handles GET /users/{username}.
func (h *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.svc.PublicProfile(r.Context(), r.PathValue("username"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, profile)
}

// Me handles GET /me and returns the account plus roles and permissions.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalOrUnauthorized(w, r)
	if !ok {
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
	principal, ok := principalOrUnauthorized(w, r)
	if !ok {
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
	principal, ok := principalOrUnauthorized(w, r)
	if !ok {
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
	principal, ok := principalOrUnauthorized(w, r)
	if !ok {
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

// UploadAvatar handles POST /me/avatar (multipart form, field "file").
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "avatars", h.svc.SetAvatar)
}

// UploadBanner handles POST /me/banner (multipart form, field "file").
func (h *Handler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "banners", h.svc.SetBanner)
}

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request, prefix string, set func(context.Context, string, string) (*Account, error)) {
	principal, ok := principalOrUnauthorized(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(storage.MaxImageBytes); err != nil {
		httpx.WriteValidation(w, "arquivo inválido ou muito grande")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteValidation(w, "campo 'file' é obrigatório")
		return
	}
	defer file.Close()
	if header.Size > storage.MaxImageBytes {
		httpx.WriteValidation(w, "o arquivo excede o limite de 5 MB")
		return
	}
	contentType, ext, err := storage.DetectImage(file)
	if err != nil {
		httpx.WriteValidation(w, err.Error())
		return
	}
	key := prefix + "/" + principal.UserID + "-" + id.New() + ext
	url, err := h.storage.Save(r.Context(), key, file, header.Size, contentType)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	account, err := set(r.Context(), principal.UserID, url)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	response.OK(w, account)
}
