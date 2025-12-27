// Package cli implements the patrolbot CLI.
package cli

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	configCmd "github.com/iocalebs/patrolbot/internal/cli/commands/config"
	"github.com/spf13/cobra"
)

// Execute runs the CLI.
func Execute() {
	rootCmd := &cobra.Command{
		Use:          "patrolbot",
		Short:        "PatrolBot assists with patrolling MediaWiki sites.",
		Long:         "",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		configCmd.NewCommand(),
	)
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	err := fang.Execute(context.Background(), rootCmd, fang.WithColorSchemeFunc(fang.AnsiColorScheme))
	if err != nil {
		os.Exit(1)
	}
}
