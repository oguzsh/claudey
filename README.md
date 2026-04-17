# Claudey

Production-ready agents, skills, hooks, commands, and rules for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

> **Credit:** This project is based on [everything-claude-code](https://github.com/affaan-m/everything-claude-code) created by Affaan Mustafa.

## What is Claudey?

Claudey is a Claude Code plugin that gives your coding sessions a complete development toolkit out of the box. Instead of writing one-off prompts, you get structured workflows -- planning, TDD, code review, security scanning, session persistence, and more -- all wired together with event-driven hooks and a Go binary runtime.

- **24 skills** -- reusable domain knowledge for patterns, testing, deployment, and security
- **29 slash commands** -- `/plan`, `/tdd`, `/code-review`, `/verify`, and more
- **18 rule files** -- always-on guidelines for coding style, security, git workflow, and testing
- **Event-driven hooks** -- auto-format on edit, session persistence, console.log detection, compaction hints
- **Session persistence** -- automatic state save/restore across conversations
- **Go binary backend** -- compiled hook runner, package manager detection, session management

## Quick Start

```bash
/plugin marketplace add oguzsh/claudey
/plugin install claudey@oguzsh/claudey
```

Then try your first command:

```
/plan I need to add user authentication
```

## Installation

### Plugin Install (recommended)

```bash
/plugin install claudey@oguzsh/claudey
```

This installs everything: agents, skills, commands, hooks, rules, and the Go binary.

### Manual Copy

```bash
git clone https://github.com/oguzsh/claudey.git
cd claudey

# Copy what you need to ~/.claude/
cp -r skills/ ~/.claude/skills/
cp -r commands/ ~/.claude/commands/
cp -r rules/ ~/.claude/rules/
cp -r hooks/ ~/.claude/hooks/
```

## Architecture

```
claudey/
├── bin/             # Build and test scripts
├── commands/        # 29 slash commands (Markdown)
├── contexts/        # Context modes: dev, research, review
├── hooks/           # Event-driven hooks (JSON config + Go handlers)
│   └── bin/         # Compiled hook binary (claudey-hooks)
├── internal/        # Go source: hooks, session, package manager, git, etc.
├── mcp-configs/     # MCP server presets
├── rules/           # 18 always-on rule files
│   ├── common/      #   Language-agnostic (security, testing, git, etc.)
│   ├── python/      #   Python-specific
│   └── typescript/  #   TypeScript-specific
├── schemas/         # JSON schemas for hooks, plugins, package manager config
├── scripts/         # Cross-platform utilities
├── skills/          # 24 workflow skills (Markdown)
├── install.sh       # Rules installer (Claude + Cursor targets)
├── go.mod           # Go module definition
└── CLAUDE.md        # Project instructions for Claude Code
```

**Two-layer architecture:** Markdown/JSON configuration files define what agents, skills, commands, and rules do. A compiled Go binary (`bin/claudey`) handles the runtime -- hook execution, session management, package manager detection, and file operations.

## Components Reference

### Skills

<details>
<summary>11 skills grouped by category</summary>

**Patterns and Architecture**

| Skill                          | Description                                                |
| ------------------------------ | ---------------------------------------------------------- |
| `content-hash-cache-pattern`   | SHA-256 content hash caching for expensive file processing |
| `regex-vs-llm-structured-text` | Decision framework for regex vs LLM when parsing text      |

**Learning and Evolution**

| Skill                 | Description                                    |
| --------------------- | ---------------------------------------------- |
| `continuous-learning` | Instinct-based learning with confidence scores |
| `strategic-compact`   | Manual context compaction at logical intervals |

</details>

### Commands

<details>
<summary>29 slash commands grouped by category</summary>

**Core Workflow**

| Command   | Description                                                   |
| --------- | ------------------------------------------------------------- |
| `/verify` | Run comprehensive verification (build, lint, tests, security) |

**Learning and Evolution**

| Command            | Description                                                   |
| ------------------ | ------------------------------------------------------------- |
| `/learn`           | Extract reusable patterns from the current session            |
| `/learn-eval`      | Extract patterns with self-evaluation and smart save location |
| `/evolve`          | Cluster related instincts into skills, commands, or agents    |
| `/skill-create`    | Generate skills from local git history                        |
| `/instinct-status` | Show all learned instincts with confidence levels             |
| `/instinct-export` | Export instincts for sharing                                  |
| `/instinct-import` | Import instincts from teammates or other sources              |

**Session and Maintenance**

| Command     | Description                                   |
| ----------- | --------------------------------------------- |
| `/sessions` | List, load, alias, and manage session history |

</details>

### Rules

| Directory           | Files                                                                               | Scope               |
| ------------------- | ----------------------------------------------------------------------------------- | ------------------- |
| `rules/common/`     | agents, coding-style, git-workflow, hooks, patterns, performance, security, testing | All languages       |
| `rules/python/`     | coding-style, hooks, patterns, security, testing                                    | Python projects     |
| `rules/typescript/` | coding-style, hooks, patterns, security, testing                                    | TypeScript projects |

### Hooks

| Event                      | Hook                     | What It Does                                     |
| -------------------------- | ------------------------ | ------------------------------------------------ |
| `SessionStart`             | `session-start`          | Load previous context and detect package manager |
| `SessionEnd`               | `session-end`            | Persist session state                            |
| `SessionEnd`               | `evaluate-session`       | Evaluate session for extractable patterns        |
| `PreToolUse (Bash)`        | `git-push-reminder`      | Reminder before git push to review changes       |
| `PreToolUse (Write)`       | `block-random-docs`      | Block creation of random .md files               |
| `PreToolUse (Edit\|Write)` | `suggest-compact`        | Suggest manual compaction at logical intervals   |
| `PreCompact`               | `pre-compact`            | Save state before context compaction             |
| `PostToolUse (Bash)`       | `pr-created-log`         | Log PR URL and provide review command            |
| `PostToolUse (Edit)`       | `post-edit-format`       | Auto-format JS/TS files with Prettier            |
| `PostToolUse (Edit)`       | `post-edit-typecheck`    | TypeScript check after editing .ts/.tsx files    |
| `PostToolUse (Edit)`       | `post-edit-console-warn` | Warn about console.log statements                |
| `Stop`                     | `check-console-log`      | Check for console.log in modified files          |

### Commands

Create a Markdown file in `commands/`:

```markdown
---
description: What this command does when invoked.
---

# Command Name

Instructions for the agent when this command is invoked.
```

### Rules

Create a Markdown file in `rules/common/` (all languages) or `rules/<language>/`:

```markdown
# Rule Title

## Guidelines

- Guideline 1
- Guideline 2
```

Rules are always active -- they're loaded into every conversation automatically.

### Hooks

Add entries to `hooks/hooks.json`:

```json
{
	"matcher": "Edit",
	"hooks": [
		{
			"type": "command",
			"command": "echo 'File was edited'"
		}
	],
	"description": "Run after any file edit"
}
```

Hook events: `SessionStart`, `SessionEnd`, `PreToolUse`, `PostToolUse`, `PreCompact`, `Stop`.

## License

[MIT](LICENSE)
