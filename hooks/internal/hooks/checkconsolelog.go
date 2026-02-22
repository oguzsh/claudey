package hooks

import (
	"regexp"
	"strings"

	"github.com/oguzsh/everything-claude-code/internal/fileutil"
	"github.com/oguzsh/everything-claude-code/internal/gitutil"
	"github.com/oguzsh/everything-claude-code/internal/hookio"
)

var excludedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.test\.[jt]sx?$`),
	regexp.MustCompile(`\.spec\.[jt]sx?$`),
	regexp.MustCompile(`\.config\.[jt]s$`),
	regexp.MustCompile(`scripts/`),
	regexp.MustCompile(`__tests__/`),
	regexp.MustCompile(`__mocks__/`),
}

// CheckConsoleLog checks for console.log in modified JS/TS files.
func CheckConsoleLog(raw []byte) {
	if !gitutil.IsGitRepo() {
		hookio.Passthrough(raw)
		return
	}

	files := gitutil.ModifiedFiles([]string{`\.tsx?$`, `\.jsx?$`})

	// Filter out excluded patterns and non-existent files
	var filtered []string
	for _, f := range files {
		if !fileutil.Exists(f) {
			continue
		}
		excluded := false
		for _, pattern := range excludedPatterns {
			if pattern.MatchString(f) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, f)
		}
	}

	hasConsole := false
	for _, file := range filtered {
		content, ok := fileutil.ReadFile(file)
		if ok && strings.Contains(content, "console.log") {
			hookio.Logf("[Hook] WARNING: console.log found in %s", file)
			hasConsole = true
		}
	}

	if hasConsole {
		hookio.Log("[Hook] Remove console.log statements before committing")
	}

	hookio.Passthrough(raw)
}

