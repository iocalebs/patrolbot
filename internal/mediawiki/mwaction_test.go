package mediawiki_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

type actionTest struct {
	name        string
	requestFunc func(ctx context.Context, client *mediawiki.Client) error
}

//nolint:wrapcheck
func tests() []actionTest {
	return []actionTest{
		{
			name: "LogEvents",
			requestFunc: func(ctx context.Context, client *mediawiki.Client) error {
				paginator := mediawiki.NewLogEventsPaginator(client, mediawiki.LogEventsQueryParams{})
				_, err := paginator.NextPage(ctx)

				return err
			},
		},
		{
			name: "Login",
			requestFunc: func(ctx context.Context, client *mediawiki.Client) error {
				return client.Login(ctx, "")
			},
		},
		{
			name: "LoginToken",
			requestFunc: func(ctx context.Context, client *mediawiki.Client) error {
				_, err := client.LoginToken(ctx)
				return err
			},
		},
		{
			name: "RecentChanges",
			requestFunc: func(ctx context.Context, client *mediawiki.Client) error {
				paginator := mediawiki.NewRecentChangesPaginator(client, mediawiki.RecentChangesQueryParams{})
				_, err := paginator.NextPage(ctx)

				return err
			},
		},
	}
}

func TestRequestHeaders(t *testing.T) {
	t.Parallel()

	for _, test := range tests() {
		t.Run(test.name, func(t *testing.T) {
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
			mwclient := mediawiki.NewClient(cfg, httpClient, slog.Default(), userAgent)

			test.requestFunc(t.Context(), mwclient) //nolint:errcheck,gosec

			gotUserAgent := transport.lastRequest.Header.Get("User-Agent")
			if gotUserAgent != userAgent {
				t.Errorf("Expected User-Agent header %q, got %q", userAgent, gotUserAgent)
			}

			gotFrom := transport.lastRequest.Header.Get("From")
			if gotFrom != from {
				t.Errorf("Expected From header %q, got %q", from, gotFrom)
			}
		})
	}
}

func TestErrorStatusCodesNoBody(t *testing.T) {
	t.Parallel()

	for _, test := range tests() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL
			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")

			err := test.requestFunc(t.Context(), mwclient)
			want := "MediaWiki API returned error status code: 502 with empty response body"

			if err == nil {
				t.Fatalf("Expected error %q but got nil", want)
			}

			if err.Error() != want {
				t.Fatalf("Expected error %q but got %q", want, err.Error())
			}
		})
	}
}

func TestErrorStatusCodesWithBody(t *testing.T) {
	t.Parallel()

	for _, test := range tests() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("an error occurred")) //nolint:errcheck,gosec
			}))
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL
			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")

			err := test.requestFunc(t.Context(), mwclient)
			want := "MediaWiki API returned error status code: 500, response body: an error occurred"

			if err == nil {
				t.Fatalf("Expected error %q but got nil", want)
			}

			if err.Error() != want {
				t.Fatalf("Expected error %q but got %q", want, err.Error())
			}
		})
	}
}
