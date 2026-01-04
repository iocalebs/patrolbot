package mediawiki_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestLogEventsPaginator(t *testing.T) { //nolint:cyclop
	t.Parallel()

	tests := []struct {
		name           string   // test case name
		pages          [][]byte // mock API responses - each call returns the next page
		expectResults  bool     // set to true if pages are expected to have actual results
		expectError    error    // expected error value, or nil if no error expected
		expectErrorSub string   // expected error message substring
	}{
		{
			name: "ValidPages",
			pages: [][]byte{
				readFile(t, "testdata/mwlogevents1.json"),
				readFile(t, "testdata/mwlogevents2.json"),
				readFile(t, "testdata/mwlogevents3.json"),
			},
			expectResults:  true,
			expectError:    nil,
			expectErrorSub: "",
		},
		{
			name: "ResponseError",
			pages: [][]byte{
				readFile(t, "testdata/mwlogevents_error.json"),
			},
			expectResults:  false,
			expectError:    mediawiki.ErrResponseError,
			expectErrorSub: "Unrecognized value for parameter \\\"letype\\\": foo.",
		},
		{
			name: "ResponseWarnings",
			pages: [][]byte{
				readFile(t, "testdata/mwlogevents_warnings.json"),
			},
			expectResults:  true,
			expectError:    mediawiki.ErrResponseWarnings,
			expectErrorSub: "Unrecognized value for parameter \\\"leprop\\\": foo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := mockServer(t, test.pages)
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL

			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
			paginator := mediawiki.NewLogEventsPaginator(mwclient, mediawiki.LogEventsQueryParams{})

			expectedPageCount := len(test.pages)
			for pageCount := 0; paginator.HasMorePages(); pageCount++ {
				if pageCount > expectedPageCount {
					t.Fatalf("Paginator has more than the %d pages expected", expectedPageCount)
				}

				page, err := paginator.NextPage(t.Context())

				if test.expectResults && len(page.LogEvents) == 0 {
					t.Fatalf("No results in page %d", pageCount+1)
				}

				if test.expectError == nil && err != nil {
					t.Fatalf("Unexpected error querying for nth page: %d", pageCount+1)
				}

				if test.expectError != nil && !errors.Is(err, test.expectError) {
					t.Fatalf("Expected error %v, got %v", test.expectError, err)
				}

				if test.expectError != nil && !strings.Contains(err.Error(), test.expectErrorSub) {
					t.Fatalf("Expected error to contain %q, got %q", test.expectErrorSub, err.Error())
				}
			}
		})
	}
}

func TestLogEventsQueryParametersAll(t *testing.T) {
	t.Parallel()

	var url *url.URL

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url = r.URL

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}")) //nolint:errcheck,gosec
	}))
	defer srv.Close()

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL

	mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
	paginator := mediawiki.NewLogEventsPaginator(mwclient, mediawiki.LogEventsQueryParams{
		LEStart: time.Now(),
		LEEnd:   time.Now(),
		LEProp:  "foo",
		LEType:  "bar",
		LELimit: 10,
	})

	_, err := paginator.NextPage(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	urlParams := url.Query()

	expectedParams := []string{"lestart", "leend", "leprop", "letype", "lelimit"}
	for _, param := range expectedParams {
		if !urlParams.Has(param) {
			t.Errorf("Expected query URL to have param %q. URL: %s", param, url)
		}
	}
}

func TestLogEventsQueryParametersNone(t *testing.T) {
	t.Parallel()

	var url *url.URL

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url = r.URL

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}")) //nolint:errcheck,gosec
	}))
	defer srv.Close()

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL
	mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
	paginator := mediawiki.NewLogEventsPaginator(mwclient, mediawiki.LogEventsQueryParams{})

	_, err := paginator.NextPage(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	expectedQuery := "action=query&format=json&formatversion=2&list=logevents"
	if url.RawQuery != expectedQuery {
		t.Errorf("Expected URL query %q, got %q", expectedQuery, url.RawQuery)
	}
}
