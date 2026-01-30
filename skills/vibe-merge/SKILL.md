---
name: vibe-merge
description: Merge a pull request and post a merge comment. Only use when explicitly requested.
disable-model-invocation: true
allowed-tools: Bash(vibe:*), Bash(git:*)
---

# Merge Pull Request

**IMPORTANT**: Only run this when the user explicitly asks to merge.

## Pre-merge Checklist

Before merging, verify:

1. CI checks are passing: `vibe pr-status`
2. Required approvals are received
3. User has explicitly confirmed they want to merge

## Steps

1. **Check PR status**:

   ```bash
   vibe pr-status
   ```

2. **Ask for confirmation**:
   - Show the user the PR status
   - Ask: "Are you sure you want to merge this PR?"
   - Wait for explicit "yes" confirmation

3. **Merge** (only after confirmation):

   ```bash
   vibe merge [pr-number]
   ```

## Never Auto-merge

This command should NEVER be run automatically. Always require explicit user confirmation.

## Merge Methods

By default uses repository settings (usually "squash and merge"). The vibe CLI uses the repository's configured merge method.

## If Merge Fails

| Error | Solution |
|-------|----------|
| Merge conflicts | `git pull origin main`, resolve conflicts, push |
| CI not passing | Fix failures first with `vibe ci-status` |
| Missing approvals | Request reviews, wait for approval |
| Protected branch rules | Ensure all required checks pass |

## After Merge

1. Switch back to main: `git checkout main`
2. Pull latest: `git pull origin main`
3. Delete local branch: `git branch -d <branch-name>` (optional)
4. Update ClickUp ticket status to "Done" (if not auto-linked)
