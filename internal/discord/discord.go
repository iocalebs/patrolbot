// Package discord provides a Discord API client, as well as types and functions for receiving Interactions via
// HTTP webhook
package discord

//go:generate go run ../tools/gentestdata --config ../../config.yaml

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
)

// Client is a Discord API client.
type Client struct {
	httpClient *http.Client
	logger     *slog.Logger
	config     config.Discord
	userAgent  string
}

// NewClient creates a new [Client].
func NewClient(httpClient *http.Client, logger *slog.Logger, cfg config.Discord, userAgent string) *Client {
	return &Client{
		httpClient: httpClient,
		logger:     logger,
		config:     cfg,
		userAgent:  userAgent,
	}
}

func (c *Client) newRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.config.Token)
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to Discord API: %w", err)
	}

	if resp.StatusCode >= 300 { //nolint:mnd
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.WarnContext(ctx, "Failed to read response body", "error", err)

			return nil, &HTTPStatusError{
				StatusCode: resp.StatusCode,
				Body:       []byte(""),
			}
		}

		return nil, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	return resp, nil
}
