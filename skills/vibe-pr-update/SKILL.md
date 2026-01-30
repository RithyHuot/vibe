---
name: vibe-pr-update
description: Update sections of an existing pull request. Use when the user wants to improve or fix the PR description.
argument-hint: [--summary "..." --description "..." --testing "..."]
allowed-tools: Bash(vibe:*), Bash(gh:*)
---

# Update Pull Request

Update specific sections of an existing PR.

## When to Use

- PR description needs clarification after reviews
- Testing instructions were incomplete
- Summary doesn't reflect additional changes
- Fixing typos or formatting in PR body

**Don't use for:**

- Adding new commits (just push to the branch)
- Changing PR title (use: `gh pr edit --title "..."`)
- Updating code changes (push new commits instead)

## Steps

1. **Check current PR body**:

   ```bash
   gh pr view --json body
   ```

2. **Update relevant sections**:

   ```bash
   vibe pr-update --summary "..." --description "..." --testing "..."
   ```

   Only include the flags for sections you're updating.

## Available Flags

- `--summary "..."` - Update the summary section
- `--description "..."` - Update the description section
- `--testing "..."` - Update the testing instructions
- `--ticket <id>` - Update the ticket reference

## Example

Update just the testing section:

```bash
vibe pr-update --testing "
1. Log in as test user
2. Navigate to settings
3. Verify email preferences save correctly
4. Expected: Success message appears
"
```

## Piping Content

For longer content from files or multi-line strings:

```bash
# From file
cat testing-steps.md | vibe pr-update --testing -

# Multi-line string
echo "Updated description:
- Fixed auth flow
- Added validation
- Improved error handling" | vibe pr-update --description -

# From command output
git log --oneline HEAD~5..HEAD | vibe pr-update --description -
```
