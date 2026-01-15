package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CreateMessageRequestBody represents the JSON request body used in a [Client.CreateMessage] request.
type CreateMessageRequestBody struct {
	Content string `json:"content,omitempty"` // Message contents (up to 2000 characters)
}

// A Message represents a Discord message sent in a channel within Discord.
// See: https://discord.com/developers/docs/resources/message#messages-resource
type Message struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
}

// CreateMessage posts a message to a Discord channel.
// See: https://discord.com/developers/docs/resources/message#create-message
func (c *Client) CreateMessage(
	ctx context.Context,
	channelID string,
	reqBody CreateMessageRequestBody,
) (Message, error) {
	reqURL, err := url.JoinPath(c.config.API, "channels", channelID, "messages")
	if err != nil {
		return Message{}, fmt.Errorf("failed to create request URL: %w", err)
	}

	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, reqURL, &buf)
	if err != nil {
		return Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(ctx, req)
	if err != nil {
		return Message{}, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			c.logger.WarnContext(ctx, "Failed to close Discord API response body", "error", err)
		}
	}()

	var message Message

	err = json.NewDecoder(resp.Body).Decode(&message)
	if err != nil {
		return Message{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return message, nil
}
