package mediawiki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var errMediaWikiWarnings = errors.New("MediaWiki response body contains warnings")
var errMediaWikiErrors = errors.New("MediaWiki response body contains errors")

type queryResponseBody struct {
	Warnings *json.RawMessage `json:"warnings"`
	Errors   *json.RawMessage `json:"errors"`
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
			c.logger.WarnContext(ctx, "failed to close MediaWiki response body", "error", err)
		}
	}()

	var responseBody queryResponseBody

	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return fmt.Errorf("failed to decode MediaWiki response body: %w", err)
	}

	if responseBody.Errors != nil {
		return fmt.Errorf("%w: %s", errMediaWikiErrors, string(*responseBody.Errors))
	}

	if responseBody.Warnings != nil {
		return fmt.Errorf("%w: %s", errMediaWikiWarnings, string(*responseBody.Warnings))
	}

	err = json.Unmarshal(*responseBody.Query, res)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MediaWiki query response: %w", err)
	}

	return nil
}
