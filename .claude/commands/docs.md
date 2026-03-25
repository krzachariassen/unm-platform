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

## Task

$ARGUMENTS

## Example Domain Rules

Do NOT use INCA as the example domain. Use one of the approved fictional domains:
- **BookShelf** (simple) -- online bookstore
- **MediSchedule** (medium) -- healthcare appointment platform
- **FleetOps** (complex) -- vehicle fleet management

## Validation

Before declaring done:
1. All YAML examples parse: `cd backend && go run ./cmd/cli/ parse <file>`
2. All bash commands work when copy-pasted
3. All internal links resolve to existing files
4. No references to nonexistent features
5. MEMORY.md updated if new learnings discovered
