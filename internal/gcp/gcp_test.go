package gcp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iocalebs/patrolbot/internal/gcp"
)

func TestLoggerTrace(t *testing.T) {
	var logBuf bytes.Buffer

	logger := gcp.NewLogger(&logBuf)
	handler := gcp.TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "message", "key", "value")
		w.WriteHeader(http.StatusOK)
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	request.Header.Set("Traceparent", "00-4c6a2f4b3e1a9b8c7d6e5f4a3b2c1d0e-9f8e7d6c5b4a3210-01")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer res.Body.Close() //nolint:errcheck

	got := logBuf.String()

	want := `"logging.googleapis.com/trace":"projects/test-project/traces/4c6a2f4b3e1a9b8c7d6e5f4a3b2c1d0e"`
	if !strings.Contains(got, want) {
		t.Errorf("Expected log to contain %q, got %q", want, got)
	}
}
