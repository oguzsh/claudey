package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateSkills validates skill directories have SKILL.md with content.
func ValidateSkills(rootDir string) error {
	skillsDir := filepath.Join(rootDir, "skills")

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No skills directory found, skipping validation")
			return nil
		}
		return err
	}

	hasErrors := false
	validCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillMd := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillMd)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "ERROR: %s/ - Missing SKILL.md\n", entry.Name())
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: %s/SKILL.md - %s\n", entry.Name(), err)
			}
			hasErrors = true
			continue
		}

		if strings.TrimSpace(string(data)) == "" {
			fmt.Fprintf(os.Stderr, "ERROR: %s/SKILL.md - Empty file\n", entry.Name())
			hasErrors = true
			continue
		}

		validCount++
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("Validated %d skill directories\n", validCount)
	return nil
}




