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

// CreateMessage posts a message to a Discord channel.
// See: https://discord.com/developers/docs/resources/message#create-message
func (c *Client) CreateMessage(ctx context.Context, channelID string, reqBody CreateMessageRequestBody) error {
	reqURL, err := url.JoinPath(c.config.API, "channels", channelID, "messages")
	if err != nil {
		return fmt.Errorf("failed to create request URL: %w", err)
	}

	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, reqURL, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return c.do(ctx, req)
}
