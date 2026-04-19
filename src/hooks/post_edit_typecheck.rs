//! post-edit-typecheck hook — run tsc after editing .ts/.tsx files.
//!
//! Ports internal/hooks/postedittypecheck.go.

use crate::{hookio, sysutil};
use serde_json::Value;
use std::path::{Path, PathBuf};
use std::process::Command;

const MAX_WALK_DEPTH: usize = 20;
const MAX_RELEVANT_LINES: usize = 10;

pub fn post_edit_typecheck(input: Value, raw: Vec<u8>) {
    let file_path = hookio::get_tool_input_string(&input, "file_path");

    if file_path.is_empty() || !is_typescript(file_path) {
        hookio::passthrough(&raw);
        return;
    }

    let resolved = match std::fs::canonicalize(file_path) {
        Ok(p) => p,
        Err(_) => {
            hookio::passthrough(&raw);
            return;
        }
    };

    if let Some(tsconfig_dir) = find_tsconfig_dir(&resolved) {
        let output = Command::new(sysutil::npx_bin())
            .args(["tsc", "--noEmit", "--pretty", "false"])
            .current_dir(&tsconfig_dir)
            .output();

        if let Ok(out) = output {
            if !out.status.success() {
                let mut combined = String::from_utf8_lossy(&out.stdout).into_owned();
                combined.push_str(&String::from_utf8_lossy(&out.stderr));

                let rel = resolved
                    .strip_prefix(&tsconfig_dir)
                    .ok()
                    .map(|p| p.to_string_lossy().into_owned());
                let candidates: Vec<String> = [
                    Some(file_path.to_string()),
                    Some(resolved.to_string_lossy().into_owned()),
                    rel,
                ]
                .into_iter()
                .flatten()
                .filter(|s| !s.is_empty())
                .collect();

                let relevant: Vec<&str> = combined
                    .lines()
                    .filter(|line| candidates.iter().any(|c| line.contains(c)))
                    .take(MAX_RELEVANT_LINES)
                    .collect();

                if !relevant.is_empty() {
                    let base = Path::new(file_path)
                        .file_name()
                        .map(|s| s.to_string_lossy().into_owned())
                        .unwrap_or_else(|| file_path.to_string());
                    hookio::log(&format!("[Hook] TypeScript errors in {base}:"));
                    for line in relevant {
                        hookio::log(line);
                    }
                }
            }
        }
    }

    hookio::passthrough(&raw);
}

pub fn is_typescript(path: &str) -> bool {
    path.ends_with(".ts") || path.ends_with(".tsx")
}

pub fn find_tsconfig_dir(start: &Path) -> Option<PathBuf> {
    let mut dir = start.parent()?.to_path_buf();
    for _ in 0..MAX_WALK_DEPTH {
        if dir.join("tsconfig.json").is_file() {
            return Some(dir);
        }
        match dir.parent() {
            Some(p) if p != dir => dir = p.to_path_buf(),
            _ => return None,
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn is_typescript_matches_ts_variants() {
        assert!(is_typescript("foo.ts"));
        assert!(is_typescript("foo.tsx"));
        assert!(is_typescript("/abs/path.ts"));
    }

    #[test]
    fn is_typescript_rejects_non_ts() {
        assert!(!is_typescript("foo.js"));
        assert!(!is_typescript("foo.py"));
        assert!(!is_typescript("foo"));
        assert!(!is_typescript(""));
    }

    #[test]
    fn find_tsconfig_walks_up() {
        let base = crate::testutil::TempDir::new();
        let nested = base.path().join("a").join("b").join("c");
        std::fs::create_dir_all(&nested).unwrap();
        std::fs::write(base.path().join("tsconfig.json"), "{}").unwrap();

        let file = nested.join("file.ts");
        std::fs::write(&file, "").unwrap();

        let found = find_tsconfig_dir(&file).expect("tsconfig should be found");
        assert_eq!(
            std::fs::canonicalize(found).unwrap(),
            std::fs::canonicalize(base.path()).unwrap()
        );
    }

    #[test]
    fn find_tsconfig_returns_none_when_absent() {
        let base = crate::testutil::TempDir::new();
        let file = base.path().join("file.ts");
        std::fs::write(&file, "").unwrap();
        assert!(find_tsconfig_dir(&file).is_none());
    }
}
