// Package vibe provides embedded resources for the vibe CLI tool.
package vibe

import "embed"

// SkillsFS contains the embedded skills directory for Claude Code integration.
//
//go:embed skills
var SkillsFS embed.FS
