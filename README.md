# UNM Platform

**Executable architecture modeling for User Needs Mapping, Team Topologies, and AI-assisted discovery.**

UNM Platform turns User Needs Mapping from a workshop artifact into a versionable, queryable, AI-assisted engineering tool. It codifies the relationship between architecture, teams, and flow — letting you model, analyze, and evolve your organization intentionally.

## What It Does

- **Model** your architecture as code: actors, needs, capabilities, services, teams, and their relationships
- **Analyze** fragmentation, cognitive load, dependency chains, and organizational anti-patterns
- **Visualize** multiple interactive views: needs, capabilities, realization, ownership, team topologies, transitions
- **Generate** candidate models from real codebases using AI agents
- **Plan** transitions from current state to target state with step-by-step migration paths

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (Clean Architecture) |
| Frontend | React + TypeScript + Vite + Tailwind CSS v4 + shadcn/ui |
| Graph visualization | React Flow |
| API | REST (JSON) |
| DSL parsing | YAML (Phase 1) → Custom PEG parser (Phase 5) |

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Actor** | Person or system with needs (Merchant, Eater, Operator) |
| **Need** | What an actor is trying to achieve in a specific scenario |
| **Capability** | What the system/org must be able to do to support a need |
| **Service** | Concrete implementation that realizes a capability |
| **Team** | Organizational unit that owns services/capabilities (with Team Topologies type) |
| **Interaction** | How teams work together (collaboration, x-as-a-service, facilitating) |

## Project Structure

```
backend/               # Go backend (Clean Architecture)
├── cmd/               # CLI and server entrypoints
├── internal/
│   ├── domain/        # Entities, value objects, domain services (zero deps)
│   ├── usecase/       # Application business logic
│   ├── adapter/       # HTTP handlers, presenters, repositories
│   └── infrastructure/# Parsers, analyzers, AI integration
├── pkg/               # Public types
└── testdata/          # Test fixtures

frontend/              # React frontend
├── src/
│   ├── components/    # UI components (shadcn/ui), views, graph rendering
│   ├── hooks/         # Custom React hooks
│   ├── lib/           # Utilities, API client
│   ├── pages/         # Route pages
│   └── types/         # TypeScript types matching backend API

docs/                  # Project documentation
├── BACKLOG.md         # Full phased product backlog (8 phases)
├── ENGINEERING_PRINCIPLES.md  # TDD, SOLID, Clean Architecture guidelines
└── UNM_DSL_SPECIFICATION.md   # DSL syntax, meta-model, validation rules

examples/              # Example UNM models
└── inca.unm.yaml      # INCA platform example
```

## Backlog

The product is built in 8 incremental phases. Each phase delivers a testable, demonstrable artifact.

| Phase | Focus | Deliverable |
|-------|-------|-------------|
| 1 | Domain Model & YAML Parser | Go CLI: parse and validate `.unm.yaml` files |
| 2 | Querying & Analysis | CLI: fragmentation, cognitive load, dependency, gap analysis |
| 3 | REST API Server | HTTP API exposing all model operations |
| 4 | Interactive Web Frontend | React app with multiple interactive views |
| 5 | Custom DSL Parser | Full `.unm` DSL with imports, transitions, inferred mappings |
| 6 | AI-Assisted Generation | Agent pipeline: codebase → candidate UNM model |
| 7 | Transition Planning | Current vs target state modeling with migration steps |
| 8 | Platform Maturity | Database, CI/CD, plugins, federation |

See [Full Backlog](docs/BACKLOG.md) for detailed items.

## Quick Start

```bash
# 1. Set up AI credentials
cp ai.env.example ai.env
# Edit ai.env and add your OpenAI API key

# 2. Load credentials (once per terminal session)
source ai.env

# 3. Run the backend
cd backend && go run ./cmd/server/

# 4. Run the frontend (separate terminal)
cd frontend && npm install && npm run dev
```

For production mode, set `UNM_ENV=production` before starting the server.

## Configuration

The platform uses a layered config system: code defaults, `config/base.yaml`, environment-specific overrides (`config/{env}.yaml`), and `UNM_*` environment variables (highest priority).

Set the environment with `UNM_ENV`:

```bash
export UNM_ENV=production   # loads config/production.yaml on top of base.yaml
```

See [Configuration Reference](docs/CONFIGURATION.md) for the full schema, all available keys, secret management, and how to add custom environments.

## Engineering Principles

- **TDD**: Red → Green → Refactor. No production code without a failing test.
- **Clean Architecture**: Strict layer separation with dependency inversion.
- **SOLID**: Single responsibility, open/closed, interface segregation, dependency inversion.
- **KISS**: Simplest solution that works. YAGNI.

See [Engineering Principles](docs/ENGINEERING_PRINCIPLES.md) for full details.

## License

Internal — Uber Technologies, Inc.
