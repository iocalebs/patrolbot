package mediawiki_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestLogEventsPaginator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string   // test case name
		pages       [][]byte // mock API responses - each call returns the next page
		wantResults bool     // set to true if pages are expected to have actual results
		wantErr     error    // expected error value, or nil if no error expected
	}{
		{
			name: "ValidPages",
			pages: [][]byte{
				readFile(t, "testdata/mwlogevents1.json"),
				readFile(t, "testdata/mwlogevents2.json"),
				readFile(t, "testdata/mwlogevents3.json"),
			},
			wantResults: true,
			wantErr:     nil,
		},
		{
			name: "ResponseError",
			pages: [][]byte{
				readFile(t, "testdata/mwlogevents_errors.json"),
			},
			wantResults: false,
			wantErr: &mediawiki.APIError{
				Errors: []string{
					"Unrecognized value for parameter \"letype\": foo.",
				},
				Warnings: []string{},
			},
		},
		{
			name: "ResponseWarnings",
			pages: [][]byte{
				readFile(t, "testdata/mwlogevents_warnings.json"),
			},
			wantResults: true,
			wantErr: &mediawiki.APIError{
				Errors: []string{},
				Warnings: []string{
					"Unrecognized value for parameter \"leprop\": foo",
				},
			},
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

				if test.wantResults && len(page.LogEvents) == 0 {
					t.Fatalf("No results in page %d", pageCount+1)
				}

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
					t.Fatalf("Unexpected error querying for nth page: %d", pageCount+1)
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
		w.Write([]byte("{}"))
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
		w.Write([]byte("{}"))
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

	expectedQuery := "action=query&errorformat=plaintext&format=json&formatversion=2&list=logevents&uselang=user"
	if url.RawQuery != expectedQuery {
		t.Errorf("unexpected URL query string: got %q, want %q", url.RawQuery, expectedQuery)
	}
}
