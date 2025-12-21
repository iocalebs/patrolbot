package cli

import (
	"fmt"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/cobra"
)

func reportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Send a status report to a Discord server",
		Long:  ``, // TODO
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.FromContext(cmd.Context())
			cobra.CheckErr(err)

			fmt.Println(cfg) //nolint:forbidigo

			return nil
		},
	}
}
