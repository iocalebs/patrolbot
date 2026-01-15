package discord_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/discord"
)

func TestClient_CreateMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string          // test case name
		statusCode   int             // HTTP status code to return in mock API response
		responseBody []byte          // mock API response body
		wantMessage  discord.Message // expected Message return value, or zero value if error expected
		wantErr      error           // expected error value, or nil if no error expected
	}{
		{
			name:         "200",
			statusCode:   200,
			responseBody: readFile(t, "testdata/createmessage.json"),
			wantMessage: discord.Message{
				ID:        "1461062205767155837",
				ChannelID: "911393794262458421",
			},
			wantErr: nil,
		},
		{
			name:         "400 with body",
			statusCode:   400,
			responseBody: readFile(t, "testdata/createmessage_invalid.json"),
			wantMessage:  discord.Message{},
			wantErr: &discord.HTTPStatusError{
				StatusCode: 400,
				Body:       readFile(t, "testdata/createmessage_invalid.json"),
			},
		},
		{
			name:         "500 no body",
			statusCode:   500,
			responseBody: []byte(""),
			wantMessage:  discord.Message{},
			wantErr: &discord.HTTPStatusError{
				StatusCode: 500,
				Body:       []byte(""),
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

			message, err := client.CreateMessage(t.Context(), "channelID", discord.CreateMessageRequestBody{})

			diff := cmp.Diff(test.wantMessage, message)
			if diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}

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

func readFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	return data
}
