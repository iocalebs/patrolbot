package gcp_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iocalebs/patrolbot/internal/gcp"
)

func TestLoggerTrace(t *testing.T) {
	var logBuf bytes.Buffer

	logger := gcp.NewLogger(&logBuf, slog.LevelInfo)
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
		t.Errorf("Expected written logs to contain %q, got:\n%s", want, got)
	}
}

func TestLoggerLevel(t *testing.T) {
	t.Parallel()

	errMsg := "error msg"
	warnMsg := "warning msg"
	infoMsg := "info msg"
	debugMsg := "debug msg"

	tests := []struct {
		level       slog.Level
		wantMsgs    []string
		notWantMsgs []string
	}{
		{
			level:       slog.LevelDebug,
			wantMsgs:    []string{errMsg, warnMsg, infoMsg, debugMsg},
			notWantMsgs: []string{},
		},
		{
			level:       slog.LevelInfo,
			wantMsgs:    []string{errMsg, warnMsg, infoMsg},
			notWantMsgs: []string{debugMsg},
		},
		{
			level:       slog.LevelWarn,
			wantMsgs:    []string{errMsg, warnMsg},
			notWantMsgs: []string{debugMsg, infoMsg},
		},
		{
			level:       slog.LevelError,
			wantMsgs:    []string{errMsg},
			notWantMsgs: []string{debugMsg, infoMsg, warnMsg},
		},
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			t.Parallel()

			var logBuf bytes.Buffer

			logger := gcp.NewLogger(&logBuf, test.level)

			logger.Debug(debugMsg)
			logger.Info(infoMsg)
			logger.Warn(warnMsg)
			logger.Error(errMsg)

			for _, want := range test.wantMsgs {
				if !strings.Contains(logBuf.String(), want) {
					t.Errorf("Expected written logs to contain %q, got:\n%s", want, logBuf.String())
				}
			}

			for _, notWant := range test.notWantMsgs {
				if strings.Contains(logBuf.String(), notWant) {
					t.Errorf("Expected written logs to NOT contain %q, got:\n%s", notWant, logBuf.String())
				}
			}
		})
	}
}
