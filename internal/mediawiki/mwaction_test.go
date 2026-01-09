package mediawiki_test

import (
	"context"
	"errors"
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
				w.Write([]byte("{}"))
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

			var httpStatusCodeErr *mediawiki.HTTPStatusError
			if !errors.As(err, &httpStatusCodeErr) {
				t.Fatalf("Expected HTTPStatusError, got: %v", err)
			}

			wantMsg := "502 Bad Gateway"
			if httpStatusCodeErr.Error() != wantMsg {
				t.Errorf("Unexpected error message, got: %q, want %q", httpStatusCodeErr.Error(), wantMsg)
			}

			if httpStatusCodeErr.StatusCode != http.StatusBadGateway {
				t.Errorf(
					"HTTPStatusError.StatusCode = %d, want %d",
					httpStatusCodeErr.StatusCode,
					http.StatusBadGateway,
				)
			}

			if httpStatusCodeErr.Body == nil {
				t.Fatal("Expected empty HTTPStatusError.Body, got nil")
			}

			if len(httpStatusCodeErr.Body) != 0 {
				t.Fatalf("Expected empty HTTPStatusError.Body, got %q", string(httpStatusCodeErr.Body))
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
				w.Write([]byte("an error occurred"))
			}))
			defer srv.Close()

			cfg := config.Wiki{}
			cfg.Site.URL = srv.URL
			mwclient := mediawiki.NewClient(cfg, http.DefaultClient, slog.Default(), "")

			err := test.requestFunc(t.Context(), mwclient)

			var httpStatusCodeErr *mediawiki.HTTPStatusError
			if !errors.As(err, &httpStatusCodeErr) {
				t.Fatalf("Expected HTTPStatusError, got: %v", err)
			}

			wantMsg := "500 Internal Server Error"
			if httpStatusCodeErr.Error() != wantMsg {
				t.Errorf("Error() = %q, want %q", httpStatusCodeErr.Error(), wantMsg)
			}

			if httpStatusCodeErr.StatusCode != http.StatusInternalServerError {
				t.Errorf(
					"HTTPStatusError.StatusCode = %d, want %d",
					httpStatusCodeErr.StatusCode,
					http.StatusInternalServerError,
				)
			}

			wantBody := "an error occurred"
			if string(httpStatusCodeErr.Body) != wantBody {
				t.Fatalf("HTTPStatusError.Body = %q, want %q", string(httpStatusCodeErr.Body), wantBody)
			}
		})
	}
}
