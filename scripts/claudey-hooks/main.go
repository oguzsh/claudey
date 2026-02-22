// Command claudey-hooks is a single binary with subcommands for all Claude Code hooks and validators.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oguzsh/claudey/hooks/internal/hookio"
	"github.com/oguzsh/claudey/hooks/internal/hooks"
	"github.com/oguzsh/claudey/hooks/internal/validators"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: claudey-hooks <subcommand>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Hook subcommands:")
		fmt.Fprintln(os.Stderr, "  session-start          Load previous context on new session")
		fmt.Fprintln(os.Stderr, "  session-end            Persist session state on end")
		fmt.Fprintln(os.Stderr, "  pre-compact            Save state before compaction")
		fmt.Fprintln(os.Stderr, "  suggest-compact        Suggest manual compaction at intervals")
		fmt.Fprintln(os.Stderr, "  post-edit-format       Auto-format JS/TS with Prettier")
		fmt.Fprintln(os.Stderr, "  post-edit-typecheck    TypeScript check after .ts/.tsx edits")
		fmt.Fprintln(os.Stderr, "  post-edit-console-warn Warn about console.log after edits")
		fmt.Fprintln(os.Stderr, "  check-console-log      Check modified files for console.log")
		fmt.Fprintln(os.Stderr, "  evaluate-session       Evaluate session for patterns")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Inline hook subcommands:")
		fmt.Fprintln(os.Stderr, "  block-dev-server       Block dev servers outside tmux")
		fmt.Fprintln(os.Stderr, "  tmux-reminder          Remind about tmux for long commands")
		fmt.Fprintln(os.Stderr, "  git-push-reminder      Reminder before git push")
		fmt.Fprintln(os.Stderr, "  block-random-docs      Block random .md/.txt file creation")
		fmt.Fprintln(os.Stderr, "  pr-created-log         Log PR URL after creation")
		fmt.Fprintln(os.Stderr, "  build-analysis         Log after build commands")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "CLI utility subcommands:")
		fmt.Fprintln(os.Stderr, "  setup-pm               Setup preferred package manager")
		fmt.Fprintln(os.Stderr, "  skill-create-output    Render terminal UI for /skill-create")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Validator subcommands:")
		fmt.Fprintln(os.Stderr, "  validate-hooks         Validate hooks.json schema")
		fmt.Fprintln(os.Stderr, "  validate-agents        Validate agent frontmatter")
		fmt.Fprintln(os.Stderr, "  validate-commands      Validate command cross-references")
		fmt.Fprintln(os.Stderr, "  validate-rules         Validate rule files")
		fmt.Fprintln(os.Stderr, "  validate-skills        Validate skill directories")
		os.Exit(1)
	}

	subcmd := os.Args[1]

	switch subcmd {
	// --- Hook subcommands ---
	case "session-start":
		hooks.SessionStart()

	case "session-end":
		input, _, _ := readStdin()
		hooks.SessionEnd(input)

	case "pre-compact":
		hooks.PreCompact()

	case "suggest-compact":
		hooks.SuggestCompact()

	case "post-edit-format":
		input, raw, _ := readStdin()
		hooks.PostEditFormat(input, raw)

	case "post-edit-typecheck":
		input, raw, _ := readStdin()
		hooks.PostEditTypecheck(input, raw)

	case "post-edit-console-warn":
		input, raw, _ := readStdin()
		hooks.PostEditConsoleWarn(input, raw)

	case "check-console-log":
		_, raw, _ := readStdin()
		hooks.CheckConsoleLog(raw)

	case "evaluate-session":
		input, _, _ := readStdin()
		pluginRoot := findPluginRoot()
		hooks.EvaluateSession(input, pluginRoot)

	// --- Inline hook subcommands ---
	case "block-dev-server":
		input, raw, _ := readStdin()
		exitCode := hooks.BlockDevServer(input, raw)
		os.Exit(exitCode)

	case "tmux-reminder":
		input, raw, _ := readStdin()
		hooks.TmuxReminder(input, raw)

	case "git-push-reminder":
		input, raw, _ := readStdin()
		hooks.GitPushReminder(input, raw)

	case "block-random-docs":
		input, raw, _ := readStdin()
		exitCode := hooks.BlockRandomDocs(input, raw)
		os.Exit(exitCode)

	case "pr-created-log":
		input, raw, _ := readStdin()
		hooks.PRCreatedLog(input, raw)

	case "build-analysis":
		input, raw, _ := readStdin()
		hooks.BuildAnalysis(input, raw)

	// --- Validator subcommands ---
	case "validate-hooks":
		rootDir := findPluginRoot()
		if err := validators.ValidateHooks(rootDir); err != nil {
			os.Exit(1)
		}

	case "validate-agents":
		rootDir := findPluginRoot()
		if err := validators.ValidateAgents(rootDir); err != nil {
			os.Exit(1)
		}

	case "validate-commands":
		rootDir := findPluginRoot()
		if err := validators.ValidateCommands(rootDir); err != nil {
			os.Exit(1)
		}

	case "validate-rules":
		rootDir := findPluginRoot()
		if err := validators.ValidateRules(rootDir); err != nil {
			os.Exit(1)
		}

	case "validate-skills":
		rootDir := findPluginRoot()
		if err := validators.ValidateSkills(rootDir); err != nil {
			os.Exit(1)
		}

	// --- CLI utility subcommands ---
	case "setup-pm":
		hooks.SetupPackageManager(os.Args[2:])

	case "skill-create-output":
		hooks.SkillCreateOutput(os.Args[2:])

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}

// readStdin reads and parses JSON from stdin with a 5-second timeout.
func readStdin() (map[string]any, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return hookio.ReadStdinJSON(ctx, hookio.DefaultMaxSize)
}

// findPluginRoot locates the plugin root directory.
// It walks up from the binary's location or uses CLAUDE_PLUGIN_ROOT env var.
func findPluginRoot() string {
	// Check env var first
	if root := os.Getenv("CLAUDE_PLUGIN_ROOT"); root != "" {
		return root
	}

	// Try to find root relative to the binary
	exe, err := os.Executable()
	if err == nil {
		// Binary is at <root>/bin/claudey-hooks
		dir := filepath.Dir(filepath.Dir(exe))
		if _, err := os.Stat(filepath.Join(dir, "hooks", "hooks.json")); err == nil {
			return dir
		}
	}

	// Walk up from cwd
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "hooks", "hooks.json")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cwd
}
