// Package gcp provides utilties for running PatrolBot on Google Cloud Platform.
package gcp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type logHandler struct {
	*slog.JSONHandler
}

// NewLogger creates a new [slog.Logger] configured for logging on Google Cloud Platform.
func NewLogger(w io.Writer) *slog.Logger {
	handler := &logHandler{
		JSONHandler: slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource: true,
			Level:     nil, // TODO: make configurable
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				switch attr.Key {
				case slog.MessageKey:
					attr.Key = "message"
				case slog.LevelKey:
					attr.Key = "severity"
				case slog.SourceKey:
					attr.Key = "logging.googleapis.com/sourceLocation"
				}

				return attr
			},
		}),
	}

	return slog.New(handler)
}

// TraceMiddleware adds Google Cloud trace information from HTTP request headers to the request context so that
// the request can be correlated with log entries.
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

		if projectID != "" {
			traceHeader := r.Header.Get("Traceparent")
			traceParts := strings.Split(traceHeader, "-")

			if len(traceParts) > 1 && len(traceParts[1]) > 0 {
				trace := fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[1])
				ctx := contextWithTrace(r.Context(), trace)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (h *logHandler) Handle(ctx context.Context, rec slog.Record) error {
	if trace, ok := traceFromContext(ctx); ok {
		rec = rec.Clone()
		rec.AddAttrs(slog.String("logging.googleapis.com/trace", trace))
	}

	return h.JSONHandler.Handle(ctx, rec) //nolint:wrapcheck
}
