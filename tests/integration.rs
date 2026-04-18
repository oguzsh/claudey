use std::io::Write;
use std::process::{Command, Stdio};

/// Run the binary with the given subcommand, optionally piping JSON to stdin.
/// Returns `(success, stdout, stderr, exit_code)`.
fn run(subcmd: &str, stdin_json: Option<&str>) -> (bool, String, String, Option<i32>) {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_claudey"));
    if !subcmd.is_empty() {
        cmd.arg(subcmd);
    }
    cmd.stdin(Stdio::piped()).stdout(Stdio::piped()).stderr(Stdio::piped());
    let mut child = cmd.spawn().expect("spawn binary");
    if let Some(json) = stdin_json {
        let mut sin = child.stdin.take().expect("stdin");
        sin.write_all(json.as_bytes()).expect("write stdin");
    }
    let out = child.wait_with_output().expect("wait");
    (
        out.status.success(),
        String::from_utf8_lossy(&out.stdout).to_string(),
        String::from_utf8_lossy(&out.stderr).to_string(),
        out.status.code(),
    )
}

#[test]
fn prints_usage_when_no_subcommand() {
    let (ok, _, stderr, _) = run("", None);
    assert!(!ok);
    assert!(stderr.contains("Usage: claudey <subcommand>"));
}

#[test]
fn unknown_subcommand_fails_with_message() {
    let (ok, _, stderr, _) = run("definitely-not-a-hook", None);
    assert!(!ok);
    assert!(stderr.contains("Unknown subcommand"));
}

// ── routing tests: each in-scope subcommand reaches its stub ──────────────

#[test]
fn routes_session_start() {
    let (_, _, stderr, _) = run("session-start", None);
    assert!(stderr.contains("session-start"));
}

#[test]
fn routes_session_end() {
    let (_, _, stderr, _) = run("session-end", Some("{}"));
    assert!(stderr.contains("session-end"));
}

#[test]
fn routes_pre_compact() {
    let (_, _, stderr, _) = run("pre-compact", None);
    assert!(stderr.contains("pre-compact"));
}

#[test]
fn routes_suggest_compact() {
    let (_, _, stderr, _) = run("suggest-compact", None);
    assert!(stderr.contains("suggest-compact"));
}

#[test]
fn routes_post_edit_format() {
    let (_, _, stderr, _) = run("post-edit-format", Some(r#"{"tool_input":{"file_path":"a.ts"}}"#));
    assert!(stderr.contains("post-edit-format"));
}

#[test]
fn routes_post_edit_typecheck() {
    let (_, _, stderr, _) = run("post-edit-typecheck", Some(r#"{"tool_input":{"file_path":"a.ts"}}"#));
    assert!(stderr.contains("post-edit-typecheck"));
}

#[test]
fn routes_post_edit_console_warn() {
    let (_, _, stderr, _) = run(
        "post-edit-console-warn",
        Some(r#"{"tool_input":{"file_path":"a.ts"}}"#),
    );
    assert!(stderr.contains("post-edit-console-warn"));
}

#[test]
fn routes_check_console_log() {
    let (_, _, stderr, _) = run("check-console-log", Some("{}"));
    assert!(stderr.contains("check-console-log"));
}

#[test]
fn routes_evaluate_session() {
    let (_, _, stderr, _) = run("evaluate-session", Some("{}"));
    assert!(stderr.contains("evaluate-session"));
}

#[test]
fn routes_git_push_reminder() {
    let (_, _, stderr, _) = run(
        "git-push-reminder",
        Some(r#"{"tool_input":{"command":"git push"}}"#),
    );
    assert!(stderr.contains("git-push-reminder"));
}

#[test]
fn routes_block_random_docs() {
    // Stub returns 0 → success.
    let (_, _, stderr, code) = run(
        "block-random-docs",
        Some(r#"{"tool_input":{"file_path":"README.md"}}"#),
    );
    assert!(stderr.contains("block-random-docs"));
    assert_eq!(code, Some(0));
}

#[test]
fn routes_pr_created_log() {
    let (_, _, stderr, _) = run(
        "pr-created-log",
        Some(r#"{"tool_input":{"command":"gh pr create"}}"#),
    );
    assert!(stderr.contains("pr-created-log"));
}
