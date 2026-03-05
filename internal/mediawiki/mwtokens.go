package mediawiki

import (
	"context"
	"net/http"
)

// Tokens represents one or more MediaWiki API tokens.
type Tokens struct {
	Login  string `json:"logintoken"`
	Patrol string `json:"patroltoken"`
}

type queryResponseTokens struct {
	baseResponse

	Query struct {
		Tokens Tokens `json:"tokens"`
	}
}

// Tokens retrieves tokens from [API:Tokens] given a type string.
//
// [API:Tokens]: https://www.mediawiki.org/wiki/API:Tokens
func (c *Client) Tokens(ctx context.Context, tokenType string) (Tokens, error) {
	req, err := c.newRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return Tokens{}, err
	}

	q := req.URL.Query()
	q.Set("action", "query")
	q.Set("meta", "tokens")
	q.Set("type", tokenType)
	req.URL.RawQuery = q.Encode()

	var res queryResponseTokens

	err = c.do(ctx, req, &res)
	if err != nil {
		return Tokens{}, err
	}

	return res.Query.Tokens, nil
}
