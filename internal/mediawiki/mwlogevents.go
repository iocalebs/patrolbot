package mediawiki

import (
	"context"
	"strconv"
	"time"
)

// LogEventsQueryParams represents the query parameters for [API:Logevents] requests.
//
// [API:Logevents]: https://www.mediawiki.org/wiki/API:Logevents
type LogEventsQueryParams struct {
	// Which properties to include in the response.
	LEProp string

	// Filter log entries to only this type.
	LEType string

	// The timestamp to start enumerating from.
	LEStart time.Time

	// The timestamp to end enumerating.
	LEEnd time.Time

	// How many total events to return. The value must be between 1 and 500. Default: 10.
	LELimit int
}

// LogEvent represents an entry in [Special:Log].
//
// [Special:Log]: https://www.mediawiki.org/wiki/Help:Log
type LogEvent struct{}

// LogEventsPage is a page of results from [API:Logevents].
//
// [API:Logevents]: https://www.mediawiki.org/wiki/API:Logevents
type LogEventsPage struct {
	LogEvents []LogEvent
}

type logEventsResponse struct {
	baseResponse

	Continue struct {
		LEContinue string `json:"lecontinue"`
	} `json:"continue"`

	Query struct {
		LogEvents []LogEvent `json:"logevents"`
	}
}

// LogEventsPaginator is a paginator for [API:Logevents].
//
// [API:Logevents]: https://www.mediawiki.org/wiki/API:Logevents
type LogEventsPaginator struct {
	client    *Client
	params    LogEventsQueryParams
	cursor    string
	firstPage bool
}

// NewLogEventsPaginator returns a new LogEventsPaginator.
func NewLogEventsPaginator(client *Client, params LogEventsQueryParams) *LogEventsPaginator {
	return &LogEventsPaginator{
		client:    client,
		params:    params,
		cursor:    "",
		firstPage: true,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available.
func (p *LogEventsPaginator) HasMorePages() bool {
	return p.firstPage || p.cursor != ""
}

// NextPage requests the next page of logs.
func (p *LogEventsPaginator) NextPage(ctx context.Context) (LogEventsPage, error) {
	req, err := p.client.newQuery(ctx)
	if err != nil {
		return LogEventsPage{}, err
	}

	query := req.URL.Query()
	query.Set("list", "logevents")

	if p.params.LEProp != "" {
		query.Set("leprop", p.params.LEProp)
	}

	if p.params.LEType != "" {
		query.Set("letype", p.params.LEType)
	}

	if !p.params.LEStart.IsZero() {
		query.Set("lestart", p.params.LEStart.Format(time.RFC3339))
	}

	if !p.params.LEEnd.IsZero() {
		query.Set("leend", p.params.LEEnd.Format(time.RFC3339))
	}

	if p.params.LELimit != 0 {
		query.Set("lelimit", strconv.Itoa(p.params.LELimit))
	}

	if !p.firstPage {
		query.Set("lecontinue", p.cursor)
	}

	req.URL.RawQuery = query.Encode()

	var res logEventsResponse

	err = p.client.do(ctx, req, &res)

	page := LogEventsPage{
		LogEvents: res.Query.LogEvents,
	}
	p.cursor = res.Continue.LEContinue
	p.firstPage = false

	return page, err
}
