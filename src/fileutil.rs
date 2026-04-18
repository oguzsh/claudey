//! fileutil — cross-platform file operations for hooks.
//!
//! Ports internal/fileutil/fileutil.go and internal/fileutil/replace.go.
//! Uses only `std` + `regex` (no `walkdir`).

#![allow(dead_code)]

use regex::Regex;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::time::SystemTime;

// ── read / write / append / exists ────────────────────────────────────────

/// Read a file as UTF-8 text, or `None` on any error.
pub fn read_file(path: &Path) -> Option<String> {
    fs::read_to_string(path).ok()
}

/// Write `content` to `path`, creating parent directories as needed.
pub fn write_file(path: &Path, content: &str) -> std::io::Result<()> {
    if let Some(parent) = path.parent() {
        if !parent.as_os_str().is_empty() {
            fs::create_dir_all(parent)?;
        }
    }
    fs::write(path, content)
}

/// Append `content` to `path`, creating parent directories as needed.
pub fn append_file(path: &Path, content: &str) -> std::io::Result<()> {
    if let Some(parent) = path.parent() {
        if !parent.as_os_str().is_empty() {
            fs::create_dir_all(parent)?;
        }
    }
    let mut f = fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(path)?;
    f.write_all(content.as_bytes())
}

/// True when `path` can be stat'd (file, dir, link — anything).
pub fn exists(path: &Path) -> bool {
    fs::metadata(path).is_ok()
}

// ── find_files ────────────────────────────────────────────────────────────

pub struct FileResult {
    pub path: PathBuf,
    pub mtime: SystemTime,
}

pub struct GrepResult {
    pub line_number: usize,
    pub content: String,
}

/// Find files in `dir` whose filename matches the glob `pattern`.
/// `max_age_days` of `0.0` disables the age filter. When `recursive`,
/// descends into subdirectories. Results are sorted newest-first by mtime.
pub fn find_files(
    dir: &Path,
    pattern: &str,
    max_age_days: f64,
    recursive: bool,
) -> Vec<FileResult> {
    if pattern.is_empty() {
        return Vec::new();
    }
    let meta = match fs::metadata(dir) {
        Ok(m) => m,
        Err(_) => return Vec::new(),
    };
    if !meta.is_dir() {
        return Vec::new();
    }
    let re = match Regex::new(&format!("^{}$", glob_to_regex(pattern))) {
        Ok(r) => r,
        Err(_) => return Vec::new(),
    };
    let now = SystemTime::now();
    let mut results = Vec::new();
    collect_files(dir, &re, max_age_days, now, recursive, &mut results);
    // Newest first.
    results.sort_by(|a, b| b.mtime.cmp(&a.mtime));
    results
}

fn collect_files(
    dir: &Path,
    re: &Regex,
    max_age_days: f64,
    now: SystemTime,
    recursive: bool,
    out: &mut Vec<FileResult>,
) {
    let entries = match fs::read_dir(dir) {
        Ok(e) => e,
        Err(_) => return,
    };
    for entry in entries.flatten() {
        let ft = match entry.file_type() {
            Ok(ft) => ft,
            Err(_) => continue,
        };
        let path = entry.path();
        if ft.is_dir() {
            if recursive {
                collect_files(&path, re, max_age_days, now, recursive, out);
            }
            continue;
        }
        if !ft.is_file() {
            continue;
        }
        let name = match entry.file_name().into_string() {
            Ok(s) => s,
            Err(_) => continue,
        };
        if !re.is_match(&name) {
            continue;
        }
        let meta = match entry.metadata() {
            Ok(m) => m,
            Err(_) => continue,
        };
        let mtime = meta.modified().unwrap_or(now);
        if max_age_days > 0.0 {
            if let Ok(age) = now.duration_since(mtime) {
                if age.as_secs_f64() / 86400.0 > max_age_days {
                    continue;
                }
            }
        }
        out.push(FileResult { path, mtime });
    }
}

// ── grep / count ──────────────────────────────────────────────────────────

/// Return `(line_number, content)` for every line matching `pattern`.
/// Line numbers are 1-indexed. Invalid regex yields an empty result.
pub fn grep_file(path: &Path, pattern: &str) -> Vec<GrepResult> {
    let content = match read_file(path) {
        Some(c) => c,
        None => return Vec::new(),
    };
    let re = match Regex::new(pattern) {
        Ok(r) => r,
        Err(_) => return Vec::new(),
    };
    let mut out = Vec::new();
    for (i, line) in content.split('\n').enumerate() {
        if re.is_match(line) {
            out.push(GrepResult {
                line_number: i + 1,
                content: line.to_string(),
            });
        }
    }
    out
}

/// Count regex occurrences in `path`. Missing file / bad regex → `0`.
pub fn count_in_file(path: &Path, pattern: &str) -> usize {
    let content = match read_file(path) {
        Some(c) => c,
        None => return 0,
    };
    let re = match Regex::new(pattern) {
        Ok(r) => r,
        Err(_) => return 0,
    };
    re.find_iter(&content).count()
}

// ── replace_* ─────────────────────────────────────────────────────────────

/// Replace the first occurrence of the exact string `search`.
/// Returns true when the file was modified.
pub fn replace_in_file(path: &Path, search: &str, replace: &str) -> bool {
    let content = match read_file(path) {
        Some(c) => c,
        None => return false,
    };
    let new_content = content.replacen(search, replace, 1);
    if new_content == content {
        return false;
    }
    write_file(path, &new_content).is_ok()
}

/// Replace every exact-string occurrence of `search`.
pub fn replace_all_in_file(path: &Path, search: &str, replace: &str) -> bool {
    let content = match read_file(path) {
        Some(c) => c,
        None => return false,
    };
    let new_content = content.replace(search, replace);
    if new_content == content {
        return false;
    }
    write_file(path, &new_content).is_ok()
}

/// Replace every regex match of `pattern`.
pub fn replace_regex_in_file(path: &Path, pattern: &str, replace: &str) -> bool {
    let content = match read_file(path) {
        Some(c) => c,
        None => return false,
    };
    let re = match Regex::new(pattern) {
        Ok(r) => r,
        Err(_) => return false,
    };
    let new_content = re.replace_all(&content, replace).to_string();
    if new_content == content {
        return false;
    }
    write_file(path, &new_content).is_ok()
}

// ── glob helper ───────────────────────────────────────────────────────────

/// Convert a shell-style glob to a regex body (no anchors).
/// Only handles `*` and `?`; escapes regex metacharacters.
pub fn glob_to_regex(pattern: &str) -> String {
    let mut out = String::with_capacity(pattern.len() * 2);
    for ch in pattern.chars() {
        match ch {
            '*' => out.push_str(".*"),
            '?' => out.push('.'),
            '.' | '+' | '^' | '$' | '{' | '}' | '(' | ')' | '|' | '[' | ']' | '\\' => {
                out.push('\\');
                out.push(ch);
            }
            _ => out.push(ch),
        }
    }
    out
}

// ── tests ─────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use crate::testutil::TempDir;

    #[test]
    fn test_read_write_file() {
        let d = TempDir::new();
        let p = d.path().join("test.txt");
        write_file(&p, "hello world").unwrap();
        assert_eq!(read_file(&p).unwrap(), "hello world");
    }

    #[test]
    fn test_read_file_missing() {
        assert!(read_file(Path::new("/nonexistent/file.txt")).is_none());
    }

    #[test]
    fn test_write_file_creates_parent_dirs() {
        let d = TempDir::new();
        let p = d.path().join("a").join("b").join("c").join("deep.txt");
        write_file(&p, "deep").unwrap();
        assert_eq!(read_file(&p).unwrap(), "deep");
    }

    #[test]
    fn test_append_file() {
        let d = TempDir::new();
        let p = d.path().join("append.txt");
        write_file(&p, "line1\n").unwrap();
        append_file(&p, "line2\n").unwrap();
        assert_eq!(read_file(&p).unwrap(), "line1\nline2\n");
    }

    #[test]
    fn test_find_files_glob_match() {
        let d = TempDir::new();
        write_file(&d.path().join("a-session.tmp"), "a").unwrap();
        write_file(&d.path().join("b-session.tmp"), "b").unwrap();
        write_file(&d.path().join("readme.md"), "readme").unwrap();

        let results = find_files(d.path(), "*-session.tmp", 0.0, false);
        assert_eq!(results.len(), 2);
    }

    #[test]
    fn test_find_files_nonexistent_dir() {
        let results = find_files(Path::new("/nonexistent"), "*.tmp", 0.0, false);
        assert!(results.is_empty());
    }

    #[test]
    fn test_find_files_recursive() {
        let d = TempDir::new();
        write_file(&d.path().join("top.tmp"), "x").unwrap();
        write_file(&d.path().join("sub/nested.tmp"), "y").unwrap();

        let shallow = find_files(d.path(), "*.tmp", 0.0, false);
        assert_eq!(shallow.len(), 1, "non-recursive should only see top.tmp");

        let deep = find_files(d.path(), "*.tmp", 0.0, true);
        assert_eq!(deep.len(), 2, "recursive should find nested.tmp too");
    }

    #[test]
    fn test_grep_file_line_numbers() {
        let d = TempDir::new();
        let p = d.path().join("test.js");
        write_file(
            &p,
            "line1\nconsole.log('debug')\nline3\nconsole.log('test')\n",
        )
        .unwrap();

        let results = grep_file(&p, r"console\.log");
        assert_eq!(results.len(), 2);
        assert_eq!(results[0].line_number, 2);
        assert_eq!(results[1].line_number, 4);
    }

    #[test]
    fn test_count_in_file() {
        let d = TempDir::new();
        let p = d.path().join("test.txt");
        write_file(
            &p,
            r#""type":"user" and "type":"user" plus "type":"assistant""#,
        )
        .unwrap();

        let n = count_in_file(&p, r#""type"\s*:\s*"user""#);
        assert_eq!(n, 2);
    }

    #[test]
    fn test_replace_in_file() {
        let d = TempDir::new();
        let p = d.path().join("test.txt");
        write_file(&p, "**Last Updated:** 10:00").unwrap();

        assert!(replace_in_file(
            &p,
            "**Last Updated:** 10:00",
            "**Last Updated:** 11:30"
        ));
        assert_eq!(read_file(&p).unwrap(), "**Last Updated:** 11:30");
    }

    #[test]
    fn test_replace_in_file_no_match_returns_false() {
        let d = TempDir::new();
        let p = d.path().join("test.txt");
        write_file(&p, "hello").unwrap();
        assert!(!replace_in_file(&p, "nope", "yes"));
    }

    #[test]
    fn test_replace_regex_in_file() {
        let d = TempDir::new();
        let p = d.path().join("test.txt");
        write_file(&p, "**Last Updated:** 10:00\nother line").unwrap();

        assert!(replace_regex_in_file(
            &p,
            r"\*\*Last Updated:\*\*.*",
            "**Last Updated:** 15:30",
        ));
        assert_eq!(
            read_file(&p).unwrap(),
            "**Last Updated:** 15:30\nother line"
        );
    }

    #[test]
    fn test_replace_all_in_file() {
        let d = TempDir::new();
        let p = d.path().join("test.txt");
        write_file(&p, "foo bar foo baz foo").unwrap();
        assert!(replace_all_in_file(&p, "foo", "qux"));
        assert_eq!(read_file(&p).unwrap(), "qux bar qux baz qux");
    }

    #[test]
    fn test_exists() {
        let d = TempDir::new();
        let p = d.path().join("exists.txt");
        write_file(&p, "content").unwrap();
        assert!(exists(&p));
        assert!(!exists(&d.path().join("nope.txt")));
    }

    #[test]
    fn test_glob_to_regex_cases() {
        let cases = [
            ("*.tmp", "session.tmp", true),
            ("*.tmp", "session.txt", false),
            ("*-session.tmp", "2024-01-01-session.tmp", true),
            ("test.?s", "test.js", true),
            ("test.?s", "test.ts", true),
            ("test.?s", "test.xy", false), // one '?' == exactly one char
        ];
        for (glob, input, want) in cases {
            let re = Regex::new(&format!("^{}$", glob_to_regex(glob))).unwrap();
            assert_eq!(re.is_match(input), want, "glob={glob:?} input={input:?}");
        }
    }

    #[test]
    fn test_find_files_sorted_newest_first() {
        let d = TempDir::new();
        let older = d.path().join("older.tmp");
        let newer = d.path().join("newer.tmp");
        write_file(&older, "o").unwrap();
        // Ensure distinct mtimes on fast filesystems.
        std::thread::sleep(std::time::Duration::from_millis(15));
        write_file(&newer, "n").unwrap();

        let results = find_files(d.path(), "*.tmp", 0.0, false);
        assert_eq!(results.len(), 2);
        assert_eq!(results[0].path, newer);
        assert_eq!(results[1].path, older);
    }
}
