# UNM Platform

**Executable architecture modeling for User Needs Mapping, Team Topologies, and AI-assisted discovery.**

UNM Platform turns User Needs Mapping from a workshop artifact into a versionable, queryable, AI-assisted engineering tool. It codifies the relationship between architecture, teams, and flow — letting you model, analyze, and evolve your organization intentionally.

## What It Does

- **Model** your architecture as code: actors, needs, capabilities, services, teams, and their relationships
- **Analyze** fragmentation, cognitive load, dependency chains, and organizational anti-patterns
- **Visualize** multiple interactive views: needs, capabilities, realization, ownership, team topologies, UNM map
- **Edit** models through a smart changeset system with validation and impact preview
- **Generate** candidate models from real codebases using AI agents
- **Advise** with AI-powered architectural insights and recommendations

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (Clean Architecture, net/http stdlib) |
| Frontend | React 19 + TypeScript + Vite 6 + Tailwind CSS v4 + shadcn/ui |
| Graph visualization | React Flow, D3.js |
| Icons | Lucide React |
| Routing | React Router |
| Config | koanf v2 (layered: defaults → YAML → env vars) |
| API | REST (JSON) |
| AI | OpenAI GPT (optional, degrades gracefully) |

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Actor** | Person or system with needs (Merchant, Eater, Operator) |
| **Need** | What an actor is trying to achieve — an outcome, not a feature |
| **Capability** | What the system must be able to do, positioned by visibility (user-facing → infrastructure) |
| **Service** | Concrete implementation that realizes capabilities |
| **Team** | Organizational unit that owns services (stream-aligned, platform, enabling, complicated-subsystem) |
| **Interaction** | How teams work together (collaboration, x-as-a-service, facilitating) |
| **Signal** | Architectural finding with severity and evidence (bottleneck, fragmentation, coupling, gap) |
| **Data Asset** | Storage or messaging infrastructure shared by services |

## Project Structure

```
backend/                    # Go backend (Clean Architecture)
├── cmd/
│   ├── server/             # HTTP API server (main entrypoint)
│   ├── cli/                # CLI tool for parsing and analysis
│   └── runquestions/       # AI test question runner
├── internal/
│   ├── domain/
│   │   ├── entity/         # Core entities (Actor, Need, Capability, Service, Team, etc.)
│   │   ├── valueobject/    # Value objects (TeamType, InteractionMode, Confidence)
│   │   └── service/        # Domain services (validation, anti-patterns, changeset applier)
│   ├── usecase/            # Application use cases
│   ├── adapter/
│   │   ├── handler/        # HTTP handlers (REST API)
│   │   ├── presenter/      # View model transformers for each view type
│   │   └── repository/     # In-memory stores
│   └── infrastructure/
│       ├── parser/         # YAML parser + custom DSL parser (PEG)
│       ├── analyzer/       # Analysis engines
│       ├── serializer/     # YAML export
│       ├── config/         # Configuration loading (koanf)
│       └── ai/             # OpenAI integration, prompt templates
├── pkg/
│   └── unmmodel/           # Shared model types (used by CLI and server)
└── testdata/               # Test fixture files (.unm.yaml)

config/                     # Environment config files
├── base.yaml               # Shared defaults
├── local.yaml              # Local dev overrides
├── production.yaml         # Production overrides
└── test.yaml               # Test overrides

frontend/                   # React frontend
├── src/
│   ├── pages/              # Route pages (DashboardPage, UploadPage, ViewPage, etc.)
│   │   └── views/          # View components (UNMMapView, NeedView, CapabilityView, etc.)
│   ├── components/
│   │   ├── ui/             # Base UI components (shadcn/ui)
│   │   ├── layout/         # App shell, sidebar, navigation
│   │   ├── graph/          # React Flow graph components
│   │   ├── changeset/      # Edit panel, action forms, smart dropdowns
│   │   └── advisor/        # AI advisor panel
│   ├── hooks/              # Custom hooks (useAIEnabled, usePageInsights)
│   ├── lib/                # API client, model context, search context, config
│   └── types/              # TypeScript type definitions

docs/                       # Project documentation
├── BACKLOG.md              # Phased product backlog
├── ENGINEERING_PRINCIPLES.md
├── UNM_DSL_SPECIFICATION.md  # Complete DSL syntax and meta-model
├── CODE_TO_DSL_AGENT.md    # Codebase-to-UNM model generation process
├── CONFIGURATION.md        # Config system reference
├── ARCHITECTURE_REVIEW.md  # Architecture review notes
└── UI_BACKLOG.md           # Frontend UI backlog

examples/                   # Example UNM models
└── inca.unm.yaml           # Reference model

scripts/                    # Utility scripts (publishing, deployment)

.claude/                    # AI agent framework (see CLAUDE.md)
├── agents/                 # Specialized AI agents
├── commands/               # Orchestrator slash commands
└── rules/                  # Engineering rules
```

## Quick Start

```bash
# 1. Run the backend (from project root)
cd backend && go run ./cmd/server/

# 2. Run the frontend (separate terminal)
cd frontend && npm install && npm run dev
```

The backend starts on port **8080**, the frontend on port **5173**.

Open http://localhost:5173, go to the Upload page, and drag-drop a `.unm.yaml` file to load a model.
A reference model is available at `examples/inca.unm.yaml`.

### AI Features (optional)

AI-powered insights and recommendations require an OpenAI API key:

```bash
# Add your key to ai.env
source ai.env
cd backend && go run ./cmd/server/
```

AI features degrade gracefully — the platform is fully functional without them.

## Configuration

Layered config system: code defaults → `config/base.yaml` → environment overrides → `UNM_*` env vars.

```bash
UNM_ENV=production go run ./cmd/server/   # loads config/production.yaml
```

See [Configuration Reference](docs/CONFIGURATION.md) for all options.

## Interactive Views

| View | Route | What It Shows |
|------|-------|---------------|
| Dashboard | `/dashboard` | Stats, health summary, signals overview |
| Signals | `/signals` | Architectural findings by category and severity |
| UNM Map | `/unm-map` | Full value chain visualization with pan/zoom |
| Need View | `/need` | Actor needs, mapping coverage, risk indicators |
| Capability View | `/capability` | Capability hierarchy by visibility or domain |
| Ownership View | `/ownership` | Service ownership by team with anti-pattern detection |
| Team Topology | `/team-topology` | Team interaction graph (Team Topologies) |
| Cognitive Load | `/cognitive-load` | Per-team cognitive load assessment |
| Realization View | `/realization` | Capability-to-service traceability |
| Edit Model | `/edit` | Smart changeset editor with validation |
| What-If Explorer | `/what-if` | Scenario modeling with impact preview |

## Engineering Principles

- **TDD**: Red → Green → Refactor. No production code without a failing test.
- **Clean Architecture**: Domain is pure Go with zero external deps. Dependencies point inward.
- **SOLID**: Single responsibility, open/closed, dependency inversion.

See [Engineering Principles](docs/ENGINEERING_PRINCIPLES.md) for full details.

## AI Agent Framework

This project uses a decomposed AI agent architecture for development.
See [CLAUDE.md](CLAUDE.md) for the full agent framework documentation.

## License

Internal — Uber Technologies, Inc.
