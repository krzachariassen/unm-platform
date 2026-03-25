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
│   ├── code-to-dsl/         # Codebase-to-UNM model generator
│   └── documentation-writer/# Docs, README, examples, tutorials
├── commands/                # Orchestrator + direct commands
│   ├── run.md               # /run <task> — AUTO-ROUTING ORCHESTRATOR
│   ├── backend.md           # /backend <task>
│   ├── frontend.md          # /frontend <task>
│   ├── fullstack.md         # /fullstack <task>
│   ├── review-ui.md         # /review-ui <scope>
│   ├── review-code.md       # /review-code <scope>
│   ├── validate.md          # /validate
│   └── docs.md              # /docs <task>
└── rules/                   # Engineering rules (auto-loaded)
    ├── clean-architecture.md
    ├── tdd.md
    ├── git-flow.md
    ├── go-conventions.md
    ├── react-conventions.md
    └── agent-teams.md
```

### How It Works

Each **agent** has:
- `AGENT.md` -- role, process, constraints
- `MEMORY.md` -- operational memory (learnings from past sessions)
- Specialized files (anti-patterns, checklists, templates)

**Rules** in `.claude/rules/` are automatically loaded and apply to all agents.

### Orchestrator (recommended)

Use `/run` for any task. The orchestrator auto-classifies intent, routes to
the right agent(s), decomposes multi-layer tasks, and spawns agent teams
for parallel work when appropriate.

```bash
# Single-agent tasks (orchestrator routes automatically)
/run Add a new analyzer for team interaction diversity
/run Fix the capability badge colors in CapabilityView
/run Update README.md with the current project structure

# Multi-agent tasks (orchestrator decomposes and coordinates)
/run Add a priority field to needs — update the domain entity, API, and UI
/run Add CSV export endpoint and a Download button in the frontend

# Reviews (orchestrator routes to reviewer agents)
/run Review the changeset system for architecture violations
/run Do a UX review of the Capability View page
```

The orchestrator determines whether to use a single agent, sequential
agents, or parallel agent teams based on the task structure.

### Direct Agent Commands

For precise control, invoke agents directly:

```bash
/backend <task>       # Go backend only
/frontend <task>      # React frontend only
/fullstack <task>     # Tightly coupled cross-stack
/docs <task>          # Documentation and examples
/review-code <scope>  # Code quality review
/review-ui <scope>    # UX review
/validate             # Full build + test validation
```

## Key Principles

- **Git Flow**: Never commit to main. All work on feature branches. See `.claude/rules/git-flow.md`.
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
- Do not commit directly to main -- always use feature branches
- Do not let two teammates edit the same file
