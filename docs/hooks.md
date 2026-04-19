# Hooks

Hooks are wired in `hooks/hooks.json`. Most entries delegate to `${CLAUDE_PLUGIN_ROOT}/bin/claudey <subcommand>` — see [`./binary.md`](./binary.md) for the subcommand catalogue and module layout. A handful of entries shell out to scripts (the `continuous-learning` skill's observer) or macOS system commands (notification sounds).

## Hook Events

Ordered as they appear in `hooks/hooks.json`.

| Event          | Matcher       | Command                                                 | Blocking?                 |
| -------------- | ------------- | ------------------------------------------------------- | ------------------------- |
| `PreToolUse`   | `Bash`        | `bin/claudey git-push-reminder`                         | No (informational)        |
| `PreToolUse`   | `Write`       | `bin/claudey block-random-docs`                         | **Yes** (exit `2` blocks) |
| `PreToolUse`   | `Edit\|Write` | `bin/claudey suggest-compact`                           | No                        |
| `PreToolUse`   | `*`           | `skills/continuous-learning/hooks/observe.sh pre`       | No                        |
| `PreCompact`   | `*`           | `bin/claudey pre-compact`                               | No                        |
| `SessionStart` | `*`           | `bin/claudey session-start`                             | No                        |
| `PostToolUse`  | `Bash`        | `bin/claudey pr-created-log`                            | No                        |
| `PostToolUse`  | `Edit`        | `bin/claudey post-edit-format`                          | No                        |
| `PostToolUse`  | `Edit`        | `bin/claudey post-edit-typecheck`                       | No                        |
| `PostToolUse`  | `Edit`        | `bin/claudey post-edit-console-warn`                    | No                        |
| `PostToolUse`  | `*`           | `skills/continuous-learning/hooks/observe.sh post`      | No                        |
| `Notification` | `*`           | `afplay /System/Library/Sounds/Glass.aiff` (macOS only) | No                        |
| `Stop`         | `*`           | `bin/claudey check-console-log`                         | No                        |
| `Stop`         | `*`           | `afplay /System/Library/Sounds/Hero.aiff` (macOS only)  | No                        |
| `SessionEnd`   | `*`           | `bin/claudey session-end`                               | No                        |
| `SessionEnd`   | `*`           | `bin/claudey evaluate-session`                          | No                        |

## Per-Subcommand Behaviour

### session-start

Reads the most recent session file from `~/.claude/sessions/*-session.tmp`, surfaces the stored summary (if any), and lists active alias entries and learned skills.

- Source: `src/hooks/session_start.rs`

### session-end

Parses the JSONL transcript referenced by the hook payload (or `CLAUDE_TRANSCRIPT_PATH`), extracts user messages, tool calls, and file paths, and writes/updates `~/.claude/sessions/{date}-{short-id}-session.tmp`.

- Source: `src/hooks/session_end.rs`

### pre-compact

Appends a compaction event line to `~/.claude/sessions/compaction-log.txt` and, if a latest session file exists, notes the compaction in it.

- Source: `src/hooks/pre_compact.rs`

### suggest-compact

Increments a per-session tool counter in the system temp directory and emits a suggestion to run `/compact` when the counter crosses a multiple of the threshold (default 50, configurable via `COMPACT_THRESHOLD`).

- Source: `src/hooks/suggest_compact.rs`

### post-edit-format

Runs `npx prettier --write <file>` for edited `.ts`, `.tsx`, `.js`, or `.jsx` files, with the current directory set to the file's parent so local Prettier config is picked up.

- Source: `src/hooks/post_edit_format.rs`

### post-edit-typecheck

For edited `.ts` / `.tsx` files, walks up to the nearest `tsconfig.json` (up to 20 levels) and runs `npx tsc --noEmit --pretty false`. Errors mentioning the edited file are logged back (up to 10 lines).

- Source: `src/hooks/post_edit_typecheck.rs`

### post-edit-console-warn

Scans the just-edited JS/TS file for `console.log` occurrences and logs the first five matches with their 1-indexed line numbers.

- Source: `src/hooks/post_edit_console_warn.rs`

### check-console-log

On session stop, uses `git diff --name-only HEAD` to find modified JS/TS files, excludes tests/specs/configs/mocks/scripts, and warns about any file that still contains `console.log`.

- Source: `src/hooks/check_console_log.rs`

### evaluate-session

Reads `{plugin_root}/skills/continuous-learning/config.json`, counts user messages in the transcript, and logs an evaluation result if the session crossed the minimum length threshold.

- Source: `src/hooks/evaluate_session.rs`

### git-push-reminder

If the `tool_input.command` being run contains `git push`, logs a reminder to review changes before the push lands.

- Source: `src/hooks/inline.rs` (`git_push_reminder`)

### block-random-docs

Inspects the `file_path` of a `Write` tool call. If the target looks like a random `.md` or `.txt` file (not `README.md` / `CLAUDE.md` / `AGENTS.md` / `CONTRIBUTING.md`, and not inside `.claude/plans/`), the hook logs a block message and returns a non-zero exit code.

- Source: `src/hooks/inline.rs` (`block_random_docs`)
- Exit codes: `0` = pass, `2` = block.

### pr-created-log

If a completed Bash call contained `gh pr create`, extracts the PR URL from the output and logs a follow-up `gh pr review` suggestion.

- Source: `src/hooks/inline.rs` (`pr_created_log`)

## Non-Rust Hooks

`skills/continuous-learning/hooks/observe.sh pre` / `observe.sh post` are part of the `continuous-learning` skill's observation layer — they run on every tool call and feed the instinct-learning system. See [`./skills.md`](./skills.md).

The `Notification` and `Stop` entries that invoke `afplay` play a macOS system sound. They are a no-op on Linux and Windows (WSL) where `afplay` is not present.

## Authoring a New Hook

- Add a matcher entry to `hooks/hooks.json`.
- If the hook needs shared utilities, add a new subcommand under `src/hooks/` and register it in both `src/hooks/mod.rs` and the dispatch table in `src/main.rs`.
- Re-build with `bash bin/build-hooks.sh` and verify with `bash bin/test.sh`.
