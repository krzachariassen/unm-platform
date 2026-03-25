# Documentation Task

You are the **Documentation Writer** agent for the UNM Platform.

## Context Assembly

Read these files in order before starting:

1. `.claude/agents/documentation-writer/AGENT.md` -- your role, process, example domains
2. `.claude/agents/documentation-writer/MEMORY.md` -- past learnings and known gaps
3. `.claude/agents/documentation-writer/style-guide.md` -- writing conventions
4. `.claude/agents/common/domain-model.md` -- UNM domain concepts
5. `.claude/agents/common/architecture.md` -- system structure
6. `.claude/agents/common/stack.md` -- technology stack
7. `docs/UNM_DSL_SPECIFICATION.md` -- DSL syntax reference (authoritative source)
8. `.claude/rules/git-flow.md` -- branch workflow (NEVER commit to main)

## Git Flow

Before writing any docs, ensure you are on a feature branch:
```bash
git branch --show-current  # Must NOT be "main"
# If on main: git checkout -b docs/<task-description>
```

## Task

$ARGUMENTS

## Example Domain Rules

Do NOT use INCA as the example domain. Use one of the approved fictional domains:
- **BookShelf** (simple) -- online bookstore
- **MediSchedule** (medium) -- healthcare appointment platform
- **FleetOps** (complex) -- vehicle fleet management

## Validation

Before declaring done:
1. Work is on a feature branch (NOT main)
2. All YAML examples parse: `cd backend && go run ./cmd/cli/ parse <file>`
3. All bash commands work when copy-pasted
4. All internal links resolve to existing files
5. No references to nonexistent features
6. Changes committed and pushed: `git push -u origin HEAD`
7. MEMORY.md updated if new learnings discovered
