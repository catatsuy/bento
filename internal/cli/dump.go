package cli

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// RunDump processes the repository path and writes its contents to standard output.
func (c *CLI) RunDump(repoPath, description string) error {
	ignorePatterns := make([]string, 0, 10)
	ignorePatterns = append(ignorePatterns, ".git/") // Default patterns to ignore the .git directory

	// Check for .aiignore file
	aiIgnorePath := filepath.Join(repoPath, ".aiignore")
	patterns, err := readIgnoreFile(aiIgnorePath, "")
	if err != nil {
		return fmt.Errorf("failed to read .aiignore file: %w", err)
	}
	ignorePatterns = append(ignorePatterns, patterns...)

	// Check for .gitignore file
	gitIgnorePath := filepath.Join(repoPath, ".gitignore")
	patterns, err = readIgnoreFile(gitIgnorePath, "")
	if err != nil {
		return fmt.Errorf("failed to read .gitignore file: %w", err)
	}
	ignorePatterns = append(ignorePatterns, patterns...)

	err = filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		for _, pattern := range ignorePatterns {
			if strings.HasPrefix(relPath, pattern) {
				return filepath.SkipDir
			}
		}

		if d.IsDir() {
			gitIgnorePath := filepath.Join(path, ".gitignore")
			if patterns, err := readIgnoreFile(gitIgnorePath, relPath); err == nil {
				ignorePatterns = append(ignorePatterns, patterns...)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk the path %s: %w", repoPath, err)
	}

	ignores := gitignore.CompileIgnoreLines(ignorePatterns...)

	dumpPrompt := `The output represents a Git repository's content in the following format:

1. Each section begins with ----.
2. The first line after ---- contains the file path and name.
3. The subsequent lines contain the file contents.
4. The repository content ends with --END--.
`

	if description != "" {
		dumpPrompt += "\n" + unescapeString(description) + "\n"
	}

	dumpPrompt += "\nAny text after --END-- should be treated as instructions, using the repository content as context.\n"

	// Write the initial explanation text
	if _, err := fmt.Fprintln(c.outStream, dumpPrompt); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Walk through the repository and process files
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
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
		if ignores.MatchesPath(relPath) {
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

func unescapeString(input string) string {
	replacer := strings.NewReplacer(
		`\\`, `\`,
		`\n`, "\n",
		`\t`, "\t",
		`\r`, "\r",
	)
	return replacer.Replace(input)
}

// readIgnoreFile reads ignore patterns from the specified file and optionally prepends the provided directory to the patterns.
func readIgnoreFile(filePath string, dir string) ([]string, error) {
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
			patterns = append(patterns, filepath.Join(dir, line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
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
