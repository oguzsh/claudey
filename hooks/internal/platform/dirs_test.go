package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDir_CreatesNewDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir(%q) returned unexpected error: %v", dir, err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory to exist at %q, got error: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", dir)
	}
}

func TestEnsureDir_ExistingDirectoryIsNoOp(t *testing.T) {
	dir := t.TempDir() // already exists

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir(%q) on existing dir returned unexpected error: %v", dir, err)
	}
}

func TestEnsureDir_ReturnsErrorForInvalidPath(t *testing.T) {
	// Create a regular file, then try to create a directory inside it,
	// which should fail.
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "file.txt")

	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("setup: could not create file: %v", err)
	}

	invalidDir := filepath.Join(filePath, "subdir")
	err := EnsureDir(invalidDir)
	if err == nil {
		t.Fatalf("EnsureDir(%q) expected error when parent is a file, got nil", invalidDir)
	}
}

