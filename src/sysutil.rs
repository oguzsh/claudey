//! sysutil — command lookup + shell execution helpers.
//!
//! Ports internal/sysutil/sysutil.go. Uses only `std` (no `which` crate).

#![allow(dead_code)]

use std::path::Path;
use std::process::Command;

/// Result of running a shell command. Output is the trimmed combined
/// stdout + stderr.
pub struct CommandResult {
    pub success: bool,
    pub output: String,
}

/// Returns true when `cmd` is found in `PATH` and is executable.
/// Accepts an absolute or relative path directly (bypasses `PATH`).
pub fn command_exists(cmd: &str) -> bool {
    if cmd.is_empty() {
        return false;
    }
    let has_sep = cmd.contains('/') || (cfg!(windows) && cmd.contains('\\'));
    if has_sep {
        return is_executable(Path::new(cmd));
    }
    let path = match std::env::var_os("PATH") {
        Some(p) => p,
        None => return false,
    };
    let sep = if cfg!(windows) { ';' } else { ':' };
    for part in path.to_string_lossy().split(sep) {
        if part.is_empty() {
            continue;
        }
        let candidate = Path::new(part).join(cmd);
        if is_executable(&candidate) {
            return true;
        }
    }
    false
}

#[cfg(unix)]
fn is_executable(p: &Path) -> bool {
    use std::os::unix::fs::PermissionsExt;
    match std::fs::metadata(p) {
        Ok(m) => m.is_file() && (m.permissions().mode() & 0o111 != 0),
        Err(_) => false,
    }
}

#[cfg(windows)]
fn is_executable(p: &Path) -> bool {
    if p.is_file() {
        return true;
    }
    // Try each PATHEXT suffix — matches what Go's exec.LookPath does.
    if let Some(pathext) = std::env::var_os("PATHEXT") {
        for ext in pathext.to_string_lossy().split(';') {
            let mut with_ext = p.as_os_str().to_owned();
            with_ext.push(ext);
            if Path::new(&with_ext).is_file() {
                return true;
            }
        }
    }
    false
}

#[cfg(not(any(unix, windows)))]
fn is_executable(p: &Path) -> bool {
    p.is_file()
}

/// Runs a shell command and returns the combined stdout+stderr output (trimmed)
/// plus the success flag. On Windows uses `cmd /C`, otherwise `sh -c`.
///
/// Only call with trusted, hard-coded command strings — the argument is passed
/// verbatim to the shell.
pub fn run_command(cmd: &str, dir: Option<&Path>) -> CommandResult {
    let mut c = if cfg!(target_os = "windows") {
        let mut x = Command::new("cmd");
        x.arg("/C").arg(cmd);
        x
    } else {
        let mut x = Command::new("sh");
        x.arg("-c").arg(cmd);
        x
    };
    if let Some(d) = dir {
        c.current_dir(d);
    }
    match c.output() {
        Ok(o) => {
            let mut s = String::from_utf8_lossy(&o.stdout).into_owned();
            s.push_str(&String::from_utf8_lossy(&o.stderr));
            CommandResult {
                success: o.status.success(),
                output: s.trim().to_string(),
            }
        }
        Err(_) => CommandResult {
            success: false,
            output: String::new(),
        },
    }
}

/// Returns the platform-specific `npx` launcher name.
pub fn npx_bin() -> &'static str {
    if cfg!(target_os = "windows") {
        "npx.cmd"
    } else {
        "npx"
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_command_exists_known_command() {
        // `sh` is present on every POSIX system and in Windows' Git Bash / WSL.
        #[cfg(unix)]
        assert!(
            command_exists("sh"),
            "command_exists('sh') should be true on unix"
        );
        #[cfg(windows)]
        assert!(
            command_exists("cmd"),
            "command_exists('cmd') should be true on Windows"
        );
    }

    #[test]
    fn test_command_exists_unknown_command() {
        assert!(
            !command_exists("nonexistent_command_12345"),
            "command_exists should be false for a nonexistent command"
        );
    }

    #[test]
    fn test_command_exists_empty_string() {
        assert!(!command_exists(""), "empty string must not match");
    }

    #[test]
    fn test_run_command_echo() {
        let r = run_command("echo hello", None);
        assert!(r.success, "echo hello should succeed");
        assert_eq!(r.output, "hello");
    }

    #[test]
    fn test_run_command_failure() {
        // `false` always exits non-zero on Unix. On Windows, `exit 1` under cmd.
        #[cfg(unix)]
        let r = run_command("false", None);
        #[cfg(windows)]
        let r = run_command("exit 1", None);
        assert!(!r.success, "failing command should report success=false");
    }

    #[test]
    fn test_run_command_cwd_is_respected() {
        let base = crate::testutil::TempDir::new();
        let expected = base.path().canonicalize().unwrap_or_else(|_| base.path().to_path_buf());
        #[cfg(unix)]
        let r = run_command("pwd", Some(base.path()));
        #[cfg(windows)]
        let r = run_command("cd", Some(base.path()));
        assert!(r.success);
        let got = std::path::Path::new(r.output.trim())
            .canonicalize()
            .unwrap_or_else(|_| std::path::PathBuf::from(r.output.trim()));
        assert_eq!(got, expected, "cwd override should change the shell's pwd");
    }

    #[test]
    fn test_npx_bin_per_platform() {
        #[cfg(target_os = "windows")]
        assert_eq!(npx_bin(), "npx.cmd");
        #[cfg(not(target_os = "windows"))]
        assert_eq!(npx_bin(), "npx");
    }
}
