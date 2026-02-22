package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/oguzsh/claudey/internal/hookio"
	"github.com/oguzsh/claudey/internal/platform"
)

// SuggestCompact suggests manual compaction at logical intervals.
func SuggestCompact() {
	sessionID := os.Getenv("CLAUDE_SESSION_ID")
	if sessionID == "" {
		sessionID = "default"
	}
	counterFile := filepath.Join(platform.TempDir(), "claude-tool-count-"+sessionID)

	threshold := 50
	if envThreshold := os.Getenv("COMPACT_THRESHOLD"); envThreshold != "" {
		if v, err := strconv.Atoi(envThreshold); err == nil && v > 0 && v <= 10000 {
			threshold = v
		}
	}

	count := 1

	// Read and update counter file using fd-based operations
	f, err := os.OpenFile(counterFile, os.O_RDWR|os.O_CREATE, 0o644)
	if err == nil {
		buf := make([]byte, 64)
		n, _ := f.Read(buf)
		if n > 0 {
			parsed, err := strconv.Atoi(strings.TrimSpace(string(buf[:n])))
			if err == nil && parsed > 0 && parsed <= 1000000 {
				count = parsed + 1
			}
		}
		f.Truncate(0)
		f.Seek(0, 0)
		fmt.Fprint(f, count)
		f.Close()
	}

	if count == threshold {
		hookio.Logf("[StrategicCompact] %d tool calls reached - consider /compact if transitioning phases", threshold)
	}

	if count > threshold && (count-threshold)%25 == 0 {
		hookio.Logf("[StrategicCompact] %d tool calls - good checkpoint for /compact if context is stale", count)
	}
}
