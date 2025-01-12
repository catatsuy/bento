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

	err := cli.RunDump(repoPath)
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

	err := cli.RunDump(repoPath)
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

	err := cli.RunDump(repoPath)
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

	err := cli.RunDump(repoPath)
	if err != nil {
		t.Fatalf("RunDump failed: %v", err)
	}

	if strings.Contains(outStream.String(), "----\nai_ignored_file.txt") {
		t.Errorf("files specified in .aiignore should not be included in the output")
	}
}

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		patterns []string
		expected bool
	}{
		{
			name:     "Ignore file with exact match",
			filePath: "test.txt",
			patterns: []string{"test.txt"},
			expected: true,
		},
		{
			name:     "Ignore file in directory",
			filePath: "src/test.txt",
			patterns: []string{"src/"},
			expected: true,
		},
		{
			name:     "Do not ignore file outside directory",
			filePath: "main.go",
			patterns: []string{"src/"},
			expected: false,
		},
		{
			name:     "Match wildcard pattern",
			filePath: "main_test.go",
			patterns: []string{"*.go"},
			expected: true,
		},
		{
			name:     "Do not match mismatched wildcard pattern",
			filePath: "main_test.go",
			patterns: []string{"*.txt"},
			expected: false,
		},
		{
			name:     "Ignore file in nested directory",
			filePath: "src/sub/test.txt",
			patterns: []string{"src/"},
			expected: true,
		},
		{
			name:     "Handle invalid pattern",
			filePath: "main.go",
			patterns: []string{"[invalid pattern"},
			expected: false,
		},
		{
			name:     "Ignore .git directory",
			filePath: ".git/info/exclude",
			patterns: []string{".git/"},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ShouldIgnore(test.filePath, test.patterns)
			if result != test.expected {
				t.Errorf("shouldIgnore(%q, %q) = %v; want %v", test.filePath, test.patterns, result, test.expected)
			}
		})
	}
}
