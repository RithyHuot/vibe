---
name: add-claude-skill
description: Add a new Claude Code skill for vibe CLI. Use when adding a skill that wraps existing commands or provides workflow automation (e.g., "add a skill for code review workflow").
argument-hint: <skill-name>
allowed-tools: Write, Edit, Read, Bash(make:*), Glob, Grep, AskUserQuestion
---

# Add New Claude Code Skill

This skill automates adding a new Claude Code skill to the vibe CLI. Use this for skills that:

- Wrap existing vibe commands with AI logic
- Combine multiple commands into workflows
- Add automation or decision-making on top of CLI commands

## Prerequisites

Before starting, gather:

1. **Skill name** (e.g., "deployment-check", "ticket-analysis")
2. **Skill purpose** (what workflow does it automate?)
3. **Trigger conditions** (when should Claude invoke this?)
4. **Commands used** (which vibe commands does it call?)
5. **Tools needed** (Bash, Read, Grep, etc.)

## Steps to Execute

### 1. Gather Requirements

Ask the user:

- What is the skill name?
- What workflow does it automate?
- When should Claude invoke this skill?
- Which vibe commands will it use?
- What tools does it need access to?
- What should the output be?
- What additional information you would like to provide to generate this command?

### 2. Create Skill Directory and File

**File**: `skills/vibe-<skill-name>/SKILL.md`

**Template**:

```markdown
---
name: vibe-<skill-name>
description: <Brief description>. Use when <trigger condition>.
argument-hint: [arguments]
allowed-tools: Bash(vibe:*), Read, Grep, Glob, Write, Edit
---

# <Skill Title>

<Detailed description of what the skill automates>

## Steps

When this skill is invoked:

1. **<Step 1 Title>**
   - <Action description>
   - Command: `vibe <command>`

2. **<Step 2 Title>**
   - <Action description>
   - Analysis: <what to look for>

3. **<Step 3 Title>**
   - <Final action>
   - Output: <what to report>

## When to Use This Skill

Use this skill when:
- <Condition 1>
- <Condition 2>
- <Condition 3>

## Workflow Example

```

User: "<example request>"
Claude:

1. Runs `vibe <command1>`
2. Analyzes output
3. Runs `vibe <command2>` based on analysis
4. Reports findings

```

## Decision Logic

<Describe any decision points in the workflow>

```

If <condition>:
  -> Action A
Else:
  -> Action B

```

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| <Error> | <Cause> | <Solution> |

## Notes

- <Important note 1>
- <Important note 2>
```

### 3. Update README.md

**In Available Skills table**:

```markdown
| `vibe-<skill-name>` | <Description> | "<usage example>" |
```

### 4. Update SKILLS.md

**Add new skill section**:

Position it appropriately (usually at the end unless it fits logically elsewhere).

```markdown
### <N>. vibe-<skill-name>

**Purpose**: <Purpose>

**When to use**:

- <Trigger condition 1>
- <Trigger condition 2>

**What it does**:

- <Action 1>
- <Action 2>
- <Action 3>

**Example usage**:

```

User: "<example request>"
Claude:

1. <Step 1>
2. <Step 2>
3. <Step 3>
4. Reports: <output>

```

**Decision Points**:

- If <condition>: <action>
- Otherwise: <alternative action>

---
```

**Important**: Renumber all subsequent skills if inserting in the middle.

### 5. Update internal/skills/installer.go

**In `Uninstall()` function, add to `skillDirs` slice**:

```go
skillDirs := []string{
    // ... existing skills ...
    "vibe-<skill-name>",
}
```

**In `PrintInstallSuccess()` function, add to skills list**:

```go
fmt.Println("  - vibe-<skill-name>          <Description>")
```

### 6. Update FAQ.md (if relevant)

If the skill relates to common questions or provides workarounds, add to FAQ.

### 7. Build and Test

```bash
# Build with embedded skills
make build

# Install skills
make install

# Install to Claude Code
vibe skills

# Verify skill installed
ls ~/.claude/skills/vibe-<skill-name>

# Check skill content
cat ~/.claude/skills/vibe-<skill-name>/SKILL.md | head -20
```

## Files Checklist

Complete this checklist for every new skill:

- [ ] `skills/vibe-<skill-name>/SKILL.md` - Skill definition
- [ ] `README.md` - Updated (Available Skills table)
- [ ] `SKILLS.md` - Added skill section with proper numbering
- [ ] `internal/skills/installer.go` - Updated (Uninstall + PrintInstallSuccess)
- [ ] `FAQ.md` - Updated if relevant
- [ ] Built and installed (`make build install`)
- [ ] Skills deployed (`vibe skills`)
- [ ] Skill verified in `~/.claude/skills/`

## Skill Types

### Workflow Skills

Combine multiple commands with logic:

- Check status → analyze → take action
- Review code → run tests → report issues

### Analysis Skills

Add intelligence to command output:

- Parse logs for errors
- Summarize ticket information
- Identify patterns in failures

### Decision Skills

Add conditional logic:

- Check if PR is ready → merge or wait
- Analyze CI failure → retry or fix
- Review dependencies → safe or risky

## Allowed Tools Reference

Common tools used in skills:

- `Bash(vibe:*)` - Execute any vibe command
- `Bash(vibe:<command>*)` - Execute specific vibe command only
- `Read` - Read files from repository
- `Grep` - Search code for patterns
- `Glob` - Find files by pattern
- `Write` - Create new files
- `Edit` - Modify existing files

## Important Notes

1. **Skill naming**: Prefix with `vibe-` (e.g., `vibe-deployment-check`)
2. **Embedding**: Skills are auto-embedded via `//go:embed skills` directive
3. **Tools**: Only request tools the skill actually needs
4. **Description**: Be specific about trigger conditions
5. **Examples**: Include realistic user requests

## Example Workflow

```
User: "Add a skill that checks if a PR is ready to merge"

Claude:
1. Asks clarifying questions:
   - What checks should it perform?
   - What should it report?
   - What actions should it recommend?

2. Creates skills/vibe-pr-ready-check/SKILL.md with:
   - Steps: Check CI, reviews, conflicts
   - Decision logic: Ready vs. not ready
   - Output: Summary with action recommendation

3. Updates README.md:
   - Adds to Available Skills table

4. Updates SKILLS.md:
   - Adds new skill section (e.g., #16)
   - Documents workflow and decision points

5. Updates internal/skills/installer.go:
   - Adds to skillDirs list
   - Adds to PrintInstallSuccess

6. Builds and installs:
   - make build install
   - vibe skills

7. Verifies installation:
   - ls ~/.claude/skills/vibe-pr-ready-check
   - Shows skill now available in Claude Code

8. Reports completion with usage example
```

## Testing the Skill

After installation, test in Claude Code:

```
User: "<skill trigger phrase>"

Expected: Claude should invoke the skill and follow the documented workflow
```

Verify:

- Skill triggers on correct phrases
- Follows documented steps
- Produces expected output
- Handles errors gracefully
