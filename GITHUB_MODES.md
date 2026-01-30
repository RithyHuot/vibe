# GitHub Integration Modes

vibe supports two methods for integrating with GitHub: **API mode** and **CLI mode**. This document explains both approaches and when to use each.

## Overview

| Feature | API Mode | CLI Mode | Auto Mode |
|---------|----------|----------|-----------|
| **Authentication** | Personal Access Token | `gh auth` credentials | Either |
| **Setup Complexity** | Medium (create token) | Low (one command) | Low |
| **CI/CD Compatible** | ✅ Yes | ⚠️ Depends | ✅ Yes |
| **External Dependencies** | None | Requires `gh` CLI | None (fallback) |
| **Configuration** | Token in config | No token needed | Token optional |
| **Recommended For** | Scripts, CI/CD | Local development | Most users |

## Quick Decision Flowchart

```
Start: Which GitHub mode should I use?
│
├─ Are you setting up CI/CD or automation?
│  └─ YES → Use API Mode
│         └─ Create token, add to config/secrets
│
├─ Do you already have 'gh' CLI installed?
│  └─ YES → Use CLI Mode or Auto Mode
│         ├─ CLI Mode: Only use gh CLI
│         └─ Auto Mode: Use gh locally, API in CI
│
├─ Working on multiple repos/orgs frequently?
│  └─ YES → Use CLI Mode
│         └─ Auth once, works everywhere
│
├─ Need exact control over assignees/labels?
│  └─ YES → Use API Mode
│         └─ API replaces, CLI adds
│
└─ Unsure? → Use Auto Mode (Recommended)
            └─ Works with either gh CLI or token
```

## Performance Comparison

| Operation | API Mode | CLI Mode | Notes |
|-----------|----------|----------|-------|
| **Create PR** | ~200-500ms | ~500-800ms | CLI adds overhead for subprocess |
| **Get PR** | ~100-300ms | ~300-600ms | CLI parses JSON output |
| **List PRs** | ~200-400ms | ~400-700ms | CLI spawns new process |
| **Update PR** | ~200-400ms | ~400-700ms | Similar overhead |
| **Cached Operations** | Same | Same | Both benefit from caching |
| **Batch Operations** | Faster | Slower | API can reuse HTTP connections |

**Key Takeaways:**

- **API mode is 1.5-2x faster** due to direct HTTP calls
- **CLI mode adds ~200-400ms overhead** for process spawning
- For **interactive use**, difference is barely noticeable
- For **scripting/automation**, API mode provides better performance

## When to Use Each Mode

### Use API Mode When

- ✅ Running in CI/CD pipelines (GitHub Actions, CircleCI, etc.)
- ✅ Building automation scripts
- ✅ Performance is critical (batch operations)
- ✅ Need precise control over assignees/labels
- ✅ gh CLI is not available or cannot be installed
- ✅ Working in restricted environments (containers, sandboxes)

### Use CLI Mode When

- ✅ Local development only
- ✅ Already using gh CLI daily
- ✅ Don't want to manage tokens
- ✅ Working across multiple repositories frequently
- ✅ Need SSO or advanced GitHub auth features
- ✅ Quick setup is priority

### Use Auto Mode When

- ✅ You're unsure which to choose (recommended default)
- ✅ Team has mixed preferences
- ✅ Need flexibility (local + CI/CD)
- ✅ Want best of both worlds

## Feature Support by Mode

Some GitHub features have different support levels depending on the mode:

| Feature | API Mode | CLI Mode | Notes |
|---------|----------|----------|-------|
| **GitHub Projects v2** | ✅ Full support | ✅ Full support | Both modes now supported |
| **Issue Assignees (update)** | Replace all | Additive | See behavior differences below |
| **Issue Labels (update)** | Replace all | Additive | See behavior differences below |
| **Pull Requests** | ✅ Full support | ✅ Full support | Both modes fully supported |
| **Issue Creation** | ✅ Full support | ✅ Full support | Both modes fully supported |
| **Comments** | ✅ Full support | ✅ Full support | Both modes fully supported |

### GitHub Projects v2 Support

Both API mode and CLI mode now fully support GitHub Projects v2. API mode uses the GraphQL API internally for project operations.

**Supported Project ID Formats**:

- **Node ID**: `PVT_kwDOABC123` (fastest, direct lookup)
- **Project Number**: `5` or `#5` (requires GraphQL lookup)
- **Project Name**: `"Sprint 2024"` (requires GraphQL search)

**Examples**:

```bash
# Using node ID (fastest)
vibe issue-create --title "Bug" --project "PVT_kwDOABC123"

# Using project number
vibe issue-create --title "Bug" --project "5"
vibe issue-create --title "Bug" --project "#5"

# Using project name
vibe issue-create --title "Bug" --project "Sprint 2024"

# Adding to multiple projects
vibe issue-create --title "Bug" --project "PVT_kwDOABC123" --project "Backlog"
```

**Finding Project Node IDs**:

To get a project's node ID for fastest performance:

```bash
# List projects and their node IDs
gh api graphql -f query='
  query($owner: String!) {
    organization(login: $owner) {
      projectsV2(first: 20) {
        nodes {
          id
          title
          number
        }
      }
    }
  }
' -F owner=YOUR_ORG_NAME
```

**Error Handling**:

Projects v2 operations use a two-step process:

1. Create/update the issue via REST API
2. Add to projects via GraphQL API

If step 1 succeeds but step 2 fails (e.g., invalid project ID, missing permissions):

- The issue is still created/updated successfully
- An error message indicates which projects failed
- You can manually add to projects via GitHub UI or retry

**Required Token Permissions**:

For API mode to support Projects v2, your GitHub token needs:

- `repo` scope (for issue operations)
- `project` scope (for Projects v2 operations)

Update your token at: <https://github.com/settings/tokens>

**Rate Limiting**:

GraphQL API has separate rate limits from REST API:

- **Primary rate limit**: 5,000 points/hour
- **Each mutation**: ~1 point
- **Project lookups**: ~1 point each

Examples:

- Adding issue to 1 project (using node ID): 1 point
- Adding issue to 1 project (using name): 2 points (1 for search, 1 for mutation)
- Adding issue to 10 projects (using names): up to 20 points

**Tip**: Use project node IDs instead of names/numbers for best performance and lowest rate limit impact.

### Update Behavior Differences

When updating issues, the two modes handle assignees and labels differently:

**Assignees**:

- **API mode**: Replaces all assignees with the new list
  - Example: Issue has [alice, bob] → Update with [charlie] → Result: [charlie]
- **CLI mode**: Adds assignees to the existing list (uses `--add-assignee`)
  - Example: Issue has [alice, bob] → Update with [charlie] → Result: [alice, bob, charlie]

**Labels**:

- **API mode**: Replaces all labels with the new list
  - Example: Issue has [bug, priority] → Update with [feature] → Result: [feature]
- **CLI mode**: Adds labels to the existing list (uses `--add-label`)
  - Example: Issue has [bug, priority] → Update with [feature] → Result: [bug, priority, feature]

**Recommendation**: For exact control over assignees and labels, use API mode. For additive updates, use CLI mode.

## API Mode

Uses GitHub REST API directly with a personal access token.

### Configuration

```yaml
github:
  token: "ghp_your_token_here"
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "api"
```

### Setup

1. Create a personal access token:
   - Go to <https://github.com/settings/tokens>
   - Click "Generate new token (classic)"
   - Select scope: `repo` (full control of private repositories)
   - Copy the token

2. Add to your config:

   ```bash
   vim ~/.config/vibe/config.yaml
   # Paste token in github.token field
   ```

### When to Use

- ✅ Running in CI/CD pipelines
- ✅ Automated scripts and workflows
- ✅ When `gh` CLI is not available
- ✅ When you need fine-grained permission control
- ✅ When you want consistent behavior across environments

### Pros

- No external dependencies
- Works everywhere (local, CI/CD, containers, etc.)
- Explicit permission scopes
- Direct API access (faster for batch operations)

### Cons

- Requires creating and managing tokens
- Token needs to be stored securely
- Token rotation requires config updates
- Need to configure token in multiple places (local, CI, etc.)

## CLI Mode

Uses the GitHub CLI (`gh`) for all GitHub operations.

### Configuration

```yaml
github:
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "cli"
```

Note: No token needed in config!

### Setup

1. Install GitHub CLI:

   ```bash
   # macOS
   brew install gh

   # Windows
   winget install GitHub.cli

   # Linux
   # See: https://github.com/cli/cli/blob/trunk/docs/install_linux.md
   ```

2. Authenticate:

   ```bash
   gh auth login
   ```

3. Verify:

   ```bash
   gh auth status
   ```

### When to Use

- ✅ Local development
- ✅ When you already use `gh` CLI
- ✅ Quick setup without managing tokens
- ✅ When you want to use your existing GitHub auth
- ✅ When working across multiple repos/orgs

### Pros

- Simple authentication (one command)
- No token management
- Uses your existing GitHub credentials
- Respects gh CLI configuration
- SSO and other auth methods work automatically

### Cons

- Requires `gh` CLI to be installed
- May not work in all CI/CD environments
- Depends on external tool
- Authentication tied to local machine

## Auto Mode (Recommended)

Automatically selects the best method based on what's available.

### Configuration

```yaml
github:
  token: "ghp_your_token_here"  # Optional - used as fallback
  username: "your-username"
  owner: "org-name"
  repo: "repo-name"
  mode: "auto"  # or omit - this is the default
```

### How It Works

1. **Check for gh CLI**: If `gh auth status` succeeds → use CLI mode
2. **Check for token**: If token is configured → use API mode
3. **Error**: If neither available → show helpful error message

### When to Use

- ✅ Most users (best of both worlds)
- ✅ Developers working locally and in CI/CD
- ✅ Teams with mixed setups
- ✅ When you want flexibility

### Setup

Provide EITHER:

- `gh` CLI authentication: `gh auth login`
- OR GitHub token in config

For maximum flexibility, configure both:

- Use CLI locally (more convenient)
- Fall back to API in CI/CD (more reliable)

## Implementation Details

### How Commands Work

All GitHub operations use the same interface regardless of mode:

```go
type Client interface {
    CreatePR(ctx, req) (*PullRequest, error)
    GetPR(ctx, number) (*PullRequest, error)
    UpdatePR(ctx, number, title, body) (*PullRequest, error)
    GetPRStatus(ctx, number) (*PRStatus, error)
    ListPRs(ctx, state) ([]*PullRequest, error)
    AddComment(ctx, number, body) error
    GetPRTemplate(ctx) (string, error)
}
```

The mode selection happens once at initialization:

```go
client, err := github.NewClientWithMode(
    config.GitHub.Mode,    // "api", "cli", or "auto"
    config.GitHub.Token,   // token (optional for CLI)
    config.GitHub.Owner,
    config.GitHub.Repo,
)
```

### CLI Mode Implementation

CLI mode uses `gh` commands under the hood:

| Operation | CLI Command |
|-----------|-------------|
| Create PR | `gh pr create --title "..." --body "..." --base main --head feature` |
| Get PR | `gh pr view 123 --json number,title,body,state,...` |
| Update PR | `gh pr edit 123 --title "..." --body "..."` |
| PR Status | `gh pr view 123 --json statusCheckRollup,reviews` |
| List PRs | `gh pr list --state open --json ...` |
| Add Comment | `gh pr comment 123 --body "..."` |

### API Mode Implementation

API mode uses GitHub REST API directly:

| Operation | API Endpoint |
|-----------|--------------|
| Create PR | `POST /repos/:owner/:repo/pulls` |
| Get PR | `GET /repos/:owner/:repo/pulls/:number` |
| Update PR | `PATCH /repos/:owner/:repo/pulls/:number` |
| PR Status | Multiple endpoints for checks and reviews |
| List PRs | `GET /repos/:owner/:repo/pulls?state=:state` |
| Add Comment | `POST /repos/:owner/:repo/issues/:number/comments` |

## Migrating Between Modes

### From API to CLI

1. Authenticate with gh CLI:

   ```bash
   gh auth login
   ```

2. Update config:

   ```yaml
   github:
     # token: "ghp_xxx"  # Can remove or comment out
     mode: "cli"
   ```

3. Test:

   ```bash
   vibe pr-status  # Should work without token
   ```

### From CLI to API

1. Create personal access token at <https://github.com/settings/tokens>

2. Update config:

   ```yaml
   github:
     token: "ghp_your_new_token"
     mode: "api"
   ```

3. Test:

   ```bash
   vibe pr-status  # Should work without gh CLI
   ```

### To Auto Mode

Just set `mode: "auto"` (or remove the mode line entirely):

```yaml
github:
  token: "ghp_xxx"  # Optional
  mode: "auto"      # Or omit this line
```

## Troubleshooting

### "gh CLI is not available or not authenticated"

```bash
# Check if gh is installed
which gh

# Install if needed (macOS)
brew install gh

# Authenticate
gh auth login

# Verify
gh auth status
```

### "GitHub token is required for API mode"

Either:

1. Add token to config: `github.token: "ghp_xxx"`
2. Switch to CLI mode: `github.mode: "cli"` + `gh auth login`
3. Use auto mode: `github.mode: "auto"` (tries CLI first)

### CI/CD Issues with CLI Mode

If using CLI mode in CI/CD:

1. **Option 1**: Install and authenticate gh CLI in CI

   ```yaml
   # GitHub Actions example
   - uses: cli/setup-gh@v1
   - run: gh auth login --with-token <<< "${{ secrets.GITHUB_TOKEN }}"
   ```

2. **Option 2**: Switch to API mode for CI

   ```yaml
   github:
     token: ${{ secrets.VIBE_GITHUB_TOKEN }}
     mode: "api"
   ```

3. **Option 3**: Use auto mode with token

   ```yaml
   github:
     token: ${{ secrets.VIBE_GITHUB_TOKEN }}
     mode: "auto"  # Falls back to API in CI
   ```

### Rate Limiting

Both modes have rate limits:

- **API Mode**: 5,000 requests/hour (authenticated)
- **CLI Mode**: Same as API (uses same authentication)

Check rate limit:

```bash
# API mode
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.github.com/rate_limit

# CLI mode
gh api rate_limit
```

## Best Practices

1. **Local Development**: Use CLI mode or auto mode for simplicity
2. **CI/CD**: Use API mode with token for reliability
3. **Teams**: Use auto mode so everyone can use their preferred method
4. **Multiple Repos**: CLI mode works across repos without config changes
5. **Security**: Never commit tokens to version control
6. **Automation**: API mode for scripts that run unattended

## Environment Variables

You can override the mode via environment variable:

```bash
export VIBE_GITHUB_MODE="cli"
vibe pr-status
```

This is useful for testing or temporary overrides.

## Security Considerations

### API Mode

- Store tokens securely (use environment variables or secret managers)
- Rotate tokens regularly
- Use minimum required scopes
- Don't commit tokens to git (add to .gitignore)

### CLI Mode

- Uses system keychain (more secure)
- SSO and 2FA work automatically
- Per-device authentication
- Easy to revoke (revoke session, not token)

### Recommendations

- **Local**: CLI mode (uses secure keychain)
- **CI/CD**: API mode with secret management (GitHub Secrets, etc.)
- **Shared Servers**: API mode with restricted permissions
