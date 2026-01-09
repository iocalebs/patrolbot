// Package report implements the `report` Cobra command
package report

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/iocalebs/patrolbot/internal/cli/clierr"
	"github.com/iocalebs/patrolbot/internal/cli/flags"
	"github.com/iocalebs/patrolbot/internal/clock"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	reporter "github.com/iocalebs/patrolbot/internal/report"
	"github.com/iocalebs/patrolbot/internal/report/provider"
	"github.com/spf13/cobra"
)

// NewCommand returns the `report` Cobra command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <type>",
		Short: "Generate a report",
		Long:  "", // TODO - documentation
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := cmd.Flags().GetBool("list")
			if err != nil {
				return fmt.Errorf("error checking --list flag: %w", err)
			}

			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			reporter, err := newReporter(cfg)
			if err != nil {
				return err
			}

			if list {
				return reporter.ListReports(os.Stdout)
			}

			if len(args) == 0 {
				return clierr.UsageError("missing report type")
			}

			err = reporter.Report(cmd.Context(), args[0], os.Stdout)
			if err != nil {
				return fmt.Errorf("error generating report: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("list", false, "List available report types")
	flags.Config(cmd)

	return cmd
}

func newReporter(cfg config.Config) (*reporter.Reporter, error) {
	wiki, err := cfg.CurrentWiki()
	if err != nil {
		return nil, fmt.Errorf("invalid wiki configuration: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating MediaWiki HTTP client: %w", err)
	}

	httpClient := http.Client{
		Jar:     jar,
		Timeout: wiki.Client.Timeout,
	}
	logger := slog.Default()
	mwclient := mediawiki.NewClient(wiki, &httpClient, logger, cfg.UserAgent)
	provider := provider.New(logger, clock.Func(time.Now), mwclient, wiki)
	reporter := reporter.New(wiki.Reports, provider)

	return reporter, nil
}
