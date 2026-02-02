---
name: vibe-branch
description: Create and checkout a new branch. Use when the user wants to create a branch with or without a ticket ID (e.g., "create a branch", "vibe branch abc123xyz").
argument-hint: [ticket-id]
allowed-tools: Bash(vibe:branch*), AskUserQuestion
---

# Create and Checkout a New Branch

When the user wants to create a new branch, use the `vibe branch` command. This command can create branches with or without ticket IDs.

## Steps

### 1. Determine Branch Type

Ask the user if they haven't specified:

- **With Ticket ID**: If the user provides a ticket ID in their request, proceed directly to creating the branch
- **Without Ticket ID**: If the user wants a custom branch without a ticket ID, use AskUserQuestion to ask: "What would you like to name the branch?" or let the CLI prompt interactively

### 2. Create the Branch

**With Ticket ID**:

```bash
vibe branch <ticket-id>
```

This creates a branch in the format: `username/ticketid`

**Without Ticket ID (Interactive)**:

```bash
vibe branch
```

This will:
1. Prompt for a branch description (if not already provided via AskUserQuestion)
2. Create a branch in the format: `username/description`

## When to Use This Skill

Use this skill when:
- User says "create a branch" or "make a new branch"
- User says "vibe branch" followed by a ticket ID
- User wants to create a branch quickly without ClickUp integration
- User wants a branch with a custom description

## Comparison with `vibe` Command

| Command | Purpose | ClickUp Integration | Branch Format | Status Update |
|---------|---------|---------------------|---------------|---------------|
| `vibe branch <ticket-id>` | Quick branch creation | No | `username/ticketid` | No |
| `vibe <ticket-id>` | Full ticket workflow | Yes | `username/ticketid/task-name` | Yes (to "In Progress") |

## Notes

- The username is automatically extracted from `git config user.name`
- Usernames are sanitized (lowercase, spaces to hyphens, special chars removed)
- If the branch already exists, it will prompt to checkout the existing branch
- No ClickUp API call is made when using `vibe branch`

## Examples

```bash
# Create branch with ticket ID
vibe branch abc123xyz
# Creates: john-doe/abc123xyz

# Create branch with custom description (interactive)
vibe branch
# Prompts: "Branch description:"
# Input: fix-login-bug
# Creates: john-doe/fix-login-bug
```

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "Invalid ticket ID format" | Ticket ID is not 9 alphanumeric characters | Use valid ticket ID format |
| "Not a git repository" | Not in git repo | `cd` to repository root |
| "user.name not configured" | Git user.name not set | Run `git config user.name "Your Name"` |
| "Branch already exists" | Branch exists | Choose to checkout or cancel when prompted |
