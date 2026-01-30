// Package claude provides a client for interacting with Claude AI CLI.
package claude

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rithyhuot/vibe/internal/utils"
)

const (
	baseURL      = "https://api.anthropic.com/v1"
	defaultModel = "claude-3-5-sonnet-20241022"
	apiVersion   = "2023-06-01"
)

// Client interface defines Claude AI operations
type Client interface {
	GenerateDescription(ctx context.Context, title string) (string, error)
	EnhancePRDescription(ctx context.Context, title, changes string) (string, error)
	SummarizeChanges(ctx context.Context, diff string) (string, error)
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithContext(ctx context.Context, prompt, context string) (string, error)
}

// HTTPClient implements the Client interface using Anthropic API
type HTTPClient struct {
	httpClient *utils.HTTPClient
	apiKey     string
	model      string
}

// NewClient creates a new Claude HTTP client
func NewClient(apiKey string) *HTTPClient {
	return &HTTPClient{
		httpClient: utils.NewHTTPClient(0).WithUserAgent("vibe"),
		apiKey:     apiKey,
		model:      defaultModel,
	}
}

// headers returns the common headers for Claude API requests
func (c *HTTPClient) headers() map[string]string {
	return map[string]string{
		"x-api-key":         c.apiKey,
		"anthropic-version": apiVersion,
		"content-type":      "application/json",
	}
}

// Generate generates text from a prompt
func (c *HTTPClient) Generate(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/messages", baseURL)

	req := Request{
		Model:     c.model,
		MaxTokens: 1024,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	var resp Response
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, req, &resp, c.headers())
	if err != nil {
		return "", fmt.Errorf("failed to generate text: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return resp.Content[0].Text, nil
}

// GenerateWithContext generates text with additional context
func (c *HTTPClient) GenerateWithContext(ctx context.Context, prompt, context string) (string, error) {
	url := fmt.Sprintf("%s/messages", baseURL)

	req := Request{
		Model:     c.model,
		MaxTokens: 2048,
		System:    context,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	var resp Response
	err := c.httpClient.DoJSONRequest(ctx, "POST", url, req, &resp, c.headers())
	if err != nil {
		return "", fmt.Errorf("failed to generate text: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return resp.Content[0].Text, nil
}

// GenerateDescription generates a ticket description from a title
func (c *HTTPClient) GenerateDescription(ctx context.Context, title string) (string, error) {
	prompt := fmt.Sprintf(`Generate a concise ticket description with acceptance criteria for this task:

Title: %s

Provide:
1. A brief overview (1-2 sentences)
2. Acceptance criteria as a bulleted list
3. Keep it practical and focused

Format the output in markdown.`, title)

	return c.Generate(ctx, prompt)
}

// EnhancePRDescription generates an enhanced PR description
func (c *HTTPClient) EnhancePRDescription(ctx context.Context, title, changes string) (string, error) {
	prompt := fmt.Sprintf(`Generate a clear and concise PR description for this change:

Title: %s

Changes summary:
%s

Provide:
1. A brief summary (2-3 sentences) of what changed and why
2. Key changes as bullet points
3. Keep it technical but accessible

Format the output in markdown.`, title, changes)

	return c.Generate(ctx, prompt)
}

// SummarizeChanges summarizes git diff output
func (c *HTTPClient) SummarizeChanges(ctx context.Context, diff string) (string, error) {
	prompt := `Analyze this git diff and provide a concise summary of the changes:

Focus on:
- What files were modified
- What functionality was added/changed/removed
- Any notable patterns or refactorings

Keep the summary to 3-5 bullet points.`

	return c.GenerateWithContext(ctx, prompt, fmt.Sprintf("Git diff:\n%s", diff))
}

// CLIClient implements the Client interface using claude CLI
type CLIClient struct{}

// NewCLIClient creates a new Claude CLI client
func NewCLIClient() *CLIClient {
	return &CLIClient{}
}

// IsClaudeAvailable checks if claude CLI is available
func IsClaudeAvailable() bool {
	cmd := exec.Command("which", "claude")
	return cmd.Run() == nil
}

// Generate generates text using claude CLI
func (c *CLIClient) Generate(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "claude", "--print", prompt)
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GenerateWithContext generates text with context using claude CLI
func (c *CLIClient) GenerateWithContext(ctx context.Context, prompt, context string) (string, error) {
	cmd := exec.CommandContext(ctx, "claude", "--print", prompt)
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(context)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GenerateDescription generates a ticket description from a title
func (c *CLIClient) GenerateDescription(ctx context.Context, title string) (string, error) {
	prompt := fmt.Sprintf(`Generate a concise ticket description with acceptance criteria for this task:

Title: %s

Provide:
1. A brief overview (1-2 sentences)
2. Acceptance criteria as a bulleted list
3. Keep it practical and focused

Format the output in markdown.`, title)

	return c.Generate(ctx, prompt)
}

// EnhancePRDescription generates an enhanced PR description
func (c *CLIClient) EnhancePRDescription(ctx context.Context, title, changes string) (string, error) {
	prompt := fmt.Sprintf(`Generate a clear and concise PR description for this change:

Title: %s

Changes summary:
%s

Provide:
1. A brief summary (2-3 sentences) of what changed and why
2. Key changes as bullet points
3. Keep it technical but accessible

Format the output in markdown.`, title, changes)

	return c.Generate(ctx, prompt)
}

// SummarizeChanges summarizes git diff output
func (c *CLIClient) SummarizeChanges(ctx context.Context, diff string) (string, error) {
	prompt := `Analyze this git diff and provide a concise summary of the changes:

Focus on:
- What files were modified
- What functionality was added/changed/removed
- Any notable patterns or refactorings

Keep the summary to 3-5 bullet points.`

	return c.GenerateWithContext(ctx, prompt, fmt.Sprintf("Git diff:\n%s", diff))
}

// NewClientAuto creates a Claude client, preferring CLI if available, otherwise API
func NewClientAuto(apiKey string) Client {
	if IsClaudeAvailable() {
		return NewCLIClient()
	}
	if apiKey != "" {
		return NewClient(apiKey)
	}
	return nil
}
