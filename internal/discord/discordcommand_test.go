package discord_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
)

func TestClient_RegisterGuildCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string // test case name
		statusCode   int    // HTTP status code to return in mock API response
		responseBody []byte // mock API response body
		wantErr      error  // expected error value, or nil if no error expected
	}{
		{
			name:         "200",
			statusCode:   200,
			responseBody: readFile(t, "testdata/registerguildcommand.json"),
			wantErr:      nil,
		},
		{
			name:         "400",
			statusCode:   400,
			responseBody: readFile(t, "testdata/registerguildcommand_invalid.json"),
			wantErr: &discord.HTTPStatusError{
				StatusCode: 400,
				Body:       readFile(t, "testdata/registerguildcommand_invalid.json"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)

				if test.responseBody != nil {
					w.Write(test.responseBody)
				}
			}))
			defer srv.Close()

			cfg := config.Discord{API: srv.URL}
			client := discord.NewClient(http.DefaultClient, slog.Default(), cfg, "")

			err := client.RegisterGuildCommand(t.Context(), "guildId", discord.Command{})

			if test.wantErr != nil {
				wantType := reflect.TypeOf(test.wantErr)
				if wantType != reflect.TypeOf(err) {
					t.Fatalf("Got error of type %T, want %T", err, test.wantErr)
				}

				diff := cmp.Diff(test.wantErr, err)
				if diff != "" {
					t.Fatalf("Error mismatch (-want +got):\n%s", diff)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}
