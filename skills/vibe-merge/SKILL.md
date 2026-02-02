---
name: vibe-merge
description: Merge a pull request and post a merge comment. Only use when explicitly requested.
disable-model-invocation: true
allowed-tools: Bash(vibe:*), Bash(git:*), AskUserQuestion
---

# Merge Pull Request

**IMPORTANT**: Only run this when the user explicitly asks to merge.

## Pre-merge Checklist

Before merging, verify:

1. CI checks are passing: `vibe pr-status`
2. Required approvals are received
3. User has explicitly confirmed they want to merge

## Steps

1. **Determine PR number**:
   - If `$ARGUMENTS` contains a PR number, use it
   - If no PR number provided, try to auto-detect from current branch
   - If auto-detection fails, use AskUserQuestion to ask: "Which PR number would you like to merge?"

2. **Check PR status**:

   ```bash
   vibe pr-status [pr-number]
   ```

3. **Ask for confirmation** using AskUserQuestion:
   - Show the user the PR status (CI checks, approvals, merge status)
   - Use AskUserQuestion to ask: "Are you sure you want to merge this PR?" (Options: Yes, No, Check status again)
   - **REQUIRED**: Wait for explicit "Yes" confirmation
   - If "Check status again", return to step 2
   - If "No", abort the merge

4. **Merge** (only after confirmation):

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
