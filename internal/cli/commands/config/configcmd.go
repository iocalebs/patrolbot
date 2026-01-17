// Package config implements the "patrolbot config" command
package config

import (
	"github.com/iocalebs/patrolbot/internal/cli/commands/config/view"
	"github.com/spf13/cobra"
)

// NewCommand returns the `config` Cobra command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage PatrolBot configuration",
		Long: `Manage PatrolBot configuration.

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

	cmd.AddCommand(
		view.NewCommand(),
	)

	return cmd
}
