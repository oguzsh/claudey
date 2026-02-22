package hooks

import (
	"path/filepath"

	"github.com/oguzsh/claudey/hooks/internal/datetime"
	"github.com/oguzsh/claudey/hooks/internal/fileutil"
	"github.com/oguzsh/claudey/hooks/internal/hookio"
	"github.com/oguzsh/claudey/hooks/internal/platform"
)

// PreCompact saves state before context compaction.
func PreCompact() {
	sessionsDir := platform.SessionsDir()
	compactionLog := filepath.Join(sessionsDir, "compaction-log.txt")

	platform.EnsureDir(sessionsDir)

	timestamp := datetime.DateTimeString()
	fileutil.AppendFile(compactionLog, "["+timestamp+"] Context compaction triggered\n")

	// Note compaction in most recent session file
	sessions := fileutil.FindFiles(sessionsDir, "*-session.tmp", 0, false)
	if len(sessions) > 0 {
		timeStr := datetime.TimeString()
		fileutil.AppendFile(sessions[0].Path, "\n---\n**[Compaction occurred at "+timeStr+"]** - Context was summarized\n")
	}

	hookio.Log("[PreCompact] State saved before compaction")
}
