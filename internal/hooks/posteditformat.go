package hooks

import (
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/oguzsh/claudey/internal/hookio"
	"github.com/oguzsh/claudey/internal/sysutil"
)

var jstsFormatRe = regexp.MustCompile(`\.(ts|tsx|js|jsx)$`)

// PostEditFormat auto-formats JS/TS files with Prettier after edits.
func PostEditFormat(input map[string]any, raw []byte) {
	filePath := hookio.GetToolInputString(input, "file_path")

	if filePath != "" && jstsFormatRe.MatchString(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err == nil {
			cmd := exec.Command(sysutil.NpxBin(), "prettier", "--write", filePath)
			cmd.Dir = filepath.Dir(absPath)
			// Run with timeout, ignore errors (prettier not installed, etc.)
			cmd.Run()
		}
	}

	hookio.Passthrough(raw)
}



