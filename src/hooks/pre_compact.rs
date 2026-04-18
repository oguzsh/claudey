//! pre-compact hook — save state before context compaction.
//!
//! Ports internal/hooks/precompact.go.

use crate::{datetime, fileutil, hookio, platform};

pub fn pre_compact() {
    let sessions_dir = platform::sessions_dir();
    let compaction_log = sessions_dir.join("compaction-log.txt");

    let _ = platform::ensure_dir(&sessions_dir);

    let timestamp = datetime::datetime_string();
    let _ = fileutil::append_file(
        &compaction_log,
        &format!("[{timestamp}] Context compaction triggered\n"),
    );

    let sessions = fileutil::find_files(&sessions_dir, "*-session.tmp", 0.0, false);
    if let Some(latest) = sessions.first() {
        let time_str = datetime::time_string();
        let _ = fileutil::append_file(
            &latest.path,
            &format!(
                "\n---\n**[Compaction occurred at {time_str}]** - Context was summarized\n"
            ),
        );
    }

    hookio::log("[PreCompact] State saved before compaction");
}

// Behavior is exercised via tests/integration.rs — a unit test here would
// need to mutate $HOME globally, racing with other parallel tests.
