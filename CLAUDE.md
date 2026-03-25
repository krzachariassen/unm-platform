# CLAUDE.md -- UNM Platform

## What Is This Project

UNM Platform is an executable architecture modeling system that codifies **User Needs Mapping (UNM)** combined with **Team Topologies** into a versionable, queryable, AI-assisted platform. It turns UNM from a workshop artifact into an engineering tool.

**Go backend** (Clean Architecture) + **React frontend** (Vite + Tailwind + shadcn/ui).

## Quick Commands

```bash
# Backend
cd backend && go test ./...              # Run all tests
cd backend && go build ./cmd/server/     # Build server
cd backend && go run ./cmd/server/       # Run server (port 8080)

# Frontend
cd frontend && npm install               # Install deps
cd frontend && npx tsc --noEmit          # Type check
cd frontend && npx vite build            # Production build
cd frontend && npm run dev               # Dev server (port 5173)

# AI tests (require API key)
source ai.env && cd backend && go test ./internal/infrastructure/ai/... -v -timeout 10m
```

## Agent Framework

This project uses a **decomposed AI agent architecture**. Instead of one
monolithic context file, specialized agents get focused, composable context.

### Structure

```
.claude/
├── agents/
│   ├── common/              # Shared context modules (all agents read these)
│   │   ├── architecture.md  # System structure
│   │   ├── domain-model.md  # UNM domain concepts
│   │   ├── build-test.md    # Build and test commands
│   │   ├── safety-checklist.md  # Pre-submission checks
│   │   └── stack.md         # Technology stack
│   ├── backend-engineer/    # Go backend specialist
│   ├── frontend-engineer/   # React frontend specialist
│   ├── fullstack-engineer/  # Cross-stack feature developer
│   ├── ui-reviewer/         # UX review specialist
│   ├── code-reviewer/       # Code quality reviewer
│   └── code-to-dsl/         # Codebase-to-UNM model generator
├── commands/                # Orchestrator slash commands
│   ├── backend.md           # /backend <task>
│   ├── frontend.md          # /frontend <task>
│   ├── fullstack.md         # /fullstack <task>
│   ├── review-ui.md         # /review-ui <scope>
│   ├── review-code.md       # /review-code <scope>
│   └── validate.md          # /validate
└── rules/                   # Engineering rules (auto-loaded)
    ├── clean-architecture.md
    ├── tdd.md
    ├── go-conventions.md
    ├── react-conventions.md
    └── agent-teams.md
```

### How It Works

Each **agent** has:
- `AGENT.md` -- role, process, constraints
- `MEMORY.md` -- operational memory (learnings from past sessions)
- Specialized files (anti-patterns, checklists, templates)

Each **command** is an orchestrator that assembles the right context
for a task. Commands reference agents and common modules.

**Rules** in `.claude/rules/` are automatically loaded and apply to all agents.

### Using Commands (Claude Code)

```bash
# Backend task
claude "/backend Implement a new analyzer for team interaction diversity"

# Frontend task
claude "/frontend Add a new filter chip to the Ownership view for platform teams"

# Full-stack task
claude "/fullstack Add a new /api/v1/models/{id}/export/csv endpoint and a Download CSV button"

# UI review
claude "/review-ui Review the Capability View page for UX issues"

# Code review
claude "/review-code Review the changeset system (entity/changeset.go + service/changeset_applier.go)"

# Validation
claude "/validate Run full backend + frontend validation"
```

## Key Principles

- **TDD**: Red -> Green -> Refactor. No code without a failing test first.
- **Clean Architecture**: Domain is pure Go with zero deps. Dependencies point inward.
- **Agent Teams**: Parallelize independent work. See `.claude/rules/agent-teams.md`.

## Reference Files

- `docs/BACKLOG.md` -- Phased product backlog
- `docs/ENGINEERING_PRINCIPLES.md` -- Full engineering principles
- `docs/UNM_DSL_SPECIFICATION.md` -- DSL syntax and meta-model
- `docs/CODE_TO_DSL_AGENT.md` -- Code analysis process for model generation
- `examples/inca.unm.yaml` -- Reference model

## What NOT to Do

- Do not add dependencies to `internal/domain/` (must be pure Go)
- Do not skip writing tests first
- Do not mock OpenAI in tests (use real API or skip)
- Do not hardcode view logic into the model
- Do not mix abstraction levels in view projections
- Do not let two teammates edit the same file
