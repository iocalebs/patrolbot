package config

import (
	"log/slog"
	"time"
)

// Config represents the configuration for PatrolBot.
type Config struct {
	// Discord bot configuration
	Discord Discord `json:"discord"`

	// Logging configuration
	Log Log `json:"log"`

	// HTTP server configuration
	Server *Server `json:"server,omitempty"`

	// User-Agent HTTP header for MediaWiki and Discord API requests
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/User-Agent
	UserAgent string `json:"userAgent"`

	// Target wiki for bot commands
	Wiki string `json:"wiki"`

	// Wiki configurations
	Wikis map[string]Wiki `json:"wikis"`
}

// Discord represents the Discord bot configuration.
type Discord struct {
	// API version URL, e.g. https://discord.com/api/v10
	// See: https://discord.com/developers/docs/reference#api-versioning
	API string `json:"api"`

	// Bot token
	Token string `json:"token"`
}

// Log represents PatrolBot logging configuration.
type Log struct {
	// Level is the minimum severity level that will be logged.
	Level slog.Level `json:"level" jsonschema:"type=string,enum=DEBUG,enum=INFO,enum=WARN,enum=ERROR,default=INFO"`
}

// Server represents the configuration for the PatrolBot HTTP server.
type Server struct {
	// Port to listen on
	Port int `json:"port"`

	// HTTP server timeouts as durations, e.g. "2s" or "500ms"
	// See: https://pkg.go.dev/net/http#Server
	Timeouts Timeouts `json:"timeouts"`
}

// Timeouts represents the configuration for the PatrolBot HTTP server timeouts.
type Timeouts struct {
	ReadTimeout       time.Duration `json:"readTimeout"       jsonschema:"type=string"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" jsonschema:"type=string"`
	WriteTimeout      time.Duration `json:"writeTimeout"      jsonschema:"type=string"`
	IdleTimeout       time.Duration `json:"idleTimeout"       jsonschema:"type=string"`
}

// Wiki represents the configuration for a particular MediaWiki instance.
type Wiki struct {
	// MediaWiki site information that PatrolBot needs to be aware of.
	Site WikiSite `json:"site"`

	// Authentication credentials for the MediaWiki API
	// See https://www.mediawiki.org/wiki/API:Login
	Auth WikiAuth `json:"auth"`

	// Configuration for the HTTP client used for MediaWiki API requests.
	Client WikiClient `json:"client"`

	// Configuration for the wiki's reports
	Reports Reports `json:"reports"`
}

// WikiSite represents MediaWiki instance configuration (namely $wg variables) that PatrolBot needs to be aware of.
type WikiSite struct {
	// Wiki URL (e.g. https://en.wikipedia.org)
	URL string `json:"url"`

	// Wiki script path - typically "/w"
	// See https://www.mediawiki.org/wiki/Manual:$wgScriptPath
	ScriptPath string `json:"scriptPath"`

	// Wiki article path without /$1 ending - typically "/wiki"
	// See https://www.mediawiki.org/wiki/Manual:$wgArticlePath
	ArticlePath string `json:"articlePath"`

	// The wiki's recent changes retention period in days
	// See https://www.mediawiki.org/wiki/Manual:$wgRCMaxAge
	RCMaxAgeDays int `json:"rcMaxAgeDays"`

	// Maximum results that can be shown in the wikis' Special:RecentChanges
	// See https://www.mediawiki.org/wiki/Manual:$wgRCLinkLimits
	RCLinkLimit int `json:"rcLinkLimit"`
}

// WikiAuth represents the configuration for the bot's MediaWiki API login credentials.
type WikiAuth struct {
	// Special:BotPasswords username (e.g. PhantomCaleb@PatrolBot)
	Username string `json:"username"`

	// Special:BotPasswords password
	Password string `json:"password"`
}

// Reports represents the configuration for generating reports.
type Reports struct {
	// Path to directory containing report templates
	TemplateDir string `json:"templateDir" jsonschema:"minLength=1"`

	// Discord server (guild) ID to post reports in
	DiscordServerID string `json:"discordServerID"`

	// Defines the types of reports that can be generated using the `report` command
	Types map[string]ReportType `json:"types"`
}

// ReportType represents the configuration for a specific report type.
type ReportType struct {
	// File name of the template to use for the report
	Template string `json:"template"`

	// Discord channel to post the report to
	DiscordChannelID string `json:"discordChannelID"`

	// Data to use in the report
	Data ReportData `json:"data"`
}

// ReportData represents the configuration for the data sources used for a report type.
type ReportData struct {
	// Counts unpatrolled edits via API:RecentChanges
	PatrolExpiring *PatrolExpiring `json:"patrolExpiring,omitempty"`
}

// PatrolExpiring represents the configuration for the
// [github.com/iocalebs/patrolbot/internal/report/data.PatrolExpiring] data source.
type PatrolExpiring struct {
	// Durations of time before changes reach $wgRCMaxAge and expire from Special:RecentChanges.
	Windows []Duration `json:"windows"`
}

// WikiClient represents configuration for the HTTP client used for MediaWiki API requests.
type WikiClient struct {
	// Request timeout as a string, e.g. "2s" or "500ms"
	// A timeout of zero means no timeout
	Timeout time.Duration `json:"timeout" jsonschema:"type=string,minLength=1"`

	// From HTTP header for API requests
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/From
	From string `json:"from"`
}
