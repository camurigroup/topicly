package server

import (
	"net/http"

	"github.com/forum-framework/api/internal/shared/authctx"
	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/reqmeta"
	"github.com/forum-framework/api/internal/shared/response"
)

// withClientIP stores the client IP in the request context (for audit logs).
func withClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(reqmeta.WithIP(r.Context(), httpx.ClientIP(r))))
	})
}

// statusRecorder captures the response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// audited records an audit entry for admin routes of modules that do not
// audit themselves, once the handler succeeds (2xx). The target is taken
// from the {id}, {slug} or {username} path value.
func (s *Server) audited(action, targetType string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		h.ServeHTTP(rec, r)
		if rec.status < 200 || rec.status >= 300 {
			return
		}
		p, ok := authctx.From(r.Context())
		if !ok {
			return
		}
		target := r.PathValue("id")
		for _, key := range []string{"slug", "username"} {
			if target == "" {
				target = r.PathValue(key)
			}
		}
		s.auditSvc.Record(r.Context(), p.UserID, action, targetType, target, map[string]any{
			"method": r.Method, "path": r.URL.Path,
		})
	})
}

// anyPerm requires authentication and at least one of the given permissions.
func (s *Server) anyPerm(h http.HandlerFunc, slugs ...string) http.Handler {
	check := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := authctx.From(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "autenticação necessária")
			return
		}
		for _, slug := range slugs {
			has, err := s.rbac.HasPermission(r.Context(), p.UserID, slug)
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "internal_error", "erro interno")
				return
			}
			if has {
				h.ServeHTTP(w, r)
				return
			}
		}
		response.Error(w, http.StatusForbidden, "forbidden", "permissão insuficiente")
	})
	return s.authMW.RequireAuth(check)
}

// write requires authentication and blocks muted users.
func (s *Server) write(h http.HandlerFunc) http.Handler {
	return s.authMW.RequireAuth(s.moderation.RequireUnmuted(h))
}
