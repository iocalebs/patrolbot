// Package view implements the "patrolbot config view" command
package view

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/cobra"
)

// NewCommand returns the `config view` Cobra command.
func NewCommand() *cobra.Command {
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

	config.AddFlag(cmd.Flags())

	return cmd
}
