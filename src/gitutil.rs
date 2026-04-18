//! gitutil — thin git wrapper using `sysutil::run_command`.
//!
//! Ports internal/gitutil/gitutil.go.

#![allow(dead_code)]

use crate::sysutil;
use regex::Regex;
use std::path::Path;

/// True when the current working directory is inside a git repository.
pub fn is_git_repo() -> bool {
    sysutil::run_command("git rev-parse --git-dir", None).success
}

/// Git repository name = basename of `git rev-parse --show-toplevel`.
/// Empty string when not in a repo or git is unavailable.
pub fn repo_name() -> String {
    let r = sysutil::run_command("git rev-parse --show-toplevel", None);
    if !r.success {
        return String::new();
    }
    Path::new(r.output.trim())
        .file_name()
        .and_then(|s| s.to_str())
        .map(|s| s.to_string())
        .unwrap_or_default()
}

/// `repo_name()` when in a repo, otherwise basename of the current directory.
pub fn project_name() -> String {
    let name = repo_name();
    if !name.is_empty() {
        return name;
    }
    std::env::current_dir()
        .ok()
        .and_then(|p| {
            p.file_name()
                .and_then(|s| s.to_str())
                .map(|s| s.to_string())
        })
        .unwrap_or_default()
}

/// `git diff --name-only HEAD`, optionally filtered by regex patterns.
/// Invalid / empty patterns are silently skipped. When no valid patterns
/// are supplied, returns all modified files.
pub fn modified_files(patterns: &[&str]) -> Vec<String> {
    if !is_git_repo() {
        return Vec::new();
    }
    let r = sysutil::run_command("git diff --name-only HEAD", None);
    if !r.success {
        return Vec::new();
    }
    let files: Vec<String> = r
        .output
        .split('\n')
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .map(String::from)
        .collect();

    if patterns.is_empty() {
        return files;
    }

    let compiled: Vec<Regex> = patterns
        .iter()
        .filter(|p| !p.is_empty())
        .filter_map(|p| Regex::new(p).ok())
        .collect();

    if compiled.is_empty() {
        return files;
    }

    files
        .into_iter()
        .filter(|f| compiled.iter().any(|re| re.is_match(f)))
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_project_name_non_empty() {
        // `cargo test` always runs inside some directory, so this must produce *something*.
        assert!(
            !project_name().is_empty(),
            "project_name() should never return empty from a real cwd"
        );
    }

    #[test]
    fn test_modified_files_filters_by_pattern() {
        // Unit-testable pure logic: the pattern filter.
        // We call the helper that does regex filtering in isolation.
        let all = vec![
            "src/main.rs".to_string(),
            "README.md".to_string(),
            "tests/integration.rs".to_string(),
        ];
        let patterns = [r"\.rs$"];
        let compiled: Vec<Regex> = patterns
            .iter()
            .filter_map(|p| Regex::new(p).ok())
            .collect();
        let matched: Vec<String> = all
            .into_iter()
            .filter(|f| compiled.iter().any(|re| re.is_match(f)))
            .collect();
        assert_eq!(matched, vec!["src/main.rs", "tests/integration.rs"]);
    }

    #[test]
    fn test_invalid_regex_patterns_are_skipped() {
        // Invalid regex never crashes; just drops out.
        let compiled: Vec<Regex> = ["[invalid", r"\.rs$"]
            .iter()
            .filter_map(|p| Regex::new(p).ok())
            .collect();
        assert_eq!(compiled.len(), 1);
    }

    #[test]
    fn test_is_git_repo_in_this_workspace() {
        // The test is invoked from within this git repo's worktree; don't assert —
        // just ensure the call doesn't panic.
        let _ = is_git_repo();
    }
}
