package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/forum-framework/api/internal/config"
	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/middleware"
	"github.com/forum-framework/api/internal/modules/admin"
	"github.com/forum-framework/api/internal/modules/audit"
	"github.com/forum-framework/api/internal/modules/auth"
	"github.com/forum-framework/api/internal/modules/category"
	"github.com/forum-framework/api/internal/modules/health"
	"github.com/forum-framework/api/internal/modules/install"
	"github.com/forum-framework/api/internal/modules/moderation"
	"github.com/forum-framework/api/internal/modules/notification"
	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/modules/search"
	"github.com/forum-framework/api/internal/modules/site"
	"github.com/forum-framework/api/internal/modules/social"
	"github.com/forum-framework/api/internal/modules/topic"
	"github.com/forum-framework/api/internal/modules/user"
	"github.com/forum-framework/api/internal/storage"
)

// Server wires configuration, database, storage and the module handlers.
type Server struct {
	cfg     config.Config
	db      *sql.DB
	dialect string
	logger  *slog.Logger
	storage storage.Provider

	health      *health.Handler
	auth        *auth.Handler
	authMW      *auth.Middleware
	authLimiter *middleware.RateLimiter
	user        *user.Handler
	category    *category.Handler
	topic       *topic.Handler
	notify      *notification.Handler
	social      *social.Handler
	search      *search.Handler
	site        *site.Handler
	siteSvc     *site.Service
	rbac        *rbac.Service
	rbacAdmin   *rbac.AdminHandler
	moderation  *moderation.Handler
	audit       *audit.Handler
	auditSvc    *audit.Service
	reportLimit *middleware.RateLimiter
	admin       *admin.Handler
	install     *install.Handler
	installSvc  *install.Service
}

// New builds the Server and all module dependencies for the given dialect.
func New(cfg config.Config, db *sql.DB, dialect string, logger *slog.Logger) (*Server, error) {
	store := database.NewStore(db, dialect)

	fileStore, err := storage.New(cfg.Storage)
	if err != nil {
		return nil, err
	}

	userRepo := user.NewRepository(store)
	rbacRepo := rbac.NewRepository(store)
	tokenRepo := auth.NewRefreshTokenRepository(store)
	categoryRepo := category.NewRepository(store)
	topicRepo := topic.NewRepository(store)
	notifyRepo := notification.NewRepository(store)
	socialRepo := social.NewRepository(store)

	rbacSvc := rbac.NewService(rbacRepo)
	notifySvc := notification.NewService(notifyRepo, notification.NewHub(), logger)
	userSvc := user.NewService(userRepo)
	categorySvc := category.NewService(categoryRepo)
	topicSvc := topic.NewService(topicRepo, rbacSvc, notifySvc)
	siteSvc := site.NewService(site.NewRepository(store), topicRepo, categoryRepo)
	authSvc := auth.NewService(userRepo, tokenRepo, rbacSvc, auth.Config{
		AccessSecret: cfg.JWTAccessSecret,
		AccessTTL:    cfg.JWTAccessTTL,
		RefreshTTL:   cfg.JWTRefreshTTL,
	})
	authMW := auth.NewMiddleware(cfg.JWTAccessSecret, rbacSvc)

	// Audit + moderation (Etapa 5).
	auditSvc := audit.NewService(audit.NewRepository(store), logger)
	rbacSvc.SetAuditor(auditSvc)
	topicSvc.SetAuditor(auditSvc)
	modSvc := moderation.NewService(moderation.NewRepository(store), rbacSvc, auditSvc, notifySvc, authSvc.RevokeAllSessions)
	banCheck := func(ctx context.Context, userID string) (string, bool, error) {
		s, err := modSvc.Restriction(ctx, userID, moderation.TypeBan)
		if err != nil || s == nil {
			return "", false, err
		}
		return moderation.Describe(s), true, nil
	}
	authSvc.SetBanCheck(banCheck)
	authMW.SetBanCheck(banCheck)
	lookupUser := func(ctx context.Context, username string) (string, error) {
		u, err := userRepo.GetByUsername(ctx, username)
		if err != nil {
			return "", err
		}
		return u.ID, nil
	}

	installSvc := install.NewService(store, authSvc, siteSvc, categorySvc, os.Getenv("INSTALL_TOKEN"))

	return &Server{
		cfg:         cfg,
		db:          db,
		dialect:     dialect,
		logger:      logger,
		storage:     fileStore,
		health:      health.New(db),
		auth:        auth.NewHandler(authSvc),
		authMW:      authMW,
		authLimiter: middleware.NewRateLimiter(20, time.Minute),
		user:        user.NewHandler(userSvc, rbacSvc, fileStore),
		category:    category.NewHandler(categorySvc),
		topic:       topic.NewHandler(topicSvc),
		notify:      notification.NewHandler(notifySvc),
		social:      social.NewHandler(social.NewService(socialRepo, notifySvc)),
		search:      search.NewHandler(search.NewService(search.NewRepository(store), topicRepo)),
		site:        site.NewHandler(siteSvc),
		siteSvc:     siteSvc,
		rbac:        rbacSvc,
		rbacAdmin:   rbac.NewAdminHandler(rbacSvc, lookupUser),
		moderation:  moderation.NewHandler(modSvc),
		audit:       audit.NewHandler(auditSvc),
		auditSvc:    auditSvc,
		reportLimit: middleware.NewRateLimiter(10, time.Minute),
		admin:       admin.NewHandler(admin.NewService(admin.NewRepository(store), cfg.Storage), fileStore),
		install:     install.NewHandler(installSvc),
		installSvc:  installSvc,
	}, nil
}

// Seed applies idempotent baseline data (default roles, permissions and
// site settings).
func (s *Server) Seed(ctx context.Context) error {
	if err := s.rbac.Seed(ctx); err != nil {
		return err
	}
	if err := s.siteSvc.Seed(ctx); err != nil {
		return err
	}
	if marked, err := s.installSvc.Reconcile(ctx); err != nil {
		return err
	} else if marked {
		s.logger.Info("existing database marked as installed")
	}
	n, err := s.normalizeMediaURLs(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		s.logger.Info("normalized stored image URLs to /files/…", "count", n)
	}
	return nil
}

// HTTPServer returns a configured *http.Server with sane timeouts.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              ":" + s.cfg.APIPort,
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// fileServer returns an HTTP handler for local file serving, if supported.
func (s *Server) fileServer() (http.Handler, bool) {
	fs, ok := s.storage.(interface{ FileServer() http.Handler })
	if !ok {
		return nil, false
	}
	return fs.FileServer(), true
}
