// Package flags provides flags that are common to many (but not all) commands
package flags

import "github.com/spf13/cobra"

// Config applies config-related flags for commands that load config.
// It provides the --config flag for passing a custom config file, and all the flags for overriding specific config
// values (e.g. --wiki).
func Config(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"config",
		"c",
		"",
		"Path to config file (default locations: ./config.yaml, $HOME/.patrolbot/config.yaml)",
	)
	cmd.Flags().StringP(
		"wiki",
		"w",
		"",
		"Target wiki",
	)
}
