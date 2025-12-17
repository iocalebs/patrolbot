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

	zwEn, ok := cfg.Wikis["ZWen"]
	if !ok {
		t.Fatal("Expected wikis[\"ZWen\"] to exist")
	}

	if zwEn.APIURL == "" {
		t.Fatal("Expected wikis[\"ZWen\"].APIURL to not be empty")
	}
}
