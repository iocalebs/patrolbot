package mediawiki_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func mockServer(t *testing.T, statusCode int, pages [][]byte) *httptest.Server {
	t.Helper()

	curPage := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if curPage >= len(pages) {
			t.Fatalf("Requested invalid nth page: %d", curPage+1)
		}

		w.WriteHeader(statusCode)

		page := pages[curPage]
		if page != nil {
			w.Write(pages[curPage]) //nolint:errcheck,gosec
		}

		curPage++
	}))

	return srv
}

func TestRecentChangesPaginator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string   // test case name
		statusCode     int      // HTTP status code to return in mock API response
		pages          [][]byte // mock API responses - each call returns the next page
		expectResults  bool     // set to true if pages are expected to have actual results
		expectError    error    // expected error value, or nil if no error expected
		expectErrorSub string   // expected error message substring
	}{
		{
			name:       "ValidPages",
			statusCode: 200,
			pages: [][]byte{
				readFile(t, "testdata/mwrecentchanges_unpatrolled1.json"),
				readFile(t, "testdata/mwrecentchanges_unpatrolled2.json"),
				readFile(t, "testdata/mwrecentchanges_unpatrolled3.json"),
			},
			expectResults:  true,
			expectError:    nil,
			expectErrorSub: "",
		},
		{
			name:       "ResponseError",
			statusCode: 200,
			pages: [][]byte{
				readFile(t, "testdata/mwrecentchanges_error.json"),
			},
			expectResults:  false,
			expectError:    mediawiki.ErrResponseError,
			expectErrorSub: "Incorrect parameter - mutually exclusive values may not be supplied.",
		},
		{
			name:       "ResponseWarnings",
			statusCode: 200,
			pages: [][]byte{
				readFile(t, "testdata/mwrecentchanges_warnings.json"),
			},
			expectResults:  true,
			expectError:    mediawiki.ErrResponseWarnings,
			expectErrorSub: "Unrecognized value for parameter \\\"rcshow\\\": foo",
		},
		{
			name:       "500Error_WithBody",
			statusCode: http.StatusInternalServerError,
			pages: [][]byte{
				[]byte("An error occurred"),
			},
			expectResults:  false,
			expectError:    mediawiki.ErrResponseStatusCode,
			expectErrorSub: "An error occurred",
		},
		{
			name:       "502Error_WithoutBody",
			statusCode: http.StatusBadGateway,
			pages: [][]byte{
				nil,
			},
			expectResults:  false,
			expectError:    mediawiki.ErrResponseStatusCode,
			expectErrorSub: "empty response body",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := mockServer(t, test.statusCode, test.pages)
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL

			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, *slog.Default(), "")
			paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{})

			expectedPageCount := len(test.pages)
			for pageCount := 0; paginator.HasMorePages(); pageCount++ {
				if pageCount > expectedPageCount {
					t.Fatalf("Paginator has more than the %d pages expected", expectedPageCount)
				}

				page, err := paginator.NextPage(t.Context())

				if test.expectResults && len(page.RecentChanges) == 0 {
					t.Fatalf("No results in page %d", pageCount+1)
				}

				if test.expectError == nil && err != nil {
					t.Fatalf("Unexpected error querying for nth page: %d", pageCount+1)
				}

				if test.expectError != nil && !errors.Is(err, test.expectError) {
					t.Fatalf("Expected error %v, got %v", test.expectError, err)
				}
			}
		})
	}
}

func TestRecentChangesQueryParametersAll(t *testing.T) {
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

	mwclient := mediawiki.NewClient(cfg, http.DefaultClient, *slog.Default(), "")
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{
		RCStart: time.Now(),
		RCEnd:   time.Now(),
		RCShow:  "foo",
		RCLimit: 10,
	})

	_, err := paginator.NextPage(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	urlParams := url.Query()

	expectedParams := []string{"rcstart", "rcend", "rcshow", "rclimit"}
	for _, param := range expectedParams {
		if !urlParams.Has(param) {
			t.Errorf("Expected query URL to have param %q. URL: %s", param, url)
		}
	}
}

func TestRecentChangesQueryParametersNone(t *testing.T) {
	t.Parallel()

	var url *url.URL

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url = r.URL

		w.Write([]byte("{}")) //nolint:errcheck,gosec
	}))
	defer srv.Close()

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL

	mwclient := mediawiki.NewClient(cfg, http.DefaultClient, *slog.Default(), "")
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{})

	_, err := paginator.NextPage(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	expectedQuery := "action=query&format=json&formatversion=2&list=recentchanges"
	if url.RawQuery != expectedQuery {
		t.Errorf("Expected URL query %q, got %q", expectedQuery, url.RawQuery)
	}
}

func TestRecentChangesRequestHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("{}")) //nolint:errcheck,gosec
	}))
	defer srv.Close()

	userAgent := "myAgent"
	from := "foo@example.test"

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL
	cfg.Client.From = from

	httpClient, transport := newHTTPClient()
	mwclient := mediawiki.NewClient(cfg, httpClient, *slog.Default(), userAgent)
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{})

	paginator.NextPage(t.Context()) //nolint:errcheck,gosec

	gotUserAgent := transport.lastRequest.Header.Get("User-Agent")
	if gotUserAgent != userAgent {
		t.Errorf("Expected User-Agent header %q, got %q", userAgent, gotUserAgent)
	}

	gotFrom := transport.lastRequest.Header.Get("From")
	if gotFrom != from {
		t.Errorf("Expected From header %q, got %q", userAgent, gotUserAgent)
	}
}
