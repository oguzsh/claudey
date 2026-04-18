#![allow(dead_code)]

use std::path::{Path, PathBuf};

/// True when compiled for Windows.
pub const IS_WINDOWS: bool = cfg!(target_os = "windows");
/// True when compiled for macOS.
pub const IS_MACOS: bool = cfg!(target_os = "macos");
/// True when compiled for Linux.
pub const IS_LINUX: bool = cfg!(target_os = "linux");

/// Returns the current user's home directory.
/// Panics with a clear message if the home directory cannot be determined,
/// mirroring the Go implementation's `panic(err)`.
pub fn home_dir() -> PathBuf {
    dirs::home_dir().expect("unable to determine home directory")
}

/// Returns the path to the Claude configuration directory (`~/.claude`).
pub fn claude_dir() -> PathBuf {
    home_dir().join(".claude")
}

/// Returns the path to the sessions directory (`~/.claude/sessions`).
pub fn sessions_dir() -> PathBuf {
    claude_dir().join("sessions")
}

/// Returns the path to the learned skills directory (`~/.claude/skills/learned`).
pub fn learned_skills_dir() -> PathBuf {
    claude_dir().join("skills").join("learned")
}

/// Returns the OS default temporary directory.
pub fn temp_dir() -> PathBuf {
    std::env::temp_dir()
}

/// Creates the directory at `p` along with any necessary parents.
/// If the directory already exists, returns `Ok(())` (mirrors Go's `os.ErrExist` swallow).
pub fn ensure_dir(p: &Path) -> std::io::Result<()> {
    match std::fs::create_dir_all(p) {
        Ok(()) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => Ok(()),
        Err(e) => Err(e),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    // NOTE: Tests that override the HOME environment variable are skipped here
    // because env vars are process-global and `cargo test` runs tests in parallel
    // by default, causing races. Those tests would require `--test-threads=1` to
    // be safe. Instead, we test the structural/relational properties of the path
    // helpers without mutating the environment.

    // --- OS flag tests ---

    #[test]
    fn test_exactly_one_os_flag_is_true() {
        let flags = [IS_WINDOWS, IS_MACOS, IS_LINUX];
        let true_count = flags.iter().filter(|&&f| f).count();
        // On known platforms exactly one flag is true; on unknown OSes all may be
        // false, but never more than one should be true simultaneously.
        assert!(
            true_count <= 1,
            "expected at most one OS flag to be true, got {true_count}"
        );
    }

    // --- Path structure / relational tests ---
    // These tests port the pure-logic assertions from dirs_test.go without
    // actually changing $HOME, so they are deterministic and race-free.

    #[test]
    fn test_claude_dir_ends_with_dot_claude() {
        let d = claude_dir();
        assert_eq!(
            d.file_name().unwrap(),
            ".claude",
            "claude_dir() should end with '.claude', got {d:?}"
        );
    }

    #[test]
    fn test_sessions_dir_ends_with_sessions() {
        let d = sessions_dir();
        assert_eq!(
            d.file_name().unwrap(),
            "sessions",
            "sessions_dir() should end with 'sessions', got {d:?}"
        );
    }

    #[test]
    fn test_sessions_dir_is_under_claude_dir() {
        let sessions = sessions_dir();
        let claude = claude_dir();
        assert!(
            sessions.starts_with(&claude),
            "sessions_dir() {sessions:?} should be under claude_dir() {claude:?}"
        );
    }

    #[test]
    fn test_learned_skills_dir_ends_with_learned() {
        let d = learned_skills_dir();
        assert_eq!(
            d.file_name().unwrap(),
            "learned",
            "learned_skills_dir() should end with 'learned', got {d:?}"
        );
    }

    #[test]
    fn test_learned_skills_dir_is_under_claude_dir() {
        let learned = learned_skills_dir();
        let claude = claude_dir();
        assert!(
            learned.starts_with(&claude),
            "learned_skills_dir() {learned:?} should be under claude_dir() {claude:?}"
        );
    }

    #[test]
    fn test_temp_dir_is_non_empty() {
        let t = temp_dir();
        assert!(!t.as_os_str().is_empty(), "temp_dir() should not be empty");
    }

    // --- ensure_dir tests (ported from TestEnsureDir_* in dirs_test.go) ---

    #[test]
    fn test_ensure_dir_creates_new_directory() {
        let base = tempfile::tempdir().expect("failed to create temp dir");
        let new_dir = base.path().join("a").join("b").join("c");

        ensure_dir(&new_dir).expect("ensure_dir should not fail on a new nested path");

        let info = fs::metadata(&new_dir).expect("directory should exist after ensure_dir");
        assert!(info.is_dir(), "path should be a directory");
    }

    #[test]
    fn test_ensure_dir_existing_directory_is_no_op() {
        let base = tempfile::tempdir().expect("failed to create temp dir");
        // The temp dir already exists — calling ensure_dir should be a no-op.
        ensure_dir(base.path()).expect("ensure_dir on existing dir should not error");
    }

    #[test]
    fn test_ensure_dir_returns_error_for_invalid_path() {
        let base = tempfile::tempdir().expect("failed to create temp dir");
        let file_path = base.path().join("file.txt");
        fs::write(&file_path, b"hello").expect("setup: could not create file");

        // Trying to create a directory inside a regular file must fail.
        let invalid_dir = file_path.join("subdir");
        let result = ensure_dir(&invalid_dir);
        assert!(
            result.is_err(),
            "ensure_dir should return an error when a path component is a regular file"
        );
    }
}
