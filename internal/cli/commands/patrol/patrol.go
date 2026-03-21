// Package patrol implements the "patrolbot patrol" command
package patrol

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"

	"github.com/charmbracelet/lipgloss"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	"github.com/iocalebs/patrolbot/internal/patroller"
	"github.com/spf13/cobra"
)

// NewCommand returns the `patrol` Cobra command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patrol",
		Short: "Mark edits as patrolled",
		Long:  "Iterate over unpatrolled revisions in Special:RecentChanges confirm whether to mark them as patrolled.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			latest, err := cmd.Flags().GetBool("latest")
			if err != nil {
				return fmt.Errorf("error parsing flags: %w", err)
			}

			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return fmt.Errorf("error parsing flags: %w", err)
			}

			user, err := cmd.Flags().GetString("user")
			if err != nil {
				return fmt.Errorf("error parsing flags: %w", err)
			}

			opts := patroller.Opts{
				Latest:    latest,
				Namespace: namespace,
				User:      user,
			}

			patroller, err := newPatroller(cmd, cfg, opts)
			if err != nil {
				return err
			}

			return patroller.run()
		},
	}

	config.AddFlag(cmd.Flags())
	cmd.Flags().Bool(
		"latest",
		false,
		"Patrol newest revisions first (default: oldest first)",
	)
	cmd.Flags().String(
		"namespace",
		"",
		"Namespace to patrol (default: all namespaces)",
	)
	cmd.Flags().StringP(
		"user",
		"u",
		"",
		"User to patrol (default: all users)",
	)
	cmd.Flags().StringP(
		"wiki",
		"w",
		"",
		"Target wiki",
	)

	return cmd
}

func newPatroller(cmd *cobra.Command, cfg config.Config, opts patroller.Opts) (*CLIPatroller, error) {
	wiki, err := cfg.CurrentWiki()
	if err != nil {
		return nil, fmt.Errorf("invalid wiki configuration: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating MediaWiki HTTP client: %w", err)
	}

	httpClient := &http.Client{
		Jar:     jar,
		Timeout: wiki.Client.Timeout,
	}
	mwclient := mediawiki.NewClient(wiki, httpClient, slog.Default(), cfg.UserAgent)

	patroller, err := patroller.New(mwclient, opts)
	if err != nil {
		return nil, fmt.Errorf("error initializing patroller: %w", err)
	}

	cliPatroller := newCLIPatroller(patroller, cmd, theme())

	return cliPatroller, nil
}

func theme() Theme {
	var (
		cyan    = lipgloss.Color("6")
		green   = lipgloss.Color("2")
		indigo  = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
		magenta = lipgloss.Color("5")
		red     = lipgloss.Color("1")
	)

	return Theme{
		error:       lipgloss.NewStyle().Foreground(red).Bold(true),
		hunk:        lipgloss.NewStyle().Foreground(cyan),
		lineAdded:   lipgloss.NewStyle().Foreground(green),
		lineRemoved: lipgloss.NewStyle().Foreground(red),
		prompt:      lipgloss.NewStyle().Foreground(indigo).Bold(true),
		title:       lipgloss.NewStyle().Foreground(magenta),
	}
}
