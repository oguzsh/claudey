package hooks

import (
	"fmt"
	"os"

	"github.com/oguzsh/claudey/internal/pkgmanager"
)

func showSetupPMHelp() {
	fmt.Println(`
Package Manager Setup for Claude Code

Usage:
  claudey-hooks setup-pm [options] [package-manager]

Options:
  --detect        Detect and show current package manager
  --global <pm>   Set global preference (saves to ~/.claude/package-manager.json)
  --project <pm>  Set project preference (saves to .claude/package-manager.json)
  --list          List available package managers
  --help          Show this help message

Package Managers:
  npm             Node Package Manager (default with Node.js)
  pnpm            Fast, disk space efficient package manager
  yarn            Classic Yarn package manager
  bun             All-in-one JavaScript runtime & toolkit`)
}

func SetupPackageManager(args []string) {
	if len(args) == 0 || contains(args, "--help") || contains(args, "-h") {
		showSetupPMHelp()
		os.Exit(0)
	}

	if contains(args, "--detect") {
		detectAndShow()
		os.Exit(0)
	}

	if contains(args, "--list") {
		listAvailable()
		os.Exit(0)
	}

	globalIdx := indexOf(args, "--global")
	if globalIdx != -1 {
		if globalIdx+1 >= len(args) || stringsHasPrefix(args[globalIdx+1], "-") {
			fmt.Fprintln(os.Stderr, "Error: --global requires a package manager name")
			os.Exit(1)
		}
		setGlobal(args[globalIdx+1])
		os.Exit(0)
	}

	projectIdx := indexOf(args, "--project")
	if projectIdx != -1 {
		if projectIdx+1 >= len(args) || stringsHasPrefix(args[projectIdx+1], "-") {
			fmt.Fprintln(os.Stderr, "Error: --project requires a package manager name")
			os.Exit(1)
		}
		setProject(args[projectIdx+1])
		os.Exit(0)
	}

	// If just a package manager name is provided, set it globally
	pmName := args[0]
	// Basic validation if it's one of the supported ones
	switch pmName {
	case "npm", "pnpm", "yarn", "bun":
		setGlobal(pmName)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown option or package manager \"%s\"\n", pmName)
		showSetupPMHelp()
		os.Exit(1)
	}
}

func detectAndShow() {
	res := pkgmanager.Detect("")
	available := pkgmanager.GetAvailable()
	fromLock := pkgmanager.DetectFromLockFile("")
	fromPkg := pkgmanager.DetectFromPackageJSON("")

	fmt.Println("\n=== Package Manager Detection ===\n")

	fmt.Println("Current selection:")
	fmt.Printf("  Package Manager: %s\n", res.Name)
	fmt.Printf("  Source: %s\n\n", res.Source)

	fmt.Println("Detection results:")
	if fromPkg == "" {
		fmt.Println("  From package.json: not specified")
	} else {
		fmt.Printf("  From package.json: %s\n", fromPkg)
	}
	if fromLock == "" {
		fmt.Println("  From lock file: not found")
	} else {
		fmt.Printf("  From lock file: %s\n", fromLock)
	}

	envVar := os.Getenv("CLAUDE_PACKAGE_MANAGER")
	if envVar == "" {
		fmt.Println("  Environment var: not set\n")
	} else {
		fmt.Printf("  Environment var: %s\n\n", envVar)
	}

	fmt.Println("Available package managers:")
	allPMs := []string{"npm", "pnpm", "yarn", "bun"}
	for _, pmName := range allPMs {
		installed := contains(available, pmName)
		indicator := "✗"
		if installed {
			indicator = "✓"
		}
		current := ""
		if pmName == res.Name {
			current = " (current)"
		}
		fmt.Printf("  %s %s%s\n", indicator, pmName, current)
	}

	fmt.Println("\nCommands:")
	fmt.Printf("  Install: %s\n", res.Config.InstallCmd)
	fmt.Printf("  Run script: %s [script-name]\n", res.Config.RunCmd)
	fmt.Printf("  Execute binary: %s [binary-name]\n\n", res.Config.ExecCmd)
}

func listAvailable() {
	available := pkgmanager.GetAvailable()
	res := pkgmanager.Detect("")

	fmt.Println("\nAvailable Package Managers:\n")

	// Create dummy configs to show info
	dummyConfigs := map[string]pkgmanager.Config{
		"npm":  {LockFile: "package-lock.json", InstallCmd: "npm install", RunCmd: "npm run"},
		"pnpm": {LockFile: "pnpm-lock.yaml", InstallCmd: "pnpm install", RunCmd: "pnpm"},
		"yarn": {LockFile: "yarn.lock", InstallCmd: "yarn", RunCmd: "yarn"},
		"bun":  {LockFile: "bun.lockb", InstallCmd: "bun install", RunCmd: "bun run"},
	}

	for _, pmName := range []string{"npm", "pnpm", "yarn", "bun"} {
		config := dummyConfigs[pmName]
		installed := contains(available, pmName)
		current := ""
		if pmName == res.Name {
			current = " (current)"
		}

		fmt.Printf("%s%s\n", pmName, current)
		instStr := "No"
		if installed {
			instStr = "Yes"
		}
		fmt.Printf("  Installed: %s\n", instStr)
		fmt.Printf("  Lock file: %s\n", config.LockFile)
		fmt.Printf("  Install: %s\n", config.InstallCmd)
		fmt.Printf("  Run: %s\n\n", config.RunCmd)
	}
}

func setGlobal(pmName string) {
	available := pkgmanager.GetAvailable()
	if !contains(available, pmName) {
		fmt.Printf("Warning: %s is not installed on your system\n", pmName)
	}

	err := pkgmanager.SetPreferred(pmName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n✓ Global preference set to: %s\n", pmName)
	fmt.Println("  Saved to: ~/.claude/package-manager.json\n")
}

func setProject(pmName string) {
	err := pkgmanager.SetProjectPreferred(pmName, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n✓ Project preference set to: %s\n", pmName)
	fmt.Println("  Saved to: .claude/package-manager.json\n")
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func indexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return -1
}

func stringsHasPrefix(s, prefix string) bool {
	if len(s) >= len(prefix) && s[0:len(prefix)] == prefix {
		return true
	}
	return false
}
