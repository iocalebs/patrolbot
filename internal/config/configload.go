package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

//go:embed config.yaml
var config string

// ErrMissingEnvVariables indicates that required environment variables are undefined or empty.
var ErrMissingEnvVariables = errors.New("required environment variables are undefined or empty")

// Load loads the application configuration from the embedded YAML file.
func Load() (*Config, error) {
	expandedConfig, expandErr := expandEnv(config)

	var cfg Config

	err := yaml.Unmarshal([]byte(expandedConfig), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, expandErr
}

func expandEnv(config string) (string, error) {
	missing := []string{}

	// Strip comment lines which may provide false positives for env expansion
	// e.g. # yaml-language-server: $schema=config.schema.json
	reg := regexp.MustCompile(`^\s*#[^\n]+\n`)
	config = reg.ReplaceAllString(config, "")

	expanded := os.Expand(config, func(envVar string) string {
		val := os.Getenv(envVar)
		// Allow undefined env variables for wiki-specific config so that theapp can still
		// run even if not all wikis have the env variables configured,
		// which may well be the case when running locally
		if val == "" && !strings.HasPrefix(envVar, "MW_") {
			missing = append(missing, envVar)
		}

		return val
	})

	var err error

	if len(missing) > 0 {
		vars := strings.Join(missing, ", ")
		err = fmt.Errorf("%w: %s", ErrMissingEnvVariables, vars)
	}

	return expanded, err
}
