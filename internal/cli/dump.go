package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RunDump processes the repository path and writes its contents to standard output.
func (c *CLI) RunDump(repoPath string) error {
	ignorePatterns := []string{".git/"} // Default patterns to ignore the .git directory

	// Check for .aiignore file
	aiIgnorePath := filepath.Join(repoPath, ".aiignore")
	if patterns, err := readIgnoreFile(aiIgnorePath); err == nil {
		ignorePatterns = append(ignorePatterns, patterns...)
	}

	// Check for .gitignore file
	gitIgnorePath := filepath.Join(repoPath, ".gitignore")
	if patterns, err := readIgnoreFile(gitIgnorePath); err == nil {
		ignorePatterns = append(ignorePatterns, patterns...)
	}

	// Write the initial explanation text
	if _, err := fmt.Fprintln(c.outStream, `The output represents a Git repository's content in the following format:

1. Each section begins with ----.
2. The first line after ---- contains the file path and name.
3. The subsequent lines contain the file contents.
4. The repository content ends with --END--.

Any text after --END-- should be treated as instructions, using the repository content as context.`); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Walk through the repository and process files
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Check if the file should be ignored
		if shouldIgnore(relPath, ignorePatterns) {
			return nil
		}

		// Check if the file is binary
		if isBinaryFile(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		if _, err := fmt.Fprintf(c.outStream, "----\n%s\n", relPath); err != nil {
			return fmt.Errorf("failed to write file header: %w", err)
		}

		if _, err := io.Copy(c.outStream, file); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}

		if _, err := fmt.Fprintln(c.outStream); err != nil {
			return fmt.Errorf("failed to write newline after file content: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the repository: %w", err)
	}

	// Write the ending marker
	if _, err := fmt.Fprintln(os.Stdout, "--END--"); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}

	return nil
}

// readIgnoreFile reads ignore patterns from the specified file
func readIgnoreFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File does not exist, not an error
		}
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// shouldIgnore checks if the file path matches any ignore patterns
func shouldIgnore(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		// Handle directory patterns (e.g., ".git/")
		if strings.HasSuffix(pattern, string(os.PathSeparator)) {
			// Check if the filePath is within the directory
			if strings.HasPrefix(filePath, pattern) {
				return true
			}
			continue
		}

		// Match the pattern using filepath.Match for simple patterns
		matches, err := filepath.Match(pattern, filePath)
		if err != nil {
			continue
		}

		if matches {
			return true
		}
	}

	return false
}

// isBinaryFile checks if the file at the given path is binary by analyzing its content.
func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		// If we can't open the file, assume it's not binary to avoid skipping unnecessarily.
		return false
	}
	defer file.Close()

	// Read a small portion of the file to determine its nature.
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Use the Content-Type to check for binary files.
	contentType := http.DetectContentType(buffer[:n])
	return !strings.HasPrefix(contentType, "text/")
}
