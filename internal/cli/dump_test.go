package cli_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/catatsuy/bento/internal/cli"
)

func TestRunDump(t *testing.T) {
	repoPath := "testdata/repo"
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream, inputStream, nil, false)

	err := cli.RunDump(repoPath, "")
	if err != nil {
		t.Fatalf("RunDump failed: %v", err)
	}

	if !strings.Contains(outStream.String(), "This is a text file.") {
		t.Errorf("expected header to be included in output")
	}

	if !strings.Contains(outStream.String(), "--END--") {
		t.Errorf("output should end with --END--")
	}
}

func TestRunDumpWithIgnorePatterns(t *testing.T) {
	repoPath := "testdata/repo_with_ignore"
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream, inputStream, nil, false)

	err := cli.RunDump(repoPath, "")
	if err != nil {
		t.Fatalf("RunDump failed: %v", err)
	}

	if strings.Contains(outStream.String(), "----\nignored_file.txt") {
		t.Errorf("ignored files should not be included in the output")
	}
}

func TestRunDumpIgnoresBinaryFiles(t *testing.T) {
	repoPath := "testdata/repo_with_binaries"
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream, inputStream, nil, false)

	err := cli.RunDump(repoPath, "")
	if err != nil {
		t.Fatalf("RunDump failed: %v", err)
	}

	if strings.Contains(outStream.String(), "binary_file.png") {
		t.Errorf("binary files should not be included in the output")
	}
}

func TestRunDumpAIIgnore(t *testing.T) {
	repoPath := "testdata/repo_with_aiignore"
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream, inputStream, nil, false)

	err := cli.RunDump(repoPath, "")
	if err != nil {
		t.Fatalf("RunDump failed: %v", err)
	}

	if strings.Contains(outStream.String(), "----\nai_ignored_file.txt") {
		t.Errorf("files specified in .aiignore should not be included in the output")
	}
}

func TestRunDump_WithDescription(t *testing.T) {
	repoPath := "testdata/repo"
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream, inputStream, nil, false)

	description := "This is a test description."
	err := cli.RunDump(repoPath, description)
	if err != nil {
		t.Fatalf("RunDump failed: %v", err)
	}

	// Check that the output includes the description
	if !strings.Contains(outStream.String(), description) {
		t.Fatalf("Expected description %q to be in output, got: %q", description, outStream.String())
	}
}
