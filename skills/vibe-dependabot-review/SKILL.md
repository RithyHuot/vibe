---
name: vibe-dependabot-review
description: Review Dependabot PRs for breaking changes and create a draft PR with fixes. Use when user says "review dependabot", "check dependabot PR", or wants to handle dependency updates safely.
argument-hint: [pr-number]
allowed-tools: Bash(vibe:*), Bash(gh:*), Bash(git:*), Read, Grep, Glob, Edit, Write, AskUserQuestion
---

# Dependabot PR Review & Fix

## Purpose

Automatically review Dependabot pull requests to identify potential breaking changes and create a draft PR with necessary fixes. This skill helps safely update dependencies by:

- **Detecting Breaking Changes**: Identify major version bumps, API changes, deprecations
- **Analyzing Impact**: Understand how changes affect the codebase
- **Creating Fixes**: Make necessary code changes to maintain compatibility
- **Running Tests**: Verify fixes don't introduce new issues
- **Draft PR**: Create a reviewable PR with all fixes

## When to Use

- When Dependabot creates a PR for dependency updates
- Before merging dependency updates
- When dependency updates are causing CI failures
- To proactively identify compatibility issues

## Review Process

### 1. Identify Dependabot PR

First, determine which PR to review:

**If PR number provided in `$ARGUMENTS`:**

- Use it directly

**If no PR number provided:**

1. Check if current branch is from Dependabot
2. If not on a Dependabot branch, list recent Dependabot PRs:

   ```bash
   gh pr list --author "app/dependabot" --limit 10 --json number,title,createdAt,headRefName
   ```

3. Use AskUserQuestion to ask: "Which Dependabot PR would you like to review?" (Provide list of PR numbers with titles)

**Dependabot PR Patterns:**

- Author: `app/dependabot` or `dependabot[bot]`
- Branch name: `dependabot/<ecosystem>/<package>-<version>`
- Title format: `Bump <package> from <old-version> to <new-version>`

### 2. Analyze Dependency Changes

Extract and analyze the dependency changes:

```bash
# Get the PR details and diff
gh pr view $PR_NUMBER --json body,commits

# Get the actual code diff
gh pr diff $PR_NUMBER

# For Go projects, check go.mod changes
gh pr diff $PR_NUMBER --name-only | grep -E "(go.mod|go.sum)"
gh pr diff $PR_NUMBER -- go.mod
```

**What to look for in the diff:**

- [ ] Direct dependencies vs indirect dependencies
- [ ] Major version changes (v1.x.x → v2.x.x)
- [ ] Minor version changes (v1.2.x → v1.3.x)
- [ ] Patch version changes (v1.2.3 → v1.2.4)
- [ ] Multiple dependencies updated together
- [ ] New dependencies added
- [ ] Dependencies removed

### 3. Assess Breaking Change Risk

Determine the risk level based on semantic versioning and change type:

#### 🔴 HIGH RISK (Major Version Bump)

Major version changes (e.g., v1.x → v2.x) likely contain breaking changes:

**For each major version change:**

1. **Find the changelog/release notes:**

   ```bash
   # Get repository URL from go.mod or package manager
   # Construct GitHub releases URL
   gh api repos/<owner>/<repo>/releases --jq '.[0] | {tag_name, name, body}'
   ```

2. **Look for these indicators:**
   - "BREAKING CHANGE" in release notes
   - "Breaking Changes" section
   - Deprecated APIs removed
   - Function signature changes
   - Package/module renames
   - Configuration format changes
   - Minimum version requirements changed

3. **Search codebase for usage:**

   ```bash
   # Find all imports/uses of the updated package
   grep -r "github.com/package/name" --include="*.go"

   # For Go, find all files importing the package
   grep -r "\"github.com/package/name" --include="*.go"
   ```

#### 🟡 MEDIUM RISK (Minor Version or Multiple Patches)

Minor version changes or multiple patch updates may have subtle issues:

- New features might conflict with existing code
- Behavior changes in edge cases
- Performance characteristics changed
- Deprecation warnings introduced

#### 🟢 LOW RISK (Single Patch Version)

Patch version changes are typically safe:

- Bug fixes
- Security patches
- Documentation updates

### 4. Investigate Breaking Changes

For HIGH and MEDIUM risk changes:

#### Step A: Read Release Notes/Changelog

```bash
# Fetch release notes from GitHub
gh api repos/<owner>/<repo>/releases/tags/<new-version> --jq '.body'

# Or fetch CHANGELOG.md
gh api repos/<owner>/<repo>/contents/CHANGELOG.md --jq '.content' | base64 -d

# Look for migration guides
gh api repos/<owner>/<repo>/contents/MIGRATION.md --jq '.content' | base64 -d
```

#### Step B: Identify Affected Code

Search for usage of changed APIs:

```bash
# Use Grep tool to find imports
# Use Read tool to examine usage in context
# Use Glob tool to find all relevant files

# For Go projects, common patterns:
# - Changed function signatures
# - Removed methods
# - Renamed packages
# - Changed interfaces
```

#### Step C: Check Test Files

```bash
# Find test files
find . -name "*_test.go"

# Look for tests that might be affected
grep -r "PackageName" --include="*_test.go"
```

### 5. Create Fix Branch

Create a new branch based on the Dependabot PR:

```bash
# Fetch the Dependabot PR branch
git fetch origin $DEPENDABOT_BRANCH

# Create new fix branch from Dependabot branch
# Format: fix/dependabot-<package>-<version>
BRANCH_NAME="fix/dependabot-$(echo $PACKAGE_NAME | tr '/' '-')-$NEW_VERSION"
git checkout -b $BRANCH_NAME origin/$DEPENDABOT_BRANCH
```

### 6. Implement Fixes

Based on the breaking changes identified, implement necessary fixes:

#### Common Go Fixes

**API Signature Changes:**

```go
// Old: func Process(ctx context.Context, data string) error
// New: func Process(ctx context.Context, data string, opts ...Option) error

// Fix: Add empty options
err := pkg.Process(ctx, data)  // Old
err := pkg.Process(ctx, data, pkg.DefaultOptions()...)  // New
```

**Package Renames:**

```go
// Old: import "github.com/pkg/old"
// New: import "github.com/pkg/new"

// Use Edit tool to update all imports
```

**Interface Changes:**

```go
// If interface added new methods, implement them:
type MyHandler struct {}

// Add new required method
func (h *MyHandler) NewMethod() error {
    return nil
}
```

**Configuration Changes:**

```go
// Old config format → New config format
// Update configuration structs and initialization
```

#### Fix Implementation Steps

1. **Use Read tool** to understand current implementation
2. **Use Edit tool** to make precise changes
3. **Preserve existing behavior** where possible
4. **Add comments** explaining why changes were needed
5. **Update tests** if test expectations changed

### 7. Verify Fixes

Run tests to ensure fixes work:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific packages that use updated dependencies
go test ./internal/services/...

# Check if build succeeds
go build ./...

# Run linter
go vet ./...
```

**If tests fail:**

1. Review test output for specific failures
2. Identify root cause
3. Make additional fixes
4. Re-run tests
5. Iterate until all tests pass

**If tests pass:**

1. Verify coverage didn't decrease
2. Check for new compiler warnings
3. Run integration tests if available

### 8. Create Draft PR

Create a draft PR with your fixes:

```bash
# Stage all changes
git add .

# Create detailed commit message
git commit -m "fix: adapt code for <package> <new-version> update

- Updated API calls to match new signatures
- Implemented new required interface methods
- Updated configuration for new format
- Adjusted tests for behavior changes

Addresses breaking changes in dependabot PR #$DEPENDABOT_PR_NUMBER"

# Push the fix branch
git push -u origin $BRANCH_NAME

# Create draft PR
gh pr create \
  --title "Fix breaking changes from $PACKAGE_NAME $NEW_VERSION update" \
  --body "$(cat <<EOF
## Summary

This PR contains fixes for breaking changes introduced by updating $PACKAGE_NAME from $OLD_VERSION to $NEW_VERSION in PR #$DEPENDABOT_PR_NUMBER.

## Breaking Changes Addressed

[List each breaking change and how it was fixed]

- **Change 1**: Description of breaking change
  - **Fix**: How it was addressed
  - **Files**: List affected files

- **Change 2**: Description of breaking change
  - **Fix**: How it was addressed
  - **Files**: List affected files

## Testing

- [ ] All existing tests pass
- [ ] No new compiler warnings
- [ ] Verified functionality with manual testing
- [ ] Checked for performance regressions

## Related PRs

- Dependabot PR: #$DEPENDABOT_PR_NUMBER

## Review Notes

This PR should be reviewed and merged BEFORE the Dependabot PR. Once this is merged, the Dependabot PR can be safely merged or rebased.

---

🤖 Generated with vibe-dependabot-review skill
EOF
)" \
  --draft \
  --base main
```

### 9. Provide Summary

After creating the draft PR, provide a summary to the user:

```markdown
## Dependabot Review Complete

**Dependabot PR**: #$DEPENDABOT_PR_NUMBER - $PACKAGE_NAME $OLD_VERSION → $NEW_VERSION

### Risk Assessment: [HIGH/MEDIUM/LOW]

### Breaking Changes Found:
1. [Breaking change description]
2. [Breaking change description]

### Fixes Applied:
- Updated X files
- Modified Y API calls
- Implemented Z new methods

### Test Results:
✅ All tests passing
✅ Build successful
✅ No new warnings

### Draft PR Created:
**PR #XXX**: [PR URL]

### Next Steps:
1. Review the draft PR
2. Request reviews from team
3. Merge the fix PR first
4. Then merge or rebase the Dependabot PR
```

## Risk Decision Tree

Use this decision tree to determine the review depth:

```
Is it a major version bump?
├─ YES → HIGH RISK
│   ├─ Read full changelog
│   ├─ Search for breaking changes
│   ├─ Create fix branch
│   └─ Make necessary fixes
│
└─ NO → Is it a minor version bump or multiple dependencies?
    ├─ YES → MEDIUM RISK
    │   ├─ Skim changelog for notable changes
    │   ├─ Run tests on Dependabot branch
    │   └─ Create fix branch if tests fail
    │
    └─ NO → LOW RISK (single patch)
        └─ Quick review, likely safe to merge

```

## Language-Specific Patterns

### Go Breaking Changes

Common breaking changes in Go dependencies:

- **Function signature changes**: Added/removed/reordered parameters
- **Interface changes**: New required methods
- **Package renames**: Import path changes
- **Type changes**: Struct field changes, type aliases
- **Error handling**: New error types, changed error messages
- **Context requirements**: Functions now require context.Context
- **Configuration**: Changed config struct or initialization

**Detection:**

```bash
# Find all imports of the package
grep -r "\"$PACKAGE_PATH\"" --include="*.go" .

# Check for struct initialization
grep -r "$STRUCT_NAME{" --include="*.go" .

# Check for function calls
grep -r "$PACKAGE_NAME\.$FUNCTION_NAME" --include="*.go" .
```

### GitHub Actions Breaking Changes

For `github-actions` ecosystem updates:

- **Input/output changes**: Action inputs renamed or removed
- **Workflow syntax**: New required fields
- **Runner changes**: Different runtime requirements
- **Permissions**: New permission requirements

**Detection:**

```bash
# Find all uses of the action
grep -r "uses: $ACTION_NAME@" .github/workflows/

# Read workflow files
find .github/workflows -name "*.yml" -o -name "*.yaml"
```

## Example Scenarios

### Scenario 1: Major Version Bump (cobra v1.8.0 → v2.0.0)

```bash
# 1. Identify change
gh pr view 123 --json title
# Output: "Bump github.com/spf13/cobra from 1.8.0 to 2.0.0"

# 2. Fetch release notes
gh api repos/spf13/cobra/releases/tags/v2.0.0 --jq '.body'
# Shows: "Breaking: Command.Run signature changed to include error return"

# 3. Find affected code
grep -r "cobra.Command" --include="*.go" .
# Find: internal/commands/*.go use cobra.Command

# 4. Create fix branch
git checkout -b fix/dependabot-cobra-v2 origin/dependabot/go_modules/github.com/spf13/cobra-2.0.0

# 5. Fix code
# Use Edit tool to update all Run functions to RunE

# 6. Test
go test ./...

# 7. Create draft PR
gh pr create --draft --title "Fix cobra v2 breaking changes" ...
```

### Scenario 2: Multiple Patch Updates (Low Risk)

```bash
# 1. Identify changes
gh pr view 124 --json title
# Output: "Bump go dependencies - security patches"

# 2. Quick check
gh pr diff 124 -- go.mod
# Shows: Only patch version bumps

# 3. Run tests on Dependabot branch
git fetch origin dependabot/go_modules/...
git checkout origin/dependabot/go_modules/...
go test ./...

# 4. If tests pass
# → Report: "LOW RISK - Patch updates only, tests pass, safe to merge"

# 5. If tests fail
# → Investigate failures and create fix branch
```

## Important Notes

- **Always test fixes**: Run full test suite before creating PR
- **Preserve behavior**: Don't add new features while fixing breaking changes
- **Clear commits**: Use descriptive commit messages explaining what broke and how it's fixed
- **Link PRs**: Always reference the Dependabot PR in your fix PR
- **Merge order**: Fix PR should be merged BEFORE Dependabot PR
- **Ask for help**: If changes are complex, ask the user for guidance
- **Security updates**: Prioritize security-related dependency updates

## Workflow Integration

This skill works well with:

- `vibe-pr-status`: Check Dependabot PR status
- `vibe-ci-status`: Debug test failures
- `vibe-code-review`: Review the fixes you made
- `vibe-merge`: Merge PRs in correct order

## Tips for Success

1. **Read release notes thoroughly**: Don't skip this step for major versions
2. **Search broadly**: Breaking changes might affect unexpected parts of the codebase
3. **Test incrementally**: Fix and test one breaking change at a time
4. **Document decisions**: Add comments explaining non-obvious fixes
5. **Check dependencies**: A breaking change might affect multiple packages
6. **Consider rollback**: If fixes are too complex, consider staying on old version
7. **Communicate**: Keep the team informed about significant dependency changes

## Error Handling

Common issues and solutions:

| Issue | Cause | Solution |
|-------|-------|----------|
| Can't fetch Dependabot PR | Wrong PR number or auth issue | Verify PR exists: `gh pr list` |
| Tests fail after fix | Incomplete fix or new issue | Review test output, debug specific failures |
| Merge conflict | Dependabot PR outdated | Rebase Dependabot PR first: ask user |
| Can't find breaking changes | Release notes unclear | Search codebase for package usage, test empirically |
| Multiple breaking changes | Complex update | Fix one at a time, create separate commits |

## Advanced Options

### Custom Test Commands

If project uses custom test commands:

```bash
# Make test
make test

# Docker-based tests
docker-compose run test

# Integration tests
./scripts/run-integration-tests.sh
```

### Skip Low-Risk PRs

For patch updates with passing tests:

```bash
# Quick validation
go test ./... && go build ./...

# If passing, report to user
# "✅ Patch update verified, safe to merge Dependabot PR directly"
```

### Batch Multiple Dependabot PRs

If multiple low-risk Dependabot PRs exist:

```bash
# List all Dependabot PRs
gh pr list --author "app/dependabot" --json number,title

# Create single branch updating all
# Test all changes together
# Create single PR with all updates
```
