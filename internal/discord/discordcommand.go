package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CommandOptionType represents an  Application Command option type.
type CommandOptionType uint16

// Application Command option types
// https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-structure
const (
	CommandOptionTypeSubcommand = iota + 1
	CommandOptionTypeSubcommandGroup
	CommandOptionTypeString
)

// Command represents an Application Command.
// https://discord.com/developers/docs/interactions/application-commands#application-command-object
type Command struct {
	// 1-32 character name
	// https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-naming
	Name string `json:"name"`

	// Description for CHAT_INPUT commands, 1-100 characters. Empty string for USER and MESSAGE commands
	Description string `json:"description"`

	// Parameters for the command, max of 25
	Options []CommandOption `json:"options"`
}

// CommandOption represents an option for an Application Command.
// https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-structure
type CommandOption struct {
	// Type of option
	Type CommandOptionType `json:"type"`

	// 1-32 character name
	// https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-naming
	Name string `json:"name"`

	// 1-100 character description
	Description string `json:"description"`

	// Whether the parameter is required or optional
	Required bool `json:"required"`

	// Choices for the user to pick from, max 25
	Choices []CommandOptionChoice `json:"choices"`
}

// CommandOptionChoice represents a choice for an Application Command option.
// https://discord.com/developers/docs/interactions/application-commands#application-command-object-application-command-option-choice-structure
type CommandOptionChoice struct {
	// 1-100 character choice name
	Name string `json:"name"`

	// Value for the choice, up to 100 characters
	Value string `json:"value"`
}

// RegisterGuildCommand registers a command within the specified Discord guild (server).
func (c *Client) RegisterGuildCommand(ctx context.Context, guild string, cmd Command) error {
	reqURL, err := url.JoinPath(c.config.API, "applications", c.config.AppID, "guilds", guild, "commands")
	if err != nil {
		return fmt.Errorf("failed to create request URL: %w", err)
	}

	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(cmd)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, reqURL, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(ctx, req)
	if err != nil {
		return err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			c.logger.WarnContext(ctx, "Failed to close Discord API response body", "error", err)
		}
	}()

	return nil
}
