package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func reportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Send a status report to a Discord server",
		Long:  ``, // TODO
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("TODO") //nolint:forbidigo
		},
	}
}
