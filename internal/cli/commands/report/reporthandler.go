package report

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"os"
	"slices"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/iocalebs/patrolbot/internal/clock"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
	"github.com/iocalebs/patrolbot/internal/errdefs"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	"github.com/iocalebs/patrolbot/internal/report/provider"
	"github.com/iocalebs/patrolbot/internal/report/reporter"
)

type handler struct {
	config config.Config
	logger *slog.Logger
	stdout io.Writer
	stderr io.Writer
}

type opts struct {
	list        bool
	sendDiscord bool
}

const discordTimeout = 10 * time.Second

func (h *handler) run(ctx context.Context, args []string, opts opts) error {
	wiki, err := h.config.CurrentWiki()
	if err != nil {
		return fmt.Errorf("invalid wiki configuration: %w", err)
	}

	if opts.list {
		return h.listReports(wiki)
	}

	reporter, err := h.newReporter(wiki)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return errdefs.UsageError("missing report type")
	}

	var buf bytes.Buffer

	reportConfig, err := reporter.Report(ctx, args[0], &buf)
	if err != nil {
		return fmt.Errorf("error generating report: %w", err)
	}

	_, err = h.stdout.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error writing report to stdout: %w", err)
	}

	if !opts.sendDiscord {
		postToDiscord, err := confirmPostToDiscord(ctx)
		if err != nil {
			return err
		}

		if !postToDiscord {
			return nil
		}
	}

	return h.postToDiscord(ctx, reportConfig.DiscordChannelID, wiki.Reports.DiscordServerID, buf.String())
}

func (h *handler) listReports(wiki config.Wiki) error {
	keys := slices.Collect(maps.Keys(wiki.Reports.Types))
	slices.Sort(keys)

	for _, key := range keys {
		_, err := h.stdout.Write([]byte(key + "\n"))
		if err != nil {
			return fmt.Errorf("error listing reports: %w", err)
		}
	}

	return nil
}

func (h *handler) newReporter(wiki config.Wiki) (*reporter.Reporter, error) {
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

			h.logger.Error("Error parsing $MOCK_TIME %q: %v", mockTime, now)
		}

		return time.Now()
	})

	mwclient := mediawiki.NewClient(wiki, &httpClient, h.logger, h.config.UserAgent)
	provider := provider.New(h.logger, clock, mwclient, wiki)
	reporter := reporter.New(wiki.Reports, provider)

	return reporter, nil
}

func (h *handler) postToDiscord(
	ctx context.Context,
	discordChannelID string,
	discordServerID string,
	content string,
) error {
	httpClient := &http.Client{
		Timeout: discordTimeout,
	}
	discordClient := discord.NewClient(httpClient, h.logger, h.config.Discord, h.config.UserAgent)

	req := discord.CreateMessageRequestBody{
		Content: content,
	}

	msg, err := discordClient.CreateMessage(ctx, discordChannelID, req)
	if err != nil {
		return fmt.Errorf("error posting report to Discord: %w", err)
	}

	msgLink := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", discordServerID, msg.ChannelID, msg.ID)

	_, err = h.stderr.Write([]byte("Sent Discord message: " + msgLink + "\n"))
	if err != nil {
		return fmt.Errorf("error writing to stderr: %w", err)
	}

	return nil
}

func confirmPostToDiscord(ctx context.Context) (bool, error) {
	var postToDiscord bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Post to Discord?").
				Value(&postToDiscord),
		),
	).WithAccessible(true)

	err := form.RunWithContext(ctx)
	if err != nil {
		return false, fmt.Errorf("prompt failed: %w", err)
	}

	return postToDiscord, nil
}
