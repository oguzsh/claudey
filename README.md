# Claudey

Production-ready agents, skills, hooks, commands, and rules for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

> **Credit:** This project is based on [everything-claude-code](https://github.com/affaan-m/everything-claude-code) created by Affaan Mustafa.

## What is Claudey?

Claudey is a Claude Code plugin that gives your coding sessions a complete development toolkit out of the box. Instead of writing one-off prompts, you get structured workflows -- planning, TDD, code review, security scanning, session persistence, and more -- all wired together with event-driven hooks and a Go binary runtime.

- **11 specialized agents** -- planner, code reviewer, TDD guide, security reviewer, and more
- **24 skills** -- reusable domain knowledge for patterns, testing, deployment, and security
- **29 slash commands** -- `/plan`, `/tdd`, `/code-review`, `/verify`, and more
- **18 rule files** -- always-on guidelines for coding style, security, git workflow, and testing
- **Event-driven hooks** -- auto-format on edit, session persistence, console.log detection, compaction hints
- **MCP server presets** -- pre-configured Model Context Protocol integrations
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

### Rules-Only via install.sh

If you only need the rule files (coding style, security, testing guidelines):

```bash
# Install common + TypeScript rules to ~/.claude/rules/
./install.sh typescript

# Install multiple languages
./install.sh typescript python

```

### Interactive Setup

Copy the `configure-claudey` skill, then tell Claude:

```
configure claudey
```

This walks you through selecting which components to install.

### Manual Copy

```bash
git clone https://github.com/oguzsh/claudey.git
cd claudey

# Copy what you need to ~/.claude/
cp -r agents/ ~/.claude/agents/
cp -r skills/ ~/.claude/skills/
cp -r commands/ ~/.claude/commands/
cp -r rules/ ~/.claude/rules/
cp -r hooks/ ~/.claude/hooks/
cp -r contexts/ ~/.claude/contexts/
```

## Architecture

```
claudey/
├── agents/          # 11 specialized subagents (Markdown + YAML frontmatter)
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

### Agents

| Agent                  | Model  | Purpose                                                      |
| ---------------------- | ------ | ------------------------------------------------------------ |
| `planner`              | Opus   | Implementation planning for complex features and refactoring |
| `architect`            | Opus   | System design, scalability, and technical decision-making    |
| `tdd-guide`            | Sonnet | Test-driven development with write-tests-first enforcement   |
| `code-reviewer`        | Sonnet | Code quality, security, and maintainability review           |
| `security-reviewer`    | Sonnet | OWASP Top 10 vulnerability detection and remediation         |
| `build-error-resolver` | Sonnet | Fix build and type errors with minimal diffs                 |
| `database-reviewer`    | Sonnet | PostgreSQL query optimization, schema design, security       |
| `python-reviewer`      | Sonnet | PEP 8 compliance, Pythonic idioms, type hints                |
| `e2e-runner`           | Sonnet | E2E testing with Playwright                                  |
| `refactor-cleaner`     | Sonnet | Dead code cleanup and consolidation                          |
| `doc-updater`          | Haiku  | Documentation and codemap generation                         |

### Skills

<details>
<summary>24 skills grouped by category</summary>

**Patterns and Architecture**

| Skill                          | Description                                                             |
| ------------------------------ | ----------------------------------------------------------------------- |
| `api-design`                   | REST API design patterns, resource naming, status codes, pagination     |
| `backend-patterns`             | Backend architecture, database optimization, server-side best practices |
| `frontend-patterns`            | React, Next.js, state management, performance optimization              |
| `postgres-patterns`            | PostgreSQL query optimization, schema design, indexing, security        |
| `docker-patterns`              | Docker and Docker Compose patterns for local dev and CI                 |
| `deployment-patterns`          | CI/CD pipelines, health checks, rollback strategies                     |
| `content-hash-cache-pattern`   | SHA-256 content hash caching for expensive file processing              |
| `cost-aware-llm-pipeline`      | LLM API cost optimization, model routing, budget tracking               |
| `regex-vs-llm-structured-text` | Decision framework for regex vs LLM when parsing text                   |
| `iterative-retrieval`          | Progressive context retrieval for subagent workflows                    |

**Coding Standards**

| Skill              | Description                                                    |
| ------------------ | -------------------------------------------------------------- |
| `coding-standards` | Universal standards for TypeScript, JavaScript, React, Node.js |
| `python-patterns`  | PEP 8, type hints, Pythonic idioms                             |

**Testing**

| Skill               | Description                                                |
| ------------------- | ---------------------------------------------------------- |
| `tdd-workflow`      | Test-driven development: RED -> GREEN -> REFACTOR          |
| `e2e-testing`       | Playwright E2E patterns, Page Object Model, CI integration |
| `python-testing`    | pytest strategies, fixtures, mocking, parametrization      |
| `eval-harness`      | Eval-driven development (EDD) framework                    |
| `verification-loop` | Comprehensive verification system for sessions             |

**Security**

| Skill             | Description                                                |
| ----------------- | ---------------------------------------------------------- |
| `security-review` | Authentication, user input, secrets, API endpoint security |
| `security-scan`   | Scan `.claude/` directory for misconfigurations            |

**Learning and Evolution**

| Skill                    | Description                                    |
| ------------------------ | ---------------------------------------------- |
| `continuous-learning`    | Extract reusable patterns from sessions        |
| `continuous-learning-v2` | Instinct-based learning with confidence scores |
| `search-first`           | Research-before-coding workflow                |
| `strategic-compact`      | Manual context compaction at logical intervals |

**Setup**

| Skill               | Description                  |
| ------------------- | ---------------------------- |
| `configure-claudey` | Interactive installer wizard |

</details>

### Commands

<details>
<summary>29 slash commands grouped by category</summary>

**Core Workflow**

| Command           | Description                                                                    |
| ----------------- | ------------------------------------------------------------------------------ |
| `/plan`           | Create step-by-step implementation plan; waits for confirmation                |
| `/tdd`            | Test-driven development: scaffold, test first, implement, verify 80%+ coverage |
| `/code-review`    | Security and quality review of uncommitted changes                             |
| `/build-fix`      | Incrementally fix build and type errors with minimal changes                   |
| `/verify`         | Run comprehensive verification (build, lint, tests, security)                  |
| `/checkpoint`     | Create or verify a checkpoint in your workflow                                 |
| `/e2e`            | Generate and run E2E tests with Playwright                                     |
| `/refactor-clean` | Identify and remove dead code with test verification                           |

**Multi-Agent**

| Command           | Description                                 |
| ----------------- | ------------------------------------------- |
| `/multi-plan`     | Multi-model collaborative planning          |
| `/multi-execute`  | Multi-model collaborative execution         |
| `/multi-workflow` | Full multi-model development workflow       |
| `/multi-backend`  | Backend-focused multi-model workflow        |
| `/multi-frontend` | Frontend-focused multi-model workflow       |
| `/orchestrate`    | Sequential agent workflow for complex tasks |

**Testing and Coverage**

| Command          | Description                                               |
| ---------------- | --------------------------------------------------------- |
| `/test-coverage` | Analyze coverage, identify gaps, generate missing tests   |
| `/python-review` | Python-specific code review (PEP 8, type hints, security) |

**Learning and Evolution**

| Command            | Description                                                   |
| ------------------ | ------------------------------------------------------------- |
| `/learn`           | Extract reusable patterns from the current session            |
| `/learn-eval`      | Extract patterns with self-evaluation and smart save location |
| `/eval`            | Manage eval-driven development workflow                       |
| `/evolve`          | Cluster related instincts into skills, commands, or agents    |
| `/skill-create`    | Generate skills from local git history                        |
| `/instinct-status` | Show all learned instincts with confidence levels             |
| `/instinct-export` | Export instincts for sharing                                  |
| `/instinct-import` | Import instincts from teammates or other sources              |

**Session and Maintenance**

| Command            | Description                                            |
| ------------------ | ------------------------------------------------------ |
| `/sessions`        | List, load, alias, and manage session history          |
| `/pm2`             | Auto-analyze project and generate PM2 service commands |
| `/setup-pm`        | Configure preferred package manager                    |
| `/update-codemaps` | Generate token-lean architecture documentation         |
| `/update-docs`     | Sync documentation with codebase                       |

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
| `PreToolUse (Bash)`        | `block-dev-server`       | Block dev servers outside tmux                   |
| `PreToolUse (Bash)`        | `tmux-reminder`          | Reminder to use tmux for long-running commands   |
| `PreToolUse (Bash)`        | `git-push-reminder`      | Reminder before git push to review changes       |
| `PreToolUse (Write)`       | `block-random-docs`      | Block creation of random .md files               |
| `PreToolUse (Edit\|Write)` | `suggest-compact`        | Suggest manual compaction at logical intervals   |
| `PreCompact`               | `pre-compact`            | Save state before context compaction             |
| `PostToolUse (Bash)`       | `pr-created-log`         | Log PR URL and provide review command            |
| `PostToolUse (Bash)`       | `build-analysis`         | Async build analysis (background)                |
| `PostToolUse (Edit)`       | `post-edit-format`       | Auto-format JS/TS files with Prettier            |
| `PostToolUse (Edit)`       | `post-edit-typecheck`    | TypeScript check after editing .ts/.tsx files    |
| `PostToolUse (Edit)`       | `post-edit-console-warn` | Warn about console.log statements                |
| `Stop`                     | `check-console-log`      | Check for console.log in modified files          |

## Typical Workflow

```
1. /plan        Plan the implementation
2. /tdd         Write tests first, then implement
3. /code-review Review code quality and security
4. /verify      Run build, lint, tests, security checks
5. commit       Ship it
```

The agent orchestration flow:

```
User request
  │
  ├─ /plan ──────────► planner (Opus)
  │                        │
  │                        ▼
  ├─ /tdd ───────────► tdd-guide (Sonnet)
  │                        │
  │                        ▼
  ├─ /code-review ───► code-reviewer (Sonnet)
  │                    security-reviewer (Sonnet)  ← parallel
  │                        │
  │                        ▼
  └─ /verify ────────► build + lint + tests
```

## Creating Your Own Components

### Agents

Create a Markdown file in `agents/` with YAML frontmatter:

```markdown
---
name: my-agent
description: What this agent does and when to use it.
tools: ["Read", "Grep", "Glob"]
model: sonnet
---

You are an expert in [domain].

## Your Role

- Specific responsibility 1
- Specific responsibility 2
```

Model choices: `opus` (deep reasoning), `sonnet` (best coding), `haiku` (fast/cheap).

### Skills

Create `skills/<name>/SKILL.md`:

```markdown
---
name: my-skill
description: One-sentence description.
---

## When to Activate

- Trigger condition 1
- Trigger condition 2

## How It Works

1. Step 1
2. Step 2

## Examples

...
```

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

## Configuration

### MCP Servers

Copy entries from `mcp-configs/mcp-servers.json` into your `~/.claude.json` under `mcpServers` to enable pre-configured integrations.

### Contexts

Three context modes in `contexts/`:

| Context       | Purpose                              |
| ------------- | ------------------------------------ |
| `dev.md`      | Active development with full tooling |
| `research.md` | Exploration and analysis             |
| `review.md`   | Code review focus                    |

#### ZSH Aliases

```bash
# Daily development
alias claude-dev='claude --system-prompt "$(cat ~/.claude/contexts/dev.md)"'

# PR review mode
alias claude-review='claude --system-prompt "$(cat ~/.claude/contexts/review.md)"'

# Research/exploration mode
alias claude-research='claude --system-prompt "$(cat ~/.claude/contexts/research.md)"'
```

### Package Manager

Claudey auto-detects your package manager (npm, pnpm, yarn, bun). To override:

```bash
export CLAUDE_PACKAGE_MANAGER=pnpm
```

Or use the `/setup-pm` command interactively.

## Testing

```bash
# Run all validations
./bin/test.sh
```

Requires Go 1.22+ for building the binary. The test suite validates hook definitions, agent schemas, command formats, and rule files.

## Contributing

- **Conventional commits**: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- **File naming**: lowercase with hyphens (`python-reviewer.md`, `tdd-workflow.md`)
- **Run tests before PRs**: `./bin/test.sh`
- **Component formats**: see [Creating Your Own Components](#creating-your-own-components) above

## License

[MIT](LICENSE)
