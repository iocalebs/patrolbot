package mediawiki_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string           // test case name
		statusCode   int              // HTTP status code to return in mock API response
		responseBody []byte           // mock API response body
		wantTokens   mediawiki.Tokens // expected Tokens value
		wantErr      error            // expected error value, or nil if no error expected
	}{
		{
			name:         "ValidLoginToken",
			statusCode:   http.StatusOK,
			responseBody: readFile(t, "testdata/mwtokens_login.json"),
			wantTokens: mediawiki.Tokens{
				Login: "8af20f35764bee599652a5d5d9e804d469444fb1+\\",
			},
			wantErr: nil,
		},
		{
			name:         "WarningResponse",
			statusCode:   http.StatusOK,
			responseBody: readFile(t, "testdata/mwtokens_warnings.json"),
			wantTokens:   mediawiki.Tokens{},
			wantErr: &mediawiki.APIError{
				Errors: []string{},
				Warnings: []string{
					"Unrecognized value for parameter \"type\": foo",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)

				if test.responseBody != nil {
					w.Write(test.responseBody)
				}
			}))
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL

			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
			tokens, err := mwclient.Tokens(context.Background(), "login")

			if test.wantErr != nil {
				wantType := reflect.TypeOf(test.wantErr)
				if wantType != reflect.TypeOf(err) {
					t.Fatalf("got error of type %T, want %T", err, test.wantErr)
				}

				diff := cmp.Diff(test.wantErr, err)
				if diff != "" {
					t.Fatalf("Error mismatch (-want +got):\n%s", diff)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			diff := cmp.Diff(test.wantTokens, tokens)
			if diff != "" {
				t.Fatalf("Tokens() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
