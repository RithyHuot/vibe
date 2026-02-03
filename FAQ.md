# Frequently Asked Questions (FAQ)

## Table of Contents

- [General Questions](#general-questions)
- [Configuration Questions](#configuration-questions)
- [Troubleshooting](#troubleshooting)
- [Workflow Questions](#workflow-questions)
- [Integration Questions](#integration-questions)
- [Performance Questions](#performance-questions)

## General Questions

### Can I use vibecli with multiple ClickUp workspaces?

No, vibecli works with **one ClickUp workspace at a time**. The workspace is set via the `clickup.workspace_id` in your configuration:

```yaml
clickup:
  workspace_id: "1234567"  # Single ClickUp workspace
  team_id: "1234567"       # Team ID (often same as workspace_id)
```

**To switch to a different ClickUp workspace:**

1. Edit `.vibe.yaml` and change the `workspace_id` value
2. Or run `vibe init` to reconfigure from scratch
3. Restart vibe commands

**Note about the `workspaces:` config section:**

The `workspaces:` array in your config is **not** for multiple ClickUp workspaces. It's for organizing folders and sprint detection **within your single ClickUp workspace**:

```yaml
workspaces:
  - name: "Engineering"      # Just a label for reference
    folder_id: "123456789"   # Folder ID within your workspace
    sprint_patterns:          # Patterns to identify sprint folders
      - "Sprint \\d+ \\("
```

**For multiple ClickUp workspaces:**

If you work with multiple ClickUp workspaces (e.g., different clients or projects), maintain separate project directories, each with its own `.vibe.yaml` file:

```
~/client-a/.vibe.yaml  # workspace_id: "111111"
~/client-b/.vibe.yaml  # workspace_id: "222222"
```

### Does vibecli support GitLab or Bitbucket?

Currently, vibecli only supports GitHub for pull request and issue management. ClickUp integration works regardless of your Git provider, but GitHub-specific features (PR creation, issue sync, etc.) require GitHub.

If you're using GitLab or Bitbucket, you can still use:

- ClickUp ticket management (`vibe workon`, `vibe ticket`, `vibe comment`)
- Branch creation and Git workflows
- CircleCI status checking

### Where does vibecli store its cache?

vibecli currently uses **in-memory caching only**. There is no persistent cache stored on disk.

**What is cached:**

- Sprint folder data (1 hour TTL)

**What is NOT cached:**

- ClickUp task data
- GitHub API responses
- CircleCI workflow data

All cache data is lost when the application exits.

### Can I use vibecli without ClickUp?

No, vibecli is designed to work with ClickUp as the primary task management system. The core workflow revolves around ClickUp tickets. However, you can use the GitHub and CircleCI features independently through their respective commands.

### How do I update vibecli to the latest version?

If installed via Homebrew:

```bash
brew update
brew upgrade vibecli
```

If installed from source:

```bash
cd /path/to/vibecli
git pull
make install
```

Check your current version:

```bash
vibe --version
```

## Configuration Questions

### What tokens do I need to configure?

vibecli requires the following tokens:

1. **ClickUp API Token** (Required)
   - Get it from: <https://app.clickup.com/settings/apps>
   - Permissions needed: Read/Write tasks, comments
   - Set via: `CLICKUP_API_TOKEN` environment variable

2. **GitHub Token** (Required for GitHub features)
   - Get it from: <https://github.com/settings/tokens>
   - Permissions needed: `repo`, `read:org`, `workflow`
   - Set via: `GITHUB_TOKEN` environment variable

3. **CircleCI Token** (Optional, for CI features)
   - Get it from: <https://app.circleci.com/settings/user/tokens>
   - Permissions needed: Read-only access
   - Set via: `CIRCLECI_TOKEN` environment variable

4. **Claude API Token** (Optional, for AI features)
   - Get it from: <https://console.anthropic.com/>
   - Set via: `ANTHROPIC_API_KEY` environment variable

### Why am I getting "unauthorized" errors?

Common causes:

1. **Token not set**: Ensure environment variables are exported in your shell profile

   ```bash
   # Add to ~/.zshrc or ~/.bashrc
   export CLICKUP_API_TOKEN="your_token_here"
   export GITHUB_TOKEN="your_token_here"
   ```

2. **Token expired**: GitHub tokens can expire. Check <https://github.com/settings/tokens>

3. **Insufficient permissions**: Verify your token has required scopes
   - GitHub: `repo`, `read:org`, `workflow`
   - ClickUp: Full access to tasks and comments

4. **Wrong workspace**: Ensure your ClickUp token has access to the configured workspace

### How do I choose between GitHub modes?

vibecli offers three GitHub modes. See [GITHUB_MODES.md](GITHUB_MODES.md) for detailed comparison.

**API Mode:**

- Uses GitHub REST API directly with your token
- Best for CI/CD and automation
- Works everywhere

**CLI Mode:**

- Uses GitHub CLI (`gh`) under the hood
- Best for local development
- No token needed in config

**Auto Mode (Recommended):**

- Automatically uses CLI if available, falls back to API
- Best for most users

Configure in `.vibe.yaml`:

```yaml
github:
  mode: auto  # Options: "api", "cli", or "auto" (default)
```

### Can I configure default PR settings?

Limited PR configuration is available in `.vibe.yaml`:

```yaml
github:
  mode: auto
  # Note: draft_pr and auto_merge are not currently supported
  # PRs are created as regular (non-draft) by default
```

Git configuration affects branch names:

```yaml
git:
  base_branch: main           # Base branch for PRs (default: main)
  branch_prefix: "username"   # Prefix for ticket branches
```

### Can I read ClickUp custom fields?

Yes, vibecli can read custom fields from ClickUp tasks. Custom fields are automatically included when fetching task details with `vibe ticket`.

However, there is currently **no configuration to set or filter by custom fields**. Custom fields are read-only through the ClickUp API integration.

To view custom fields, use:

```bash
vibe ticket TICKET_ID
```

The output will include any custom fields like "Type", "Domain", "Priority", etc., if they're set on the ticket.

## Troubleshooting

### vibecli says "no ticket found" when I run commands

This happens when you're not on a ticket branch. Solutions:

1. **Check current branch**: `git branch --show-current`
   - Ticket branches must match the pattern: `username/ticket-id-description`

2. **Create a ticket branch**: `vibe workon TICKET_ID`

3. **Manually switch to ticket branch**: `git checkout username/abc123-feature`

4. **View ticket without branch**: `vibe ticket TICKET_ID`

### Getting "Status does not exist" error when running vibe workon

When running `vibe workon <ticket-id>`, you might see:

```
Failed to update task status: HTTP 400: Status does not exist
```

**Why this happens:**

The status name in your config doesn't exist in your ClickUp space. Status names vary by space/list and must be valid status names configured in ClickUp.

**Solutions:**

1. **Find a valid status name:**
   - Open any ticket in ClickUp
   - Note the available status names (e.g., "doing", "on deck", "in code review")
   - Update `~/.config/vibe/config.yaml`:

   ```yaml
   defaults:
     status: "on deck"  # Use a valid status name from your ClickUp space
   ```

2. **Common ClickUp statuses:**
   - `"prioritized"` (prioritized)
   - `"on deck"` (ready to start)
   - `"doing"` (active)
   - `"backlog"` (backlog)
   - `"complete"` (completed)

3. **Disable automatic status updates** (optional):

   ```yaml
   defaults:
     # status: "in progress"  # Commented out
   ```

**Note:** The command still works even if status update fails - it successfully creates/checks out your branch. The status update is a convenience feature.

### My PR description is not formatted correctly

Common issues:

1. **Markdown rendering**: GitHub may interpret certain characters differently
   - Use ` ``` ` for code blocks
   - Escape special characters when needed

2. **Template not found**: Ensure `.github/PULL_REQUEST_TEMPLATE.md` exists

3. **Custom template**: Specify in `.vibe.yaml`:

   ```yaml
   github:
     pr_template: docs/PR_TEMPLATE.md
   ```

### CircleCI status shows "no workflows found"

Possible causes:

1. **No CircleCI configuration**: Ensure `.circleci/config.yml` exists

2. **Branch not pushed**: Push your branch first

   ```bash
   git push -u origin your-branch
   ```

3. **CircleCI not triggered**: Check CircleCI project settings

4. **Token issues**: Verify `CIRCLECI_TOKEN` is set correctly

### Commands are slow or timing out

Performance tips:

1. **Use User mode for large repos**: Edit `.vibe.yaml`

   ```yaml
   github:
     mode: user
   ```

2. **Check network connection**: API calls require internet access

3. **Restart application**: Clears in-memory cache

   ```bash
   # All cached data is in-memory only
   # Simply restart vibe to clear cache
   ```

4. **Check API rate limits**:
   - GitHub: 5,000 requests/hour (authenticated)
   - ClickUp: Rate limits vary by plan

### Git operations fail with "repository not found"

Checklist:

1. **Verify remote**: `git remote -v`
2. **Check GitHub token permissions**: Must have `repo` scope
3. **Ensure repository access**: Your GitHub account must have access
4. **Try HTTPS vs SSH**: Update remote URL if needed

   ```bash
   git remote set-url origin https://github.com/user/repo.git
   ```

### Comments are not appearing in ClickUp

Debug steps:

1. **Check ticket ID**: Ensure you're on the correct ticket branch
2. **Verify token permissions**: ClickUp token needs comment write access
3. **Check API status**: Visit <https://status.clickup.com/>
4. **Try manual comment**: Test in ClickUp web interface
5. **Check command output**: Look for error messages

## Workflow Questions

### What's the recommended workflow for starting new work?

Standard workflow:

```bash
# 1. Start work on a ticket
vibe workon TICKET_ID

# 2. Make your changes
# ... edit files ...

# 3. Commit your work
git add .
git commit -m "Your commit message"

# 4. Push and create PR
git push -u origin $(git branch --show-current)
vibe pr create

# 5. Add comments to ticket as you work
vibe comment "Updated the user authentication flow"

# 6. Check CI status
vibe ci-status

# 7. When ready, merge
vibe merge
```

### How do I work on multiple tickets simultaneously?

Use Git branches to switch between tickets. Vibe automatically detects uncommitted changes and prompts you to stash them.

```bash
# Start first ticket
vibe workon TICKET_1
# ... work on ticket 1 ...

# Switch to second ticket (vibe will prompt to stash uncommitted changes)
vibe workon TICKET_2
# ⚠️  You have 3 uncommitted change(s)
# ? What would you like to do?
#   > Stash changes
#     Cancel
# ✓ Changes stashed successfully

# ... work on ticket 2 ...

# Switch back to ticket 1
vibe workon TICKET_1

# Restore stashed changes
git stash pop
git commit -m "Work in progress on ticket 2"

# Switch back to first ticket
git checkout username/ticket-1-description
# OR use vibe workon again
vibe workon TICKET_1
```

**Note:** Vibe automatically handles uncommitted changes by prompting you to stash them when switching branches. You can also manually stash changes:

```bash
# Manually stash if needed
git stash

# View stashed changes
git stash list

# Restore latest stash
git stash pop

# Restore specific stash
git stash apply stash@{1}
```

### What happens if I have uncommitted changes when switching branches?

Vibe automatically detects uncommitted changes and prompts you before switching branches:

```bash
vibe branch new-feature

⚠️  You have 3 uncommitted change(s)

? What would you like to do?
  > Stash changes
    Cancel

Stashing uncommitted changes...
✓ Changes stashed successfully
  (Use 'git stash pop' to restore them later)
```

**Behavior:**

- **Stash changes**: Your changes are safely stored and you can proceed with branch checkout
- **Cancel**: Operation is cancelled, you stay on current branch with uncommitted changes
- **Untracked files**: Not stashed automatically (only modified/added/deleted files are stashed)

**To restore stashed changes:**

```bash
git stash pop  # Apply and remove from stash
# or
git stash apply  # Apply but keep in stash
```

### Troubleshooting: Stash-Related Issues

#### What if stash fails?

If stashing fails, you'll see an error message with details. Common causes:

1. **Merge conflicts in working directory**

   ```bash
   # Resolve conflicts first, then try again
   git status  # Check conflicted files
   # Fix conflicts manually
   git add <resolved-files>
   ```

2. **Permission issues**

   ```bash
   # Check file permissions
   ls -la
   # Ensure you have write access to .git directory
   ```

3. **Corrupted git repository**

   ```bash
   # Verify repository integrity
   git fsck
   ```

#### How do I recover from a failed stash?

If something goes wrong during stashing:

```bash
# View all stashes
git stash list

# See what's in a specific stash
git stash show stash@{0}

# Apply a specific stash
git stash apply stash@{0}

# Drop a stash if needed
git stash drop stash@{0}

# Emergency: recover dropped stash (within ~90 days)
git fsck --unreachable | grep commit | cut -d' ' -f3 | xargs git log --oneline
```

#### What happens to merge conflicts in stashed changes?

When you apply a stash with conflicts:

```bash
# Apply stash
git stash pop

# If conflicts occur:
# 1. Git marks conflicted files
# 2. Resolve conflicts manually
# 3. Stage resolved files
git add <resolved-files>

# 4. Stash is NOT automatically dropped
# 5. Drop it manually after resolving
git stash drop
```

#### Can I stash specific files?

Yes, use git directly:

```bash
# Stash specific files
git stash push -m "Partial work" file1.go file2.go

# Stash everything except certain files
git stash push --keep-index

# Stash including untracked files
git stash push -u -m "Including untracked"
```

### Can I create a PR without a ClickUp ticket?

Yes, but you'll need to use Git directly:

```bash
# Create a branch manually
git checkout -b feature/my-feature

# Make changes and commit
git add .
git commit -m "Add new feature"

# Push branch
git push -u origin feature/my-feature

# Create PR using GitHub CLI or web interface
gh pr create --title "My Feature" --body "Description"
```

Note: vibecli PR commands work best with ticket-based branches.

### How do I handle PR review feedback?

```bash
# 1. Make requested changes
# ... edit files ...

# 2. Commit changes
git add .
git commit -m "Address review feedback"

# 3. Push updates
git push

# 4. Add comment to ticket
vibe comment "Addressed review feedback: updated error handling"

# 5. Update PR description if needed
vibe pr-update --body "Updated implementation based on feedback"
```

### What happens when I run `vibe start`?

`vibe start` provides an interactive way to start work on a ticket:

1. **With ticket ID**: `vibe start TICKET_ID` - Works like `vibe workon TICKET_ID`
2. **Without ticket ID**: Prompts you to enter a ticket ID or search term
3. Creates/switches to the ticket branch
4. Updates ticket status (configured in `defaults.status`)

**Note:** The search functionality is not yet fully implemented. Currently, you need to enter the ticket ID directly when prompted.

It's essentially an interactive version of `vibe workon TICKET_ID`.

## Integration Questions

### How does vibecli integrate with ClickUp?

vibecli uses the ClickUp API v2 to:

- Fetch ticket details and metadata
- Update ticket status
- Add comments and attachments
- Read custom fields
- Manage task assignments

Data flow:

1. You run a command (e.g., `vibe workon TICKET_ID`)
2. vibecli calls ClickUp API to fetch ticket data
3. Local branch is created with ticket information
4. Subsequent commands update ClickUp via API

### How does vibecli integrate with GitHub?

vibecli uses the GitHub REST API and GraphQL API to:

- Create and manage pull requests
- List and update issues
- Check CI status from GitHub Actions
- Manage labels, assignees, and milestones

Two modes available:

- **Repository mode**: Queries all repository PRs
- **User mode**: Queries only your PRs (faster for large repos)

### Does vibecli work with GitHub Actions?

**Limited support.** The `vibe pr-status` command can show GitHub Actions check status when viewing a PR (as part of the PR's status checks). However, the dedicated CI commands (`vibe ci-status` and `vibe ci-failure`) are **CircleCI-only**.

**What works:**

- `vibe pr-status` - Shows GitHub Actions checks as part of PR status

**What requires CircleCI:**

- `vibe ci-status` - CircleCI pipelines and workflows only
- `vibe ci-failure` - CircleCI job logs only

If you use GitHub Actions exclusively (no CircleCI), you can still use `vibe pr-status` to check CI status, but you won't have detailed CI logs like you get with CircleCI integration.

### Can I use vibecli with private repositories?

Yes, vibecli works with private repositories. Ensure:

1. Your `GITHUB_TOKEN` has `repo` scope (grants full repository access)
2. Your GitHub account has access to the repository
3. Git remote is configured correctly

### How does caching work?

vibecli implements **minimal in-memory caching**:

**What is cached**:

- **Sprint data**: 1 hour TTL for sprint folder lookups

**What is NOT cached**:

- ClickUp tasks, workspaces, lists
- GitHub PRs, issues
- CircleCI workflows, jobs

**Cache behavior**:

- All caching is in-memory only (no files on disk)
- Cache is lost when application exits
- No persistent cache to manage

**Clear cache**:

```bash
# All cache is in-memory only
# Simply restart the application to clear cache
```

**Future Enhancement**: Per-service caching with configurable TTLs and `--no-cache` flag may be added in the future.

## Performance Questions

### Why is the first command sometimes slower?

The first command may need to:

1. Fetch workspace and task data from ClickUp
2. Load repository information from GitHub
3. Retrieve user and organization details

Since vibecli uses minimal caching (sprint data only), most commands make fresh API calls each time. Performance depends primarily on:

- Network latency to API services
- API response times
- Whether GitHub CLI mode is being used (can be faster for local development)

### How can I make vibecli faster?

Optimization tips:

1. **Use CLI mode** for faster GitHub operations:

   ```yaml
   # .vibe.yaml
   github:
     mode: cli  # or "auto" to use CLI when available
   ```

2. **Use ticket branches**: Commands are faster when context is known

3. **Reduce concurrent API calls**: Wait for operations to complete

4. **Update regularly**: New versions include performance improvements

### Does vibecli work offline?

Limited functionality:

- **Works offline**: Git operations, reading cached data
- **Requires internet**: ClickUp, GitHub, CircleCI API calls

When offline:

- Cannot fetch new ticket data
- Cannot create or update PRs
- Cannot check CI status
- Can still commit and work locally

### How many API calls does a typical workflow use?

Example workflow API usage:

```bash
vibe workon TICKET_ID     # 2-3 API calls (ClickUp)
vibe pr create            # 3-5 API calls (GitHub)
vibe comment "message"    # 1 API call (ClickUp)
vibe ci-status            # 2-3 API calls (CircleCI)
vibe merge                # 2-3 API calls (GitHub)
```

Total: ~10-15 API calls per complete workflow

**Note**: vibecli uses minimal caching, so most commands make fresh API calls each time.

## Additional Resources

- [README.md](README.md) - Installation and basic usage
- [GITHUB_MODES.md](GITHUB_MODES.md) - Detailed GitHub mode comparison
- [SKILLS.md](SKILLS.md) - Claude Code skills documentation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contributing guidelines
- [SECURITY.md](SECURITY.md) - Security best practices

## Still Have Questions?

- **Bug reports**: [GitHub Issues](https://github.com/vibecli/vibecli/issues)
- **Feature requests**: [GitHub Discussions](https://github.com/vibecli/vibecli/discussions)
- **Documentation improvements**: Submit a PR!
