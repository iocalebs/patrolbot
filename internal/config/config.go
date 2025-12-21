// Package config provides functionality to load static app config from a YAML file.
package config

//go:generate go run ../tools/genschema/genschema.go

// Config represents the entire application configuration.
type Config struct {
	Wikis map[string]Wiki `json:"wikis"`
}

// Wiki represents bot configuration for a particular MediaWiki instance.
type Wiki struct {
	APIURL   string `json:"apiUrl"`
	Username string `json:"username"`
	Password string `json:"password"`
}
