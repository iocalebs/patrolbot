package main_test

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/iocalebs/patrolbot/internal/httpstub"
	"github.com/rogpeppe/go-internal/testscript"
)

var update = flag.Bool("update", false, "update golden files") //nolint:gochecknoglobals

func TestCommands(t *testing.T) {
	t.Parallel()

	params := testscript.Params{
		Dir:           "testdata/integration-tests",
		UpdateScripts: *update,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"scrub": scrub,
			"stub":  httpstub.Cmd("stub"),
		},
		Setup: func(env *testscript.Env) error {
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Error getting current working directory: %v", err)
			}

			// Add host path and patrolbot binary to test PATH
			// Assuming binary was already built in go test's working dir
			hostPath := os.Getenv("PATH")
			testPath := hostPath + string(os.PathListSeparator) + cwd
			env.Setenv("PATH", testPath)

			srv := httpstub.InitServer(env, func(dump []byte) []byte {
				addr := env.Getenv("STUB_ADDR")
				dump = bytes.ReplaceAll(dump, []byte(addr), []byte("$STUB_ADDR"))

				return dump
			})
			env.Setenv("STUB_ADDR", srv.Addr)

			return nil
		},
	}

	testscript.Run(t, params)
}

func scrub(ts *testscript.TestScript, _ bool, args []string) {
	file := args[0]
	scrubbed := ts.ReadFile(file)
	scrubbed = scrubWorkDir(ts, scrubbed)
	scrubbed = scrubVHS(scrubbed)
	scrubbed = scrubMockServerAddr(ts, scrubbed)
	scrubbed = trimTrailingWhitespace(scrubbed)

	var err error

	switch file {
	case "stdout":
		_, err = ts.Stdout().Write([]byte(scrubbed))
		if err != nil {
			ts.Fatalf("scrub: %v", err)
		}

		// Calling Stdout() flushes both stdout and stderr, therefore stderr needs to be copied over (and vice-versa)
		_, err := ts.Stderr().Write([]byte(ts.ReadFile("stderr")))
		if err != nil {
			ts.Fatalf("scrub: %v", err)
		}

	case "stderr":
		_, err = ts.Stderr().Write([]byte(scrubbed))
		if err != nil {
			ts.Fatalf("scrub: %v", err)
		}

		_, err = ts.Stdout().Write([]byte(ts.ReadFile("stdout")))
		if err != nil {
			ts.Fatalf("scrub: %v", err)
		}

	default:
		err = os.WriteFile(ts.MkAbs(file), []byte(scrubbed), 0600)
		if err != nil {
			ts.Fatalf("scrub: %v", err)
		}
	}
}

func trimTrailingWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}

	return strings.Join(lines, "\n")
}

// Replaces $WORK value (a random test dir) with placeholder so that tests that output subdirs can use the
// `cmp` testscript command. `cmpenv` is not ideal as it cannot update golden files and can expand
// text that is not actually an env variable (e.g. /$1 as used in `patrolbot init` help).
func scrubWorkDir(ts *testscript.TestScript, text string) string {
	workDir := ts.Getenv("WORK")
	text = strings.ReplaceAll(text, workDir, "$WORK")

	return text
}

// Scrub mock server address from output given that it uses a random port.
func scrubMockServerAddr(ts *testscript.TestScript, text string) string {
	addr := ts.Getenv("STUB_ADDR")
	if addr != "" {
		text = strings.ReplaceAll(text, addr, "$STUB_ADDR")
	}

	return text
}

// Scrub empty frames from the start of a charmbracelet/vhs recording.
func scrubVHS(recording string) string {
	separator := "────────────────────────────────────────────────────────────────────────────────"
	frames := strings.Split(recording, separator)

	firstFrame := 0

	for _, frame := range frames {
		// Clean the block to see if it's "empty"
		// We remove the shell prompt '>' and all whitespace
		contentCheck := strings.TrimSpace(frame)
		contentCheck = strings.TrimPrefix(contentCheck, ">")
		contentCheck = strings.TrimSpace(contentCheck)

		// If we haven't found the app yet and this block is still just whitespace/prompt, skip it
		if contentCheck != "" {
			break
		}

		firstFrame++
	}

	// Join the kept blocks back together
	return strings.Join(frames[firstFrame:], separator)
}
