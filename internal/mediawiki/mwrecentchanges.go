package mediawiki

import (
	"context"
	"strconv"
	"time"
)

// RecentChangesQueryParams represents the query parameters for [API:RecentChanges] requests.
//
// [API:RecentChanges]: https://www.mediawiki.org/wiki/API:RecentChanges
type RecentChangesQueryParams struct {
	// The timestamp to start enumerating from.
	RCStart time.Time

	// The timestamp to end enumerating.
	RCEnd time.Time

	// Show only items that meet these criteria.
	// For example, to see only minor edits done by logged-in users, set rcshow=minor|!anon.
	RCShow string

	// How many total changes to return. The value must be between 1 and 500. Default: 10.
	RCLimit int
}

// RecentChange represents an entry in [Special:RecentChanges].
//
// [Special:RecentChangs]: https://www.mediawiki.org/wiki/Help:Recent_changes
type RecentChange struct {
	Timestamp time.Time `json:"timestamp"`
}

type recentChangesQueryResponse struct {
	baseResponse

	Continue struct {
		RCContinue string `json:"rccontinue"`
	} `json:"continue"`

	Query struct {
		RecentChanges []RecentChange `json:"recentchanges"`
	}
}

// RecentChangesPage is a page of results from [API:RecentChanges].
//
// [API:RecentChanges]: https://www.mediawiki.org/wiki/API:RecentChanges
type RecentChangesPage struct {
	RecentChanges []RecentChange
}

// RecentChangesPaginator is a paginator for [API:RecentChanges].
//
// [API:RecentChanges]: https://www.mediawiki.org/wiki/API:RecentChanges
type RecentChangesPaginator struct {
	client    *Client
	params    RecentChangesQueryParams
	cursor    string
	firstPage bool
}

// NewRecentChangesPaginator returns a new RecentChangesPaginator.
func NewRecentChangesPaginator(client *Client, params RecentChangesQueryParams) *RecentChangesPaginator {
	return &RecentChangesPaginator{
		client:    client,
		params:    params,
		cursor:    "",
		firstPage: true,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available.
func (p *RecentChangesPaginator) HasMorePages() bool {
	return p.firstPage || p.cursor != ""
}

// NextPage requests the next page of recent changes.
func (p *RecentChangesPaginator) NextPage(ctx context.Context) (RecentChangesPage, error) {
	req, err := p.client.newQuery(ctx)
	if err != nil {
		return RecentChangesPage{}, err
	}

	query := req.URL.Query()
	query.Set("list", "recentchanges")

	if !p.params.RCStart.IsZero() {
		query.Set("rcstart", p.params.RCStart.Format(time.RFC3339))
	}

	if !p.params.RCEnd.IsZero() {
		query.Set("rcend", p.params.RCEnd.Format(time.RFC3339))
	}

	if p.params.RCShow != "" {
		query.Set("rcshow", p.params.RCShow)
	}

	if p.params.RCLimit != 0 {
		query.Set("rclimit", strconv.Itoa(p.params.RCLimit))
	}

	if !p.firstPage {
		query.Set("rccontinue", p.cursor)
	}

	req.URL.RawQuery = query.Encode()

	var res recentChangesQueryResponse

	err = p.client.do(ctx, req, &res)

	page := RecentChangesPage{
		RecentChanges: res.Query.RecentChanges,
	}
	p.cursor = res.Continue.RCContinue
	p.firstPage = false

	return page, err
}
