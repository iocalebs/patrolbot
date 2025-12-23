package main_test

// TODO: add tests for config view covering precedence rules

import (
	"flag"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var update = flag.Bool("update", false, "update golden files") //nolint:gochecknoglobals

func TestCommands(t *testing.T) {
	t.Parallel()

	params := testscript.Params{}
	params.Dir = "testdata/integration-tests"
	params.UpdateScripts = *update
	params.Setup = func(env *testscript.Env) error {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error getting current working directory: %v", err)
		}

		// Add patrolbot binary to PATH
		// Assuming binary was already built in go test's working dir
		env.Setenv("PATH", cwd)

		return nil
	}

	testscript.Run(t, params)
}
