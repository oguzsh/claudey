// Package validators provides CI validation for plugin components.
package validators

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var validEvents = map[string]bool{
	"PreToolUse": true, "PostToolUse": true, "PreCompact": true,
	"SessionStart": true, "SessionEnd": true, "Stop": true,
	"Notification": true, "SubagentStop": true,
}

// ValidateHooks validates hooks.json schema.
func ValidateHooks(rootDir string) error {
	hooksFile := filepath.Join(rootDir, "hooks", "hooks.json")

	data, err := os.ReadFile(hooksFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No hooks.json found, skipping validation")
			return nil
		}
		return fmt.Errorf("reading hooks.json: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid JSON in hooks.json: %w", err)
	}

	hooks, ok := raw["hooks"]
	if !ok {
		hooks = raw
	}

	hooksObj, ok := hooks.(map[string]any)
	if !ok {
		return fmt.Errorf("hooks.json must be an object")
	}

	hasErrors := false
	totalMatchers := 0

	for eventType, matchersRaw := range hooksObj {
		if eventType == "$schema" {
			continue
		}
		if !validEvents[eventType] {
			fmt.Fprintf(os.Stderr, "ERROR: Invalid event type: %s\n", eventType)
			hasErrors = true
			continue
		}

		matchers, ok := matchersRaw.([]any)
		if !ok {
			fmt.Fprintf(os.Stderr, "ERROR: %s must be an array\n", eventType)
			hasErrors = true
			continue
		}

		for i, matcherRaw := range matchers {
			matcher, ok := matcherRaw.(map[string]any)
			if !ok {
				fmt.Fprintf(os.Stderr, "ERROR: %s[%d] is not an object\n", eventType, i)
				hasErrors = true
				continue
			}

			if _, ok := matcher["matcher"]; !ok {
				fmt.Fprintf(os.Stderr, "ERROR: %s[%d] missing 'matcher' field\n", eventType, i)
				hasErrors = true
			}

			hooksArr, ok := matcher["hooks"].([]any)
			if !ok {
				fmt.Fprintf(os.Stderr, "ERROR: %s[%d] missing 'hooks' array\n", eventType, i)
				hasErrors = true
			} else {
				for j, hookRaw := range hooksArr {
					label := fmt.Sprintf("%s[%d].hooks[%d]", eventType, i, j)
					if validateHookEntry(hookRaw, label) {
						hasErrors = true
					}
				}
			}
			totalMatchers++
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("Validated %d hook matchers\n", totalMatchers)
	return nil
}

func validateHookEntry(hookRaw any, label string) bool {
	hook, ok := hookRaw.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: %s is not an object\n", label)
		return true
	}

	hasErrors := false

	hookType, _ := hook["type"].(string)
	if hookType == "" {
		fmt.Fprintf(os.Stderr, "ERROR: %s missing or invalid 'type' field\n", label)
		hasErrors = true
	}

	// Validate async field
	if v, exists := hook["async"]; exists {
		if _, ok := v.(bool); !ok {
			fmt.Fprintf(os.Stderr, "ERROR: %s 'async' must be a boolean\n", label)
			hasErrors = true
		}
	}

	// Validate timeout field
	if v, exists := hook["timeout"]; exists {
		if num, ok := v.(float64); !ok || num < 0 {
			fmt.Fprintf(os.Stderr, "ERROR: %s 'timeout' must be a non-negative number\n", label)
			hasErrors = true
		}
	}

	// Validate command field
	cmd := hook["command"]
	switch c := cmd.(type) {
	case string:
		if len(c) == 0 || len(trimStr(c)) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: %s missing or invalid 'command' field\n", label)
			hasErrors = true
		}
	case []any:
		if len(c) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: %s missing or invalid 'command' field\n", label)
			hasErrors = true
		} else {
			for _, s := range c {
				str, ok := s.(string)
				if !ok || str == "" {
					fmt.Fprintf(os.Stderr, "ERROR: %s missing or invalid 'command' field\n", label)
					hasErrors = true
					break
				}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "ERROR: %s missing or invalid 'command' field\n", label)
		hasErrors = true
	}

	return hasErrors
}

func trimStr(s string) string {
	result := ""
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result += string(c)
		}
	}
	return result
}




