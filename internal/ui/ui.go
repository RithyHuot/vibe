// Package ui provides user interface utilities for the vibe CLI.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Color scheme
var (
	Success = color.New(color.FgGreen)
	Error   = color.New(color.FgRed)
	Warning = color.New(color.FgYellow)
	Info    = color.New(color.FgBlue)
	Dim     = color.New(color.Faint)
	Bold    = color.New(color.Bold)
	Cyan    = color.New(color.FgCyan)
)

// CreateSpinner creates a new spinner with consistent styling
func CreateSpinner(text string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" %s", text)
	return s
}

// ShowSuccess prints a success message with checkmark
func ShowSuccess(message string) {
	_, _ = Success.Printf("✓ %s\n", message)
}

// ShowError prints an error message with X mark
func ShowError(message string) {
	_, _ = Error.Printf("✗ %s\n", message)
}

// ShowWarning prints a warning message with warning symbol
func ShowWarning(message string) {
	_, _ = Warning.Printf("⚠ %s\n", message)
}

// ShowInfo prints an info message with info symbol
func ShowInfo(message string) {
	_, _ = Info.Printf("ℹ %s\n", message)
}

// ShowBox prints content in a simple box
func ShowBox(content string, title string) {
	width := 60
	if len(title) > 0 {
		width = maxInt(width, len(title)+4)
	}

	// Calculate content lines and adjust width if needed
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if len(line)+4 > width {
			width = len(line) + 4
		}
	}

	border := strings.Repeat("─", width-2)

	fmt.Println()
	if title != "" {
		_, _ = Cyan.Printf("┌─ %s %s┐\n", title, strings.Repeat("─", width-len(title)-5))
	} else {
		_, _ = Cyan.Printf("┌%s┐\n", border)
	}

	for _, line := range lines {
		padding := width - len(line) - 4
		fmt.Printf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	_, _ = Cyan.Printf("└%s┘\n", border)
	fmt.Println()
}

// Table creates a formatted table
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

// Render prints the table
func (t *Table) Render() {
	// Simple table rendering without tablewriter
	// Print header
	headerLine := strings.Join(t.headers, " | ")
	fmt.Println(headerLine)
	fmt.Println(strings.Repeat("-", len(headerLine)))

	// Print rows
	for _, row := range t.rows {
		fmt.Println(strings.Join(row, " | "))
	}
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// FormatList formats a list of items with bullets
func FormatList(items []string, bullet string) string {
	if bullet == "" {
		bullet = "•"
	}

	var result []string
	for _, item := range items {
		result = append(result, fmt.Sprintf("  %s %s", bullet, item))
	}

	return strings.Join(result, "\n")
}

// ConfirmPrompt prints a confirmation prompt (note: doesn't actually prompt, just formats)
func ConfirmPrompt(message string, defaultYes bool) string {
	if defaultYes {
		return fmt.Sprintf("%s (Y/n): ", message)
	}
	return fmt.Sprintf("%s (y/N): ", message)
}

// PrintSeparator prints a horizontal separator
func PrintSeparator(char string, length int) {
	if char == "" {
		char = "─"
	}
	if length <= 0 {
		length = 60
	}
	_, _ = Dim.Println(strings.Repeat(char, length))
}

// PrintHeader prints a section header
func PrintHeader(text string) {
	fmt.Println()
	_, _ = Bold.Println(text)
	fmt.Println()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
