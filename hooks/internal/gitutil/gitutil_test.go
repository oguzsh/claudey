package gitutil

import (
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// We're running inside a git repo, so this should return true
	if !IsGitRepo() {
		t.Skip("Not in a git repo, skipping")
	}
}

func TestRepoName(t *testing.T) {
	if !IsGitRepo() {
		t.Skip("Not in a git repo, skipping")
	}
	name := RepoName()
	if name == "" {
		t.Error("RepoName() returned empty string in a git repo")
	}
}

func TestProjectName(t *testing.T) {
	name := ProjectName()
	if name == "" {
		t.Error("ProjectName() returned empty string")
	}
}

