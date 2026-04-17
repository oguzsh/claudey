package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oguzsh/claudey/internal/datetime"
	"github.com/oguzsh/claudey/internal/fileutil"
	"github.com/oguzsh/claudey/internal/gitutil"
	"github.com/oguzsh/claudey/internal/hookio"
	"github.com/oguzsh/claudey/internal/platform"
)

// SessionSummary holds extracted session data.
type SessionSummary struct {
	UserMessages  []string
	ToolsUsed     []string
	FilesModified []string
	TotalMessages int
}

// extractSessionSummary reads a JSONL transcript and extracts key information.
func extractSessionSummary(transcriptPath string) *SessionSummary {
	content, ok := fileutil.ReadFile(transcriptPath)
	if !ok {
		return nil
	}

	lines := strings.Split(content, "\n")
	var userMessages []string
	toolsUsed := make(map[string]bool)
	filesModified := make(map[string]bool)
	parseErrors := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			parseErrors++
			continue
		}

		// Collect user messages
		entryType, _ := entry["type"].(string)
		role, _ := entry["role"].(string)
		var msgRole string
		if msg, ok := entry["message"].(map[string]any); ok {
			msgRole, _ = msg["role"].(string)
		}

		if entryType == "user" || role == "user" || msgRole == "user" {
			text := extractText(entry)
			if text != "" {
				if len(text) > 200 {
					text = text[:200]
				}
				userMessages = append(userMessages, text)
			}
		}

		// Collect tool names and modified files (direct entries)
		if entryType == "tool_use" || entry["tool_name"] != nil {
			toolName := strOr(entry["tool_name"], entry["name"])
			if toolName != "" {
				toolsUsed[toolName] = true
			}
			filePath := extractFilePath(entry, "tool_input")
			if filePath == "" {
				filePath = extractFilePath(entry, "input")
			}
			if filePath != "" && (toolName == "Edit" || toolName == "Write") {
				filesModified[filePath] = true
			}
		}

		// Extract tool uses from assistant content blocks
		if entryType == "assistant" {
			if msg, ok := entry["message"].(map[string]any); ok {
				if content, ok := msg["content"].([]any); ok {
					for _, block := range content {
						if blockMap, ok := block.(map[string]any); ok {
							if blockType, _ := blockMap["type"].(string); blockType == "tool_use" {
								toolName, _ := blockMap["name"].(string)
								if toolName != "" {
									toolsUsed[toolName] = true
								}
								if inputMap, ok := blockMap["input"].(map[string]any); ok {
									if fp, ok := inputMap["file_path"].(string); ok && fp != "" {
										if toolName == "Edit" || toolName == "Write" {
											filesModified[fp] = true
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if parseErrors > 0 {
		hookio.Logf("[SessionEnd] Skipped %d/%d unparseable transcript lines", parseErrors, len(lines))
	}

	if len(userMessages) == 0 {
		return nil
	}

	// Last 10 user messages
	start := 0
	if len(userMessages) > 10 {
		start = len(userMessages) - 10
	}

	tools := mapKeys(toolsUsed, 20)
	files := mapKeys(filesModified, 30)

	return &SessionSummary{
		UserMessages:  userMessages[start:],
		ToolsUsed:     tools,
		FilesModified: files,
		TotalMessages: len(userMessages),
	}
}

func extractText(entry map[string]any) string {
	// Try message.content first
	if msg, ok := entry["message"].(map[string]any); ok {
		return contentToString(msg["content"])
	}
	return contentToString(entry["content"])
}

func contentToString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	if arr, ok := v.([]any); ok {
		var parts []string
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.TrimSpace(strings.Join(parts, " "))
	}
	return ""
}

func extractFilePath(entry map[string]any, key string) string {
	if input, ok := entry[key].(map[string]any); ok {
		if fp, ok := input["file_path"].(string); ok {
			return fp
		}
	}
	return ""
}

func strOr(vals ...any) string {
	for _, v := range vals {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func mapKeys(m map[string]bool, limit int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
		if len(keys) >= limit {
			break
		}
	}
	return keys
}

func buildSummarySection(summary *SessionSummary) string {
	var b strings.Builder

	b.WriteString("## Session Summary\n\n")

	b.WriteString("### Tasks\n")
	for _, msg := range summary.UserMessages {
		escaped := strings.ReplaceAll(msg, "\n", " ")
		escaped = strings.ReplaceAll(escaped, "`", "\\`")
		b.WriteString("- " + escaped + "\n")
	}
	b.WriteString("\n")

	if len(summary.FilesModified) > 0 {
		b.WriteString("### Files Modified\n")
		for _, f := range summary.FilesModified {
			b.WriteString("- " + f + "\n")
		}
		b.WriteString("\n")
	}

	if len(summary.ToolsUsed) > 0 {
		b.WriteString("### Tools Used\n")
		b.WriteString(strings.Join(summary.ToolsUsed, ", ") + "\n\n")
	}

	b.WriteString(fmt.Sprintf("### Stats\n- Total user messages: %d\n", summary.TotalMessages))

	return b.String()
}

// SessionEnd persists session state when session ends.
func SessionEnd(input map[string]any) {
	// Get transcript path from input or env
	transcriptPath, _ := input["transcript_path"].(string)
	if transcriptPath == "" {
		transcriptPath = os.Getenv("CLAUDE_TRANSCRIPT_PATH")
	}

	sessionsDir := platform.SessionsDir()
	today := datetime.DateString()
	shortID := sessionIDShort()
	sessionFile := filepath.Join(sessionsDir, today+"-"+shortID+"-session.tmp")

	platform.EnsureDir(sessionsDir)

	currentTime := datetime.TimeString()

	// Try to extract summary from transcript
	var summary *SessionSummary
	if transcriptPath != "" {
		if fileutil.Exists(transcriptPath) {
			summary = extractSessionSummary(transcriptPath)
		} else {
			hookio.Logf("[SessionEnd] Transcript not found: %s", transcriptPath)
		}
	}

	if fileutil.Exists(sessionFile) {
		// Update existing session file
		fileutil.ReplaceRegexInFile(sessionFile, `\*\*Last Updated:\*\*.*`, "**Last Updated:** "+currentTime)

		// Replace blank template with summary if available
		if summary != nil {
			existing, ok := fileutil.ReadFile(sessionFile)
			if ok && strings.Contains(existing, "[Session context goes here]") {
				templateRe := `## Current State\s*\n\s*\[Session context goes here\][\s\S]*?### Context to Load\s*\n` + "```\\s*\n\\[relevant files\\]\\s*\n```"
				fileutil.ReplaceRegexInFile(sessionFile, templateRe, buildSummarySection(summary))
			}
		}

		hookio.Logf("[SessionEnd] Updated session file: %s", sessionFile)
	} else {
		// Create new session file
		summarySection := ""
		if summary != nil {
			summarySection = buildSummarySection(summary)
		} else {
			summarySection = "## Current State\n\n[Session context goes here]\n\n### Completed\n- [ ]\n\n### In Progress\n- [ ]\n\n### Notes for Next Session\n-\n\n### Context to Load\n```\n[relevant files]\n```"
		}

		template := fmt.Sprintf("# Session: %s\n**Date:** %s\n**Started:** %s\n**Last Updated:** %s\n\n---\n\n%s\n",
			today, today, currentTime, currentTime, summarySection)

		fileutil.WriteFile(sessionFile, template)
		hookio.Logf("[SessionEnd] Created session file: %s", sessionFile)
	}
}

// sessionIDShort returns a short session identifier.
func sessionIDShort() string {
	sessionID := os.Getenv("CLAUDE_SESSION_ID")
	if sessionID != "" && len(sessionID) > 0 {
		if len(sessionID) > 8 {
			return sessionID[len(sessionID)-8:]
		}
		return sessionID
	}
	if name := gitutil.ProjectName(); name != "" {
		return name
	}
	return "default"
}
