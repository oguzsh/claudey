# Skills

Skills are self-contained directories under `skills/`, each with a `SKILL.md` file whose YAML frontmatter declares the skill's `name` and `description`. Claude Code loads them via the `skills` key in `.claude-plugin/plugin.json`. There is no registration step — dropping a new directory with a `SKILL.md` is enough.

## Skill Catalogue

| Skill                           | Description                                                                          | Path                                   |
| ------------------------------- | ------------------------------------------------------------------------------------ | -------------------------------------- |
| `content-hash-cache-pattern`    | Cache expensive file processing by SHA-256 of content — path-independent, auto-invalidating | `skills/content-hash-cache-pattern/`   |
| `continuous-learning`           | Instinct-based learning system: observes sessions, scores confidence, evolves into skills/commands/agents | `skills/continuous-learning/`          |
| `generating-mermaid-diagrams`   | Render `.mmd` source files to PNG/SVG via the local `mmdc` CLI                       | `skills/mermaid-diagrams/`             |
| `pr-describe`                   | Generate a structured PR description from the current branch's git history and diff | `skills/pr-describe/`                  |
| `pr-review-approve`             | Review and optionally approve a pull request by number or URL                        | `skills/pr-review-approve/`            |
| `regex-vs-llm-structured-text`  | Decision framework for parsing structured text: regex first, LLM only for low-confidence edges | `skills/regex-vs-llm-structured-text/` |
| `security-scan`                 | Audit `.claude/` configuration for injection risks and misconfigurations using AgentShield | `skills/security-scan/`                |
| `strategic-compact`             | Suggest manual `/compact` at logical workflow breakpoints rather than arbitrary auto-compaction | `skills/strategic-compact/`            |

## Multi-file Skills

Most skills ship only a `SKILL.md`, but a skill may carry additional assets next to it:

- `continuous-learning` contains `agents/`, `hooks/`, and `scripts/` subdirectories plus a `config.json`. Its `hooks/observe.sh` script is wired directly into `hooks/hooks.json` — see [`./hooks.md`](./hooks.md#non-rust-hooks).
- `strategic-compact` ships a `suggest-compact.sh` alongside its `SKILL.md`.

Both patterns are fine: a skill is a directory, not a single file.

## Authoring a New Skill

- Create `skills/<name>/SKILL.md` with YAML frontmatter — at minimum `name` and `description`; optionally `version`, `allowed_tools`, and so on.
- Skills are auto-discovered through the `skills` glob in `plugin.json`. No separate registration step.
- For skills with code, scripts, or config, co-locate them in the same directory so the skill stays self-contained.

## See Also

- [Slash Commands](./commands.md) — several commands (`/learn`, `/learn-eval`, `/evolve`, `/instinct-*`) pair with the `continuous-learning` skill.
- [Hooks](./hooks.md) — skills can ship their own hook entries, as `continuous-learning` does with its `observe.sh` observer.
