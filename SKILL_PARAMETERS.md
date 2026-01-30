# Claude Skills - Parameter Usage Guide

This guide explains how to pass and handle parameters in Claude Code skills for the vibe CLI.

## Overview

Claude Code skills support parameters through:

1. **Declaration**: Using `argument-hint` in skill frontmatter
2. **Access**: Using the `$ARGUMENTS` variable in skill commands
3. **Usage**: Claude automatically passes user-provided arguments

## Basic Parameter Usage

### 1. Declare Parameters in Frontmatter

```yaml
---
name: vibe-skill-name
description: Skill description
argument-hint: [parameter-name]  # or <required-param> [optional-param]
allowed-tools: Bash(vibe:*)
---
```

**Notation**:

- `[optional]` - Square brackets indicate optional parameters
- `<required>` - Angle brackets indicate required parameters
- `[param...]` - Ellipsis indicates multiple values accepted

### 2. Access Parameters with `$ARGUMENTS`

Inside your skill, use `$ARGUMENTS` to access passed parameters:

```bash
# Simple usage - pass directly to command
vibe pr-status $ARGUMENTS

# Conditional usage
if [ -n "$ARGUMENTS" ]; then
  PR_NUMBER="$ARGUMENTS"
  gh pr view $PR_NUMBER
else
  # Handle case with no arguments
  gh pr view
fi
```

### 3. User Invocation

When users invoke your skill, Claude automatically captures and passes arguments:

```
User: "review dependabot PR 456"
Claude: Runs skill with $ARGUMENTS = "456"

User: "check PR status"
Claude: Runs skill with $ARGUMENTS = "" (empty)
```

## Examples from Existing Skills

### Example 1: Optional PR Number

**File**: `skills/vibe-pr-status/SKILL.md`

```yaml
---
name: vibe-pr-status
argument-hint: [pr-number]
---
```

```bash
# Direct usage - passes PR number if provided
vibe pr-status $ARGUMENTS

# If no PR number provided, checks current branch's PR
```

**Invocations**:

- `"check PR status"` → `$ARGUMENTS = ""` → checks current branch
- `"check PR 123 status"` → `$ARGUMENTS = "123"` → checks PR #123

### Example 2: Required Issue Number

**File**: `skills/vibe-issue-update/SKILL.md`

```yaml
---
name: vibe-issue-update
argument-hint: <issue-number>
---
```

```bash
# Issue number is required
vibe issue-update $ARGUMENTS --state closed
```

**Invocations**:

- `"close issue 456"` → `$ARGUMENTS = "456"`
- `"close issue"` → Claude will prompt user for issue number

### Example 3: Multiple Parameters

**File**: `skills/vibe-issue/SKILL.md`

```yaml
---
name: vibe-issue
argument-hint: [issue-number] [--comments]
---
```

```bash
# Can pass multiple arguments
vibe issue $ARGUMENTS
```

**Invocations**:

- `"show issue 123"` → `$ARGUMENTS = "123"`
- `"show issue 123 with comments"` → `$ARGUMENTS = "123 --comments"`

### Example 4: Conditional Parameter Handling

**File**: `skills/vibe-dependabot-review/SKILL.md`

```yaml
---
name: vibe-dependabot-review
argument-hint: [pr-number]
---
```

```bash
# Check if argument provided
if [ -n "$ARGUMENTS" ]; then
  # User specified PR number
  PR_NUMBER="$ARGUMENTS"
  gh pr view $PR_NUMBER
else
  # No PR number - list Dependabot PRs
  gh pr list --author "app/dependabot" --limit 10
fi
```

## Advanced Patterns

### Pattern 1: Parsing Multiple Arguments

```bash
# Split arguments into variables
read -r ISSUE_NUMBER FLAGS <<< "$ARGUMENTS"

# Use separately
vibe issue-update $ISSUE_NUMBER $FLAGS
```

### Pattern 2: Providing Defaults

```bash
# Use default if no argument provided
PR_NUMBER="${ARGUMENTS:-current}"

if [ "$PR_NUMBER" = "current" ]; then
  vibe pr-status
else
  vibe pr-status $PR_NUMBER
fi
```

### Pattern 3: Validating Arguments

```bash
# Check if argument is a number
if [ -n "$ARGUMENTS" ] && ! [[ "$ARGUMENTS" =~ ^[0-9]+$ ]]; then
  echo "Error: PR number must be numeric"
  exit 1
fi

vibe pr-status $ARGUMENTS
```

### Pattern 4: Flags and Options

```bash
# Handle flags alongside positional arguments
TICKET_ID=$(echo "$ARGUMENTS" | awk '{print $1}')
FLAGS=$(echo "$ARGUMENTS" | cut -d' ' -f2-)

vibe ticket $TICKET_ID $FLAGS
```

## Parameter Naming Conventions

Use consistent naming in `argument-hint`:

| Type | Format | Example |
|------|--------|---------|
| PR number | `[pr-number]` | PR identifier (123, 456) |
| Issue number | `[issue-number]` | Issue identifier (789) |
| Ticket ID | `[ticket-id]` | ClickUp ticket (abc123xyz) |
| Branch name | `[branch]` | Git branch name |
| Job number | `[job-number]` | CI/CD job identifier |
| Flags | `[flags]` | Optional CLI flags |

## Best Practices

### 1. Clear Documentation

Always document what parameters your skill accepts:

```markdown
## Usage

```bash
vibe skill-name [pr-number]
```

**Parameters**:

- `pr-number` (optional): Pull request number. If omitted, uses current branch's PR.

**Examples**:

```bash
# Check current branch's PR
vibe skill-name

# Check specific PR
vibe skill-name 123
```

```

### 2. Handle Missing Parameters Gracefully

```bash
if [ -z "$ARGUMENTS" ]; then
  echo "No PR number provided. Checking current branch..."
  # Fallback behavior
else
  # Use provided argument
  PR_NUMBER="$ARGUMENTS"
fi
```

### 3. Provide Helpful Error Messages

```bash
if [ -z "$ARGUMENTS" ]; then
  echo "Error: Issue number required"
  echo "Usage: vibe issue-update <issue-number> [flags]"
  exit 1
fi
```

### 4. Support Both Interactive and Direct Usage

```bash
if [ -n "$ARGUMENTS" ]; then
  # Direct usage with argument
  TICKET_ID="$ARGUMENTS"
else
  # Interactive mode - let vibe CLI handle selection
  echo "Starting interactive ticket selection..."
fi

vibe ticket $TICKET_ID
```

## Testing Parameters

### Test Case 1: With Arguments

```
User: "review dependabot PR 123"
Expected: $ARGUMENTS = "123"
Command: gh pr view 123
```

### Test Case 2: Without Arguments

```
User: "review dependabot"
Expected: $ARGUMENTS = ""
Behavior: List all Dependabot PRs
```

### Test Case 3: Multiple Arguments

```
User: "show issue 456 with comments"
Expected: $ARGUMENTS = "456 --comments"
Command: vibe issue 456 --comments
```

## Common Patterns by Skill Type

### PR/Issue Skills

```yaml
argument-hint: [pr-number]
# or
argument-hint: [issue-number]
```

### Ticket Skills

```yaml
argument-hint: [ticket-id]
```

### CI/CD Skills

```yaml
argument-hint: [branch]
# or
argument-hint: [job-number]
```

### Action Skills (Merge, Close, etc.)

```yaml
argument-hint: <resource-id>  # Required
```

## Debugging Parameters

If parameters aren't working as expected:

1. **Check the frontmatter**: Ensure `argument-hint` is set correctly
2. **Test with echo**: Add `echo "Arguments: $ARGUMENTS"` to see what's passed
3. **Verify invocation**: Make sure user's request includes the parameter
4. **Check parsing**: Ensure your bash script correctly handles the arguments

## Summary

| Aspect | Implementation |
|--------|----------------|
| **Declaration** | `argument-hint: [param-name]` in frontmatter |
| **Access** | Use `$ARGUMENTS` variable in skill commands |
| **Optional** | `[param]` - Handle empty `$ARGUMENTS` |
| **Required** | `<param>` - Validate `$ARGUMENTS` not empty |
| **Multiple** | Parse `$ARGUMENTS` with bash string operations |
| **Flags** | Include in `$ARGUMENTS` string |

## Real-World Example

Here's a complete skill showing parameter usage:

```yaml
---
name: vibe-custom-skill
description: Custom skill with optional parameter
argument-hint: [identifier]
allowed-tools: Bash(vibe:*), Bash(gh:*)
---

# Custom Skill

## Usage

```bash
# With parameter
vibe custom-skill 123

# Without parameter (interactive)
vibe custom-skill
```

## Implementation

```bash
# Check if parameter provided
if [ -n "$ARGUMENTS" ]; then
  # Direct mode with parameter
  IDENTIFIER="$ARGUMENTS"
  echo "Using identifier: $IDENTIFIER"

  # Run command with parameter
  gh api repos/owner/repo/items/$IDENTIFIER
else
  # Interactive mode - list options
  echo "No identifier provided. Showing list..."
  gh api repos/owner/repo/items --jq '.[].id'

  # User can select from list
fi
```

This pattern makes your skill flexible and user-friendly!
