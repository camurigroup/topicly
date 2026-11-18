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
	api.HandleFunc("GET /discover", s.topic.Discover)
	api.Handle("PUT /topics/{id}/pin", s.perm(rbac.PermTopicsPin, s.topic.Pin))
	api.Handle("PUT /topics/{id}/lock", s.perm(rbac.PermTopicsLock, s.topic.Lock))
	api.Handle("PUT /topics/{id}/feature", s.perm(rbac.PermTopicsPin, s.topic.Feature))

	// Replies.
	api.Handle("GET /topics/{id}/replies", optional(http.HandlerFunc(s.topic.ListReplies)))
	api.Handle("POST /topics/{id}/replies", auth(http.HandlerFunc(s.topic.CreateReply)))
	api.Handle("PATCH /replies/{id}", auth(http.HandlerFunc(s.topic.UpdateReply)))
	api.Handle("DELETE /replies/{id}", auth(http.HandlerFunc(s.topic.DeleteReply)))
	api.Handle("POST /replies/{id}/like", auth(http.HandlerFunc(s.topic.LikeReply)))

	// Global search (filters applied in the database).
	api.HandleFunc("GET /search", s.search.Search)

	// Social: follow users and categories, follower lists and counters.
	api.Handle("POST /users/{username}/follow", auth(http.HandlerFunc(s.social.FollowUser)))
	api.Handle("GET /users/{username}/stats", optional(http.HandlerFunc(s.social.UserStats)))
	api.HandleFunc("GET /users/{username}/followers", s.social.Followers)
	api.HandleFunc("GET /users/{username}/following", s.social.Following)
	api.Handle("POST /categories/{slug}/follow", auth(http.HandlerFunc(s.social.FollowCategory)))
	api.Handle("GET /categories/{slug}/stats", optional(http.HandlerFunc(s.social.CategoryStats)))

	// Notifications (live stream via SSE).
	api.Handle("GET /notifications", auth(http.HandlerFunc(s.notify.List)))
	api.Handle("GET /notifications/unread-count", auth(http.HandlerFunc(s.notify.UnreadCount)))
	api.Handle("GET /notifications/stream", auth(http.HandlerFunc(s.notify.Stream)))
	api.Handle("PUT /notifications/read-all", auth(http.HandlerFunc(s.notify.MarkAllRead)))
	api.Handle("PUT /notifications/{id}/read", auth(http.HandlerFunc(s.notify.MarkRead)))
	api.Handle("DELETE /notifications/{id}", auth(http.HandlerFunc(s.notify.Delete)))
	api.Handle("POST /admin/notifications", s.perm(rbac.PermSettingsManage, s.notify.Broadcast))

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
