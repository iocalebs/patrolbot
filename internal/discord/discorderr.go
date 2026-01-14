package discord

import (
	"fmt"
	"net/http"
)

// HTTPStatusError indicates the Discord API responded with a non-2xx status code.
type HTTPStatusError struct {
	StatusCode int    // HTTP status code
	Body       []byte // Response body, if any
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode))
}
