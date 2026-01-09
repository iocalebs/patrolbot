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
		name         string // test case name
		statusCode   int    // HTTP status code to return in mock API response
		responseBody []byte // mock API response body
		wantErr      error  // expected error value, or nil if no error expected
		wantErrSub   string // // expected error message substring
	}{
		{

			name:         "Success",
			statusCode:   http.StatusOK,
			responseBody: readFile(t, "testdata/mwlogin_success.json"),
			wantErr:      nil,
		},
		{
			name:         "FailedWrongToken",
			statusCode:   http.StatusOK,
			responseBody: readFile(t, "testdata/mwlogin_wrongtoken.json"),
			wantErr:      mediawiki.ErrLoginFailed,
			wantErrSub:   "WrongToken",
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

			err := mwclient.Login(context.Background(), "token")

			if test.wantErr != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", test.wantErr)
				}

				if !errors.Is(err, test.wantErr) {
					t.Fatalf("Expected error %v, got %v", test.wantErr, err)
				}

				if !strings.Contains(err.Error(), test.wantErrSub) {
					t.Fatalf("Expected error message to contain %q, got %q", test.wantErr, err.Error())
				}
			}
		})
	}
}
