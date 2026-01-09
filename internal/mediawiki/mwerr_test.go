package mediawiki_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func TestHTTPStatusError(t *testing.T) {
	t.Parallel()

	err := &mediawiki.HTTPStatusError{
		StatusCode: 500,
		Body:       []byte("An error occurred"),
	}

	want := "500 Internal Server Error"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *mediawiki.APIError
		want string
	}{
		{
			name: "errors only",
			err: &mediawiki.APIError{
				Errors:   []string{"Error 1"},
				Warnings: []string{},
			},
			want: "\n- Error 1",
		},
		{
			name: "warnings only",
			err: &mediawiki.APIError{
				Errors:   []string{},
				Warnings: []string{"Warning 1"},
			},
			want: "\n- Warning 1",
		},
		{
			name: "errors and warnings",
			err: &mediawiki.APIError{
				Errors:   []string{"Error 1", "Error 2"},
				Warnings: []string{"Warning 1", "Warning 2"},
			},
			want: "\n- Error 1\n- Error 2\n- Warning 1\n- Warning 2",
		},
		{
			name: "nil slices",
			err:  &mediawiki.APIError{},
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			diff := cmp.Diff(test.want, test.err.Error())
			if diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
