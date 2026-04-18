//! inline hooks — git_push_reminder, block_random_docs, pr_created_log.
//!
//! Ports internal/hooks/inline.go.

use crate::hookio;
use regex::Regex;
use serde_json::Value;
use std::sync::LazyLock;

static GIT_PUSH_RE: LazyLock<Regex> = LazyLock::new(|| Regex::new(r"git push").unwrap());
static RANDOM_DOC_RE: LazyLock<Regex> = LazyLock::new(|| Regex::new(r"\.(md|txt)$").unwrap());
static ALLOWED_DOC_RE: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r"(README|CLAUDE|AGENTS|CONTRIBUTING)\.md$").unwrap());
static PLANS_PATH_RE: LazyLock<Regex> = LazyLock::new(|| Regex::new(r"\.claude/plans/").unwrap());
static PR_URL_RE: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r"https://github\.com/[^/]+/[^/]+/pull/\d+").unwrap());
static GH_PR_CREATE_RE: LazyLock<Regex> = LazyLock::new(|| Regex::new(r"gh pr create").unwrap());

pub fn git_push_reminder(input: Value, raw: Vec<u8>) {
    let cmd = hookio::get_tool_input_string(&input, "command");
    if GIT_PUSH_RE.is_match(cmd) {
        hookio::log("[Hook] Review changes before push...");
        hookio::log("[Hook] Continuing with push (remove this hook to add interactive review)");
    }
    hookio::passthrough(&raw);
}

pub fn block_random_docs(input: Value, raw: Vec<u8>) -> i32 {
    let file_path = hookio::get_tool_input_string(&input, "file_path");
    if is_blocked_doc(file_path) {
        hookio::log("[Hook] BLOCKED: Unnecessary documentation file creation");
        hookio::log(&format!("[Hook] File: {file_path}"));
        hookio::log("[Hook] Use README.md for documentation instead");
        hookio::passthrough(&raw);
        return 2;
    }
    hookio::passthrough(&raw);
    0
}

pub fn pr_created_log(input: Value, raw: Vec<u8>) {
    let cmd = hookio::get_tool_input_string(&input, "command");
    if GH_PR_CREATE_RE.is_match(cmd) {
        let output = hookio::get_tool_output_string(&input, "output");
        if let Some(url) = extract_pr_url(output) {
            hookio::log(&format!("[Hook] PR created: {url}"));

            let parts: Vec<&str> = url.split('/').collect();
            if parts.len() >= 7 {
                let repo = format!("{}/{}", parts[3], parts[4]);
                let pr = parts[6];
                hookio::log(&format!(
                    "[Hook] To review: gh pr review {pr} --repo {repo}"
                ));
            }
        }
    }
    hookio::passthrough(&raw);
}

pub fn is_blocked_doc(file_path: &str) -> bool {
    !file_path.is_empty()
        && RANDOM_DOC_RE.is_match(file_path)
        && !ALLOWED_DOC_RE.is_match(file_path)
        && !PLANS_PATH_RE.is_match(file_path)
}

pub fn extract_pr_url(output: &str) -> Option<&str> {
    PR_URL_RE.find(output).map(|m| m.as_str())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn blocks_random_md_txt() {
        assert!(is_blocked_doc("notes.md"));
        assert!(is_blocked_doc("docs/foo.md"));
        assert!(is_blocked_doc("scratch.txt"));
    }

    #[test]
    fn allows_known_doc_names() {
        assert!(!is_blocked_doc("README.md"));
        assert!(!is_blocked_doc("path/to/CLAUDE.md"));
        assert!(!is_blocked_doc("AGENTS.md"));
        assert!(!is_blocked_doc("CONTRIBUTING.md"));
    }

    #[test]
    fn allows_plans_directory() {
        assert!(!is_blocked_doc(".claude/plans/2026-04-19.md"));
        assert!(!is_blocked_doc("/home/u/.claude/plans/todo.md"));
    }

    #[test]
    fn allows_non_doc_extensions() {
        assert!(!is_blocked_doc("foo.rs"));
        assert!(!is_blocked_doc("foo.ts"));
        assert!(!is_blocked_doc(""));
    }

    #[test]
    fn extract_pr_url_happy_path() {
        let out = "Created https://github.com/oguzsh/claudey/pull/42 successfully";
        assert_eq!(
            extract_pr_url(out),
            Some("https://github.com/oguzsh/claudey/pull/42")
        );
    }

    #[test]
    fn extract_pr_url_absent_returns_none() {
        assert!(extract_pr_url("nothing to see").is_none());
    }
}
