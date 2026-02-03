# Architecture

This document describes the design and architecture of vibecli, a production-quality Go CLI tool for streamlining developer workflows.

## Table of Contents

- [Overview](#overview)
- [Package Structure](#package-structure)
- [Component Interactions](#component-interactions)
- [Design Patterns](#design-patterns)
- [Data Flow](#data-flow)
- [Integration Patterns](#integration-patterns)
- [Error Handling Strategy](#error-handling-strategy)
- [Caching Strategy](#caching-strategy)
- [Security Considerations](#security-considerations)

## Overview

vibecli is designed around a **command-driven architecture** that integrates three core external services:

1. **ClickUp** - Project management and task tracking
2. **GitHub** - Source code management and pull requests
3. **CircleCI** - Continuous integration and deployment

The architecture emphasizes:

- **Separation of concerns** through clear package boundaries
- **Consistent patterns** for command implementation
- **Dependency injection** via context objects
- **Caching** for performance optimization
- **Security** through input validation and sanitization

## Package Structure

```
vibe/
├── cmd/vibe/              # Application entry point
│   └── main.go            # CLI initialization, root command setup
├── internal/              # Private application code
│   ├── commands/          # Command implementations (Cobra commands)
│   │   ├── comment.go     # Add comments to ClickUp tickets
│   │   ├── init.go        # Initialize configuration
│   │   ├── issue*.go      # GitHub issue management
│   │   ├── pr*.go         # GitHub pull request management
│   │   ├── ci*.go         # CircleCI integration
│   │   ├── ticket.go      # View ticket details
│   │   ├── workon.go      # Start work on ticket (create branch)
│   │   ├── start.go       # Interactive ticket selection
│   │   ├── merge.go       # PR merge automation
│   │   └── skills.go      # Claude Code skills management
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config struct and loading
│   │   └── defaults.go    # Default configuration values
│   ├── models/            # Data structures
│   │   ├── task.go        # ClickUp task models
│   │   ├── issue.go       # GitHub issue models
│   │   └── pr.go          # Pull request models
│   ├── services/          # External service integrations
│   │   ├── clickup/       # ClickUp API client
│   │   ├── github/        # GitHub API client (REST + GraphQL)
│   │   ├── circleci/      # CircleCI API client
│   │   ├── claude/        # Claude API integration
│   │   └── git/           # Git operations via go-git
│   ├── utils/             # Shared utilities
│   │   ├── cache.go       # In-memory caching with TTL
│   │   ├── http.go        # HTTP client utilities
│   │   ├── validation.go  # Input validation and sanitization
│   │   └── branch.go      # Branch name generation
│   ├── ui/                # User interface components
│   │   ├── colors.go      # Color definitions
│   │   ├── spinner.go     # Loading indicators
│   │   └── prompt.go      # Interactive prompts
│   └── skills/            # Claude Code skill definitions
│       └── *.skill.md     # Markdown skill files
└── .vibe.yaml            # Project configuration example
```

### Key Package Responsibilities

**cmd/vibe/**

- Application bootstrap
- Root command initialization
- Global flags and middleware

**internal/commands/**

- Individual command implementations
- Command-specific flags and validation
- User interaction and output formatting

**internal/services/**

- External API integration
- HTTP request/response handling
- Service-specific error handling

**internal/models/**

- Data transfer objects (DTOs)
- JSON marshaling/unmarshaling
- Model validation

**internal/utils/**

- Cross-cutting concerns
- Reusable helper functions
- Caching and HTTP utilities

**internal/config/**

- Configuration file parsing
- Environment variable handling
- Configuration validation

## Component Interactions

```
┌─────────────────────────────────────────────────────────────┐
│                         User / CLI                           │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Commands Layer                            │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ ticket  │  │ workon  │  │   pr    │  │   ci    │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
└───────┼────────────┼────────────┼────────────┼─────────────┘
        │            │            │            │
        └────────────┴────────────┴────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  CommandContext                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Config, GitRepo, ClickUpClient, GitHubClient,        │   │
│  │ ClaudeClient                                         │   │
│  └──────────────────────────────────────────────────────┘   │
└───────────┬──────────────┬──────────────┬──────────────────┘
            │              │              │
            ▼              ▼              ▼
┌─────────────────┐ ┌─────────────┐ ┌──────────────┐
│  ClickUp API    │ │ GitHub API  │ │  Claude API  │
│                 │ │             │ │              │
│ • Tasks         │ │ • PRs       │ │ • AI Gen     │
│ • Comments      │ │ • Issues    │ │ • Review     │
│ • Workspaces    │ │ • Reviews   │ │              │
└─────────────────┘ └─────────────┘ └──────────────┘

Note: CircleCI integration exists in commands but not in CommandContext
```

### Interaction Flow Example: `vibe workon`

1. **User runs command**: `vibe workon abc123`
2. **Command layer**: `workon.go` receives ticket ID
3. **Validation**: Ticket ID format validated via `utils.IsTicketID()`
4. **ClickUp fetch**: `ClickUpClient.GetTask()` retrieves ticket
5. **Cache check**: Result cached in `utils.Cache` with TTL
6. **Branch generation**: `utils.GenerateBranchName()` creates branch name
7. **Git operation**: `GitRepo.CreateBranch()` creates branch
8. **Status update**: `ClickUpClient.UpdateTask()` sets status (configured in `defaults.status`)
9. **Output**: Formatted success message displayed to user

## Design Patterns

### 1. Command Factory Pattern

Every command follows a consistent factory pattern:

```go
func NewCommandCommand(ctx *CommandContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "command [args]",
        Short: "One-liner description",
        Long:  `Detailed description with examples`,
        RunE: func(_ *cobra.Command, args []string) error {
            return runCommand(ctx, args)
        },
    }
    return cmd
}
```

**Benefits:**

- Consistent command structure
- Easy testing and mocking
- Clear separation of setup vs execution

### 2. Context Pattern

`CommandContext` provides dependency injection:

```go
type CommandContext struct {
    Config        *config.Config
    ClickUpClient clickup.Client
    GitHubClient  github.Client
    GitRepo       git.Repository
    ClaudeClient  claude.Client
}
```

**Benefits:**

- Explicit dependencies
- Easy mocking for tests
- Shared resources across commands

### 3. Options Pattern

Commands with flags use options structs:

```go
type PRCreateOptions struct {
    Title       string
    Body        string
    Draft       bool
    BaseBranch  string
}
```

**Benefits:**

- Type-safe flag handling
- Clear command configuration
- Easy to extend with new flags

### 4. Service Layer Pattern

External integrations are abstracted behind interfaces:

```go
type Client interface {
    GetTask(ctx context.Context, taskID string) (*models.Task, error)
    UpdateTask(ctx context.Context, taskID string, req *TaskUpdateRequest) error
}
```

**Benefits:**

- Testable with mocks
- Swappable implementations
- Clear contract definition

### 5. Strategy Pattern

GitHub client supports multiple modes (API, CLI, Auto):

```go
type GitHubClient interface {
    ListPRs(ctx context.Context, opts *ListPRsOptions) ([]*PullRequest, error)
}

// API mode: Uses GitHub REST/GraphQL API directly
// CLI mode: Uses GitHub CLI (gh) under the hood
// Auto mode: Uses CLI if available, falls back to API
```

**Benefits:**

- Flexible authentication methods
- Adapts to environment (local vs CI/CD)
- Runtime mode switching

### 6. In-Memory Caching

Caching is implemented via a simple in-memory cache utility:

```go
type Cache struct {
    items map[string]*CacheEntry
    mu    sync.RWMutex
    ttl   time.Duration
}

// Usage example
cache := utils.NewCache(5 * time.Minute)
if value, found := cache.Get(key); found {
    return value
}
// Fetch from API and cache
cache.Set(key, result)
```

**Benefits:**

- Simple TTL-based expiration
- Thread-safe with sync.RWMutex
- No external dependencies

**Note:** Caching is in-memory only. There is no persistent file-based cache.

## Data Flow

### Ticket-based Workflow

```
User: vibe workon abc123
       │
       ▼
┌──────────────────┐
│ Validate Input   │
│ • Format check   │
│ • Sanitization   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐     ┌──────────────┐
│ Check Cache      │────▶│ Cache Hit?   │
└────────┬─────────┘     └──────┬───────┘
         │ Miss                  │ Hit
         ▼                       ▼
┌──────────────────┐     ┌──────────────┐
│ ClickUp API Call │     │ Return Cache │
│ GET /task/{id}   │     └──────────────┘
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Parse Response   │
│ • JSON decode    │
│ • Model mapping  │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Store in Cache   │
│ TTL: 5 minutes   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Generate Branch  │
│ user/id/desc     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Git Operations   │
│ • Create branch  │
│ • Checkout       │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Update Task      │
│ Status: Progress │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Display Success  │
└──────────────────┘
```

### PR Creation Flow

```
User: vibe pr create
       │
       ▼
┌──────────────────┐
│ Get Current      │
│ Branch & Commits │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Extract Ticket   │
│ ID from Branch   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Fetch ClickUp    │
│ Task (cached)    │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐     ┌──────────────┐
│ AI Description?  │────▶│ Claude API   │
│ (Optional)       │     │ Analyze Diff │
└────────┬─────────┘     └──────────────┘
         │
         ▼
┌──────────────────┐
│ Load PR Template │
│ (if exists)      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Generate PR Body │
│ • Summary        │
│ • Description    │
│ • Testing        │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ GitHub API Call  │
│ POST /pulls      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Display PR URL   │
└──────────────────┘
```

## Integration Patterns

### ClickUp Integration

**Authentication:** Bearer token via `Authorization` header

**Key Endpoints:**

- `GET /api/v2/task/{task_id}` - Fetch task details
- `PUT /api/v2/task/{task_id}` - Update task
- `POST /api/v2/task/{task_id}/comment` - Add comment
- `GET /api/v2/team/{team_id}/space` - List spaces
- `GET /api/v2/folder/{folder_id}/list` - List lists

**Rate Limiting:** Handled via exponential backoff

**Caching:** 5-minute TTL for task data

### GitHub Integration

**Authentication:**
- **API mode:** Personal access token via `Authorization: token {token}`
- **CLI mode:** Uses `gh auth` credentials (no token in config needed)
- **Auto mode:** Tries CLI first, falls back to API

**Three Integration Modes:**

1. **API Mode** - Direct REST/GraphQL API calls with token
   - Best for CI/CD and automation
   - Requires token in config or environment

2. **CLI Mode** - Uses GitHub CLI (`gh`) under the hood
   - Best for local development
   - No token needed in config
   - Executes `gh` commands like `gh pr create`, `gh pr view`, etc.

3. **Auto Mode** (default) - Automatic mode selection
   - Uses CLI if `gh auth status` succeeds
   - Falls back to API if CLI not available
   - Best for most users

**Key API Endpoints (API mode):**

- `POST /repos/{owner}/{repo}/pulls` - Create PR
- `GET /repos/{owner}/{repo}/pulls/{number}` - Get PR
- `PATCH /repos/{owner}/{repo}/pulls/{number}` - Update PR
- `POST /repos/{owner}/{repo}/issues/{number}/comments` - Add comment
- GraphQL `/graphql` - Complex queries

**Key CLI Commands (CLI mode):**

- `gh pr create` - Create PR
- `gh pr view` - Get PR details
- `gh pr edit` - Update PR
- `gh pr comment` - Add comment

**Rate Limiting:** 5,000 requests/hour (authenticated, both modes)

**Caching:** In-memory only (no persistent cache)

### CircleCI Integration

**Authentication:** Circle-Token header

**Key Endpoints:**

- `GET /api/v2/project/{project-slug}/pipeline` - List pipelines
- `GET /api/v2/workflow/{workflow-id}` - Get workflow
- `GET /api/v2/workflow/{workflow-id}/job` - List jobs
- `GET /api/v2/project/{project-slug}/job/{job-number}` - Job details
- `GET /api/v2/project/{project-slug}/job/{job-number}/artifacts` - Artifacts

**Rate Limiting:** Varies by plan

**Caching:** 2-minute TTL for CI data

### Git Integration

**Library:** go-git (pure Go implementation)

**Operations:**

- Branch creation and switching
- Commit history inspection
- Remote detection and parsing
- Working tree status

**No external git binary required** - all operations in-process

## Error Handling Strategy

### Layered Error Handling

1. **Service Layer**: Wrap errors with context

   ```go
   if err != nil {
       return nil, fmt.Errorf("failed to fetch task %s: %w", taskID, err)
   }
   ```

2. **Command Layer**: User-friendly messages

   ```go
   if err != nil {
       return fmt.Errorf("could not create branch: %w\nTip: Check if branch already exists", err)
   }
   ```

3. **Main Layer**: Exit codes and output

   ```go
   if err := cmd.Execute(); err != nil {
       os.Exit(1)
   }
   ```

### Error Types

- **Validation Errors**: User input issues (clear messages, suggestions)
- **API Errors**: External service failures (retry logic, fallbacks)
- **Git Errors**: Repository issues (helpful tips)
- **Configuration Errors**: Setup problems (guidance to fix)

### Retry Logic

Exponential backoff for transient failures:

```go
for attempt := 0; attempt < maxRetries; attempt++ {
    resp, err := makeRequest()
    if err == nil {
        return resp, nil
    }
    if !isRetryable(err) {
        return nil, err
    }
    time.Sleep(backoff(attempt))
}
```

## Caching Strategy

### In-Memory Cache

**Implementation:** `sync.Map` with TTL tracking

**Key Patterns:**

```go
type CacheEntry struct {
    Value      interface{}
    Expiration time.Time
}

func (c *Cache) Get(key string) interface{}
func (c *Cache) Set(key string, value interface{}, ttl time.Duration)
func (c *Cache) Clear()
```

### Cache Implementation

Currently, vibecli implements **minimal in-memory caching**:

- **Sprint Cache**: 1 hour TTL for sprint folder data
- **No persistent cache**: All caching is in-memory only
- **No per-service caching**: API services (ClickUp, GitHub, CircleCI) make direct calls without caching layers

**Note**: The architecture supports adding per-service caching in the future with configurable TTLs.

### Cache Invalidation

**Automatic Expiration:**

- TTL-based cleanup on access
- Expired entries removed when accessed

**Manual Invalidation:**

- Currently limited - only sprint cache can be cleared programmatically
- No persistent cache files to delete (all in-memory)
- Application restart clears all caches

**Future Enhancement:**

- `--no-cache` flag for bypassing cache on specific commands
- `vibe cache clear` command for manual cache invalidation

## Security Considerations

### Input Validation

**Branch Names:**

```go
func ValidateBranchName(name string) error {
    if name == "" {
        return fmt.Errorf("branch name cannot be empty")
    }

    // Prevent command injection - blocks: ;&|><$`(){}[]\
    if shellMetacharPattern.MatchString(name) {
        return fmt.Errorf("branch name contains unsafe characters")
    }

    // Prevent path traversal
    if strings.Contains(name, "..") {
        return fmt.Errorf("branch name contains invalid '..' sequence")
    }

    // No leading/trailing slashes
    if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
        return fmt.Errorf("branch name cannot start or end with '/'")
    }

    return nil
}
```

**Ticket IDs:**

```go
func IsTicketID(id string) bool {
    // ClickUp ticket IDs are 9 alphanumeric characters
    matched, _ := regexp.MatchString(`^[a-z0-9]{9}$`, id)
    return matched
}
```

### Token Handling

- **Never logged**: Tokens excluded from debug output
- **Environment preferred**: `CLICKUP_API_TOKEN`, `GITHUB_TOKEN`
- **Config file security**: New config files created with 0600 permissions
- **No default tokens**: Users must provide their own

### API Security

- **HTTPS only**: All external APIs use TLS
- **Token rotation**: Encourage regular token updates
- **Minimal permissions**: Documentation specifies minimum required scopes
- **No token storage**: Tokens read from env/config, not persisted elsewhere

### Git Security

- **No destructive operations by default**: Warn before force push, hard reset
- **Branch name sanitization**: Prevent command injection
- **Remote verification**: Validate remote URLs before push

## Performance Optimizations

### Parallel Operations

Where possible, operations run concurrently:

```go
go fetchTask()
go fetchPR()
```

### Lazy Loading

Data fetched only when needed:

- PR lists loaded on first access
- CI status checked on demand

### Minimal API Calls

- Batch operations where API supports
- Use GraphQL for complex GitHub queries (single request vs multiple REST calls)
- Minimal caching currently implemented (sprint data only)

### Streaming

Large responses streamed rather than buffered:

- CI logs streamed to terminal
- Large PR descriptions handled efficiently

## Testing Strategy

### Unit Tests

- Service layer fully mocked
- Utilities independently tested
- Models validated for JSON correctness

### Integration Tests

- Test against real APIs (optional, via env flag)
- Mock HTTP server for deterministic tests
- Git operations tested against temp repos

### End-to-End Tests

- Smoke tests for core workflows
- Test configuration loading
- Validate command help text

## Future Architecture Considerations

### Potential Improvements

1. **Plugin System**: Allow custom commands and integrations
2. **Database**: Persistent cache with SQLite for offline support
3. **Webhooks**: React to external events (PR merged, CI failed)
4. **Multi-workspace**: Support multiple ClickUp workspaces simultaneously
5. **GraphQL Server**: Expose vibecli data for other tools
6. **Background Sync**: Daemon mode for continuous cache updates

### Scalability

Current architecture supports:

- **Small teams**: < 10 developers, low API volume
- **Medium teams**: 10-50 developers, moderate API volume

For larger teams, consider:

- Shared cache (Redis)
- Rate limit coordination
- Dedicated backend service

## Contributing

When contributing to vibecli architecture:

1. **Follow existing patterns**: Use command factory, options, context
2. **Add tests**: Unit tests for new functionality
3. **Update docs**: Keep this document current
4. **Consider performance**: Use caching, avoid unnecessary API calls
5. **Security first**: Validate all inputs, sanitize outputs

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## References

- [Cobra Documentation](https://cobra.dev/) - CLI framework
- [go-git](https://github.com/go-git/go-git) - Git operations
- [ClickUp API v2](https://clickup.com/api) - Task management
- [GitHub REST API](https://docs.github.com/en/rest) - Code hosting
- [GitHub GraphQL API](https://docs.github.com/en/graphql) - Advanced queries
- [CircleCI API v2](https://circleci.com/docs/api/v2/) - CI/CD integration
