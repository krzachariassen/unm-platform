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
cd frontend && npm run build             # Full build (tsc -b + vite build)
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
│   ├── documentation-writer/# Docs, README, examples, tutorials
│   └── backlog-manager/     # Backlog maintenance
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

## Cursor Rules (synced)

The `.cursor/rules/` directory mirrors this agent framework for Cursor IDE:

| Rule | Activation | Syncs With |
|------|-----------|------------|
| `ai-engineer.mdc` | Always | `CLAUDE.md` + `commands/run.md` (routing, git-flow) |
| `backend.mdc` | `backend/**` | Backend engineer agent + rules |
| `frontend.mdc` | `frontend/**` | Frontend engineer agent + rules |
| `backlog.mdc` | `**/BACKLOG*.md` | Backlog manager agent |

When updating agent rules in `.claude/`, keep `.cursor/rules/` in sync.

## Key Principles

- **Git Flow**: Never commit to main. All work on feature branches. See `.claude/rules/git-flow.md`.
- **TDD**: Red -> Green -> Refactor. No code without a failing test first.
- **Clean Architecture**: Domain is pure Go with zero deps. Dependencies point inward.
- **Agent Teams**: Parallelize independent work. See `.claude/rules/agent-teams.md`.
- **Backlog**: Single source in `docs/BACKLOG.md`. **This file is owned by the backlog-manager agent. No other agent or orchestrator may edit it directly.** All backlog edits — adding items, restructuring phases, updating checkboxes, inlining details — MUST go through the backlog-manager agent. Three states: `[ ]` not started, `[~]` in progress, `[x]` done. Orchestrator requests the backlog-manager to mark items `[~]` when work begins and `[x]` (with date) when merged. Never bulk-update at the end — update after each item. Never mark `[x]` items that were skipped or deferred. This is non-negotiable.
- **Validation Pipeline**: Run `/validate` after all code changes. Mandatory, not optional.
- **Security**: Read `.claude/agents/common/security.md`. No secrets in code. No PII in logs.

## Memory Curation Policy

Each agent's `MEMORY.md` is subject to these guardrails:

- **30-entry hard cap** per agent. When reached, oldest non-promoted entries are removed.
- **Reusable knowledge only**: entries must be about platform behavior, not current task progress.
- **Structured format**: each entry needs date, context/service area, and a clear learning.
- **Curation** (see `.claude/agents/AGENT_OWNERSHIP.md` §2):
  - **Promote**: entry was useful in a subsequent task → move to `anti-patterns.md`
  - **Keep**: potentially useful, not yet validated
  - **Prune**: older than 2 months with no reuse

## Reference Files

- `docs/BACKLOG.md` -- Work items, phased roadmap, and Recently Completed
- `docs/ARCHITECTURE_EVOLUTION.md` -- Backend: data hierarchy, tenancy, auth, API routes
- `docs/FRONTEND_EVOLUTION.md` -- Frontend: navigation layers, tab structure, platform chrome
- `docs/PRODUCT_ROADMAP.md` -- User-facing milestones and capabilities
- `.claude/agents/AGENT_OWNERSHIP.md` -- Agent ownership and memory curation
- `docs/UNM_DSL_SPECIFICATION.md` -- DSL syntax and meta-model
- `examples/nexus.unm.yaml` -- Reference model

## What NOT to Do

- **Do not implement code directly** — for any coding task, invoke the appropriate specialist agent via the Skill tool: `/backend`, `/frontend`, `/fullstack`, `/docs`. Use `/run` if you are unsure which agent to route to. You (Claude Code) are the orchestrator — your job is to route and coordinate, not to write source files, tests, or build scripts yourself.
- **Do not edit `docs/BACKLOG.md` directly** — all backlog changes (adding items, restructuring phases, marking progress, adding detail) MUST go through the backlog-manager agent. If you find yourself opening `BACKLOG.md` for a write operation, stop and invoke the backlog-manager instead. This applies to orchestrators, specialist agents, and Cursor sessions alike.
- Do not add dependencies to `internal/domain/` (must be pure Go)
- Do not skip writing tests first
- Do not mock OpenAI in tests (use real API or skip)
- Do not hardcode view logic into the model
- Do not mix abstraction levels in view projections
- Do not commit directly to main -- always use feature branches
- Do not let two teammates edit the same file
