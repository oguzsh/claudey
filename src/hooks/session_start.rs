//! session-start hook — load previous context at session start.
//!
//! Ports internal/hooks/sessionstart.go.

use crate::{aliases, fileutil, hookio, platform};

pub fn session_start() {
    let sessions_dir = platform::sessions_dir();
    let learned_dir = platform::learned_skills_dir();

    let _ = platform::ensure_dir(&sessions_dir);
    let _ = platform::ensure_dir(&learned_dir);

    let recent = fileutil::find_files(&sessions_dir, "*-session.tmp", 7.0, false);
    if !recent.is_empty() {
        let latest = &recent[0];
        hookio::log(&format!(
            "[SessionStart] Found {} recent session(s)",
            recent.len()
        ));
        hookio::log(&format!("[SessionStart] Latest: {}", latest.path.display()));

        if let Some(content) = fileutil::read_file(&latest.path) {
            if !content.contains("[Session context goes here]") {
                hookio::output_string(&format!("Previous session summary:\n{content}"));
            }
        }
    }

    let learned = fileutil::find_files(&learned_dir, "*.md", 0.0, false);
    if !learned.is_empty() {
        hookio::log(&format!(
            "[SessionStart] {} learned skill(s) available in {}",
            learned.len(),
            learned_dir.display()
        ));
    }

    let alias_list = aliases::list(None, 5);
    if !alias_list.is_empty() {
        let names: Vec<&str> = alias_list.iter().map(|a| a.name.as_str()).collect();
        hookio::log(&format!(
            "[SessionStart] {} session alias(es) available: {}",
            alias_list.len(),
            names.join(", ")
        ));
        hookio::log("[SessionStart] Use /sessions load <alias> to continue a previous session");
    }
}
