package hooks

import (
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/oguzsh/claudey/internal/hookio"
)

var devServerRe = regexp.MustCompile(`(npm run dev\b|pnpm( run)? dev\b|yarn dev\b|bun run dev\b)`)
var longRunningRe = regexp.MustCompile(`(npm (install|test)|pnpm (install|test)|yarn (install|test)?|bun (install|test)|cargo build|make\b|docker\b|pytest|vitest|playwright)`)
var gitPushRe = regexp.MustCompile(`git push`)
var randomDocRe = regexp.MustCompile(`\.(md|txt)$`)
var allowedDocRe = regexp.MustCompile(`(README|CLAUDE|AGENTS|CONTRIBUTING)\.md$`)
var plansPathRe = regexp.MustCompile(`\.claude/plans/`)
var prURLRe = regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/pull/\d+`)
var ghPrCreateRe = regexp.MustCompile(`gh pr create`)
var buildCmdRe = regexp.MustCompile(`(npm run build|pnpm build|yarn build)`)

// BlockDevServer blocks dev server commands outside tmux (PreToolUse Bash).
// Exits with code 2 to block the command.
func BlockDevServer(input map[string]any, raw []byte) int {
	if runtime.GOOS == "windows" {
		hookio.Passthrough(raw)
		return 0
	}

	cmd := hookio.GetToolInputString(input, "command")
	if devServerRe.MatchString(cmd) {
		hookio.Log("[Hook] BLOCKED: Dev server must run in tmux for log access")
		hookio.Log("[Hook] Use: tmux new-session -d -s dev \"npm run dev\"")
		hookio.Log("[Hook] Then: tmux attach -t dev")
		hookio.Passthrough(raw)
		return 2
	}

	hookio.Passthrough(raw)
	return 0
}

// TmuxReminder reminds about tmux for long-running commands (PreToolUse Bash).
func TmuxReminder(input map[string]any, raw []byte) {
	if runtime.GOOS == "windows" {
		hookio.Passthrough(raw)
		return
	}

	cmd := hookio.GetToolInputString(input, "command")
	if longRunningRe.MatchString(cmd) {
		// Check if already in tmux
		if !inTmux() {
			hookio.Log("[Hook] Consider running in tmux for session persistence")
			hookio.Log("[Hook] tmux new -s dev  |  tmux attach -t dev")
		}
	}

	hookio.Passthrough(raw)
}

func inTmux() bool {
	return os.Getenv("TMUX") != ""
}

// GitPushReminder reminds before git push (PreToolUse Bash).
func GitPushReminder(input map[string]any, raw []byte) {
	cmd := hookio.GetToolInputString(input, "command")
	if gitPushRe.MatchString(cmd) {
		hookio.Log("[Hook] Review changes before push...")
		hookio.Log("[Hook] Continuing with push (remove this hook to add interactive review)")
	}

	hookio.Passthrough(raw)
}

// BlockRandomDocs blocks creation of random .md/.txt files (PreToolUse Write).
// Returns exit code 2 to block.
func BlockRandomDocs(input map[string]any, raw []byte) int {
	filePath := hookio.GetToolInputString(input, "file_path")
	if filePath != "" && randomDocRe.MatchString(filePath) &&
		!allowedDocRe.MatchString(filePath) &&
		!plansPathRe.MatchString(filePath) {
		hookio.Log("[Hook] BLOCKED: Unnecessary documentation file creation")
		hookio.Logf("[Hook] File: %s", filePath)
		hookio.Log("[Hook] Use README.md for documentation instead")
		hookio.Passthrough(raw)
		return 2
	}

	hookio.Passthrough(raw)
	return 0
}

// PRCreatedLog logs PR URL after PR creation (PostToolUse Bash).
func PRCreatedLog(input map[string]any, raw []byte) {
	cmd := hookio.GetToolInputString(input, "command")
	if ghPrCreateRe.MatchString(cmd) {
		output := hookio.GetToolOutputString(input, "output")
		if m := prURLRe.FindString(output); m != "" {
			hookio.Logf("[Hook] PR created: %s", m)

			// Extract repo and PR number for review command
			parts := strings.Split(m, "/")
			if len(parts) >= 7 {
				repo := parts[3] + "/" + parts[4]
				pr := parts[6]
				hookio.Logf("[Hook] To review: gh pr review %s --repo %s", pr, repo)
			}
		}
	}

	hookio.Passthrough(raw)
}

// BuildAnalysis logs after build commands (PostToolUse Bash, async).
func BuildAnalysis(input map[string]any, raw []byte) {
	cmd := hookio.GetToolInputString(input, "command")
	if buildCmdRe.MatchString(cmd) {
		hookio.Log("[Hook] Build completed - async analysis running in background")
	}

	hookio.Passthrough(raw)
}



