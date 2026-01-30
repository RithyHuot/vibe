package config

import "time"

// GitHub mode constants
const (
	GitHubModeAPI  = "api"
	GitHubModeCLI  = "cli"
	GitHubModeAuto = "auto"
)

// Config represents the application configuration
type Config struct {
	ClickUp    ClickUpConfig     `yaml:"clickup" mapstructure:"clickup" validate:"required"`
	GitHub     GitHubConfig      `yaml:"github" mapstructure:"github" validate:"required"`
	Git        GitConfig         `yaml:"git" mapstructure:"git" validate:"required"`
	CircleCI   CircleCIConfig    `yaml:"circleci" mapstructure:"circleci"`
	Claude     ClaudeConfig      `yaml:"claude" mapstructure:"claude"`
	Workspaces []WorkspaceConfig `yaml:"workspaces" mapstructure:"workspaces" validate:"required,min=1"`
	Defaults   DefaultsConfig    `yaml:"defaults" mapstructure:"defaults"`
	AI         AIConfig          `yaml:"ai" mapstructure:"ai"`
	UI         UIConfig          `yaml:"ui" mapstructure:"ui"`
}

// ClickUpConfig holds ClickUp API configuration
type ClickUpConfig struct {
	APIToken    string `yaml:"api_token" mapstructure:"api_token" validate:"required"`
	UserID      string `yaml:"user_id" mapstructure:"user_id" validate:"required"`
	WorkspaceID string `yaml:"workspace_id" mapstructure:"workspace_id" validate:"required"`
	TeamID      string `yaml:"team_id" mapstructure:"team_id" validate:"required"`
}

// GitHubConfig holds GitHub configuration
type GitHubConfig struct {
	Token    string `yaml:"token" mapstructure:"token"`
	Username string `yaml:"username" mapstructure:"username" validate:"required"`
	Owner    string `yaml:"owner" mapstructure:"owner" validate:"required"`
	Repo     string `yaml:"repo" mapstructure:"repo" validate:"required"`
	Mode     string `yaml:"mode" mapstructure:"mode"` // "api", "cli", or "auto" (default: auto)
}

// GitConfig holds Git-related configuration
type GitConfig struct {
	BranchPrefix string `yaml:"branch_prefix" mapstructure:"branch_prefix" validate:"required"`
	BaseBranch   string `yaml:"base_branch" mapstructure:"base_branch" validate:"required"`
}

// CircleCIConfig holds CircleCI API configuration
type CircleCIConfig struct {
	APIToken string `yaml:"api_token" mapstructure:"api_token"`
}

// ClaudeConfig holds Claude AI configuration
type ClaudeConfig struct {
	APIKey string `yaml:"api_key" mapstructure:"api_key"`
}

// WorkspaceConfig represents a ClickUp workspace
type WorkspaceConfig struct {
	Name           string   `yaml:"name" mapstructure:"name" validate:"required"`
	FolderID       string   `yaml:"folder_id" mapstructure:"folder_id" validate:"required"`
	SprintPatterns []string `yaml:"sprint_patterns" mapstructure:"sprint_patterns" validate:"required,min=1"`
}

// DefaultsConfig holds default values
type DefaultsConfig struct {
	Status string `yaml:"status" mapstructure:"status"`
}

// AIConfig holds AI feature configuration
type AIConfig struct {
	Enabled              bool `yaml:"enabled" mapstructure:"enabled"`
	GenerateDescriptions bool `yaml:"generate_descriptions" mapstructure:"generate_descriptions"`
}

// UIConfig holds UI preferences
type UIConfig struct {
	ColorEnabled bool `yaml:"color_enabled" mapstructure:"color_enabled"`
}

// HTTPClientConfig holds HTTP client configuration
type HTTPClientConfig struct {
	Timeout     time.Duration
	MaxRetries  int
	UserAgent   string
	EnableDebug bool
}

// DefaultHTTPClientConfig returns default HTTP client configuration
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		UserAgent:   "vibe",
		EnableDebug: false,
	}
}
