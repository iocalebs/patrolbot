// Package report implements the `report` Cobra command
package report

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/iocalebs/patrolbot/internal/cli/clierr"
	"github.com/iocalebs/patrolbot/internal/cli/flags"
	"github.com/iocalebs/patrolbot/internal/clock"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	reporter "github.com/iocalebs/patrolbot/internal/report"
	"github.com/iocalebs/patrolbot/internal/report/provider"
	"github.com/spf13/cobra"
)

const discordTimeout = 10 * time.Second

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

			sendDiscord, err := cmd.Flags().GetBool("send-discord")
			if err != nil {
				return fmt.Errorf("error checking --send-discord flag: %w", err)
			}

			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			logger := slog.Default()

			reporter, err := newReporter(cfg, logger)
			if err != nil {
				return err
			}

			if list {
				return reporter.ListReports(os.Stdout)
			}

			if len(args) == 0 {
				return clierr.UsageError("missing report type")
			}

			discordClient := newDiscordClient(cfg, logger)

			return report(cmd, reporter, args[0], discordClient, sendDiscord)
		},
	}

	cmd.Flags().Bool("list", false, "List available report types")
	cmd.Flags().Bool(
		"send-discord",
		false,
		"Send report to configured Discord channel without prompting for confirmation first",
	)
	flags.Config(cmd)

	return cmd
}

func newReporter(cfg config.Config, logger *slog.Logger) (*reporter.Reporter, error) {
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

	clock := clock.Func(func() time.Time {
		mockTime := os.Getenv("MOCK_TIME")
		if mockTime != "" {
			now, err := time.Parse(time.RFC3339, mockTime)
			if err == nil {
				return now
			}

			logger.Error("Error parsing $MOCK_TIME %q: %v", mockTime, now)
		}

		return time.Now()
	})

	mwclient := mediawiki.NewClient(wiki, &httpClient, logger, cfg.UserAgent)
	provider := provider.New(logger, clock, mwclient, wiki)
	reporter := reporter.New(wiki.Reports, provider)

	return reporter, nil
}

func newDiscordClient(cfg config.Config, logger *slog.Logger) *discord.Client {
	httpClient := &http.Client{
		Timeout: discordTimeout,
	}
	discordClient := discord.NewClient(httpClient, logger, cfg.Discord, cfg.UserAgent)

	return discordClient
}

func report(
	cmd *cobra.Command,
	reporter *reporter.Reporter,
	reportType string,
	discordClient *discord.Client,
	sendDiscord bool,
) error {
	var buf bytes.Buffer

	reportConfig, err := reporter.Report(cmd.Context(), reportType, &buf)
	if err != nil {
		return fmt.Errorf("error generating report: %w", err)
	}

	_, err = cmd.OutOrStdout().Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error writing report to stdout: %w", err)
	}

	if !sendDiscord {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Post to Discord?").
					Value(&sendDiscord),
			),
		).WithAccessible(true)

		err = form.RunWithContext(cmd.Context())
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
	}

	if !sendDiscord {
		return nil
	}

	req := discord.CreateMessageRequestBody{
		Content: buf.String(),
	}

	err = discordClient.CreateMessage(cmd.Context(), reportConfig.ChannelID, req)
	if err != nil {
		return fmt.Errorf("error posting report to Discord: %w", err)
	}

	_, err = cmd.ErrOrStderr().Write([]byte("Sent Discord message\n")) // TODO: message link
	if err != nil {
		return fmt.Errorf("error writing to stderr: %w", err)
	}

	return nil
}
