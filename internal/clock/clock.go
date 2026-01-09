// Package clock provides an interface around the standard library's [time] package so that unit tests can inject a mock
// clock.
package clock

import "time"

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
