package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
)

const testChannel = "911393794262458421"
const testGuild = "911393793666863116"

type discordGenerator struct {
	client    *discord.Client
	config    config.Config
	transport *capturingTransport
}

func discordTestData(cfg config.Config, overwrite bool) error {
	transport := &capturingTransport{
		rt:        http.DefaultTransport,
		overwrite: overwrite,
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	gen := &discordGenerator{
		config:    cfg,
		client:    discord.NewClient(httpClient, slog.Default(), cfg.Discord, cfg.UserAgent),
		transport: transport,
	}

	generators := [](func(context.Context) error){
		gen.createMessage,
		gen.createMessageInvalid,
		gen.registerGuildCommand,
		gen.registerGuildCommandInvalid,
	}

	for _, generator := range generators {
		err := generator(context.TODO())

		var httpStatusError *discord.HTTPStatusError
		if errors.As(err, &httpStatusError) {
			return fmt.Errorf("%w, body: %s", httpStatusError, httpStatusError.Body)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (g *discordGenerator) writeCapture(outPath string) error {
	return g.transport.writeCapture(outPath)
}

func (g *discordGenerator) createMessage(ctx context.Context) error {
	_, err := g.client.CreateMessage(ctx, testChannel, discord.CreateMessageRequestBody{
		Content: "test",
	})
	if err != nil {
		return err
	}

	return g.writeCapture("testdata/createmessage.json")
}

func (g *discordGenerator) createMessageInvalid(ctx context.Context) error {
	_, err := g.client.CreateMessage(ctx, testChannel, discord.CreateMessageRequestBody{})

	var httpStatusError *discord.HTTPStatusError
	if errors.As(err, &httpStatusError) {
		if httpStatusError.StatusCode != http.StatusBadRequest {
			return err
		}
	} else if err != nil {
		return err
	}

	return g.writeCapture("testdata/createmessage_invalid.json")
}

func (g *discordGenerator) registerGuildCommand(ctx context.Context) error {
	cmd := discord.Command{
		Name:        "report",
		Description: "Generate a report",
		Options: []discord.CommandOption{
			{
				Type:        discord.CommandOptionTypeSubcommand,
				Description: "description",
				Name:        "expiring",
			},
		},
	}

	err := g.client.RegisterGuildCommand(ctx, testGuild, cmd)
	if err != nil {
		return err
	}

	return g.writeCapture("testdata/registerguildcommand.json")
}

func (g *discordGenerator) registerGuildCommandInvalid(ctx context.Context) error {
	err := g.client.RegisterGuildCommand(ctx, testGuild, discord.Command{})

	var httpStatusError *discord.HTTPStatusError
	if errors.As(err, &httpStatusError) {
		if httpStatusError.StatusCode != http.StatusBadRequest {
			return err
		}
	} else if err != nil {
		return err
	}

	return g.writeCapture("testdata/registerguildcommand_invalid.json")
}
