package main_test

// TODO: add tests for config view covering precedence rules

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/tools/txtar"
)

const integrationTestDir = "testdata/integration-tests"

//go:embed testdata/integration-tests/*
var integrationTests embed.FS

var update bool //nolint:gochecknoglobals

type integrationTest struct {
	comment []byte
	command []byte
	stdout  []byte
	stderr  []byte
}

func TestMain(m *testing.M) {
	flag.BoolVar(&update, "update", false, "Update golden files")
	m.Run()
}

// TestCommands uses the a variation of the file-based approach described in https://research.swtch.com/testing,
// Each file is a test, with the input command and expected outputs (stdout, stderr) stored in one file using txtar.
//
// This allows inputs to be neatly displayed alongside expected outputs. It also allows new tests to be added without
// changing any Go test code. If the output changes unexpectedly, the test failure will display a unified diff of
// expected vs. actual output.
//
// In keeping with the "golden files" approach, running go test with the -update flag updates the stdout and stderr
// in the txtar files with current output. git diff then shows a helpful diff of what has changed in terms of command
// output.
func TestCommands(t *testing.T) {
	t.Parallel()

	entries, err := integrationTests.ReadDir(integrationTestDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(integrationTestDir, entry.Name())

			test := parseTestFile(t, file)

			stdout := bytes.Buffer{}
			stderr := bytes.Buffer{}

			// #nosec G204 -- trusted inputs from test suite
			cmd := exec.CommandContext(t.Context(), "sh", "-c", string(test.command))
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Run()

			var exitErr *exec.ExitError
			if err != nil && (!errors.As(err, &exitErr) || cmd.ProcessState.ExitCode() > 1) {
				t.Fatal("Error executing test command: %w", err)
			}

			test.stdout = stdout.Bytes()
			test.stderr = stderr.Bytes()
			got := formatTestFile(t, test)

			if update {
				err := os.WriteFile(file, got, 0600)
				if err != nil {
					t.Fatalf("Error updating test file %s", file)
				}

				return
			}

			// Simpler to compare whole test files than to extract stdout and stderr from test files and run diffs on
			// each. This way the line numbers in the diff align with the checked in test file.
			diff := diffTestFile(t, file, got)
			if diff != "" {
				t.Errorf("Output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func parseTestFile(t *testing.T, path string) integrationTest {
	t.Helper()

	archive, err := txtar.ParseFile(path)
	if err != nil {
		t.Fatalf("Error parsing integration test file: %v", err)
	}

	test := integrationTest{}
	test.comment = archive.Comment

	for _, file := range archive.Files {
		switch file.Name {
		case "command":
			test.command = file.Data
		case "stdout":
			test.stdout = file.Data
		case "stderr":
			test.stderr = file.Data
		default:
			t.Errorf("unrecognized file %s in txtar %s", file.Name, path)
		}
	}

	if len(test.command) == 0 {
		t.Fatalf("Error parsing integration test file: missing command in test file %s", path)
	}

	return test
}

func formatTestFile(t *testing.T, test integrationTest) []byte {
	t.Helper()

	archive := &txtar.Archive{
		Comment: test.comment,
		Files: []txtar.File{
			{Name: "command", Data: test.command},
			{Name: "stdout", Data: test.stdout},
			{Name: "stderr", Data: test.stderr},
		},
	}

	return txtar.Format(archive)
}

func diffTestFile(t *testing.T, wantPath string, got []byte) string {
	t.Helper()

	dir := t.TempDir()
	gotPath := filepath.Join(dir, "got.txt")

	err := os.WriteFile(gotPath, got, 0600)
	if err != nil {
		t.Fatalf("Error writing %s: %v", gotPath, err)
	}

	// #nosec G204 -- trusted inputs from test suite
	cmd := exec.CommandContext(t.Context(), "diff", "-u", wantPath, gotPath)
	diff, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	if err != nil && (!errors.As(err, &exitErr) || cmd.ProcessState.ExitCode() > 1) {
		t.Fatalf("Error diffing text: %v, out: %s", err, diff)
	}

	if len(diff) == 0 {
		return ""
	}

	// Strip first two header lines
	diffParts := bytes.SplitN(diff, []byte{'\n'}, 3)
	if len(diffParts) < 3 {
		t.Fatal("Error procesing diff output")
	}

	return string(diffParts[2])
}
