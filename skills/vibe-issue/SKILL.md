---
name: vibe-issue
description: View GitHub issue details including title, description, assignees, labels, milestone, state, and optionally comments. Use when user asks to "show issue", "view issue #123", or needs issue details.
allowed-tools: Bash(vibe:*), Bash(gh:*), AskUserQuestion
---

# View GitHub Issue

## Steps

1. **Determine issue number**:
   - If user provided an issue number in `$ARGUMENTS`, use it
   - If no issue number provided, check if it can be auto-detected from current branch name
   - If auto-detection fails or no branch context, use AskUserQuestion to ask: "Which issue number would you like to view?"

2. **Check for comments flag**:
   - If user wants to see comments, use AskUserQuestion to ask: "Would you like to include comments?" (Options: Yes, No)

3. **Run the command**:

```bash
vibe issue [issue-number] [--comments]
```

## Examples

**View specific issue**:

```bash
vibe issue 123
```

**View with comments**:

```bash
vibe issue 123 --comments
```

**Auto-detect from branch name**:

```bash
vibe issue
```

If no issue number is provided, the command attempts to extract it from the current branch name.

## Branch Name Patterns

The command recognizes these patterns:

- `issue-123` → extracts 123
- `123-fix-bug` → extracts 123
- `username/issue-123/description` → extracts 123
- `fix-issue-456` → extracts 456

## Output

Displays:

- Issue number and title
- State (OPEN/CLOSED)
- Author
- Assignees
- Labels
- Milestone (if set)
- Projects (if assigned)
- Timestamps (created, updated, closed)
- Full description/body
- Comments (if `--comments` flag used)

## Use Cases

- Check issue details before starting work
- Review issue requirements
- Read discussions in comments
- Verify issue state and assignments

## Branch Creation

After viewing an issue, you'll be prompted to create a branch:

- Automatically generates branch name: `username/issue-123/title-slug`
- Creates and checks out the new branch
- Checks out existing branch if it already exists
- Allows you to immediately start working on the issue

**Workflow:**

1. View issue: `vibe issue 123`
2. Review details
3. Confirm branch creation when prompted
4. Start working on the issue
