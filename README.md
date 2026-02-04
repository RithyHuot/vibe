# vibe

A production-quality Go CLI tool that streamlines developer workflow by integrating ClickUp (project management), GitHub (code repository), and CircleCI (CI/CD).

## Features

### Ticket Management

- 🎫 **Fetch Tasks**: Get task details from ClickUp with full metadata
- 📝 **Status Updates**: Automatically update task status when starting work
- 💬 **Comments**: Add comments to ClickUp tasks from the terminal
- 🔍 **Interactive Selection**: Browse and select tickets from your workspace
- 🎯 **Sprint Detection**: Smart sprint folder identification with date parsing

### Git & Branch Management

- 🌿 **Auto Branch Creation**: Generate standardized branches: `username/ticketid/description`
- ✅ **Branch Validation**: Security checks for safe branch names
- 🔄 **Branch Switching**: Seamlessly switch between feature branches
- 💾 **Smart Stashing**: Automatically prompts to stash uncommitted changes before checkout
- 📊 **Status Tracking**: View working tree status and changes

### Pull Requests

- 🚀 **PR Creation**: Interactive and non-interactive PR creation
- 📋 **Template Support**: Auto-populate from `.github/PULL_REQUEST_TEMPLATE.md`
- 🤖 **AI Descriptions**: Generate PR descriptions from git diff using Claude
- 👀 **Status Monitoring**: Track reviews, CI checks, and merge readiness
- ✏️ **PR Updates**: Edit titles and descriptions with section-aware updates
- 🔀 **Merge Automation**: Trigger merge via `/merge` comments

### Issue Management

- 📋 **List Issues**: Browse and filter issues by state (open, closed, all)
- 🔍 **Interactive Selection**: Select issues from list to view full details
- 🌿 **Branch Creation**: Create branches directly from issue view to start working
- 📝 **Issue Creation**: Create issues with metadata (labels, assignees, milestone, projects)
- 📋 **Template Support**: Auto-populate from `.github/ISSUE_TEMPLATE.md`
- ✏️ **Issue Updates**: Modify title, description, state, and metadata
- 🏷️ **Full Metadata**: Support for assignees, labels, milestones, and GitHub Projects
- 🔢 **Branch Integration**: Auto-detect issue numbers from branch names
- 💬 **Comments**: View issue comments with `--comments` flag

### CI/CD Integration

- 🔄 **CircleCI Monitoring**: Real-time pipeline and workflow status
- ❌ **Failure Analysis**: Detailed error logs and test failure reports
- 🎨 **Visual Indicators**: Color-coded status display
- 📊 **Test Results**: View failed tests with error messages

### AI Features

- 🤖 **Claude Integration**: Dual support for Claude API and CLI
- 📝 **Smart Descriptions**: AI-generated PR descriptions from code changes
- 🔍 **Code Review**: Comprehensive review for bugs, security, performance, and best practices
- 🎯 **Context Analysis**: Analyze git diffs for comprehensive summaries
- 💡 **Interactive Opt-in**: Choose when to use AI features

### Developer Experience

- 🎨 **Rich UI**: Colors, spinners, tables, and formatted output
- ⚡ **Fast Performance**: In-memory caching with TTL
- 🔒 **Secure**: Input sanitization and branch name validation
- 🐛 **Debug Mode**: Detailed HTTP request logging
- 📦 **Zero Config**: Sensible defaults with optional customization

## Quick Start at a Glance

Get up and running with vibe in under 5 minutes:

```bash
# 1. Install
go install github.com/rithyhuot/vibe/cmd/vibe@latest

# 2. Initialize configuration
vibe init

# 3. Set up API tokens (edit ~/.config/vibe/config.yaml)
# - Get ClickUp token: https://app.clickup.com/settings/apps
# - Get GitHub token: https://github.com/settings/tokens

# 4. Start working on a ticket
vibe workon abc123          # Replace abc123 with your ticket ID

# 5. Make your changes, then create a PR
git add .
git commit -m "Your changes"
git push
vibe pr create

# 6. Check CI status
vibe ci-status

# 7. Merge when ready
vibe merge
```

**Common workflows:**

- `vibe ticket` - View current ticket details
- `vibe comment "message"` - Add comment to ticket
- `vibe pr-status` - Check PR approval and CI status
- `vibe issues` - Browse GitHub issues
- `vibe start` - Interactive ticket selection

See the full [Usage Guide](#commands) below for detailed examples and options.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/rithyhuot/vibe.git
cd vibe

# Build
make build

# Install
make install
```

### Using Go

```bash
go install github.com/rithyhuot/vibe/cmd/vibe@latest
```

## Updating

### Update from Source

If you installed from source, pull the latest changes and rebuild:

```bash
cd vibe
git pull origin main
make build
make install
```

### Update from Go Install

Simply re-run the install command to get the latest version:

```bash
go install github.com/rithyhuot/vibe/cmd/vibe@latest
```

### Check Version

Verify your installed version:

```bash
vibe --version
```

## Uninstalling

### Remove Binary

The `vibe` binary is installed in your Go bin directory. Remove it with:

```bash
rm $(which vibe)
```

Or manually from the Go bin directory:

```bash
# Find your Go bin directory
go env GOPATH

# Remove the binary (typically in $GOPATH/bin or $GOBIN)
rm $GOPATH/bin/vibe
# or
rm ~/go/bin/vibe
```

### Remove Configuration

To also remove your configuration file:

```bash
rm -rf ~/.config/vibe/
```

### Remove Source (if installed from source)

If you cloned the repository and want to remove it:

```bash
# Navigate to parent directory
cd ..
rm -rf vibe
```

### Complete Cleanup

Remove everything related to vibe:

```bash
# Remove binary
rm $(which vibe)

# Remove configuration
rm -rf ~/.config/vibe/

# Remove source (if applicable)
rm -rf ~/path/to/vibe

# Clear Go cache (optional)
go clean -cache -modcache -i github.com/rithyhuot/vibe/...
```

## Claude Code Integration

The vibe CLI includes skills for [Claude Code](https://claude.ai/code) that enable seamless integration with AI-assisted development workflows. These skills allow Claude Code to interact with your ClickUp tickets, GitHub PRs, and CI/CD pipelines.

### Installing Skills Globally

To make vibe skills available in **all your projects**, install them globally during initialization:

```bash
vibe init --install-skills
```

Or install them separately anytime:

```bash
vibe skills
```

This copies all skills to `~/.claude/skills/` making them available to Claude Code in every project you work on.

### Updating Skills

When you update the vibe CLI to a new version, update your skills to get the latest features:

```bash
vibe skills
```

This overwrites the existing skills with the latest versions embedded in your binary.

### Uninstalling Skills

To remove vibe skills from Claude Code:

```bash
vibe skills --uninstall
```

This removes all vibe skills from `~/.claude/skills/`. You can reinstall them anytime with `vibe skills`.

### Available Skills

| Skill | Description | Usage |
|-------|-------------|-------|
| `vibe` | Start work on a ClickUp ticket | "vibe 86b7x5453" |
| `vibe-workon` | Start work on a ClickUp ticket (explicit) | "vibe workon 86b7x5453" |
| `vibe-branch` | Create and checkout a new branch | "create a branch" or "vibe branch abc123xyz" |
| `vibe-ticket` | Get context on current ticket | "what am I working on?" |
| `vibe-comment` | Add comment to ticket | "vibe comment <text>" |
| `vibe-pr` | Create a pull request | "create a PR" |
| `vibe-pr-status` | Check PR status | "check PR status" |
| `vibe-pr-update` | Update PR description | "update PR description" |
| `vibe-merge` | Merge a pull request | "merge the PR" |
| `vibe-ci-status` | Check CircleCI status | "check CI status" |
| `vibe-issues` | List GitHub issues | "list all issues" |
| `vibe-issue` | View issue details | "show issue #123" |
| `vibe-issue-create` | Create a new issue | "create an issue" |
| `vibe-issue-update` | Update existing issue | "close issue #123" |
| `vibe-code-review` | Perform comprehensive code review | "review my code" |
| `vibe-dependabot-review` | Review Dependabot PRs and create fixes | "review dependabot PR 123" |
| `add-claude-skill` | Add a new Claude Code skill for vibe CLI | "add a skill for code review workflow" |
| `add-command-skill` | Add a new vibe command with associated skill | "add a new command called status" |

### Using Skills with Claude Code

Once installed globally, Claude Code can use these skills in any project:

1. **Start work on tickets**: Automatically create branches and fetch ticket context
2. **Create PRs**: Generate comprehensive PR descriptions from your changes
3. **Review code**: Comprehensive code review for bugs, security, and best practices
4. **Manage issues**: Browse, create, and update GitHub issues
5. **Monitor CI**: Check build status and analyze failures
6. **Update tickets**: Add comments and track progress

The skills integrate seamlessly with the vibe CLI commands, providing an AI-enhanced development workflow.

### Learn More

See [SKILLS.md](SKILLS.md) for detailed documentation on:

- How each skill works
- When to use each skill
- Workflow examples
- Best practices
- Troubleshooting

### Project-Specific Skills

If you prefer project-specific skills instead of global installation, you can manually copy the `skills/` directory from this repository to any project. Each skill contains a `SKILL.md` file that defines how Claude Code should use the vibe CLI.

## Quick Start

### 1. Initialize Configuration

```bash
vibe init
```

This creates a configuration file at `~/.config/vibe/config.yaml` and optionally sets up shell autocomplete.

**Note:** You can run `vibe init` multiple times safely. If the config already exists, it will skip creation and continue with other setup steps like skills installation and shell completion.

### 2. Configure Your Settings

Edit the config file with your credentials. The configuration file is located at `~/.config/vibe/config.yaml`.

#### Required Configuration

```yaml
# ClickUp Configuration (REQUIRED)
clickup:
  api_token: "pk_your_token"        # REQUIRED: Get from https://app.clickup.com/settings/apps
  user_id: "12345678"               # REQUIRED: Your ClickUp user ID
  workspace_id: "1234567"           # REQUIRED: Your workspace ID
  team_id: "1234567"                # REQUIRED: Your team ID (often same as workspace_id)

# GitHub Configuration (REQUIRED for PR/issue features)
github:
  token: "ghp_your_token"           # REQUIRED for API mode: Get from https://github.com/settings/tokens
                                    # Permissions needed: repo, read:org, workflow
  username: "your-username"         # REQUIRED: Your GitHub username
  owner: "org-name"                 # REQUIRED: GitHub org or user that owns the repo
  repo: "repo-name"                 # REQUIRED: Repository name
  mode: "auto"                      # OPTIONAL: "api", "cli", or "auto" (default: auto)

# Git Configuration (REQUIRED)
git:
  branch_prefix: "your-username"    # REQUIRED: Prefix for branch names (usually your username)
  base_branch: "main"               # OPTIONAL: Default branch (default: main)

# CircleCI Configuration (OPTIONAL - only needed for CI features)
circleci:
  api_token: "your_circleci_token"  # OPTIONAL: Get from https://app.circleci.com/settings/user/tokens
                                    # Only needed if you use 'vibe ci-status' or 'vibe ci-failure'

# Workspace Configuration (OPTIONAL - for sprint detection)
workspaces:
  - name: "Engineering"             # OPTIONAL: Workspace name for reference
    folder_id: "123456789"          # OPTIONAL: Folder ID for sprint detection
    sprint_patterns:                # OPTIONAL: Regex patterns to identify sprint folders
      - "Sprint \\d+ \\("

# Default Settings (OPTIONAL - sensible defaults provided)
defaults:
  status: "doing"             # OPTIONAL: Status to set when starting work
                                    # Must be a valid status name in your ClickUp space
                                    # Common values: "doing", "on deck", "backlog"
                                    # Comment out to disable automatic status updates
```

#### Environment Variables (Alternative to config file)

You can also use environment variables instead of storing tokens in the config file:

```bash
# Add to ~/.zshrc or ~/.bashrc
export CLICKUP_API_TOKEN="pk_your_token"
export GITHUB_TOKEN="ghp_your_token"
export CIRCLECI_TOKEN="your_circleci_token"
export ANTHROPIC_API_KEY="your_claude_api_key"  # For AI features
```

**Note:** Environment variables take precedence over config file values.

#### GitHub Mode Options

vibe supports two methods for GitHub integration:

**API Mode** (default if you have a token):

- Uses GitHub REST API directly
- Requires a personal access token
- Works everywhere, including CI/CD

**CLI Mode** (if you have `gh` CLI installed):

- Uses the GitHub CLI (`gh`) under the hood
- No token needed in config (uses `gh auth` credentials)
- Simpler authentication
- Respects your `gh` CLI settings

**Auto Mode** (recommended):

- Automatically chooses CLI if available, falls back to API
- Set `mode: "auto"` (default) to use this

To use CLI mode exclusively:

```yaml
github:
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "cli"
```

Make sure you're authenticated with the CLI:

```bash
gh auth login
gh auth status  # Verify authentication
```

#### Getting API Tokens

- **ClickUp**: <https://app.clickup.com/settings/apps>
- **GitHub**: <https://github.com/settings/tokens> (needs `repo` scope)
- **CircleCI**: <https://app.circleci.com/settings/user/tokens> (optional)
- **Claude**: <https://console.anthropic.com/> (optional, for AI features)

### 3. Enable Shell Autocomplete (Optional but Recommended)

For command-line autocomplete with TAB completion:

```bash
# Zsh (macOS default)
mkdir -p ~/.zsh/completions
vibe completion zsh > ~/.zsh/completions/_vibe

# Add to ~/.zshrc (if not already present):
# fpath=(~/.zsh/completions $fpath)
# autoload -U compinit && compinit

exec zsh

# Bash (requires bash-completion to be installed)
vibe completion bash > $(brew --prefix)/etc/bash_completion.d/vibe

# See 'vibe completion --help' for other shells
```

### 4. Start Working on a Ticket

```bash
# Start working on a ticket by ID
vibe abc123xyz

# This will:
# 1. Fetch the task from ClickUp
# 2. Create a branch: username/abc123xyz/task-name
# 3. Checkout the branch
# 4. Update task status (configured in defaults.status)
```

## Commands

### `vibe init`

Initialize configuration file and set up vibe.

```bash
# Create global config file
vibe init

# Create local project override config
vibe init --local

# Create config and install skills globally
vibe init --install-skills

# Overwrite existing config file
vibe init --force

# Combine flags
vibe init --force --install-skills

# Run init again to install skills or set up completion (skips existing config)
vibe init --install-skills
```

**Options:**

- `-l, --local`: Create local `.vibe.yaml` override file in current directory
- `--install-skills`: Install Claude Code skills globally to `~/.claude/skills/`
- `-f, --force`: Overwrite existing configuration file without prompting

**Behavior with Existing Config:**

You can run `vibe init` multiple times. If a config file already exists:

- **Interactive mode**: Prompts you to overwrite or skip, then continues with skills/completion setup
- **Non-interactive mode**: Skips config creation, then continues with skills/completion setup
- **With `--force` flag**: Overwrites config without prompting, then continues with setup

This allows you to:

- Re-run init to install skills: `vibe init --install-skills`
- Re-run init to set up shell completion: `vibe init`
- Update config without affecting skills: `vibe init --force`

### `vibe completion`

Generate shell completion scripts for command-line autocomplete.

```bash
# Generate completion script for your shell
vibe completion [bash|zsh|fish|powershell]
```

#### Setup Instructions

**Zsh (macOS default):**

```bash
# One-time setup
mkdir -p ~/.zsh/completions
vibe completion zsh > ~/.zsh/completions/_vibe

# Add to ~/.zshrc (if not already present):
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit && compinit' >> ~/.zshrc

# Restart shell
exec zsh
```

**Bash (requires bash-completion package):**

```bash
# macOS with Homebrew (requires: brew install bash-completion)
vibe completion bash > $(brew --prefix)/etc/bash_completion.d/vibe

# Linux (requires bash-completion package)
sudo vibe completion bash > /etc/bash_completion.d/vibe

# Or install to user directory (no sudo):
mkdir -p ~/.bash_completion.d
vibe completion bash > ~/.bash_completion.d/vibe
echo 'source ~/.bash_completion.d/vibe' >> ~/.bashrc

# Restart your terminal
```

**Fish:**

```bash
vibe completion fish > ~/.config/fish/completions/vibe.fish
```

**PowerShell:**

```powershell
vibe completion powershell > vibe.ps1
# Source this file from your PowerShell profile
```

#### Quick Test (Temporary)

Try autocomplete immediately without installation:

```bash
# Zsh (requires compinit to be loaded first)
autoload -U compinit && compinit
source <(vibe completion zsh)

# Bash (requires bash-completion package)
source <(vibe completion bash)
```

Once installed, you can:

- Tab-complete commands: `vibe <TAB>`
- Tab-complete subcommands: `vibe pr-<TAB>`
- See available options when you press TAB

### `vibe skills`

Manage Claude Code skills (install, update, or uninstall).

```bash
# Install or update skills globally
vibe skills

# Uninstall skills
vibe skills --uninstall

# Check if skills are installed
ls ~/.claude/skills/
```

Skills are installed to `~/.claude/skills/` and are available in all your projects.

### `vibe <ticket-id>`

Start working on a ticket. Fetches the task, creates a branch, and updates status.

```bash
vibe abc123xyz
```

### `vibe ticket [ticket-id]`

View ticket details. If no ticket ID is provided, uses the current branch.

```bash
# View current ticket
vibe ticket

# View specific ticket
vibe ticket abc123xyz
```

### `vibe comment <text>`

Add a comment to the current ticket.

```bash
# From command line
vibe comment "Fixed the bug"

# From stdin
echo "Implemented feature" | vibe comment
```

### `vibe start`

Interactively select and start working on a ticket.

```bash
vibe start
```

This command will:

1. Show you a list of available tickets
2. Let you select a ticket interactively
3. Create and checkout a branch
4. Update the ticket status

### `vibe branch [ticket-id]`

Create and checkout a new branch with or without a ticket ID.

```bash
# Create branch with ticket ID
vibe branch abc123xyz

# Interactive mode - prompts for branch description
vibe branch
```

This command creates simple branches without ClickUp integration:

- **With ticket ID**: Creates branch in format `username/ticketid`
- **Without ticket ID**: Prompts for description and creates `username/description`
- Username is automatically extracted from `git config user.name`
- No ClickUp API call is made (unlike `vibe <ticket-id>`)
- **Uncommitted changes**: Automatically prompts to stash before checkout

**Comparison with `vibe <ticket-id>`:**

| Command | ClickUp Integration | Branch Format | Status Update |
|---------|---------------------|---------------|---------------|
| `vibe branch <ticket-id>` | No | `username/ticketid` | No |
| `vibe <ticket-id>` | Yes | `username/ticketid/task-name` | Yes (to "In Progress") |

### `vibe pr`

Create a pull request for the current branch.

```bash
# Interactive mode (recommended)
vibe pr

# Non-interactive with flags
vibe pr --title "Add feature" --summary "Summary" --yes

# Create as draft
vibe pr --draft

# Use AI to generate description
vibe pr --ai
```

### `vibe pr-status [pr-number]`

Check the status of a pull request.

```bash
# Check current branch's PR
vibe pr-status

# Check specific PR
vibe pr-status 123
```

### `vibe pr-update [pr-number]`

Update a pull request's title or description.

```bash
# Update current branch's PR interactively
vibe pr-update

# Update specific PR
vibe pr-update 123
```

### `vibe issues`

List GitHub issues with optional filtering.

```bash
# List open issues (default)
vibe issues

# List closed issues
vibe issues --state closed

# List all issues
vibe issues --state all

# Interactive selection mode
vibe issues --select

# Limit number of issues
vibe issues --limit 50
```

**Options:**

- `--state`: Filter by state (`open`, `closed`, `all`). Default: `open`
- `--limit`: Maximum number of issues to display. Default: `30`
- `-s, --select`: Enable interactive selection to view full issue details

In interactive mode (`--select`), you can browse the list and select an issue to view its complete details including description, labels, assignees, and comments. After viewing, you'll be prompted to optionally create a branch for the issue to start working on it immediately.

### `vibe issue [issue-number]`

View GitHub issue details.

```bash
# View specific issue
vibe issue 123

# View with comments
vibe issue 123 --comments

# Auto-detect from branch name
vibe issue
```

**Options:**

- `-c, --comments`: Include comments in output

**Branch name patterns:**

If no issue number is provided, the command attempts to extract it from the current branch name:

- `issue-123` → 123
- `123-fix-bug` → 123
- `username/issue-123/description` → 123
- `fix-issue-456` → 456

**Output includes:**

- Issue number, title, and URL
- State (OPEN/CLOSED)
- Author and assignees
- Labels and milestone
- Projects (if assigned)
- Timestamps (created, updated, closed)
- Full description
- Comments (if `--comments` flag used)

**Branch creation:**

After viewing an issue, you'll be prompted with the option to create a branch for that issue. This allows you to:

- Quickly start working on an issue
- Automatically generate a standardized branch name (e.g., `username/issue-123/title-slug`)
- Check out existing branches if they already exist

### `vibe issue-create`

Create a new GitHub issue.

```bash
# Interactive mode (prompts for all fields)
vibe issue-create

# Non-interactive mode
vibe issue-create --yes --title "Bug: Login fails" --body "Description..."

# With metadata
vibe issue-create --yes \
  --title "Add dark mode" \
  --body-file feature.md \
  --labels feature,enhancement \
  --assignees user1,user2 \
  --milestone "v2.0" \
  --projects "Project Name"
```

**Options:**

- `--title`: Issue title (required in non-interactive mode)
- `--body`: Issue description
- `--body-file`: Read description from file
- `--assignees`: Comma-separated list of assignees
- `--labels`: Comma-separated list of labels
- `--milestone`: Milestone name
- `--projects`: Comma-separated list of project names
- `-y, --yes`: Skip confirmation prompts

**Features:**

- Template support: Auto-loads from `.github/ISSUE_TEMPLATE.md`
- Interactive mode: Uses editor for description
- Preview: Shows what will be created before confirmation
- Validation: Ensures title is provided

### `vibe issue-update <issue-number>`

Update an existing GitHub issue.

```bash
# Close an issue
vibe issue-update 123 --state closed

# Reopen an issue
vibe issue-update 123 --state open

# Update title and description
vibe issue-update 123 --title "New title" --body "Updated description"

# Update assignees (replaces existing)
vibe issue-update 123 --assignees user1,user2

# Update labels (replaces existing)
vibe issue-update 123 --labels bug,urgent,priority

# Set milestone
vibe issue-update 123 --milestone "v1.0"

# Combine multiple updates
vibe issue-update 123 \
  --state closed \
  --labels fixed,verified \
  --assignees rithyhuot
```

**Options:**

- `--title`: Update issue title
- `--body`: Update issue description
- `--state`: Change state (`open` or `closed`)
- `--assignees`: Update assignees (replaces existing)
- `--labels`: Update labels (replaces existing)
- `--milestone`: Set or change milestone
- `--projects`: Update projects (replaces existing)

**Important:**

- At least one flag is required
- State must be either "open" or "closed"
- **CLI Mode Note**: When using `gh` CLI, assignees and labels are additive. For true replacement behavior, use API mode (`github.mode: "api"` in config)
- In API mode: Assignees, labels, and projects **replace** existing values
- Issue number is required as an argument

### `vibe merge [pr-number]`

Post a `/merge` comment to trigger merge automation.

```bash
# Merge current branch's PR
vibe merge

# Merge specific PR
vibe merge 123
```

### `vibe ci-status [branch]`

Check CircleCI status for a branch.

```bash
# Check current branch
vibe ci-status

# Check specific branch
vibe ci-status feature/my-feature
```

### `vibe ci-failure [job-number]`

View detailed failure logs from a failed CI job.

```bash
# Show first failed job
vibe ci-failure

# Show specific job
vibe ci-failure 12345
```

## Configuration

### Full Configuration Example

```yaml
clickup:
  api_token: "pk_xxx"
  user_id: "12345678"
  workspace_id: "1234567"
  team_id: "1234567"

github:
  token: "ghp_xxx"  # Optional if using CLI mode
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "auto"  # Options: "api", "cli", or "auto"

git:
  branch_prefix: "your-username"
  base_branch: "main"

circleci:
  api_token: "circle_xxx"

claude:
  api_key: "sk-ant-xxx"

workspaces:
  - name: "Engineering"
    folder_id: "123456789"
    sprint_patterns:
      - "Sprint \\d+ \\("

defaults:
  status: "in progress"  # Must be a valid status in your ClickUp space

ai:
  enabled: true
  generate_descriptions: true

ui:
  color_enabled: true
```

### GitHub Configuration Details

vibe supports two methods for GitHub integration:

#### API Mode

Uses GitHub REST API directly with a personal access token.

```yaml
github:
  token: "ghp_xxx"
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "api"
```

**Pros:**

- Works in any environment (local, CI/CD, scripts)
- No external dependencies
- Fine-grained control over permissions

**Cons:**

- Requires creating and managing a personal access token
- Token needs `repo` scope

Get your token from: <https://github.com/settings/tokens>

#### CLI Mode

Uses the GitHub CLI (`gh`) for all GitHub operations.

```yaml
github:
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "cli"
```

**Pros:**

- No token needed in config file
- Uses your existing `gh auth` credentials
- Simpler setup and authentication
- Respects your gh CLI configuration

**Cons:**

- Requires `gh` CLI to be installed and authenticated
- May not work in some CI/CD environments

Setup:

```bash
# Install gh CLI (if not already installed)
brew install gh  # macOS
# or see: https://cli.github.com/

# Authenticate
gh auth login

# Verify
gh auth status
```

#### Auto Mode (Recommended)

Automatically detects the best method:

```yaml
github:
  token: "ghp_xxx"  # Optional
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "auto"  # Default if not specified
```

**How it works:**

1. If `gh` CLI is available and authenticated → uses CLI mode
2. If token is configured → uses API mode
3. If neither → shows error with setup instructions

This gives you the best of both worlds - use CLI locally for convenience, and API in CI/CD.

### AI Configuration Details

vibe supports AI-powered features using Claude. You have two options:

**Option 1: Claude API (Recommended)**

```yaml
claude:
  api_key: "sk-ant-xxx"
```

Get your API key from: <https://console.anthropic.com/>

**Option 2: Claude CLI (Automatic Fallback)**

If you have the Claude CLI installed, vibe will automatically detect and use it when no API key is configured.

**AI Features:**

- **PR Description Generation**: Analyzes your git diff and generates comprehensive descriptions
- **Smart Summaries**: Creates bullet-pointed summaries of changes
- **Context-Aware**: Understands code structure and suggests appropriate testing notes
- **Interactive**: Always gives you a chance to edit AI-generated content

**Disable AI:**

```yaml
ai:
  enabled: false
```

Or use flags:

```bash
vibe pr --no-ai  # Disable for a single command
```

### Environment Variables

You can override configuration values using environment variables:

```bash
export VIBE_CLICKUP_TOKEN="pk_xxx"
export VIBE_GITHUB_TOKEN="ghp_xxx"
export VIBE_CIRCLECI_TOKEN="circle_xxx"
export VIBE_CLAUDE_API_KEY="sk-ant-xxx"
```

### Custom Config Location

```bash
vibe --config /path/to/config.yaml <command>
```

### Local Config Overrides

You can create a `.vibe.yaml` file in your project directory to override global settings for that project only. This is useful when:

- Working on multiple repositories with different GitHub orgs/repos
- Using different settings per project (branch prefix, git settings, etc.)
- Temporarily disabling AI features for specific projects

**Creating a local config:**

```bash
# Create a local .vibe.yaml with example overrides
vibe init --local

# Or create manually in your project root
touch .vibe.yaml
```

**Example `.vibe.yaml` (only include fields you want to override):**

```yaml
# Override GitHub repo for this project
github:
  owner: "different-org"
  repo: "different-repo"

# Override git settings
git:
  branch_prefix: "feature"
  base_branch: "develop"

# Use CLI mode for GitHub in this project
github:
  mode: "cli"

# Disable AI features for this project
ai:
  enabled: false
```

**How it works:**

1. Settings are loaded in priority order:
   - CLI flags (highest)
   - Local `.vibe.yaml` (project-specific)
   - Global `~/.config/vibe/config.yaml`
   - Environment variables
   - Defaults (lowest)

2. Only include the fields you want to override - you don't need the full config structure

3. The local config merges on top of the global config, so you only need to specify what changes

**Common use cases:**

```yaml
# Working on an open-source project with a different repo
github:
  owner: "opensource-org"
  repo: "project-name"

# Using a feature branch workflow
git:
  base_branch: "develop"

# Project with different workspace
workspaces:
  - name: "Client Project"
    folder_id: "987654321"
    sprint_patterns:
      - "Sprint \\d+"
```

**Note:** Add `.vibe.yaml` to your `.gitignore` if it contains sensitive overrides, or commit it to share project-specific settings with your team.

### Debug Mode

Enable debug logging for HTTP requests:

```bash
export VIBE_DEBUG=true
vibe <command>
```

## Workflows

### Complete Feature Development

```bash
# 1. Start working on a ticket
vibe abc123xyz
# Creates branch: username/abc123xyz/feature-name
# Updates ticket status (per config)

# 2. Make your changes
# ... code, commit, etc ...

# 3. Add a progress comment
vibe comment "Implemented core functionality, working on tests"

# 4. Create a pull request with AI
vibe pr --ai
# AI analyzes your changes and generates a comprehensive description
# You can edit before submitting

# 5. Check CI status
vibe ci-status
# Shows pipeline, workflow, and job status
# Displays any test failures

# 6. View failure details if needed
vibe ci-failure
# Shows full error logs from failed jobs

# 7. Update PR after changes
vibe pr-update
# Edit title or description sections

# 8. Check PR status
vibe pr-status
# Shows reviews, CI checks, and merge readiness

# 9. Merge when ready
vibe merge
# Posts /merge comment to trigger automation
```

### Quick Bug Fix

```bash
# Start from ticket ID
vibe xyz789abc

# Make your fix
git add .
git commit -m "Fix: resolve null pointer exception"

# Quick PR
vibe pr --title "Fix null pointer in login" --yes

# Monitor until green
vibe ci-status
```

### Interactive Workflow

```bash
# Browse and select a ticket
vibe start

# Work on it...

# Create PR interactively
vibe pr
# Interactive prompts guide you through:
# - Title
# - Summary
# - Description
# - Testing notes
# - Draft vs ready
# - AI enhancement
```

## Development

### Prerequisites

- Go 1.24 or higher
- Git
- GitHub CLI (`gh`) for some features

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Lint

```bash
make lint
```

### Format

```bash
make fmt
```

### Security Scan

```bash
make vulncheck

# Control output options (traces, color, version, verbose)
VULNCHECK_SHOW=color make vulncheck
```

### Run All Checks

```bash
make check
```

### Pre-PR Checks

```bash
make pre-pr
```

### View All Available Commands

```bash
make help
```

### Install Locally

```bash
make install
```

## Project Structure

```
vibe/
├── cmd/vibe/           # Main entry point
│   └── main.go
├── internal/
│   ├── commands/         # Command implementations
│   ├── services/         # External integrations
│   │   ├── clickup/      # ClickUp API client
│   │   ├── github/       # GitHub operations
│   │   ├── circleci/     # CircleCI monitoring
│   │   ├── claude/       # Claude API/CLI client
│   │   └── git/          # Git operations
│   ├── config/           # Configuration
│   ├── models/           # Data models
│   ├── ui/               # UI utilities
│   └── utils/            # Utilities
├── skills/               # Claude Code integration skills
│   ├── vibe/
│   ├── vibe-workon/
│   ├── vibe-branch/
│   ├── vibe-ticket/
│   ├── vibe-comment/
│   ├── vibe-pr/
│   ├── vibe-pr-status/
│   ├── vibe-pr-update/
│   ├── vibe-merge/
│   ├── vibe-ci-status/
│   ├── vibe-issues/
│   ├── vibe-issue/
│   ├── vibe-issue-create/
│   ├── vibe-issue-update/
│   ├── vibe-code-review/
│   ├── vibe-dependabot-review/
│   ├── add-claude-skill/
│   └── add-command-skill/
├── Makefile
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Additional Documentation

### User Documentation

- [FAQ](FAQ.md) - Frequently asked questions and troubleshooting
- [GitHub Integration Modes](GITHUB_MODES.md) - Detailed guide on API vs CLI mode
- [Claude Code Skills](SKILLS.md) - Using skills with Claude Code
- [Security Policy](SECURITY.md) - Security best practices and vulnerability reporting

### Developer Documentation

- [Architecture](ARCHITECTURE.md) - System design, patterns, and component interactions
- [Contributing](CONTRIBUTING.md) - How to contribute to the project
- [Release Process](RELEASE.md) - Release and versioning information

## License

MIT

## Troubleshooting

### Common Issues

#### "Failed to load configuration"

**Problem**: Configuration file is missing or invalid.

**Solution**:

```bash
# Initialize configuration
vibe init

# Edit with your credentials
vim ~/.config/vibe/config.yaml
```

#### "Failed to open repository"

**Problem**: Not in a Git repository.

**Solution**:

```bash
# Make sure you're in a git repo
git status

# Or initialize one
git init
```

#### "Could not extract ticket ID from branch"

**Problem**: Current branch doesn't follow the expected format.

**Solution**:

- Expected format: `prefix/ticketid/description`
- Example: `john/abc123xyz/add-feature`
- Use `vibe <ticket-id>` to create a properly formatted branch

#### PR Creation Fails

**Problem**: GitHub authentication or permissions issue.

**Solution**:

**If using CLI mode:**

```bash
# Test GitHub CLI authentication
gh auth status

# Re-authenticate if needed
gh auth login

# Verify the repository
gh repo view
```

**If using API mode:**

```bash
# Verify token has 'repo' scope in GitHub settings
# Check: https://github.com/settings/tokens

# Test the token
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.github.com/user
```

**Switch modes if needed:**

```yaml
# In config.yaml, try changing:
github:
  mode: "cli"  # or "api"
```

#### GitHub Mode Issues

**Problem**: "gh CLI is not available or not authenticated"

**Solution**:

```bash
# Install gh CLI
brew install gh  # macOS
# For other OS: https://cli.github.com/

# Authenticate
gh auth login

# Or switch to API mode in config
github:
  token: "ghp_your_token"
  mode: "api"
```

**Problem**: "GitHub token is required for API mode"

**Solution**:

```bash
# Option 1: Add token to config
github:
  token: "ghp_xxx"
  mode: "api"

# Option 2: Switch to CLI mode
github:
  mode: "cli"
# Then: gh auth login

# Option 3: Use auto mode (tries CLI first, then API)
github:
  token: "ghp_xxx"  # as backup
  mode: "auto"
```

#### CircleCI Status Not Showing

**Problem**: CircleCI API token is invalid or missing.

**Solution**:

1. Get a personal API token from <https://app.circleci.com/settings/user/tokens>
2. Add to your config:

```yaml
circleci:
  api_token: "your-token-here"
```

#### AI Features Not Working

**Problem**: Claude API key is invalid or CLI not available.

**Solution**:

```bash
# Option 1: Use Claude API
# Get key from https://console.anthropic.com/
# Add to config:
claude:
  api_key: "sk-ant-xxx"

# Option 2: Install Claude CLI (if available in your region)
# vibe will auto-detect it

# Option 3: Disable AI features
vibe pr --no-ai
```

### Debug Mode

Enable debug logging for detailed HTTP requests and responses:

```bash
export VIBE_DEBUG=true
vibe <command>
```

This will show:

- Full HTTP requests (URLs, headers, body)
- Response status codes and bodies
- Timing information
- Error stack traces

### Getting Help

1. **Check the documentation**: Read this README thoroughly
2. **Search existing issues**: <https://github.com/rithyhuot/vibe/issues>
3. **Enable debug mode**: Set `VIBE_DEBUG=true` to see detailed logs
4. **Create an issue**: Provide:
   - Your OS and version
   - Go version (`go version`)
   - vibe version (`vibe --version`)
   - Full error message
   - Steps to reproduce
   - Debug logs (if applicable)

## Credits

Built with:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [go-git](https://github.com/go-git/go-git) - Git operations
- [go-github](https://github.com/google/go-github) - GitHub API
- [Survey](https://github.com/AlecAivazis/survey) - Interactive prompts
- [Color](https://github.com/fatih/color) - Terminal colors
