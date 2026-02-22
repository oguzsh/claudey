// Package sysutil provides system utilities for command execution.
package sysutil

import (
	"os/exec"
	"runtime"
	"strings"
)

// CommandExists checks if a command exists in PATH.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// CommandResult holds the result of running a command.
type CommandResult struct {
	Success bool
	Output  string
}

// RunCommand runs a shell command and returns its output.
// Only use with trusted, hardcoded commands.
func RunCommand(cmd string, dir string) CommandResult {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("sh", "-c", cmd)
	}

	if dir != "" {
		c.Dir = dir
	}

	output, err := c.CombinedOutput()
	if err != nil {
		return CommandResult{
			Success: false,
			Output:  strings.TrimSpace(string(output)),
		}
	}

	return CommandResult{
		Success: true,
		Output:  strings.TrimSpace(string(output)),
	}
}

// NpxBin returns "npx.cmd" on Windows, "npx" otherwise.
func NpxBin() string {
	if runtime.GOOS == "windows" {
		return "npx.cmd"
	}
	return "npx"
}




