package clock_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/iocalebs/patrolbot/internal/clock"
)

func TestMock(t *testing.T) {
	mockTime := time.Unix(0, 0).UTC()
	t.Setenv("MOCK_TIME", mockTime.Format(time.RFC3339))

	clock := clock.Mock(slog.Default())
	got := clock.Now()

	if got != mockTime {
		t.Errorf("got %s, want %s", got.Format(time.RFC3339), mockTime.Format(time.RFC3339))
	}
}
