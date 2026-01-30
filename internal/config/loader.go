package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Load reads and validates the configuration
// Priority order: CLI flag > local .vibe.yaml > global config > env vars > defaults
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config location
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "vibe")
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Environment variable overrides
	v.SetEnvPrefix("VIBE")
	v.AutomaticEnv()

	// Bind specific env variables
	_ = v.BindEnv("clickup.api_token", "VIBE_CLICKUP_TOKEN")
	_ = v.BindEnv("github.token", "VIBE_GITHUB_TOKEN")
	_ = v.BindEnv("circleci.api_token", "VIBE_CIRCLECI_TOKEN")
	_ = v.BindEnv("claude.api_key", "VIBE_CLAUDE_API_KEY")

	// Read global config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Try to merge local .vibe.yaml if it exists
	localConfigPath := findLocalConfig()
	if localConfigPath != "" {
		if err := v.MergeInConfig(); err == nil {
			// Successfully merged, but viper won't automatically pick up the file
			// So we need to manually merge it
			localViper := viper.New()
			localViper.SetConfigFile(localConfigPath)
			if err := localViper.ReadInConfig(); err == nil {
				// Merge the local config settings into the main viper instance
				for _, key := range localViper.AllKeys() {
					v.Set(key, localViper.Get(key))
				}
			}
		}
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	if cfg.Defaults.Status == "" {
		cfg.Defaults.Status = "In Progress"
	}
	if cfg.UI.ColorEnabled {
		cfg.UI.ColorEnabled = true
	}

	// Set GitHub mode default
	if cfg.GitHub.Mode == "" {
		cfg.GitHub.Mode = GitHubModeAuto
	}

	// Validate GitHub mode
	if err := validateGitHubMode(&cfg.GitHub); err != nil {
		return nil, err
	}

	// Validate
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// findLocalConfig searches for .vibe.yaml in the current directory
func findLocalConfig() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	localConfigPath := filepath.Join(cwd, ".vibe.yaml")
	if _, err := os.Stat(localConfigPath); err == nil {
		return localConfigPath
	}

	return ""
}

// GetConfigDir returns the default config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "vibe"), nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// validateGitHubMode validates the GitHub configuration mode
func validateGitHubMode(cfg *GitHubConfig) error {
	// Validate mode value
	validModes := map[string]bool{
		GitHubModeAPI:  true,
		GitHubModeCLI:  true,
		GitHubModeAuto: true,
	}

	if !validModes[cfg.Mode] {
		return fmt.Errorf("invalid github.mode: %s (must be '%s', '%s', or '%s')",
			cfg.Mode, GitHubModeAPI, GitHubModeCLI, GitHubModeAuto)
	}

	// Mode-specific validation
	switch cfg.Mode {
	case GitHubModeCLI:
		// In CLI mode, token is not needed (will be ignored)
		// Keep the token in config as it may be useful for fallback or documentation
		// but it won't be used

	case GitHubModeAPI:
		// API mode requires a token
		if cfg.Token == "" {
			return fmt.Errorf("github.token is required when using API mode")
		}

	case GitHubModeAuto:
		// Auto mode can work with either CLI or token
		// No validation needed here - will be checked at runtime
	}

	return nil
}
