// Package patroller provides functionality for iterating through unpatrolled revisions in a MediaWiki backlog.
package patroller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

// A Patroller iterates through unpatrolled revisions in a MediaWiki backlog, provides information about each revision,
// and allows the caller to mark revisions as patrolled.
type Patroller struct {
	mwclient    *mediawiki.Client
	rcPaginator *mediawiki.RecentChangesPaginator

	current     mediawiki.RecentChange
	err         error
	patrolToken string
}

// Opts provides the options for configuring a [Patroller].
type Opts struct {
	// If true, the patroller will start with the latest unpatrolled revision and work backwards.
	// If false, it will start with the oldest unpatrolled revision and work forwards.
	Latest bool

	// If non-empty, the patroller will only return revisions made by this user.
	User string
}

// New returns a new [Patroller].
func New(mwclient *mediawiki.Client, opts Opts) *Patroller {
	params := mediawiki.RecentChangesQueryParams{
		RCLimit: 1,
		RCProp:  []string{"ids", "loginfo", "timestamp", "title", "user"},
		RCShow:  "!patrolled",
	}
	if opts.Latest {
		params.RCDir = "older"
	} else {
		params.RCDir = "newer"
	}

	if opts.User != "" {
		params.RCUser = opts.User
	}

	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	return &Patroller{
		mwclient:    mwclient,
		rcPaginator: paginator,
	}
}

// Next advances to the next unpatrolled revision in the backlog.
// It returns false when there are no more revisions to patrol, or if an error occurred querying the MediaWiki API.
// After Next returns false, the [Patroller.Err] method will return the first error that occurred.
func (p *Patroller) Next(ctx context.Context) bool {
	if p.patrolToken == "" {
		tokens, err := p.mwclient.Tokens(ctx, "login|patrol")
		if err != nil {
			err = fmt.Errorf("MediaWiki login error: %w", err)
			p.setErr(err)

			return false
		}

		err = p.mwclient.Login(ctx, tokens.Login)
		if err != nil {
			err = fmt.Errorf("MediaWiki login error: %w", err)
			p.setErr(err)

			return false
		}

		p.patrolToken = tokens.Patrol
	}

	if !p.rcPaginator.HasMorePages() {
		return false
	}

	page, err := p.rcPaginator.NextPage(ctx)
	if err != nil {
		p.setErr(err)
		return false
	}

	if len(page.RecentChanges) == 0 {
		return false
	}

	p.current = page.RecentChanges[0]

	return true
}

// Err returns the first non-EOF error that was encountered by the [Patroller], if any.
func (p *Patroller) Err() error {
	if errors.Is(p.err, io.EOF) {
		return nil
	}

	return p.err
}

// Timestamp returns the timestamp of the revision the patroller is currently on.
func (p *Patroller) Timestamp() time.Time {
	return p.current.Timestamp
}

// Title returns the page title of the revision the patroller is currently on.
func (p *Patroller) Title() string {
	return p.current.Title
}

// URL returns a URL to the diff of the revision the patroller is currently on.
func (p *Patroller) URL() (string, error) {
	wiki := p.mwclient.Wiki().Site

	url, err := url.Parse(wiki.URL)
	if err != nil {
		return "", fmt.Errorf("error generating site URL: %w", err)
	}

	url = url.JoinPath(wiki.ScriptPath, "index.php")
	q := url.Query()
	q.Add("title", p.current.Title)
	q.Add("diff", strconv.Itoa(p.current.RevisionID))
	q.Add("oldid", strconv.Itoa(p.current.OldRevisionID))
	url.RawQuery = q.Encode()

	return url.String(), nil
}

// User returns the user of the revision the patroller is currently on.
func (p *Patroller) User() string {
	return p.current.User
}

func (p *Patroller) setErr(err error) {
	if p.err == nil || errors.Is(p.err, io.EOF) {
		p.err = err
	}
}
