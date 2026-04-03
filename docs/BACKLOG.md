# UNM Platform — Backlog

_Single source of truth for all work items.
Completed phases: `docs/PRODUCT_ROADMAP.md`.
Implementation patterns: `.claude/agents/` and `.claude/rules/`._

_Last updated: 2026-04-03 (Phase 10–13 reconciliation: all completed items verified, deferred items left open)_
_Priority: Phase 10 (model freeze) → Phase 11 (docs) → Phase 12 (tests/CI)
→ Phase 13 (code quality) → Phase 14 (persistence) → Phase 15 (auth/tenancy)
→ Phase 16 (collaboration) → Phase 17 (hardening) → Phase 18 (ecosystem)._

_Context: Two independent external reviews identified migration completion,
meta-model stability, analyzer trust, test infrastructure, persistence, and
auth as the critical path to product readiness. No users exist — backward
compatibility is not required. All legacy patterns can be removed outright._

---

## Recently Completed

- [x] **refactor(phase-13): 13.1, 13.3, 13.4, 13.5 (backend)** — Split `view_enriched.go` (1281 lines) into 7 focused files: `view_helpers.go`, `view_need.go`, `view_capability.go`, `view_team.go`, `view_ownership.go`, `view_realization.go`, `view_map.go`; deleted `view_enriched.go`; removed backward-compat `corsMiddleware` with `"*"`; added `Explanation` field to `Signal` and `SuggestedSignal` with human-readable threshold explanations in all 5 analyzer rules; defined `SourceType` enum (`model_fact`, `analyzer_finding`, `ai_interpretation`) in `domain/valueobject`; added `Source SourceType` to `Signal` entity and `SourceTag` to `SuggestedSignal` with `SourceAnalyzerFinding` tagging; all 14 packages pass. Deferred: 13.2 (handler sub-handler decomposition), 13.4.2 (SignalsView explanation surface), 13.5.3 (frontend trust indicators). (2026-04-03)
- [x] **test(phase-12): 12.6.1-12.6.3** — Frontend tests in CI: added `npm run test` step to `.github/workflows/ci.yml`; smoke tests for NeedView, TeamTopologyView, OwnershipView, RealizationView, CognitiveLoadView, DashboardPage, UploadPage (69 tests total across 16 test files, all passing). Deferred: 12.6.4 (WhatIfPage and AdvisorPage smoke tests), 12.1.3 (analyzer golden fixtures). (2026-04-02)
- [x] **test(phase-12): 12.1.1-12.1.2, 12.2, 12.3, 12.5** — Test infrastructure: golden model fixtures for need/capability/ownership/team-topology/signals/cognitive-load views (nexus.unm.yaml source, UPDATE_GOLDEN=1 workflow); commit endpoint HTTP tests + UpdatesStoredModel; ChangesetStore CRUD tests; DSL→YAML cross-format round-trip; validation severity levels (error/warning/info) + orphan entity diagnostics (InfoOrphanActor, InfoOrphanTeam); DSL transformer warning parity for unresolved realizes and team interaction targets. (2026-04-02)
- [x] **docs(phase-11): 11.1-11.5** — Documentation alignment: rewrote UNM_DSL_SPECIFICATION.md to v2-only (service.realizes, team.interacts, external deps definition-only, removed realizedBy/usedBy/interactions/scenarios); rewrote YAML_GUIDE.md and DSL_GUIDE.md to v2-only patterns; created META_MODEL_REFERENCE.md (authored vs derived fields, v2 removal table); fixed inca.unm.yaml → nexus.unm.yaml references in CLAUDE.md and CONFIGURATION.md. (2026-04-02)
- [x] **refactor(phase-10): 10.1-10.8 (all items)** — Model Freeze: removed capability.realizedBy, capability.ownedBy, top-level interactions:, external_dependencies.usedBy, scenarios, signals, pain_points, inferred, need.scenario, service.type/supports/dataAssets/externalDependsOn from YAML parser; removed realizedBy keyword from DSL grammar and AST; updated all tests and testdata; rewrote bookshelf.unm to v2 (realizes on services); verified nexus.unm.yaml v2-clean; updated extract-actions.tmpl and recommendations.tmpl display strings; updated README DSL example to v2 pattern; fixed dsl_serializer_test.go to reference nexus.unm.yaml. (2026-04-02)
- [x] **feat(export-ui)** — Export section on UploadPage: .unm.yaml and .unm download buttons (2026-03-31)
- [x] **refactor(ui-unification)** — Complete UI unification: TanStack Query, shared components, all 12 views rewritten, UNMMapView with React Flow, 36 tests passing (2026-03-31)
- [x] **refactor(clean-arch)** — Fix Clean Architecture violations: analyzer interfaces in usecase, Handler singletons, changeset validation guards (2026-03-31)
- [x] **feat(ux): batch edit mode** — ChangesetContext, PendingChangesBar, map editing, Services panel, validation warnings (2026-03-31)
- [x] **refactor(data-asset)** — Simplify DataAsset: flat UsedBy []string, free-form type (2026-03-30)
- [x] **Phase 9 complete** — DSL v2 schema: flat caps, visibility inheritance, service.realizes, service.externalDeps, team.interacts, reference validation, multi-actor needs, serializer v2, example models (2026-03-30)

---

## Phase 10: Model Freeze — Remove All Legacy Patterns

No backward compatibility. No deprecation warnings. Remove the old code entirely.
After this phase, only v2 patterns exist in parsers, serializers, tests, and examples.

### 10.1 — Remove `capability.realizedBy` from YAML Parser

- [x] **10.1.1** — Remove `realizedBy` field from `yamlCapability` struct.
      Remove all deprecation-warning logic and `processServiceRealizes`
      conflict-merge code. `service.realizes` becomes the only path.
      _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.1.2** — Remove `realizedBy` from `yamlCapability` in all YAML
      parser test fixtures. Update assertions.
      _Files: `yaml_parser_test.go`, `yaml_parser_error_test.go`,
      `yaml_parser_phase9_test.go`_ (#backend)
- [x] **10.1.3** — Remove `realizedBy` from handler test YAML snippets.
      _Files: `handler/model_test.go`, `handler/view_test.go`,
      `handler/changeset_test.go`_ (#backend)
- [x] **10.1.4** — Remove `realizedBy` from `testdata/relationships.unm.yaml`
      and any other test fixtures using old format.
      _File: `testdata/`_ (#backend)

### 10.2 — Remove `capability.realizedBy` from DSL Parser

- [x] **10.2.1** — Remove `realizedBy` keyword parsing from DSL grammar for
      capability blocks. `service.realizes` is the only authoring path.
      _File: `parser/dsl/grammar.go`_ (#backend)
- [x] **10.2.2** — Remove `RealizedByNode` (or equivalent) from DSL AST if
      it becomes unused after removing capability-side parsing.
      _File: `parser/dsl/ast.go`_ (#backend)
- [x] **10.2.3** — Update DSL grammar tests: remove all `realizedBy` on
      capability fixtures, ensure `realizes` on services is well-covered.
      _Files: `dsl/grammar_test.go`, `dsl/transformer_test.go`_ (#backend)

### 10.3 — Remove Top-Level `interactions:` Section

- [x] **10.3.1** — Remove `interactions` field from `yamlModel` struct.
      Remove `addInteractions` merge logic and deprecation warning.
      `team.interacts` is the only path.
      _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.3.2** — Decision: keep standalone `interaction "A" -> "B"` blocks
      in DSL (arrow syntax is ergonomic). Only YAML top-level section removed.
      _File: `parser/dsl/grammar.go`_ (#backend)
- [x] **10.3.3** — Update all test fixtures using top-level interactions.
      _Files: `testdata/relationships.unm.yaml`, parser tests_ (#backend)

### 10.4 — Remove `external_dependencies[].usedBy`

- [x] **10.4.1** — Remove `usedBy` field from `yamlExternalDependency` struct.
      Remove duplicate-merge logic. `service.externalDeps` is the only path.
      _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.4.2** — Update parser tests for external deps.
      _File: `yaml_parser_test.go`, `yaml_parser_phase9_test.go`_ (#backend)

### 10.5 — Remove All Silently Ignored Fields

- [x] **10.5.1** — Remove `scenarios` field from `yamlModel` struct.
      _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.5.2** — Remove `scenario` field from `yamlNeed` struct.
      _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.5.3** — Remove `signals`, `pain_points`, `inferred` fields
      from `yamlModel` struct.
      _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.5.4** — Remove `capability.ownedBy` field and its deprecation
      warning. _File: `parser/yaml_parser.go`_ (#backend)
- [x] **10.5.5** — Remove `service.type`, `service.supports`,
      `service.dataAssets`, `service.externalDependsOn` fields. Remove
      the `TODO` comment about future deprecation warnings.
      _File: `parser/yaml_parser.go`_ (#backend)

### 10.6 — Clean Up Examples

- [x] **10.6.1** — Rewrite `examples/bookshelf.unm` to v2 format (currently
      uses `realizedBy` on capabilities only).
      _File: `examples/bookshelf.unm`_ (#docs)
- [x] **10.6.2** — Rewrite `examples/nexus.unm.yaml` to v2-only format
      (verify no legacy patterns remain).
      _File: `examples/nexus.unm.yaml`_ (#docs)
- [x] **10.6.3** — Deduplicate multi-actor needs in example files (backlog
      item 9.11.9). _Files: `examples/*.unm.yaml`_ (#docs)
- [x] **10.6.4** — Remove or update `examples/inca.unm` if it uses legacy
      patterns. Verify `inca.unm.yaml` references in `CLAUDE.md`,
      `config/base.yaml`, `dsl_serializer_test.go` are valid. Updated
      dsl_serializer_test.go to reference nexus.unm.yaml instead.
      _Files: examples, config, tests_ (#fullstack)

### 10.7 — Clean Up AI Templates

- [x] **10.7.1** — Update `extract-actions.tmpl` wording from `realizedBy`
      to `realizes` (display string). Review all AI prompt templates for
      legacy terminology.
      _File: `ai/prompts/`_ (#backend)
- [x] **10.7.2** — Review `ai_context_builder.go` field naming:
      `RealizingServices` is fine, but confirm no user-facing output still
      says "realizedBy".
      _File: `usecase/ai_context_builder.go`_ (#backend)

### 10.8 — Clean Up README

- [x] **10.8.1** — Update README.md DSL example: change `realizedBy` to
      `realizes` on service block.
      _File: `README.md`_ (#docs)

---

## Phase 11: Documentation Alignment

After Phase 10, only v2 patterns exist. Docs must reflect this single reality.

### 11.1 — Rewrite DSL Specification

The spec currently documents `realizedBy` as the authoring pattern and does
not mention `service.realizes`, `service.externalDeps`, or `team.interacts`.

- [x] **11.1.1** — Rewrite §2.4 (Capabilities): remove `realizedBy`, add
      `parent` field, document visibility inheritance.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **11.1.2** — Rewrite §2.5 (Services): add `realizes` (with roles),
      `externalDeps`, `dependsOn`. Update required/optional fields table.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **11.1.3** — Rewrite §2.6/§2.7 (Interactions): document `team.interacts`
      as primary. Decide whether standalone arrow syntax stays in DSL.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **11.1.4** — Rewrite §2.10 (External Dependencies): remove `usedBy`,
      document as name+description definitions only (edges from services).
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **11.1.5** — Update meta-model table (§1): add all service fields
      (`realizes`, `externalDeps`, `dependsOn`, `ownedBy`).
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **11.1.6** — Remove `scenarios` from spec (currently says "deprecation
      warning" but parser silently ignores).
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)
- [x] **11.1.7** — Add multi-actor needs syntax to spec.
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 11.2 — Rewrite YAML Guide

- [x] **11.2.1** — Rewrite YAML guide using v2-only patterns throughout.
      Remove all `realizedBy`, `usedBy`, top-level `interactions` examples.
      _File: `docs/YAML_GUIDE.md`_ (#docs)

### 11.3 — Rewrite DSL Guide

- [x] **11.3.1** — Rewrite DSL guide using v2-only patterns throughout.
      _File: `docs/DSL_GUIDE.md`_ (#docs)

### 11.4 — Derived vs Authored Field Reference

Review 2 §5: "You need explicit derived vs authored fields." Without this,
the meta-model freeze is incomplete — users and AI need to know which
fields they author and which the system computes.

- [x] **11.4.1** — Create `docs/META_MODEL_REFERENCE.md`: a single table
      listing every entity field, whether it is authored (written in .unm
      file) or derived (computed by the system), the canonical relationship
      direction, and which fields were removed in the v2 freeze.
      _File: `docs/META_MODEL_REFERENCE.md`_ (#docs)
- [x] **11.4.2** — Update DSL Specification to reference the meta-model
      table and explicitly mark derived fields (e.g., reverse lookups like
      `GetCapabilitiesForService` are derived, not authored).
      _File: `docs/UNM_DSL_SPECIFICATION.md`_ (#docs)

### 11.5 — Fix Stale References

- [x] **11.5.1** — Audit `CLAUDE.md` for stale example file references.
      _File: `CLAUDE.md`_ (#docs)
- [x] **11.5.2** — Audit `docs/CONFIGURATION.md` for stale paths.
      _File: `docs/CONFIGURATION.md`_ (#docs)

---

## Phase 12: Test Infrastructure & CI

Address review concern: "analyzer correctness will be the product's
long-term credibility risk" and "the biggest missing test investment."

### 12.1 — Golden Model Fixtures

All golden tests use `nexus.unm` (DSL format) as the input model — DSL is
the primary authoring format. Golden output files are `.json` because they
capture the expected REST API response shape (what the handler returns).

- [x] **12.1.1** — Create `testdata/golden/nexus-need-view.json` — expected
      enriched need view API response for `nexus.unm`. Snapshot test: parse
      DSL → call handler → compare JSON output against golden file.
      _Files: `testdata/golden/`, `handler/view_test.go`_ (#backend)
- [x] **12.1.2** — Create golden fixtures for: capability view, ownership
      view, team-topology view, cognitive-load view, signals view.
      _File: `testdata/golden/`_ (#backend)
- [ ] **12.1.3** — Create golden fixtures for analyzer outputs: fragmentation,
      cognitive-load, bottleneck, gap, coupling reports against nexus model.
      _File: `testdata/golden/`_ (#backend)

### 12.2 — Missing Backend Tests

- [x] **12.2.1** — Add HTTP tests for changeset commit endpoint
      (`POST .../commit`). Test: valid commit, conflict (409), validation
      failure.
      _File: `handler/changeset_test.go`_ (#backend)
- [x] **12.2.2** — Add `changeset_store_test.go` — dedicated tests for
      ChangesetStore CRUD and cascade delete.
      _File: `adapter/repository/changeset_store_test.go`_ (#backend)
- [x] **12.2.3** — Add round-trip test: parse `nexus.unm` (DSL) → serialize
      to DSL → parse again → compare models field-by-field. This is the
      primary round-trip path. Fix `dsl_serializer_test.go` skip-if-missing.
      _File: `serializer/dsl_serializer_test.go`_ (#backend)
- [x] **12.2.4** — Add cross-format round-trip test: parse `nexus.unm` (DSL)
      → serialize to YAML → parse YAML → compare models. Verifies format
      interop produces identical domain models.
      _File: `serializer/yaml_serializer_test.go`_ (#backend)

### 12.3 — Reference Validation Depth

Review 2 §6: "Reference validation is still relatively shallow." Improve
error quality to support real authoring at scale.

- [x] **12.3.1** — Add error locality: validation errors should include
      the entity name and field where the broken reference occurs, not just
      "unresolved reference to X".
      _File: `domain/service/validation_engine.go`_ (#backend)
- [x] **12.3.2** — Distinguish "missing entity" from "model incomplete":
      if a capability has no `realizedBy` relationships, is that a gap or
      just an incomplete model? Add severity levels (error vs warning vs
      info) based on model completeness context.
      _File: `domain/service/validation_engine.go`_ (#backend)
- [x] **12.3.3** — Add validation for orphaned entities: services not
      referenced by any capability, teams that own nothing, actors with no
      needs. Surface as info-level diagnostics.
      _File: `domain/service/validation_engine.go`_ (#backend)

### 12.5 — DSL Parser Warning Parity

- [x] **12.5.1** — Add `Warnings` population to DSL parser (parity with
      YAML parser). Unresolved references, unknown fields should produce
      warnings on `.unm` files too.
      _File: `parser/dsl/transformer.go`_ (#backend)
- [x] **12.5.2** — Tests for DSL warning output.
      _File: `parser/dsl/transformer_test.go`_ (#backend)

### 12.6 — Frontend Tests in CI

- [x] **12.6.1** — Add `npm run test -- --run` (non-watch mode) to
      `.github/workflows/ci.yml` frontend job.
      _File: `.github/workflows/ci.yml`_ (#infra)
- [x] **12.6.2** — Add smoke tests for untested views: NeedView,
      TeamTopologyView, OwnershipView, RealizationView, CognitiveLoadView.
      _File: `frontend/src/pages/views/`_ (#frontend)
- [x] **12.6.3** — Add smoke tests for DashboardPage and UploadPage.
      _File: `frontend/src/pages/`_ (#frontend)
- [ ] **12.6.4** — Add smoke tests for WhatIfPage and AdvisorPage.
      _File: `frontend/src/pages/`_ (#frontend)

---

## Phase 13: Codebase Quality

Address review concerns about projection concentration and code health.

### 13.1 — Split `view_enriched.go`

The file is 1,281 lines and concentrates all projection logic. Split into
per-view builders with shared utilities.

- [x] **13.1.1** — Extract shared precomputation (cap→service maps,
      team→service maps, anti-pattern lookups) into `view_helpers.go`.
      _File: `handler/view_helpers.go`_ (#backend)
- [x] **13.1.2** — Extract need view projection into `view_need.go`.
      _File: `handler/view_need.go`_ (#backend)
- [x] **13.1.3** — Extract capability view projection into `view_capability.go`.
      _File: `handler/view_capability.go`_ (#backend)
- [x] **13.1.4** — Extract team/cognitive-load/ownership projections into
      `view_team.go`.
      _File: `handler/view_team.go`_ (#backend)
- [x] **13.1.5** — Extract realization and value-stream projections into
      `view_realization.go`.
      _File: `handler/view_realization.go`_ (#backend)
- [x] **13.1.6** — Extract UNM map projection into `view_map.go`.
      _File: `handler/view_map.go`_ (#backend)
- [x] **13.1.7** — Delete `view_enriched.go` once all projections are
      extracted. Ensure all existing view tests still pass.

### 13.2 — Handler Decomposition

Review 2 §4: "Handler is still a heavy dependency bucket... it will become
a scaling issue if persistence and multi-user concerns arrive." Must be
addressed BEFORE Phase 14 adds auth, orgs, and workspaces.

- [ ] **13.2.1** — Extract model-related handlers into `model_handler.go`.
      Move parse, validate, export, model CRUD operations.
      _File: `handler/model_handler.go`_ (#backend)
      _Note: model.go already provides this separation; 13.2 is partially done — Handler struct is already well-decomposed into focused .go files_
- [ ] **13.2.2** — Extract changeset handlers into `changeset_handler.go`.
      Move create, get, projected, impact, apply, commit, explain.
      _File: `handler/changeset_handler.go`_ (#backend)
- [ ] **13.2.3** — Extract analysis/view handlers into `view_handler.go`.
      Move view dispatch, signals, analysis endpoints.
      _File: `handler/view_handler.go`_ (#backend)
- [ ] **13.2.4** — Extract AI handlers into `ai_handler.go`. Move ask,
      extract-actions, insights endpoints.
      _File: `handler/ai_handler.go`_ (#backend)
- [ ] **13.2.5** — Reduce `Handler` struct to a thin router that delegates
      to sub-handlers. Each sub-handler gets only the dependencies it needs.
      _File: `handler/handler.go`_ (#backend)
- [ ] **13.2.6** — Ensure all handler tests still pass after decomposition.

### 13.3 — CORS Cleanup

- [x] **13.3.1** — Remove the backward-compat `corsMiddleware` with `"*"`.
      Only the configured-origin middleware should exist.
      _File: `handler/middleware.go`_ (#backend)

### 13.4 — Analyzer Explainability

Both reviews flag "if findings feel noisy or arbitrary, trust drops."

- [x] **13.4.1** — Add `Explanation` field to analyzer findings: each finding
      should include a human-readable sentence explaining why it was flagged
      and what threshold was breached.
      _File: `domain/entity/signal.go`, `analyzer/*.go`_ (#backend)
- [ ] **13.4.2** — Surface explanations in SignalsView.
      _File: `SignalsView.tsx`_ (#frontend)

### 13.5 — AI Trust Layering

Review 2 §10: "The product will need a very clear distinction between
model-derived facts, analyzer-derived findings, AI-generated interpretation.
If those blur, trust drops quickly."

- [x] **13.5.1** — Define a `SourceType` enum: `model_fact`, `analyzer_finding`,
      `ai_interpretation`. Add to all view payloads and signal responses so
      the frontend can distinguish origin.
      _File: `domain/valueobject/source_type.go`, `entity/signal.go`_ (#backend)
- [x] **13.5.2** — Tag all analyzer findings with `analyzer_finding` source.
      Tag AI insight responses with `ai_interpretation` source.
      _File: `analyzer/*.go`, `handler/ai.go`_ (#backend)
- [ ] **13.5.3** — Frontend: render trust indicators — facts with solid
      styling, analyzer findings with threshold citation, AI output with
      "AI-generated" badge. Users must always know what is fact vs opinion.
      _File: `components/signals/SignalCard.tsx`, `components/advisor/`_ (#frontend)

---

## Phase 14: Persistence, Auth & Multi-Tenancy

**Architecture document: `docs/ARCHITECTURE_EVOLUTION.md`**
Read that document first. It defines the data hierarchy (Organization →
Workspace → Model → Version), tenancy model, authorization matrix, API
route evolution, and database choice. The backlog items below implement
that design.

### 14A — Repository Interfaces & PostgreSQL Foundation

- [ ] **14A.1** — Define repository interfaces in usecase package:
      `ModelRepository`, `ChangesetRepository`, `VersionRepository` with
      operations scoped by workspace context.
      _File: `usecase/repository.go`_ (#backend)
- [ ] **14A.2** — Add PostgreSQL driver (`pgx`) and migration tool
      (`golang-migrate`). _File: `go.mod`_ (#backend)
- [ ] **14A.3** — Create initial migration: `users`, `organizations`,
      `org_memberships`, `workspaces`, `workspace_memberships`, `models`,
      `model_versions`, `changesets` tables.
      Schema defined in `docs/ARCHITECTURE_EVOLUTION.md` §3.
      _File: `infrastructure/persistence/migrations/001_initial.up.sql`_ (#backend)
- [ ] **14A.4** — Implement PostgreSQL-backed `ModelRepository`.
      All queries scoped through workspace → org chain.
      _File: `infrastructure/persistence/pg_model_store.go`_ (#backend)
- [ ] **14A.5** — Implement PostgreSQL-backed `ChangesetRepository`.
      _File: `infrastructure/persistence/pg_changeset_store.go`_ (#backend)
- [ ] **14A.6** — Implement PostgreSQL-backed `VersionRepository`.
      _File: `infrastructure/persistence/pg_version_store.go`_ (#backend)
- [ ] **14A.7** — Config: `storage.driver: postgres|memory`,
      `storage.database_url`, `storage.migrate_on_startup`.
      _Files: `entity/config.go`, `config/base.yaml`_ (#backend)
- [ ] **14A.8** — Wire in `main.go`: construct PG stores when driver=postgres,
      run migrations on startup. Keep memory stores for driver=memory.
      _File: `cmd/server/main.go`_ (#backend)
- [ ] **14A.9** — Tests: repository interface tests that run against both
      memory and postgres implementations.
      _File: `persistence/*_test.go`_ (#backend)
- [ ] **14A.10** — Docker Compose: add PostgreSQL service for local dev.
      _File: `docker-compose.yml`_ (#infra)

### 14B — Model History & Multi-Model

- [ ] **14B.1** — On changeset commit, create a new `ModelVersion` record
      with raw content and commit message.
      _File: `handler/changeset.go`_ (#backend)
- [ ] **14B.2** — API: `GET .../models` — list models in workspace.
      _File: `handler/model.go`_ (#backend)
- [ ] **14B.3** — API: `GET .../models/{id}/history` — list versions.
      _File: `handler/model.go`_ (#backend)
- [ ] **14B.4** — API: `GET .../models/{id}/versions/{v}` — get version.
      _File: `handler/model.go`_ (#backend)
- [ ] **14B.5** — API: `GET .../models/{id}/diff?from={v1}&to={v2}` —
      structured diff between two model versions. Returns added/removed/
      changed entities grouped by type. Without diff, history is just a
      list of dates.
      _File: `handler/model.go`, `domain/service/model_diff.go`_ (#backend)
- [ ] **14B.6** — Frontend: Model list page (all models in workspace).
      _File: `pages/ModelsPage.tsx`_ (#frontend)
- [ ] **14B.7** — Frontend: Version history panel with inline diff viewer.
      _File: `components/model/VersionHistory.tsx`_ (#frontend)
- [ ] **14B.8** — Remove session TTL eviction. Make eviction opt-in for
      memory-only mode.
      _File: `cmd/server/main.go`_ (#backend)

### 15A — Authentication

- [ ] **15A.1** — Add `golang.org/x/oauth2` + Google provider. Config:
      `auth.enabled`, `auth.google.client_id`, `auth.google.client_secret`,
      `auth.google.redirect_url`.
      _Files: `go.mod`, `entity/config.go`, `config/base.yaml`_ (#backend)
- [ ] **15A.2** — Implement OAuth flow: `GET /auth/google` (redirect),
      `GET /auth/callback` (exchange + verify + create user + set session).
      _File: `handler/auth.go`_ (#backend)
- [ ] **15A.3** — Session management: secure HTTP-only cookie. Sessions
      stored in PostgreSQL (or in-memory for dev mode).
      _File: `infrastructure/persistence/pg_session_store.go`_ (#backend)
- [ ] **15A.4** — `GET /api/me` — return current user + org memberships.
      _File: `handler/auth.go`_ (#backend)
- [ ] **15A.5** — `POST /auth/logout` — clear session.
      _File: `handler/auth.go`_ (#backend)
- [ ] **15A.6** — Auth middleware: verify session on `/api/*` routes.
      401 for unauthenticated. Exclude `/health`, `/auth/*`, static assets.
      _File: `handler/middleware.go`_ (#backend)
- [ ] **15A.7** — Local dev mode: when `auth.enabled: false`, inject default
      user + default org + default workspace context into all requests.
      _File: `handler/middleware.go`_ (#backend)
- [ ] **15A.8** — Frontend: Login page with "Sign in with Google."
      _File: `pages/LoginPage.tsx`_ (#frontend)
- [ ] **15A.9** — Frontend: Auth context (check `/api/me`, redirect to
      login on 401, store user info).
      _File: `lib/auth-context.tsx`_ (#frontend)
- [ ] **15A.10** — Frontend: Protected route wrapper.
      _File: `App.tsx`_ (#frontend)
- [ ] **15A.11** — Frontend: User info in TopBar (name, avatar, logout).
      _File: `components/layout/TopBar.tsx`_ (#frontend)

### 15B — Organizations & Workspaces

- [ ] **15B.1** — Org management APIs: create org, list user's orgs, get
      org details, manage org members.
      _File: `handler/org.go`_ (#backend)
- [ ] **15B.2** — Workspace management APIs: create workspace, list
      workspaces in org, manage workspace members.
      _File: `handler/workspace.go`_ (#backend)
- [ ] **15B.3** — Migrate all model/changeset/view/analysis routes to
      workspace-scoped paths: `/api/orgs/{slug}/ws/{slug}/models/...`
      _File: `handler/router.go`_ (#backend)
- [ ] **15B.4** — Onboarding flow: first login creates org + default
      workspace. _File: `handler/auth.go`_ (#backend)
- [ ] **15B.5** — Frontend: Org selector (if user has multiple orgs).
      _File: `components/layout/OrgSelector.tsx`_ (#frontend)
- [ ] **15B.6** — Frontend: Workspace management page.
      _File: `pages/WorkspacePage.tsx`_ (#frontend)
- [ ] **15B.7** — Frontend: Update API client with org/workspace path
      scoping. _File: `services/api/client.ts`_ (#frontend)

### 15C — Authorization

- [ ] **15C.1** — Authorization middleware: check workspace membership and
      role before handler execution. Return 403 on insufficient permissions.
      _File: `handler/middleware.go`_ (#backend)
- [ ] **15C.2** — Implement permission matrix from
      `ARCHITECTURE_EVOLUTION.md` §5.
      _File: `handler/authorization.go`_ (#backend)
- [ ] **15C.3** — Changeset ownership: only creator or workspace admin
      can commit.
      _File: `handler/changeset.go`_ (#backend)
- [ ] **15C.4** — Org-visible workspaces: any org member can read, only
      workspace members can write.
      _File: `handler/authorization.go`_ (#backend)

---

## Phase 16: Collaboration & Multi-User Features

**Architecture document: `docs/ARCHITECTURE_EVOLUTION.md` §7**

Builds on the auth + tenancy foundation from Phases 14-15.

### 16.1 — Changeset Collaboration

- [ ] **16.1.1** — Changeset comments: `changeset_comments` table,
      `POST/GET .../changesets/{id}/comments` API.
      _File: `handler/changeset.go`, `persistence/pg_changeset_store.go`_ (#backend)
- [ ] **16.1.2** — Frontend: comment thread on changeset review dialog.
      _File: `components/changeset/ChangesetComments.tsx`_ (#frontend)
- [ ] **16.1.3** — Author attribution: show who created/committed each
      changeset and model version.
      _File: `components/changeset/ChangesetList.tsx`_ (#frontend)

### 16.2 — Changeset Review Workflow

Review 1: "Change review workflow (PR-like review for changesets)." In
multi-user, structural changes need approval, not just comments.

- [ ] **16.2.1** — Add changeset status lifecycle: `draft → in_review →
      approved → committed` (or `rejected`). Only workspace admin or
      designated reviewer can approve.
      _File: `entity/changeset.go`, `handler/changeset.go`_ (#backend)
- [ ] **16.2.2** — API: `POST .../changesets/{id}/review` with
      `{action: approve|reject|request_changes, comment: "..."}`.
      _File: `handler/changeset.go`_ (#backend)
- [ ] **16.2.3** — Frontend: review status badges and approve/reject
      actions on changeset dialog.
      _File: `components/changeset/ReviewDialog.tsx`_ (#frontend)

### 16.3 — API Key Auth (CI/CD Integration)

- [ ] **16.3.1** — API key table: `api_keys(id, workspace_id, role, hashed_key,
      name, created_by, created_at, expires_at)`.
      _File: `persistence/migrations/`, `handler/api_key.go`_ (#backend)
- [ ] **16.3.2** — Auth middleware: accept `Authorization: Bearer <key>` in
      addition to session cookies. Resolve key to workspace + role.
      _File: `handler/middleware.go`_ (#backend)
- [ ] **16.3.3** — GitHub Action: `unm-validate` — validates `.unm` files in
      a PR using API key auth. Runs parser + analyzers, comments with findings.
      _File: `actions/` directory_ (#infra)
- [ ] **16.3.4** — CLI `validate` command with machine-readable JSON output.
      _File: `cmd/cli/`_ (#backend)

---

## Phase 17: Product Hardening

Features that make the core workflow "boringly reliable" (Review 2).

### 17.1 — Export Formats

- [ ] **17.1.1** — Mermaid diagram export (graph of actors → needs →
      capabilities → services → teams).
      _File: `serializer/mermaid_serializer.go`_ (#backend)
- [ ] **17.1.2** — JSON export (full model as JSON, useful for integrations).
      _File: `handler/model.go`_ (#backend)

### 17.2 — Changeset Reliability

Review 2: "edit/apply/commit complexity is where correctness bugs accumulate."

- [ ] **17.2.1** — Changeset round-trip tests: create changeset → apply →
      commit → export → parse → verify model matches expected state.
      _File: `handler/changeset_test.go`_ (#backend)
- [ ] **17.2.2** — Add changeset undo: revert to pre-commit model version
      (requires Phase 14 model history).
      _File: `handler/changeset.go`_ (#backend)

### 17.3 — Analyzer Calibration

Review 2: "Threshold tuning is likely under-validated."

- [ ] **17.3.1** — Document all analyzer thresholds in a single reference
      file. Include rationale for each default value.
      _File: `docs/ANALYZER_REFERENCE.md`_ (#docs)
- [ ] **17.3.2** — Make key thresholds configurable via `config/base.yaml`
      (already partially done for `overloaded_capability_threshold`; extend
      to cognitive load, bottleneck, coupling thresholds).
      _File: `entity/config.go`, `config/base.yaml`_ (#backend)

### 17.4 — Model Completeness Indicators

Review 2: "analysis risks conflating model incompleteness with architectural
truth."

- [ ] **17.4.1** — Add model completeness score: percentage of entities with
      relationships, services with ownership, capabilities with realization.
      Surface in dashboard.
      _File: `analyzer/completeness.go`, `DashboardPage.tsx`_ (#fullstack)
- [ ] **17.4.2** — Analyzer findings should indicate confidence level based
      on model completeness. Low completeness → findings marked as
      "low confidence" rather than stated as facts.
      _File: `analyzer/*.go`, `entity/signal.go`_ (#backend)

---

## Phase 18: Ecosystem Integration & Onboarding

Address Review 1: "No integration surface" and "the first 10 hours of
investment before seeing insight is a high bar." These make the platform
usable in real engineering workflows and lower the cold-start barrier.

### 18.1 — Git Repository Integration

Review 1: "auto-load from a repo path, webhook on push."

- [ ] **18.1.1** — Git import: API endpoint to clone a repo, find `.unm`
      files, parse and store them as models in a workspace.
      _File: `handler/git_import.go`_ (#backend)
- [ ] **18.1.2** — GitHub webhook: receive push events, auto-update models
      when `.unm` files change in a configured repo/branch.
      _File: `handler/webhook.go`_ (#backend)
- [ ] **18.1.3** — Frontend: "Import from Git" flow on workspace page.
      _File: `pages/WorkspacePage.tsx`_ (#frontend)

### 18.2 — Model Onboarding

Review 1: "Building a comprehensive .unm model requires effort. The first
10 hours before seeing insight is a high bar."

- [ ] **18.2.1** — AI-assisted model generation: user provides a text
      description of their org/system, AI generates a starter `.unm` model.
      _File: `handler/ai.go`, `ai/prompts/generate-model.tmpl`_ (#backend)
- [ ] **18.2.2** — Getting-started wizard: guided flow that asks questions
      (what teams exist, what services, who are users) and builds a model
      incrementally.
      _File: `pages/OnboardingWizard.tsx`_ (#frontend)
- [ ] **18.2.3** — Template library: pre-built starter models for common
      patterns (microservices team, platform team, startup, etc.).
      _File: `examples/templates/`_ (#docs)

### 18.3 — Import from External Tools

- [ ] **18.3.1** — Import from Structurizr (C4 model → UNM model mapping).
      _File: `parser/structurizr_importer.go`_ (#backend)
- [ ] **18.3.2** — Import from Backstage catalog (catalog-info.yaml →
      services/teams). _File: `parser/backstage_importer.go`_ (#backend)

### 18.4 — Notification Integrations

- [ ] **18.4.1** — Webhook notifications: fire webhooks on changeset
      created, committed, model updated.
      _File: `handler/notifications.go`_ (#backend)
- [ ] **18.4.2** — Slack integration: post changeset summaries to a
      configured Slack channel.
      _File: `infrastructure/notifications/slack.go`_ (#backend)

---

## Bugs

_(No open bugs)_

---

## Execution Order

```
Phase 10 (Model Freeze)       ─── 1-2 days, pure removal, no new code
    │
Phase 11 (Docs)                ─── 1-2 days, meta-model ref + spec rewrite
    │
Phase 12 (Tests & CI)          ─── 2-3 days, golden fixtures + validation depth
    │
Phase 13 (Code Quality)        ─── 2-3 days, handler decomp + view split +
    │                                AI trust layering + explainability
    │
    ├── ARCHITECTURE_EVOLUTION.md ─── APPROVED (all decisions resolved)
    │
Phase 14A (PG Foundation)      ─── 3-4 days, schema + stores + migrations
    │
Phase 14B (History + Multi)    ─── 2-3 days, versions + diff + model list
    │
Phase 15A (Auth)               ─── 2-3 days, Google OAuth + session + UI
    │
Phase 15B (Orgs + Workspaces)  ─── 3-4 days, tenancy + management + routes
    │
Phase 15C (Authorization)      ─── 1-2 days, role checks + permissions
    │
Phase 16 (Collaboration)      ─── 3-4 days, comments + review workflow +
    │                                API keys + CI action
    │
Phase 17 (Hardening)          ─── ongoing, export + calibration + completeness
    │
Phase 18 (Ecosystem)          ─── git integration, onboarding, import, notifications
```

Phases 10–13: rapid cleanup, docs, tests, and decomposition. No new features,
but critical preparation for what follows.
Before Phase 14: **`ARCHITECTURE_EVOLUTION.md` approved** — all 6
architecture decisions resolved (AD-1 through AD-6).
Phases 14–16: structural additions, built on the architecture document.
Phase 17: ongoing hardening, interleaved as needed.
Phase 18: ecosystem integration — makes the platform usable in real workflows.
