// Package serve implements the "patrolbot serve" command
package serve

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/gcp"
	"github.com/iocalebs/patrolbot/internal/server"
	"github.com/spf13/cobra"
)

// NewCommand returns the `serve` Cobra subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a Discord webhook server",
		Long:  "Start an HTTP server to receive and respond to Discord Interactions via webhooks.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			srv, err := newServer(cfg)
			if err != nil {
				return err
			}

			return srv.Start(cmd.Context())
		},
	}

	config.AddFlag(cmd.Flags())

	return cmd
}

func newServer(cfg config.Config) (*server.Server, error) {
	logger := gcp.NewLogger(os.Stderr, cfg.Log.Level)

	srv, err := server.NewServer(cfg, logger, gcp.TraceMiddleware, logRequestCancellation(logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	return srv, nil
}

func logRequestCancellation(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			done := make(chan struct{})

			go func() {
				select {
				case <-ctx.Done():
					logger.ErrorContext(r.Context(), "request cancelled")
				case <-done:
				}
			}()

			next.ServeHTTP(w, r)
			close(done)
		})
	}
}
