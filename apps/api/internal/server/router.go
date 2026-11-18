package server

import (
	"net/http"

	"github.com/forum-framework/api/internal/middleware"
	"github.com/forum-framework/api/internal/modules/rbac"
)

// routes builds the HTTP handler with all routes and the middleware chain.
func (s *Server) routes() http.Handler {
	root := http.NewServeMux()

	root.HandleFunc("GET /health", s.health.Live)
	root.HandleFunc("GET /ready", s.health.Ready)
	if fs, ok := s.fileServer(); ok {
		root.Handle("GET /files/", fs)
	}

	api := http.NewServeMux()
	api.HandleFunc("GET /ping", s.health.Ping)

	auth := s.authMW.RequireAuth
	optional := s.authMW.OptionalAuth

	// Auth (public, rate-limited).
	limit := s.authLimiter.Limit
	api.Handle("POST /auth/register", limit(http.HandlerFunc(s.auth.Register)))
	api.Handle("POST /auth/login", limit(http.HandlerFunc(s.auth.Login)))
	api.Handle("POST /auth/refresh", limit(http.HandlerFunc(s.auth.Refresh)))
	api.HandleFunc("POST /auth/logout", s.auth.Logout)

	// Public profiles.
	api.HandleFunc("GET /users/{username}", s.user.PublicProfile)

	// Account endpoints.
	api.Handle("GET /me", auth(http.HandlerFunc(s.user.Me)))
	api.Handle("PATCH /me", auth(http.HandlerFunc(s.user.UpdateProfile)))
	api.Handle("PUT /me/password", auth(http.HandlerFunc(s.user.ChangePassword)))
	api.Handle("PUT /me/email", auth(http.HandlerFunc(s.user.ChangeEmail)))
	api.Handle("POST /me/avatar", auth(http.HandlerFunc(s.user.UploadAvatar)))
	api.Handle("POST /me/banner", auth(http.HandlerFunc(s.user.UploadBanner)))

	// Categories.
	api.HandleFunc("GET /categories", s.category.List)
	api.HandleFunc("GET /categories/{slug}", s.category.GetBySlug)
	api.Handle("GET /admin/categories", s.perm(rbac.PermCategoriesManage, s.category.ListAll))
	api.Handle("POST /categories", s.perm(rbac.PermCategoriesManage, s.category.Create))
	api.Handle("PATCH /categories/{id}", s.perm(rbac.PermCategoriesManage, s.category.Update))
	api.Handle("DELETE /categories/{id}", s.perm(rbac.PermCategoriesManage, s.category.Delete))

	// Topics (reads use optional auth to enrich viewer flags).
	api.Handle("GET /topics", optional(http.HandlerFunc(s.topic.ListTopics)))
	api.Handle("GET /topics/{id}", optional(http.HandlerFunc(s.topic.GetTopic)))
	api.Handle("POST /topics", auth(http.HandlerFunc(s.topic.CreateTopic)))
	api.Handle("PATCH /topics/{id}", auth(http.HandlerFunc(s.topic.UpdateTopic)))
	api.Handle("DELETE /topics/{id}", auth(http.HandlerFunc(s.topic.DeleteTopic)))
	api.Handle("POST /topics/{id}/like", auth(http.HandlerFunc(s.topic.LikeTopic)))
	api.Handle("POST /topics/{id}/follow", auth(http.HandlerFunc(s.topic.FollowTopic)))
	api.Handle("PUT /topics/{id}/pin", s.perm(rbac.PermTopicsPin, s.topic.Pin))
	api.Handle("PUT /topics/{id}/lock", s.perm(rbac.PermTopicsLock, s.topic.Lock))
	api.Handle("PUT /topics/{id}/feature", s.perm(rbac.PermTopicsPin, s.topic.Feature))

	// Replies.
	api.Handle("GET /topics/{id}/replies", optional(http.HandlerFunc(s.topic.ListReplies)))
	api.Handle("POST /topics/{id}/replies", auth(http.HandlerFunc(s.topic.CreateReply)))
	api.Handle("PATCH /replies/{id}", auth(http.HandlerFunc(s.topic.UpdateReply)))
	api.Handle("DELETE /replies/{id}", auth(http.HandlerFunc(s.topic.DeleteReply)))
	api.Handle("POST /replies/{id}/like", auth(http.HandlerFunc(s.topic.LikeReply)))

	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return middleware.Chain(root,
		middleware.RequestID,
		middleware.Logger(s.logger),
		middleware.Recover(s.logger),
		middleware.CORS(s.cfg.CORSAllowedOrigins),
	)
}

// perm composes RequireAuth + RequirePermission around a handler func.
func (s *Server) perm(slug string, h http.HandlerFunc) http.Handler {
	return s.authMW.RequireAuth(s.authMW.RequirePermission(slug)(http.HandlerFunc(h)))
}
