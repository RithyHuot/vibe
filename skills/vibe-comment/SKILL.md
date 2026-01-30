---
name: vibe-comment
description: Add a comment to the current ClickUp ticket. Use when the user says "vibe comment" or wants to add a comment to the ticket.
argument-hint: <comment text>
allowed-tools: Bash(vibe:*)
---

# Add Comment to Ticket

Add a comment to the current ClickUp ticket:

```bash
vibe comment "$ARGUMENTS"
```

If the user provides comment text, use it directly. If $ARGUMENTS is empty, ask: "What would you like to comment on the ticket?"

## When to Use

- Documenting a decision made during implementation
- Noting blockers or dependencies
- Updating stakeholders on progress
- Recording testing results

## Comment Best Practices

- Be specific: "Updated API endpoint to handle null values" not "fixed bug"
- Include context: "Blocked on PR #123 for auth changes"
- Mention testing: "Tested locally, verified logout clears session"

## Examples

```bash
# Quick comment
vibe comment "Fixed validation logic in user form"

# Multi-line comment
vibe comment "
Completed acceptance criteria:
- [x] Users can log in
- [x] Session persists
- [ ] Password reset (pending API)
"
```

## Alternative: Pipe Content

For longer comments or content from files:

```bash
# From file
cat notes.txt | vibe comment

# From multi-line string
echo "Comment content here" | vibe comment
```
