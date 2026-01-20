package main_test

import (
	"bytes"
	"flag"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

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
			"sleep": sleep,
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

func sleep(ts *testscript.TestScript, _ bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("usage: sleep duration")
	}

	d, err := time.ParseDuration(args[0])
	if err != nil {
		ts.Fatalf("invalid duration: %v", err)
	}

	time.Sleep(d)
}

func scrub(ts *testscript.TestScript, _ bool, args []string) {
	file := args[0]
	scrubbed := ts.ReadFile(file)
	scrubbed = scrubEnv(ts, scrubbed)
	scrubbed = scrubTimestamps(scrubbed)
	scrubbed = scrubVHS(scrubbed)
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

// Replace environment variables with random element ($WORK, $STUB_ADDR) with placeholders
// Scrub mock server URL from output given that it uses a random port.
func scrubEnv(ts *testscript.TestScript, text string) string {
	env := []string{"STUB_ADDR", "WORK"}

	for _, envVar := range env {
		val := ts.Getenv(envVar)
		if val != "" {
			text = strings.ReplaceAll(text, val, "$"+envVar)
		}
	}

	return text
}

// Set all timestamps to epoch time.
func scrubTimestamps(text string) string {
	// Pattern for RFC3339: 2026-01-20T09:28:27.051219-05:00
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})`)
	epoch := time.Unix(0, 0).UTC().Format(time.RFC3339Nano)

	return re.ReplaceAllString(text, epoch)
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

func trimTrailingWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}

	return strings.Join(lines, "\n")
}
