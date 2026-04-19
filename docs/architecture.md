# Architecture

Claudey is a Claude Code plugin built in two layers: plain Markdown and JSON assets that Claude Code loads directly (skills, slash commands, rules, hook wiring), and a single Rust binary (`bin/claudey`) that every hook in `hooks/hooks.json` dispatches to via subcommands. There is no runtime framework — the binary is just a CLI that reads a JSON payload on stdin, does its work, and exits.

## Layout

```
claudey/
├── .claude-plugin/      # Plugin manifest (plugin.json)
├── bin/                 # Build and test scripts; bin/claudey is built here (gitignored)
├── commands/            # Slash command definitions (Markdown)
├── hooks/               # hooks.json — maps Claude Code events to bin/claudey subcommands
├── rules/               # Always-on rule files, split by language scope
├── skills/              # Skill directories, each with SKILL.md
├── src/                 # Rust source for the claudey binary
├── tests/               # Integration tests for the binary
├── Cargo.toml           # Rust crate definition
├── install.sh           # Rules installer (Claude + Cursor targets)
└── README.md            # Landing page; links into docs/
```

## Plugin Manifest

`.claude-plugin/plugin.json` declares the plugin to Claude Code:

- `name`: `claudey`
- `version`: `2.0.0`
- `skills`: `["./skills/", "./commands/"]` — both directories are discovered as skills; slash commands are skills that declare a `command:` field.
- `agents`: lists 10 agent Markdown files under `./agents/*.md`.

**Known gap:** the `agents/` directory does not currently exist in the repo, so the agent entries in `plugin.json` point at files that will need to be added before the manifest's agent list is satisfiable. Treat this as a pending follow-up, not live functionality.

## Runtime Components

### Markdown/JSON Assets

Skills, slash commands, rules, and the hook wiring (`hooks/hooks.json`) are plain files that Claude Code reads at session start. There is no compilation or registration step — adding a new skill or command is a matter of dropping a file in the right directory and committing. See [`./skills.md`](./skills.md), [`./commands.md`](./commands.md), [`./rules.md`](./rules.md), and [`./hooks.md`](./hooks.md) for per-subsystem catalogues.

### Rust Binary

Every hook delegates to `${CLAUDE_PLUGIN_ROOT}/bin/claudey <subcommand>`. The binary is a single-purpose CLI: it parses a hook payload from stdin, runs the relevant subcommand, and writes back through stderr (for user-visible logs) or stdout (for passthrough). Source lives under `src/` with one file per subcommand in `src/hooks/`. See [`./binary.md`](./binary.md).

## Hook Flow

A typical invocation:

```
Claude Code event (e.g. PostToolUse on Edit)
        │
        ▼
hooks/hooks.json  ── matcher matches "Edit" ──▶  "${CLAUDE_PLUGIN_ROOT}/bin/claudey post-edit-format"
        │
        ▼
bin/claudey                                     stdin: { "tool_input": { "file_path": ... }, ... }
  ├─ reads JSON payload (≤ 1 MiB, 5 s timeout)
  ├─ routes to src/hooks/post_edit_format.rs
  ├─ runs prettier against the edited file
  ├─ writes [Hook] log lines to stderr
  ├─ passes raw input through to stdout (unchanged)
  └─ exits with status code (0 = continue, non-zero = signal to Claude Code)
```

Non-zero exit codes have meaning: `block-random-docs` returns `2` to signal a blocked write, while other hooks always exit `0` and use stderr to convey information.

## Cross-References

- [Rust Binary](./binary.md) — `bin/claudey` subcommand reference and module layout.
- [Hooks](./hooks.md) — every wiring in `hooks/hooks.json` with behaviour notes.
- [Skills](./skills.md) — catalogue of everything under `skills/`.
- [Slash Commands](./commands.md) — catalogue of everything under `commands/`.
- [Rules](./rules.md) — always-on rules split by language scope.
