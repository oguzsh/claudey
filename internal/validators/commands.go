package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var cmdRefRe = regexp.MustCompile("`/([a-z][-a-z0-9]*)`")
var agentPathRefRe = regexp.MustCompile(`agents/([a-z][-a-z0-9]*)\.md`)
var skillRefRe = regexp.MustCompile(`skills/([a-z][-a-z0-9]*)/`)
var workflowLineRe = regexp.MustCompile(`^([a-z][-a-z0-9]*(?:\s*->\s*[a-z][-a-z0-9]*)+)$`)
var createsLineRe = regexp.MustCompile(`(?i)creates:|would create:`)
var codeBlockRe = regexp.MustCompile("(?s)```.*?```")

// ValidateCommands validates command markdown files.
func ValidateCommands(rootDir string) error {
	commandsDir := filepath.Join(rootDir, "commands")
	agentsDir := filepath.Join(rootDir, "agents")
	skillsDir := filepath.Join(rootDir, "skills")

	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No commands directory found, skipping validation")
			return nil
		}
		return err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}

	// Build valid sets
	validCommands := make(map[string]bool)
	for _, f := range files {
		validCommands[strings.TrimSuffix(f, ".md")] = true
	}

	validAgents := make(map[string]bool)
	if agentEntries, err := os.ReadDir(agentsDir); err == nil {
		for _, e := range agentEntries {
			if strings.HasSuffix(e.Name(), ".md") {
				validAgents[strings.TrimSuffix(e.Name(), ".md")] = true
			}
		}
	}

	validSkills := make(map[string]bool)
	if skillEntries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range skillEntries {
			if e.IsDir() {
				validSkills[e.Name()] = true
			}
		}
	}

	hasErrors := false
	warnCount := 0

	for _, file := range files {
		filePath := filepath.Join(commandsDir, file)
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s - %s\n", file, err)
			hasErrors = true
			continue
		}

		content := string(data)
		if strings.TrimSpace(content) == "" {
			fmt.Fprintf(os.Stderr, "ERROR: %s - Empty command file\n", file)
			hasErrors = true
			continue
		}

		// Strip code blocks before cross-reference checking
		contentNoCode := codeBlockRe.ReplaceAllString(content, "")

		// Check command references line by line
		for _, line := range strings.Split(contentNoCode, "\n") {
			if createsLineRe.MatchString(line) {
				continue
			}
			for _, m := range cmdRefRe.FindAllStringSubmatch(line, -1) {
				if !validCommands[m[1]] {
					fmt.Fprintf(os.Stderr, "ERROR: %s - references non-existent command /%s\n", file, m[1])
					hasErrors = true
				}
			}
		}

		// Check agent path references
		for _, m := range agentPathRefRe.FindAllStringSubmatch(contentNoCode, -1) {
			if !validAgents[m[1]] {
				fmt.Fprintf(os.Stderr, "ERROR: %s - references non-existent agent agents/%s.md\n", file, m[1])
				hasErrors = true
			}
		}

		// Check skill references
		for _, m := range skillRefRe.FindAllStringSubmatch(contentNoCode, -1) {
			if !validSkills[m[1]] {
				fmt.Fprintf(os.Stderr, "WARN: %s - references skill directory skills/%s/ (not found locally)\n", file, m[1])
				warnCount++
			}
		}

		// Check workflow diagram references
		for _, line := range strings.Split(contentNoCode, "\n") {
			if m := workflowLineRe.FindStringSubmatch(line); m != nil {
				agents := strings.Split(m[1], "->")
				for _, agent := range agents {
					agent = strings.TrimSpace(agent)
					if !validAgents[agent] {
						fmt.Fprintf(os.Stderr, "ERROR: %s - workflow references non-existent agent \"%s\"\n", file, agent)
						hasErrors = true
					}
				}
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	msg := fmt.Sprintf("Validated %d command files", len(files))
	if warnCount > 0 {
		msg += fmt.Sprintf(" (%d warnings)", warnCount)
	}
	fmt.Println(msg)
	return nil
}




