// Package clock provides an interface around the standard library's [time] package so that unit tests can inject a mock
// clock.
package clock

import (
	"log/slog"
	"os"
	"time"
)

// Clock represents a source of the current time.
type Clock interface {
	Now() time.Time
}

// Func is an adapter that allows a function returning [time.Time] (such as [time.Now]) to satisfy the Clock interface.
type Func func() time.Time

// Now returns the time produced by the underlying function.
func (f Func) Now() time.Time {
	return f()
}

// Mock returns a [Clock] that reads the MOCK_TIME env variable if set to an RFC3339 timestamp,
// otherwise it uses [time.Now].
func Mock(logger *slog.Logger) Clock { //nolint:ireturn
	return Func(func() time.Time {
		mockTime := os.Getenv("MOCK_TIME")
		if mockTime != "" {
			now, err := time.Parse(time.RFC3339, mockTime)
			if err == nil {
				logger.Debug("Using mock time", "MOCK_TIME", mockTime)
				return now
			}

			logger.Error("Error parsing $MOCK_TIME", "MOCK_TIME", mockTime, "err", err)
		}

		logger.Debug("MOCK_TIME not set, using current time")

		return time.Now()
	})
}
