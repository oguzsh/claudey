# Plugins and Marketplaces

---

## Marketplaces

Marketplaces are repositories of installable plugins.

### Adding a Marketplace

```bash
# Add official Anthropic marketplace
claude plugin marketplace add https://github.com/anthropics/claude-plugins-official

# Add community marketplaces
claude plugin marketplace add https://github.com/EveryInc/compound-engineering-plugin
```

### Recommended Marketplaces

| Marketplace                  | Source                               |
| ---------------------------- | ------------------------------------ |
| claude-plugins-official      | `anthropics/claude-plugins-official` |
| claude-code-plugins          | `anthropics/claude-code`             |
| Mixedbread-Grep              | `mixedbread-ai/mgrep`                |
| obra/superpowers-marketplace | `obra/superpowers-marketplace`       |
| compound-engineering-plugin  | `compound-engineering-plugin`        |

---

## Installing Plugins

```bash
# Open plugins browser
/plugins

# Or install directly
claude plugin install typescript-lsp@claude-plugins-official
```

### Recommended Plugins

**Development:**

- `typescript-lsp` - TypeScript intelligence
- `pyright-lsp` - Python type checking
- `hookify` - Create hooks conversationally
- `code-simplifier` - Refactor code

**Code Quality:**

- `code-review` - Code review
- `pr-review-toolkit` - PR automation
- `security-guidance` - Security checks

**Search:**

- `mgrep` - Enhanced search (better than ripgrep)
- `context7` - Live documentation lookup

**Workflow:**

- `commit-commands` - Git workflow
- `frontend-design` - UI patterns
- `feature-dev` - Feature development
- `superpowers-plugin` - Superpowers plugin
- `compound-engineering-plugin` - Compound engineering plugin

---

## Quick Setup

```bash
# Add marketplaces
claude plugin marketplace add https://github.com/anthropics/claude-plugins-official
claude plugin marketplace add https://github.com/mixedbread-ai/mgrep
claude plugin marketplace add https://github.com/EveryInc/compound-engineering-plugin
claude plugin marketplace add obra/superpowers-marketplace

# Open /plugins and install what you need
```

---

## Plugin Files Location

```
~/.claude/plugins/
|-- cache/                    # Downloaded plugins
|-- installed_plugins.json    # Installed list
|-- known_marketplaces.json   # Added marketplaces
|-- marketplaces/             # Marketplace data
```

