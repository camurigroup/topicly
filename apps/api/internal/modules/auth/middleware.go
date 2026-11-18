package auth

import (
	"context"
	"net/http"

	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
	"github.com/forum-framework/api/internal/shared/security"
)

// BanCheck reports whether a user is currently banned, with a message
// describing the ban. It is provided by the moderation module.
type BanCheck func(ctx context.Context, userID string) (message string, banned bool, err error)

// Middleware guards routes with JWT authentication and permission checks.
type Middleware struct {
	secret   string
	rbac     *rbac.Service
	banCheck BanCheck
}

// NewMiddleware builds the auth Middleware.
func NewMiddleware(secret string, rbacSvc *rbac.Service) *Middleware {
	return &Middleware{secret: secret, rbac: rbacSvc}
}

// SetBanCheck enables ban enforcement: banned users are rejected even with a
// still-valid access token (the ban is effective immediately).
func (m *Middleware) SetBanCheck(check BanCheck) { m.banCheck = check }

// banned reports whether userID is banned (false when no check is set).
func (m *Middleware) banned(ctx context.Context, userID string) (string, bool, error) {
	if m.banCheck == nil {
		return "", false, nil
	}
	return m.banCheck(ctx, userID)
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
		msg, isBanned, err := m.banned(r.Context(), claims.Subject)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "internal_error", "erro interno")
			return
		}
		if isBanned {
			response.Error(w, http.StatusForbidden, "banned", msg)
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

// OptionalAuth sets the principal when a valid token is present but never
// rejects the request. Used on public reads that enrich output for logged-in users.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := httpx.BearerToken(r); token != "" {
			claims, err := security.ParseAccessToken(m.secret, token)
			if err == nil {
				if _, isBanned, berr := m.banned(r.Context(), claims.Subject); berr != nil || isBanned {
					next.ServeHTTP(w, r) // banned viewers browse as anonymous
					return
				}
				ctx := authctx.With(r.Context(), authctx.Principal{UserID: claims.Subject, Username: claims.Username})
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
