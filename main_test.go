package main_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var update = flag.Bool("update", false, "update golden files") //nolint:gochecknoglobals

type mockHandler struct {
	httpMethod        string
	requestURIPattern *regexp.Regexp
	statusCode        int
	dumpRequest       func(*http.Request)
	responseBody      []byte
}

func TestCommands(t *testing.T) {
	t.Parallel()

	mocks := []mockHandler{}

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
				uri := r.URL.RequestURI()
				for _, mock := range mocks {
					if mock.requestURIPattern.MatchString(uri) {
						mock.dumpRequest(r)
						w.WriteHeader(mock.statusCode)

						if mock.responseBody != nil {
							w.Write(mock.responseBody)
						}

						return
					}
				}

				env.T().Fatal(fmt.Sprintf("Not implemented: %s %s", r.Method, uri))
			}))

			addr := strings.Replace(srv.URL, "http://", "", 1)
			env.Setenv("MOCKSERVER_ADDR", addr)

			return nil
		},
	}

	testscript.Run(t, params)
}

func mock(mocks *[]mockHandler) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, _ bool, args []string) {
		if len(args) < 3 {
			ts.Fatalf("usage: mock httpMethod requestURIPattern mockResponseStatus [mockResponseBodyPath] [requestOutPath]")
		}

		method := args[0]

		requestURIPattern, err := regexp.Compile(args[1])
		if err != nil {
			ts.Fatalf("Error compiling regex pattern: %v", err)
		}

		statusCode, err := strconv.Atoi(args[2])
		if err != nil {
			ts.Fatalf("Error parsing status code: %v", err)
		}

		var body []byte
		if len(args) > 3 {
			body = []byte(ts.ReadFile(args[3]))
		}

		*mocks = append([]mockHandler{{
			httpMethod:        method,
			requestURIPattern: requestURIPattern,
			statusCode:        statusCode,
			dumpRequest: func(r *http.Request) {
				if len(args) < 5 {
					return
				}

				path := ts.MkAbs(args[4])

				dump, err := httputil.DumpRequest(r, false)
				if err != nil {
					ts.Fatalf("Failed to dump HTTP request: %v", err)
				}

				var body []byte
				if r.Header.Get("Content-Type") == "application/json" {
					var data map[string]any

					err = json.NewDecoder(r.Body).Decode(&data)
					if err != nil {
						ts.Fatalf("Failed to unmarshal JSON request body: %v", err)
					}

					body, err = json.MarshalIndent(data, "", "  ")
					if err != nil {
						ts.Fatalf("Failed to re-marshal JSON request body: %v", err)
					}
				} else if r.Body != nil {
					body, err = io.ReadAll(r.Body)
					if err != nil {
						ts.Fatalf("Failed to read request body: %v", err)
					}
				}

				if body != nil {
					r.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore request body so it can be read again
					dump = slices.Concat(dump, body)
				}

				dump = []byte(scrubMockServerAddr(ts, string(dump)))

				// httputil.DumpRequest uses \r\n per RFC9112 spec but txtar parses these as newlines
				dump = []byte(strings.ReplaceAll(string(dump), "\r\n", "\n"))

				// txtar file is parsed with trailling newline - we add one here to avoid cmp failures
				dump = append(dump, '\n')

				err = os.WriteFile(path, dump, 0600)
				if err != nil {
					ts.Fatalf("Failed to write request to file: %v", err)
				}
			},
			responseBody: body,
		}}, *mocks...)
	}
}

func scrub(ts *testscript.TestScript, _ bool, args []string) {
	scrubbed := ts.ReadFile(args[0])
	scrubbed = scrubWorkDir(ts, scrubbed)
	scrubbed = scrubVHS(scrubbed)
	scrubbed = scrubMockServerAddr(ts, scrubbed)
	scrubbed = trimTrailingWhitespace(scrubbed)

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
	url := ts.Getenv("MOCKSERVER_ADDR")
	if url != "" {
		text = strings.ReplaceAll(text, url, "$MOCKSERVER_ADDR")
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
