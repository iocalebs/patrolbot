package config_test

import (
	"testing"

	"github.com/iocalebs/patrolbot/internal/config"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.MediaWiki.APIURL == "" {
		t.Fatal("MediaWiki API URL is empty")
	}
}
