// Package config provides functionality to load static app config from a YAML file.
package config

//go:generate go run ../../tools/genschema/genschema.go

import (
	_ "embed"
	"fmt"

	"sigs.k8s.io/yaml"
)

// Config represents the entire application configuration.
type Config struct {
	Wikis map[string]Wiki `json:"wikis"`
}

// Wiki represents bot configuration for a particular MediaWiki instance.
type Wiki struct {
	APIURL string `json:"apiUrl"`
}

//go:embed config.yaml
var configBytes []byte

// Load loads the application configuration from the embedded YAML file.
func Load() (*Config, error) {
	var cfg Config

	err := yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
