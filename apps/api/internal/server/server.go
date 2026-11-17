package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/forum-framework/api/internal/config"
)

// Server wires configuration, database and logger into the HTTP layer.
type Server struct {
	cfg    config.Config
	db     *sql.DB
	logger *slog.Logger
}

// New builds a Server.
func New(cfg config.Config, db *sql.DB, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, db: db, logger: logger}
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
