---
name: continuous-learning
description: Instinct-based learning system that observes sessions via hooks, creates atomic instincts with confidence scoring, and evolves them into skills/commands/agents.
version: 2.0.0
---

# Continuous Learning - Instinct-Based Architecture

An advanced learning system that turns your Claude Code sessions into reusable knowledge through atomic "instincts" - small learned behaviors with confidence scoring.

## When to Activate

- Setting up automatic learning from Claude Code sessions
- Configuring instinct-based behavior extraction via hooks
- Tuning confidence thresholds for learned behaviors
- Reviewing, exporting, or importing instinct libraries
- Evolving instincts into full skills, commands, or agents

## What's In This Version

| Feature     | Description                               |
| ----------- | ----------------------------------------- |
| Observation | PreToolUse/PostToolUse (100% reliable)    |
| Analysis    | Background agent (Haiku)                  |
| Granularity | Atomic "instincts"                        |
| Confidence  | 0.3-0.9 weighted                          |
| Evolution   | Instincts → cluster → skill/command/agent |
| Sharing     | Export/import instincts                   |

## The Instinct Model

An instinct is a small learned behavior:

```yaml
---
id: prefer-functional-style
trigger: "when writing new functions"
confidence: 0.7
domain: "code-style"
source: "session-observation"
---

# Prefer Functional Style

## Action
Use functional patterns over classes when appropriate.

## Evidence
- Observed 5 instances of functional pattern preference
- User corrected class-based approach to functional on 2025-01-15
```

**Properties:**

- **Atomic** — one trigger, one action
- **Confidence-weighted** — 0.3 = tentative, 0.9 = near certain
- **Domain-tagged** — code-style, testing, git, debugging, workflow, etc.
- **Evidence-backed** — tracks what observations created it

## How It Works

```
Session Activity
      │
      │ Hooks capture prompts + tool use (100% reliable)
      ▼
┌─────────────────────────────────────────┐
│         observations.jsonl              │
│   (prompts, tool calls, outcomes)       │
└─────────────────────────────────────────┘
      │
      │ Observer agent reads (background, Haiku)
      ▼
┌─────────────────────────────────────────┐
│          PATTERN DETECTION              │
│   • User corrections → instinct         │
│   • Error resolutions → instinct        │
│   • Repeated workflows → instinct       │
└─────────────────────────────────────────┘
      │
      │ Creates/updates
      ▼
┌─────────────────────────────────────────┐
│         instincts/personal/             │
│   • prefer-functional.md (0.7)          │
│   • always-test-first.md (0.9)          │
│   • use-zod-validation.md (0.6)         │
└─────────────────────────────────────────┘
      │
      │ /evolve clusters
      ▼
┌─────────────────────────────────────────┐
│              evolved/                   │
│   • commands/new-feature.md             │
│   • skills/testing-workflow.md          │
│   • agents/refactor-specialist.md       │
└─────────────────────────────────────────┘
```

## Quick Start

### 1. Enable Observation Hooks

Add to your `~/.claude/settings.json`.

**If installed as a plugin** (recommended):

```json
{
	"hooks": {
		"PreToolUse": [
			{
				"matcher": "*",
				"hooks": [
					{
						"type": "command",
						"command": "${CLAUDE_PLUGIN_ROOT}/skills/continuous-learning/hooks/observe.sh pre"
					}
				]
			}
		],
		"PostToolUse": [
			{
				"matcher": "*",
				"hooks": [
					{
						"type": "command",
						"command": "${CLAUDE_PLUGIN_ROOT}/skills/continuous-learning/hooks/observe.sh post"
					}
				]
			}
		]
	}
}
```

**If installed manually** to `~/.claude/skills`:

```json
{
	"hooks": {
		"PreToolUse": [
			{
				"matcher": "*",
				"hooks": [
					{
						"type": "command",
						"command": "~/.claude/skills/continuous-learning-v2/hooks/observe.sh pre"
					}
				]
			}
		],
		"PostToolUse": [
			{
				"matcher": "*",
				"hooks": [
					{
						"type": "command",
						"command": "~/.claude/skills/continuous-learning-v2/hooks/observe.sh post"
					}
				]
			}
		]
	}
}
```

### 2. Initialize Directory Structure

The Python CLI will create these automatically, but you can also create them manually:

```bash
mkdir -p ~/.claude/book/{instincts/{personal,inherited},evolved/{agents,skills,commands}}
touch ~/.claude/book/observations.jsonl
```

### 3. Use the Instinct Commands

```bash
/instinct-status     # Show learned instincts with confidence scores
/evolve              # Cluster related instincts into skills/commands
/instinct-export     # Export instincts for sharing
/instinct-import     # Import instincts from others
```

## Commands

| Command                   | Description                                    |
| ------------------------- | ---------------------------------------------- |
| `/instinct-status`        | Show all learned instincts with confidence     |
| `/evolve`                 | Cluster related instincts into skills/commands |
| `/instinct-export`        | Export instincts for sharing                   |
| `/instinct-import <file>` | Import instincts from others                   |

## Configuration

Edit `config.json`:

```json
{
	"version": "2.0",
	"observation": {
		"enabled": true,
		"store_path": "~/.claude/book/observations.jsonl",
		"max_file_size_mb": 10,
		"archive_after_days": 7
	},
	"instincts": {
		"personal_path": "~/.claude/book/instincts/personal/",
		"inherited_path": "~/.claude/book/instincts/inherited/",
		"min_confidence": 0.3,
		"auto_approve_threshold": 0.7,
		"confidence_decay_rate": 0.05
	},
	"observer": {
		"enabled": true,
		"model": "haiku",
		"run_interval_minutes": 5,
		"patterns_to_detect": [
			"user_corrections",
			"error_resolutions",
			"repeated_workflows",
			"tool_preferences"
		]
	},
	"evolution": {
		"cluster_threshold": 3,
		"evolved_path": "~/.claude/book/evolved/"
	}
}
```

## File Structure

```
~/.claude/book/
├── identity.json           # Your profile, technical level
├── observations.jsonl      # Current session observations
├── observations.archive/   # Processed observations
├── instincts/
│   ├── personal/           # Auto-learned instincts
│   └── inherited/          # Imported from others
└── evolved/
    ├── agents/             # Generated specialist agents
    ├── skills/             # Generated skills
    └── commands/           # Generated commands
```

## Confidence Scoring

Confidence evolves over time:

| Score | Meaning      | Behavior                      |
| ----- | ------------ | ----------------------------- |
| 0.3   | Tentative    | Suggested but not enforced    |
| 0.5   | Moderate     | Applied when relevant         |
| 0.7   | Strong       | Auto-approved for application |
| 0.9   | Near-certain | Core behavior                 |

**Confidence increases** when:

- Pattern is repeatedly observed
- User doesn't correct the suggested behavior
- Similar instincts from other sources agree

**Confidence decreases** when:

- User explicitly corrects the behavior
- Pattern isn't observed for extended periods
- Contradicting evidence appears

## Privacy

- Observations stay **local** on your machine
- Only **instincts** (patterns) can be exported
- No actual code or conversation content is shared
- You control what gets exported

---

_Instinct-based learning: teaching Claude your patterns, one observation at a time._
