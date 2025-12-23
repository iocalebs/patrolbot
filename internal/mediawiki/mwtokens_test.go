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

func TestLoginToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string          // test case name
		statusCode          int             // HTTP status code to return in mock API response
		responseBody        []byte          // mock API response body
		expectedToken       mediawiki.Token // expected token value
		expectedError       error           // expected error value, or nil if no error expected
		expectedErrorSubstr string          // expected error message substring
	}{
		{
			name:                "200_ValidLoginToken",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwtokens_login.json"),
			expectedToken:       "8af20f35764bee599652a5d5d9e804d469444fb1+\\",
			expectedError:       nil,
			expectedErrorSubstr: "",
		},
		{
			name:                "200_WarningResponse",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwaction_warnings.json"),
			expectedToken:       "",
			expectedError:       mediawiki.ErrResponseWarnings,
			expectedErrorSubstr: "Unrecognized value for parameter",
		},
		{
			name:                "200_WarningError",
			statusCode:          http.StatusOK,
			responseBody:        readFile(t, "testdata/mwaction_error.json"),
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

			cfg := config.Wiki{
				URL: srv.URL,
			}
			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, *slog.Default())
			token, err := mwclient.LoginToken(context.Background())

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

			if token != test.expectedToken {
				t.Fatalf("Invalid token value: got %q, want %q", token, test.expectedToken)
			}
		})
	}
}
