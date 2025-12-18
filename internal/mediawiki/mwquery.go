package mediawiki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type queryResponse struct {
	baseResponse

	Query *json.RawMessage `json:"query"`
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

func (c *Client) doQuery(ctx context.Context, req *http.Request, query any) error {
	var res queryResponse

	err := c.do(ctx, req, &res)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*res.Query, &query)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MediaWiki query response: %w", err)
	}

	return nil
}
