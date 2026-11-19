package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forum-framework/api/internal/config"
	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/server"
	"github.com/forum-framework/api/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	if err := config.LoadDotEnv(envFile); err != nil {
		logger.Error("could not read env file", "file", envFile, "error", err)
		os.Exit(1)
	}
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	dialect, err := database.Dialect(cfg.DB.Driver)
	if err != nil {
		logger.Error("invalid database driver", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.DB)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	srv, err := server.New(cfg, db, dialect, logger)
	if err != nil {
		logger.Error("server init failed", "error", err)
		os.Exit(1)
	}

	// Verify connectivity, run migrations and seed baseline data. If the
	// database is unreachable the server still boots so /health works and
	// /ready reports the outage.
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	pingErr := db.PingContext(pingCtx)
	cancel()
	if pingErr != nil {
		logger.Warn("database not reachable yet; migrations and seed skipped", "error", pingErr)
	} else {
		n, err := database.Migrate(db, dialect, migrations.FS, dialect)
		if err != nil {
			logger.Error("migration failed", "error", err)
			os.Exit(1)
		}
		logger.Info("migrations applied", "count", n, "dialect", dialect)

		seedCtx, seedCancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := srv.Seed(seedCtx); err != nil {
			seedCancel()
			logger.Error("seed failed", "error", err)
			os.Exit(1)
		}
		seedCancel()
		logger.Info("baseline data seeded")
	}

	httpServer := srv.HTTPServer()

	go func() {
		logger.Info("api listening", "addr", httpServer.Addr, "env", cfg.AppEnv)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("stopped")
}
