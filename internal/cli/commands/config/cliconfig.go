// Package config implements the config command
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/cobra"
)

// NewCommand returns the `config` subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage bot configuration",
		Long: `Manage patrolbot configuration.
Configuration sources (from highest to lowest priority):

  1. Command-line flags
     Example: --wiki wikipedia

  2. Environment variables
     Example: PATROLBOT_WIKI=wikipedia

  3. Configuration file (YAML)
     Only one file is loaded, chosen from:
       a. Path specified by --config
       b. ./config.yaml
       c. $HOME/.patrolbot/config.yaml`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(view())

	return cmd
}

func view() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View merged configuration",
		Long: "View JSON representation of the final configuration used by PatrolBot " +
			"after resolving command-line flags, environment variables, and config files.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			err = encoder.Encode(cfg)
			if err != nil {
				return fmt.Errorf("error JSON-encoding config: %w", err)
			}

			return nil
		},
	}

	return cmd
}
