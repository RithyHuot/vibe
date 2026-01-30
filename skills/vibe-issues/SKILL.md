---
name: vibe-issues
description: List GitHub issues with optional filtering and interactive selection. Use when user asks to "list issues", "show issues", "see all issues", or wants to browse issues.
allowed-tools: Bash(vibe:*), Bash(gh:*)
---

# List GitHub Issues

## Usage

```bash
vibe issues [--state open|closed|all] [--select]
```

## Examples

**List open issues** (default):

```bash
vibe issues
```

**List closed issues**:

```bash
vibe issues --state closed
```

**List all issues**:

```bash
vibe issues --state all
```

**Interactive selection mode**:

```bash
vibe issues --select
```

With `--select` flag, the command displays issues in an interactive list where you can:

- Select an issue to view full details
- Choose whether to include comments
- View complete issue information including description, labels, assignees, and metadata
- Create a branch for the issue to start working immediately

## Output

The command displays:

- Issue number
- Title
- State (OPEN/CLOSED)
- Labels (up to 2, then "...")
- Assignees (up to 2, then "...")

## Notes

- Default state is "open"
- Default limit is 30 issues (use `--limit` flag to change)
- Issues are displayed in a table format
- Interactive mode (`--select`) allows viewing full issue details without needing to know the issue number

## Workflow: Browse Issues and Start Working

Interactive mode provides a seamless workflow for picking up work:

1. List issues: `vibe issues --select`
2. Browse and select an issue
3. Choose whether to include comments
4. Review the full issue details
5. Optionally create a branch when prompted
6. Start working immediately

This workflow is ideal for:

- Triaging issues and starting work
- Exploring available work items
- Quickly context-switching between issues
