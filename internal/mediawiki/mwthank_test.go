package mediawiki_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestThank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		params       mediawiki.ThankParams
		statusCode   int
		responseBody string
		wantQuery    url.Values
		wantForm     url.Values
		wantErr      error
	}{
		{
			name:         "Revision",
			params:       mediawiki.ThankParams{RevisionID: 42},
			statusCode:   http.StatusOK,
			responseBody: `{}`,
			wantQuery:    thankQuery(),
			wantForm: url.Values{
				"rev":    {"42"},
				"source": {"patrolbot"},
				"token":  {"csrftoken"},
			},
		},
		{
			name:         "Log",
			params:       mediawiki.ThankParams{LogID: 99},
			statusCode:   http.StatusOK,
			responseBody: `{}`,
			wantQuery:    thankQuery(),
			wantForm: url.Values{
				"log":    {"99"},
				"source": {"patrolbot"},
				"token":  {"csrftoken"},
			},
		},
		{
			name:         "APIError",
			params:       mediawiki.ThankParams{RevisionID: 42},
			statusCode:   http.StatusOK,
			responseBody: `{"errors":[{"text":"The recipient cannot be thanked"}]}`,
			wantQuery:    thankQuery(),
			wantForm: url.Values{
				"rev":    {"42"},
				"source": {"patrolbot"},
				"token":  {"csrftoken"},
			},
			wantErr: &mediawiki.APIError{
				Errors:   []string{"The recipient cannot be thanked"},
				Warnings: []string{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv, gotQuery, gotForm := newThankServer(t, test.statusCode, test.responseBody)
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL

			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
			err := mwclient.Thank(context.Background(), "csrftoken", test.params)

			assertThankErr(t, err, test.wantErr)

			diff := cmp.Diff(test.wantQuery, *gotQuery)
			if diff != "" {
				t.Fatalf("query mismatch (-want +got):\n%s", diff)
			}

			diff = cmp.Diff(test.wantForm, *gotForm)
			if diff != "" {
				t.Fatalf("form mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func thankQuery() url.Values {
	return url.Values{
		"action":        {"thank"},
		"errorformat":   {"plaintext"},
		"format":        {"json"},
		"formatversion": {"2"},
		"uselang":       {"user"},
	}
}

func newThankServer(t *testing.T, statusCode int, responseBody string) (*httptest.Server, *url.Values, *url.Values) {
	t.Helper()

	var (
		gotQuery url.Values
		gotForm  url.Values
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}

		gotQuery = r.URL.Query()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		gotForm, err = url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))

	return srv, &gotQuery, &gotForm
}

func assertThankErr(t *testing.T, gotErr error, wantErr error) {
	t.Helper()

	if wantErr == nil {
		if gotErr != nil {
			t.Fatalf("unexpected error: %v", gotErr)
		}

		return
	}

	wantType := reflect.TypeOf(wantErr)
	if wantType != reflect.TypeOf(gotErr) {
		t.Fatalf("got error of type %T, want %T", gotErr, wantErr)
	}

	diff := cmp.Diff(wantErr, gotErr)
	if diff != "" {
		t.Fatalf("Error mismatch (-want +got):\n%s", diff)
	}
}
