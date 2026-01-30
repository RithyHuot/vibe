---
name: vibe-code-review
description: Perform comprehensive code review of changes. Use when user asks to "review code", "review my changes", "check my code", or before creating PRs. Reviews for bugs, security, performance, best practices, and readability.
allowed-tools: Bash(git:*), Bash(vibe:*), Read, Grep
---

# Code Review

## Purpose

Perform a comprehensive code review of changes to identify:

- **Bugs and Logic Errors**: Potential runtime errors, edge cases, null pointer issues
- **Security Vulnerabilities**: SQL injection, XSS, command injection, exposed secrets
- **Performance Issues**: N+1 queries, inefficient algorithms, memory leaks
- **Best Practices**: Code style, naming conventions, SOLID principles
- **Readability**: Complex logic, missing documentation, unclear variable names
- **Test Coverage**: Missing tests, untested edge cases
- **Type Safety**: Missing error handling, incorrect types

## When to Use

- Before creating a pull request
- After making significant changes
- When user explicitly asks for code review
- As part of development workflow
- Before committing changes

## Review Process

### 1. Gather Context

First, understand what changed:

```bash
# Get list of changed files
git diff --name-only

# Get the actual diff (staged and unstaged)
git diff HEAD

# Or compare with main branch
git diff main...HEAD

# Get commit history for context
git log --oneline -10
```

### 2. Read Relevant Files

For context, read the full files that were modified (not just the diff):

```bash
# Use Read tool to view entire files
# This helps understand the broader context
```

### 3. Perform Review

Analyze each file systematically:

#### Security Review

Check for:

- [ ] Hard-coded credentials or API keys
- [ ] SQL injection vulnerabilities (concatenated queries)
- [ ] XSS vulnerabilities (unescaped user input)
- [ ] Command injection (shell command construction)
- [ ] Path traversal vulnerabilities
- [ ] Insecure cryptography or random number generation
- [ ] Exposed sensitive data in logs
- [ ] Missing input validation
- [ ] Insecure dependencies

#### Bug Detection

Check for:

- [ ] Null/undefined pointer dereferences
- [ ] Array index out of bounds
- [ ] Race conditions
- [ ] Resource leaks (unclosed files, connections)
- [ ] Infinite loops
- [ ] Off-by-one errors
- [ ] Type mismatches
- [ ] Missing error handling
- [ ] Incorrect error handling (swallowed exceptions)
- [ ] Logic errors in conditionals

#### Performance Review

Check for:

- [ ] N+1 query problems
- [ ] Inefficient loops (nested O(n²) operations)
- [ ] Unnecessary database queries
- [ ] Missing indexes on database queries
- [ ] Large memory allocations
- [ ] Blocking operations in async code
- [ ] Missing caching opportunities
- [ ] Inefficient string concatenation
- [ ] Redundant computations

#### Code Quality

Check for:

- [ ] Code duplication (DRY principle)
- [ ] Functions that are too long (>50 lines)
- [ ] High cyclomatic complexity
- [ ] Poor naming (unclear variables/functions)
- [ ] Magic numbers (use constants)
- [ ] Dead code (unused functions/variables)
- [ ] Commented-out code
- [ ] Missing documentation for public APIs
- [ ] Inconsistent code style
- [ ] Deep nesting (>3 levels)

#### Best Practices

Check for:

- [ ] SOLID principles violations
- [ ] Separation of concerns
- [ ] Single responsibility principle
- [ ] Dependency injection opportunities
- [ ] Error messages are user-friendly
- [ ] Logging is appropriate (not too verbose/quiet)
- [ ] Configuration is externalized
- [ ] Backward compatibility considerations
- [ ] API design consistency

#### Testing

Check for:

- [ ] Missing unit tests for new code
- [ ] Missing edge case tests
- [ ] Missing error case tests
- [ ] Test coverage decreased
- [ ] Tests are flaky or fragile
- [ ] Integration tests needed
- [ ] Missing test documentation

### 4. Provide Structured Feedback

Organize findings by:

1. **Critical Issues** 🔴 - Must fix (security, bugs)
2. **Important** 🟡 - Should fix (performance, quality)
3. **Suggestions** 🔵 - Nice to have (style, optimization)

For each finding:

- **Location**: File path and line number(s)
- **Issue**: Clear description of the problem
- **Impact**: Why it matters
- **Suggestion**: How to fix it
- **Example**: Show corrected code if helpful

### 5. Example Feedback Format

```markdown
## Code Review Findings

### 🔴 Critical Issues

**File: src/auth/login.go:45-52**
- **Issue**: SQL injection vulnerability
- **Details**: User input is directly concatenated into SQL query
- **Impact**: Attackers could access or modify database
- **Fix**: Use parameterized queries:
  ```go
  query := "SELECT * FROM users WHERE email = $1"
  db.Query(query, email)
  ```

### 🟡 Important

**File: src/api/handler.go:120-135**

- **Issue**: N+1 query problem
- **Details**: Loop makes individual database queries
- **Impact**: Performance degrades with scale (O(n) queries)
- **Fix**: Use a single query with JOIN or IN clause

### 🔵 Suggestions

**File: src/utils/helper.go:78**

- **Issue**: Function is too long (85 lines)
- **Details**: Hard to understand and test
- **Suggestion**: Break into smaller, focused functions

```

## Review Levels

### Quick Review (5 min)
- Focus on critical security and bug issues only
- Scan for obvious problems
- Check for exposed secrets

### Standard Review (15 min)
- Full security and bug review
- Performance check
- Major code quality issues
- Best practices

### Thorough Review (30+ min)
- All categories
- Read full context of files
- Check test coverage
- Suggest refactoring opportunities
- Review documentation

## Language-Specific Checks

### Go
- Error handling (every error checked)
- Goroutine leaks
- Mutex usage (deadlocks)
- Context cancellation
- Defer usage

### JavaScript/TypeScript
- Promise rejection handling
- Memory leaks (event listeners)
- XSS vulnerabilities
- Type safety (TypeScript)
- Async/await usage

### Python
- Exception handling
- Resource cleanup (with statements)
- Type hints
- List comprehension vs loops
- Virtual environment

### Java
- Exception handling
- Resource cleanup (try-with-resources)
- Thread safety
- Stream API usage
- Memory management

## Tips

- **Be constructive**: Frame feedback positively
- **Explain why**: Don't just point out issues, explain the impact
- **Provide examples**: Show how to fix issues when possible
- **Prioritize**: Focus on critical issues first
- **Consider context**: Understand the purpose before suggesting changes
- **Ask questions**: If something is unclear, ask for clarification
- **Praise good code**: Acknowledge well-written code too

## Important Notes

- Do NOT modify files during review (use Read, not Edit)
- Focus on changes, but consider surrounding context
- Security issues are highest priority
- Performance issues depend on context (critical path vs rare code)
- Style suggestions should align with existing codebase patterns
- Always consider backwards compatibility
- Test code differently than production code

## Workflow Integration

This skill works well with:
- `vibe-pr`: Review before creating PR
- `vibe-issue`: Review while working on an issue
- Use after making changes but before committing
- Integrate into pre-commit hooks

## Example Usage

```bash
# Review all changes
git diff HEAD

# Review specific commit
git show <commit-hash>

# Review changes since main
git diff main...HEAD

# Review staged changes only
git diff --staged
```

Then perform the review following the structured process above.
