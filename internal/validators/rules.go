package validators

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ValidateRules validates rule markdown files.
func ValidateRules(rootDir string) error {
	rulesDir := filepath.Join(rootDir, "rules")

	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		fmt.Println("No rules directory found, skipping validation")
		return nil
	}

	hasErrors := false
	validatedCount := 0

	filepath.WalkDir(rulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(rulesDir, path)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s - %s\n", relPath, err)
			hasErrors = true
			return nil
		}

		if strings.TrimSpace(string(data)) == "" {
			fmt.Fprintf(os.Stderr, "ERROR: %s - Empty rule file\n", relPath)
			hasErrors = true
			return nil
		}

		validatedCount++
		return nil
	})

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("Validated %d rule files\n", validatedCount)
	return nil
}




