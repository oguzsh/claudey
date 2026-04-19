# Rust Binary (bin/claudey)

A single Rust binary dispatches every hook in the plugin via subcommands. Source lives under `src/` and builds with a stable Rust toolchain; the compiled binary is not checked into git — each machine builds its own.

## Build & Install

```bash
bash bin/build-hooks.sh    # cargo build --release + install -m 0755 target/release/claudey bin/claudey
bash bin/test.sh           # cargo fmt --check + cargo clippy --all-targets -- -D warnings + cargo test
```

`bin/claudey` and `bin/claudey.exe` are gitignored. Run `bin/build-hooks.sh` once on each machine after cloning the repo.

## Subcommands

Invoked as `${CLAUDE_PLUGIN_ROOT}/bin/claudey <subcommand>` from `hooks/hooks.json`.

| Subcommand               | Purpose                                                                | Source                                      |
| ------------------------ | ---------------------------------------------------------------------- | ------------------------------------------- |
| `session-start`          | Surface the previous session's summary and learned-skill count         | `src/hooks/session_start.rs`                |
| `session-end`            | Write/update the session tmp file from the JSONL transcript            | `src/hooks/session_end.rs`                  |
| `pre-compact`            | Append a compaction event entry before context compaction              | `src/hooks/pre_compact.rs`                  |
| `suggest-compact`        | Warn at tool-count thresholds to prompt manual `/compact`              | `src/hooks/suggest_compact.rs`              |
| `post-edit-format`       | Run `prettier --write` on edited JS/TS files                           | `src/hooks/post_edit_format.rs`             |
| `post-edit-typecheck`    | Run `tsc --noEmit` after editing `.ts` / `.tsx` files                  | `src/hooks/post_edit_typecheck.rs`          |
| `post-edit-console-warn` | Warn when `console.log` is present in the just-edited JS/TS file       | `src/hooks/post_edit_console_warn.rs`       |
| `check-console-log`      | Scan modified files (excluding tests/scripts/mocks) for `console.log`  | `src/hooks/check_console_log.rs`            |
| `evaluate-session`       | Evaluate whether the session is long enough to extract learnings       | `src/hooks/evaluate_session.rs`             |
| `git-push-reminder`      | Log a reminder when a `git push` Bash command is about to run          | `src/hooks/inline.rs` (`git_push_reminder`) |
| `block-random-docs`      | Return exit code `2` to block random `.md`/`.txt` writes               | `src/hooks/inline.rs` (`block_random_docs`) |
| `pr-created-log`         | Capture the PR URL from `gh pr create` output and log a review command | `src/hooks/inline.rs` (`pr_created_log`)    |

## Module Layout

| Module         | Purpose                                                  |
| -------------- | -------------------------------------------------------- |
| `main.rs`      | Subcommand dispatcher + plugin-root discovery            |
| `hookio.rs`    | Stdin JSON read, stderr logging, stdout passthrough      |
| `platform.rs`  | OS flags, home/claude/sessions/learned-skills directories |
| `datetime.rs`  | Local date/time/datetime string formatting               |
| `sysutil.rs`   | Command existence check, shell runner, `npx` binary path |
| `fileutil.rs`  | Read/write/append/find/grep/replace helpers              |
| `gitutil.rs`   | `is_repo`, `repo_name`, `project_name`, `modified_files` |
| `aliases.rs`   | Session-alias registry load/save/list                    |
| `testutil.rs`  | `TempDir` RAII helper for tests                          |
| `hooks/mod.rs` | Re-exports each hook's entry point                       |
| `hooks/*.rs`   | One file per subcommand (see table above)                |

## Plugin Root Discovery

`src/main.rs:find_plugin_root` resolves `${CLAUDE_PLUGIN_ROOT}` with a three-step fallback:

1. `$CLAUDE_PLUGIN_ROOT` environment variable, if set and non-empty.
2. `exe.parent().parent()` — walk up from the binary itself. If that directory contains `hooks/hooks.json`, use it.
3. Current working directory, walked upward up to 10 levels, looking for the first ancestor that contains `hooks/hooks.json`.

If all three fail, the current directory is returned as a last-resort default.

## Hook I/O Contract

- **Input**: hook JSON payload on stdin (up to 1 MiB, 5 s timeout). Missing or malformed JSON is treated as an empty object rather than an error.
- **Output**: stdout passes the raw input through unchanged unless the hook explicitly writes a replacement.
- **Logs**: go to stderr, typically with a `[Hook] ` prefix. These are the messages the user sees during a Claude Code session.
- **Exit code**: `0` means "continue normally". A non-zero code (e.g. `block-random-docs` returns `2`) signals to Claude Code that the tool call should be blocked.

## First-Run Setup

The plugin ships as Rust source, not a prebuilt binary, so `bin/claudey` has to be built per-machine. Two shell scripts mediate this:

- **`bin/setup.sh`** — run manually in your terminal after installing the plugin. If `cargo` is missing, prompts to install Rust via `rustup` (one-line Y/n). Then delegates to `bin/build-hooks.sh` to build and install the binary. Run once per machine.
- **`bin/session-start-guard.sh`** — invoked automatically at the start of every Claude Code session, wrapped around `bin/claudey session-start` in `hooks/hooks.json`. If the binary is missing, prints a one-line nudge to run `bin/setup.sh` and exits (the SessionStart hook fails visibly; Claude itself keeps running). If the binary exists but `src/` / `Cargo.toml` / `Cargo.lock` are newer, silently rebuilds via `bin/build-hooks.sh` and streams cargo's output to stderr.

The binary's own existence is the "setup done" marker — no separate flag file. Manual rebuilds still work: `bash bin/build-hooks.sh` any time.

**Platform support:** macOS and Linux/WSL. Windows native is not supported.
