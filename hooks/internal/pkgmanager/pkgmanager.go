// Package pkgmanager provides package manager detection and selection.
package pkgmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/oguzsh/claudey/hooks/internal/fileutil"
	"github.com/oguzsh/claudey/hooks/internal/platform"
	"github.com/oguzsh/claudey/hooks/internal/sysutil"
)

// Config holds the commands for a package manager.
type Config struct {
	Name       string
	LockFile   string
	InstallCmd string
	RunCmd     string
	ExecCmd    string
	TestCmd    string
	BuildCmd   string
	DevCmd     string
}

// Result holds the detected package manager and its source.
type Result struct {
	Name   string
	Config Config
	Source string
}

var managers = map[string]Config{
	"npm": {
		Name: "npm", LockFile: "package-lock.json",
		InstallCmd: "npm install", RunCmd: "npm run", ExecCmd: "npx",
		TestCmd: "npm test", BuildCmd: "npm run build", DevCmd: "npm run dev",
	},
	"pnpm": {
		Name: "pnpm", LockFile: "pnpm-lock.yaml",
		InstallCmd: "pnpm install", RunCmd: "pnpm", ExecCmd: "pnpm dlx",
		TestCmd: "pnpm test", BuildCmd: "pnpm build", DevCmd: "pnpm dev",
	},
	"yarn": {
		Name: "yarn", LockFile: "yarn.lock",
		InstallCmd: "yarn", RunCmd: "yarn", ExecCmd: "yarn dlx",
		TestCmd: "yarn test", BuildCmd: "yarn build", DevCmd: "yarn dev",
	},
	"bun": {
		Name: "bun", LockFile: "bun.lockb",
		InstallCmd: "bun install", RunCmd: "bun run", ExecCmd: "bunx",
		TestCmd: "bun test", BuildCmd: "bun run build", DevCmd: "bun run dev",
	},
}

var detectionPriority = []string{"pnpm", "bun", "yarn", "npm"}

func configPath() string {
	return filepath.Join(platform.ClaudeDir(), "package-manager.json")
}

// DetectFromLockFile checks for lock files in projectDir.
func DetectFromLockFile(projectDir string) string {
	for _, name := range detectionPriority {
		cfg := managers[name]
		if fileutil.Exists(filepath.Join(projectDir, cfg.LockFile)) {
			return name
		}
	}
	return ""
}

// DetectFromPackageJSON checks the packageManager field in package.json.
func DetectFromPackageJSON(projectDir string) string {
	content, ok := fileutil.ReadFile(filepath.Join(projectDir, "package.json"))
	if !ok {
		return ""
	}

	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return ""
	}

	if pkg.PackageManager == "" {
		return ""
	}

	name := strings.Split(pkg.PackageManager, "@")[0]
	if _, ok := managers[name]; ok {
		return name
	}
	return ""
}

// GetAvailable returns package managers installed on the system.
func GetAvailable() []string {
	var available []string
	for name := range managers {
		if sysutil.CommandExists(name) {
			available = append(available, name)
		}
	}
	return available
}

// Detect returns the package manager for the current project.
// Detection priority:
// 1. CLAUDE_PACKAGE_MANAGER env var
// 2. Project-specific .claude/package-manager.json
// 3. package.json packageManager field
// 4. Lock file detection
// 5. Global ~/.claude/package-manager.json
// 6. Default to npm
func Detect(projectDir string) Result {
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}

	// 1. Environment variable
	if envPm := os.Getenv("CLAUDE_PACKAGE_MANAGER"); envPm != "" {
		if cfg, ok := managers[envPm]; ok {
			return Result{Name: envPm, Config: cfg, Source: "environment"}
		}
	}

	// 2. Project-specific config
	projConfigPath := filepath.Join(projectDir, ".claude", "package-manager.json")
	if content, ok := fileutil.ReadFile(projConfigPath); ok {
		var cfg struct {
			PackageManager string `json:"packageManager"`
		}
		if json.Unmarshal([]byte(content), &cfg) == nil && cfg.PackageManager != "" {
			if c, ok := managers[cfg.PackageManager]; ok {
				return Result{Name: cfg.PackageManager, Config: c, Source: "project-config"}
			}
		}
	}

	// 3. package.json packageManager field
	if name := DetectFromPackageJSON(projectDir); name != "" {
		return Result{Name: name, Config: managers[name], Source: "package.json"}
	}

	// 4. Lock file detection
	if name := DetectFromLockFile(projectDir); name != "" {
		return Result{Name: name, Config: managers[name], Source: "lock-file"}
	}

	// 5. Global config
	if content, ok := fileutil.ReadFile(configPath()); ok {
		var cfg struct {
			PackageManager string `json:"packageManager"`
		}
		if json.Unmarshal([]byte(content), &cfg) == nil && cfg.PackageManager != "" {
			if c, ok := managers[cfg.PackageManager]; ok {
				return Result{Name: cfg.PackageManager, Config: c, Source: "global-config"}
			}
		}
	}

	// 6. Default to npm
	return Result{Name: "npm", Config: managers["npm"], Source: "default"}
}

// SetPreferred saves the global package manager preference.
func SetPreferred(name string) error {
	if _, ok := managers[name]; !ok {
		return fmt.Errorf("unknown package manager: %s", name)
	}

	data := map[string]string{
		"packageManager": name,
		"setAt":          time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.MarshalIndent(data, "", "  ")
	return fileutil.WriteFile(configPath(), string(b))
}

// SetProjectPreferred saves the project package manager preference.
func SetProjectPreferred(name string, projectDir string) error {
	if _, ok := managers[name]; !ok {
		return fmt.Errorf("unknown package manager: %s", name)
	}

	data := map[string]string{
		"packageManager": name,
		"setAt":          time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.MarshalIndent(data, "", "  ")
	return fileutil.WriteFile(filepath.Join(projectDir, ".claude", "package-manager.json"), string(b))
}

var safeNameRegex = regexp.MustCompile(`^[@a-zA-Z0-9_./-]+$`)
var safeArgsRegex = regexp.MustCompile(`^[@a-zA-Z0-9\s_./:=,'"*+\-]+$`)

// GetRunCommand returns the command to run a named script.
func GetRunCommand(script string, projectDir string) (string, error) {
	if script == "" {
		return "", fmt.Errorf("script name must be a non-empty string")
	}
	if !safeNameRegex.MatchString(script) {
		return "", fmt.Errorf("script name contains unsafe characters: %s", script)
	}

	pm := Detect(projectDir)
	switch script {
	case "install":
		return pm.Config.InstallCmd, nil
	case "test":
		return pm.Config.TestCmd, nil
	case "build":
		return pm.Config.BuildCmd, nil
	case "dev":
		return pm.Config.DevCmd, nil
	default:
		return pm.Config.RunCmd + " " + script, nil
	}
}

// GetExecCommand returns the command to execute a package binary.
func GetExecCommand(binary string, args string, projectDir string) (string, error) {
	if binary == "" {
		return "", fmt.Errorf("binary name must be a non-empty string")
	}
	if !safeNameRegex.MatchString(binary) {
		return "", fmt.Errorf("binary name contains unsafe characters: %s", binary)
	}
	if args != "" && !safeArgsRegex.MatchString(args) {
		return "", fmt.Errorf("arguments contain unsafe characters: %s", args)
	}

	pm := Detect(projectDir)
	cmd := pm.Config.ExecCmd + " " + binary
	if args != "" {
		cmd += " " + args
	}
	return cmd, nil
}

// SelectionPrompt returns a help message for package manager selection.
func SelectionPrompt() string {
	var b strings.Builder
	b.WriteString("[PackageManager] No package manager preference detected.\n")
	b.WriteString("Supported package managers: npm, pnpm, yarn, bun\n")
	b.WriteString("\nTo set your preferred package manager:\n")
	b.WriteString("  - Global: Set CLAUDE_PACKAGE_MANAGER environment variable\n")
	b.WriteString("  - Or add to ~/.claude/package-manager.json: {\"packageManager\": \"pnpm\"}\n")
	b.WriteString("  - Or add to package.json: {\"packageManager\": \"pnpm@8\"}\n")
	b.WriteString("  - Or add a lock file to your project (e.g., pnpm-lock.yaml)\n")
	return b.String()
}

// CommandPattern generates a regex pattern matching commands for all package managers.
func CommandPattern(action string) string {
	action = strings.TrimSpace(action)
	var patterns []string

	switch action {
	case "dev":
		patterns = []string{
			`npm run dev`,
			`pnpm( run)? dev`,
			`yarn dev`,
			`bun run dev`,
		}
	case "install":
		patterns = []string{
			`npm install`,
			`pnpm install`,
			`yarn( install)?`,
			`bun install`,
		}
	case "test":
		patterns = []string{
			`npm test`,
			`pnpm test`,
			`yarn test`,
			`bun test`,
		}
	case "build":
		patterns = []string{
			`npm run build`,
			`pnpm( run)? build`,
			`yarn build`,
			`bun run build`,
		}
	default:
		escaped := regexp.QuoteMeta(action)
		patterns = []string{
			`npm run ` + escaped,
			`pnpm( run)? ` + escaped,
			`yarn ` + escaped,
			`bun run ` + escaped,
		}
	}

	return "(" + strings.Join(patterns, "|") + ")"
}
