package mediawiki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrResponseWarnings indicates that the MediaWiki API response body contained warnings.
	ErrResponseWarnings = errors.New("MediaWiki response body contains warnings")

	// ErrResponseError indicates that the MediaWiki API response body contained errors.
	ErrResponseError = errors.New("MediaWiki response body contains errors")

	// ErrQueryNotOK indicates that the MediaWiki API returned a non-200 OK status code.
	ErrQueryNotOK = errors.New("MediaWiki API returned non-OK status code")
)

type queryResponseBody struct {
	Warnings *json.RawMessage `json:"warnings"`
	Error    *json.RawMessage `json:"error"`
	Query    *json.RawMessage `json:"query"`
}

func (c *Client) newQuery(ctx context.Context) (*http.Request, error) {
	req, err := c.newRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("action", "query")
	q.Set("format", "json")
	q.Set("formatversion", "2")
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (c *Client) doQuery(ctx context.Context, req *http.Request, res any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request to MediaWiki API: %w", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			c.logger.WarnContext(ctx, "Failed to close MediaWiki response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.WarnContext(ctx, "Failed to read response body", "error", err)

			return fmt.Errorf("%w: %d", ErrQueryNotOK, resp.StatusCode)
		}

		if len(bodyBytes) == 0 {
			return fmt.Errorf("%w: %d with empty response body", ErrQueryNotOK, resp.StatusCode)
		}

		return fmt.Errorf("%w: %d, response body: %s", ErrQueryNotOK, resp.StatusCode, string(bodyBytes))
	}

	var responseBody queryResponseBody

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return fmt.Errorf("failed to decode MediaWiki response body: %w", err)
	}

	if responseBody.Error != nil {
		return fmt.Errorf("%w: %s", ErrResponseError, string(*responseBody.Error))
	}

	if responseBody.Warnings != nil {
		return fmt.Errorf("%w: %s", ErrResponseWarnings, string(*responseBody.Warnings))
	}

	err = json.Unmarshal(*responseBody.Query, res)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MediaWiki query response: %w", err)
	}

	return nil
}
