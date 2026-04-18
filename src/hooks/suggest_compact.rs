//! suggest-compact hook — suggest manual compaction at tool-call intervals.
//!
//! Ports internal/hooks/suggestcompact.go.

use crate::{hookio, platform};
use std::fs::OpenOptions;
use std::io::{Read, Seek, SeekFrom, Write};

const DEFAULT_THRESHOLD: i64 = 50;
const MAX_THRESHOLD: i64 = 10_000;
const MAX_COUNT: i64 = 1_000_000;

pub fn suggest_compact() {
    let session_id = session_id();
    let counter_file = platform::temp_dir().join(format!("claude-tool-count-{session_id}"));

    let threshold = threshold_from_env();

    let count = increment_counter(&counter_file);

    if count == threshold {
        hookio::log(&format!(
            "[StrategicCompact] {threshold} tool calls reached - consider /compact if transitioning phases"
        ));
    }

    if count > threshold && (count - threshold) % 25 == 0 {
        hookio::log(&format!(
            "[StrategicCompact] {count} tool calls - good checkpoint for /compact if context is stale"
        ));
    }
}

fn session_id() -> String {
    match std::env::var("CLAUDE_SESSION_ID") {
        Ok(s) if !s.is_empty() => s,
        _ => "default".to_string(),
    }
}

fn threshold_from_env() -> i64 {
    match std::env::var("COMPACT_THRESHOLD") {
        Ok(s) => match s.parse::<i64>() {
            Ok(v) if v > 0 && v <= MAX_THRESHOLD => v,
            _ => DEFAULT_THRESHOLD,
        },
        Err(_) => DEFAULT_THRESHOLD,
    }
}

fn increment_counter(path: &std::path::Path) -> i64 {
    let mut count: i64 = 1;
    if let Ok(mut f) = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(path)
    {
        let mut buf = [0u8; 64];
        let n = f.read(&mut buf).unwrap_or(0);
        if n > 0 {
            if let Ok(s) = std::str::from_utf8(&buf[..n]) {
                if let Ok(parsed) = s.trim().parse::<i64>() {
                    if parsed > 0 && parsed <= MAX_COUNT {
                        count = parsed + 1;
                    }
                }
            }
        }
        let _ = f.set_len(0);
        let _ = f.seek(SeekFrom::Start(0));
        let _ = write!(f, "{count}");
    }
    count
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn increment_counter_starts_at_one() {
        let base = crate::testutil::TempDir::new();
        let f = base.path().join("counter");
        assert_eq!(increment_counter(&f), 1);
    }

    #[test]
    fn increment_counter_increments_existing() {
        let base = crate::testutil::TempDir::new();
        let f = base.path().join("counter");
        std::fs::write(&f, b"5").unwrap();
        assert_eq!(increment_counter(&f), 6);
    }

    #[test]
    fn increment_counter_resets_on_garbage() {
        let base = crate::testutil::TempDir::new();
        let f = base.path().join("counter");
        std::fs::write(&f, b"not-a-number").unwrap();
        assert_eq!(increment_counter(&f), 1);
    }

    #[test]
    fn increment_counter_resets_when_over_max() {
        let base = crate::testutil::TempDir::new();
        let f = base.path().join("counter");
        std::fs::write(&f, format!("{}", MAX_COUNT + 1)).unwrap();
        assert_eq!(increment_counter(&f), 1);
    }

    #[test]
    fn increment_counter_persists_value() {
        let base = crate::testutil::TempDir::new();
        let f = base.path().join("counter");
        increment_counter(&f);
        increment_counter(&f);
        let read = std::fs::read_to_string(&f).unwrap();
        assert_eq!(read.trim(), "2");
    }
}
