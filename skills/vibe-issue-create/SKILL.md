---
name: vibe-issue-create
description: Create a new GitHub issue. Use when user says "create an issue", "report a bug", "file an issue", "request a feature", or wants to create a new issue.
disable-model-invocation: true
allowed-tools: Bash(vibe:*), Bash(gh:*), Read, AskUserQuestion
---

# Create a GitHub Issue

## Steps

1. **Gather context**:
   - Read issue template if it exists: `.github/ISSUE_TEMPLATE.md`
   - Understand what the user wants to report (bug, feature, question, etc.)

2. **Ask for required and optional information** using AskUserQuestion if not already provided:

   **Required:**
   - **Title**: "What is the issue title?" (Clear, concise summary)

   **Optional but recommended:**
   - **Issue Type**: "What type of issue is this?" (Options: Bug, Feature, Enhancement, Question, Documentation)
   - **Description Details**: If it's a bug, ask:
     - "What is the problem or unexpected behavior?"
     - "What are the steps to reproduce?"
     - "What was the expected behavior?"
   - If it's a feature, ask:
     - "What feature would you like to add?"
     - "Why is this feature needed?"
     - "Do you have implementation suggestions?"

   **Metadata (optional):**
   - **Labels**: "What labels should be added?" (e.g., bug, feature, urgent, high-priority)
   - **Assignees**: "Who should be assigned to this issue?" (GitHub usernames)
   - **Milestone**: "Which milestone should this be added to?" (e.g., v1.0, Sprint 5)
   - **Projects**: "Which GitHub project should this be added to?"

3. **Generate content** (you are responsible for generating this):
   - **Title**: Clear, concise summary of the issue
   - **Description**: Detailed explanation including:
     - Problem description or feature request
     - Steps to reproduce (for bugs)
     - Expected vs actual behavior (for bugs)
     - Suggested implementation (for features)
     - Any relevant context or screenshots
   - **Metadata**:
     - Labels: e.g., `bug`, `feature`, `question`, `urgent`
     - Assignees: usernames of people who should work on it
     - Milestone: version/sprint to target
     - Projects: GitHub Projects to add the issue to

4. **Get explicit user approval**:
   - **REQUIRED**: Show the user the complete issue details (title, description, metadata)
   - **REQUIRED**: Ask: "Ready to create this issue?"
   - **REQUIRED**: Wait for explicit confirmation (e.g., "yes", "create it", "go ahead")
   - **Do NOT proceed** until you receive explicit approval
   - **Exception**: Skip approval if user's original request explicitly included `--yes` flag

5. **Create the issue** after receiving approval:

   ```bash
   vibe issue-create -y --title "..." --body "..." --labels bug,urgent --assignees user1
   ```

   - Always use `-y` flag to skip confirmation (user already approved)
   - Use `--body-file` for long descriptions (recommended)
   - Add labels with `--labels label1,label2`
   - Add assignees with `--assignees user1,user2`
   - Add milestone with `--milestone "v1.0"`
   - Add to projects with `--projects "Project Name"`

## Examples

**Bug report**:
```bash
vibe issue-create -y --title "Login fails with invalid redirect" \
  --body-file bug_description.md \
  --labels bug,urgent \
  --assignees rithyhuot
```

**Feature request**:
```bash
vibe issue-create -y --title "Add dark mode support" \
  --body "Users want dark mode..." \
  --labels feature,enhancement \
  --milestone "v2.0"
```

**Interactive mode** (no flags):
```bash
vibe issue-create
```

This will prompt for all fields interactively.

## Important Notes

- Title is required
- Description is optional but recommended
- Labels help with organization and filtering
- Assignees can be added during creation or later
- Issue template will be pre-filled if it exists
- The issue URL will be displayed after creation

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| "Authentication failed" | GitHub token expired | Run `gh auth login` |
| "Title is required" | Missing title | Add `--title "..."` |
| "Invalid label" | Label doesn't exist | Create label first or use existing ones |
| "User not found" | Invalid assignee | Check username spelling |
