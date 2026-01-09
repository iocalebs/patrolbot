package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	"github.com/iocalebs/patrolbot/internal/report/data"
)

const rcLimitMax = 500

func (p *Provider) patrolExpiring(ctx context.Context, cfg config.PatrolExpiring) (data.PatrolExpiring, error) {
	// Query API:RecentChanges using the largest expiry window for the time range, then filter for each window
	// oldest edit in range = now - $wgRCMaxAge
	// newest edit in range = rcend + max window duration
	oldest := p.clock.Now().UTC().Add(-time.Duration(p.wiki.Site.RCMaxAgeDays) * 24 * time.Hour)

	var newest time.Time

	windowStarts := make([]time.Time, len(cfg.Windows))
	for i, duration := range cfg.Windows {
		windowStarts[i] = duration.AddTo(oldest)
		if windowStarts[i].After(newest) {
			newest = windowStarts[i]
		}
	}

	paginator := mediawiki.NewRecentChangesPaginator(p.mwclient, mediawiki.RecentChangesQueryParams{
		RCDir:   "newer",
		RCStart: oldest,
		RCEnd:   newest,
		RCLimit: rcLimitMax,
		RCShow:  "!patrolled",
	})

	counts := make([]int, len(cfg.Windows))

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			p.logError(ctx, "Error executing API:RecentChanges query", err)
			return data.PatrolExpiring{}, fmt.Errorf("error executing API:RecentChanges query: %w", err)
		}

		for i := range counts {
			for _, change := range page.RecentChanges {
				if change.Timestamp.After(windowStarts[i]) {
					break
				}

				counts[i]++
			}
		}
	}

	windows := make([]data.ExpiryWindow, len(cfg.Windows))
	for i := range windows {
		windows[i] = data.ExpiryWindow{
			Duration: cfg.Windows[i],
			Count:    counts[i],
		}
	}

	return data.PatrolExpiring{
		Windows: windows,
	}, nil
}
