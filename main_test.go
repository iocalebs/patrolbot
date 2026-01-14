package main_test

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var update = flag.Bool("update", false, "update golden files") //nolint:gochecknoglobals

type mockResponse struct {
	requestURIPattern *regexp.Regexp
	statusCode        int
	responseBody      []byte
}

func TestCommands(t *testing.T) {
	t.Parallel()

	mocks := []mockResponse{}

	params := testscript.Params{
		Dir:           "testdata/integration-tests",
		UpdateScripts: *update,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"mock":  mock(&mocks),
			"scrub": scrub,
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

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for _, mock := range mocks {
					if mock.requestURIPattern.MatchString(r.URL.RequestURI()) {
						w.WriteHeader(mock.statusCode)
						w.Write(mock.responseBody)

						return
					}
				}

				w.WriteHeader(http.StatusNotImplemented)
			}))
			env.Setenv("MOCK_SERVER_URL", srv.URL)

			return nil
		},
	}

	testscript.Run(t, params)
}

func mock(mocks *[]mockResponse) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, _ bool, args []string) {
		re, err := regexp.Compile(args[0])
		if err != nil {
			ts.Fatalf("Error compiling regex pattern: %v", err)
		}

		statusCode, err := strconv.Atoi(args[1])
		if err != nil {
			ts.Fatalf("Error parsing status code: %v", err)
		}

		*mocks = append([]mockResponse{{
			requestURIPattern: re,
			statusCode:        statusCode,
			responseBody:      []byte(ts.ReadFile(args[2])),
		}}, *mocks...)
	}
}

func scrub(ts *testscript.TestScript, _ bool, args []string) {
	scrubbed := ts.ReadFile(args[0])
	scrubbed = scrubWorkDir(scrubbed, ts.Getenv("WORK"))
	scrubbed = scrubVHS(scrubbed)
	scrubbed = scrubMockServerURL(ts, scrubbed)

	var outPath string
	if len(args) > 1 {
		outPath = ts.MkAbs(args[1])
	} else {
		outPath = ts.MkAbs(args[0])
	}

	err := os.WriteFile(outPath, []byte(scrubbed), 0600)
	if err != nil {
		ts.Fatalf("scrub: %v", err)
	}
}

// Replaces $WORK value (a random test dir) with placeholder so that tests that output subdirs can use the
// `cmp` testscript command. `cmpenv` is not ideal as it cannot update golden files and can expand
// text that is not actually an env variable (e.g. /$1 as used in `patrolbot init` help).
func scrubWorkDir(file string, workDir string) string {
	// Need to account for newlines wrapping the path, due to how fang wraps error messages.
	workDir = regexp.QuoteMeta(workDir)
	rePattern := strings.Join(strings.Split(workDir, ""), `(?:\s+|\n\s*)?`)
	re := regexp.MustCompile(rePattern)

	scrubbed := re.ReplaceAllString(file, "$$WORK")

	// Fixes test failures caused by whitespace differences
	lines := strings.Split(scrubbed, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}

	scrubbed = strings.Join(lines, "\n")

	return scrubbed
}

// Scrub mock server URL from output given that it uses a random port.
func scrubMockServerURL(ts *testscript.TestScript, text string) string {
	url := ts.Getenv("MOCK_SERVER_URL")
	if url != "" {
		text = strings.ReplaceAll(text, url, "$MOCK_SERVER_URL")
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
