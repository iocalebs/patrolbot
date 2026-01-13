// Package clierr provides custom error types, allowing the CLI error handling logic to be tailored to the error type.
package clierr

import "strings"

// UsageError represents an error resulting from incorrect command usage by the user.
type UsageError string

// MultiError indicates a combination of several related errors.
type MultiError struct {
	msg    string
	causes []string
}

// NewMultiError creates a new [MutliError].
func NewMultiError(msg string, causes []string) *MultiError {
	return &MultiError{
		msg:    msg,
		causes: causes,
	}
}

func (e UsageError) Error() string {
	return string(e)
}

func (e MultiError) Error() string {
	var builder strings.Builder

	if e.msg != "" {
		builder.WriteString(e.msg + ":")
	}

	if e.causes != nil {
		for _, cause := range e.causes {
			builder.WriteString("\n- " + cause)
		}
	}

	return builder.String()
}
