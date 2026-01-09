// Package data exports the type of the [ReportData] object in PatrolBot report templates.
package data

import "github.com/iocalebs/patrolbot/internal/config"

// ReportData contains the data used to render a report template.
type ReportData struct {
	// An error message indicating an error occurred while querying the MediaWiki API.
	Error string

	// Wiki configuration - URL, article path, etc.
	// The password field is redacted
	Config config.Wiki

	// Counts of unpatrolled edits in danger of aging out of Special:RecentChanges
	// [github.com/iocalebs/patrolbot/internal/config.ReportSources.PatrolExpiring] must be set for this data to be
	// sourced for the report
	PatrolExpiring PatrolExpiring
}

// PatrolExpiring contains counts of unpatrolled edits that will reach [$wgRCMaxage] within the given time windows.
//
// [$wgRCMaxage]: https://www.mediawiki.org/wiki/Manual:$wgRCMaxAge
type PatrolExpiring struct {
	// Counts of edits expiring in specified time windows
	Windows []ExpiryWindow
}

// An ExpiryWindow contains a duration of time and a count of the unpatrolled edits that will reach [$wgRCMaxage]
// and expire from [Special:RecentChanges] by the end of the given time period.
//
// [$wgRCMaxage]: https://www.mediawiki.org/wiki/Manual:$wgRCMaxAge
// [Special:RecentChanges]: https://www.mediawiki.org/wiki/Manual:Recent_changes
type ExpiryWindow struct {
	Count    int
	Duration config.Duration
}
