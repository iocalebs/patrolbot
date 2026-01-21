// Package register implements the "patrolbot register" command
package register

import (
	"context"
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

// NewCommand returns the `register` Cobra command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register Discord bot commands",
		Long: "Register the /report command in each configured Discord server with its corresponding report types " +
			"as choices.",
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

	for wiki, wikiCfg := range cfg.Wikis {
		guild := wikiCfg.Reports.DiscordServerID
		if guild == "" {
			continue
		}

		reportTypes := slices.Collect(maps.Keys(wikiCfg.Reports.Types))

		choices := make([]discord.CommandOptionChoice, len(reportTypes))
		for i, reportType := range reportTypes {
			choices[i] = discord.CommandOptionChoice{
				Name:  reportType,
				Value: reportType,
			}
		}

		if len(choices) == 0 {
			continue
		}

		cmd := discord.Command{
			Name:        "report",
			Description: "Generate a report using data from the MediaWiki API",
			Options: []discord.CommandOption{
				{
					Type:        discord.CommandOptionTypeString,
					Name:        "type",
					Description: "Report type",
					Required:    true,
					Choices:     choices,
				},
			},
		}

		err := discordClient.RegisterGuildCommand(ctx, guild, cmd)
		if err != nil {
			return fmt.Errorf("error registering /report command for wiki %s server %s: %w", wiki, guild, err)
		}

		logger.InfoContext(
			ctx,
			"registered /report command", "wiki",
			wiki,
			"guildID",
			guild,
			"choices",
			strings.Join(reportTypes, ","),
		)
	}

	return nil
}
