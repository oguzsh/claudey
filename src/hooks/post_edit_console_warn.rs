//! post-edit-console-warn hook — warn about console.log in edited JS/TS files.
//!
//! Ports internal/hooks/posteditconsolewarn.go.

use crate::{fileutil, hookio};
use serde_json::Value;
use std::path::Path;

const MAX_REPORTED: usize = 5;

pub fn post_edit_console_warn(input: Value, raw: Vec<u8>) {
    let file_path = hookio::get_tool_input_string(&input, "file_path");

    if !file_path.is_empty() && is_jsts(file_path) {
        if let Some(matches) = scan_for_console_log(Path::new(file_path)) {
            if !matches.is_empty() {
                hookio::log(&format!("[Hook] WARNING: console.log found in {file_path}"));
                for line in matches.iter().take(MAX_REPORTED) {
                    hookio::log(line);
                }
                hookio::log("[Hook] Remove console.log before committing");
            }
        }
    }

    hookio::passthrough(&raw);
}

pub fn is_jsts(path: &str) -> bool {
    path.ends_with(".ts")
        || path.ends_with(".tsx")
        || path.ends_with(".js")
        || path.ends_with(".jsx")
}

pub fn scan_for_console_log(path: &Path) -> Option<Vec<String>> {
    let content = fileutil::read_file(path)?;
    let mut out = Vec::new();
    for (i, line) in content.split('\n').enumerate() {
        if line.contains("console.log") {
            out.push(format!("{}: {}", i + 1, line.trim()));
        }
    }
    Some(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn scan_finds_console_log_lines_with_one_indexed_numbers() {
        let base = crate::testutil::TempDir::new();
        let file = base.path().join("a.ts");
        std::fs::write(
            &file,
            "const a = 1;\nconsole.log('x');\nconst b = 2;\nconsole.log('y');\n",
        )
        .unwrap();

        let matches = scan_for_console_log(&file).expect("file should be readable");
        assert_eq!(matches.len(), 2);
        assert!(matches[0].starts_with("2:"));
        assert!(matches[1].starts_with("4:"));
    }

    #[test]
    fn scan_returns_empty_for_clean_file() {
        let base = crate::testutil::TempDir::new();
        let file = base.path().join("a.ts");
        std::fs::write(&file, "const a = 1;\n").unwrap();
        assert!(scan_for_console_log(&file).unwrap().is_empty());
    }

    #[test]
    fn scan_missing_file_returns_none() {
        let missing = Path::new("/nonexistent/claudey-test-missing.ts");
        assert!(scan_for_console_log(missing).is_none());
    }

    #[test]
    fn is_jsts_accepts_known_extensions() {
        assert!(is_jsts("a.ts"));
        assert!(is_jsts("a.tsx"));
        assert!(is_jsts("a.js"));
        assert!(is_jsts("a.jsx"));
        assert!(!is_jsts("a.py"));
        assert!(!is_jsts("a"));
    }
}
