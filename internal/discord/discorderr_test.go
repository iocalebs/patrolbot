package discord_test

import (
	"testing"

	"github.com/iocalebs/patrolbot/internal/discord"
)

func TestHTTPStatusError(t *testing.T) {
	t.Parallel()

	err := &discord.HTTPStatusError{
		StatusCode: 500,
		Body:       []byte("An error occurred"),
	}

	want := "500 Internal Server Error"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
