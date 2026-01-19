// Package interactions implements the /interactions route handler
package interactions

import "net/http"

// NewHandler returns the [http.Handler] for the /interactions route.
func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
}
