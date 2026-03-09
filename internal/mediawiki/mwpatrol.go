package mediawiki

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PatrolParams represents the parameters for a [API:Patrol] request.
//
// [API:Patrol]: https://www.mediawiki.org/wiki/API:Patrol
type PatrolParams struct {
	// RecentChangeID is the recent changes ID to patrol.
	// Either RecentChangeID or RevisionID must be specified.
	RecentChangeID int

	// RevisionID is the revision ID to patrol.
	// Either RecentChangeID or RevisionID must be specified.
	RevisionID int
}

// Patrol marks a revision as patrolled using [API:Patrol].
// Requires either RecentChangeID or RevisionID to be specified in params.
//
// [API:Patrol]: https://www.mediawiki.org/wiki/API:Patrol
func (c *Client) Patrol(ctx context.Context, token string, params PatrolParams) error {
	formBody := url.Values{
		"token": {token},
	}

	if params.RecentChangeID != 0 {
		formBody.Set("rcid", strconv.Itoa(params.RecentChangeID))
	}

	if params.RevisionID != 0 {
		formBody.Set("revid", strconv.Itoa(params.RevisionID))
	}

	req, err := c.newRequest(ctx, http.MethodPost, strings.NewReader(formBody.Encode()))
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Set("action", "patrol")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var res baseResponse

	err = c.do(ctx, req, &res)
	if err != nil {
		return err
	}

	return nil
}
