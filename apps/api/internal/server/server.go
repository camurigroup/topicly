package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/forum-framework/api/internal/config"
	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/modules/auth"
	"github.com/forum-framework/api/internal/modules/health"
	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/modules/user"
)

// Server wires configuration, database and the module handlers.
type Server struct {
	cfg    config.Config
	db     *sql.DB
	logger *slog.Logger

	health *health.Handler
	auth   *auth.Handler
	authMW *auth.Middleware
	user   *user.Handler
	rbac   *rbac.Service
}

// New builds the Server and all module dependencies for the given dialect.
func New(cfg config.Config, db *sql.DB, dialect string, logger *slog.Logger) *Server {
	store := database.NewStore(db, dialect)

	userRepo := user.NewRepository(store)
	rbacRepo := rbac.NewRepository(store)
	tokenRepo := auth.NewRefreshTokenRepository(store)

	rbacSvc := rbac.NewService(rbacRepo)
	userSvc := user.NewService(userRepo)
	authSvc := auth.NewService(userRepo, tokenRepo, rbacSvc, auth.Config{
		AccessSecret: cfg.JWTAccessSecret,
		AccessTTL:    cfg.JWTAccessTTL,
		RefreshTTL:   cfg.JWTRefreshTTL,
	})

	return &Server{
		cfg:    cfg,
		db:     db,
		logger: logger,
		health: health.New(db),
		auth:   auth.NewHandler(authSvc),
		authMW: auth.NewMiddleware(cfg.JWTAccessSecret, rbacSvc),
		user:   user.NewHandler(userSvc, rbacSvc),
		rbac:   rbacSvc,
	}
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
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
