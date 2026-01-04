package mediawiki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	// ErrResponseWarnings indicates that the MediaWiki API response body contained warnings.
	ErrResponseWarnings = errors.New("MediaWiki API response body contains warnings")

	// ErrResponseError indicates that the MediaWiki API response body contained errors.
	ErrResponseError = errors.New("MediaWiki API response body contains errors")
)

// HTTPStatusError represents a non-2xx HTTP response from the MediaWiki API.
type HTTPStatusError struct {
	StatusCode int    // HTTP status code
	Body       []byte // Response body, if any
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("MediaWiki API error response: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

type baseResponse struct {
	Warnings json.RawMessage `json:"warnings"`
	Error    json.RawMessage `json:"error"`
}

type response interface {
	warnings() json.RawMessage
	errors() json.RawMessage
}

func (b baseResponse) warnings() json.RawMessage {
	return b.Warnings
}

func (b baseResponse) errors() json.RawMessage {
	return b.Error
}

func (c *Client) newRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	apiURL, err := url.JoinPath(c.config.Site.URL, c.config.Site.ScriptPath, "api.php")
	if err != nil {
		return nil, fmt.Errorf("failed to create API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create MediaWiki request: %w", err)
	}

	req.Header.Set("From", c.config.Client.From)
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
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

func (c *Client) do(ctx context.Context, req *http.Request, res response) error {
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

	if resp.StatusCode >= 300 { //nolint:mnd
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.WarnContext(ctx, "Failed to read response body", "error", err)

			return HTTPStatusError{
				StatusCode: resp.StatusCode,
				Body:       []byte(""),
			}
		}

		return HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return fmt.Errorf("failed to decode MediaWiki response body: %w", err)
	}

	if errs := res.errors(); len(errs) != 0 {
		return fmt.Errorf("%w: %s", ErrResponseError, string(errs))
	}

	if warnings := res.warnings(); len(warnings) != 0 {
		return fmt.Errorf("%w: %s", ErrResponseWarnings, string(warnings))
	}

	return nil
}
