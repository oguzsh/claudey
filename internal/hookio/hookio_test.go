package hookio

import (
	"testing"
)

func TestGetToolInputString(t *testing.T) {
	input := map[string]any{
		"tool_input": map[string]any{
			"file_path": "/src/main.ts",
			"command":   "npm run dev",
		},
	}

	if got := GetToolInputString(input, "file_path"); got != "/src/main.ts" {
		t.Errorf("GetToolInputString(file_path) = %q, want /src/main.ts", got)
	}
	if got := GetToolInputString(input, "command"); got != "npm run dev" {
		t.Errorf("GetToolInputString(command) = %q, want npm run dev", got)
	}
	if got := GetToolInputString(input, "missing"); got != "" {
		t.Errorf("GetToolInputString(missing) = %q, want empty", got)
	}
}

func TestGetToolInputString_NoToolInput(t *testing.T) {
	input := map[string]any{"other": "data"}
	if got := GetToolInputString(input, "file_path"); got != "" {
		t.Errorf("GetToolInputString on no tool_input = %q, want empty", got)
	}
}

func TestGetToolOutputString(t *testing.T) {
	input := map[string]any{
		"tool_output": map[string]any{
			"output": "https://github.com/owner/repo/pull/42",
		},
	}

	if got := GetToolOutputString(input, "output"); got != "https://github.com/owner/repo/pull/42" {
		t.Errorf("GetToolOutputString(output) = %q, want PR URL", got)
	}
}

func TestGetToolInput_NilMap(t *testing.T) {
	if got := GetToolInput(nil); got != nil {
		t.Errorf("GetToolInput(nil) = %v, want nil", got)
	}
}




