// Package serve implements the "patrolbot serve" command
package serve

import (
	"fmt"
	"os"

	"github.com/iocalebs/patrolbot/internal/cli/clierr"
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
	err := validateServerConfig(cfg.Server)
	if err != nil {
		return nil, err
	}

	logger := gcp.NewLogger(os.Stderr, cfg.Log.Level)

	return server.NewServer(*cfg.Server, logger, gcp.TraceMiddleware), nil
}

func validateServerConfig(cfg *config.Server) error {
	if cfg == nil {
		return clierr.NewMultiError("invalid configuration", []string{".server not set"})
	}

	errs := []string{}

	if cfg.Port == 0 {
		errs = append(errs, ".server.port not set")
	}

	if cfg.Timeouts.ReadHeaderTimeout <= 0 {
		errs = append(
			errs,
			".server.timeouts.readHeaderTimeout must be greater than zero to guard against slowloris attacks",
		)
	}

	if len(errs) > 0 {
		return clierr.NewMultiError("invalid configuration", errs)
	}

	return nil
}
