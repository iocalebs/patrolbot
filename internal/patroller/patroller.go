// Package patroller provides functionality for iterating through unpatrolled revisions in a MediaWiki backlog.
package patroller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

var errNoThankTarget = errors.New("current change has no revision or log ID")

// A Patroller iterates through unpatrolled revisions in a MediaWiki backlog, provides information about each revision,
// and allows the caller to mark revisions as patrolled.
type Patroller struct {
	mwclient    *mediawiki.Client
	rcPaginator *mediawiki.RecentChangesPaginator

	current     mediawiki.RecentChange
	csrfToken   string
	err         error
	patrolToken string
}

// Opts provides the options for configuring a [Patroller].
type Opts struct {
	// If true, the patroller will start with the latest unpatrolled revision and work backwards.
	// If false, it will start with the oldest unpatrolled revision and work forwards.
	Latest bool

	// If non-empty, the patroller will only return revisions in this namespace.
	Namespace string

	// If non-empty, the patroller will only return revisions made by this user.
	User string
}

// New returns a new [Patroller].
func New(mwclient *mediawiki.Client, opts Opts) (*Patroller, error) {
	params := mediawiki.RecentChangesQueryParams{
		RCLimit: 1,
		RCProp:  []string{"comment", "ids", "loginfo", "timestamp", "title", "user"},
		RCShow:  "!patrolled",
	}
	if opts.Latest {
		params.RCDir = "older"
	} else {
		params.RCDir = "newer"
	}

	if opts.Namespace != "" {
		ns, err := strconv.Atoi(opts.Namespace)
		if err != nil {
			return nil, fmt.Errorf("invalid namespace number: %w", err)
		}

		params.RCNamespace = []int{ns}
	}

	if opts.User != "" {
		params.RCUser = opts.User
	}

	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	return &Patroller{
		mwclient:    mwclient,
		rcPaginator: paginator,
	}, nil
}

// Next advances to the next unpatrolled revision in the backlog.
// It returns false when there are no more revisions to patrol, or if an error occurred querying the MediaWiki API.
// After Next returns false, the [Patroller.Err] method will return the first error that occurred.
func (p *Patroller) Next(ctx context.Context) bool {
	if p.patrolToken == "" {
		p.patrolToken = p.login(ctx)
		if p.err != nil {
			return false
		}
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

// Comment returns the comment of the revision the patroller is currently on.
func (p *Patroller) Comment() string {
	return p.current.Comment
}

// Diff returns a unified diff of the revision the patroller is currently on.
func (p *Patroller) Diff(ctx context.Context) (string, error) {
	if p.current.Type == "new" {
		text, err := p.mwclient.RevisionText(ctx, mediawiki.RevisionQueryParams{
			RevisionID: p.current.RevisionID,
		})
		if err != nil {
			return "", fmt.Errorf("error fetching new page revision text: %w", err)
		}

		return newPageDiff(text), nil
	}

	diff, err := p.mwclient.Compare(ctx, mediawiki.CompareQueryParams{
		DiffType: mediawiki.DiffTypeUnified,
		FromRev:  p.current.OldRevisionID,
		ToRev:    p.current.RevisionID,
	})
	if err != nil {
		return "", fmt.Errorf("error fetching diff: %w", err)
	}

	return diff, nil
}

func newPageDiff(text string) string {
	lines := strings.Split(text, "\n")

	var sb strings.Builder

	fmt.Fprintf(&sb, "@@ -0,0 +1,%d @@\n", len(lines))

	for i, line := range lines {
		sb.WriteString("+")
		sb.WriteString(line)

		if i != len(lines)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// LogAction returns the string "upload" if the current revision is a new file upload, "overwrite" if it's an upload
// on an existing file, or an empty string if it's a regular edit.
func (p *Patroller) LogAction() string {
	return p.current.LogAction
}

// MarkPatrolled marks the revision the patroller is currently on as patrolled.
func (p *Patroller) MarkPatrolled(ctx context.Context) error {
	err := p.mwclient.Patrol(ctx, p.patrolToken, mediawiki.PatrolParams{
		RevisionID: p.current.RevisionID,
	})
	if err != nil {
		return fmt.Errorf("error marking revision as patrolled: %w", err)
	}

	return nil
}

// Thank sends a thanks notification for the revision the patroller is currently on.
func (p *Patroller) Thank(ctx context.Context) error {
	params := mediawiki.ThankParams{}

	switch {
	case p.current.RevisionID != 0:
		params.RevisionID = p.current.RevisionID
	case p.current.LogID != 0:
		params.LogID = p.current.LogID
	default:
		return errNoThankTarget
	}

	if p.csrfToken == "" {
		tokens, err := p.mwclient.Tokens(ctx, "csrf")
		if err != nil {
			return fmt.Errorf("failed to retrieve CSRF token: %w", err)
		}

		p.csrfToken = tokens.CSRF
	}

	err := p.mwclient.Thank(ctx, p.csrfToken, params)
	if err != nil {
		return fmt.Errorf("error thanking user: %w", err)
	}

	return nil
}

// RevisionID returns the revision ID of the revision the patroller is currently on.
func (p *Patroller) RevisionID() int {
	return p.current.RevisionID
}

// RevisionType returns the type of the revision the patroller is currently on (e.g. "edit", "new", "log").
func (p *Patroller) RevisionType() string {
	return p.current.Type
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

func (p *Patroller) login(ctx context.Context) string {
	// Step 1: Get login token
	tokens, err := p.mwclient.Tokens(ctx, "login")
	if err != nil {
		err = fmt.Errorf("MediaWiki login error: %w", err)
		p.setErr(err)

		return ""
	}

	// Step 2: Log in with the login token
	err = p.mwclient.Login(ctx, tokens.Login)
	if err != nil {
		err = fmt.Errorf("MediaWiki login error: %w", err)
		p.setErr(err)

		return ""
	}

	// Step 3: Get patrol token after login
	tokens, err = p.mwclient.Tokens(ctx, "patrol")
	if err != nil {
		err = fmt.Errorf("failed to retrieve patrol token: %w", err)
		p.setErr(err)

		return ""
	}

	return tokens.Patrol
}

func (p *Patroller) setErr(err error) {
	if p.err == nil || errors.Is(p.err, io.EOF) {
		p.err = err
	}
}
