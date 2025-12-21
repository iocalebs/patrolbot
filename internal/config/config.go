// Package config provides functionality to load config from YAML files,
// environment variables, and command-line arguments.
package config

// Config represents the PatrolBot config.
type Config struct {
	Wiki  string          `json:"wiki"`  // Target wiki for bot commands
	Wikis map[string]Wiki `json:"wikis"` // Wiki configurations
}

// Wiki represents bot configuration for a particular MediaWiki instance.
type Wiki struct {
	APIURL   string `json:"apiUrl"`
	Username string `json:"username"`
	Password string `json:"password"`
}
