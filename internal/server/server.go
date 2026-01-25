// Package server implements the HTTP server for handling Discord interactions via webhooks.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/errdefs"
)

// Server represents the HTTP server.
type Server struct {
	logger *slog.Logger
	server *http.Server
}

// NewServer creates a new [Server] instance.
func NewServer(cfg config.Config, logger *slog.Logger, middleware ...func(http.Handler) http.Handler) (*Server, error) {
	err := validateServerConfig(cfg)
	if err != nil {
		return nil, err
	}

	handler, err := mux(cfg)
	if err != nil {
		return nil, err
	}

	for _, middlewareFunc := range middleware {
		handler = middlewareFunc(handler)
	}

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Server.Port),
		ReadTimeout:       cfg.Server.Timeouts.ReadTimeout,
		ReadHeaderTimeout: cfg.Server.Timeouts.ReadHeaderTimeout,
		WriteTimeout:      cfg.Server.Timeouts.WriteTimeout,
		IdleTimeout:       cfg.Server.Timeouts.IdleTimeout,
		Handler:           handler,
	}

	return &Server{
		logger: logger,
		server: srv,
	}, nil
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

func validateServerConfig(cfg config.Config) error {
	if cfg.Server == nil {
		return errdefs.NewConfigError(".server not set")
	}

	errs := []string{}

	if cfg.Server.Port == 0 {
		errs = append(errs, ".server.port not set")
	}

	if cfg.Server.Timeouts.ReadHeaderTimeout <= 0 {
		errs = append(
			errs,
			".server.timeouts.readHeaderTimeout must be greater than zero to guard against slowloris attacks",
		)
	}

	if len(errs) > 0 {
		return errdefs.NewConfigError(errs...)
	}

	return nil
}
