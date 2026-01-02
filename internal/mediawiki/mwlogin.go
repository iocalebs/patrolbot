package mediawiki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ErrLoginFailed indicates that the Login action failed.
// See https://www.mediawiki.org/wiki/API:Login#Possible_errors
var ErrLoginFailed = errors.New("Login failed")

type login struct {
	Result string `json:"result"`
}

type loginResponseBody struct {
	baseResponse

	Login json.RawMessage `json:"login"`
}

// Login authenticates the client via [bot password] using [API:Login].
// If successful, a session cookie is created for the underlying HTTP client.
//
// [bot password]: https://www.mediawiki.org/wiki/Manual:Bot_passwords
// [API:Login]: https://www.mediawiki.org/wiki/API:Login
func (c *Client) Login(ctx context.Context, token Token) error {
	formBody := url.Values{
		"action":        {"login"},
		"lgname":        {c.config.Auth.Username},
		"lgpassword":    {c.config.Auth.Password},
		"lgtoken":       {string(token)},
		"format":        {"json"},
		"formatversion": {"2"},
	}.Encode()

	req, err := c.newRequest(ctx, http.MethodPost, strings.NewReader(formBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var res loginResponseBody

	err = c.do(ctx, req, &res)
	if err != nil {
		return err
	}

	var loginData login

	err = json.Unmarshal(res.Login, &loginData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal login action response: %w", err)
	}

	if loginData.Result != "Success" {
		return fmt.Errorf("%w: %v", ErrLoginFailed, string(res.Login))
	}

	return nil
}
