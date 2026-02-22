package hooks

import (
	"strings"

	"github.com/oguzsh/claudey/hooks/internal/aliases"
	"github.com/oguzsh/claudey/hooks/internal/fileutil"
	"github.com/oguzsh/claudey/hooks/internal/hookio"
	"github.com/oguzsh/claudey/hooks/internal/pkgmanager"
	"github.com/oguzsh/claudey/hooks/internal/platform"
)

// SessionStart loads previous context and detects package manager.
func SessionStart() {
	sessionsDir := platform.SessionsDir()
	learnedDir := platform.LearnedSkillsDir()

	platform.EnsureDir(sessionsDir)
	platform.EnsureDir(learnedDir)

	// Check for recent session files (last 7 days)
	recentSessions := fileutil.FindFiles(sessionsDir, "*-session.tmp", 7, false)

	if len(recentSessions) > 0 {
		latest := recentSessions[0]
		hookio.Logf("[SessionStart] Found %d recent session(s)", len(recentSessions))
		hookio.Logf("[SessionStart] Latest: %s", latest.Path)

		content, ok := fileutil.ReadFile(latest.Path)
		if ok && !strings.Contains(content, "[Session context goes here]") {
			hookio.Output("Previous session summary:\n" + content)
		}
	}

	// Check for learned skills
	learnedSkills := fileutil.FindFiles(learnedDir, "*.md", 0, false)
	if len(learnedSkills) > 0 {
		hookio.Logf("[SessionStart] %d learned skill(s) available in %s", len(learnedSkills), learnedDir)
	}

	// Check for session aliases
	aliasList := aliases.List("", 5)
	if len(aliasList) > 0 {
		names := make([]string, len(aliasList))
		for i, a := range aliasList {
			names[i] = a.Name
		}
		hookio.Logf("[SessionStart] %d session alias(es) available: %s", len(aliasList), strings.Join(names, ", "))
		hookio.Log("[SessionStart] Use /sessions load <alias> to continue a previous session")
	}

	// Detect package manager
	pm := pkgmanager.Detect("")
	hookio.Logf("[SessionStart] Package manager: %s (%s)", pm.Name, pm.Source)

	if pm.Source == "default" {
		hookio.Log("[SessionStart] No package manager preference found.")
		hookio.Log(pkgmanager.SelectionPrompt())
	}
}
