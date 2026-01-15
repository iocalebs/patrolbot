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

func discordTestData(cfg config.Config, overwrite bool) error {
	client, capturer := setupDiscord(cfg, overwrite)
	generators := [](func(context.Context, *discord.Client, *capturingTransport) error){
		createMessage,
		createMessageInvalid,
	}

	for _, generator := range generators {
		err := generator(context.TODO(), client, capturer)

		var httpStatusError *discord.HTTPStatusError
		if errors.As(err, &httpStatusError) {
			return fmt.Errorf("%w, body: %s", httpStatusError, httpStatusError.Body)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func setupDiscord(cfg config.Config, overwrite bool) (*discord.Client, *capturingTransport) {
	transport := &capturingTransport{
		rt:        http.DefaultTransport,
		overwrite: overwrite,
	}
	client := &http.Client{
		Transport: transport,
	}
	discordClient := discord.NewClient(client, slog.Default(), cfg.Discord, cfg.UserAgent)

	return discordClient, transport
}

func createMessage(ctx context.Context, client *discord.Client, transport *capturingTransport) error {
	_, err := client.CreateMessage(ctx, testChannel, discord.CreateMessageRequestBody{
		Content: "test",
	})
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/createmessage.json")
}

func createMessageInvalid(ctx context.Context, client *discord.Client, transport *capturingTransport) error {
	_, err := client.CreateMessage(ctx, testChannel, discord.CreateMessageRequestBody{})

	var httpStatusError *discord.HTTPStatusError
	if errors.As(err, &httpStatusError) {
		if httpStatusError.StatusCode != http.StatusBadRequest {
			return err
		}
	} else if err != nil {
		return err
	}

	return transport.writeCapture("testdata/createmessage_invalid.json")
}
