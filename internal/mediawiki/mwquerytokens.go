package mediawiki

import "context"

type queryResponseTokens struct {
	Tokens struct {
		LoginToken string `json:"logintoken"`
	} `json:"tokens"`
}

// LoginToken retrieves a login token from [API:Tokens].
//
// [API:Tokens]: https://www.mediawiki.org/wiki/API:Tokens
func (c *Client) LoginToken(ctx context.Context) (string, error) {
	req, err := c.newQuery(ctx)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("meta", "tokens")
	q.Set("type", "login")
	req.URL.RawQuery = q.Encode()

	var tokens queryResponseTokens

	err = c.doQuery(ctx, req, &tokens)
	if err != nil {
		return "", err
	}

	return tokens.Tokens.LoginToken, nil
}
