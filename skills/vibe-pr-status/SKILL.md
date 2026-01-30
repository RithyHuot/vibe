---
name: vibe-pr-status
description: Check the status of a pull request including CI checks and approvals.
argument-hint: [pr-number]
allowed-tools: Bash(vibe:*)
---

# Check PR Status

Check the status of a pull request:

```bash
vibe pr-status $ARGUMENTS
```

If no PR number is provided, it checks the PR for the current branch.

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
