---
name: vibe-issue-update
description: Update an existing GitHub issue including title, body, state, assignees, labels, milestone, and projects. Use when user wants to "update issue", "close issue", "change issue", or modify issue metadata.
allowed-tools: Bash(vibe:*), Bash(gh:*)
---

# Update GitHub Issue

## Usage

```bash
vibe issue-update <issue-number> [flags]
```

## Common Operations

**Close an issue**:

```bash
vibe issue-update 123 --state closed
```

**Reopen an issue**:

```bash
vibe issue-update 123 --state open
```

**Update title and description**:

```bash
vibe issue-update 123 --title "New title" --body "Updated description"
```

**Update assignees** (replaces existing):

```bash
vibe issue-update 123 --assignees user1,user2
```

**Update labels** (replaces existing):

```bash
vibe issue-update 123 --labels bug,urgent,priority
```

**Set milestone**:

```bash
vibe issue-update 123 --milestone "v1.0"
```

**Add to projects**:

```bash
vibe issue-update 123 --projects "Project Name"
```

**Combine multiple updates**:

```bash
vibe issue-update 123 \
  --state closed \
  --labels fixed,verified \
  --assignees rithyhuot
```

## Flags

- `--title`: Update issue title
- `--body`: Update issue description/body
- `--state`: Change state (`open` or `closed`)
- `--assignees`: Update assignees (comma-separated, replaces existing)
- `--labels`: Update labels (comma-separated, replaces existing)
- `--milestone`: Set or change milestone
- `--projects`: Update projects (comma-separated)

## Important Notes

- At least one flag is required
- State must be either "open" or "closed"
- **CLI Mode Limitation**: When using gh CLI, assignees and labels are **additive**, not replaced
  - For true replacement behavior, use API mode: `github.mode: "api"` in config
- In API mode: Assignees, labels, and projects **replace** existing values
- Issue number is required (cannot be auto-detected)

## Use Cases

**Triage workflow**:

1. View issue: `vibe issue 123`
2. Assign and label: `vibe issue-update 123 --assignees user1 --labels bug,high-priority`
3. Add to sprint: `vibe issue-update 123 --milestone "Sprint 5"`

**Close resolved issue**:

```bash
vibe issue-update 123 --state closed --labels fixed
```

**Reassign issue**:

```bash
vibe issue-update 123 --assignees newuser
```

**Update details**:

```bash
vibe issue-update 123 --title "Better title" --body "Updated description with more context"
```

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| "Invalid state" | State not open/closed | Use `--state open` or `--state closed` |
| "No updates provided" | No flags specified | Add at least one update flag |
| "Issue not found" | Invalid issue number | Check issue number exists |
| "Authentication failed" | GitHub token expired | Run `gh auth login` |
