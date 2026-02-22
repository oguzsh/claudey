package fileutil

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestReadWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := WriteFile(path, "hello world"); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	content, ok := ReadFile(path)
	if !ok {
		t.Fatal("ReadFile returned not ok")
	}
	if content != "hello world" {
		t.Errorf("ReadFile = %q, want 'hello world'", content)
	}
}

func TestReadFile_Missing(t *testing.T) {
	_, ok := ReadFile("/nonexistent/file.txt")
	if ok {
		t.Error("ReadFile should return false for missing file")
	}
}

func TestAppendFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.txt")

	if err := WriteFile(path, "line1\n"); err != nil {
		t.Fatal(err)
	}
	if err := AppendFile(path, "line2\n"); err != nil {
		t.Fatal(err)
	}

	content, ok := ReadFile(path)
	if !ok {
		t.Fatal("ReadFile failed")
	}
	if content != "line1\nline2\n" {
		t.Errorf("content = %q, want 'line1\\nline2\\n'", content)
	}
}

func TestFindFiles(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	WriteFile(filepath.Join(dir, "a-session.tmp"), "a")
	WriteFile(filepath.Join(dir, "b-session.tmp"), "b")
	WriteFile(filepath.Join(dir, "readme.md"), "readme")

	results := FindFiles(dir, "*-session.tmp", 0, false)
	if len(results) != 2 {
		t.Errorf("FindFiles found %d files, want 2", len(results))
	}
}

func TestFindFiles_EmptyDir(t *testing.T) {
	results := FindFiles("/nonexistent", "*.tmp", 0, false)
	if results != nil {
		t.Errorf("FindFiles on nonexistent dir should return nil, got %v", results)
	}
}

func TestGrepFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.js")
	WriteFile(path, "line1\nconsole.log('debug')\nline3\nconsole.log('test')\n")

	results := GrepFile(path, `console\.log`)
	if len(results) != 2 {
		t.Errorf("GrepFile found %d matches, want 2", len(results))
	}
	if results[0].LineNumber != 2 {
		t.Errorf("first match line = %d, want 2", results[0].LineNumber)
	}
}

func TestCountInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	WriteFile(path, `"type":"user" and "type":"user" plus "type":"assistant"`)

	count := CountInFile(path, `"type"\s*:\s*"user"`)
	if count != 2 {
		t.Errorf("CountInFile = %d, want 2", count)
	}
}

func TestReplaceInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	WriteFile(path, "**Last Updated:** 10:00")

	ok := ReplaceInFile(path, "**Last Updated:** 10:00", "**Last Updated:** 11:30")
	if !ok {
		t.Error("ReplaceInFile returned false")
	}

	content, _ := ReadFile(path)
	if content != "**Last Updated:** 11:30" {
		t.Errorf("content = %q", content)
	}
}

func TestReplaceRegexInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	WriteFile(path, "**Last Updated:** 10:00\nother line")

	ok := ReplaceRegexInFile(path, `\*\*Last Updated:\*\*.*`, "**Last Updated:** 15:30")
	if !ok {
		t.Error("ReplaceRegexInFile returned false")
	}

	content, _ := ReadFile(path)
	expected := "**Last Updated:** 15:30\nother line"
	if content != expected {
		t.Errorf("content = %q, want %q", content, expected)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")
	WriteFile(path, "content")

	if !Exists(path) {
		t.Error("Exists returned false for existing file")
	}
	if Exists(filepath.Join(dir, "nope.txt")) {
		t.Error("Exists returned true for nonexistent file")
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		glob  string
		input string
		want  bool
	}{
		{"*.tmp", "session.tmp", true},
		{"*.tmp", "session.txt", false},
		{"*-session.tmp", "2024-01-01-session.tmp", true},
		{"test.?s", "test.js", true},
		{"test.?s", "test.ts", true},
	}

	for _, tt := range tests {
		re := "^" + globToRegex(tt.glob) + "$"
		matched, _ := regexp.MatchString(re, tt.input)
		if matched != tt.want {
			t.Errorf("globToRegex(%q) match %q = %v, want %v", tt.glob, tt.input, matched, tt.want)
		}
	}
}

func init() {
	// Suppress unused import warnings
	_ = os.TempDir
	_ = regexp.MatchString
}

