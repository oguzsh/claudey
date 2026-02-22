// Package gitutil provides git repository detection and modified file listing.
package gitutil

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/oguzsh/everything-claude-code/internal/sysutil"
)

// IsGitRepo returns true if the current directory is in a git repository.
func IsGitRepo() bool {
	return sysutil.RunCommand("git rev-parse --git-dir", "").Success
}

// RepoName returns the git repository name (basename of toplevel).
func RepoName() string {
	result := sysutil.RunCommand("git rev-parse --show-toplevel", "")
	if !result.Success {
		return ""
	}
	return filepath.Base(result.Output)
}

// ProjectName returns the project name from git repo or cwd basename.
func ProjectName() string {
	if name := RepoName(); name != "" {
		return name
	}
	result := sysutil.RunCommand("pwd", "")
	if result.Success {
		return filepath.Base(result.Output)
	}
	return ""
}

// ModifiedFiles returns git modified files, optionally filtered by regex patterns.
// Invalid patterns are silently skipped.
func ModifiedFiles(patterns []string) []string {
	if !IsGitRepo() {
		return nil
	}

	result := sysutil.RunCommand("git diff --name-only HEAD", "")
	if !result.Success {
		return nil
	}

	files := strings.Split(result.Output, "\n")
	var filtered []string
	for _, f := range files {
		if strings.TrimSpace(f) != "" {
			filtered = append(filtered, f)
		}
	}

	if len(patterns) == 0 {
		return filtered
	}

	// Compile patterns, skip invalid ones
	var compiled []*regexp.Regexp
	for _, p := range patterns {
		if p == "" {
			continue
		}
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		compiled = append(compiled, re)
	}

	if len(compiled) == 0 {
		return filtered
	}

	var matched []string
	for _, f := range filtered {
		for _, re := range compiled {
			if re.MatchString(f) {
				matched = append(matched, f)
				break
			}
		}
	}

	return matched
}

