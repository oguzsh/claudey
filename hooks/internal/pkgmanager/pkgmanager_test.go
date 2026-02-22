package pkgmanager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oguzsh/claudey/internal/fileutil"
)

func TestDetectFromLockFile(t *testing.T) {
	dir := t.TempDir()

	// No lock file
	if got := DetectFromLockFile(dir); got != "" {
		t.Errorf("DetectFromLockFile (empty dir) = %q, want empty", got)
	}

	// Create pnpm lock file
	fileutil.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), "")
	if got := DetectFromLockFile(dir); got != "pnpm" {
		t.Errorf("DetectFromLockFile (pnpm) = %q, want pnpm", got)
	}
}

func TestDetectFromPackageJSON(t *testing.T) {
	dir := t.TempDir()

	// No package.json
	if got := DetectFromPackageJSON(dir); got != "" {
		t.Errorf("DetectFromPackageJSON (no file) = %q, want empty", got)
	}

	// With packageManager field
	fileutil.WriteFile(filepath.Join(dir, "package.json"), `{"packageManager": "pnpm@8.6.0"}`)
	if got := DetectFromPackageJSON(dir); got != "pnpm" {
		t.Errorf("DetectFromPackageJSON = %q, want pnpm", got)
	}
}

func TestDetect_EnvVar(t *testing.T) {
	os.Setenv("CLAUDE_PACKAGE_MANAGER", "yarn")
	defer os.Unsetenv("CLAUDE_PACKAGE_MANAGER")

	result := Detect(t.TempDir())
	if result.Name != "yarn" {
		t.Errorf("Detect with env = %q, want yarn", result.Name)
	}
	if result.Source != "environment" {
		t.Errorf("source = %q, want environment", result.Source)
	}
}

func TestDetect_Default(t *testing.T) {
	os.Unsetenv("CLAUDE_PACKAGE_MANAGER")
	result := Detect(t.TempDir())
	if result.Name != "npm" {
		t.Errorf("Detect default = %q, want npm", result.Name)
	}
	if result.Source != "default" {
		t.Errorf("source = %q, want default", result.Source)
	}
}

func TestGetRunCommand(t *testing.T) {
	os.Setenv("CLAUDE_PACKAGE_MANAGER", "npm")
	defer os.Unsetenv("CLAUDE_PACKAGE_MANAGER")

	tests := []struct {
		script string
		want   string
	}{
		{"install", "npm install"},
		{"test", "npm test"},
		{"build", "npm run build"},
		{"dev", "npm run dev"},
		{"lint", "npm run lint"},
	}

	for _, tt := range tests {
		got, err := GetRunCommand(tt.script, "")
		if err != nil {
			t.Errorf("GetRunCommand(%q) error: %v", tt.script, err)
			continue
		}
		if got != tt.want {
			t.Errorf("GetRunCommand(%q) = %q, want %q", tt.script, got, tt.want)
		}
	}
}

func TestGetRunCommand_Unsafe(t *testing.T) {
	_, err := GetRunCommand("test; rm -rf /", "")
	if err == nil {
		t.Error("GetRunCommand should reject unsafe script names")
	}
}

func TestGetExecCommand(t *testing.T) {
	os.Setenv("CLAUDE_PACKAGE_MANAGER", "npm")
	defer os.Unsetenv("CLAUDE_PACKAGE_MANAGER")

	got, err := GetExecCommand("prettier", "--write .", "")
	if err != nil {
		t.Fatalf("GetExecCommand error: %v", err)
	}
	if got != "npx prettier --write ." {
		t.Errorf("GetExecCommand = %q, want 'npx prettier --write .'", got)
	}
}

func TestCommandPattern(t *testing.T) {
	pattern := CommandPattern("dev")
	if pattern == "" {
		t.Error("CommandPattern returned empty string")
	}
	// Should contain all PM variants
	for _, pm := range []string{"npm", "pnpm", "yarn", "bun"} {
		if !containsStr(pattern, pm) {
			t.Errorf("CommandPattern('dev') missing %s", pm)
		}
	}
}

func TestSelectionPrompt(t *testing.T) {
	prompt := SelectionPrompt()
	if prompt == "" {
		t.Error("SelectionPrompt returned empty string")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
