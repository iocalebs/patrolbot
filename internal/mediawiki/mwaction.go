package mediawiki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type mediaWikiError struct {
	Text string `json:"text"`
}

type baseResponse struct {
	Warnings []mediaWikiError `json:"warnings"`
	Errors   []mediaWikiError `json:"errors"`
}

type response interface {
	warnings() []string
	errors() []string
}

func (b baseResponse) warnings() []string {
	warnings := make([]string, len(b.Warnings))
	for i, warning := range b.Warnings {
		warnings[i] = warning.Text
	}

	return warnings
}

func (b baseResponse) errors() []string {
	errs := make([]string, len(b.Errors))
	for i, err := range b.Errors {
		errs[i] = err.Text
	}

	return errs
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

	q := req.URL.Query()
	q.Set("errorformat", "plaintext")
	q.Set("format", "json")
	q.Set("formatversion", "2")
	q.Set("uselang", "user")
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

			return &HTTPStatusError{
				StatusCode: resp.StatusCode,
				Body:       []byte(""),
			}
		}

		return &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return fmt.Errorf("failed to decode MediaWiki response body: %w", err)
	}

	errs := res.errors()
	warnings := res.warnings()

	if len(errs) > 0 || len(warnings) > 0 {
		return &APIError{
			Errors:   errs,
			Warnings: warnings,
		}
	}

	return nil
}
