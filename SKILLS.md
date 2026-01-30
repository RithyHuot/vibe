# Claude Code Skills for vibe CLI

This document describes the Claude Code skills available for the vibe CLI. These skills enable Claude Code to seamlessly integrate with your development workflow.

## What are Claude Code Skills?

Claude Code skills are AI-powered integrations that allow Claude Code to interact with your development tools. Each skill defines how Claude should use a specific vibe CLI command, including when to use it and what context to provide.

## Available Skills

### 1. vibe

**Purpose**: Start work on a ClickUp ticket

**When to use**: User provides a ticket ID (e.g., "vibe 86b7x5453")

**What it does**:

- Creates or checks out a branch for the ticket
- Automatically fetches ticket context
- Updates task status to "In Progress"

**Example usage**:

```
User: "vibe 86b7x5453"
Claude: Runs `vibe 86b7x5453`, then `vibe ticket` to get context
```

---

### 2. vibe-ticket

**Purpose**: Get context on the current ClickUp ticket

**When to use**:

- User asks "what am I working on?"
- Starting a coding session on an existing branch
- Need to understand requirements

**What it does**:

- Fetches ticket title, status, URL, and description
- Displays acceptance criteria
- Provides context for implementation

**Example usage**:

```
User: "What am I working on?"
Claude: Runs `vibe ticket` and summarizes the requirements
```

---

### 3. vibe-comment

**Purpose**: Add a comment to the current ClickUp ticket

**When to use**: User wants to add a comment to the ticket

**What it does**:

- Posts a comment to the ticket in ClickUp
- Supports both direct text and piped content

**Example usage**:

```
User: "Add a comment saying I fixed the bug"
Claude: Runs `vibe comment "Fixed the bug"`
```

---

### 4. vibe-pr

**Purpose**: Create a pull request

**When to use**:

- User says "create a PR" or "open a PR"
- Code is ready for review

**What it does**:

- Gathers context from git diff and commit history
- Offers to run a code review first
- Generates PR title, summary, description, and testing instructions
- Gets explicit user approval before creating
- Creates the PR with all sections filled out

**Example usage**:

```
User: "Create a PR"
Claude:
1. Reads PR template
2. Analyzes changes
3. Offers code review
4. Generates PR content
5. Shows preview and asks for approval
6. Creates PR: `vibe pr -y --title "..." --summary "..." --description "..." --testing "..."`
```

**Important**: Always requires explicit user approval before creating the PR.

---

### 5. vibe-pr-status

**Purpose**: Check pull request status

**When to use**: User wants to know the status of a PR

**What it does**:

- Shows CI check status
- Displays review/approval status
- Indicates merge readiness

**Example usage**:

```
User: "Is my PR ready to merge?"
Claude: Runs `vibe pr-status` and reports the status
```

---

### 6. vibe-pr-update

**Purpose**: Update sections of an existing pull request

**When to use**: User wants to improve or fix PR description

**What it does**:

- Updates specific sections (summary, description, testing)
- Preserves other sections unchanged

**Example usage**:

```
User: "Update the PR description to mention the new API endpoint"
Claude:
1. Checks current PR: `gh pr view --json body`
2. Updates description: `vibe pr-update --description "..."`
```

---

### 7. vibe-merge

**Purpose**: Merge a pull request

**When to use**: Only when user explicitly requests to merge

**What it does**:

- Checks PR status and readiness
- Requires explicit user confirmation
- Posts /merge comment to trigger automation

**Example usage**:

```
User: "Merge the PR"
Claude:
1. Checks status: `vibe pr-status`
2. Shows status and asks for confirmation
3. After "yes": `vibe merge`
```

**Important**: Never auto-merges. Always requires explicit confirmation.

---

### 8. vibe-ci-status

**Purpose**: Check CircleCI status and help fix failures

**When to use**: User wants to check CI status or investigate failures

**What it does**:

- Shows all workflows and job status
- For failures, fetches full error output
- Helps identify and fix the issue

**Example usage**:

```
User: "Why is CI failing?"
Claude:
1. Checks status: `vibe ci-status`
2. Gets failure details: `vibe ci-failure`
3. Analyzes error, finds relevant code
4. Suggests fix
```

---

## Skill Configuration

Each skill is defined by a `SKILL.md` file in the `skills/<skill-name>/` directory. The file includes:

- **name**: Skill identifier
- **description**: When and how to use the skill
- **argument-hint**: Expected arguments (optional)
- **allowed-tools**: Tools Claude can use within the skill
- **disable-model-invocation**: Skip model calls for simple tasks (optional)

## Installing Skills

To make vibe skills available globally in **all your projects**:

### Option 1: Install During Initialization

```bash
vibe init --install-skills
```

### Option 2: Install Separately

```bash
vibe skills
```

Both methods copy the skills to `~/.claude/skills/` making them available to Claude Code everywhere.

### Option 3: Project-Specific Installation

If you prefer to keep skills local to specific projects, manually copy the `skills/` directory to each project:

```bash
# In your project directory
cp -r /path/to/vibe/skills ./
```

## Updating Skills

After updating vibe to a new version, update your installed skills to get the latest features and fixes:

```bash
vibe skills
```

This command:

- Overwrites existing skills with the latest versions
- Preserves any other skills in `~/.claude/skills/`
- Only affects vibe skills

**When to update:**

- After upgrading the vibe CLI
- When new skill features are released
- If you notice skills behaving unexpectedly

## Uninstalling Skills

To remove vibe skills from Claude Code:

```bash
vibe skills --uninstall
```

This command:

- Removes all vibe skills from `~/.claude/skills/`
- Does not affect other skills you may have installed
- Does not remove the vibe CLI itself
- Can be reinstalled anytime with `vibe skills`

**Why uninstall:**

- You no longer use Claude Code
- You prefer project-specific skills
- Troubleshooting skill conflicts

## Using Skills in Claude Code

1. **Install vibe CLI**: Follow the installation instructions in the README
2. **Configure vibe**: Run `vibe init` and set up your credentials
3. **Install skills globally**: Run `vibe init --install-skills` or `vibe skills`
4. **Use Claude Code**: Skills are now available in all your projects

## Workflow Examples

### Starting Work on a New Ticket

```
You: "vibe ABC123"
Claude:
- Runs vibe ABC123
- Fetches ticket details
- Shows you what you're working on
- Helps you understand the requirements
```

### Creating a PR with AI Assistance

```
You: "I'm ready to create a PR"
Claude:
- Analyzes your changes
- Offers to review the code
- Generates comprehensive PR description
- Shows you the preview
- Waits for your approval
- Creates the PR
```

### Fixing CI Failures

```
You: "CI is failing, can you help?"
Claude:
- Checks CI status
- Gets failure logs
- Identifies the issue
- Locates the problematic code
- Suggests a fix
- Implements the fix if you approve
```

### Checking Work Status

```
You: "What's the status of my work?"
Claude:
- Shows current ticket details
- Checks PR status if one exists
- Shows CI status if PR is open
- Summarizes what's left to do
```

## Best Practices

1. **Start with ticket context**: Use `vibe <ticket-id>` at the beginning of your session
2. **Check before merging**: Always review PR status before merging
3. **Let Claude help with PRs**: The AI can generate comprehensive descriptions
4. **Use CI checks**: Let Claude analyze failures and suggest fixes
5. **Keep tickets updated**: Use comments to track progress

## Contributing Skills

If you create custom skills for your workflow:

1. Create a new directory in `skills/`
2. Add a `SKILL.md` file with the skill definition
3. Test the skill with Claude Code
4. Submit a PR with your new skill

## Quick Reference

| Task | Command |
|------|---------|
| Install skills | `vibe skills` |
| Update skills | `vibe skills` |
| Uninstall skills | `vibe skills --uninstall` |
| Check installation | `ls ~/.claude/skills/` |
| Install during init | `vibe init --install-skills` |

## Troubleshooting

### Skill not working?

**Basic checks:**

1. Ensure vibe CLI is installed: `which vibe`
2. Check configuration: `vibe init`
3. Verify skills are installed: `ls ~/.claude/skills/`
4. Update to latest skills: `vibe skills`
5. Verify you're in a git repository

**Common issues:**

| Issue | Solution |
|-------|----------|
| "vibe: command not found" | Install vibe CLI: `go install github.com/rithyhuot/vibe/cmd/vibe@latest` |
| Skill doesn't appear | Reinstall: `vibe skills`, restart Claude Code |
| "unauthorized" error | Check API tokens in `~/.config/vibe/config.yaml` |
| "no ticket found" | Ensure you're on a ticket branch (e.g., `user/ticket-id/description`) |
| "repository not found" | Verify GitHub config in `~/.config/vibe/config.yaml` |

### Skills out of date?

After upgrading vibe CLI:

```bash
vibe skills
```

This ensures your skills match your CLI version.

### Want to start fresh?

Completely remove and reinstall:

```bash
vibe skills --uninstall
vibe skills
```

### Permission issues?

Skills use the `allowed-tools` directive to specify which tools Claude can use. If a skill fails:

1. Check the `allowed-tools` list in the skill's `SKILL.md`
2. Ensure the tool is available in your environment
3. Check that you have proper permissions

### Skill Output Examples

**Successful execution:**

```
✓ Successfully fetched ticket #abc123
✓ Created branch: username/abc123-feature-name
✓ Updated task status to: In Progress
```

**Error output:**

```
✗ Error: Invalid ticket ID format: abcd
  Expected 9 alphanumeric characters
```

### Debug Mode

Enable debug output to see what commands are being executed:

```bash
export DEBUG=true
# Now use Claude Code skills and check output
```

### Expected Behavior

**vibe-ticket skill:**

- Should display ticket title, status, assignees, and URL
- Takes 2-5 seconds depending on API response time
- Works from any ticket branch

**vibe-pr-status skill:**

- Shows CI check status, review approvals, and merge readiness
- Requires an open PR for the current branch
- Updates reflect real-time GitHub data

**vibe-code-review skill:**

- Analyzes uncommitted or staged changes
- Provides feedback on bugs, security, performance, and best practices
- May take 10-30 seconds for comprehensive review

### Getting Help

If issues persist:

1. Check [FAQ.md](FAQ.md) for common questions
2. Review [ARCHITECTURE.md](ARCHITECTURE.md) for system details
3. File an issue at: <https://github.com/rithyhuot/vibe/issues>

## Learn More

- [vibe CLI Documentation](README.md)
- [Claude Code Documentation](https://claude.ai/code)
- [Contributing Guide](CONTRIBUTING.md)
