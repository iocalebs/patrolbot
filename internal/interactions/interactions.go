// Package interactions provides the functionality for handling Discord interactions.
// See: https://discord.com/developers/docs/interactions/receiving-and-responding
package interactions

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"

	"github.com/iocalebs/patrolbot/internal/clock"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	"github.com/iocalebs/patrolbot/internal/report/provider"
	"github.com/iocalebs/patrolbot/internal/report/reporter"
)

// A Handler handles Discord interactions.
type Handler struct {
	logger        *slog.Logger
	invalidConfig map[string]error
	reporters     map[string]*reporter.Reporter
}

// NewHandler creates a new [Handler] instance.
func NewHandler(cfg config.Config, logger *slog.Logger) (*Handler, error) {
	invalidConfig := map[string]error{}
	reporters := map[string]*reporter.Reporter{}

	for wiki, wikiCfg := range cfg.Wikis {
		cfg.Wiki = wiki

		guildID := wikiCfg.Reports.DiscordServerID
		if guildID == "" {
			continue
		}

		_, err := cfg.CurrentWiki()
		if err != nil {
			invalidConfig[guildID] = err
			continue
		}

		reporter, err := newReporter(cfg, wikiCfg, logger)
		if err != nil {
			return nil, err
		}

		reporters[guildID] = reporter
	}

	return &Handler{
		logger:        logger,
		invalidConfig: invalidConfig,
		reporters:     reporters,
	}, nil
}

// Respond handles an interaction and returns an interaction response.
func (h *Handler) Respond(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	response, err := h.respond(ctx, interaction)
	if err != "" {
		h.logger.ErrorContext(ctx, err)

		return discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: "Error: " + err,
			},
		}
	}

	return response
}

func (h *Handler) respond(ctx context.Context, interaction discord.Interaction) (discord.InteractionResponse, string) {
	switch interaction.Type {
	case discord.InteractionTypePing:
		return discord.InteractionResponse{
			Type: discord.InteractionCallbackTypePong,
		}, ""
	case discord.InteractionTypeApplicationCommand:
		command := interaction.Data.Name
		if command == "report" {
			return h.report(ctx, interaction)
		}

		return discord.InteractionResponse{}, fmt.Sprintf("unsupported command `%s`", command)
	}

	return discord.InteractionResponse{}, fmt.Sprintf("unsupported interaction type %d", interaction.Type)
}

func (h *Handler) report(ctx context.Context, interaction discord.Interaction) (discord.InteractionResponse, string) {
	reporter, ok := h.reporters[interaction.Data.GuildID]
	if !ok {
		err, ok := h.invalidConfig[interaction.Data.GuildID]
		if !ok {
			return discord.InteractionResponse{}, "PatrolBot is not configured for this server"
		}

		return discord.InteractionResponse{}, err.Error()
	}

	if len(interaction.Data.Options) < 1 {
		return discord.InteractionResponse{}, "no report type provided"
	}

	var buffer bytes.Buffer

	reportType := interaction.Data.Options[0].Name

	_, err := reporter.Report(ctx, reportType, &buffer)
	if err != nil {
		return discord.InteractionResponse{}, err.Error()
	}

	return discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionResponseData{
			Content: buffer.String(),
		},
	}, ""
}

func newReporter(cfg config.Config, wiki config.Wiki, logger *slog.Logger) (*reporter.Reporter, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating MediaWiki HTTP client: %w", err)
	}

	httpClient := http.Client{
		Jar:     jar,
		Timeout: wiki.Client.Timeout,
	}

	mwclient := mediawiki.NewClient(wiki, &httpClient, logger, cfg.UserAgent)
	provider := provider.New(logger, clock.Mock(logger), mwclient, wiki)
	reporter := reporter.New(wiki.Reports, provider)

	return reporter, nil
}
