# Claudey Documentation Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a `docs/` folder containing one page per subsystem (architecture, binary, hooks, skills, commands, rules) and rewrite the root `README.md` as a short, link-first overview that routes readers into `docs/`.

**Architecture:** Split documentation by subsystem so each page has a single responsibility and can be maintained independently. Root README becomes a map to the deep-dive docs — it no longer tries to inventory every skill/command/hook inline. `docs/README.md` serves as the table of contents for `docs/`.

**Tech Stack:** Plain Markdown (GitHub-flavored). No build pipeline. All content is hand-written; relative links only so pages remain valid on any git host.

---

## Context

The current `README.md` (committed as `0541d22`) has partially stale counts:
- Claims "24 skills" and "29 slash commands" — the repo actually ships 8 skills and 11 commands (post-migration).
- Claims "18 rule files" — needs verification by counting files under `rules/common`, `rules/python`, `rules/typescript`.
- Has a trimmed-down "Components Reference" section that only lists a subset of skills/commands and mixes them with incomplete tables.

The Go → Rust migration (commits `dbd8a01`, `d22fb96`, `0541d22`) also invalidated most architecture prose. A documentation rewrite is the cleanest way to reset the public surface.

### Key source-of-truth files the writer must read

Every doc page has a canonical source the writer must consult instead of inventing content:

| Doc page             | Canonical sources                                                 |
| -------------------- | ----------------------------------------------------------------- |
| `architecture.md`    | `.claude-plugin/plugin.json`, `hooks/hooks.json`, `Cargo.toml`, `src/main.rs` |
| `binary.md`          | `src/main.rs` (dispatch table), `src/hooks/mod.rs`, `Cargo.toml`  |
| `hooks.md`           | `hooks/hooks.json`, `src/hooks/*.rs` (each file's module doc)     |
| `skills.md`          | Each `skills/<name>/SKILL.md` frontmatter `name`/`description`    |
| `commands.md`        | Each `commands/<name>.md` frontmatter or first heading            |
| `rules.md`           | `rules/README.md`, file listing of `rules/common/`, `rules/python/`, `rules/typescript/` |

---

## Scope Check

One subsystem (documentation). No scope split needed.

---

## File Structure

```
docs/
├── README.md           # Index — links to every other doc page below
├── architecture.md     # System overview: plugin layout + Rust binary + hook flow
├── binary.md           # bin/claudey reference: subcommands, build, module map
├── hooks.md            # Every hook in hooks.json with event, matcher, purpose
├── skills.md           # Every directory under skills/ with purpose
├── commands.md         # Every file under commands/ with purpose
└── rules.md            # rules/common, rules/python, rules/typescript index
```

Plus one file at the repo root:

```
README.md               # Rewritten — short overview + links into docs/
```

**Files that change together:** `docs/README.md` must stay in sync with the presence of sibling pages; a new doc page → an entry added to `docs/README.md` → a link added to root `README.md`. Task 8 is explicit about keeping these in sync.

**Single-responsibility check:** Each doc file covers exactly one subsystem. If a section in one page requires more than a paragraph about another subsystem, link to that subsystem's page instead of duplicating.

---

## Plan-wide Conventions

- Use GitHub-flavored Markdown (tables, fenced code blocks, task lists allowed).
- Use **relative links** between docs (e.g., `./hooks.md`, not absolute URLs).
- Use backticked code for file paths, subcommand names, and hook event names.
- Date format in any "last updated" line: `YYYY-MM-DD` (today: `2026-04-19`).
- Do not document the missing `agents/` directory — `plugin.json` references files that don't exist; note this as a gap in `architecture.md` Task 1, but don't create `docs/agents.md`.
- Commit after each task with a conventional commit: `docs: add <file>` or `docs: rewrite README`.

---

## Task 1: Draft `docs/architecture.md`

**Files:**
- Create: `docs/architecture.md`

- [ ] **Step 1: Define required contents checklist**

`docs/architecture.md` MUST contain:
1. H1: `# Architecture`
2. One-paragraph intro: Claudey is a Claude Code plugin; two layers (Markdown/JSON assets + Rust binary).
3. H2 `## Layout` — an ASCII tree of the top-level directories that actually exist in the repo (verified against `ls -la`). Include: `.claude-plugin/`, `bin/`, `commands/`, `hooks/`, `rules/`, `skills/`, `src/`, `tests/`, `Cargo.toml`, `README.md`, `install.sh`. Do NOT include `agents/` (directory is empty / files not yet written).
4. H2 `## Plugin Manifest` — list what `.claude-plugin/plugin.json` declares: name, version, skills/commands paths, agents list. Flag that the agents list references files that do not exist yet.
5. H2 `## Runtime Components` — two subsections:
   - H3 `### Markdown/JSON Assets` — paragraph explaining that skills, commands, rules, and hook wiring are plain files Claude Code loads directly. Link to `./skills.md`, `./commands.md`, `./rules.md`, `./hooks.md`.
   - H3 `### Rust Binary` — paragraph explaining that all hooks dispatch through `bin/claudey <subcommand>`. Link to `./binary.md`.
6. H2 `## Hook Flow` — prose description of the end-to-end flow: Claude Code event → `hooks/hooks.json` → `${CLAUDE_PLUGIN_ROOT}/bin/claudey <subcommand>` → stdin JSON read → subcommand executes → stdout passthrough / stderr log / exit code. One paragraph, plus a short Mermaid or ASCII sequence diagram if helpful.
7. H2 `## Cross-References` — bulleted list of four links: binary, hooks, skills, commands, rules.

- [ ] **Step 2: Verify sources before writing**

Run these to confirm the tree and manifest content are accurate:
```bash
ls -la
cat .claude-plugin/plugin.json
cat Cargo.toml
```

Write down: actual top-level directory names, `plugin.json` fields, Rust crate name.

- [ ] **Step 3: Write `docs/architecture.md`**

Produce a ~60-100-line markdown file following the checklist. Write the ASCII tree from the `ls -la` output in Step 2. Write the manifest section from the `plugin.json` output.

- [ ] **Step 4: Verify required sections are present**

Run:
```bash
grep -E "^# Architecture$|^## Layout$|^## Plugin Manifest$|^## Runtime Components$|^## Hook Flow$|^## Cross-References$" docs/architecture.md
```
Expected: all 6 headings print.

Also verify no link rot before sibling pages exist — relative links to `./binary.md`, `./hooks.md`, `./skills.md`, `./commands.md`, `./rules.md` are intentional and will be satisfied by later tasks.

- [ ] **Step 5: Commit**

```bash
git add docs/architecture.md
git commit -m "docs: add architecture overview"
```

---

## Task 2: Draft `docs/binary.md`

**Files:**
- Create: `docs/binary.md`

- [ ] **Step 1: Define required contents checklist**

`docs/binary.md` MUST contain:
1. H1: `# Rust Binary (bin/claudey)`
2. Intro paragraph: single binary dispatches all hooks via subcommands. Built from `src/`. Not checked into git — build per machine with `bin/build-hooks.sh`.
3. H2 `## Build & Install` — fenced code block showing the two commands:
   ```bash
   bash bin/build-hooks.sh    # cargo build --release + install to bin/claudey
   bash bin/test.sh           # cargo fmt --check + clippy + cargo test
   ```
4. H2 `## Subcommands` — one table with columns `Subcommand | Purpose | Source` and exactly these 12 rows (in the order below):

   | Subcommand               | Purpose                                                                 | Source                                         |
   | ------------------------ | ----------------------------------------------------------------------- | ---------------------------------------------- |
   | `session-start`          | Surface the previous session's summary and learned-skill count          | `src/hooks/session_start.rs`                   |
   | `session-end`            | Write/update the session tmp file from the JSONL transcript             | `src/hooks/session_end.rs`                     |
   | `pre-compact`            | Append a compaction event entry before context compaction               | `src/hooks/pre_compact.rs`                     |
   | `suggest-compact`        | Warn at tool-count thresholds to prompt manual `/compact`               | `src/hooks/suggest_compact.rs`                 |
   | `post-edit-format`       | Run `prettier --write` on edited JS/TS files                            | `src/hooks/post_edit_format.rs`                |
   | `post-edit-typecheck`    | Run `tsc --noEmit` after editing `.ts` / `.tsx` files                   | `src/hooks/post_edit_typecheck.rs`             |
   | `post-edit-console-warn` | Warn when `console.log` is present in the just-edited JS/TS file        | `src/hooks/post_edit_console_warn.rs`          |
   | `check-console-log`      | Scan modified files (excluding tests/scripts/mocks) for `console.log`   | `src/hooks/check_console_log.rs`               |
   | `evaluate-session`       | Evaluate whether the session is long enough to extract learnings        | `src/hooks/evaluate_session.rs`                |
   | `git-push-reminder`      | Log a reminder when a `git push` Bash command is about to run           | `src/hooks/inline.rs` (`git_push_reminder`)    |
   | `block-random-docs`      | Return exit code `2` to block random `.md`/`.txt` writes                | `src/hooks/inline.rs` (`block_random_docs`)    |
   | `pr-created-log`         | Capture the PR URL from `gh pr create` output and log a review command  | `src/hooks/inline.rs` (`pr_created_log`)       |

5. H2 `## Module Layout` — table with columns `Module | Purpose` covering all files under `src/` (non-test `.rs` files at the crate root, plus `src/hooks/`):

   | Module            | Purpose                                             |
   | ----------------- | --------------------------------------------------- |
   | `main.rs`         | Subcommand dispatcher + plugin-root discovery       |
   | `hookio.rs`       | Stdin JSON read, stderr logging, stdout passthrough |
   | `platform.rs`     | OS flags, home/claude/sessions/learned-skills dirs  |
   | `datetime.rs`     | Local date/time/datetime string formatting          |
   | `sysutil.rs`      | Command existence check, shell runner, `npx` binary |
   | `fileutil.rs`     | Read/write/append/find/grep/replace                 |
   | `gitutil.rs`      | `is_repo`, `repo_name`, `project_name`, `modified_files` |
   | `aliases.rs`      | Session-alias registry load/save/list               |
   | `testutil.rs`     | `TempDir` RAII helper for tests                     |
   | `hooks/mod.rs`    | Re-exports each hook's entry point                  |
   | `hooks/*.rs`      | One file per subcommand (see table above)           |

6. H2 `## Plugin Root Discovery` — describe the three-step fallback: `$CLAUDE_PLUGIN_ROOT` env → `exe.parent().parent()` if it contains `hooks/hooks.json` → cwd walk up to 10 levels looking for the same. Cite `src/main.rs:find_plugin_root`.
7. H2 `## Hook I/O Contract` — 4-bullet list:
   - Input: hook JSON payload on stdin (up to 1 MiB, 5 s timeout).
   - Output: stdout passes raw input through unchanged unless the hook explicitly writes a replacement.
   - Logs: go to stderr with `[Hook] `-style prefixes.
   - Exit code: non-zero (e.g., `2`) signals "block the tool call"; `0` means "continue".

- [ ] **Step 2: Verify sources**

```bash
grep -n 'match subcmd' src/main.rs
ls src/ src/hooks/
```
Confirm the subcommand names in the dispatch match the plan's table verbatim.

- [ ] **Step 3: Write `docs/binary.md`**

- [ ] **Step 4: Verify subcommand table completeness**

```bash
for sub in session-start session-end pre-compact suggest-compact \
           post-edit-format post-edit-typecheck post-edit-console-warn \
           check-console-log evaluate-session git-push-reminder \
           block-random-docs pr-created-log; do
  grep -q "\`$sub\`" docs/binary.md || echo "MISSING: $sub"
done
```
Expected: no "MISSING" lines.

- [ ] **Step 5: Commit**

```bash
git add docs/binary.md
git commit -m "docs: add rust binary reference"
```

---

## Task 3: Draft `docs/hooks.md`

**Files:**
- Create: `docs/hooks.md`

- [ ] **Step 1: Define required contents checklist**

`docs/hooks.md` MUST contain:
1. H1: `# Hooks`
2. Intro paragraph: wired via `hooks/hooks.json`; most hooks delegate to `bin/claudey <subcommand>`.
3. H2 `## Hook Events` — single table ordered exactly as in `hooks/hooks.json`, columns `Event | Matcher | Command | Blocking?`. Include all entries from `hooks.json`. Use this row set:

   | Event          | Matcher       | Command                                                         | Blocking?                  |
   | -------------- | ------------- | --------------------------------------------------------------- | -------------------------- |
   | `PreToolUse`   | `Bash`        | `bin/claudey git-push-reminder`                                 | No (informational)         |
   | `PreToolUse`   | `Write`       | `bin/claudey block-random-docs`                                 | **Yes** (exit `2` blocks)  |
   | `PreToolUse`   | `Edit\|Write` | `bin/claudey suggest-compact`                                   | No                         |
   | `PreToolUse`   | `*`           | `skills/continuous-learning/hooks/observe.sh pre`               | No                         |
   | `PreCompact`   | `*`           | `bin/claudey pre-compact`                                       | No                         |
   | `SessionStart` | `*`           | `bin/claudey session-start`                                     | No                         |
   | `PostToolUse`  | `Bash`        | `bin/claudey pr-created-log`                                    | No                         |
   | `PostToolUse`  | `Edit`        | `bin/claudey post-edit-format`                                  | No                         |
   | `PostToolUse`  | `Edit`        | `bin/claudey post-edit-typecheck`                               | No                         |
   | `PostToolUse`  | `Edit`        | `bin/claudey post-edit-console-warn`                            | No                         |
   | `PostToolUse`  | `*`           | `skills/continuous-learning/hooks/observe.sh post`              | No                         |
   | `Notification` | `*`           | `afplay /System/Library/Sounds/Glass.aiff` (macOS only)         | No                         |
   | `Stop`         | `*`           | `bin/claudey check-console-log`                                 | No                         |
   | `Stop`         | `*`           | `afplay /System/Library/Sounds/Hero.aiff` (macOS only)          | No                         |
   | `SessionEnd`   | `*`           | `bin/claudey session-end`                                       | No                         |
   | `SessionEnd`   | `*`           | `bin/claudey evaluate-session`                                  | No                         |

4. H2 `## Per-Subcommand Behaviour` — one H3 per Rust-dispatched hook. Each H3:
   - Heading: `### <subcommand>`
   - 1-2 sentences summarising what the hook does.
   - One bullet line: `Source: \`src/hooks/<file>.rs\``.
   - For blocking hooks (`block-random-docs`), one extra line starting `Exit codes:` listing `0 = pass, 2 = block`.

   The H3s to include (12 total, matching Task 2's table):
   `session-start`, `session-end`, `pre-compact`, `suggest-compact`, `post-edit-format`, `post-edit-typecheck`, `post-edit-console-warn`, `check-console-log`, `evaluate-session`, `git-push-reminder`, `block-random-docs`, `pr-created-log`.

5. H2 `## Non-Rust Hooks` — two paragraphs:
   - `observe.sh pre` / `observe.sh post`: part of the `continuous-learning` skill — link `./skills.md#continuous-learning`.
   - `afplay` Notification and Stop hooks: play a system sound on macOS. No-op on Linux/Windows (the binary is missing).

6. H2 `## Authoring a New Hook` — 3-bullet checklist:
   - Add a matcher entry to `hooks/hooks.json`.
   - If the hook needs shared utilities, add a new subcommand under `src/hooks/` and register it in `src/hooks/mod.rs` + `src/main.rs` dispatch table.
   - Re-build with `bash bin/build-hooks.sh` and verify with `bash bin/test.sh`.

- [ ] **Step 2: Verify every `hooks.json` command has a row**

```bash
grep -oE '"command":\s*"[^"]+"' hooks/hooks.json | sort -u
```
Cross-check each line against the Step 1 table.

- [ ] **Step 3: Write `docs/hooks.md`**

- [ ] **Step 4: Verify required sections are present**

```bash
grep -E "^# Hooks$|^## Hook Events$|^## Per-Subcommand Behaviour$|^## Non-Rust Hooks$|^## Authoring a New Hook$" docs/hooks.md
```
Expected: 5 matching lines.

```bash
for sub in session-start session-end pre-compact suggest-compact \
           post-edit-format post-edit-typecheck post-edit-console-warn \
           check-console-log evaluate-session git-push-reminder \
           block-random-docs pr-created-log; do
  grep -q "^### $sub$" docs/hooks.md || echo "MISSING H3: $sub"
done
```
Expected: no "MISSING" lines.

- [ ] **Step 5: Commit**

```bash
git add docs/hooks.md
git commit -m "docs: add hooks reference"
```

---

## Task 4: Draft `docs/skills.md`

**Files:**
- Create: `docs/skills.md`

- [ ] **Step 1: Define required contents checklist**

`docs/skills.md` MUST contain:
1. H1: `# Skills`
2. Intro paragraph: skills are self-contained `SKILL.md` directories under `skills/`; Claude Code loads them via the `skills` key in `plugin.json`.
3. H2 `## Skill Catalogue` — a single table with columns `Skill | Description | Path` and exactly these 8 rows, one per directory in `skills/`:

   | Skill                         | Description                                                                          | Path                                        |
   | ----------------------------- | ------------------------------------------------------------------------------------ | ------------------------------------------- |
   | `content-hash-cache-pattern`  | Cache expensive file processing by SHA-256 of content                                | `skills/content-hash-cache-pattern/`        |
   | `continuous-learning`         | Instinct-based learning system with confidence scoring; evolves into skills/commands | `skills/continuous-learning/`               |
   | `generating-mermaid-diagrams` | Render `.mmd` files to PNG/SVG via local `mmdc` CLI                                  | `skills/mermaid-diagrams/`                  |
   | `pr-describe`                 | Generate a structured PR description from the current branch                         | `skills/pr-describe/`                       |
   | `pr-review-approve`           | Review and optionally approve a pull request                                         | `skills/pr-review-approve/`                 |
   | `regex-vs-llm-structured-text`| Decision framework: regex first, LLM only for low-confidence edges                   | `skills/regex-vs-llm-structured-text/`      |
   | `security-scan`               | Audit `.claude/` configuration using AgentShield                                     | `skills/security-scan/`                     |
   | `strategic-compact`           | Suggest manual `/compact` at logical workflow breakpoints                            | `skills/strategic-compact/`                 |

   Read the `description:` field from each `skills/<name>/SKILL.md` frontmatter to confirm the description text; keep it under ~120 characters per row.

4. H2 `## Multi-file Skills` — paragraph calling out `continuous-learning` (has subdirs `agents/`, `hooks/`, `scripts/`, plus `config.json`) and `strategic-compact` (ships a `suggest-compact.sh`). Explain that a skill can carry additional assets next to `SKILL.md`.
5. H2 `## Authoring a New Skill` — 3-bullet list:
   - Create `skills/<name>/SKILL.md` with YAML frontmatter (`name`, `description`; optional `version`, `allowed_tools`).
   - Skill is auto-discovered — no separate registration step.
   - For skills with code/scripts, co-locate them in the same directory.
6. H2 `## See Also` — link to `./commands.md` (skills often pair with slash commands) and `./hooks.md` (some skills ship hooks — e.g., `continuous-learning/hooks/observe.sh`).

- [ ] **Step 2: Verify names and descriptions**

```bash
for d in skills/*/; do
  echo "=== $d"
  head -6 "$d/SKILL.md"
done
```
Copy `name`/`description` frontmatter values into the table rows. `generating-mermaid-diagrams` vs directory name `mermaid-diagrams` is intentional — use the frontmatter `name` in the first column.

- [ ] **Step 3: Write `docs/skills.md`**

- [ ] **Step 4: Verify every skill directory has a row**

```bash
for d in skills/*/; do
  name=$(basename "$d")
  grep -q "skills/$name/" docs/skills.md || echo "MISSING: $name"
done
```
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add docs/skills.md
git commit -m "docs: add skills catalogue"
```

---

## Task 5: Draft `docs/commands.md`

**Files:**
- Create: `docs/commands.md`

- [ ] **Step 1: Define required contents checklist**

`docs/commands.md` MUST contain:
1. H1: `# Slash Commands`
2. Intro paragraph: slash commands are Markdown files in `commands/`; Claude Code exposes them as `/command-name`.
3. H2 `## Command Catalogue` — single table, columns `Command | Description | Source`. Exactly 11 rows, alphabetised by command name:

   | Command           | Description                                                                            | Source                        |
   | ----------------- | -------------------------------------------------------------------------------------- | ----------------------------- |
   | `/evolve`         | Cluster related instincts into skills, commands, or agents                             | `commands/evolve.md`          |
   | `/gsd-workflow`   | Run the complete GSD phase lifecycle (discuss → plan → execute → verify) in one command| `commands/gsd-workflow.md`    |
   | `/instinct-export`| Export instincts for sharing with teammates or other projects                          | `commands/instinct-export.md` |
   | `/instinct-import`| Import instincts from teammates, Skill Creator, or other sources                       | `commands/instinct-import.md` |
   | `/instinct-status`| Show all learned instincts with their confidence levels                                | `commands/instinct-status.md` |
   | `/learn`          | Extract reusable patterns from the current session                                     | `commands/learn.md`           |
   | `/learn-eval`     | Extract patterns with self-evaluation and smart save location                          | `commands/learn-eval.md`      |
   | `/refactor-clean` | Safely identify and remove dead code with test verification at every step              | `commands/refactor-clean.md`  |
   | `/sessions`       | List, load, alias, and edit session history                                            | `commands/sessions.md`        |
   | `/skill-create`   | Analyse local git history to extract patterns and generate `SKILL.md`                  | `commands/skill-create.md`    |
   | `/verify`         | Run comprehensive verification on the current codebase state                           | `commands/verify.md`          |

4. H2 `## Command Frontmatter` — short paragraph: commands may declare `name`, `description`, `argument-hint`, `allowed-tools` (or `allowed_tools`) in YAML frontmatter. Commands without frontmatter are still loaded; the first heading becomes the display name.
5. H2 `## Authoring a New Command` — 3-bullet list:
   - Create `commands/<name>.md`.
   - Add `---` frontmatter with `name` and `description` for discoverability.
   - Commit — no separate registration step.
6. H2 `## Related` — link `./skills.md` (learning-family commands pair with the `continuous-learning` skill).

- [ ] **Step 2: Verify descriptions against command file frontmatter**

```bash
for f in commands/*.md; do
  echo "=== $f"
  head -8 "$f"
done
```
Use `description` field where present; otherwise summarise from the first heading.

- [ ] **Step 3: Write `docs/commands.md`**

- [ ] **Step 4: Verify every command file has a row**

```bash
for f in commands/*.md; do
  name=$(basename "$f" .md)
  grep -q "/${name}" docs/commands.md || echo "MISSING: /$name"
done
```
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add docs/commands.md
git commit -m "docs: add slash commands catalogue"
```

---

## Task 6: Draft `docs/rules.md`

**Files:**
- Create: `docs/rules.md`

- [ ] **Step 1: Define required contents checklist**

`docs/rules.md` MUST contain:
1. H1: `# Rules`
2. Intro paragraph: rules are always-on Markdown instructions loaded into every session; split by scope into `common/`, `python/`, `typescript/`.
3. H2 `## Rule Set` — three sub-tables, one per scope:

   **`rules/common/` — all languages**

   | File                 | Topic                                                           |
   | -------------------- | --------------------------------------------------------------- |
   | `agents.md`          | How and when to use specialised agents                          |
   | `coding-style.md`    | Immutability, file organisation, error handling                 |
   | `git-workflow.md`    | Commit message format, PR workflow                              |
   | `hooks.md`           | Hook types, TodoWrite practices, auto-accept policy             |
   | `patterns.md`        | Skeleton projects, design patterns (repository, API envelope)   |
   | `performance.md`     | Model selection, context-window management, build troubleshooting|
   | `security.md`        | Secret management, mandatory pre-commit security checks         |
   | `testing.md`         | 80 % coverage minimum, TDD workflow, failure triage             |

   **`rules/python/` — Python projects**

   | File                 | Topic                   |
   | -------------------- | ----------------------- |
   | `coding-style.md`    | Python coding style     |
   | `hooks.md`           | Python-specific hooks   |
   | `patterns.md`        | Python design patterns  |
   | `security.md`        | Python security         |
   | `testing.md`         | Python testing          |

   **`rules/typescript/` — TypeScript projects**

   | File                 | Topic                        |
   | -------------------- | ---------------------------- |
   | `coding-style.md`    | TypeScript coding style      |
   | `hooks.md`           | TypeScript-specific hooks    |
   | `patterns.md`        | TypeScript design patterns   |
   | `security.md`        | TypeScript security          |
   | `testing.md`         | TypeScript testing           |

   **Note:** Topic cells in the Python/TypeScript tables should be one line. If `rules/README.md` already captures the topic in a different wording, use that wording to stay consistent.

4. H2 `## Loading` — paragraph: Claude Code reads every file under `rules/` at session start. Users drop into `~/.claude/rules/` (see root README) or install via the plugin system.
5. H2 `## Authoring a New Rule` — 3-bullet list:
   - Choose scope: `rules/common/` (always) vs `rules/<language>/` (conditional).
   - Keep files small and focused — one topic per file.
   - Commit; auto-discovered.

- [ ] **Step 2: Verify rule file listing**

```bash
find rules -type f -name '*.md' | sort
cat rules/README.md
```
Cross-check each file against the tables. If any file is missing from the tables, add a row. If any table row references a missing file, remove it.

- [ ] **Step 3: Write `docs/rules.md`**

- [ ] **Step 4: Verify every rule file has a row**

```bash
for f in $(find rules -type f -name '*.md' ! -name 'README.md'); do
  grep -q "\`$(basename "$f")\`" docs/rules.md || echo "MISSING: $f"
done
```
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add docs/rules.md
git commit -m "docs: add rules reference"
```

---

## Task 7: Draft `docs/README.md` (index)

**Files:**
- Create: `docs/README.md`

- [ ] **Step 1: Define required contents checklist**

`docs/README.md` MUST contain:
1. H1: `# Claudey Documentation`
2. Intro paragraph: table of contents for everything under `docs/`; the root `README.md` is the landing page and points here.
3. H2 `## Pages` — bulleted list, exactly these 6 entries in this order, each a relative link + one-line purpose:
   - `[Architecture](./architecture.md)` — system overview, plugin layout, hook flow.
   - `[Rust Binary](./binary.md)` — `bin/claudey` subcommands, module layout, build.
   - `[Hooks](./hooks.md)` — every entry in `hooks/hooks.json` with behaviour notes.
   - `[Skills](./skills.md)` — catalogue of every directory under `skills/`.
   - `[Slash Commands](./commands.md)` — catalogue of every file under `commands/`.
   - `[Rules](./rules.md)` — always-on rules split by language scope.
4. H2 `## Where to Start` — three-item bullet list:
   - New to the plugin: start at `./architecture.md`.
   - Adding or debugging a hook: start at `./hooks.md`, then `./binary.md`.
   - Writing a skill or slash command: start at `./skills.md` or `./commands.md`.
5. No trailing "last updated" line (git log is authoritative; adding a manual timestamp will rot).

- [ ] **Step 2: Verify sibling files exist**

```bash
ls docs/
```
Expected: architecture.md, binary.md, hooks.md, skills.md, commands.md, rules.md all present. If any are missing, **stop** and complete the corresponding earlier task first.

- [ ] **Step 3: Write `docs/README.md`**

- [ ] **Step 4: Verify all six links resolve**

```bash
for target in architecture.md binary.md hooks.md skills.md commands.md rules.md; do
  grep -q "\./$target" docs/README.md || echo "MISSING LINK: $target"
  test -f "docs/$target" || echo "MISSING FILE: docs/$target"
done
```
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add docs/README.md
git commit -m "docs: add docs index"
```

---

## Task 8: Rewrite root `README.md`

**Files:**
- Modify: `README.md` (full rewrite — overwrite with `Write`)

- [ ] **Step 1: Define required contents checklist**

The new root `README.md` MUST contain:
1. H1: `# Claudey`
2. Tagline paragraph (1-2 sentences). Example: "Agents, skills, hooks, commands, and rules for Claude Code — powered by a single Rust hook binary."
3. Credit line (1 sentence) to the upstream project, matching the current README's wording. Move the `> Credit:` blockquote verbatim.
4. H2 `## What's Inside` — short bullet list, max 6 bullets, each a count + link into `docs/`. Counts must match reality (verified in Step 2):
   - `[8 skills](./docs/skills.md)` — reusable domain knowledge.
   - `[11 slash commands](./docs/commands.md)` — `/learn`, `/verify`, `/sessions`, and more.
   - `[Event-driven hooks](./docs/hooks.md)` — auto-format on edit, session persistence, compaction hints.
   - `[Rule files](./docs/rules.md)` split by language scope.
   - `[Rust hook binary](./docs/binary.md)` — single `bin/claudey` dispatches every hook.
   - `[Architecture overview](./docs/architecture.md)`.
5. H2 `## Quick Start` — keep the existing `/plugin marketplace add` / `/plugin install` code block from the current README.
6. H2 `## Manual Install` — clone + build, matching the flow added in commit `0541d22`:
   ```bash
   git clone https://github.com/oguzsh/claudey.git
   cd claudey
   bash bin/build-hooks.sh          # builds bin/claudey locally (Rust toolchain required)
   # Copy what you need to ~/.claude/
   cp -r skills/ ~/.claude/skills/
   cp -r commands/ ~/.claude/commands/
   cp -r rules/ ~/.claude/rules/
   cp -r hooks/ ~/.claude/hooks/
   ```
   With a one-line note: "`bin/claudey` is not checked into git — each machine builds its own."
7. H2 `## Documentation` — short paragraph: "All subsystem documentation lives in [`docs/`](./docs/README.md)." Then the same 6 `./docs/<file>` links from Task 7's page list, as a bullet list.
8. H2 `## License` — `[MIT](LICENSE)`.

Remove all existing "Components Reference", architecture ASCII trees, and per-skill/per-command tables from the old README — they are now owned by `docs/`.

- [ ] **Step 2: Verify counts and links before writing**

```bash
ls skills/ | wc -l            # expected: 8
ls commands/*.md | wc -l      # expected: 11
ls docs/                      # expected: 6 sibling .md files + README.md
```
If any count diverges from Step 1, update the README count to match reality — never the other way around.

- [ ] **Step 3: Overwrite `README.md`**

Use the Write tool to replace the entire file with content that satisfies the Step 1 checklist. Keep it short: target ~80 lines.

- [ ] **Step 4: Verify all docs links resolve and stale content is gone**

```bash
for target in architecture.md binary.md hooks.md skills.md commands.md rules.md; do
  grep -q "docs/$target" README.md || echo "MISSING LINK: $target"
done

# Stale content that must not appear in the new README:
grep -E "go\.mod|internal/|Go binary|claudey-hooks|24 skills|29 slash commands|18 rule files" README.md \
  && echo "STALE CONTENT FOUND" || echo "clean"
```
Expected:
- No "MISSING LINK" lines.
- Final line prints `clean`.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: rewrite root README as docs/ entry point"
```

---

## Task 9: End-to-end verification

**Files:** (verification only; no files created)

- [ ] **Step 1: Verify every required doc file exists**

```bash
for f in docs/README.md docs/architecture.md docs/binary.md \
         docs/hooks.md docs/skills.md docs/commands.md docs/rules.md \
         README.md; do
  test -f "$f" || echo "MISSING: $f"
done
```
Expected: no output.

- [ ] **Step 2: Verify no broken relative links**

```bash
# Extract all ./*.md and ./docs/*.md references from README and docs/, check each target exists.
grep -rhoE '\./[A-Za-z0-9_/.-]+\.md' README.md docs/*.md | sort -u | while read target; do
  # Resolve relative to the file that referenced it; for simplicity check from repo root.
  # Strip leading ./ and any docs/ prefix normalisation already done upstream.
  stripped=${target#./}
  if [ -f "$stripped" ] || [ -f "docs/$stripped" ]; then
    :
  else
    echo "BROKEN: $target"
  fi
done
```
Expected: no "BROKEN" lines. If a link is broken because it lives inside `docs/` and references a sibling, also run:
```bash
grep -rhoE '\./[A-Za-z0-9_-]+\.md' docs/*.md | sort -u | while read t; do
  test -f "docs/${t#./}" || echo "BROKEN in docs/: $t"
done
```

- [ ] **Step 3: Verify the plan itself is still reachable**

```bash
test -f docs/superpowers/plans/2026-04-19-project-documentation.md && echo "plan present" || echo "MISSING plan"
```
Expected: `plan present`.

- [ ] **Step 4: Smoke-build to ensure no prose accidentally shadowed a real file**

```bash
bash bin/test.sh
```
Expected: all checks pass (this is a docs change; if this fails, the docs change broke something unrelated — investigate).

- [ ] **Step 5: Final commit if anything drifted**

If Steps 1-4 caught and the executor fixed anything, land the fixes as a single follow-up commit:

```bash
git add -A
git commit -m "docs: fix link rot and stale counts after review"
```

If nothing drifted, skip this step.

---

## Verification

After all tasks, a reader opening the repo can:
1. Read the root `README.md` in under two minutes and understand what Claudey is and how to install it (Task 8).
2. Click into `docs/README.md` and find every subsystem page within one hop (Task 7).
3. For any hook event in `hooks/hooks.json`, find the matching behaviour notes by name in `docs/hooks.md` (Task 3).
4. For any subcommand in `src/main.rs`'s dispatch table, find the matching module and purpose in `docs/binary.md` (Task 2).
5. For any directory in `skills/` or file in `commands/`, find a one-line description in `docs/skills.md` or `docs/commands.md` (Tasks 4, 5).
6. For any file in `rules/`, find it listed in `docs/rules.md` (Task 6).

---

## Reference — Existing Artefacts

- Plugin manifest: `.claude-plugin/plugin.json`
- Hook wiring: `hooks/hooks.json`
- Rust crate: `Cargo.toml`, `src/main.rs`, `src/hooks/*.rs`
- Build scripts: `bin/build-hooks.sh`, `bin/test.sh`
- Rules index: `rules/README.md`
- Skills root: `skills/<name>/SKILL.md`
- Commands root: `commands/<name>.md`
