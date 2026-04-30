package mediawiki

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ThankParams represents the parameters for an [API:Thank] request.
//
// [API:Thank]: https://www.mediawiki.org/wiki/Extension:Thanks#API_documentation
type ThankParams struct {
	// LogID is the log ID to thank.
	// Either LogID or RevisionID must be specified.
	LogID int

	// RevisionID is the revision ID to thank.
	// Either LogID or RevisionID must be specified.
	RevisionID int
}

// Thank sends a thanks notification using [API:Thank].
//
// [API:Thank]: https://www.mediawiki.org/wiki/Extension:Thanks#API_documentation
func (c *Client) Thank(ctx context.Context, token string, params ThankParams) error {
	formBody := url.Values{
		"source": {"patrolbot"},
		"token":  {token},
	}

	switch {
	case params.RevisionID != 0:
		formBody.Set("rev", strconv.Itoa(params.RevisionID))
	case params.LogID != 0:
		formBody.Set("log", strconv.Itoa(params.LogID))
	}

	req, err := c.newRequest(ctx, http.MethodPost, strings.NewReader(formBody.Encode()))
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Set("action", "thank")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var res baseResponse

	err = c.do(ctx, req, &res)
	if err != nil {
		return err
	}

	return nil
}
