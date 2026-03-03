---
name: pr-describe
description: Generate a structured PR description from the current branch's changes. Use when the user says "generate PR description", "describe PR", "PR summary", "pr-describe", "describe my changes", or "write PR description".
---

# PR Description Generator

Generate a comprehensive, structured pull request description by analyzing the current branch's git history and diff against the base branch.

## Workflow

### Step 1: Detect base branch and gather changes

1. Determine the base branch. Default to `main` unless the user specifies otherwise.
2. Run these git commands to gather all changes on the current branch:

```bash
# Get current branch name
git rev-parse --abbrev-ref HEAD

# Get commit log since diverging from base
git log main...HEAD --oneline --no-merges

# Get full diff against base
git diff main...HEAD
```

3. If the diff is very large, also run `git diff main...HEAD --stat` for a file-level summary to help with the high-level overview.

### Step 2: Analyze changes and generate description

Analyze all commits and the full diff to populate each section of the template below. Be thorough - read the actual code changes, not just commit messages.

**Guidelines for each section:**

- **What?** - Concise summary of what the PR does. Use bullet points for multiple changes. Focus on observable behavior changes and new capabilities.
- **Why?** - Explain the motivation. What problem does this solve? What ticket or initiative does it relate to? Infer from branch name, commit messages, and code context.
- **Architecture** - Include a mermaid diagram ONLY when changes involve architectural patterns: new modules, service layers, API routes, data flow changes, new abstractions, or significant structural reorganization. Skip this section entirely for simple bug fixes, config changes, or minor updates.
- **How it works?** - Explain the implementation approach. Walk through the key changes and how they connect. Mention important design decisions.
- **How can I test the branch?** - Provide concrete, actionable test steps. Include specific commands to run tests, endpoints to hit, or scenarios to verify. Reference specific test files if tests were added/modified.

### Step 3: Output the description

Output the formatted description using this template:

```markdown
## What?

[Bullet points summarizing what changed]

## Why?

[Motivation and context]

## Architecture

[Mermaid diagram - ONLY if architecturally significant changes. Otherwise omit this entire section.]

## How it works?

[Implementation walkthrough]

## How can I test the branch?

[Concrete test steps, commands, and scenarios]
```

### Step 4: Ask for feedback

After outputting the description, ask the user if they want to:
- Adjust any section
- Add more detail to a specific area
- Copy it to clipboard

Do NOT create a PR or push any code. This skill only generates the description text.
