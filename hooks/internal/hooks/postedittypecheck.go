package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/oguzsh/everything-claude-code/internal/hookio"
	"github.com/oguzsh/everything-claude-code/internal/sysutil"
)

var tsExtRe = regexp.MustCompile(`\.(ts|tsx)$`)

// PostEditTypecheck runs TypeScript type checking after editing .ts/.tsx files.
func PostEditTypecheck(input map[string]any, raw []byte) {
	filePath := hookio.GetToolInputString(input, "file_path")

	if filePath != "" && tsExtRe.MatchString(filePath) {
		resolvedPath, err := filepath.Abs(filePath)
		if err != nil || !fileExists(resolvedPath) {
			hookio.Passthrough(raw)
			return
		}

		// Walk up to find tsconfig.json (max 20 levels)
		dir := filepath.Dir(resolvedPath)
		root := filepath.VolumeName(dir) + string(filepath.Separator)
		depth := 0

		for dir != root && depth < 20 {
			if fileExists(filepath.Join(dir, "tsconfig.json")) {
				break
			}
			dir = filepath.Dir(dir)
			depth++
		}

		if fileExists(filepath.Join(dir, "tsconfig.json")) {
			cmd := exec.Command(sysutil.NpxBin(), "tsc", "--noEmit", "--pretty", "false")
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()

			if err != nil {
				// tsc exits non-zero when there are errors
				relPath, _ := filepath.Rel(dir, resolvedPath)
				candidates := map[string]bool{
					filePath:     true,
					resolvedPath: true,
					relPath:      true,
				}

				lines := strings.Split(string(output), "\n")
				var relevantLines []string
				for _, line := range lines {
					for candidate := range candidates {
						if strings.Contains(line, candidate) {
							relevantLines = append(relevantLines, line)
							break
						}
					}
					if len(relevantLines) >= 10 {
						break
					}
				}

				if len(relevantLines) > 0 {
					hookio.Logf("[Hook] TypeScript errors in %s:", filepath.Base(filePath))
					for _, line := range relevantLines {
						hookio.Log(line)
					}
				}
			}
		}
	}

	hookio.Passthrough(raw)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

