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
	api.Handle("POST /categories", s.audited("category.create", "category", s.perm(rbac.PermCategoriesManage, s.category.Create)))
	api.Handle("PATCH /categories/{id}", s.audited("category.update", "category", s.perm(rbac.PermCategoriesManage, s.category.Update)))
	api.Handle("DELETE /categories/{id}", s.audited("category.delete", "category", s.perm(rbac.PermCategoriesManage, s.category.Delete)))

	// Topics (reads use optional auth to enrich viewer flags; writes are
	// blocked for muted users).
	api.Handle("GET /topics", optional(http.HandlerFunc(s.topic.ListTopics)))
	api.Handle("GET /topics/{id}", optional(http.HandlerFunc(s.topic.GetTopic)))
	api.Handle("POST /topics", s.write(s.topic.CreateTopic))
	api.Handle("PATCH /topics/{id}", s.write(s.topic.UpdateTopic))
	api.Handle("DELETE /topics/{id}", auth(http.HandlerFunc(s.topic.DeleteTopic)))
	api.Handle("POST /topics/{id}/like", s.write(s.topic.LikeTopic))
	api.Handle("POST /topics/{id}/follow", s.write(s.topic.FollowTopic))
	api.HandleFunc("GET /discover", s.topic.Discover)
	api.Handle("PUT /topics/{id}/pin", s.audited("topic.pin", "topic", s.perm(rbac.PermTopicsPin, s.topic.Pin)))
	api.Handle("PUT /topics/{id}/lock", s.audited("topic.lock", "topic", s.perm(rbac.PermTopicsLock, s.topic.Lock)))
	api.Handle("PUT /topics/{id}/feature", s.audited("topic.feature", "topic", s.perm(rbac.PermTopicsPin, s.topic.Feature)))

	// Replies.
	api.Handle("GET /topics/{id}/replies", optional(http.HandlerFunc(s.topic.ListReplies)))
	api.Handle("POST /topics/{id}/replies", s.write(s.topic.CreateReply))
	api.Handle("PATCH /replies/{id}", s.write(s.topic.UpdateReply))
	api.Handle("DELETE /replies/{id}", auth(http.HandlerFunc(s.topic.DeleteReply)))
	api.Handle("POST /replies/{id}/like", s.write(s.topic.LikeReply))

	// Home page and administrable site texts.
	api.HandleFunc("GET /home", s.site.Home)
	api.HandleFunc("GET /site/settings", s.site.Settings)
	api.Handle("GET /admin/site/settings", s.perm(rbac.PermSettingsManage, s.site.AdminSettings))
	api.Handle("PUT /admin/site/settings", s.audited("site.settings", "site", s.perm(rbac.PermSettingsManage, s.site.UpdateSettings)))

	// Global search (filters applied in the database).
	api.HandleFunc("GET /search", s.search.Search)

	// Social: follow users and categories, follower lists and counters.
	api.Handle("POST /users/{username}/follow", s.write(s.social.FollowUser))
	api.Handle("GET /users/{username}/stats", optional(http.HandlerFunc(s.social.UserStats)))
	api.HandleFunc("GET /users/{username}/followers", s.social.Followers)
	api.HandleFunc("GET /users/{username}/following", s.social.Following)
	api.Handle("POST /categories/{slug}/follow", s.write(s.social.FollowCategory))
	api.Handle("GET /categories/{slug}/stats", optional(http.HandlerFunc(s.social.CategoryStats)))

	// Notifications (live stream via SSE).
	api.Handle("GET /notifications", auth(http.HandlerFunc(s.notify.List)))
	api.Handle("GET /notifications/unread-count", auth(http.HandlerFunc(s.notify.UnreadCount)))
	api.Handle("GET /notifications/stream", auth(http.HandlerFunc(s.notify.Stream)))
	api.Handle("PUT /notifications/read-all", auth(http.HandlerFunc(s.notify.MarkAllRead)))
	api.Handle("PUT /notifications/{id}/read", auth(http.HandlerFunc(s.notify.MarkRead)))
	api.Handle("DELETE /notifications/{id}", auth(http.HandlerFunc(s.notify.Delete)))
	api.Handle("POST /admin/notifications", s.audited("notification.broadcast", "site", s.perm(rbac.PermSettingsManage, s.notify.Broadcast)))

	s.moderationRoutes(api)

	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return middleware.Chain(root,
		middleware.RequestID,
		withClientIP,
		middleware.Logger(s.logger),
		middleware.Recover(s.logger),
		middleware.CORS(s.cfg.CORSAllowedOrigins),
	)
}

// moderationRoutes registers RBAC management, moderation, reports and audit
// routes (Etapa 5). Every handler re-checks hierarchy rules server-side.
func (s *Server) moderationRoutes(api *http.ServeMux) {
	ra := s.rbacAdmin
	api.Handle("GET /admin/roles", s.perm(rbac.PermRolesManage, ra.ListRoles))
	api.Handle("POST /admin/roles", s.perm(rbac.PermRolesManage, ra.CreateRole))
	api.Handle("PATCH /admin/roles/{id}", s.perm(rbac.PermRolesManage, ra.UpdateRole))
	api.Handle("PUT /admin/roles/{id}/permissions", s.perm(rbac.PermRolesManage, ra.SetRolePermissions))
	api.Handle("DELETE /admin/roles/{id}", s.perm(rbac.PermRolesManage, ra.DeleteRole))
	api.Handle("GET /admin/permissions", s.anyPerm(ra.ListPermissions, rbac.PermRolesManage, rbac.PermPermissionsManage))
	api.Handle("POST /admin/permissions", s.perm(rbac.PermPermissionsManage, ra.CreatePermission))
	api.Handle("PATCH /admin/permissions/{id}", s.perm(rbac.PermPermissionsManage, ra.SetPermissionActive))
	api.Handle("DELETE /admin/permissions/{id}", s.perm(rbac.PermPermissionsManage, ra.DeletePermission))
	api.Handle("GET /admin/users/{username}/roles", s.perm(rbac.PermRolesManage, ra.UserRoles))
	api.Handle("PUT /admin/users/{username}/roles", s.perm(rbac.PermRolesManage, ra.SetUserRoles))

	m := s.moderation
	anyMod := []string{rbac.PermUsersBan, rbac.PermUsersMute, rbac.PermWarningsApply, rbac.PermRolesManage}
	api.Handle("GET /admin/users/{username}/moderation", s.anyPerm(m.UserSummary, anyMod...))
	api.Handle("POST /admin/users/{username}/ban", s.perm(rbac.PermUsersBan, m.Ban))
	api.Handle("POST /admin/users/{username}/mute", s.perm(rbac.PermUsersMute, m.Mute))
	api.Handle("POST /admin/users/{username}/warnings", s.perm(rbac.PermWarningsApply, m.Warn))
	api.Handle("GET /admin/sanctions", s.anyPerm(m.ListSanctions, rbac.PermUsersBan, rbac.PermUsersMute))
	api.Handle("POST /admin/sanctions/{id}/revoke", s.anyPerm(m.Revoke, rbac.PermUsersBan, rbac.PermUsersMute))

	api.Handle("POST /reports", s.authMW.RequireAuth(s.reportLimit.Limit(http.HandlerFunc(m.CreateReport))))
	api.Handle("GET /admin/reports", s.perm(rbac.PermReportsManage, m.ListReports))
	api.Handle("GET /admin/reports/counts", s.perm(rbac.PermReportsManage, m.ReportCounts))
	api.Handle("PATCH /admin/reports/{id}", s.perm(rbac.PermReportsManage, m.UpdateReport))

	api.Handle("GET /admin/audit", s.perm(rbac.PermLogsView, s.audit.List))
	api.Handle("GET /admin/audit/actions", s.perm(rbac.PermLogsView, s.audit.Actions))
}

// perm composes RequireAuth + RequirePermission around a handler func.
func (s *Server) perm(slug string, h http.HandlerFunc) http.Handler {
	return s.authMW.RequireAuth(s.authMW.RequirePermission(slug)(http.HandlerFunc(h)))
}
