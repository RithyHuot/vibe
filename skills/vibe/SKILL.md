---
name: vibe
description: Start work on a ClickUp ticket. Use when the user says "vibe" followed by a ticket ID (e.g., "vibe 86b7x5453").
argument-hint: <ticket-id>
allowed-tools: Bash(vibe:*), AskUserQuestion
---

# Start Work on a ClickUp Ticket

When the user provides a ticket ID, run the vibe command to create or checkout the branch for that ticket.

## Steps

1. **Check for ticket ID**:
   - If `$ARGUMENTS` contains a ticket ID, use it directly
   - If `$ARGUMENTS` is empty, use AskUserQuestion to ask: "What is the ClickUp ticket ID?" (Format: 9 alphanumeric characters, e.g., "86b7x5453")

2. Run the vibe command with the ticket ID:

   ```bash
   vibe $ARGUMENTS
   ```

3. After the branch is ready, automatically get the ticket context:

   ```bash
   vibe ticket
   ```

4. Use the ticket information (title, description, acceptance criteria) to understand what the user is trying to accomplish.

5. If the vibe command fails, check:
   - Is the ticket ID valid?
   - Are you in a git repository?
   - Run `git status` to verify repository state

## Notes

- The `vibe <ticket-id>` command may prompt for user input if the branch already exists (to choose: checkout, recreate, or cancel)
- If the branch doesn't exist, it creates it automatically
- Always run `vibe ticket` after to bring context into the session

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "No ticket found" | Invalid ticket ID | Verify ticket ID in ClickUp |
| "Not a git repository" | Not in git repo | `cd` to repository root |
| "Authentication failed" | ClickUp API token expired | Update token in config |
