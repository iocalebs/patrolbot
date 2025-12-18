// Package mediawiki provides a client for interacting with the MediaWiki Action API.
package mediawiki

import (
	"log/slog"
	"net/http"

	"github.com/iocalebs/patrolbot/internal/config"
)

// A Client is a MediaWiki API client.
type Client struct {
	config     config.Wiki
	httpClient *http.Client
	logger     slog.Logger
}

// NewClient creates a new MediaWiki API client.
func NewClient(config config.Wiki, httpClient *http.Client, logger slog.Logger) *Client {
	client := &Client{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}

	return client
}
