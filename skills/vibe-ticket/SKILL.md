---
name: vibe-ticket
description: Get context on the current ClickUp ticket. Use when the user asks "what am I working on?", needs ticket context, or says "vibe ticket".
allowed-tools: Bash(vibe:*)
---

# Get Ticket Context

Run this command to get information about the current task from ClickUp:

```bash
vibe ticket
```

This outputs:

- Ticket title
- Status
- URL
- Description (including acceptance criteria)

## When to Use

- Starting a new coding session on an existing branch
- The user asks "what am I working on?" or similar
- You need to understand requirements before implementing
- The user explicitly says "vibe ticket"

Use the ticket information to understand what the user is trying to accomplish and any acceptance criteria or requirements.

## Example Output

```
Ticket: Add user authentication [In Progress]
URL: https://app.clickup.com/t/86b7x5453
Description:
  Implement OAuth2 login flow with GitHub integration

Acceptance Criteria:
  - Users can log in with GitHub
  - Session persists for 24 hours
  - Logout clears session completely
```
