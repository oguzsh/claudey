package hooks

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

// ANSI color codes
const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
	ansiWhite   = "\x1b[37m"
	ansiGray    = "\x1b[90m"
	ansiBgCyan  = "\x1b[46m"
)

func bold(s string) string { return ansiBold + s + ansiReset }
func cyan(s string) string { return ansiCyan + s + ansiReset }
func green(s string) string { return ansiGreen + s + ansiReset }
func yellow(s string) string { return ansiYellow + s + ansiReset }
func magenta(s string) string { return ansiMagenta + s + ansiReset }
func gray(s string) string { return ansiGray + s + ansiReset }
func white(s string) string { return ansiWhite + s + ansiReset }
func red(s string) string { return ansiRed + s + ansiReset }
func dim(s string) string { return ansiDim + s + ansiReset }

// Box drawing characters
var boxChars = map[string]string{
	"topLeft":     "╭",
	"topRight":    "╮",
	"bottomLeft":  "╰",
	"bottomRight": "╯",
	"horizontal":  "─",
	"vertical":    "│",
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Helpers

func stripAnsi(str string) string {
	// Simple ANSI stripper
	var result strings.Builder
	inAnsi := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\x1b' {
			inAnsi = true
			continue
		}
		if inAnsi {
			if str[i] == 'm' {
				inAnsi = false
			}
			continue
		}
		result.WriteByte(str[i])
	}
	return result.String()
}

func box(title, content string, width int) string {
	lines := strings.Split(content, "\n")
	titleLength := len(stripAnsi(title))

	padding := width - titleLength - 5
	if padding < 0 {
		padding = 0
	}

	top := fmt.Sprintf("%s%s %s %s%s", boxChars["topLeft"], boxChars["horizontal"], bold(cyan(title)), strings.Repeat(boxChars["horizontal"], padding), boxChars["topRight"])
	bottom := fmt.Sprintf("%s%s%s", boxChars["bottomLeft"], strings.Repeat(boxChars["horizontal"], width-2), boxChars["bottomRight"])

	var middle strings.Builder
	for i, line := range lines {
		lineLength := len(stripAnsi(line))
		padding := width - 4 - lineLength
		if padding < 0 {
			padding = 0
		}
		middle.WriteString(fmt.Sprintf("%s %s%s %s", boxChars["vertical"], line, strings.Repeat(" ", padding), boxChars["vertical"]))
		if i < len(lines)-1 {
			middle.WriteString("\n")
		}
	}

	return fmt.Sprintf("%s\n%s\n%s", top, middle.String(), bottom)
}

func progressBar(percent float64, width int) string {
	filled := int(math.Round(float64(width) * percent / 100))
	if filled > width {
		filled = width
	} else if filled < 0 {
		filled = 0
	}
	empty := width - filled

	bar := green(strings.Repeat("█", filled)) + gray(strings.Repeat("░", empty))
	return fmt.Sprintf("%s %s%%", bar, bold(fmt.Sprintf("%.0f", percent)))
}

type Step struct {
	Name     string
	Duration time.Duration
}

func animateProgress(label string, steps []Step) {
	fmt.Printf("\n%s %s...\n", cyan("⏳"), label)

	for i, step := range steps {
		frame := spinnerFrames[i%len(spinnerFrames)]
		fmt.Printf("   %s %s", gray(frame), step.Name)
		time.Sleep(step.Duration)
		fmt.Printf("\r   %s %s\n", green("✓"), step.Name)
	}
}

func header(repoName string) {
	subtitle := fmt.Sprintf("Extracting patterns from %s", cyan(repoName))
	subtitleLen := len(stripAnsi(subtitle))
	pad := 59 - subtitleLen
	if pad < 0 {
		pad = 0
	}

	fmt.Println("\n")
	fmt.Println(bold(magenta("╔════════════════════════════════════════════════════════════════╗")))
	fmt.Println(bold(magenta("║")) + bold("  🔮 ECC Skill Creator                                          ") + bold(magenta("║")))
	fmt.Println(bold(magenta("║")) + fmt.Sprintf("     %s%s", subtitle, strings.Repeat(" ", pad)) + bold(magenta("║")))
	fmt.Println(bold(magenta("╚════════════════════════════════════════════════════════════════╝")))
	fmt.Println("")
}

func SkillCreateOutput(args []string) {
	// If it fails or we want to pass repo name properly
	repoName := "PMX"
	if len(args) > 0 {
		repoName = args[0]
	}

	header(repoName)

	steps := []Step{
		{Name: "Parsing git history...", Duration: 300 * time.Millisecond},
		{Name: fmt.Sprintf("Found %s commits", yellow("200")), Duration: 200 * time.Millisecond},
		{Name: "Analyzing commit patterns...", Duration: 400 * time.Millisecond},
		{Name: "Detecting file co-changes...", Duration: 300 * time.Millisecond},
		{Name: "Identifying workflows...", Duration: 400 * time.Millisecond},
		{Name: "Extracting architecture patterns...", Duration: 300 * time.Millisecond},
	}
	animateProgress("Analyzing Repository", steps)

	fmt.Println("\n")
	resultsContent := fmt.Sprintf(`
%s %s
%s       %s
%s     %s
%s    %s
`, bold("Commits Analyzed:"), yellow("200"),
		bold("Time Range:"), gray("Nov 2024 - Jan 2025"),
		bold("Contributors:"), cyan("4"),
		bold("Files Tracked:"), green("847"))
	fmt.Println(box("📊 Analysis Results", resultsContent, 60))

	fmt.Println("\n")
	fmt.Println(bold(cyan("🔍 Key Patterns Discovered:")))
	fmt.Println(gray(strings.Repeat("─", 50)))

	patterns := []struct {
		name       string
		trigger    string
		confidence float64
		evidence   string
	}{
		{"Conventional Commits", "when writing commit messages", 85, "Found in 150/200 commits (feat:, fix:, refactor:)"},
		{"Client/Server Component Split", "when creating Next.js pages", 90, "Observed in markets/, premarkets/, portfolio/"},
		{"Service Layer Architecture", "when adding backend logic", 85, "Business logic in services/, not routes/"},
		{"TDD with E2E Tests", "when adding features", 75, "9 E2E test files, test(e2e) commits common"},
	}

	for i, pattern := range patterns {
		confBar := progressBar(pattern.confidence, 15)
		fmt.Printf(`
  %s %s
     %s %s
     %s %s
     %s
`, bold(yellow(fmt.Sprintf("%d.", i+1))), bold(pattern.name),
			gray("Trigger:"), pattern.trigger,
			gray("Confidence:"), confBar,
			dim(pattern.evidence))
	}

	instincts := []struct{ name string; confidence float64 }{
		{"pmx-conventional-commits", 85},
		{"pmx-client-component-pattern", 90},
		{"pmx-service-layer", 85},
		{"pmx-e2e-test-location", 90},
		{"pmx-package-manager", 95},
		{"pmx-hot-path-caution", 90},
	}

	fmt.Println("\n")
	var instBuilder strings.Builder
	for i, inst := range instincts {
		instBuilder.WriteString(fmt.Sprintf("%s %s %s", yellow(fmt.Sprintf("%d.", i+1)), bold(inst.name), gray(fmt.Sprintf("(%d%%)", int(inst.confidence)))))
		if i < len(instincts)-1 {
			instBuilder.WriteString("\n")
		}
	}
	fmt.Println(box("🧠 Instincts Generated", instBuilder.String(), 60))

	// Output paths
	skillPath := fmt.Sprintf(".claude/skills/%s-patterns/SKILL.md", strings.ToLower(repoName))
	instinctsPath := fmt.Sprintf(".claude/homunculus/instincts/inherited/%s-instincts.yaml", strings.ToLower(repoName))

	fmt.Println("\n")
	fmt.Println(bold(green("✨ Generation Complete!")))
	fmt.Println(gray(strings.Repeat("─", 50)))
	fmt.Printf(`
  %s %s
     %s

  %s %s
     %s

`, green("📄"), bold("Skill File:"), cyan(skillPath), green("🧠"), bold("Instincts File:"), cyan(instinctsPath))

	nextSteps := fmt.Sprintf(`
%s Review the generated SKILL.md
%s Import instincts: %s
%s View learned patterns: %s
%s Evolve into skills: %s
`, yellow("1."), yellow("2."), cyan("/instinct-import <path>"), yellow("3."), cyan("/instinct-status"), yellow("4."), cyan("/evolve"))
	fmt.Println(box("📋 Next Steps", nextSteps, 60))
	fmt.Println("\n")

	fmt.Println(gray(strings.Repeat("─", 60)))
	fmt.Println(dim("  Powered by Everything Claude Code • ecc.tools"))
	fmt.Println(dim("  GitHub App: github.com/apps/skill-creator"))
	fmt.Println("\n")
}

