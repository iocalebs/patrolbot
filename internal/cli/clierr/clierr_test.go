package clierr_test

import (
	"testing"

	"github.com/iocalebs/patrolbot/internal/cli/clierr"
)

func TestMultiError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		msg    string
		causes []string
		want   string
	}{
		{
			name:   "Full error with message and multiple causes",
			msg:    "validation failed",
			causes: []string{"invalid email", "password too short"},
			want:   "validation failed:\n- invalid email\n- password too short",
		},
		{
			name:   "No message, only causes",
			msg:    "",
			causes: []string{"unauthorized", "timeout"},
			want:   "\n- unauthorized\n- timeout",
		},
		{
			name:   "Message with no causes",
			msg:    "system failure",
			causes: nil,
			want:   "system failure:",
		},
		{
			name:   "Empty causes slice",
			msg:    "no issues found",
			causes: []string{},
			want:   "no issues found:",
		},
		{
			name:   "Completely empty error",
			msg:    "",
			causes: nil,
			want:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := clierr.NewMultiError(test.msg, test.causes)
			got := err.Error()

			if got != test.want {
				t.Errorf("MultiError.Error() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestUsageError_Error(t *testing.T) {
	t.Parallel()

	msg := "missing required flag"

	err := clierr.UsageError(msg)
	if err.Error() != msg {
		t.Errorf("UsageError.Error() = %q, want %q", err.Error(), msg)
	}
}
