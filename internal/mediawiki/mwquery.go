package mediawiki

import (
	"context"
	"net/http"
)

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
