package server

import (
	"net/http"

	"github.com/forum-framework/api/internal/middleware"
	"github.com/forum-framework/api/internal/modules/health"
)

// routes builds the HTTP handler with all routes and the middleware chain.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	healthHandler := health.New(s.db)
	mux.HandleFunc("GET /health", healthHandler.Live)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// Versioned API group: /api/v1/*
	api := http.NewServeMux()
	api.HandleFunc("GET /ping", healthHandler.Ping)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return middleware.Chain(mux,
		middleware.RequestID,
		middleware.Logger(s.logger),
		middleware.Recover(s.logger),
		middleware.CORS(s.cfg.CORSAllowedOrigins),
	)
}
