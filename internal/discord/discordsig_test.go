package discord_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/iocalebs/patrolbot/internal/discord"
)

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte("foo")

	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate public/private key pair: %v", err)
	}

	tests := []struct {
		name       string
		addHeaders func(*http.Request)
		want       bool
	}{
		{
			name: "valid",
			want: true,
			addHeaders: func(r *http.Request) {
				bytes := append([]byte(timestamp), body...)
				sig := ed25519.Sign(private, bytes)
				r.Header.Add("X-Signature-Ed25519", hex.EncodeToString(sig))
				r.Header.Add("X-Signature-Timestamp", timestamp)
			},
		},
		{
			name: "no ed25519 header",
			want: false,
			addHeaders: func(r *http.Request) {
				bytes := append([]byte(timestamp), body...)
				sig := ed25519.Sign(private, bytes)
				r.Header.Add("X-Signature-Ed25519", hex.EncodeToString(sig))
			},
		},
		{
			name: "no timestamp header",
			want: false,
			addHeaders: func(r *http.Request) {
				r.Header.Add("X-Signature-Timestamp", timestamp)
			},
		},
		{
			name: "invalid ed25519 header",
			want: false,
			addHeaders: func(r *http.Request) {
				_, private, err := ed25519.GenerateKey(nil)
				if err != nil {
					t.Fatalf("Failed to generate public/private key pair: %v", err)
				}

				bytes := append([]byte(timestamp), body...)
				sig := ed25519.Sign(private, bytes)

				r.Header.Add("X-Signature-Ed25519", hex.EncodeToString(sig))
				r.Header.Add("X-Signature-Timestamp", timestamp)
			},
		},
		{
			name: "invalid timestamp header",
			want: false,
			addHeaders: func(r *http.Request) {
				bytes := append([]byte(timestamp), body...)
				sig := ed25519.Sign(private, bytes)
				r.Header.Add("X-Signature-Ed25519", hex.EncodeToString(sig))
				r.Header.Add("X-Signature-Timestamp", strconv.FormatInt(time.Unix(0, 0).Unix(), 10))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/interactions", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			test.addHeaders(req)

			got := discord.VerifySignature(req, public)
			if got != test.want {
				t.Fatalf("got %t, want %t", got, test.want)
			}
		})
	}
}
