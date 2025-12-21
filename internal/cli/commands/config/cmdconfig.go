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
	cmd := &cobra.Command{}
	cmd.Use = "config"
	cmd.Short = "Manage bot configuration"
	cmd.Long = `Manage patrolbot configuration.
	
patrolbot sources config from the following places, in order of highest to lowest priority:
	1. Command-line flags (e.g. --wiki wikipedia)
	2. Environment variables (e.g. PATROLBOT_WIKI=wikipedia)
	3. config.yaml file indicated by the --config flag
	4. config.yaml file in current directory
	5. config.yaml file in $HOME/.patrolbot.`
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}

	cmd.AddCommand(view())

	return cmd
}

func view() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "view"
	cmd.Short = "View merged configuration"
	cmd.Long = `View JSON representation of the merged configuration struct used by the bot.`
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.FromContext(cmd.Context())
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		err = encoder.Encode(cfg)
		if err != nil {
			return fmt.Errorf("error JSON-encoding config: %w", err)
		}

		return nil
	}

	return cmd
}
