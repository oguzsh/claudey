//! session-end hook — persists session state at session end.
//!
//! Ports internal/hooks/sessionend.go.

use crate::{datetime, fileutil, gitutil, hookio, platform};
use serde_json::Value;
use std::collections::BTreeSet;
use std::path::Path;

pub struct SessionSummary {
    pub user_messages: Vec<String>,
    pub tools_used: Vec<String>,
    pub files_modified: Vec<String>,
    pub total_messages: usize,
}

pub fn session_end(input: Value) {
    let transcript_path = get_str_field(&input, "transcript_path").to_string();
    let transcript_path = if transcript_path.is_empty() {
        std::env::var("CLAUDE_TRANSCRIPT_PATH").unwrap_or_default()
    } else {
        transcript_path
    };

    let sessions_dir = platform::sessions_dir();
    let today = datetime::date_string();
    let short_id = session_id_short();
    let session_file = sessions_dir.join(format!("{today}-{short_id}-session.tmp"));

    let _ = platform::ensure_dir(&sessions_dir);

    let current_time = datetime::time_string();

    let mut summary = None;
    if !transcript_path.is_empty() {
        let tp = Path::new(&transcript_path);
        if fileutil::exists(tp) {
            summary = extract_session_summary(tp);
        } else {
            hookio::log(&format!(
                "[SessionEnd] Transcript not found: {transcript_path}"
            ));
        }
    }

    if fileutil::exists(&session_file) {
        fileutil::replace_regex_in_file(
            &session_file,
            r"\*\*Last Updated:\*\*.*",
            &format!("**Last Updated:** {current_time}"),
        );

        if let Some(s) = &summary {
            if let Some(existing) = fileutil::read_file(&session_file) {
                if existing.contains("[Session context goes here]") {
                    // Replace the whole placeholder section with the real summary.
                    let template_re = r"(?s)## Current State\s*\n\s*\[Session context goes here\].*?### Context to Load\s*\n```\s*\n\[relevant files\]\s*\n```";
                    fileutil::replace_regex_in_file(
                        &session_file,
                        template_re,
                        &build_summary_section(s),
                    );
                }
            }
        }

        hookio::log(&format!(
            "[SessionEnd] Updated session file: {}",
            session_file.display()
        ));
    } else {
        let summary_section = match &summary {
            Some(s) => build_summary_section(s),
            None => String::from(
                "## Current State\n\n[Session context goes here]\n\n### Completed\n- [ ]\n\n### In Progress\n- [ ]\n\n### Notes for Next Session\n-\n\n### Context to Load\n```\n[relevant files]\n```",
            ),
        };
        let template = format!(
            "# Session: {today}\n**Date:** {today}\n**Started:** {current_time}\n**Last Updated:** {current_time}\n\n---\n\n{summary_section}\n"
        );
        if let Err(e) = fileutil::write_file(&session_file, &template) {
            hookio::log(&format!("[SessionEnd] Failed to write session file: {e}"));
            return;
        }
        hookio::log(&format!(
            "[SessionEnd] Created session file: {}",
            session_file.display()
        ));
    }
}

pub fn extract_session_summary(transcript_path: &Path) -> Option<SessionSummary> {
    let content = fileutil::read_file(transcript_path)?;
    let lines: Vec<&str> = content.split('\n').collect();
    let mut user_messages: Vec<String> = Vec::new();
    let mut tools_used: BTreeSet<String> = BTreeSet::new();
    let mut files_modified: BTreeSet<String> = BTreeSet::new();
    let mut parse_errors = 0usize;

    for line in &lines {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }
        let entry: Value = match serde_json::from_str(line) {
            Ok(v) => v,
            Err(_) => {
                parse_errors += 1;
                continue;
            }
        };

        let entry_type = entry.get("type").and_then(|v| v.as_str()).unwrap_or("");
        let role = entry.get("role").and_then(|v| v.as_str()).unwrap_or("");
        let msg_role = entry
            .get("message")
            .and_then(|m| m.get("role"))
            .and_then(|v| v.as_str())
            .unwrap_or("");

        if entry_type == "user" || role == "user" || msg_role == "user" {
            let text = extract_text(&entry);
            if !text.is_empty() {
                let truncated: String = text.chars().take(200).collect();
                user_messages.push(truncated);
            }
        }

        // Direct tool_use entries.
        let has_tool_name = entry.get("tool_name").is_some();
        if entry_type == "tool_use" || has_tool_name {
            let tool_name = str_or(&entry, &["tool_name", "name"]);
            if !tool_name.is_empty() {
                tools_used.insert(tool_name.to_string());
            }
            let file_path = extract_file_path(&entry, "tool_input");
            let file_path = if file_path.is_empty() {
                extract_file_path(&entry, "input")
            } else {
                file_path
            };
            if !file_path.is_empty() && (tool_name == "Edit" || tool_name == "Write") {
                files_modified.insert(file_path);
            }
        }

        // Assistant messages carry tool_use blocks in message.content[].
        if entry_type == "assistant" {
            if let Some(content) = entry
                .get("message")
                .and_then(|m| m.get("content"))
                .and_then(|c| c.as_array())
            {
                for block in content {
                    if block.get("type").and_then(|v| v.as_str()) != Some("tool_use") {
                        continue;
                    }
                    let tool_name = block.get("name").and_then(|v| v.as_str()).unwrap_or("");
                    if !tool_name.is_empty() {
                        tools_used.insert(tool_name.to_string());
                    }
                    if tool_name == "Edit" || tool_name == "Write" {
                        if let Some(fp) = block
                            .get("input")
                            .and_then(|i| i.get("file_path"))
                            .and_then(|v| v.as_str())
                        {
                            if !fp.is_empty() {
                                files_modified.insert(fp.to_string());
                            }
                        }
                    }
                }
            }
        }
    }

    if parse_errors > 0 {
        hookio::log(&format!(
            "[SessionEnd] Skipped {parse_errors}/{} unparseable transcript lines",
            lines.len()
        ));
    }

    if user_messages.is_empty() {
        return None;
    }

    let total_messages = user_messages.len();
    let start = total_messages.saturating_sub(10);
    let last_user = user_messages[start..].to_vec();

    let tools: Vec<String> = tools_used.into_iter().take(20).collect();
    let files: Vec<String> = files_modified.into_iter().take(30).collect();

    Some(SessionSummary {
        user_messages: last_user,
        tools_used: tools,
        files_modified: files,
        total_messages,
    })
}

fn extract_text(entry: &Value) -> String {
    if let Some(content) = entry.get("message").and_then(|m| m.get("content")) {
        return content_to_string(content);
    }
    if let Some(content) = entry.get("content") {
        return content_to_string(content);
    }
    String::new()
}

fn content_to_string(v: &Value) -> String {
    if let Some(s) = v.as_str() {
        return s.trim().to_string();
    }
    if let Some(arr) = v.as_array() {
        let mut parts = Vec::new();
        for item in arr {
            if let Some(text) = item.get("text").and_then(|t| t.as_str()) {
                parts.push(text.to_string());
            }
        }
        return parts.join(" ").trim().to_string();
    }
    String::new()
}

fn extract_file_path(entry: &Value, key: &str) -> String {
    entry
        .get(key)
        .and_then(|i| i.get("file_path"))
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .unwrap_or_default()
}

fn str_or<'a>(entry: &'a Value, keys: &[&str]) -> &'a str {
    for k in keys {
        if let Some(s) = entry.get(*k).and_then(|v| v.as_str()) {
            if !s.is_empty() {
                return s;
            }
        }
    }
    ""
}

fn get_str_field<'a>(v: &'a Value, field: &str) -> &'a str {
    v.get(field).and_then(|f| f.as_str()).unwrap_or("")
}

pub fn build_summary_section(s: &SessionSummary) -> String {
    let mut out = String::new();
    out.push_str("## Session Summary\n\n");
    out.push_str("### Tasks\n");
    for msg in &s.user_messages {
        let escaped = msg.replace('\n', " ").replace('`', "\\`");
        out.push_str("- ");
        out.push_str(&escaped);
        out.push('\n');
    }
    out.push('\n');

    if !s.files_modified.is_empty() {
        out.push_str("### Files Modified\n");
        for f in &s.files_modified {
            out.push_str("- ");
            out.push_str(f);
            out.push('\n');
        }
        out.push('\n');
    }

    if !s.tools_used.is_empty() {
        out.push_str("### Tools Used\n");
        out.push_str(&s.tools_used.join(", "));
        out.push_str("\n\n");
    }

    out.push_str(&format!(
        "### Stats\n- Total user messages: {}\n",
        s.total_messages
    ));
    out
}

fn session_id_short() -> String {
    if let Ok(sid) = std::env::var("CLAUDE_SESSION_ID") {
        if !sid.is_empty() {
            let len = sid.chars().count();
            if len > 8 {
                return sid.chars().skip(len - 8).collect();
            }
            return sid;
        }
    }
    let name = gitutil::project_name();
    if !name.is_empty() {
        return name;
    }
    "default".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::testutil::TempDir;

    fn write_jsonl(path: &Path, lines: &[&str]) {
        let data = lines.join("\n");
        std::fs::write(path, data).unwrap();
    }

    #[test]
    fn extract_from_empty_transcript_returns_none() {
        let d = TempDir::new();
        let p = d.path().join("t.jsonl");
        std::fs::write(&p, "").unwrap();
        assert!(extract_session_summary(&p).is_none());
    }

    #[test]
    fn extract_user_messages_truncated_at_200() {
        let d = TempDir::new();
        let p = d.path().join("t.jsonl");
        let long = "x".repeat(500);
        let line = format!(
            r#"{{"type":"user","message":{{"role":"user","content":"{long}"}}}}"#
        );
        write_jsonl(&p, &[&line]);
        let s = extract_session_summary(&p).unwrap();
        assert_eq!(s.user_messages.len(), 1);
        assert_eq!(s.user_messages[0].chars().count(), 200);
    }

    #[test]
    fn extract_tool_uses_from_assistant_blocks() {
        let d = TempDir::new();
        let p = d.path().join("t.jsonl");
        write_jsonl(
            &p,
            &[
                r#"{"type":"user","message":{"role":"user","content":"go"}}"#,
                r#"{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Edit","input":{"file_path":"/x.rs"}}]}}"#,
                r#"{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}"#,
            ],
        );
        let s = extract_session_summary(&p).unwrap();
        assert!(s.tools_used.iter().any(|t| t == "Edit"));
        assert!(s.tools_used.iter().any(|t| t == "Bash"));
        assert!(s.files_modified.iter().any(|f| f == "/x.rs"));
    }

    #[test]
    fn user_messages_capped_at_ten_last() {
        let d = TempDir::new();
        let p = d.path().join("t.jsonl");
        let mut lines = Vec::new();
        for i in 0..15 {
            lines.push(format!(
                r#"{{"type":"user","message":{{"role":"user","content":"msg-{i}"}}}}"#
            ));
        }
        let refs: Vec<&str> = lines.iter().map(|s| s.as_str()).collect();
        write_jsonl(&p, &refs);
        let s = extract_session_summary(&p).unwrap();
        assert_eq!(s.user_messages.len(), 10);
        assert_eq!(s.total_messages, 15);
        assert_eq!(s.user_messages.first().unwrap(), "msg-5");
        assert_eq!(s.user_messages.last().unwrap(), "msg-14");
    }

    #[test]
    fn build_summary_escapes_backticks_and_newlines() {
        let s = SessionSummary {
            user_messages: vec!["hello `world`\nwith newline".to_string()],
            tools_used: vec!["Edit".to_string()],
            files_modified: vec!["/x.rs".to_string()],
            total_messages: 1,
        };
        let md = build_summary_section(&s);
        assert!(md.contains("- hello \\`world\\` with newline"));
        assert!(md.contains("### Files Modified"));
        assert!(md.contains("/x.rs"));
        assert!(md.contains("- Total user messages: 1"));
    }

    #[test]
    fn session_id_short_prefers_env_last_8() {
        std::env::set_var("CLAUDE_SESSION_ID", "abcdef1234567890");
        assert_eq!(session_id_short(), "34567890");
        std::env::remove_var("CLAUDE_SESSION_ID");
    }
}
