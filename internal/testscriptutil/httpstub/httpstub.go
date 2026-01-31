// Package httpstub provides functionality for stubbing an HTTP server and spying on requests in testscript tests.
package httpstub

import (
	"bytes"
	"encoding/json"
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

	"github.com/rogpeppe/go-internal/testscript"
)

// Server represents an [httptest.Server] that can dump requests and respond with stubs provided via testscript command.
type Server struct {
	Addr  string // Server address (e.g. 127.0.0.1:8080)
	stubs []stub
}

type stub struct {
	httpMethod        string
	requestURIPattern *regexp.Regexp
	statusCode        int
	responseBody      []byte
	requestOutPath    string
	ts                *testscript.TestScript
}

type serverKey struct{}

// InitServer initializes a [Server] in the testscript environment.
func InitServer(env *testscript.Env, mutateDump func([]byte) []byte) *Server {
	srv := &Server{
		stubs: []stub{},
	}

	httpsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.RequestURI()
		for _, stub := range srv.stubs {
			if r.Method == stub.httpMethod && stub.requestURIPattern.MatchString(uri) {
				if stub.requestOutPath != "" {
					dumpRequest(stub.ts, r, stub.requestOutPath, mutateDump)
				}

				w.WriteHeader(stub.statusCode)

				if stub.responseBody != nil {
					_, err := w.Write(stub.responseBody)
					if err != nil {
						env.T().Fatal(fmt.Sprintf("Failed to write stub response: %v", err))
					}
				}

				return
			}
		}

		env.T().Fatal(fmt.Sprintf("Not implemented: %s %s", r.Method, uri))
	}))

	srv.Addr = strings.Replace(httpsrv.URL, "http://", "", 1)

	env.Values[serverKey{}] = srv

	return srv
}

// Cmd returns a [github.com/rogpeppe/go-internal/testscript] command for registering HTTP stubs and dumping
// corresponding HTTP requests to files.
//
// The stub server must be initialized with [InitServer] in the testscript.Params.Setup.
func Cmd(cmdName string) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, _ bool, args []string) {
		if len(args) < 3 {
			ts.Fatalf(
				"usage: %s method requestURIPattern stubResponseStatus [stubResponseBodyPath] [requestOutPath]",
				cmdName,
			)
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

		var requestOutPath string
		if len(args) > 4 {
			requestOutPath = ts.MkAbs(args[4])
		}

		srv, ok := ts.Value(serverKey{}).(*Server)
		if !ok {
			ts.Fatalf("httpstub server not initialized in Setup")
		}

		srv.addStub(stub{
			httpMethod:        method,
			requestURIPattern: requestURIPattern,
			statusCode:        statusCode,
			responseBody:      body,
			requestOutPath:    requestOutPath,
			ts:                ts,
		})
	}
}

func (s *Server) addStub(stb stub) {
	// Stubs are prepended so that the latest stub is matched first
	s.stubs = append([]stub{stb}, s.stubs...)
}

func dumpRequest(ts *testscript.TestScript, r *http.Request, path string, mutateDump func([]byte) []byte) {
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

	// httputil.DumpRequest uses \r\n per RFC9112 spec but txtar parses these as newlines
	dump = bytes.ReplaceAll(dump, []byte("\r\n"), []byte("\n"))

	// txtar file is parsed with trailling newline - we add one here to avoid cmp failures
	dump = append(dump, '\n')

	if mutateDump != nil {
		dump = mutateDump(dump)
	}

	err = os.WriteFile(path, dump, 0600)
	if err != nil {
		ts.Fatalf("Failed to write request to file: %v", err)
	}
}
