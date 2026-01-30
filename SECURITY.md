# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 1.0   | :x:                |

**Recommendation:** Always use the latest version of vibecli to ensure you have the most recent security updates.

## Security Best Practices for Users

### Token Storage

**DO:**

- ✅ Store tokens in environment variables
- ✅ Use config file with restrictive permissions (600)
- ✅ Rotate tokens regularly (every 90 days recommended)
- ✅ Use tokens with minimum required permissions

**DON'T:**

- ❌ Commit tokens to git repositories
- ❌ Share tokens in chat or email
- ❌ Use tokens with broader permissions than needed
- ❌ Store tokens in publicly accessible locations

### Recommended Token Permissions

#### GitHub Token

Minimum required scopes:

- `repo` - Full control of private repositories (required for private repos)
- `read:org` - Read org and team membership (for org repos)
- `workflow` - Update GitHub Actions workflows (for CI integration)

**Public repos only:**

- `public_repo` - Access public repositories (instead of full `repo`)

#### ClickUp Token

Minimum required permissions:

- Read/Write access to tasks
- Read/Write access to comments
- Read access to workspaces and spaces

#### CircleCI Token

Minimum required permissions:

- Read-only access (for viewing status)
- Build trigger (if you want to trigger builds)

### Environment Variables

Store sensitive tokens as environment variables:

```bash
# Add to ~/.zshrc or ~/.bashrc
export CLICKUP_API_TOKEN="pk_your_token_here"
export GITHUB_TOKEN="ghp_your_token_here"
export CIRCLECI_TOKEN="your_circleci_token"
export ANTHROPIC_API_KEY="your_claude_api_key"
```

**Advantages:**

- Not committed to version control
- Can be different per machine
- Easier to rotate
- Shell history can be cleared

### Configuration File Security

If using `~/.config/vibe/config.yaml`:

```bash
# Set restrictive permissions
chmod 600 ~/.config/vibe/config.yaml

# Verify permissions
ls -la ~/.config/vibe/config.yaml
# Should show: -rw------- (owner read/write only)
```

**Never commit this file to git!**

Add to your `.gitignore`:

```
.vibe.yaml
.vibe.yml
config.yaml
```

### Token Rotation

Rotate tokens regularly to limit exposure:

**GitHub:**

1. Go to <https://github.com/settings/tokens>
2. Delete old token
3. Generate new token with same scopes
4. Update environment variable or config file

**ClickUp:**

1. Go to <https://app.clickup.com/settings/apps>
2. Regenerate API token
3. Update environment variable or config file

**CircleCI:**

1. Go to <https://app.circleci.com/settings/user/tokens>
2. Revoke old token
3. Create new token
4. Update environment variable or config file

### Least Privilege Principle

Only grant the minimum permissions needed:

- Use read-only tokens where possible
- Don't use admin-level tokens for routine operations
- Create separate tokens for different tools
- Review and audit token permissions quarterly

## Security Features in vibecli

### Branch Name Validation

vibecli sanitizes branch names to prevent command injection:

```go
// Prevents malicious characters
ValidateBranchName("user/ticket/feature")  // ✅ Valid
ValidateBranchName("user;rm -rf /")        // ❌ Rejected
```

**Protected against:**

- Command injection via branch names
- Path traversal attacks
- Shell metacharacter exploits

### Input Sanitization

All user inputs are validated:

- **Ticket IDs:** Must match ClickUp format (9 alphanumeric characters)
- **PR numbers:** Must be valid integers
- **File paths:** Checked for path traversal attempts
- **URLs:** Validated against expected patterns

### Secure API Communication

All external API calls use HTTPS:

- ✅ TLS 1.2+ required
- ✅ Certificate verification enabled
- ✅ No insecure fallback to HTTP
- ✅ Timeouts prevent hanging connections

### Token Handling

Tokens are handled securely:

- Never logged (even in debug mode)
- Not included in error messages
- Excluded from crash reports
- Passed only via secure headers

### Git Operations

Git operations are sandboxed:

- No shell command execution for git operations (uses go-git library)
- Branch names validated before creation
- Remote URLs verified before push operations
- Destructive operations require confirmation

## Known Security Considerations

### Local Cache

vibecli uses **in-memory caching only**:

- **Location:** Memory only (no files on disk)
- **Contains:** Sprint folder data with 1-hour TTL
- **No sensitive data:** Tokens are NOT cached
- **Lifetime:** Cache cleared when application exits

**To clear cache:**

```bash
# Cache is in-memory only
# Simply restart the application to clear cache
```

### Network Requests

vibecli makes network requests to:

- ClickUp API (`api.clickup.com`)
- GitHub API (`api.github.com`)
- CircleCI API (`circleci.com`)
- Claude API (`api.anthropic.com`) - optional

**Network security:**

- All requests use HTTPS
- DNS resolution via system resolver
- No data sent to third parties
- User-Agent identifies vibecli version

### Local File Access

vibecli reads/writes:

- Configuration: `~/.config/vibe/config.yaml`
- Git repository: Current working directory
- Claude skills: `~/.claude/skills/`

**Permissions needed:**

- Read/write to config directory
- Read/write to git repository
- Read/write to Claude skills directory (optional)

**Note:** No persistent cache files are created. All caching is in-memory only.

## Dependency Security

We take dependency security seriously:

- Regular dependency updates
- Automated vulnerability scanning (Dependabot)
- Direct dependencies minimized
- Security patches applied promptly

**Review dependencies:**

```bash
go list -m all  # List all dependencies
go mod graph    # View dependency tree
```

## Audit Trail

vibecli operations that modify state:

| Operation | Logged Where | Auditable |
|-----------|--------------|-----------|
| Create branch | Git history | ✅ Yes |
| Create PR | GitHub audit log | ✅ Yes |
| Update ticket | ClickUp activity | ✅ Yes |
| Add comment | ClickUp activity | ✅ Yes |
| Merge PR | GitHub audit log | ✅ Yes |
| Trigger CI | CircleCI audit log | ✅ Yes |

**All operations are traceable** through external service audit logs.

## Incident Response

If your tokens are compromised:

### Immediate Actions

1. **Revoke tokens immediately:**
   - GitHub: <https://github.com/settings/tokens>
   - ClickUp: <https://app.clickup.com/settings/apps>
   - CircleCI: <https://app.circleci.com/settings/user/tokens>

2. **Check audit logs:**
   - GitHub: <https://github.com/settings/security-log>
   - ClickUp: Check workspace activity
   - CircleCI: Review recent builds

3. **Generate new tokens** with minimum required permissions

4. **Update configuration:**

   ```bash
   # Update environment variables
   export GITHUB_TOKEN="new_token_here"
   export CLICKUP_API_TOKEN="new_token_here"

   # Or update config file
   vi ~/.config/vibe/config.yaml
   ```

5. **Review recent activity** for unauthorized actions

### Prevention

- Enable 2FA on all services
- Use password manager for token storage
- Set up token expiration (where supported)
- Monitor audit logs regularly
- Review token permissions quarterly

## Security Checklist

Use this checklist to maintain security:

- [ ] Tokens stored in environment variables or secure config
- [ ] Config file has restrictive permissions (600)
- [ ] `.vibe.yaml` added to `.gitignore`
- [ ] Tokens have minimum required permissions
- [ ] 2FA enabled on GitHub, ClickUp, CircleCI
- [ ] Using latest version of vibecli
- [ ] Dependencies up to date
- [ ] Audit logs reviewed monthly
- [ ] Tokens rotated every 90 days

## Security Updates

Subscribe to security advisories:

- Watch this repository for releases
- Enable notifications for security advisories
- Follow release notes for security patches

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [GitHub Security Best Practices](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure)
- [ClickUp Security](https://clickup.com/security)
- [CircleCI Security](https://circleci.com/security/)

## Questions?

For security questions:

- Review: [FAQ.md](FAQ.md)
- Documentation: [README.md](README.md)

---

**Last Updated:** 2026-01-30

**Note:** This security policy is subject to updates. Check back regularly for the latest guidance.
