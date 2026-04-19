# Rules

Rules are always-on Markdown instructions loaded into every Claude Code session. They're split into a `common/` layer (language-agnostic principles) and language-specific directories that extend the common rules with framework and tooling specifics. The installer (`./install.sh <language>`) copies the common rules plus the chosen language directories into `~/.claude/rules/`.

## Rule Set

### `rules/common/` — all languages

| File              | Topic                                                            |
| ----------------- | ---------------------------------------------------------------- |
| `agents.md`       | How and when to use specialised agents                           |
| `coding-style.md` | Immutability, file organisation, error handling                  |
| `git-workflow.md` | Commit message format, PR workflow                               |
| `hooks.md`        | Hook types, TodoWrite practices, auto-accept policy              |
| `patterns.md`     | Skeleton projects, design patterns (repository, API envelope)    |
| `performance.md`  | Model selection, context-window management, build troubleshooting |
| `security.md`     | Secret management, mandatory pre-commit security checks          |
| `testing.md`      | 80 % coverage minimum, TDD workflow, failure triage              |

### `rules/python/` — Python projects

| File              | Topic                      |
| ----------------- | -------------------------- |
| `coding-style.md` | Python coding style        |
| `hooks.md`        | Python-specific hooks      |
| `patterns.md`     | Python design patterns     |
| `security.md`     | Python security            |
| `testing.md`      | Python testing             |

### `rules/typescript/` — TypeScript / JavaScript projects

| File              | Topic                         |
| ----------------- | ----------------------------- |
| `coding-style.md` | TypeScript coding style       |
| `hooks.md`        | TypeScript-specific hooks     |
| `patterns.md`     | TypeScript design patterns    |
| `security.md`     | TypeScript security           |
| `testing.md`      | TypeScript testing            |

Language directories extend the common rules with code examples and tool-specific guidance. Each language file typically opens with a `> This file extends [common/xxx.md](../common/xxx.md) with <Language> specific content.` reference.

## Loading

Claude Code reads every file under `~/.claude/rules/` at session start. Users either run `./install.sh <language>` to copy the chosen rule sets, or copy the directories manually (keeping `common/` and language dirs separate — flattening them breaks the `../common/` references).

## Authoring a New Rule

- Pick a scope: `rules/common/` for universal guidance, or `rules/<language>/` for anything language-specific.
- Keep files small and focused — one topic per file, following the existing filenames (`coding-style.md`, `testing.md`, etc.) so language dirs mirror `common/`.
- Commit — Claude Code picks up the new file on the next session.
