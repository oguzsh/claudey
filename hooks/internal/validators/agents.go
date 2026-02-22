package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var requiredAgentFields = []string{"model", "tools"}
var validModels = map[string]bool{"haiku": true, "sonnet": true, "opus": true}

// ValidateAgents validates agent markdown files have required frontmatter.
func ValidateAgents(rootDir string) error {
	agentsDir := filepath.Join(rootDir, "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No agents directory found, skipping validation")
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

	hasErrors := false

	for _, file := range files {
		filePath := filepath.Join(agentsDir, file)
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s - %s\n", file, err)
			hasErrors = true
			continue
		}

		content := string(data)
		frontmatter := extractFrontmatter(content)

		if frontmatter == nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s - Missing frontmatter\n", file)
			hasErrors = true
			continue
		}

		for _, field := range requiredAgentFields {
			val, ok := frontmatter[field]
			if !ok || strings.TrimSpace(val) == "" {
				fmt.Fprintf(os.Stderr, "ERROR: %s - Missing required field: %s\n", file, field)
				hasErrors = true
			}
		}

		if model, ok := frontmatter["model"]; ok && !validModels[model] {
			fmt.Fprintf(os.Stderr, "ERROR: %s - Invalid model '%s'. Must be one of: haiku, sonnet, opus\n", file, model)
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("Validated %d agent files\n", len(files))
	return nil
}

// extractFrontmatter parses YAML-like frontmatter from markdown.
func extractFrontmatter(content string) map[string]string {
	// Strip BOM
	content = strings.TrimPrefix(content, "\uFEFF")

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")

	if !strings.HasPrefix(content, "---\n") {
		return nil
	}

	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return nil
	}

	fm := make(map[string]string)
	lines := strings.Split(content[4:4+end], "\n")
	for _, line := range lines {
		idx := strings.Index(line, ":")
		if idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			fm[key] = value
		}
	}

	return fm
}

