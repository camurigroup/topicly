package server

import (
	"net/http"

	"github.com/forum-framework/api/internal/middleware"
)

// routes builds the HTTP handler with all routes and the middleware chain.
func (s *Server) routes() http.Handler {
	root := http.NewServeMux()

	// Infra endpoints (outside the versioned API).
	root.HandleFunc("GET /health", s.health.Live)
	root.HandleFunc("GET /ready", s.health.Ready)

	api := http.NewServeMux()
	api.HandleFunc("GET /ping", s.health.Ping)

	// Auth (public).
	api.HandleFunc("POST /auth/register", s.auth.Register)
	api.HandleFunc("POST /auth/login", s.auth.Login)
	api.HandleFunc("POST /auth/refresh", s.auth.Refresh)
	api.HandleFunc("POST /auth/logout", s.auth.Logout)

	// Public profiles.
	api.HandleFunc("GET /users/{username}", s.user.PublicProfile)

	// Authenticated account endpoints.
	auth := s.authMW.RequireAuth
	api.Handle("GET /me", auth(http.HandlerFunc(s.user.Me)))
	api.Handle("PATCH /me", auth(http.HandlerFunc(s.user.UpdateProfile)))
	api.Handle("PUT /me/password", auth(http.HandlerFunc(s.user.ChangePassword)))
	api.Handle("PUT /me/email", auth(http.HandlerFunc(s.user.ChangeEmail)))

	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return middleware.Chain(root,
		middleware.RequestID,
		middleware.Logger(s.logger),
		middleware.Recover(s.logger),
		middleware.CORS(s.cfg.CORSAllowedOrigins),
	)
}
