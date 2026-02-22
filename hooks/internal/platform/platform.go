// Package platform provides OS detection and directory path helpers
// for the Claude Code hook system.
package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// OS detection flags derived from runtime.GOOS.
var (
	IsWindows = runtime.GOOS == "windows"
	IsMacOS   = runtime.GOOS == "darwin"
	IsLinux   = runtime.GOOS == "linux"
)

// HomeDir returns the current user's home directory.
// It panics if the home directory cannot be determined.
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("platform: unable to determine home directory: " + err.Error())
	}
	return home
}

// ClaudeDir returns the path to the Claude configuration directory (~/.claude).
func ClaudeDir() string {
	return filepath.Join(HomeDir(), ".claude")
}

// SessionsDir returns the path to the sessions directory (~/.claude/sessions).
func SessionsDir() string {
	return filepath.Join(ClaudeDir(), "sessions")
}

// LearnedSkillsDir returns the path to the learned skills directory (~/.claude/skills/learned).
func LearnedSkillsDir() string {
	return filepath.Join(ClaudeDir(), "skills", "learned")
}

// TempDir returns the OS default temporary directory.
func TempDir() string {
	return os.TempDir()
}

