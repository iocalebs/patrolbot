package config_test

import (
	"testing"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/pflag"
)

// See testdata/integration-tests/config_view.txt for further coverage.
func TestLoadErrNoConfigFlag(t *testing.T) {
	t.Parallel()

	fs := pflag.NewFlagSet("empty", pflag.ContinueOnError)
	_, err := config.Load(fs)

	want := "error reading config flag: flag accessed but not defined: config"
	if err == nil || err.Error() != want {
		t.Errorf("Expected error %q, got: %v", want, err)
	}
}
