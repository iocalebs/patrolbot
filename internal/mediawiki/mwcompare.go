package mediawiki

import (
	"context"
	"net/http"
	"strconv"
)

// DiffType represents the type of diff in comparison requests.
type DiffType string

const (
	// DiffTypeText produces a diff in plain text format.
	DiffTypeText DiffType = "inline"

	// DiffTypeHTML produces a diff in an HTML table format.
	DiffTypeHTML DiffType = "table"

	// DiffTypeUnified produces a diff in unified format.
	DiffTypeUnified DiffType = "unified"
)

// CompareQueryParams represents the query parameters for [API:Compare] requests.
//
// [API:Compare]: https://www.mediawiki.org/wiki/API:Compare
type CompareQueryParams struct {
	// Return the comparison formatted as inline HTML.
	// Default: table.
	DiffType DiffType

	// First revision to compare.
	FromRev int

	// Second revision to compare.
	ToRev int
}

type queryResponseCompare struct {
	baseResponse

	Compare struct {
		Body string `json:"body"`
	}
}

// Compare executes an [API:Compare] request with the given query parameters.
//
// [API:Compare]: https://www.mediawiki.org/wiki/API:Compare
func (c *Client) Compare(ctx context.Context, params CompareQueryParams) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("action", "compare")
	q.Set("fromrev", strconv.Itoa(params.FromRev))
	q.Set("torev", strconv.Itoa(params.ToRev))

	if params.DiffType != "" {
		q.Set("difftype", string(params.DiffType))
	}

	req.URL.RawQuery = q.Encode()

	var res queryResponseCompare

	err = c.do(ctx, req, &res)
	if err != nil {
		return "", err
	}

	return res.Compare.Body, nil
}
