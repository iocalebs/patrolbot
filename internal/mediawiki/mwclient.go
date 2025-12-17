// Package mediawiki provides a client for interacting with the MediaWiki API.
package mediawiki

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
)

// A Client is a MediaWiki API client.
type Client struct {
	config     config.MediaWiki
	httpClient *http.Client
	logger     slog.Logger
}

// NewClient creates a new MediaWiki API client.
func NewClient(config config.MediaWiki, httpClient *http.Client, logger slog.Logger) *Client {
	client := &Client{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}

	return client
}

func (c *Client) newRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.config.APIURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create MediaWiki request: %w", err)
	}

	req.Header.Set("User-Agent", "PatrolBot/1.0 (+https://github.com/iocalebs/patrolbot; phantomcalebs@gmail.com)")

	return req, nil
}
