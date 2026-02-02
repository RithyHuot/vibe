---
name: vibe-pr-status
description: Check the status of a pull request including CI checks and approvals.
argument-hint: [pr-number]
allowed-tools: Bash(vibe:*), AskUserQuestion
---

# Check PR Status

## Steps

1. **Determine PR number**:
   - If `$ARGUMENTS` contains a PR number, use it
   - If no PR number provided, try to auto-detect from current branch
   - If auto-detection fails, use AskUserQuestion to ask: "Which PR number would you like to check?"

2. **Check the status**:

```bash
vibe pr-status $ARGUMENTS
```

If no PR number is provided and current branch has a PR, it checks that PR automatically.

## Output Includes

- **CI Checks**: All workflow statuses (passed/failed/pending)
- **Reviews**: Approval/change requests status
- **Merge Status**: Whether PR can be merged
- **Required Checks**: Which checks are required and their status

## Common Status Scenarios

| Status | Meaning | Next Action |
|--------|---------|-------------|
| ✅ All checks passing | Ready to merge | Review code, then `vibe merge` |
| ❌ CI failing | Tests/builds failed | Use `vibe ci-status` to debug |
| ⏳ Pending | Jobs still running | Wait and check again |
| 🔍 Changes requested | Reviewer wants updates | Address feedback, push changes |
| ⚠️ Merge conflict | Branch out of sync | Merge main: `git merge origin/main` |

## Workflow Integration

After checking status:

- If CI fails → `vibe ci-status` to debug
- If approved and passing → `vibe merge` to merge
- If needs updates → make changes, push, check status again
