// Package cli implements the patrolbot CLI.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the CLI.
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "patrolbot",
		Short: "patrolbot assists with patrolling tasks on MediaWiki sites.",
		Long:  "",
	}

	rootCmd.AddCommand(reportCmd())

	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
