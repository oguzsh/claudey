// Package hookio provides the stdin JSON protocol and stdout/stderr helpers
// used by Claude Code hooks.
package hookio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	DefaultMaxSize = 1024 * 1024 // 1MB
)

// ReadStdinJSON reads JSON from stdin with a size limit.
// Returns the parsed JSON as a map, the raw bytes, or an error.
// On timeout or parse failure, returns an empty map (matches JS behavior).
func ReadStdinJSON(ctx context.Context, maxSize int) (map[string]any, []byte, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}

	type result struct {
		data []byte
		err  error
	}

	ch := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, int64(maxSize)))
		ch <- result{data, err}
	}()

	select {
	case <-ctx.Done():
		return map[string]any{}, nil, nil
	case r := <-ch:
		if r.err != nil {
			return map[string]any{}, nil, nil
		}
		if len(r.data) == 0 {
			return map[string]any{}, r.data, nil
		}
		var parsed map[string]any
		if err := json.Unmarshal(r.data, &parsed); err != nil {
			return map[string]any{}, r.data, nil
		}
		return parsed, r.data, nil
	}
}

// Log writes a message to stderr (visible to user in Claude Code).
func Log(message string) {
	fmt.Fprintln(os.Stderr, message)
}

// Logf writes a formatted message to stderr.
func Logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Output writes data to stdout (returned to Claude).
// If v is a string, it is written directly.
// Otherwise, it is JSON-encoded.
func Output(v any) {
	switch val := v.(type) {
	case string:
		fmt.Println(val)
	case []byte:
		os.Stdout.Write(val)
		os.Stdout.Write([]byte("\n"))
	default:
		data, err := json.Marshal(val)
		if err != nil {
			fmt.Println("{}")
			return
		}
		os.Stdout.Write(data)
		os.Stdout.Write([]byte("\n"))
	}
}

// Passthrough writes raw bytes to stdout (for hooks that pass stdin through).
func Passthrough(data []byte) {
	os.Stdout.Write(data)
}

// GetToolInput extracts tool_input from parsed hook JSON.
func GetToolInput(input map[string]any) map[string]any {
	if ti, ok := input["tool_input"].(map[string]any); ok {
		return ti
	}
	return nil
}

// GetToolInputString extracts a string field from tool_input.
func GetToolInputString(input map[string]any, field string) string {
	ti := GetToolInput(input)
	if ti == nil {
		return ""
	}
	if s, ok := ti[field].(string); ok {
		return s
	}
	return ""
}

// GetToolOutput extracts tool_output from parsed hook JSON.
func GetToolOutput(input map[string]any) map[string]any {
	if to, ok := input["tool_output"].(map[string]any); ok {
		return to
	}
	return nil
}

// GetToolOutputString extracts a string field from tool_output.
func GetToolOutputString(input map[string]any, field string) string {
	to := GetToolOutput(input)
	if to == nil {
		return ""
	}
	if s, ok := to[field].(string); ok {
		return s
	}
	return ""
}




