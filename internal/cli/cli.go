// Package cli implements the patrolbot CLI.
package cli

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/iocalebs/patrolbot/internal/cli/commands/config"
	initCmd "github.com/iocalebs/patrolbot/internal/cli/commands/init"
	"github.com/iocalebs/patrolbot/internal/cli/commands/report"
	"github.com/iocalebs/patrolbot/internal/cli/commands/serve"
	"github.com/spf13/cobra"
)

// Execute runs the CLI.
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "patrolbot",
		Short: "PatrolBot assists with patrolling MediaWiki sites.",
		Long:  "",
	}

	rootCmd.AddCommand(
		config.NewCommand(),
		initCmd.NewCommand(),
		report.NewCommand(),
		serve.NewCommand(),
	)
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
		fang.WithErrorHandler(errHandler),
	)
	if err != nil {
		os.Exit(1)
	}
}
