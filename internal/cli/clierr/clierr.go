// Package clierr provides custom error types, allowing the CLI error handling logic to be tailored to the error type.
package clierr

// UsageError represents an error resulting from incorrect command usage by the user.
type UsageError string

func (e UsageError) Error() string {
	return string(e)
}
