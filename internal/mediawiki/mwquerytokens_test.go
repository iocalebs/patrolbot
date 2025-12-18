package mediawiki_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestLoginToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string          // test case name
		statusCode          int             // HTTP status code to return in mock API response
		responseBody        []byte          // name of file in testdata/ containing the mock API response
		expectedToken       mediawiki.Token // expected token value
		expectedError       error           // expected error value, or nil if no error expected
		expectedErrorSubstr string          // expected error message substring
	}{
		{
			name:                "200_ValidLoginToken",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwquerytokens_logintoken.json"),
			expectedToken:       "ddd90efd9edb06839437470c561b86b06943086f+\\",
			expectedError:       nil,
			expectedErrorSubstr: "",
		},
		{
			name:                "200_WarningResponse",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwquerytokens_warning.json"),
			expectedToken:       "",
			expectedError:       mediawiki.ErrResponseWarnings,
			expectedErrorSubstr: "Unrecognized value for parameter",
		},
		{
			name:                "200_WarningError",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwquerytokens_error.json"),
			expectedToken:       "",
			expectedError:       mediawiki.ErrResponseError,
			expectedErrorSubstr: "Unrecognized value for parameter",
		},
		{
			name:                "500_WithBody",
			statusCode:          http.StatusInternalServerError,
			responseBody:        []byte("An error occurred"),
			expectedToken:       "",
			expectedError:       mediawiki.ErrResponseStatusCode,
			expectedErrorSubstr: "An error occurred",
		},
		{
			name:                "502_WithoutBody",
			statusCode:          http.StatusBadGateway,
			responseBody:        nil,
			expectedToken:       "",
			expectedError:       mediawiki.ErrResponseStatusCode,
			expectedErrorSubstr: "empty response body",
		},
	}

	for _, tc := range tests { //nolint:varnamelen
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)

				if tc.responseBody != nil {
					w.Write(tc.responseBody) //nolint:errcheck,gosec
				}
			}))
			defer srv.Close()

			cfg := config.Wiki{ //nolint:exhaustruct
				APIURL: srv.URL,
			}
			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, *slog.Default())
			token, err := mwclient.LoginToken(context.Background())

			if tc.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tc.expectedError)
				}

				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("Expected error %v, got %v", tc.expectedError, err)
				}

				if !strings.Contains(err.Error(), tc.expectedErrorSubstr) {
					t.Fatalf("Expected error message to contain %q, got %q", tc.expectedErrorSubstr, err.Error())
				}
			}

			if token != tc.expectedToken {
				t.Fatalf("Invalid token value: got %q, want %q", token, tc.expectedToken)
			}
		})
	}
}

func readFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	return data
}
