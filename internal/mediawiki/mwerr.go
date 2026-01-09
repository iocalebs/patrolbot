package mediawiki

import (
	"fmt"
	"net/http"
	"strings"
)

// HTTPStatusError indicates the MediaWiki API responded with a non-2xx status code.
type HTTPStatusError struct {
	StatusCode int    // HTTP status code
	Body       []byte // Response body, if any
}

// APIError indicates the MediaWiki API response body contained errors and/or warnings.
// See: https://www.mediawiki.org/wiki/API:Errors_and_warnings
type APIError struct {
	Errors   []string
	Warnings []string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

func (e *APIError) Error() string {
	var builder strings.Builder

	// Deliberately avoiding adding any text other that what's in the MediaWiki response for better i18n
	if e.Errors != nil {
		for _, err := range e.Errors {
			builder.WriteString("\n- " + err)
		}
	}

	if e.Warnings != nil {
		for _, warn := range e.Warnings {
			builder.WriteString("\n- " + warn)
		}
	}

	return builder.String()
}
