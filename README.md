# Claudey

Agents, skills, hooks, commands, and rules for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) -- powered by a single Rust hook binary.

> **Credit:** This project is based on [everything-claude-code](https://github.com/affaan-m/everything-claude-code) created by Affaan Mustafa.

## What's Inside

- [8 skills](./docs/skills.md) -- reusable domain knowledge for patterns, testing, PR workflow, and security.
- [11 slash commands](./docs/commands.md) -- `/learn`, `/verify`, `/sessions`, `/refactor-clean`, and more.
- [Event-driven hooks](./docs/hooks.md) -- auto-format on edit, session persistence, compaction hints, console.log detection.
- [Rule files](./docs/rules.md) split by language scope (`common/`, `python/`, `typescript/`).
- [Rust hook binary](./docs/binary.md) -- single `bin/claudey` dispatches every hook via subcommands.
- [Architecture overview](./docs/architecture.md) -- how the pieces fit together.

## Quick Start

```bash
/plugin marketplace add oguzsh/claudey
/plugin install claudey@oguzsh/claudey
```

Then try your first command:

```
/learn
```

## Manual Install

```bash
git clone https://github.com/oguzsh/claudey.git
cd claudey

# One-time setup: prompts to install Rust via rustup if missing, then builds bin/claudey
bash bin/setup.sh

# Copy what you need to ~/.claude/
cp -r skills/ ~/.claude/skills/
cp -r commands/ ~/.claude/commands/
cp -r rules/ ~/.claude/rules/
cp -r hooks/ ~/.claude/hooks/
```

`bin/claudey` is not checked into git — each machine builds its own binary from source via `bash bin/setup.sh`. After the first build, new sessions auto-rebuild whenever `src/` changes (no prompt, streamed to stderr).

## Documentation

All subsystem documentation lives in [`docs/`](./docs/README.md):

- [Architecture](./docs/architecture.md)
- [Rust Binary](./docs/binary.md)
- [Hooks](./docs/hooks.md)
- [Skills](./docs/skills.md)
- [Slash Commands](./docs/commands.md)
- [Rules](./docs/rules.md)

## License

[MIT](LICENSE)
