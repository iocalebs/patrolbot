// Package register implements the "patrolbot register" command
package register

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
	"github.com/spf13/cobra"
)

const discordTimeout = 10 * time.Second

// ErrRegisterCommands indicates an error response from the Discord API when registering commands.
var ErrRegisterCommands = errors.New("failed to register one or more Discord commands")

// NewCommand returns the `register` Cobra command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register Discord bot commands",
		Long: "Register the /report command in each configured Discord server with its corresponding report types " +
			"as subcommand options.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(cmd.Flags())
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			return registerCommands(cmd.Context(), cmd.ErrOrStderr(), cfg)
		},
	}

	config.AddFlag(cmd.Flags())

	return cmd
}

func registerCommands(ctx context.Context, w io.Writer, cfg config.Config) error {
	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: cfg.Log.Level,
	}))
	httpClient := &http.Client{
		Timeout: discordTimeout,
	}
	discordClient := discord.NewClient(httpClient, logger, cfg.Discord, cfg.UserAgent)

	hasErrors := false

	keys := maps.Keys(cfg.Wikis)
	wikis := slices.Collect(keys)
	slices.Sort(wikis) // iterate in deterministic order to simplify testing

	for _, wiki := range wikis {
		wikiCfg := cfg.Wikis[wiki]
		logger := logger.With("wiki", wiki)

		ok := registerReportCommand(ctx, wikiCfg, discordClient, logger)
		if !ok {
			hasErrors = true
		}
	}

	if hasErrors {
		return ErrRegisterCommands
	}

	return nil
}

func registerReportCommand(
	ctx context.Context,
	wiki config.Wiki,
	discordClient *discord.Client,
	logger *slog.Logger,
) bool {
	guild := wiki.Reports.DiscordServerID
	if guild == "" {
		return true
	}

	reportTypes := slices.Collect(maps.Keys(wiki.Reports.Types))
	slices.Sort(reportTypes)

	options := make([]discord.CommandOption, len(reportTypes))
	for i, reportType := range reportTypes {
		options[i] = discord.CommandOption{
			Type:        discord.CommandOptionTypeSubcommand,
			Name:        reportType,
			Description: wiki.Reports.Types[reportType].Description,
		}
	}

	logger = logger.With("guildID", guild)

	cmd := discord.Command{
		Name:        "report",
		Description: "Generate a report using data from the MediaWiki API",
		Options:     options,
	}

	err := discordClient.RegisterGuildCommand(ctx, guild, cmd)
	if err != nil {
		var httpStatusError *discord.HTTPStatusError
		if errors.As(err, &httpStatusError) {
			logger.ErrorContext(ctx, "error registering /report command", "err", err, "body", httpStatusError.Body)
		} else {
			logger.ErrorContext(ctx, "error registering /report command", "err", err)
		}

		return false
	}

	logger.InfoContext(ctx, "registered /report command", "options", strings.Join(reportTypes, ","))

	return true
}
