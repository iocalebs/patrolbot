package mediawiki_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string // test case name
		statusCode          int    // HTTP status code to return in mock API response
		responseBody        []byte // mock API response body
		expectedError       error  // expected error value, or nil if no error expected
		expectedErrorSubstr string // expected error message substring
	}{
		{

			name:                "200_Success",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwlogin_success.json"),
			expectedError:       nil,
			expectedErrorSubstr: "",
		},
		{
			name:                "200_FailedWrongToken",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwlogin_failed_wrongtoken.json"),
			expectedError:       mediawiki.ErrLoginFailed,
			expectedErrorSubstr: "WrongToken",
		},
		{
			name:                "200_WarningResponse",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwaction_warnings.json"),
			expectedError:       mediawiki.ErrResponseWarnings,
			expectedErrorSubstr: "Unrecognized value for parameter",
		},
		{
			name:                "200_WarningError",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwaction_error.json"),
			expectedError:       mediawiki.ErrResponseError,
			expectedErrorSubstr: "Unrecognized value for parameter",
		},
		{
			name:                "500_WithBody",
			statusCode:          http.StatusInternalServerError,
			responseBody:        []byte("An error occurred"),
			expectedError:       mediawiki.ErrResponseStatusCode,
			expectedErrorSubstr: "An error occurred",
		},
		{
			name:                "502_WithoutBody",
			statusCode:          http.StatusBadGateway,
			responseBody:        nil,
			expectedError:       mediawiki.ErrResponseStatusCode,
			expectedErrorSubstr: "empty response body",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)

				if test.responseBody != nil {
					w.Write(test.responseBody) //nolint:errcheck,gosec
				}
			}))
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL

			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, *slog.Default())

			err := mwclient.Login(context.Background(), "token")

			if test.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", test.expectedError)
				}

				if !errors.Is(err, test.expectedError) {
					t.Fatalf("Expected error %v, got %v", test.expectedError, err)
				}

				if !strings.Contains(err.Error(), test.expectedErrorSubstr) {
					t.Fatalf("Expected error message to contain %q, got %q", test.expectedErrorSubstr, err.Error())
				}
			}
		})
	}
}

func TestLoginRequestHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("{}")) //nolint:errcheck,gosec
	}))
	defer srv.Close()

	userAgent := "myAgent"
	from := "foo@example.test"

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL
	cfg.Client.UserAgent = userAgent
	cfg.Client.From = from

	httpClient, transport := newHTTPClient()
	mwclient := mediawiki.NewClient(cfg, httpClient, *slog.Default())

	mwclient.Login(context.Background(), "token") //nolint:errcheck,gosec

	gotUserAgent := transport.lastRequest.Header.Get("User-Agent")
	if gotUserAgent != userAgent {
		t.Errorf("Expected User-Agent header %q, got %q", userAgent, gotUserAgent)
	}

	gotFrom := transport.lastRequest.Header.Get("From")
	if gotFrom != from {
		t.Errorf("Expected From header %q, got %q", userAgent, gotUserAgent)
	}
}
