// Package server implements the HTTP server for handling Discord interactions via webhooks.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/iocalebs/patrolbot/internal/config"
)

// Server represents the HTTP server.
type Server struct {
	logger *slog.Logger
	server *http.Server
}

// NewServer creates a new [Server] instance.
func NewServer(cfg config.Server, logger *slog.Logger, middleware ...func(http.Handler) http.Handler) *Server {
	handler := newMux()
	for _, middlewareFunc := range middleware {
		handler = middlewareFunc(handler)
	}

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		ReadTimeout:       cfg.Timeouts.ReadTimeout,
		ReadHeaderTimeout: cfg.Timeouts.ReadHeaderTimeout,
		WriteTimeout:      cfg.Timeouts.WriteTimeout,
		IdleTimeout:       cfg.Timeouts.IdleTimeout,
		Handler:           handler,
	}

	return &Server{
		logger: logger,
		server: srv,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Starting HTTP server", "addr", s.server.Addr)

	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}
