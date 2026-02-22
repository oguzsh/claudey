package aliases

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oguzsh/claudey/internal/platform"
)

func setupTestHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	// Override HOME so platform.ClaudeDir() points to our temp dir
	t.Setenv("HOME", dir)
	// Ensure .claude dir exists
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
}

func TestLoadDefault(t *testing.T) {
	setupTestHome(t)
	f := Load()
	if f.Version != aliasVersion {
		t.Errorf("Version = %q, want %q", f.Version, aliasVersion)
	}
	if len(f.Aliases) != 0 {
		t.Errorf("Aliases count = %d, want 0", len(f.Aliases))
	}
}

func TestSetAndResolve(t *testing.T) {
	setupTestHome(t)

	result := Set("myalias", "/path/to/session", "My Session")
	if !result.Success {
		t.Fatalf("Set failed: %s", result.Error)
	}
	if !result.IsNew {
		t.Error("expected IsNew = true")
	}

	entry := Resolve("myalias")
	if entry == nil {
		t.Fatal("Resolve returned nil")
	}
	if entry.SessionPath != "/path/to/session" {
		t.Errorf("SessionPath = %q", entry.SessionPath)
	}
	if entry.Title != "My Session" {
		t.Errorf("Title = %q", entry.Title)
	}
}

func TestSet_Validation(t *testing.T) {
	setupTestHome(t)

	tests := []struct {
		alias string
		path  string
		err   string
	}{
		{"", "/path", "empty"},
		{"a", "", "empty path"},
		{"invalid!name", "/path", "unsafe chars"},
		{"list", "/path", "reserved"},
	}

	for _, tt := range tests {
		result := Set(tt.alias, tt.path, "")
		if result.Success {
			t.Errorf("Set(%q, %q) should fail (%s)", tt.alias, tt.path, tt.err)
		}
	}
}

func TestDelete(t *testing.T) {
	setupTestHome(t)

	Set("todelete", "/path", "")
	result := Delete("todelete")
	if !result.Success {
		t.Fatalf("Delete failed: %s", result.Error)
	}

	if entry := Resolve("todelete"); entry != nil {
		t.Error("alias should be deleted")
	}
}

func TestDelete_NotFound(t *testing.T) {
	setupTestHome(t)
	result := Delete("nonexistent")
	if result.Success {
		t.Error("Delete should fail for nonexistent alias")
	}
}

func TestList(t *testing.T) {
	setupTestHome(t)

	Set("alpha", "/path/a", "Alpha")
	Set("beta", "/path/b", "Beta")

	aliases := List("", 0)
	if len(aliases) != 2 {
		t.Errorf("List count = %d, want 2", len(aliases))
	}

	// Test search
	aliases = List("alp", 0)
	if len(aliases) != 1 {
		t.Errorf("List(search='alp') count = %d, want 1", len(aliases))
	}

	// Test limit
	aliases = List("", 1)
	if len(aliases) != 1 {
		t.Errorf("List(limit=1) count = %d, want 1", len(aliases))
	}
}

func TestRename(t *testing.T) {
	setupTestHome(t)

	Set("oldname", "/path", "Title")
	result := Rename("oldname", "newname")
	if !result.Success {
		t.Fatalf("Rename failed: %s", result.Error)
	}

	if Resolve("oldname") != nil {
		t.Error("old alias should not exist")
	}
	if Resolve("newname") == nil {
		t.Error("new alias should exist")
	}
}

func TestResolveSessionAlias(t *testing.T) {
	setupTestHome(t)

	Set("myalias", "/path/to/session", "")

	if got := ResolveSessionAlias("myalias"); got != "/path/to/session" {
		t.Errorf("ResolveSessionAlias = %q, want /path/to/session", got)
	}
	if got := ResolveSessionAlias("notanalias"); got != "notanalias" {
		t.Errorf("ResolveSessionAlias (passthrough) = %q, want notanalias", got)
	}
}

func TestCleanup(t *testing.T) {
	setupTestHome(t)

	Set("exists", "/exists", "")
	Set("gone", "/gone", "")

	checked, removed := Cleanup(func(path string) bool {
		return path == "/exists"
	})

	if checked != 2 {
		t.Errorf("checked = %d, want 2", checked)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	if Resolve("gone") != nil {
		t.Error("cleaned alias should be removed")
	}
	if Resolve("exists") == nil {
		t.Error("existing alias should remain")
	}
}

// Ensure platform package works with test home
func init() {
	_ = platform.ClaudeDir
}



