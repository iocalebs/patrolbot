package errdefs_test

import (
	"testing"

	"github.com/iocalebs/patrolbot/internal/errdefs"
)

func TestConfigError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		causes []string
		want   string
	}{
		{
			name:   "one cause",
			causes: []string{`.wikis["zwen"].site.url not set`},
			want:   `invalid configuration: .wikis["zwen"].site.url not set`,
		},
		{
			name:   "multiple causes",
			causes: []string{`.wikis["zwen"].auth.username not set`, `.wikis["zwen"].auth.password not set`},
			want: `invalid configuration:
- .wikis["zwen"].auth.username not set
- .wikis["zwen"].auth.password not set`,
		},
		{
			name:   "no causes",
			causes: []string{},
			want:   "invalid configuration",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := errdefs.NewConfigError(test.causes...)
			got := err.Error()

			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestUsageError(t *testing.T) {
	t.Parallel()

	msg := "missing required flag"

	err := errdefs.UsageError(msg)
	if err.Error() != msg {
		t.Errorf("got %q, want %q", err.Error(), msg)
	}
}
