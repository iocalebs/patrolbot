// Package report implements the "patrolbot report" command
package report

import (
	"fmt"
	"log/slog"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/cobra"
)

// NewCommand returns the `report` Cobra command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <type>",
		Short: "Generate a report",
		Long:  "Generate a report using data from MediaWiki Action API queries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := cmd.Flags().GetBool("list")
			if err != nil {
				return fmt.Errorf("error checking --list flag: %w", err)
			}

			sendDiscord, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return fmt.Errorf("error checking --yes flag: %w", err)
			}

			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			handler := &handler{
				config: cfg,
				logger: slog.Default(),
				stdout: cmd.OutOrStdout(),
				stderr: cmd.ErrOrStderr(),
			}

			return handler.run(cmd.Context(), args, opts{
				list:        list,
				sendDiscord: sendDiscord,
			})
		},
	}

	config.AddFlag(cmd.Flags())
	cmd.Flags().Bool("list", false, "List available report types")
	cmd.Flags().StringP(
		"wiki",
		"w",
		"",
		"Target wiki",
	)
	cmd.Flags().BoolP(
		"yes",
		"y",
		false,
		"Automatically answer \"yes\" to any prompts",
	)

	return cmd
}
