package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/oguzsh/claudey/hooks/internal/fileutil"
	"github.com/oguzsh/claudey/hooks/internal/hookio"
)

var jstsExtRe = regexp.MustCompile(`\.(ts|tsx|js|jsx)$`)
var consoleLogRe = regexp.MustCompile(`console\.log`)

// PostEditConsoleWarn warns about console.log statements after edits.
func PostEditConsoleWarn(input map[string]any, raw []byte) {
	filePath := hookio.GetToolInputString(input, "file_path")

	if filePath != "" && jstsExtRe.MatchString(filePath) {
		content, ok := fileutil.ReadFile(filePath)
		if ok {
			lines := strings.Split(content, "\n")
			var matches []string

			for idx, line := range lines {
				if consoleLogRe.MatchString(line) {
					matches = append(matches, fmt.Sprintf("%d: %s", idx+1, strings.TrimSpace(line)))
				}
			}

			if len(matches) > 0 {
				hookio.Logf("[Hook] WARNING: console.log found in %s", filePath)
				limit := 5
				if len(matches) < limit {
					limit = len(matches)
				}
				for _, m := range matches[:limit] {
					hookio.Log(m)
				}
				hookio.Log("[Hook] Remove console.log before committing")
			}
		}
	}

	hookio.Passthrough(raw)
}
