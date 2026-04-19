//! post-edit-format hook — auto-format JS/TS files with Prettier.
//!
//! Ports internal/hooks/posteditformat.go.

use crate::{hookio, sysutil};
use serde_json::Value;
use std::path::Path;
use std::process::Command;

pub fn post_edit_format(input: Value, raw: Vec<u8>) {
    let file_path = hookio::get_tool_input_string(&input, "file_path");

    if !file_path.is_empty() && should_format(file_path) {
        let path = Path::new(file_path);
        if let Ok(abs) = std::fs::canonicalize(path).or_else(|_| absolute(path)) {
            if let Some(dir) = abs.parent() {
                let _ = Command::new(sysutil::npx_bin())
                    .args(["prettier", "--write", file_path])
                    .current_dir(dir)
                    .output();
            }
        }
    }

    hookio::passthrough(&raw);
}

pub fn should_format(path: &str) -> bool {
    path.ends_with(".ts")
        || path.ends_with(".tsx")
        || path.ends_with(".js")
        || path.ends_with(".jsx")
}

fn absolute(p: &Path) -> std::io::Result<std::path::PathBuf> {
    if p.is_absolute() {
        Ok(p.to_path_buf())
    } else {
        Ok(std::env::current_dir()?.join(p))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_format_ts_and_js_variants() {
        assert!(should_format("foo.ts"));
        assert!(should_format("foo.tsx"));
        assert!(should_format("foo.js"));
        assert!(should_format("foo.jsx"));
        assert!(should_format("/abs/path/file.ts"));
    }

    #[test]
    fn should_not_format_other_extensions() {
        assert!(!should_format("foo.py"));
        assert!(!should_format("foo.md"));
        assert!(!should_format("foo.rs"));
        assert!(!should_format("foo"));
        assert!(!should_format(""));
    }

    #[test]
    fn should_not_format_mismatched_suffix() {
        assert!(!should_format("foo.tsv"));
        assert!(!should_format("foo.json"));
    }
}
