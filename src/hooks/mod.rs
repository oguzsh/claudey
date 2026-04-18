//! hooks — subcommand implementations dispatched by `main.rs`.
//!
//! Every public fn here mirrors a Go hook from `internal/hooks/*.go`.

#![allow(dead_code)]

mod check_console_log;
mod evaluate_session;
mod inline;
mod post_edit_console_warn;
mod post_edit_format;
mod post_edit_typecheck;
mod pre_compact;
mod session_end;
mod session_start;
mod suggest_compact;
pub use check_console_log::check_console_log;
pub use evaluate_session::evaluate_session;
pub use inline::{block_random_docs, git_push_reminder, pr_created_log};
pub use post_edit_console_warn::post_edit_console_warn;
pub use post_edit_format::post_edit_format;
pub use post_edit_typecheck::post_edit_typecheck;
pub use pre_compact::pre_compact;
pub use session_end::session_end;
pub use session_start::session_start;
pub use suggest_compact::suggest_compact;
