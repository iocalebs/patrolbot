package config_test

import (
	"testing"

	"github.com/iocalebs/patrolbot/internal/config"
)

func TestLoad(t *testing.T) {
	username := "username"
	password := "password"

	t.Setenv("MW_USERNAME_ZWEN", username)
	t.Setenv("MW_PASSWORD_ZWEN", password)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	zwEn, ok := cfg.Wikis["zw_en"]
	if !ok {
		t.Fatal("Expected wikis[\"zw_en\"] to exist")
	}

	if zwEn.APIURL == "" {
		t.Fatal("Expected wikis[\"zw_en\"].APIURL to not be empty")
	}

	if zwEn.Username != username {
		t.Fatalf("Expected wikis[\"zw_en\"].Username to be %q, got %q", username, zwEn.Username)
	}

	if zwEn.Password != password {
		t.Fatalf("Expected wikis[\"zw_en\"].Password to be %q, got %q", password, zwEn.Password)
	}
}
