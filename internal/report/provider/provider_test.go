package provider_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/clock"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	"github.com/iocalebs/patrolbot/internal/report/data"
	"github.com/iocalebs/patrolbot/internal/report/provider"
)

func TestProvider_Data_Config(t *testing.T) {
	t.Parallel()

	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:          "https://zeldawiki.wiki",
			ArticlePath:  "/wiki",
			ScriptPath:   "/w",
			RCMaxAgeDays: 90,
			RCLinkLimit:  1000,
		},
		Auth: config.WikiAuth{
			Username: "username",
			Password: "password",
		},
	}

	provider := provider.New(slog.Default(), nil, nil, wiki)
	got := provider.Data(t.Context(), config.ReportData{})

	if got.Config.Auth.Password != "" {
		t.Fatalf("Expected password to be redacted from config in report data, got %q", got.Config.Auth.Password)
	}

	if wiki.Auth.Password == "" {
		t.Fatal("Expected original config object to not be mutated")
	}

	want := data.ReportData{
		Config: wiki,
	}
	want.Config.Auth.Password = ""

	diff := cmp.Diff(want, got)
	if diff != "" {
		t.Fatalf("result mismatch (-want +got):\n%s", diff)
	}
}

func TestProvider_Data_NoMediaWiki(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:        srv.URL,
			ScriptPath: "/w",
		},
	}
	mwclient := mediawiki.NewClient(wiki, http.DefaultClient, slog.Default(), "")
	provider := provider.New(slog.Default(), clock.Func(time.Now), mwclient, wiki)

	got := provider.Data(t.Context(), config.ReportData{
		PatrolExpiring: nil,
	})

	if got.Error != "" {
		t.Fatalf(
			"Expected no MediaWiki API call when not needed for report, but got API-related error: %v",
			got.Error,
		)
	}
}

func TestProvider_Data_TokensError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("an error occurred"))
	}))
	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:        srv.URL,
			ScriptPath: "/w",
		},
	}
	mwclient := mediawiki.NewClient(wiki, http.DefaultClient, slog.Default(), "")

	var logbuf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&logbuf, nil))

	provider := provider.New(logger, clock.Func(time.Now), mwclient, wiki)
	reportData := provider.Data(t.Context(), config.ReportData{
		PatrolExpiring: &config.PatrolExpiring{},
	})

	want := "Error logging into MediaWiki: 500 Internal Server Error"
	if reportData.Error != want {
		t.Errorf("Unexpected error message:\ngot: %q\nwant: %q", reportData.Error, want)
	}

	want = "msg=\"Error obtaining login token from MediaWiki API:Tokens\" " +
		"err=\"500 Internal Server Error\" statusCode=500 body=\"an error occurred\""
	if !strings.Contains(logbuf.String(), want) {
		t.Errorf("Log output did not contain the expected error log.\ngot: %q\nwant: %q", logbuf.String(), want)
	}
}

func TestProvider_Data_LoginError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("action") == "query" && q.Get("meta") == "tokens" && q.Get("type") == "login":
			w.Write([]byte(`
					{
						"query": {
							"tokens": {
								"logintoken": "8af20f35764bee599652a5d5d9e804d469444fb1+\\"
							}
						}
					}
				`))
		case q.Get("action") == "login":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("an error occurred"))
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))

	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:        srv.URL,
			ScriptPath: "/w",
		},
	}
	mwclient := mediawiki.NewClient(wiki, http.DefaultClient, slog.Default(), "")

	var logbuf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&logbuf, nil))

	provider := provider.New(logger, clock.Func(time.Now), mwclient, wiki)
	reportData := provider.Data(t.Context(), config.ReportData{
		PatrolExpiring: &config.PatrolExpiring{},
	})

	want := "Error logging into MediaWiki: 500 Internal Server Error"
	if reportData.Error != want {
		t.Errorf("Unexpected error message:\ngot: %q\nwant: %q", reportData.Error, want)
	}

	want = "msg=\"Error executing login via MediaWiki API:Login\" " +
		"err=\"500 Internal Server Error\" statusCode=500 body=\"an error occurred\""
	if !strings.Contains(logbuf.String(), want) {
		t.Errorf("Log output did not contain the expected error log.\ngot: %q\nwant: %q", logbuf.String(), want)
	}
}

func TestProvider_Data_RecentChangesError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("action") == "query" && q.Get("meta") == "tokens" && q.Get("type") == "login":
			w.Write([]byte(`
					{
						"query": {
							"tokens": {
								"logintoken": "8af20f35764bee599652a5d5d9e804d469444fb1+\\"
							}
						}
					}
				`))
		case r.FormValue("action") == "login":
			w.Write([]byte(`
					{
						"login": {
							"result": "Success",
							"lguserid": 45359339,
							"lgusername": "PhantomCaleb"
						}
					}
				`))
		default:
			w.Write([]byte(`
				{
					"errors": [
						{
							"code": "show",
							"text": "Incorrect parameter - mutually exclusive values may not be supplied.",
							"module": "query+recentchanges"
						}
					]
				}
			`))
		}
	}))
	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:          srv.URL,
			ScriptPath:   "/w",
			RCMaxAgeDays: 30,
		},
	}
	mwclient := mediawiki.NewClient(wiki, http.DefaultClient, slog.Default(), "")

	var logbuf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&logbuf, nil))

	provider := provider.New(logger, clock.Func(time.Now), mwclient, wiki)
	reportData := provider.Data(t.Context(), config.ReportData{
		PatrolExpiring: &config.PatrolExpiring{
			Windows: []config.Duration{
				{Days: 1},
			},
		},
	})

	want := "Error executing API:RecentChanges query: " +
		"\n- Incorrect parameter - mutually exclusive values may not be supplied."
	if reportData.Error != want {
		t.Errorf("Unexpected error message:\ngot: %q\nwant: %q", reportData.Error, want)
	}

	want = "msg=\"Error executing API:RecentChanges query\" " +
		"err=\"\\n- Incorrect parameter - mutually exclusive values may not be supplied.\"\n"
	if !strings.Contains(logbuf.String(), want) {
		t.Errorf("Log output did not contain the expected error log.\ngot: %q\nwant: %q", logbuf.String(), want)
	}
}

func TestProvider_Data_PatrolExpiring(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("action") == "query" && q.Get("meta") == "tokens" && q.Get("type") == "login":
			w.Write([]byte(`
				{
					"query": {
						"tokens": {
							"logintoken": "8af20f35764bee599652a5d5d9e804d469444fb1+\\"
						}
					}
				}
			`))
		case q.Get("action") == "login":
			w.Write([]byte(`
				{
					"login": {
						"result": "Success",
						"lguserid": 45359339,
						"lgusername": "PhantomCaleb"
					}
				}
			`))
		case q.Get("action") == "query" && q.Get("list") == "recentchanges":
			w.Write([]byte(`
				{
					"query": {
						"recentchanges": [
							{
								"timestamp": "2025-12-01T05:00:00Z"
							},
							{
								"timestamp": "2025-12-02T04:00:00Z"
							},
							{
								"timestamp": "2025-12-08T04:00:00Z"
							}
						]
					}
				}
			`))
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:          srv.URL,
			ScriptPath:   "/w",
			RCMaxAgeDays: 30,
		},
	}
	mwclient := mediawiki.NewClient(wiki, http.DefaultClient, slog.Default(), "")
	mockClock := clock.Func(func() time.Time {
		return time.Date(2025, 12, 31, 4, 0, 0, 0, time.UTC)
	})
	provider := provider.New(slog.Default(), mockClock, mwclient, wiki)
	reportConfig := config.ReportData{
		PatrolExpiring: &config.PatrolExpiring{
			Windows: []config.Duration{
				{Hours: 12},
				{Days: 1},
				{Weeks: 1},
			},
		},
	}

	got := provider.Data(t.Context(), reportConfig)
	want := data.PatrolExpiring{
		Windows: []data.ExpiryWindow{
			{
				Count:    1,
				Duration: config.Duration{Hours: 12},
			},
			{
				Count:    2,
				Duration: config.Duration{Days: 1},
			},
			{
				Count:    3,
				Duration: config.Duration{Weeks: 1},
			},
		},
	}

	diff := cmp.Diff(want, got.PatrolExpiring)
	if diff != "" {
		t.Fatalf("result mismatch (-want +got):\n%s", diff)
	}
}

// Calling API:Login twice in a session results in an pi-login-fail-badsessionprovider error.
func TestProvider_Data_LoginOnce(t *testing.T) {
	t.Parallel()

	countReqTokens := 0
	countReqLogin := 0
	countReqRC := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("action") == "query" && q.Get("meta") == "tokens" && q.Get("type") == "login":
			countReqTokens++
			w.Write([]byte(`
				{
					"query": {
						"tokens": {
							"logintoken": "8af20f35764bee599652a5d5d9e804d469444fb1+\\"
						}
					}
				}
			`))
		case q.Get("action") == "login":
			countReqLogin++
			w.Write([]byte(`
				{
					"login": {
						"result": "Success",
						"lguserid": 45359339,
						"lgusername": "PhantomCaleb"
					}
				}
			`))
		case q.Get("action") == "query" && q.Get("list") == "recentchanges":
			countReqRC++
			w.Write([]byte(`
				{
					"query": {
						"recentchanges": []
					}
				}
			`))
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	wiki := config.Wiki{
		Site: config.WikiSite{
			URL:          srv.URL,
			ScriptPath:   "/w",
			RCMaxAgeDays: 30,
		},
	}
	mwclient := mediawiki.NewClient(wiki, http.DefaultClient, slog.Default(), "")
	provider := provider.New(slog.Default(), clock.Func(time.Now), mwclient, wiki)

	reportConfig := config.ReportData{
		PatrolExpiring: &config.PatrolExpiring{
			Windows: []config.Duration{
				{Days: 1},
			},
		},
	}
	_ = provider.Data(t.Context(), reportConfig)
	_ = provider.Data(t.Context(), reportConfig)

	if countReqTokens != 1 {
		t.Errorf("got %d API:Tokens requests, want %d", countReqTokens, 1)
	}

	if countReqLogin != 1 {
		t.Errorf("got %d API:Login requests, want %d", countReqLogin, 1)
	}

	if countReqRC != 2 {
		t.Errorf("got %d API:RecentChanges requests, want %d", countReqRC, 2)
	}
}
