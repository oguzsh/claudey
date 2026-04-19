//! hookio — stdin/stdout hook I/O protocol for Claude Code hooks.
//!
//! Ports internal/hookio/hookio.go.

#![allow(dead_code)]

use serde::Serialize;
use serde_json::Value;
use std::io::{Read, Write};
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

pub const DEFAULT_MAX_SIZE: usize = 1024 * 1024; // 1 MB

/// Read JSON from stdin with a timeout and size limit.
///
/// On timeout, read error, or parse error, returns an empty object and
/// (where available) the raw bytes that were read.  Never panics.
pub fn read_stdin_json(timeout: Duration, max_size: usize) -> (Value, Vec<u8>) {
    let max = if max_size == 0 {
        DEFAULT_MAX_SIZE
    } else {
        max_size
    };

    let (tx, rx) = mpsc::channel::<Vec<u8>>();
    // NOTE: the reader thread is intentionally detached. This hook binary is
    // short-lived, so a blocked stdin read is released by process exit. Do not
    // adapt this into a long-lived process without adding proper cancellation.
    thread::spawn(move || {
        let mut buf = Vec::new();
        let _ = std::io::stdin().take(max as u64).read_to_end(&mut buf);
        let _ = tx.send(buf);
    });

    let raw = match rx.recv_timeout(timeout) {
        Ok(b) => b,
        Err(_) => return (Value::Object(Default::default()), Vec::new()),
    };

    if raw.is_empty() {
        return (Value::Object(Default::default()), raw);
    }

    match serde_json::from_slice::<Value>(&raw) {
        Ok(v @ Value::Object(_)) => (v, raw),
        _ => (Value::Object(Default::default()), raw),
    }
}

/// Write a message to stderr (visible to the user in Claude Code).
pub fn log(msg: &str) {
    let _ = writeln!(std::io::stderr(), "{msg}");
}

/// Write a string to stdout followed by a newline.
pub fn output_string(s: &str) {
    println!("{s}");
}

/// Write bytes to stdout followed by a newline.
pub fn output_bytes(b: &[u8]) {
    let mut out = std::io::stdout();
    let _ = out.write_all(b);
    let _ = out.write_all(b"\n");
}

/// Serialize `v` as JSON and write it to stdout via [`output_bytes`].
/// On serialization error, falls back to writing `{}` (matches Go's behaviour).
pub fn output_json<T: Serialize>(v: &T) {
    match serde_json::to_vec(v) {
        Ok(data) => output_bytes(&data),
        Err(_) => output_bytes(b"{}"),
    }
}

/// Write raw bytes to stdout with NO trailing newline.
pub fn passthrough(data: &[u8]) {
    let _ = std::io::stdout().write_all(data);
}

// ── field accessors ────────────────────────────────────────────────────────

/// Extract the `tool_input` object from a parsed hook JSON value.
pub fn get_tool_input(v: &Value) -> Option<&serde_json::Map<String, Value>> {
    v.get("tool_input").and_then(|t| t.as_object())
}

/// Extract a string field from `tool_input`, returning `""` if absent or wrong type.
pub fn get_tool_input_string<'a>(v: &'a Value, field: &str) -> &'a str {
    get_tool_input(v)
        .and_then(|m| m.get(field))
        .and_then(|f| f.as_str())
        .unwrap_or("")
}

/// Extract the `tool_output` object from a parsed hook JSON value.
pub fn get_tool_output(v: &Value) -> Option<&serde_json::Map<String, Value>> {
    v.get("tool_output").and_then(|t| t.as_object())
}

/// Extract a string field from `tool_output`, returning `""` if absent or wrong type.
pub fn get_tool_output_string<'a>(v: &'a Value, field: &str) -> &'a str {
    get_tool_output(v)
        .and_then(|m| m.get(field))
        .and_then(|f| f.as_str())
        .unwrap_or("")
}

// ── tests ──────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    // ── get_tool_input_string ─────────────────────────────────────────────

    /// Happy path: fields present and are strings.
    /// Ports TestGetToolInputString (file_path + command assertions).
    #[test]
    fn test_get_tool_input_string_happy_path() {
        let v = json!({
            "tool_input": {
                "file_path": "/src/main.ts",
                "command": "npm run dev"
            }
        });
        assert_eq!(get_tool_input_string(&v, "file_path"), "/src/main.ts");
        assert_eq!(get_tool_input_string(&v, "command"), "npm run dev");
    }

    /// Missing field inside tool_input returns "".
    /// Ports TestGetToolInputString (missing field assertion).
    #[test]
    fn test_get_tool_input_string_missing_field() {
        let v = json!({
            "tool_input": {
                "file_path": "/src/main.ts"
            }
        });
        assert_eq!(get_tool_input_string(&v, "missing"), "");
    }

    /// Wrong type (not a string) returns "".
    #[test]
    fn test_get_tool_input_string_wrong_type() {
        let v = json!({
            "tool_input": {
                "count": 42
            }
        });
        assert_eq!(get_tool_input_string(&v, "count"), "");
    }

    /// No `tool_input` key in the object returns "".
    /// Ports TestGetToolInputString_NoToolInput.
    #[test]
    fn test_get_tool_input_string_no_tool_input() {
        let v = json!({ "other": "data" });
        assert_eq!(get_tool_input_string(&v, "file_path"), "");
    }

    /// Null `tool_input` field returns "".
    #[test]
    fn test_get_tool_input_string_null_field() {
        let v = json!({
            "tool_input": {
                "file_path": null
            }
        });
        assert_eq!(get_tool_input_string(&v, "file_path"), "");
    }

    // ── get_tool_input (object) ───────────────────────────────────────────

    /// `tool_input` present returns Some.
    #[test]
    fn test_get_tool_input_present() {
        let v = json!({ "tool_input": { "k": "v" } });
        assert!(get_tool_input(&v).is_some());
    }

    /// Missing `tool_input` returns None.
    /// Ports TestGetToolInput_NilMap (nil map → None).
    #[test]
    fn test_get_tool_input_missing() {
        let v = json!({});
        assert!(get_tool_input(&v).is_none());
    }

    /// `tool_input` is not an object (e.g. a string) returns None.
    #[test]
    fn test_get_tool_input_not_object() {
        let v = json!({ "tool_input": "not-an-object" });
        assert!(get_tool_input(&v).is_none());
    }

    // ── get_tool_output_string ────────────────────────────────────────────

    /// Happy path: field present and is a string.
    /// Ports TestGetToolOutputString.
    #[test]
    fn test_get_tool_output_string_happy_path() {
        let v = json!({
            "tool_output": {
                "output": "https://github.com/owner/repo/pull/42"
            }
        });
        assert_eq!(
            get_tool_output_string(&v, "output"),
            "https://github.com/owner/repo/pull/42"
        );
    }

    /// Missing field inside tool_output returns "".
    #[test]
    fn test_get_tool_output_string_missing_field() {
        let v = json!({ "tool_output": {} });
        assert_eq!(get_tool_output_string(&v, "output"), "");
    }

    /// No `tool_output` key returns "".
    #[test]
    fn test_get_tool_output_string_no_tool_output() {
        let v = json!({ "other": "data" });
        assert_eq!(get_tool_output_string(&v, "output"), "");
    }

    /// Wrong type (not a string) returns "".
    #[test]
    fn test_get_tool_output_string_wrong_type() {
        let v = json!({ "tool_output": { "count": 99 } });
        assert_eq!(get_tool_output_string(&v, "count"), "");
    }

    /// Null field returns "".
    #[test]
    fn test_get_tool_output_string_null_field() {
        let v = json!({ "tool_output": { "output": null } });
        assert_eq!(get_tool_output_string(&v, "output"), "");
    }
}
