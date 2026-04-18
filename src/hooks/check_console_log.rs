//! check-console-log hook — scan modified JS/TS files for console.log.
//!
//! Ports internal/hooks/checkconsolelog.go.

use crate::{fileutil, gitutil, hookio};
use regex::Regex;
use std::path::Path;

pub fn check_console_log(raw: Vec<u8>) {
    if !gitutil::is_git_repo() {
        hookio::passthrough(&raw);
        return;
    }

    let files = gitutil::modified_files(&[r"\.tsx?$", r"\.jsx?$"]);

    let excluded = excluded_patterns();

    let filtered: Vec<String> = files
        .into_iter()
        .filter(|f| Path::new(f).exists() && !is_excluded(f, &excluded))
        .collect();

    let mut has_console = false;
    for file in &filtered {
        if let Some(content) = fileutil::read_file(Path::new(file)) {
            if content.contains("console.log") {
                hookio::log(&format!("[Hook] WARNING: console.log found in {file}"));
                has_console = true;
            }
        }
    }

    if has_console {
        hookio::log("[Hook] Remove console.log statements before committing");
    }

    hookio::passthrough(&raw);
}

fn excluded_patterns() -> Vec<Regex> {
    [
        r"\.test\.[jt]sx?$",
        r"\.spec\.[jt]sx?$",
        r"\.config\.[jt]s$",
        r"scripts/",
        r"__tests__/",
        r"__mocks__/",
    ]
    .iter()
    .map(|p| Regex::new(p).expect("valid static regex"))
    .collect()
}

pub fn is_excluded(path: &str, patterns: &[Regex]) -> bool {
    patterns.iter().any(|p| p.is_match(path))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn excludes_test_files() {
        let p = excluded_patterns();
        assert!(is_excluded("foo.test.ts", &p));
        assert!(is_excluded("foo.test.tsx", &p));
        assert!(is_excluded("foo.test.js", &p));
    }

    #[test]
    fn excludes_spec_and_config() {
        let p = excluded_patterns();
        assert!(is_excluded("foo.spec.ts", &p));
        assert!(is_excluded("vite.config.ts", &p));
        assert!(is_excluded("jest.config.js", &p));
    }

    #[test]
    fn excludes_tests_and_mocks_dirs() {
        let p = excluded_patterns();
        assert!(is_excluded("src/__tests__/x.ts", &p));
        assert!(is_excluded("src/__mocks__/x.ts", &p));
        assert!(is_excluded("scripts/run.ts", &p));
    }

    #[test]
    fn does_not_exclude_source_files() {
        let p = excluded_patterns();
        assert!(!is_excluded("src/app.ts", &p));
        assert!(!is_excluded("lib/helper.js", &p));
        assert!(!is_excluded("component.tsx", &p));
    }
}
