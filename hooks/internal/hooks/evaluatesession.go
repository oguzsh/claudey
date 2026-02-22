package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/oguzsh/everything-claude-code/internal/fileutil"
	"github.com/oguzsh/everything-claude-code/internal/hookio"
	"github.com/oguzsh/everything-claude-code/internal/platform"
)

// EvaluateSession evaluates a session for extractable patterns.
// pluginRoot is the root directory of the plugin (for finding config).
func EvaluateSession(input map[string]any, pluginRoot string) {
	// Get transcript path from input or env
	transcriptPath, _ := input["transcript_path"].(string)
	if transcriptPath == "" {
		transcriptPath = os.Getenv("CLAUDE_TRANSCRIPT_PATH")
	}

	// Load config
	configFile := filepath.Join(pluginRoot, "skills", "continuous-learning", "config.json")

	minSessionLength := 10
	learnedSkillsPath := platform.LearnedSkillsDir()

	if content, ok := fileutil.ReadFile(configFile); ok {
		var config struct {
			MinSessionLength  int    `json:"min_session_length"`
			LearnedSkillsPath string `json:"learned_skills_path"`
		}
		if err := json.Unmarshal([]byte(content), &config); err != nil {
			hookio.Logf("[ContinuousLearning] Failed to parse config: %s, using defaults", err)
		} else {
			if config.MinSessionLength > 0 {
				minSessionLength = config.MinSessionLength
			}
			if config.LearnedSkillsPath != "" {
				path := config.LearnedSkillsPath
				if strings.HasPrefix(path, "~") {
					home := platform.HomeDir()
					path = home + path[1:]
				}
				learnedSkillsPath = path
			}
		}
	}

	platform.EnsureDir(learnedSkillsPath)

	if transcriptPath == "" || !fileutil.Exists(transcriptPath) {
		return
	}

	// Count user messages
	messageCount := fileutil.CountInFile(transcriptPath, `"type"\s*:\s*"user"`)

	if messageCount < minSessionLength {
		hookio.Logf("[ContinuousLearning] Session too short (%d messages), skipping", messageCount)
		return
	}

	hookio.Logf("[ContinuousLearning] Session has %d messages - evaluate for extractable patterns", messageCount)
	hookio.Logf("[ContinuousLearning] Save learned skills to: %s", learnedSkillsPath)
}

