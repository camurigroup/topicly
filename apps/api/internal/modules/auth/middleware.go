package auth

import (
	"net/http"

	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
	"github.com/forum-framework/api/internal/shared/security"
)

// Middleware guards routes with JWT authentication and permission checks.
type Middleware struct {
	secret string
	rbac   *rbac.Service
}

// NewMiddleware builds the auth Middleware.
func NewMiddleware(secret string, rbacSvc *rbac.Service) *Middleware {
	return &Middleware{secret: secret, rbac: rbacSvc}
}

// RequireAuth validates the bearer access token and stores the principal.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := httpx.BearerToken(r)
		if token == "" {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
			return
		}
		claims, err := security.ParseAccessToken(m.secret, token)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "token inválido ou expirado")
			return
		}
		ctx := authctx.With(r.Context(), authctx.Principal{UserID: claims.Subject, Username: claims.Username})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission ensures the principal holds the given permission slug.
// It must be chained after RequireAuth.
func (m *Middleware) RequirePermission(slug string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := authctx.From(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
				return
			}
			has, err := m.rbac.HasPermission(r.Context(), principal.UserID, slug)
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "internal_error", "erro interno")
				return
			}
			if !has {
				response.Error(w, http.StatusForbidden, "forbidden", "permissão insuficiente")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
