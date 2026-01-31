---
name: add-command-skill
description: Add a new vibe command with associated Claude Code skill. Use when adding a new command to the vibe CLI (e.g., "add a new command called status").
argument-hint: <command-name>
allowed-tools: Write, Edit, Read, Bash(make:*), Glob, Grep, AskUserQuestion
---

# Add New Command with Associated Skill

This skill automates adding a new command to the vibe CLI along with its associated Claude Code skill.

## Prerequisites

Before starting, gather:

1. **Command name** (e.g., "status", "deploy", "review")
2. **Command purpose** (what does it do?)
3. **Command arguments** (what inputs does it take?)
4. **Skill description** (when should Claude use this?)

## Steps to Execute

### 1. Gather Requirements

Ask the user:

- What is the command name?
- What does the command do?
- What arguments/flags does it need?
- When should the skill be invoked?
- What output should it provide?
- What additional information you would like to provide to generate this command?

### 2. Create Command File

**File**: `internal/commands/<command-name>.go`

**Template**:

```go
package commands

import (
    "fmt"
    "github.com/spf13/cobra"
)

// New<Command>Command creates the <command> command
func New<Command>Command(ctx *CommandContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "<command> [args]",
        Short: "Brief description",
        Long: `Detailed description with examples.`,
        RunE: func(cobraCmd *cobra.Command, args []string) error {
            // Get context from the command's context value (set by PreRunE)
            ctx = getCommandContext(cobraCmd, ctx)
            return run<Command>(ctx, args)
        },
    }

    // Add flags here
    // cmd.Flags().StringVar(&opts.Flag, "flag", "", "Description")

    return cmd
}

func run<Command>(ctx *CommandContext, args []string) error {
    // Implementation
    return nil
}
```

### 3. Register Command in main.go

**File**: `cmd/vibe/main.go`

**Add in `addConfigDependentCommands()` function** (after creating the command with New<Command>Command):

```go
// <Command> command
<command>Cmd := commands.New<Command>Command(dummyCtx)
<command>Cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
    ctx, err := getContext()
    if err != nil {
        return err
    }
    // Store context in cobra's context so RunE can access it
    cmd.SetContext(context.WithValue(cmd.Context(), commandContextKey, ctx))
    return nil
}
```

**Update `rootCmd.AddCommand()` at the end**:

```go
rootCmd.AddCommand(..., <command>Cmd)
```

### 4. Create Skill Directory and File

**File**: `skills/vibe-<command>/SKILL.md`

**Template**:

```markdown
---
name: vibe-<command>
description: <Brief description>. Use when <trigger condition>.
argument-hint: [arguments]
allowed-tools: Bash(vibe:<command>*)
---

# <Command Title>

<Detailed description of what the skill does>

## Steps

1. <Step 1>
2. <Step 2>
3. <Step 3>

## When to Use This Skill

Use this skill when:
- <Condition 1>
- <Condition 2>
- <Condition 3>

## Examples

```bash
# Example 1
vibe <command> <args>

# Example 2
vibe <command> --flag value
```

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| <Error> | <Cause> | <Solution> |

```

### 5. Update README.md

**Sections to update**:

1. **Features section** - Add relevant feature bullet if applicable
2. **Commands section** - Add new command documentation:
   ```markdown
   ### `vibe <command> [args]`

   <Description>

   ```bash
   # Example usage
   vibe <command> <args>
   ```

   ```

3. **Available Skills table** - Add row:
   ```markdown
   | `vibe-<command>` | <Description> | "<usage>" |
   ```

1. **Common workflows** (if relevant) - Add to quick start examples

### 6. Update SKILLS.md

**Add new skill section**:

```markdown
### <N>. vibe-<command>

**Purpose**: <Purpose>

**When to use**: <Trigger conditions>

**What it does**:

- <Action 1>
- <Action 2>
- <Action 3>

**Example usage**:

```

User: "<example request>"
Claude: Runs `vibe <command>` and <does something>

```
```

**Important**: Renumber all subsequent skills if inserting in the middle.

### 7. Update internal/skills/installer.go

**In `Uninstall()` function, add to `skillDirs` slice**:

```go
skillDirs := []string{
    // ... existing skills ...
    "vibe-<command>",
    // ... more skills ...
}
```

**In `PrintInstallSuccess()` function, add to skills list**:

```go
fmt.Println("  - vibe-<command>             <Description>")
```

### 8. Update FAQ.md (if relevant)

If the command works without ClickUp or has special considerations, update the FAQ.

### 9. Build and Test

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Install
make install

# Test command
vibe <command> --help

# Install skills
vibe skills

# Verify skill installed
ls ~/.claude/skills/vibe-<command>
```

## Files Checklist

Complete this checklist for every new command:

- [ ] `internal/commands/<command-name>.go` - Command implementation
- [ ] `cmd/vibe/main.go` - Command registration
- [ ] `skills/vibe-<command>/SKILL.md` - Skill definition
- [ ] `README.md` - Updated (features, commands, skills table)
- [ ] `SKILLS.md` - Added skill section
- [ ] `internal/skills/installer.go` - Updated (Uninstall + PrintInstallSuccess)
- [ ] `FAQ.md` - Updated if relevant
- [ ] Built and tested (`make build test lint`)
- [ ] Skills installed and verified (`vibe skills`)

## Important Notes

1. **Command naming**: Use kebab-case (e.g., `vibe-status`, not `vibeStatus`)
2. **Skill naming**: Prefix with `vibe-` (e.g., `vibe-status`)
3. **Autocomplete**: Handled automatically by Cobra
4. **Embedding**: Skills are auto-embedded via `//go:embed skills` directive
5. **Context**: Always get context via `getCommandContext()` in RunE
6. **Numbering**: Keep skill numbering sequential in SKILLS.md

## Example Workflow

```
User: "Add a new command called deploy that deploys the application"

Claude:
1. Asks clarifying questions (arguments, flags, behavior)
2. Creates internal/commands/deploy.go
3. Registers in cmd/vibe/main.go
4. Creates skills/vibe-deploy/SKILL.md
5. Updates README.md (features, commands, skills)
6. Updates SKILLS.md (adds skill section)
7. Updates internal/skills/installer.go
8. Builds and tests: make build test lint
9. Installs skills: vibe skills
10. Reports completion with verification steps
```
