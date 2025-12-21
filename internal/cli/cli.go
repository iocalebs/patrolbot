// Package cli implements the patrolbot CLI.
package cli

import (
	"fmt"
	"os"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/cobra"
)

// Execute runs the CLI.
func Execute() {
	rootCmd := &cobra.Command{
		Use:          "patrolbot",
		Short:        "patrolbot assists with patrolling tasks on MediaWiki sites.",
		Long:         "",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return initConfig(cmd)
		},
	}

	initFlags(rootCmd)
	addCommands(rootCmd)
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringP(
		"config",
		"c",
		"",
		"Path to config file (default locations: ./config.yaml, $HOME/.patrolbot/config.yaml)",
	)
	rootCmd.PersistentFlags().StringP(
		"wiki",
		"w",
		"",
		"Target wiki for bot operations",
	)
}

func addCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		reportCmd(),
	)
}

func initConfig(cmd *cobra.Command) error {
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	cfg, err := config.Load(cfgFile, cmd.Flags())
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	ctx := config.NewContext(cmd.Context(), cfg)
	cmd.SetContext(ctx)

	return nil
}
