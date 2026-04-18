//! hooks — subcommand implementations dispatched by `main.rs`.
//!
//! Every public fn here mirrors a Go hook from `internal/hooks/*.go`.
//! Implementations start as stubs and will be fleshed out task-by-task.

#![allow(dead_code)]

use serde_json::Value;

mod post_edit_format;
mod post_edit_typecheck;
mod pre_compact;
mod session_end;
mod session_start;
mod suggest_compact;
pub use post_edit_format::post_edit_format;
pub use post_edit_typecheck::post_edit_typecheck;
pub use pre_compact::pre_compact;
pub use session_end::session_end;
pub use session_start::session_start;
pub use suggest_compact::suggest_compact;

pub fn post_edit_console_warn(_input: Value, _raw: Vec<u8>) {
    eprintln!("post-edit-console-warn stub");
}

pub fn check_console_log(_raw: Vec<u8>) {
    eprintln!("check-console-log stub");
}

pub fn evaluate_session(_input: Value, _plugin_root: String) {
    eprintln!("evaluate-session stub");
}

pub fn git_push_reminder(_input: Value, _raw: Vec<u8>) {
    eprintln!("git-push-reminder stub");
}

pub fn block_random_docs(_input: Value, _raw: Vec<u8>) -> i32 {
    eprintln!("block-random-docs stub");
    0
}

pub fn pr_created_log(_input: Value, _raw: Vec<u8>) {
    eprintln!("pr-created-log stub");
}
