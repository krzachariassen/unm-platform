# UNM Platform — Backlog

_Single source of truth for all work items.
Completed phases: `docs/PRODUCT_ROADMAP.md`.
Implementation patterns: `.claude/agents/` and `.claude/rules/`._

_Last updated: 2026-03-29_

---

## Recently Completed

- [x] **Phase 6.12** — Architecture refactoring: extract use cases, HandlerDeps, view registry, dead code removal, frontend dedup + typed API (2026-03-29)
- [x] **Phase 5 DSL** — Feature parity with YAML parser: outcome, size, via, typed usedBy structs, colon shorthand, external/data aliases (2026-03-29)
- [x] Sync .cursor/rules/ with .claude/ agent framework (2026-03-17)
- [x] Fix service load dimension — backend per-person ratio (2026-03-17)
- [x] Add team interactions and service dependsOn to inca-louie.unm.yaml (2026-03-17)

---

## Phase 6.10: External Dependencies in Views + Quality Hardening

### Part A: External Dependencies in Views

External dependencies (e.g., Cadence, Kafka) are invisible single points of
failure. The bottleneck analyzer ignores external dependency fan-in entirely.

- [ ] **6.10.1** — Extend bottleneck analyzer: external dependency fan-in detection.
      Count services per external dep, flag critical (>=5) and warning (>=3).
      _File: `analyzer/bottleneck.go`_ (#backend)
- [ ] **6.10.2** — Add external dependency signals to signals view API.
      New `critical_external_deps` field in organization layer.
      _File: `handler/signals.go`_ (#backend)
- [ ] **6.10.3** — Add external deps to Capability View backend. Aggregate
      external deps across each capability's services.
      _File: `handler/view_enriched.go`_ (#backend)
- [ ] **6.10.4** — Add external dependency nodes to UNM Map backend response.
      _File: `handler/view_enriched.go`_ (#backend)
- [ ] **6.10.5** — Render external deps in SignalsView — "External Dependency
      Concentration" section with color-coded badges.
      _File: `SignalsView.tsx`_ (#frontend)
- [ ] **6.10.6** — Render external deps in CapabilityView detail panel.
      _File: `CapabilityView.tsx`_ (#frontend)
- [ ] **6.10.7** — Render external dependency nodes in UNMMapView.
      _File: `UNMMapView.tsx`_ (#frontend)

### Part B: Quality Hardening — CRITICAL

- [ ] **6.10.8** — Model state persistence across page refresh. Persist
      modelId + parseResult to localStorage, restore on mount, verify with
      backend. _File: `model-context.tsx`_ (#frontend)
- [ ] **6.10.9** — Insights endpoint returns HTTP 200 on internal errors.
      Distinguish "no findings" from "AI failed" in the response.
      _Files: `insights.go`, `usePageInsights.ts`_ (#fullstack)
- [ ] **6.10.10** — Impact analyzer uses hardcoded default config instead of
      server config for cognitive load thresholds. Inject AnalysisConfig.
      _Files: `impact.go`, `main.go`_ (#backend)

### Part B: Quality Hardening — MAJOR

- [ ] **6.10.11** — AI Advisor page not in sidebar. Add nav item.
      _File: `Sidebar.tsx`_ (#frontend)
- [ ] **6.10.12** — Delete ~300 lines of dead legacy view builder functions.
      _File: `handler/view.go`_ (#backend)
- [ ] **6.10.13** — Dashboard silently hides signals on API failure. Show
      fallback message. _File: `DashboardPage.tsx`_ (#frontend)
- [ ] **6.10.14** — Frontend api.ts type definitions incomplete. Views use
      `as unknown as` casts. Add typed view fetch helpers.
      _File: `api.ts`, all view pages_ (#frontend)
- [ ] **6.10.15** — AI client ignores config-resolved API key, reads env var
      independently. Pass resolved key via constructor.
      _Files: `openai_client.go`, `main.go`_ (#backend)
- [ ] **6.10.16** — No panic recovery middleware. Add recovery that catches
      panics, logs stack trace, returns 500.
      _File: `middleware.go`_ (#backend)

### Part B: Quality Hardening — MINOR

- [ ] **6.10.17** — Missing handler tests for health, signals, insights,
      middleware, debug endpoints. (#backend)
- [ ] **6.10.18** — No frontend tests. Add vitest + smoke tests for
      model-context, api.ts, and 3 major views. (#frontend)
- [ ] **6.10.19** — Empty AI question validation. Return 400 on empty/whitespace.
      _File: `ai.go`_ (#backend)
- [ ] **6.10.20** — Inconsistent error handling in frontend API client.
      Standardize all functions to use extractError. _File: `api.ts`_ (#frontend)
- [ ] **6.10.21** — Config handler duplicates JSON encoding instead of using
      writeJSON helper. _File: `config_handler.go`_ (#backend)
- [ ] **6.10.22** — `@import` PostCSS warning. Move Google Fonts import above
      @tailwind directives. _File: `index.css`_ (#frontend)
- [ ] **6.10.23** — http.Client in OpenAI client has no default timeout.
      Add 120s safety net. _File: `openai_client.go`_ (#backend)

---

## Phase 7: Transformation & Transition Planning

Model current vs target organizational states with step-by-step transition plans.

- [ ] **7.1** — State snapshots: named model snapshots (current, target-q3) (#backend)
- [ ] **7.2** — Delta analysis: compare two snapshots, produce structured diff (#backend)
- [ ] **7.3** — Transition step modeling: define actions (move, merge, split) (#backend)
- [ ] **7.4** — Transition validation: verify plan transforms current to target (#backend)
- [ ] **7.5** — Transition view: side-by-side or overlay with step-through (#fullstack)
- [ ] **7.6** — Impact analysis: show downstream impacts of proposed changes (#backend)

---

## Phase 8: Platform Maturity & Ecosystem

Production-grade features for real organizational adoption.

- [ ] **8.1** — Model persistence (SQLite, expandable to PostgreSQL) (#backend)
- [ ] **8.2** — CI/CD integration (GitHub Action for .unm validation in PRs) (#infra)
- [ ] **8.3** — Export formats (JSON, Mermaid, PlantUML, Structurizr DSL) (#backend)
- [ ] **8.4** — Import from other tools (Structurizr, C4, existing docs) (#backend)
- [ ] **8.5** — Multi-model support (multiple models in one instance) (#fullstack)
- [ ] **8.6** — Model history and audit trail (#backend)
- [ ] **8.7** — Role-based access control (#fullstack)
- [ ] **8.8** — Hosted deployment with auth (#infra)
