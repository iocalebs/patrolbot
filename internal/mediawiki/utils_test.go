package mediawiki_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type captureTransport struct {
	base        http.RoundTripper
	lastRequest *http.Request
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lastRequest = req.Clone(req.Context())
	return t.base.RoundTrip(req) //nolint:wrapcheck
}

func newHTTPClient() (*http.Client, *captureTransport) {
	transport := &captureTransport{
		base: http.DefaultTransport,
	}
	client := &http.Client{
		Transport: transport,
	}

	return client, transport
}

func readFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	return data
}
