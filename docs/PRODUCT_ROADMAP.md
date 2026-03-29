# UNM Platform — Product Roadmap

_Phase status and goals. For work items and checklists, see `docs/BACKLOG.md`._

## Completed Phases

| Phase | Goal | Status |
|-------|------|--------|
| **1** — Domain Model & YAML Parser | Core UNM domain model in Go, YAML parser, validation engine, CLI | Done |
| **1.9** — UNM Framework Compliance | Drop Scenario, add outcome, refine visibility, one-direction relationships | Done |
| **2** — Model Querying & Analysis | Query engine, fragmentation/cognitive-load/dependency/gap analyzers | Done |
| **2.5** — Advanced Analysis | Bottleneck, coupling, complexity analyzers | Done |
| **3** — REST API Server | HTTP server with model lifecycle, analysis, and view endpoints | Done |
| **4** — Interactive Web Frontend | React app with 8 views: UNM Map, Need, Capability, Ownership, Realization, Team Topology, Cognitive Load, Dashboard | Done |
| **4.5** — View API Enrichment | Move frontend graph traversal to backend presenters. Frontend displays, backend computes. | Done |
| **4.6** — Interactive Cognitive Load & Team Topology | Stacked-bar cognitive load dashboard, interactive team topology graph | Done |
| **4.7** — Deep Capability & Ownership Views | Dependency chains, team badges, deep-linked capability pages | Done |
| **4.8** — Ownership View UX Redesign | Lane-based team ownership, service table, cross-team detection | Done |
| **4.9** — External Dependencies in Views | External deps in ownership and realization views | Done |
| **4.10** — Value Chain Risk & Signals View | Value chain traversal, signals API, signals view, stream coherence | Done |
| **5** — Custom DSL Parser | Hand-rolled recursive descent parser, AST, transformer to UNMModel | Done |
| **6** — AI-Powered Interactive Platform | Changeset engine, AI advisor, what-if, prompt library, chat UI | Done |
| **6.5** — Platform Configuration System | koanf-based layered config, typed structs, zero hardcoded values | Done |
| **6.9** — AI Per-Page Insights | Per-view AI-generated insights with prompt templates | Done |

## In Progress / Remaining

| Phase | Goal | Status |
|-------|------|--------|
| **6.10** — External Deps + Quality Hardening | External dep signals, critical bug fixes, dead code cleanup, test gaps | Partial |
| **6.12** — Architecture Refactoring | Use case extraction, DIP, registry pattern, HandlerDeps, dead code, type safety | Not started |
| **7** — Transformation & Transition Planning | State snapshots, delta analysis, transition steps, impact analysis | Not started |
| **8** — Platform Maturity & Ecosystem | Database persistence, CI/CD, export formats, multi-model, RBAC | Not started |

## Phase Dependency Graph

```
Phases 1–6.9: DONE
  ↓
Phase 6.10 (Quality Hardening)     ← in progress
  ↓
Phase 6.12 (Architecture Refactor) ← next
  ↓
Phase 7 (Transitions)              ← benefits from changeset engine
  ↓
Phase 8 (Platform Maturity)        ← requires all above
```
