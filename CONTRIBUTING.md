# Contributing to vibe

Thank you for your interest in contributing to vibe! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors. We expect:

- Respectful communication
- Constructive feedback
- Focus on the code and ideas, not the person
- Welcoming and inclusive language

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Git
- GitHub CLI (`gh`) - optional but recommended
- ClickUp account (for testing integrations)
- GitHub account

### Development Setup

1. **Fork the repository**

   ```bash
   # Fork via GitHub UI or gh CLI
   gh repo fork rithyhuot/vibe --clone
   ```

2. **Clone your fork**

   ```bash
   git clone https://github.com/YOUR_USERNAME/vibe.git
   cd vibe
   ```

3. **Add upstream remote**

   ```bash
   git remote add upstream https://github.com/rithyhuot/vibe.git
   ```

4. **Install dependencies**

   ```bash
   go mod download
   ```

5. **Build the project**

   ```bash
   make build
   ```

6. **Run tests**

   ```bash
   make test
   ```

## Making Changes

### Branching Strategy

We use a feature branch workflow:

1. **Create a feature branch from `main`**

   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**

   - Write clean, idiomatic Go code
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

3. **Commit your changes**

   ```bash
   git add .
   git commit -m "Add feature: description of changes"
   ```

### Commit Message Guidelines

We follow conventional commit messages:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

Examples:

```
feat: add support for custom sprint patterns
fix: handle nil pointer in task status update
docs: update README with new configuration options
test: add unit tests for branch utilities
```

## Code Style

### General Guidelines

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code (automatically done by `make fmt`)
- Use `golangci-lint` for linting (run `make lint`)
- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### Go Best Practices

1. **Error Handling**

   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to get task: %w", err)
   }

   // Bad
   if err != nil {
       return err
   }
   ```

2. **Context Usage**

   ```go
   // Always pass context as the first parameter
   func GetTask(ctx context.Context, taskID string) (*Task, error) {
       // ...
   }
   ```

3. **Interfaces**

   ```go
   // Define interfaces where they're used, not where they're implemented
   type TaskClient interface {
       GetTask(ctx context.Context, id string) (*Task, error)
   }
   ```

### Project Structure

```
vibe/
├── cmd/vibe/           # Main entry point
├── internal/             # Internal packages
│   ├── commands/         # CLI command implementations
│   ├── services/         # External service integrations
│   ├── config/           # Configuration management
│   ├── models/           # Data models
│   ├── skills/           # Skills installer utility
│   ├── ui/               # UI utilities
│   └── utils/            # Helper utilities
├── skills/               # Claude Code skill definitions (embedded in binary)
├── embedded.go           # Embeds skills directory into binary
├── Makefile              # Build automation
└── README.md             # User documentation
```

For detailed architecture documentation including design patterns, data flow, and component interactions, see [ARCHITECTURE.md](ARCHITECTURE.md).

### Adding or Modifying Skills

The `skills/` directory contains Claude Code skill definitions that are embedded into the binary at build time:

1. **Adding a new skill**: Create a new directory under `skills/` with a `SKILL.md` file
2. **Modifying skills**: Edit the `SKILL.md` file in the appropriate skill directory
3. **Testing**: Skills are automatically embedded during build - no special steps needed
4. **Installation**: Users can install skills globally with `vibe skills` or `vibe init --install-skills`

The skills are embedded using Go's `embed` package in `embedded.go` and installed to `~/.claude/skills/` when requested.

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/utils/...

# Run a specific test
go test -run TestGenerateBranchName ./internal/utils/...
```

### Writing Tests

1. **Use table-driven tests**

   ```go
   func TestFunction(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           expected string
       }{
           {
               name:     "test case 1",
               input:    "input",
               expected: "output",
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               result := Function(tt.input)
               assert.Equal(t, tt.expected, result)
           })
       }
   }
   ```

2. **Test file naming**

   - Test files should be named `*_test.go`
   - Place tests in the same package as the code being tested

3. **Use testify for assertions**

   ```go
   import "github.com/stretchr/testify/assert"

   assert.Equal(t, expected, actual)
   assert.NoError(t, err)
   assert.True(t, condition)
   ```

### Test Coverage

- Aim for at least 70% code coverage for new code
- Focus on testing critical business logic
- Test error cases and edge conditions
- Mock external dependencies (APIs, file system, etc.)

### Local Testing

Before submitting your changes, test them locally:

#### 1. Build and Install Locally

```bash
# Build the binary
make build

# Install to local Go bin
make install

# Verify installation
vibe --version
```

#### 2. Test Against Real APIs (Optional)

Create a test ClickUp workspace and GitHub repo for testing:

```bash
# Set up test configuration
cp ~/.config/vibe/config.yaml ~/.config/vibe/config-backup.yaml

# Edit config with test workspace/repo
vim ~/.config/vibe/config.yaml

# Test key workflows
vibe workon test-ticket-id
vibe pr create --draft
vibe ci-status
```

#### 3. Test Command Help

Ensure all commands have proper help text:

```bash
# Test root help
vibe --help

# Test command help
vibe workon --help
vibe pr --help
vibe issue --help
```

#### 4. Test Error Handling

Verify error messages are user-friendly:

```bash
# Invalid ticket ID
vibe workon invalid

# Missing configuration
mv ~/.config/vibe/config.yaml ~/.config/vibe/config.yaml.bak
vibe workon abc123
mv ~/.config/vibe/config.yaml.bak ~/.config/vibe/config.yaml

# No git repository
cd /tmp
vibe ticket
cd -
```

#### 5. Test with Different Configurations

Test GitHub modes:

```bash
# API mode
echo "github:\n  mode: api" >> ~/.config/vibe/config.yaml
vibe pr-status

# CLI mode (requires gh CLI)
echo "github:\n  mode: cli" >> ~/.config/vibe/config.yaml
vibe pr-status

# Auto mode (default)
echo "github:\n  mode: auto" >> ~/.config/vibe/config.yaml
vibe pr-status
```

#### 6. Integration Testing Checklist

- [ ] `vibe init` creates config file
- [ ] `vibe workon <ticket-id>` creates branch and updates status
- [ ] `vibe ticket` displays ticket details
- [ ] `vibe comment "test"` adds comment to ticket
- [ ] `vibe pr create` creates PR successfully
- [ ] `vibe pr-status` shows PR status
- [ ] `vibe issues` lists issues
- [ ] `vibe ci-status` shows CI status
- [ ] `vibe skills` installs Claude Code skills

## Code Review Checklist

When reviewing PRs or self-reviewing before submission:

### Functionality

- [ ] Code works as intended
- [ ] All acceptance criteria met
- [ ] Edge cases handled
- [ ] Error cases handled gracefully
- [ ] User-friendly error messages

### Code Quality

- [ ] Follows existing patterns
- [ ] No code duplication
- [ ] Functions are focused and single-purpose
- [ ] Variable/function names are clear
- [ ] Comments explain "why", not "what"
- [ ] No magic numbers (use constants)

### Testing

- [ ] Unit tests added for new code
- [ ] Tests cover happy path and error cases
- [ ] Tests are deterministic (no flaky tests)
- [ ] Test coverage >= 70% for new code
- [ ] Integration tests pass locally

### Security

- [ ] Input validation present
- [ ] No SQL/command injection vulnerabilities
- [ ] Secrets not hardcoded or logged
- [ ] API tokens handled securely
- [ ] Branch names sanitized

### Documentation

- [ ] README updated if needed
- [ ] ARCHITECTURE.md updated for architectural changes
- [ ] Command help text added/updated
- [ ] Examples provided for new features
- [ ] Comments added for complex logic

### Performance

- [ ] No unnecessary API calls
- [ ] Caching used where appropriate
- [ ] No blocking operations in loops
- [ ] HTTP requests have timeouts
- [ ] Efficient data structures used

### User Experience

- [ ] Commands have intuitive names
- [ ] Flags are consistent with existing commands
- [ ] Helpful error messages with suggestions
- [ ] Progress indicators for long operations
- [ ] Output is well-formatted and readable

### Git & Commits

- [ ] Commit messages are descriptive
- [ ] Commits are atomic (one logical change per commit)
- [ ] No unrelated changes included
- [ ] Branch name follows convention
- [ ] No merge commits (rebase preferred)

## Submitting Changes

### Pull Request Process

1. **Update your branch**

   ```bash
   git checkout main
   git pull upstream main
   git checkout your-feature-branch
   git rebase main
   ```

2. **Push to your fork**

   ```bash
   git push origin your-feature-branch
   ```

3. **Create a pull request**

   - Go to GitHub and create a PR from your fork to `rithyhuot/vibe:main`
   - Fill out the PR template with:
     - Description of changes
     - Related issues (if any)
     - Testing performed
     - Screenshots (if UI changes)

4. **PR Requirements**

   - All tests must pass
   - Code must pass linting (`make lint`)
   - Documentation must be updated
   - PR description must be clear and complete
   - Commit history should be clean (squash if needed)

5. **Code Review**

   - Respond to feedback promptly
   - Make requested changes
   - Push additional commits to your branch
   - Request re-review when ready

6. **Merging**

   - PRs will be merged by maintainers after approval
   - PRs are typically squash-merged to keep history clean

## Reporting Bugs

### Before Submitting

1. Check if the bug has already been reported
2. Verify you're using the latest version
3. Collect relevant information:
   - OS and version
   - Go version
   - vibe version
   - Steps to reproduce
   - Expected vs actual behavior
   - Error messages or logs

### Bug Report Template

```markdown
**Description**
Brief description of the issue

**Steps to Reproduce**
1. Run command: `vibe abc123xyz`
2. See error

**Expected Behavior**
What should happen

**Actual Behavior**
What actually happens

**Environment**
- OS: macOS 14.0
- Go: 1.24.0
- vibe: 0.1.0

**Additional Context**
Any other relevant information
```

## Requesting Features

### Before Requesting

1. Check if the feature has already been requested
2. Consider if the feature fits the project's scope
3. Think about how it would benefit other users

### Feature Request Template

```markdown
**Feature Description**
Clear description of the proposed feature

**Use Case**
Why is this feature needed? What problem does it solve?

**Proposed Solution**
How should this feature work?

**Alternatives Considered**
What other approaches did you consider?

**Additional Context**
Any other relevant information, mockups, examples, etc.
```

## Development Tips

### Useful Make Commands

```bash
make build      # Build the binary
make test       # Run tests
make lint       # Run linter
make fmt        # Format code
make install    # Install to $GOPATH/bin
make clean      # Clean build artifacts
```

### Debugging

Enable debug mode for verbose output:

```bash
export VIBE_DEBUG=true
vibe <command>
```

### Testing with Real APIs

Create a test configuration file for development:

```yaml
# ~/.config/vibe/config-dev.yaml
clickup:
  api_token: "test_token"
  # ... other test credentials
```

Use it with:

```bash
vibe --config ~/.config/vibe/config-dev.yaml <command>
```

## Questions?

If you have questions that aren't covered here:

1. Check the [README](README.md) for user documentation
2. Search existing [GitHub Issues](https://github.com/rithyhuot/vibe/issues)
3. Open a new issue with your question

## License

By contributing to vibe, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to vibe! 🚀
