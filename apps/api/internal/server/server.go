package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/forum-framework/api/internal/config"
	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/middleware"
	"github.com/forum-framework/api/internal/modules/auth"
	"github.com/forum-framework/api/internal/modules/category"
	"github.com/forum-framework/api/internal/modules/health"
	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/modules/topic"
	"github.com/forum-framework/api/internal/modules/user"
	"github.com/forum-framework/api/internal/storage"
)

// Server wires configuration, database, storage and the module handlers.
type Server struct {
	cfg     config.Config
	db      *sql.DB
	logger  *slog.Logger
	storage storage.Provider

	health      *health.Handler
	auth        *auth.Handler
	authMW      *auth.Middleware
	authLimiter *middleware.RateLimiter
	user        *user.Handler
	category    *category.Handler
	topic       *topic.Handler
	rbac        *rbac.Service
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

	rbacSvc := rbac.NewService(rbacRepo)
	userSvc := user.NewService(userRepo)
	categorySvc := category.NewService(categoryRepo)
	topicSvc := topic.NewService(topicRepo, rbacSvc)
	authSvc := auth.NewService(userRepo, tokenRepo, rbacSvc, auth.Config{
		AccessSecret: cfg.JWTAccessSecret,
		AccessTTL:    cfg.JWTAccessTTL,
		RefreshTTL:   cfg.JWTRefreshTTL,
	})

	return &Server{
		cfg:         cfg,
		db:          db,
		logger:      logger,
		storage:     fileStore,
		health:      health.New(db),
		auth:        auth.NewHandler(authSvc),
		authMW:      auth.NewMiddleware(cfg.JWTAccessSecret, rbacSvc),
		authLimiter: middleware.NewRateLimiter(20, time.Minute),
		user:        user.NewHandler(userSvc, rbacSvc, fileStore),
		category:    category.NewHandler(categorySvc),
		topic:       topic.NewHandler(topicSvc),
		rbac:        rbacSvc,
	}, nil
}

// Seed applies idempotent baseline data (default roles and permissions).
func (s *Server) Seed(ctx context.Context) error {
	return s.rbac.Seed(ctx)
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
