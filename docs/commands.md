# Slash Commands

Slash commands are Markdown files in `commands/`. Claude Code exposes them as `/command-name`, loaded via the same `skills` glob in `.claude-plugin/plugin.json` that loads skills.

## Command Catalogue

| Command            | Description                                                                | Source                        |
| ------------------ | -------------------------------------------------------------------------- | ----------------------------- |
| `/evolve`          | Cluster related instincts into skills, commands, or agents                 | `commands/evolve.md`          |
| `/gsd-workflow`    | Run the complete GSD phase lifecycle (discuss → plan → execute → verify) in one command | `commands/gsd-workflow.md`    |
| `/instinct-export` | Export instincts for sharing with teammates or other projects              | `commands/instinct-export.md` |
| `/instinct-import` | Import instincts from teammates, Skill Creator, or other sources           | `commands/instinct-import.md` |
| `/instinct-status` | Show all learned instincts with their confidence levels                    | `commands/instinct-status.md` |
| `/learn`           | Extract reusable patterns from the current session                         | `commands/learn.md`           |
| `/learn-eval`      | Extract patterns, self-evaluate quality, and pick the right save location  | `commands/learn-eval.md`      |
| `/refactor-clean`  | Identify and remove dead code with test verification at every step         | `commands/refactor-clean.md`  |
| `/sessions`        | List, load, alias, and edit session history under `~/.claude/sessions/`    | `commands/sessions.md`        |
| `/skill-create`    | Analyse local git history to extract patterns and generate `SKILL.md`      | `commands/skill-create.md`    |
| `/verify`          | Run comprehensive verification on the current codebase state               | `commands/verify.md`          |

## Command Frontmatter

Commands may declare YAML frontmatter fields such as `name`, `description`, `argument-hint`, and `allowed-tools` (also spelled `allowed_tools` in some files). Commands without frontmatter still load; the first heading in the file becomes the display name.

## Authoring a New Command

- Create `commands/<name>.md`.
- Add a `---` frontmatter block with at least `name` and `description` for discoverability.
- Commit — no separate registration step; the command is discovered on the next session start.

## Related

- [Skills](./skills.md) — the learning-family commands (`/learn`, `/learn-eval`, `/evolve`, `/instinct-*`, `/skill-create`) pair with the `continuous-learning` skill.
