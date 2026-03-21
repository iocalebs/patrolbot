package mediawiki

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ErrRevisionNotFound indicates that a revision was not found.
var ErrRevisionNotFound = errors.New("revision not found")

// RevisionQueryParams represents the query parameters for [API:Revisions] requests.
//
// [API:Revisions]: https://www.mediawiki.org/wiki/API:Revisions
type RevisionQueryParams struct {
	// The revision ID to fetch.
	RevisionID int
}

// Revision represents the subset of a MediaWiki revision needed by the app.
type Revision struct {
	Slots struct {
		Main struct {
			Content string `json:"content"`
		} `json:"main"`
	} `json:"slots"`
}

type revisionsQueryResponse struct {
	baseResponse

	Query struct {
		Pages []struct {
			Revisions []Revision `json:"revisions"`
		} `json:"pages"`
	} `json:"query"`
}

// RevisionText executes an [API:Revisions] request and returns the revision's content.
//
// [API:Revisions]: https://www.mediawiki.org/wiki/API:Revisions
func (c *Client) RevisionText(ctx context.Context, params RevisionQueryParams) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("action", "query")
	q.Set("prop", "revisions")
	q.Set("revids", strconv.Itoa(params.RevisionID))
	q.Set("rvprop", "content")
	q.Set("rvslots", "main")
	req.URL.RawQuery = q.Encode()

	var res revisionsQueryResponse

	err = c.do(ctx, req, &res)
	if err != nil {
		return "", err
	}

	if len(res.Query.Pages) == 0 || len(res.Query.Pages[0].Revisions) == 0 {
		return "", fmt.Errorf("%w: %d", ErrRevisionNotFound, params.RevisionID)
	}

	return res.Query.Pages[0].Revisions[0].Slots.Main.Content, nil
}
