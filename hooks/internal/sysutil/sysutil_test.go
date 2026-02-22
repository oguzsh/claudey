package sysutil

import (
	"runtime"
	"testing"
)

func TestCommandExists(t *testing.T) {
	// "ls" or "cmd" should always exist
	if runtime.GOOS == "windows" {
		if !CommandExists("cmd") {
			t.Error("CommandExists('cmd') = false on Windows")
		}
	} else {
		if !CommandExists("ls") {
			t.Error("CommandExists('ls') = false on Unix")
		}
	}

	if CommandExists("nonexistent_command_12345") {
		t.Error("CommandExists should return false for nonexistent command")
	}
}

func TestRunCommand(t *testing.T) {
	result := RunCommand("echo hello", "")
	if !result.Success {
		t.Error("RunCommand('echo hello') failed")
	}
	if result.Output != "hello" {
		t.Errorf("RunCommand output = %q, want 'hello'", result.Output)
	}
}

func TestRunCommand_Failure(t *testing.T) {
	result := RunCommand("false", "")
	if result.Success {
		t.Error("RunCommand('false') should fail")
	}
}

func TestNpxBin(t *testing.T) {
	bin := NpxBin()
	if runtime.GOOS == "windows" {
		if bin != "npx.cmd" {
			t.Errorf("NpxBin() = %q on Windows, want 'npx.cmd'", bin)
		}
	} else {
		if bin != "npx" {
			t.Errorf("NpxBin() = %q, want 'npx'", bin)
		}
	}
}

