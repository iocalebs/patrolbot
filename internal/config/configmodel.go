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
	// Wiki URL (e.g. https://en.wikipedia.org)
	URL string `json:"url"`

	// Special:BotPasswords username (e.g. PhantomCaleb@PatrolBot)
	Username string `json:"username"`

	// Special:BotPasswords password
	Password string `json:"password"`
}
