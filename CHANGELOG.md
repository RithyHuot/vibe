# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `vibe branch` command can now be used without a ticket ID for simple branch creation

### Fixed

- Release configuration and Dockerfile improvements for better builds

### Changed

- Updated Claude skills with improved verbiage and descriptions
- Enhanced add-command-skill with additional configuration prompts

## [0.1.0] - 2026-01-31

### Added

#### Core Features

- **Ticket Management**
  - Fetch task details from ClickUp with full metadata
  - Automatic task status updates when starting work
  - Add comments to ClickUp tasks from the terminal
  - Interactive ticket selection and browsing
  - Smart sprint folder detection with date parsing

- **Git & Branch Management**
  - Auto branch creation with standardized naming: `username/ticketid/description`
  - Branch validation with security checks for safe branch names
  - Seamless branch switching between feature branches
  - Smart stashing that automatically prompts for uncommitted changes
  - Working tree status tracking and visualization
  - Simple branch creation without ClickUp integration via `vibe branch`

- **Pull Request Management**
  - Interactive and non-interactive PR creation
  - Template support with auto-population from `.github/PULL_REQUEST_TEMPLATE.md`
  - AI-powered PR descriptions using Claude (API and CLI support)
  - PR status monitoring with reviews, CI checks, and merge readiness tracking
  - Section-aware PR updates for titles and descriptions
  - Draft PR support
  - Merge automation via `/merge` comments
  - Redundant git push detection to avoid double pushing

- **GitHub Issue Management**
  - Browse and filter issues by state (open, closed, all)
  - Interactive issue selection and viewing
  - Create branches directly from issue view
  - Issue creation with full metadata support (labels, assignees, milestone, projects)
  - Template support with auto-population from `.github/ISSUE_TEMPLATE.md`
  - Issue updates for title, description, state, and metadata
  - Auto-detect issue numbers from branch names
  - View issue comments with `--comments` flag

- **CI/CD Integration**
  - Real-time CircleCI pipeline and workflow status monitoring
  - Detailed failure analysis with error logs and test failure reports
  - Color-coded visual status indicators
  - Failed test results with error messages
  - Job-specific failure log viewing

- **AI Integration**
  - Dual support for Claude API and Claude CLI
  - AI-generated PR descriptions from code changes
  - Comprehensive code review for bugs, security, performance, and best practices
  - Context-aware analysis of git diffs
  - Interactive opt-in for AI features
  - Dependabot PR review with automated fix generation

#### Developer Experience

- **Rich Terminal UI**
  - Color-coded output for better readability
  - Progress spinners for long-running operations
  - Formatted tables for structured data
  - Interactive prompts with validation

- **Performance & Reliability**
  - In-memory caching with TTL for API responses
  - Fast operations with minimal API calls
  - Input sanitization and security validation
  - Debug mode with detailed HTTP request logging

- **Configuration & Setup**
  - Zero-config setup with sensible defaults
  - Global configuration file at `~/.config/vibe/config.yaml`
  - Local project overrides via `.vibe.yaml`
  - Environment variable support for sensitive credentials
  - Flexible GitHub integration modes (API, CLI, Auto)
  - One-command initialization with `vibe init`
  - Shell autocomplete support (bash, zsh, fish, powershell)

- **Claude Code Integration**
  - Global skill installation for all projects
  - 17+ specialized skills for workflow automation
  - Skills for ticket management, PR creation, issue handling, code review
  - Skills to add new commands and skills dynamically
  - Easy skill updates and management

#### Commands

- `vibe init` - Initialize configuration with optional global skill installation
- `vibe <ticket-id>` - Start working on a ClickUp ticket
- `vibe start` - Interactive ticket selection
- `vibe ticket [ticket-id]` - View ticket details
- `vibe comment <text>` - Add comment to current ticket
- `vibe branch [ticket-id]` - Create and checkout branch (with or without ticket)
- `vibe pr` - Create pull request (interactive and non-interactive modes)
- `vibe pr-status [pr-number]` - Check PR status with reviews and CI
- `vibe pr-update [pr-number]` - Update PR title or description
- `vibe merge [pr-number]` - Trigger merge via comment
- `vibe issues` - List GitHub issues with filtering and interactive selection
- `vibe issue [issue-number]` - View issue details with optional comments
- `vibe issue-create` - Create new GitHub issue with metadata
- `vibe issue-update <issue-number>` - Update existing issue
- `vibe ci-status [branch]` - Check CircleCI status
- `vibe ci-failure [job-number]` - View detailed CI failure logs
- `vibe skills` - Install, update, or uninstall Claude Code skills
- `vibe completion` - Generate shell completion scripts

#### Claude Code Skills

- `vibe` - Start work on a ClickUp ticket
- `vibe-branch` - Create and checkout a new branch
- `vibe-ticket` - Get context on current ticket
- `vibe-comment` - Add comment to ticket
- `vibe-pr` - Create a pull request
- `vibe-pr-status` - Check PR status
- `vibe-pr-update` - Update PR description
- `vibe-merge` - Merge a pull request
- `vibe-ci-status` - Check CircleCI status
- `vibe-issues` - List GitHub issues
- `vibe-issue` - View issue details
- `vibe-issue-create` - Create a new issue
- `vibe-issue-update` - Update existing issue
- `vibe-code-review` - Perform comprehensive code review
- `vibe-dependabot-review` - Review Dependabot PRs and create fixes
- `add-claude-skill` - Add new Claude Code skill for vibe CLI
- `add-command-skill` - Add new vibe command with associated skill

### Fixed

- Addressed linter issues for code quality
- Fixed redundant git push operations in PR creation workflow
- Fixed release configuration and Dockerfile for proper builds

### Changed

- Enhanced add-command-skill with additional prompts for better workflow
- Updated README with comprehensive documentation

### Dependencies

- Bumped `actions/setup-go` from 5 to 6
- Bumped `actions/checkout` from 4 to 6
- Bumped `codecov/codecov-action` from 4 to 5
- Bumped `golangci/golangci-lint-action` from 6 to 9

### Documentation

- Comprehensive README with feature overview and usage examples
- Detailed configuration guide with all options
- Workflow examples for common development tasks
- Troubleshooting guide with common issues and solutions
- Architecture documentation (ARCHITECTURE.md)
- Security policy (SECURITY.md)
- Contributing guidelines (CONTRIBUTING.md)
- FAQ with frequently asked questions
- GitHub modes guide (GITHUB_MODES.md)
- Skills documentation (SKILLS.md)
- Skill parameter guide (SKILL_PARAMETERS.md)
- Release process documentation (RELEASE.md)

## [0.0.0] - 2026-01-30

### Added

- Initial project structure and scaffolding
- Basic CLI framework with Cobra
- Configuration management with Viper
- Git operations with go-git
- ClickUp API client
- GitHub API client
- CircleCI API client
- Claude API and CLI integration
- Test framework setup
- CI/CD pipeline with GitHub Actions
- Linting with golangci-lint
- Code coverage with Codecov
- Dependabot configuration

[Unreleased]: https://github.com/rithyhuot/vibe/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/rithyhuot/vibe/releases/tag/v0.1.0
[0.0.0]: https://github.com/rithyhuot/vibe/commit/c69c8b1
