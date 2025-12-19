package mediawiki

import "context"

// Token represents a MediaWiki API token.
type Token string

type queryResponseTokens struct {
	baseResponse

	Query struct {
		Tokens struct {
			LoginToken string `json:"logintoken"`
		} `json:"tokens"`
	}
}

// LoginToken retrieves a login token from [API:Tokens].
//
// [API:Tokens]: https://www.mediawiki.org/wiki/API:Tokens
func (c *Client) LoginToken(ctx context.Context) (Token, error) {
	req, err := c.newQuery(ctx)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("meta", "tokens")
	q.Set("type", "login")
	req.URL.RawQuery = q.Encode()

	var res queryResponseTokens

	err = c.do(ctx, req, &res)
	if err != nil {
		return "", err
	}

	return Token(res.Query.Tokens.LoginToken), nil
}
