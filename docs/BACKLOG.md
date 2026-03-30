# UNM Platform — Backlog

_Single source of truth for all work items.
Completed phases: `docs/PRODUCT_ROADMAP.md`.
Implementation patterns: `.claude/agents/` and `.claude/rules/`._

_Last updated: 2026-03-30_
_Priority: Phase 9 cleanup + data_asset simplification._

---

## Recently Completed

- [x] **refactor(data-asset)** — Simplify DataAsset: remove ProducedBy/ConsumedBy/Access; UsedBy becomes flat `[]string`; type is free-form. Updated across all parsers, serializer, analyzer, tests, and nexus.unm.yaml (2026-03-30)
- [x] **Phase 9.11 (9.11.1–9.11.8)** — Multi-Actor Needs: `ActorNames []string`, flex YAML/DSL parsing, multi-actor view grouping, signals/value-chain updates, frontend SignalsView + api.ts (2026-03-30)
- [x] **BUG-1** — Fix `.unm` DSL file upload via UI — detect `.unm` extension, send `?format=dsl` (2026-03-30)
- [x] **Phase 9 clean-up** — Merge all Phase 9 PRs to main, replace INCA examples with Nexus (anonymized), enforce backlog update in git-flow rules (2026-03-30)
- [x] **PR #24 anonymize** — Add nexus.unm.yaml (generic marketplace model), gitignore INCA files, update all test fixtures (2026-03-30)

- [x] **Phase 9 Wave 1** — YAML parser + DSL parser Phase 9: flat caps, visibility inheritance, service.realizes, service.externalDeps, team.interacts, reference validation, --strict CLI, API warnings, frontend amber UI (PRs #27, #28, #29, #30) (2026-03-30)
- [x] **Phase 9 Wave 2** — v2 YAML serializer: flat caps with parent, realizes on services, externalDeps on services, interacts on teams (PR #31) (2026-03-30)
- [x] **Phase 9 Wave 3** — Example models v2: inca.unm.yaml + inca.unm.v2.yaml converted to v2 format, minimal.unm.yaml created, inca.unm DSL copy created (PR #32) (2026-03-30)
- [x] **Phase 6.10 (items 6.10.1–6.10.16)** — External deps in all views, quality hardening: insights HTTP 200, impact config injection, localStorage persistence, typed api.ts (2026-03-29)
- [x] **Phase 6.12** — Architecture refactoring: extract use cases, HandlerDeps, view registry, dead code removal, frontend dedup + typed API (2026-03-29)
- [x] **Phase 5 DSL** — Feature parity with YAML parser: outcome, size, via, typed usedBy structs, colon shorthand, external/data aliases (2026-03-29)
- [x] Sync .cursor/rules/ with .claude/ agent framework (2026-03-17)
- [x] Fix service load dimension — backend per-person ratio (2026-03-17)

---

## Phase 6.10: External Dependencies in Views + Quality Hardening

### Part A: External Dependencies in Views

External dependencies (e.g., Cadence, Kafka) are invisible single points of
failure. The bottleneck analyzer ignores external dependency fan-in entirely.

- [x] **6.10.1** — Extend bottleneck analyzer: external dependency fan-in detection.
      Count services per external dep, flag critical (>=5) and warning (>=3).
      _File: `analyzer/bottleneck.go`_ (#backend)
- [x] **6.10.2** — Add external dependency signals to signals view API.
      New `critical_external_deps` field in organization layer.
      _File: `handler/signals.go`_ (#backend)
- [x] **6.10.3** — Add external deps to Capability View backend. Aggregate
      external deps across each capability's services.
      _File: `handler/view_enriched.go`_ (#backend)
- [x] **6.10.4** — Add external dependency nodes to UNM Map backend response.
      _File: `handler/view_enriched.go`_ (#backend)
- [x] **6.10.5** — Render external deps in SignalsView — "External Dependency
      Concentration" section with color-coded badges.
      _File: `SignalsView.tsx`_ (#frontend)
- [x] **6.10.6** — Render external deps in CapabilityView detail panel.
      _File: `CapabilityView.tsx`_ (#frontend)
- [x] **6.10.7** — Render external dependency nodes in UNMMapView.
      _File: `UNMMapView.tsx`_ (#frontend)

### Part B: Quality Hardening — CRITICAL

- [x] **6.10.8** — Model state persistence across page refresh. Persist
      modelId + parseResult to localStorage, restore on mount, verify with
      backend. _File: `model-context.tsx`_ (#frontend)
- [x] **6.10.9** — Insights endpoint returns HTTP 200 on internal errors.
      Distinguish "no findings" from "AI failed" in the response.
      _Files: `insights.go`, `usePageInsights.ts`_ (#fullstack)
- [x] **6.10.10** — Impact analyzer uses hardcoded default config instead of
      server config for cognitive load thresholds. Inject AnalysisConfig.
      _Files: `impact.go`, `main.go`_ (#backend)

### Part B: Quality Hardening — MAJOR

- [x] **6.10.11** — AI Advisor page not in sidebar. Add nav item.
      _File: `Sidebar.tsx`_ (#frontend)
- [x] **6.10.12** — Delete ~300 lines of dead legacy view builder functions.
      _File: `handler/view.go`_ (#backend)
- [x] **6.10.13** — Dashboard silently hides signals on API failure. Show
      fallback message. _File: `DashboardPage.tsx`_ (#frontend)
- [x] **6.10.14** — Frontend api.ts type definitions incomplete. Views use
      `as unknown as` casts. Add typed view fetch helpers.
      _File: `api.ts`, all view pages_ (#frontend)
- [x] **6.10.15** — AI client ignores config-resolved API key, reads env var
      independently. Pass resolved key via constructor.
      _Files: `openai_client.go`, `main.go`_ (#backend)
- [x] **6.10.16** — No panic recovery middleware. Add recovery that catches
      panics, logs stack trace, returns 500.
      _File: `middleware.go`_ (#backend)

### Part B: Quality Hardening — MINOR

- [x] **6.10.17** — Missing handler tests for health, signals, insights,
      middleware, debug endpoints. (#backend)
- [x] **6.10.18** — No frontend tests. Add vitest + smoke tests for
      model-context, api.ts, and 3 major views. (#frontend)
- [x] **6.10.19** — Empty AI question validation. Return 400 on empty/whitespace.
      _File: `ai.go`_ (#backend)
- [x] **6.10.20** — Inconsistent error handling in frontend API client.
      Standardize all functions to use extractError. _File: `api.ts`_ (#frontend)
- [x] **6.10.21** — Config handler duplicates JSON encoding instead of using
      writeJSON helper. _File: `config_handler.go`_ (#backend)
- [x] **6.10.22** — `@import` PostCSS warning. Move Google Fonts import above
      @tailwind directives. _File: `index.css`_ (#frontend)
- [x] **6.10.23** — http.Client in OpenAI client has no default timeout.
      Add 120s safety net. _File: `openai_client.go`_ (#backend)

---

## Phase 9: DSL v2 — Schema Simplification & Authoring UX

Major refactoring of the UNM YAML/DSL schema to make it dramatically simpler
for human authoring while preserving full modeling power. No backward
compatibility required — existing YAML/UNM files will be converted to v2.

**Design decisions** (from user feedback session 2026-03-29):
- Flat > nested for humans. `parent` field over `children` nesting.
- Bottom-up > top-down for authoring. Services declare what they realize/depend on.
- Inheritance > repetition. Visibility flows from parent to child.
- Short-form > long-form as default. Descriptions are optional, not expected.
- Colocate related concepts. Interactions belong on teams, not in a separate section.
- Strict validation. Catch typos in references before they become blank UI.

### 9.1 — Flat Capabilities with `parent` Field

Allow capabilities to reference a parent by name instead of nesting under
`children`. Both forms accepted; flat is the recommended human-authoring style.

- [x] **9.1.1** — Add `parent` field to `yamlCapability` struct and parser.
      Resolve parent references after all capabilities are parsed. Error if
      parent name doesn't exist. _File: `yaml_parser.go`_ (#backend)
- [x] **9.1.2** — Add `parent` support to DSL parser.
      _File: `dsl_parser.go`, `dsl/grammar.go`_ (#backend)
- [x] **9.1.3** — Add parser tests for flat capabilities: single parent,
      multi-level, mixed flat+nested, missing parent error, circular parent.
      _File: `yaml_parser_test.go`_ (#backend)
- [ ] **9.1.4** — Update DSL specification §2.4 with flat capability syntax.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [ ] **9.1.5** — Update YAML guide with flat capability examples.
      _File: `docs/YAML_GUIDE.md`_ (#docs)

### 9.2 — Visibility Inheritance

Children inherit parent visibility when not explicitly set. Eliminates ~20
redundant `visibility:` declarations per model.

- [x] **9.2.1** — After building capability hierarchy, propagate parent
      visibility to children that have no explicit visibility. Depth-first,
      parent-first. _File: `yaml_parser.go` (buildModel)_ (#backend)
- [x] **9.2.2** — Same for DSL parser. _File: `dsl_parser.go`_ (#backend)
- [x] **9.2.3** — Parser tests: child inherits, child overrides, multi-level
      inheritance, root without visibility warns.
      _File: `yaml_parser_test.go`_ (#backend)
- [ ] **9.2.4** — Update DSL specification §1 (Value Types) with inheritance rule.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 9.3 — Services Declare `realizes` (Reverse realizedBy)

Move the capability-to-service mapping to the service side. Services declare
`realizes: ["Cap A", "Cap B"]`. More natural for human authoring — you know
your service, you state what it does.

- [x] **9.3.1** — Add `realizes` field to `yamlService` struct. In
      `buildModel`, after parsing all services and capabilities, wire
      `realizes` entries into the corresponding capability's `RealizedBy` list.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.3.2** — Support optional role on service.realizes: short form
      `"Cap A"` (default primary) and long form
      `{target: "Cap A", role: "supporting"}`.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.3.3** — Conflict detection: if both `service.realizes` and
      `capability.realizedBy` declare the same pair, warn (not error) and
      prefer the service-side declaration.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.3.4** — Deprecation warning on `capability.realizedBy` — still
      accepted, but guide users toward `service.realizes`.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.3.5** — DSL parser: add `realizes` to service block.
      _File: `dsl_parser.go`, `dsl/grammar.go`_ (#backend)
- [x] **9.3.6** — Parser tests: service.realizes basic, with role, conflict
      with capability.realizedBy, mixed old+new format.
      _File: `yaml_parser_test.go`_ (#backend)
- [ ] **9.3.7** — Update DSL specification §2.5 (Services). Mark
      `capability.realizedBy` as deprecated-but-supported.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 9.4 — External Dependencies on Services

Services declare their own external deps. The `external_dependencies` section
only defines what external systems are (name + description). Reverses the
current `usedBy` direction.

- [x] **9.4.1** — Add `externalDeps` field (list of strings) to
      `yamlService`. In `buildModel`, wire these into the corresponding
      `ExternalDependency.UsedBy` entries.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.4.2** — Keep `external_dependencies[].usedBy` for backward compat
      but emit deprecation warning. If both declare the same pair, merge.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.4.3** — DSL parser: add `externalDeps` to service block.
      _File: `dsl_parser.go`_ (#backend)
- [x] **9.4.4** — Parser tests: service.externalDeps basic, merge with
      legacy usedBy, external dep defined but no service references it.
      _File: `yaml_parser_test.go`_ (#backend)
- [ ] **9.4.5** — Update DSL specification §2.5 and §2.10.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 9.5 — Interactions on Teams (Inline)

Teams declare interactions inline. Remove the separate `interactions:` section
(keep for backward compat with deprecation warning).

- [x] **9.5.1** — Add `interacts` field to `yamlTeam`: list of
      `{with, mode, via, description}`. In `buildModel`, convert these to
      `Interaction` entities as if declared in the `interactions:` section.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.5.2** — Deprecation warning on top-level `interactions:` section.
      _File: `yaml_parser.go`_ (#backend)
- [x] **9.5.3** — DSL parser: add `interacts` to team block.
      _File: `dsl_parser.go`_ (#backend)
- [x] **9.5.4** — Parser tests: inline interactions, merge with legacy
      section, duplicate detection.
      _File: `yaml_parser_test.go`_ (#backend)
- [ ] **9.5.5** — Update DSL specification §2.6 and §2.7.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 9.6 — Remove Capability `ownedBy`

Capability-level `ownedBy` is silently dropped by the parser today. Make it
explicit: emit a deprecation warning.

- [ ] **9.6.1** — Add `ownedBy` to `yamlCapability` struct (currently not
      present). Emit deprecation warning when set. Do not wire to model.
      _File: `yaml_parser.go`_ (#backend)
- [ ] **9.6.2** — Remove `ownedBy` from capability examples in INCA/Nexus.
      _Files: `examples/*.unm.yaml`_ (#docs)
- [ ] **9.6.3** — Update DSL specification §2.4 — remove ownedBy from
      capability syntax.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 9.7 — Strict Reference Validation

Catch typos in entity name references before they cause blank UI.

- [x] **9.7.1** — Post-parse validation pass: check all `supportedBy`,
      `realizedBy`/`realizes`, `dependsOn`, `owns`, `externalDeps`, `via`
      references resolve to declared entities. Collect unresolved as warnings
      (default) or errors (strict mode).
      _File: `yaml_parser.go` or new `validator/references.go`_ (#backend)
- [x] **9.7.2** — CLI `validate` command: add `--strict` flag that promotes
      reference warnings to errors.
      _File: `cmd/cli/`_ (#backend)
- [x] **9.7.3** — API: include unresolved reference warnings in parse
      response so the UI can surface them.
      _File: `handler/parse.go`_ (#backend)
- [x] **9.7.4** — Frontend: show unresolved reference warnings on upload.
      _File: `UploadPage.tsx`_ (#frontend)
- [x] **9.7.5** — Tests for unresolved references across all relationship types.
      _File: `yaml_parser_test.go` or `validator/references_test.go`_ (#backend)

### 9.8 — Data Assets: Optional + Compact Syntax

Data assets remain supported but are explicitly optional. Add compact syntax
for simple cases.

- [ ] **9.8.1** — Document data_assets as optional in DSL spec. Explain when
      they add value (implicit coupling detection) and when to skip them.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [ ] **9.8.2** — Support compact `usedBy` syntax:
      `usedBy: {"svc-a": "read-write", "svc-b": "read"}`.
      _File: `yaml_parser.go` (yamlDataAssetRef UnmarshalYAML)_ (#backend)
- [ ] **9.8.3** — Parser tests for compact syntax.
      _File: `yaml_parser_test.go`_ (#backend)

### 9.9 — Rewrite Example Models (v2 Format)

Rewrite both example files using the new v2 conventions: flat capabilities,
service.realizes, short-form relationships, inline interactions, visibility
inheritance, service-side externalDeps.

- [x] **9.9.1** — Rewrite `examples/inca.unm.yaml` in v2 format. Validate
      with parser. Expected: ~600 lines (down from 1,168).
      _File: `examples/inca.unm.yaml`_ (#docs)
- [ ] **9.9.2** — Rewrite `examples/nexus.unm.yaml` in v2 format.
      _File: `examples/nexus.unm.yaml`_ (#docs)
- [x] **9.9.3** — Create a `examples/minimal.unm.yaml` — smallest possible
      valid model (~30 lines) for quick onboarding.
      _File: `examples/minimal.unm.yaml`_ (#docs)
- [ ] **9.9.4** — Update YAML guide with v2 examples throughout.
      _File: `docs/YAML_GUIDE.md`_ (#docs)
- [ ] **9.9.5** — Update DSL specification with v2 as primary syntax.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **9.9.6** — Create `examples/inca.unm` — DSL equivalent of inca.unm.yaml
      for format comparison (nested caps, realizedBy on caps, standalone interactions).
      _File: `examples/inca.unm`_ (#docs)

### 9.10 — Serializer / Export Alignment

The YAML serializer (export) must produce v2 format when exporting models.

- [x] **9.10.1** — Update YAML serializer to emit `realizes` on services
      instead of `realizedBy` on capabilities.
      _File: `serializer/yaml_serializer.go`_ (#backend)
- [x] **9.10.2** — Emit flat capabilities with `parent` instead of nested
      `children` in serializer output.
      _File: `serializer/yaml_serializer.go`_ (#backend)
- [x] **9.10.3** — Emit `externalDeps` on services, `interacts` on teams.
      _File: `serializer/yaml_serializer.go`_ (#backend)
- [x] **9.10.4** — Serializer tests: round-trip v1 input → parse → serialize
      → parse → compare models.
      _File: `serializer/yaml_serializer_test.go`_ (#backend)

### 9.11 — Multi-Actor Needs

`actor` field on needs becomes a list. A single need can be shared by multiple
actors instead of duplicating the entire need block.

```yaml
needs:
  - name: "View accurate, up-to-date catalogs"
    actor: ["Eater / Consumer", "Downstream Platform Team"]
    outcome: "..."
    supportedBy: [...]
```

- [x] **9.11.1** — Change `Need.ActorName string` to `ActorNames []string`
- [x] **9.11.2** — Parser: `actor` field accepts both a string and a list of strings (flex unmarshaling).
- [x] **9.11.3** — DSL parser: support `actor "A", "B"` syntax.
- [x] **9.11.4** — Need View API: needs with multiple actors appear in each actor's group.
- [x] **9.11.5** — Frontend NeedView: handle multi-actor needs (server-side grouping, no client changes needed).
- [x] **9.11.6** — Signals View: `actor_names[]` display in SignalsView.tsx and api.ts.
- [x] **9.11.7** — Signals/Cognitive Load views: `ActorNames []string` in signals_service, value_chain, analysis_runner.
- [x] **9.11.8** — Parser + domain tests for multi-actor needs.
- [ ] **9.11.9** — Convert existing example files: deduplicate needs shared
      across actors. _Files: `examples/*.unm.yaml`_ (#docs)

### Execution Order

**Wave 1** (parser, no API/UI impact): 9.1, 9.2, 9.6, 9.7, 9.11
**Wave 2** (relationship direction changes): 9.3, 9.4, 9.5
**Wave 3** (examples + docs): 9.8, 9.9, 9.10
Existing YAML/UNM files will be converted to v2 format — no backward compat needed.

---

## Bugs

- [x] **BUG-1** — Cannot upload `.unm` DSL files via the UI: server returns
      `parser: invalid YAML`. Fix: detect `.unm` extension in `UploadPage.tsx`
      and pass `?format=dsl` to the parse API. _Files: `api.ts`, `UploadPage.tsx`_ (#frontend)

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
