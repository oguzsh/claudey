---
name: gsd-workflow
description: "Run the complete GSD phase lifecycle — discuss, plan, execute (with mandatory superpowers), verify — in one command"
argument-hint: "<phase-number> [workspace-path]"
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - Task
  - TodoWrite
  - AskUserQuestion
  - Skill
---

<objective>
Full GSD phase lifecycle orchestrator with mandatory superpowers enforcement.

Sequences discuss-phase → plan-phase → superpowers-aware execute → verify-phase for a single phase. All four superpowers disciplines are enforced on every executor task — no skipping, no flags to disable them:

1. **Test-Driven Development** — RED-GREEN-REFACTOR, no production code before a failing test
2. **Git Worktrees** — isolated workspace per plan via `isolation="worktree"`
3. **Subagent-Driven Development** — spec review + code quality review after each plan
4. **Verification Before Completion** — evidence before claims, always

Use this instead of running discuss/plan/execute/verify individually when you want enforced quality on every task in the phase.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/workflow.md
@$HOME/.claude/get-shit-done/references/ui-brand.md
</execution_context>

<runtime_note>
**Copilot (VS Code):** Use `vscode_askquestions` wherever this workflow calls `AskUserQuestion`. They are equivalent — `vscode_askquestions` is the VS Code Copilot implementation of the same interactive question API.
</runtime_note>

<context>
Phase: $ARGUMENTS

Context files are resolved in-workflow using `init phase-op` and roadmap/state tool calls.
</context>

<process>
Execute the workflow from @$HOME/.claude/get-shit-done/workflows/workflow.md end-to-end.
Preserve all workflow gates (discuss checkpoint, plan checkpoint, wave execution, verification).
</process>
