//! evaluate-session hook — evaluate session for extractable patterns.
//!
//! Ports internal/hooks/evaluatesession.go.

use crate::{fileutil, hookio, platform};
use serde::Deserialize;
use serde_json::Value;
use std::path::{Path, PathBuf};

#[derive(Deserialize, Default)]
struct Config {
    #[serde(default)]
    min_session_length: i64,
    #[serde(default)]
    learned_skills_path: String,
}

pub fn evaluate_session(input: Value, plugin_root: String) {
    let transcript_path = input
        .get("transcript_path")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .filter(|s| !s.is_empty())
        .or_else(|| std::env::var("CLAUDE_TRANSCRIPT_PATH").ok())
        .unwrap_or_default();

    let config_file = Path::new(&plugin_root)
        .join("skills")
        .join("continuous-learning")
        .join("config.json");

    let mut min_session_length: i64 = 10;
    let mut learned_skills_path: PathBuf = platform::learned_skills_dir();

    if let Some(content) = fileutil::read_file(&config_file) {
        match serde_json::from_str::<Config>(&content) {
            Ok(cfg) => {
                if cfg.min_session_length > 0 {
                    min_session_length = cfg.min_session_length;
                }
                if !cfg.learned_skills_path.is_empty() {
                    learned_skills_path = expand_tilde(&cfg.learned_skills_path);
                }
            }
            Err(e) => hookio::log(&format!(
                "[ContinuousLearning] Failed to parse config: {e}, using defaults"
            )),
        }
    }

    let _ = platform::ensure_dir(&learned_skills_path);

    if transcript_path.is_empty() || !Path::new(&transcript_path).exists() {
        return;
    }

    let message_count =
        fileutil::count_in_file(Path::new(&transcript_path), r#""type"\s*:\s*"user""#) as i64;

    if message_count < min_session_length {
        hookio::log(&format!(
            "[ContinuousLearning] Session too short ({message_count} messages), skipping"
        ));
        return;
    }

    hookio::log(&format!(
        "[ContinuousLearning] Session has {message_count} messages - evaluate for extractable patterns"
    ));
    hookio::log(&format!(
        "[ContinuousLearning] Save learned skills to: {}",
        learned_skills_path.display()
    ));
}

fn expand_tilde(path: &str) -> PathBuf {
    if let Some(rest) = path.strip_prefix('~') {
        platform::home_dir().join(rest.trim_start_matches('/'))
    } else {
        PathBuf::from(path)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn expand_tilde_substitutes_home() {
        let expanded = expand_tilde("~/foo/bar");
        assert!(expanded.ends_with("foo/bar"));
        assert!(expanded.is_absolute());
    }

    #[test]
    fn expand_tilde_leaves_absolute_unchanged() {
        let expanded = expand_tilde("/etc/hosts");
        assert_eq!(expanded, PathBuf::from("/etc/hosts"));
    }

    #[test]
    fn count_user_messages_via_count_in_file() {
        let base = crate::testutil::TempDir::new();
        let transcript = base.path().join("t.jsonl");
        std::fs::write(
            &transcript,
            r#"{"type":"user","message":"a"}
{"type":"assistant","message":"b"}
{"type":"user","message":"c"}
"#,
        )
        .unwrap();

        let n = fileutil::count_in_file(&transcript, r#""type"\s*:\s*"user""#);
        assert_eq!(n, 2);
    }
}
