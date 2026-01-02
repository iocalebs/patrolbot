package config

// Config represents the PatrolBot config.
type Config struct {
	// Target wiki for bot commands
	Wiki string `json:"wiki"`

	// Wiki configurations
	Wikis map[string]Wiki `json:"wikis"`
}

// Wiki represents bot configuration for a particular MediaWiki instance.
type Wiki struct {
	// MediaWiki site information that PatrolBot needs to be aware of.
	Site WikiSite `json:"site"`

	// Authentication credentials for the MediaWiki API
	// See https://www.mediawiki.org/wiki/API:Login
	Auth WikiAuth `json:"auth"`

	// Configuration for the HTTP client used for MediaWiki API requests.
	Client WikiClient `json:"client"`
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
}

// WikiAuth represents configuration for MediaWiki API login credentials.
type WikiAuth struct {
	// Special:BotPasswords username (e.g. PhantomCaleb@PatrolBot)
	Username string `json:"username"`

	// Special:BotPasswords password
	Password string `json:"password"`
}

// WikiClient represents configuration for the HTTP client used for MediaWiki API requests.
type WikiClient struct {
	// User-Agent header for API requests
	UserAgent string `json:"userAgent"`
}
