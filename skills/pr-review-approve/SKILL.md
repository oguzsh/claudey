---
name: pr-review-approve
description: >-
  Use when the user wants to review and optionally approve a pull request.
  Triggers: "review and approve PR", "approve PR #123", PR number/URL with
  approval intent, "LGTM this PR", "check and approve".
---

# PR Review and Approve

Automated PR review via code-reviewer agent with conditional `gh pr review --approve`.

## Workflow

### Step 1: Parse PR Identifier

Extract PR number from user input. Accepted formats:
- `123` or `#123` (PR number)
- `https://github.com/owner/repo/pull/123` (full URL)
- `owner/repo#123` (shorthand)

If no argument provided, detect from current branch: `gh pr view --json number -q .number`

### Step 2: Save Current Branch

```bash
ORIGINAL_BRANCH=$(git rev-parse --abbrev-ref HEAD)
```

Store for restoration in Step 10.

### Step 3: Validate gh Authentication

```bash
gh auth status
```

If not authenticated, show:
> Not authenticated. Run `gh auth login` to authenticate, then retry.

**Stop here if auth fails.**

### Step 4: Fetch PR Metadata

```bash
gh pr view <NUMBER> --json number,title,author,baseRefName,headRefName,url,state
```

If PR not found or closed, show error and stop.
Display PR title, author, and branch info to user.

### Step 5: Checkout PR Branch

```bash
gh pr checkout <NUMBER>
```

If checkout fails, show the git error and stop. Do NOT force-checkout.

### Step 6: Run Code-Reviewer Agent

Dispatch `claudey:code-reviewer` agent with prompt:

> Review the changes in this PR. Run `git diff <baseRefName>...HEAD` to see all changes.
> Apply the full review checklist. End with the Review Summary table showing
> CRITICAL, HIGH, MEDIUM, LOW counts and verdicts.

### Step 7: Parse Review Results

From the code-reviewer output, extract severity counts:

| Severity | Count |
|----------|-------|
| CRITICAL | N |
| HIGH | N |
| MEDIUM | N |
| LOW | N |

### Step 8: Determine Approval Eligibility

**Eligible for approval** (all must be true):
- CRITICAL = 0
- HIGH = 0
- MEDIUM = 0

LOW issues are informational and do not block approval.

**Not eligible**: Any MEDIUM, HIGH, or CRITICAL issue found.

### Step 9: Show Summary and Ask User

**If eligible**, display:

```
PR #<N>: <title>
Review: PASS (0 blocking issues, N LOW informational)
```

Then use `AskUserQuestion` with options:
- **Approve** - Approve with standard review summary
- **Approve with message** - Approve with custom comment
- **Skip** - Do not approve

**If not eligible**, display:

```
PR #<N>: <title>
Review: BLOCKED (N CRITICAL, N HIGH, N MEDIUM issues)
```

List the blocking issues. Do NOT offer approval. Suggest the author fix the issues.

### Step 10: Execute Approval and Restore Branch

If user chose **Approve**:

Write a concise summary of what the PR does based on the diff reviewed in Step 6. The summary should explain the purpose and key changes (not the review outcome).

```bash
gh pr review <NUMBER> --approve --body "Reviewer Summary: <concise summary of what the PR does and its key changes>"
```

If user chose **Approve with message**, prepend "Reviewer Summary: " to their custom message and use that as the body.

Then restore original branch:

```bash
git checkout $ORIGINAL_BRANCH
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| `gh` not authenticated | Show `gh auth login` instructions, stop |
| PR not found | Show error, suggest `gh pr list`, stop |
| PR already closed/merged | Show state, stop |
| Checkout fails | Show git error, stop |
| Approval fails | Show possible causes: no write access, self-approval blocked by repo rules, branch protection |

## Common Mistakes

- **Self-approval**: GitHub blocks approving your own PR. The skill should warn if PR author matches `gh api user -q .login`.
- **Forgetting to restore branch**: Always checkout `$ORIGINAL_BRANCH` even on error.
- **Stale checkout**: If the PR branch has been force-pushed, `gh pr checkout` fetches the latest.
