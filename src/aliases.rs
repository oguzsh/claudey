//! aliases — persistent session-alias registry stored at `~/.claude/session-aliases.json`.
//!
//! Ports internal/aliases/aliases.go. Public API mirrors the Go package; the
//! `*_at` variants let tests operate on a local file path instead of mutating
//! `$HOME`, which keeps tests parallel-safe.

#![allow(dead_code)]

use crate::hookio;
use crate::platform;
use chrono::Utc;
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

pub const ALIAS_VERSION: &str = "1.0";

const RESERVED_NAMES: &[&str] = &["list", "help", "remove", "delete", "create", "set"];

// ── data ─────────────────────────────────────────────────────────────────

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct AliasEntry {
    #[serde(rename = "sessionPath", default)]
    pub session_path: String,
    #[serde(rename = "createdAt", default)]
    pub created_at: String,
    #[serde(
        rename = "updatedAt",
        default,
        skip_serializing_if = "String::is_empty"
    )]
    pub updated_at: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub title: String,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct AliasMetadata {
    #[serde(rename = "totalCount", default)]
    pub total_count: usize,
    #[serde(rename = "lastUpdated", default)]
    pub last_updated: String,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct AliasFile {
    #[serde(default)]
    pub version: String,
    #[serde(default)]
    pub aliases: BTreeMap<String, AliasEntry>,
    #[serde(default)]
    pub metadata: AliasMetadata,
}

pub fn default_file() -> AliasFile {
    AliasFile {
        version: ALIAS_VERSION.to_string(),
        aliases: BTreeMap::new(),
        metadata: AliasMetadata {
            total_count: 0,
            last_updated: Utc::now().to_rfc3339(),
        },
    }
}

/// Default on-disk location for the aliases file.
pub fn aliases_path() -> PathBuf {
    platform::claude_dir().join("session-aliases.json")
}

// ── load / save ──────────────────────────────────────────────────────────

pub fn load() -> AliasFile {
    load_from(&aliases_path())
}

pub fn load_from(path: &Path) -> AliasFile {
    let content = match std::fs::read_to_string(path) {
        Ok(c) => c,
        Err(_) => return default_file(),
    };
    let mut f: AliasFile = match serde_json::from_str(&content) {
        Ok(f) => f,
        Err(e) => {
            hookio::log(&format!("[Aliases] Error parsing aliases file: {e}"));
            return default_file();
        }
    };
    if f.version.is_empty() {
        f.version = ALIAS_VERSION.to_string();
    }
    f
}

pub fn save(f: &mut AliasFile) -> bool {
    save_to(&aliases_path(), f)
}

pub fn save_to(path: &Path, f: &mut AliasFile) -> bool {
    f.metadata.total_count = f.aliases.len();
    f.metadata.last_updated = Utc::now().to_rfc3339();

    let data = match serde_json::to_vec_pretty(f) {
        Ok(d) => d,
        Err(e) => {
            hookio::log(&format!("[Aliases] Error marshaling aliases: {e}"));
            return false;
        }
    };

    if let Some(dir) = path.parent() {
        if let Err(e) = platform::ensure_dir(dir) {
            hookio::log(&format!("[Aliases] Error creating directory: {e}"));
            return false;
        }
    }

    let tmp = with_extra_ext(path, "tmp");
    let bak = with_extra_ext(path, "bak");

    if path.exists() {
        let _ = std::fs::copy(path, &bak);
    }

    if let Err(e) = std::fs::write(&tmp, &data) {
        hookio::log(&format!("[Aliases] Error writing temp file: {e}"));
        restore_backup(&bak, path);
        return false;
    }

    // On Windows rename-over-existing fails; remove first.
    #[cfg(windows)]
    if path.exists() {
        let _ = std::fs::remove_file(path);
    }

    if let Err(e) = std::fs::rename(&tmp, path) {
        hookio::log(&format!("[Aliases] Error renaming temp file: {e}"));
        restore_backup(&bak, path);
        let _ = std::fs::remove_file(&tmp);
        return false;
    }

    let _ = std::fs::remove_file(&bak);
    true
}

fn restore_backup(bak: &Path, dst: &Path) {
    if bak.exists() {
        if let Ok(data) = std::fs::read(bak) {
            let _ = std::fs::write(dst, data);
            hookio::log("[Aliases] Restored from backup");
        }
    }
}

/// Appends an additional extension segment (e.g. `.tmp`) to the full path.
fn with_extra_ext(p: &Path, ext: &str) -> PathBuf {
    let mut s = p.as_os_str().to_owned();
    s.push(".");
    s.push(ext);
    PathBuf::from(s)
}

// ── validation helpers ───────────────────────────────────────────────────

pub fn valid_alias_name(name: &str) -> bool {
    !name.is_empty()
        && name
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || c == '_' || c == '-')
}

pub fn is_reserved(name: &str) -> bool {
    let lower = name.to_lowercase();
    RESERVED_NAMES.iter().any(|r| *r == lower)
}

// ── resolve / set / delete / rename ──────────────────────────────────────

pub fn resolve(alias: &str) -> Option<AliasEntry> {
    resolve_at(&aliases_path(), alias)
}

pub fn resolve_at(path: &Path, alias: &str) -> Option<AliasEntry> {
    if alias.is_empty() || !valid_alias_name(alias) {
        return None;
    }
    load_from(path).aliases.get(alias).cloned()
}

/// Resolve `alias` to a session path, or return the input unchanged.
pub fn resolve_session_alias(alias_or_id: &str) -> String {
    resolve_session_alias_at(&aliases_path(), alias_or_id)
}

pub fn resolve_session_alias_at(path: &Path, alias_or_id: &str) -> String {
    match resolve_at(path, alias_or_id) {
        Some(e) => e.session_path,
        None => alias_or_id.to_string(),
    }
}

#[derive(Debug, Default, Clone)]
pub struct SetResult {
    pub success: bool,
    pub is_new: bool,
    pub alias: String,
    pub session_path: String,
    pub title: String,
    pub error: String,
}

pub fn set(alias: &str, session_path: &str, title: &str) -> SetResult {
    set_at(&aliases_path(), alias, session_path, title)
}

pub fn set_at(path: &Path, alias: &str, session_path: &str, title: &str) -> SetResult {
    if alias.is_empty() {
        return err_result("Alias name cannot be empty");
    }
    if session_path.trim().is_empty() {
        return err_result("Session path cannot be empty");
    }
    if alias.len() > 128 {
        return err_result("Alias name cannot exceed 128 characters");
    }
    if !valid_alias_name(alias) {
        return err_result(
            "Alias name must contain only letters, numbers, dashes, and underscores",
        );
    }
    if is_reserved(alias) {
        return err_result(&format!("'{alias}' is a reserved alias name"));
    }

    let mut f = load_from(path);
    let existing = f.aliases.get(alias).cloned();
    let now = Utc::now().to_rfc3339();

    let entry = AliasEntry {
        session_path: session_path.to_string(),
        updated_at: now.clone(),
        title: title.to_string(),
        created_at: existing
            .as_ref()
            .map(|e| e.created_at.clone())
            .unwrap_or(now),
    };

    let is_new = existing.is_none();
    f.aliases.insert(alias.to_string(), entry);

    if !save_to(path, &mut f) {
        return err_result("Failed to save alias");
    }

    SetResult {
        success: true,
        is_new,
        alias: alias.to_string(),
        session_path: session_path.to_string(),
        title: title.to_string(),
        error: String::new(),
    }
}

#[derive(Debug, Default, Clone)]
pub struct DeleteResult {
    pub success: bool,
    pub alias: String,
    pub deleted_session_path: String,
    pub error: String,
}

pub fn delete(alias: &str) -> DeleteResult {
    delete_at(&aliases_path(), alias)
}

pub fn delete_at(path: &Path, alias: &str) -> DeleteResult {
    let mut f = load_from(path);
    let entry = match f.aliases.remove(alias) {
        Some(e) => e,
        None => {
            return DeleteResult {
                error: format!("Alias '{alias}' not found"),
                ..Default::default()
            };
        }
    };
    if !save_to(path, &mut f) {
        return DeleteResult {
            error: "Failed to delete alias".to_string(),
            ..Default::default()
        };
    }
    DeleteResult {
        success: true,
        alias: alias.to_string(),
        deleted_session_path: entry.session_path,
        error: String::new(),
    }
}

pub fn rename(old_alias: &str, new_alias: &str) -> SetResult {
    rename_at(&aliases_path(), old_alias, new_alias)
}

pub fn rename_at(path: &Path, old_alias: &str, new_alias: &str) -> SetResult {
    if new_alias.is_empty() {
        return err_result("New alias name cannot be empty");
    }
    if new_alias.len() > 128 {
        return err_result("New alias name cannot exceed 128 characters");
    }
    if !valid_alias_name(new_alias) {
        return err_result(
            "New alias name must contain only letters, numbers, dashes, and underscores",
        );
    }
    if is_reserved(new_alias) {
        return err_result(&format!("'{new_alias}' is a reserved alias name"));
    }

    let mut f = load_from(path);
    let mut entry = match f.aliases.remove(old_alias) {
        Some(e) => e,
        None => return err_result(&format!("Alias '{old_alias}' not found")),
    };
    if f.aliases.contains_key(new_alias) {
        // Put it back.
        f.aliases.insert(old_alias.to_string(), entry);
        return err_result(&format!("Alias '{new_alias}' already exists"));
    }
    entry.updated_at = Utc::now().to_rfc3339();
    f.aliases.insert(new_alias.to_string(), entry.clone());

    if !save_to(path, &mut f) {
        // Attempt rollback.
        f.aliases.remove(new_alias);
        f.aliases.insert(old_alias.to_string(), entry);
        let _ = save_to(path, &mut f);
        return err_result("Failed to save renamed alias — rolled back to original");
    }

    SetResult {
        success: true,
        alias: new_alias.to_string(),
        session_path: entry.session_path,
        ..Default::default()
    }
}

fn err_result(msg: &str) -> SetResult {
    SetResult {
        error: msg.to_string(),
        ..Default::default()
    }
}

// ── list / for_session / cleanup ─────────────────────────────────────────

#[derive(Debug, Clone, Default)]
pub struct AliasInfo {
    pub name: String,
    pub session_path: String,
    pub created_at: String,
    pub updated_at: String,
    pub title: String,
}

pub fn list(search: Option<&str>, limit: usize) -> Vec<AliasInfo> {
    list_at(&aliases_path(), search, limit)
}

pub fn list_at(path: &Path, search: Option<&str>, limit: usize) -> Vec<AliasInfo> {
    let f = load_from(path);
    let mut aliases: Vec<AliasInfo> = f
        .aliases
        .into_iter()
        .map(|(name, e)| AliasInfo {
            name,
            session_path: e.session_path,
            created_at: e.created_at,
            updated_at: e.updated_at,
            title: e.title,
        })
        .collect();

    aliases.sort_by(|a, b| {
        let ka = if a.updated_at.is_empty() {
            &a.created_at
        } else {
            &a.updated_at
        };
        let kb = if b.updated_at.is_empty() {
            &b.created_at
        } else {
            &b.updated_at
        };
        kb.cmp(ka) // newest first
    });

    if let Some(q) = search {
        if !q.is_empty() {
            let needle = q.to_lowercase();
            aliases.retain(|a| {
                a.name.to_lowercase().contains(&needle) || a.title.to_lowercase().contains(&needle)
            });
        }
    }

    if limit > 0 && aliases.len() > limit {
        aliases.truncate(limit);
    }

    aliases
}

pub fn for_session(session_path: &str) -> Vec<AliasInfo> {
    for_session_at(&aliases_path(), session_path)
}

pub fn for_session_at(path: &Path, session_path: &str) -> Vec<AliasInfo> {
    let f = load_from(path);
    f.aliases
        .into_iter()
        .filter(|(_, e)| e.session_path == session_path)
        .map(|(name, e)| AliasInfo {
            name,
            created_at: e.created_at,
            title: e.title,
            ..Default::default()
        })
        .collect()
}

pub fn cleanup<F: Fn(&str) -> bool>(session_exists: F) -> (usize, usize) {
    cleanup_at(&aliases_path(), session_exists)
}

pub fn cleanup_at<F: Fn(&str) -> bool>(path: &Path, session_exists: F) -> (usize, usize) {
    let mut f = load_from(path);
    let to_remove: Vec<String> = f
        .aliases
        .iter()
        .filter(|(_, e)| !session_exists(&e.session_path))
        .map(|(k, _)| k.clone())
        .collect();

    let checked = f.aliases.len();
    for name in &to_remove {
        f.aliases.remove(name);
    }
    if !to_remove.is_empty() {
        save_to(path, &mut f);
    }
    (checked, to_remove.len())
}

// ── tests ────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use crate::testutil::TempDir;

    fn tmp_alias_path() -> (TempDir, PathBuf) {
        let d = TempDir::new();
        let p = d.path().join("session-aliases.json");
        (d, p)
    }

    #[test]
    fn test_load_default_when_missing() {
        let (_d, p) = tmp_alias_path();
        let f = load_from(&p);
        assert_eq!(f.version, ALIAS_VERSION);
        assert!(f.aliases.is_empty());
    }

    #[test]
    fn test_set_and_resolve() {
        let (_d, p) = tmp_alias_path();
        let r = set_at(&p, "myalias", "/path/to/session", "My Session");
        assert!(r.success, "set failed: {}", r.error);
        assert!(r.is_new);

        let e = resolve_at(&p, "myalias").expect("resolve should find it");
        assert_eq!(e.session_path, "/path/to/session");
        assert_eq!(e.title, "My Session");
    }

    #[test]
    fn test_set_validation() {
        let (_d, p) = tmp_alias_path();
        let cases = [
            ("", "/path", "empty alias"),
            ("a", "", "empty path"),
            ("invalid!name", "/path", "unsafe chars"),
            ("list", "/path", "reserved word"),
        ];
        for (alias, sp, why) in cases {
            let r = set_at(&p, alias, sp, "");
            assert!(
                !r.success,
                "set_at({alias:?}, {sp:?}) should have failed ({why})"
            );
        }
    }

    #[test]
    fn test_set_alias_too_long_rejected() {
        let (_d, p) = tmp_alias_path();
        let long = "a".repeat(129);
        let r = set_at(&p, &long, "/x", "");
        assert!(!r.success);
    }

    #[test]
    fn test_delete_removes_alias() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "todelete", "/path", "");
        let r = delete_at(&p, "todelete");
        assert!(r.success, "delete failed: {}", r.error);
        assert!(resolve_at(&p, "todelete").is_none());
    }

    #[test]
    fn test_delete_nonexistent_fails() {
        let (_d, p) = tmp_alias_path();
        let r = delete_at(&p, "ghost");
        assert!(!r.success);
    }

    #[test]
    fn test_list_search_and_limit() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "alpha", "/path/a", "Alpha");
        set_at(&p, "beta", "/path/b", "Beta");

        assert_eq!(list_at(&p, None, 0).len(), 2);
        assert_eq!(list_at(&p, Some("alp"), 0).len(), 1);
        assert_eq!(list_at(&p, None, 1).len(), 1);
    }

    #[test]
    fn test_list_sorted_newest_first() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "first", "/a", "");
        std::thread::sleep(std::time::Duration::from_millis(5));
        set_at(&p, "second", "/b", "");
        let items = list_at(&p, None, 0);
        assert_eq!(items.len(), 2);
        assert_eq!(items[0].name, "second");
        assert_eq!(items[1].name, "first");
    }

    #[test]
    fn test_rename_moves_alias() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "oldname", "/path", "Title");
        let r = rename_at(&p, "oldname", "newname");
        assert!(r.success, "rename failed: {}", r.error);
        assert!(resolve_at(&p, "oldname").is_none());
        assert!(resolve_at(&p, "newname").is_some());
    }

    #[test]
    fn test_rename_collision_rejected() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "a", "/p1", "");
        set_at(&p, "b", "/p2", "");
        let r = rename_at(&p, "a", "b");
        assert!(!r.success, "rename onto existing alias must fail");
        // Original still intact.
        assert!(resolve_at(&p, "a").is_some());
    }

    #[test]
    fn test_resolve_session_alias_passthrough() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "myalias", "/path/to/session", "");
        assert_eq!(resolve_session_alias_at(&p, "myalias"), "/path/to/session");
        assert_eq!(resolve_session_alias_at(&p, "notanalias"), "notanalias");
    }

    #[test]
    fn test_cleanup_removes_dead_sessions() {
        let (_d, p) = tmp_alias_path();
        set_at(&p, "exists", "/exists", "");
        set_at(&p, "gone", "/gone", "");

        let (checked, removed) = cleanup_at(&p, |sp| sp == "/exists");
        assert_eq!(checked, 2);
        assert_eq!(removed, 1);
        assert!(resolve_at(&p, "gone").is_none());
        assert!(resolve_at(&p, "exists").is_some());
    }

    #[test]
    fn test_valid_alias_name() {
        assert!(valid_alias_name("foo"));
        assert!(valid_alias_name("Foo_bar-99"));
        assert!(!valid_alias_name(""));
        assert!(!valid_alias_name("no spaces"));
        assert!(!valid_alias_name("bang!"));
    }

    #[test]
    fn test_is_reserved_is_case_insensitive() {
        assert!(is_reserved("list"));
        assert!(is_reserved("LIST"));
        assert!(!is_reserved("alpha"));
    }

    #[test]
    fn test_save_writes_pretty_json() {
        let (_d, p) = tmp_alias_path();
        let mut f = default_file();
        f.aliases.insert(
            "x".to_string(),
            AliasEntry {
                session_path: "/s".to_string(),
                created_at: "2026-04-18T00:00:00Z".to_string(),
                updated_at: "2026-04-18T00:00:00Z".to_string(),
                title: "t".to_string(),
            },
        );
        assert!(save_to(&p, &mut f));
        let loaded = load_from(&p);
        assert_eq!(loaded.aliases.len(), 1);
        assert_eq!(loaded.aliases.get("x").unwrap().session_path, "/s");
    }
}
