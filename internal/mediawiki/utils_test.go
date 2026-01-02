package mediawiki_test

import (
	"os"
	"path/filepath"
	"testing"
)

func readFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	return data
}
