// Package mediawiki provides a client for interacting with the MediaWiki Action API.
package mediawiki

//go:generate go run ../tools/gentestdata --config ../../config.yaml

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
	userAgent  string
}

// NewClient creates a new MediaWiki API client.
func NewClient(config config.Wiki, httpClient *http.Client, logger slog.Logger, userAgent string) *Client {
	client := &Client{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		userAgent:  userAgent,
	}

	return client
}
