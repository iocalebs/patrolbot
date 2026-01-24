// Package errdefs provides error types for common PatrolBot errors that warrant special handling in the CLI or bot
// server.
package errdefs

import "strings"

// UsageError indicates an error resulting from incorrect command usage by the user.
type UsageError string

// ConfigError indicates a failure to run a command due to invalid configuration.
type ConfigError struct {
	causes []string
}

// NewConfigError creates a new [ConfigError].
func NewConfigError(causes ...string) *ConfigError {
	return &ConfigError{
		causes: causes,
	}
}

func (e UsageError) Error() string {
	return string(e)
}

func (e ConfigError) Error() string {
	var builder strings.Builder

	builder.WriteString("invalid configuration")

	if len(e.causes) == 0 {
		return builder.String()
	}

	if len(e.causes) == 1 {
		builder.WriteString(": " + e.causes[0])
		return builder.String()
	}

	builder.WriteString(":")

	for _, cause := range e.causes {
		builder.WriteString("\n- " + cause)
	}

	return builder.String()
}
