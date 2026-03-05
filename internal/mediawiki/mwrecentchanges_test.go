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

func mockServer(t *testing.T, pages [][]byte) *httptest.Server {
	t.Helper()

	curPage := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if curPage >= len(pages) {
			t.Fatalf("Requested invalid nth page: %d", curPage+1)
		}

		page := pages[curPage]
		if page != nil {
			w.Write(pages[curPage])
		}

		curPage++
	}))

	return srv
}

func TestRecentChangesPaginator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string   // test case name
		statusCode  int      // HTTP status code to return in mock API response
		pages       [][]byte // mock API responses - each call returns the next page
		wantResults bool     // set to true if pages are expected to have actual results
		wantErr     error    // expected error value, or nil if no error expected
	}{
		{
			name:       "ValidPages",
			statusCode: 200,
			pages: [][]byte{
				readFile(t, "testdata/mwrecentchanges_unpatrolled1.json"),
				readFile(t, "testdata/mwrecentchanges_unpatrolled2.json"),
				readFile(t, "testdata/mwrecentchanges_unpatrolled3.json"),
			},
			wantResults: true,
			wantErr:     nil,
		},
		{
			name:       "ResponseError",
			statusCode: 200,
			pages: [][]byte{
				readFile(t, "testdata/mwrecentchanges_errors.json"),
			},
			wantResults: false,
			wantErr: &mediawiki.APIError{
				Errors: []string{
					"Incorrect parameter - mutually exclusive values may not be supplied.",
				},
				Warnings: []string{},
			},
		},
		{
			name:       "ResponseWarnings",
			statusCode: 200,
			pages: [][]byte{
				readFile(t, "testdata/mwrecentchanges_warnings.json"),
			},
			wantResults: true,
			wantErr: &mediawiki.APIError{
				Errors: []string{},
				Warnings: []string{
					"Unrecognized value for parameter \"rcshow\": foo",
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
			paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{})

			expectedPageCount := len(test.pages)
			for pageCount := 0; paginator.HasMorePages(); pageCount++ {
				if pageCount > expectedPageCount {
					t.Fatalf("Paginator has more than the %d pages expected", expectedPageCount)
				}

				page, err := paginator.NextPage(t.Context())

				if test.wantResults && len(page.RecentChanges) == 0 {
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

func TestRecentChangesQueryParametersAll(t *testing.T) {
	t.Parallel()

	var gotURL *url.URL

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL

	time := time.Date(2026, 01, 13, 0, 0, 0, 0, time.UTC)
	mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{
		RCStart: time,
		RCEnd:   time,
		RCDir:   "newer",
		RCProp:  []string{"timestamp", "title"},
		RCUser:  "Dany36",
		RCShow:  "foo",
		RCLimit: 10,
	})

	_, err := paginator.NextPage(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	want := url.Values{
		"action":        []string{"query"},
		"errorformat":   []string{"plaintext"},
		"format":        []string{"json"},
		"formatversion": []string{"2"},
		"list":          []string{"recentchanges"},
		"rcdir":         []string{"newer"},
		"rcend":         []string{"2026-01-13T00:00:00Z"},
		"rclimit":       []string{"10"},
		"rcprop":        []string{"timestamp|title"},
		"rcuser":        []string{"Dany36"},
		"rcshow":        []string{"foo"},
		"rcstart":       []string{"2026-01-13T00:00:00Z"},
		"uselang":       []string{"user"},
	}

	diff := cmp.Diff(want, gotURL.Query())
	if diff != "" {
		t.Errorf("Query parameters mismatch (-want +got):\n%s", diff)
	}
}

func TestRecentChangesQueryParametersNone(t *testing.T) {
	t.Parallel()

	var gotURL *url.URL

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL

		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	cfg := config.Wiki{}
	cfg.Site.URL = srv.URL

	mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, mediawiki.RecentChangesQueryParams{})

	_, err := paginator.NextPage(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	want := url.Values{
		"action":        []string{"query"},
		"errorformat":   []string{"plaintext"},
		"format":        []string{"json"},
		"formatversion": []string{"2"},
		"list":          []string{"recentchanges"},
		"uselang":       []string{"user"},
	}

	diff := cmp.Diff(want, gotURL.Query())
	if diff != "" {
		t.Errorf("Query parameters mismatch (-want +got):\n%s", diff)
	}
}
