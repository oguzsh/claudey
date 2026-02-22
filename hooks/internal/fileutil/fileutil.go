// Package fileutil provides cross-platform file operations for hooks.
package fileutil

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ReadFile reads a text file, returning its content or empty string on error.
func ReadFile(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// WriteFile writes content to a file, creating parent directories as needed.
func WriteFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// AppendFile appends content to a file, creating parent directories as needed.
func AppendFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// FileResult represents a found file with its modification time.
type FileResult struct {
	Path  string
	Mtime time.Time
}

// FindFiles searches a directory for files matching a glob-like pattern.
// Options: maxAgeDays (0 = no limit), recursive.
// Results are sorted newest first.
func FindFiles(dir string, pattern string, maxAgeDays float64, recursive bool) []FileResult {
	if dir == "" || pattern == "" {
		return nil
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil
	}

	// Convert glob pattern to regex
	regexPattern := globToRegex(pattern)
	re, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		return nil
	}

	var results []FileResult
	now := time.Now()

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if path == dir {
			return nil
		}
		if d.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		if !re.MatchString(d.Name()) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if maxAgeDays > 0 {
			ageInDays := now.Sub(info.ModTime()).Hours() / 24
			if ageInDays > maxAgeDays {
				return nil
			}
		}

		results = append(results, FileResult{
			Path:  path,
			Mtime: info.ModTime(),
		})
		return nil
	}

	if recursive {
		filepath.WalkDir(dir, walkFn)
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		for _, entry := range entries {
			fullPath := filepath.Join(dir, entry.Name())
			walkFn(fullPath, entry, nil)
		}
	}

	// Sort newest first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Mtime.After(results[j].Mtime)
	})

	return results
}

// GrepFile searches for a pattern in a file and returns matching lines.
type GrepResult struct {
	LineNumber int
	Content    string
}

// GrepFile returns lines matching the regex pattern.
func GrepFile(filePath string, pattern string) []GrepResult {
	content, ok := ReadFile(filePath)
	if !ok {
		return nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	lines := strings.Split(content, "\n")
	var results []GrepResult

	for i, line := range lines {
		if re.MatchString(line) {
			results = append(results, GrepResult{
				LineNumber: i + 1,
				Content:    line,
			})
		}
	}

	return results
}

// CountInFile counts occurrences of a regex pattern in a file.
func CountInFile(filePath string, pattern string) int {
	content, ok := ReadFile(filePath)
	if !ok {
		return 0
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0
	}

	return len(re.FindAllString(content, -1))
}

// Exists returns true if the path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// globToRegex converts a simple glob pattern to a regex pattern.
func globToRegex(pattern string) string {
	var b strings.Builder
	for _, ch := range pattern {
		switch ch {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteByte('.')
		case '.', '+', '^', '$', '{', '}', '(', ')', '|', '[', ']', '\\':
			b.WriteByte('\\')
			b.WriteRune(ch)
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}

