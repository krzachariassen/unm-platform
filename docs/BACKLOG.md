# UNM Platform — Backlog

> **File ownership: backlog-manager agent ONLY.**
> This file is exclusively managed by the backlog-manager agent
> (`.claude/agents/backlog-manager/AGENT.md`). No other agent,
> orchestrator, or engineer may edit this file directly. If you need
> to add items, restructure phases, update checkboxes, or add
> implementation detail — invoke the backlog-manager agent.
> Violations break task tracking and cause duplicate or lost work.

_Single source of truth for all work items.
Completed phases: `docs/PRODUCT_ROADMAP.md`.
Implementation patterns: `.claude/agents/` and `.claude/rules/`._

_Last updated: 2026-04-03 (Phase 14B backend complete; frontend 14B.6–14B.7 pending)_
_Priority: **14C** (model list UI) → **Phase F** (frontend restructure) → **Phase 15** (auth/tenancy + UI) → **Phase 16** (collaboration + UI) → Phase 17 (hardening) → Phase 18 (ecosystem)._

_Context: Two independent external reviews identified migration completion,
meta-model stability, analyzer trust, test infrastructure, persistence, and
auth as the critical path to product readiness. No users exist — backward
compatibility is not required. All legacy patterns can be removed outright._

---

## Recently Completed

- [x] **feat(phase-FC.4-6): New Analyzer Tabs** — GapsTab in NeedsPage (5 gap categories: unmapped needs, unrealized caps, unowned services, unneeded caps, orphan services); DependenciesTab in CapabilitiesPage (stat cards, cycle path rendering, critical service path chain); InteractionsTab in TeamsPage (mode distribution bars, isolated teams, over-reliant teams, all_modes_same warning banner); new view types GapsView/DependenciesView/InteractionsView; API functions getGaps/getDependencies/getInteractions; 18 new tests, 118 total pass. FC.4–FC.6 complete. (2026-04-03)

- [x] **feat(parse): auto-detect DSL vs YAML in parse/validate endpoints** — Sniff first 64 bytes of body; content starting with `system ` or `system"` is automatically routed to the DSL parser. Explicit `?format=dsl|yaml` still takes precedence. Two new handler tests added. (2026-04-03)

- [x] **feat(phase-FA+FB.1+FB.3): Phase F — View Regrouping, Tabs & Interaction Consistency** — `UrlTabBar` component with URL ?tab= deep linking; `NeedsPage` (Overview/Traceability tabs), `CapabilitiesPage` (Hierarchy/Services tabs), `TeamsPage` (Topology/Ownership/Cognitive Load tabs); sidebar restructured from 12→10 items with new Architecture section; backward-compat redirects for all old routes; OwnershipView custom popover replaced with `SlidePanel`; QuickAction opacity standardized to 0.7 in TeamLane and GraphView; 14 new tests, 105 total pass. FA.1–FA.6, FB.1, FB.3 complete. (2026-04-03)

- [x] **feat(phase-14c): Model List & History UI (frontend)** — ModelsPage with card grid (name, created_at, version_count, Load button, empty/loading states); ModelHistoryPage with version timeline (commit message, date, compare mode); DiffViewer component (added/removed/changed entities, green/red/amber color coding, per-entity-type grouping); API functions listModels, loadStoredModel, getHistory, getDiff added to services/api/models.ts; new types ModelListItem, VersionMeta, DiffEntities, DiffResult in types/model.ts; /models and /history routes; "All Models" and "History" sidebar entries; 13 new tests, all 91 tests pass. 14C.1–14C.3 complete. (2026-04-03)

- [x] **feat(phase-14b): Model History & Multi-Model (backend)** — `ReplaceWithMessage` added to `ModelRepository` interface; changeset commit now stores description as `commit_message` in `model_versions`; `GET /api/models` list endpoint with version counts; `GET /api/models/{id}/history` with per-version metadata; `GET /api/models/{id}/versions/{v}` retrieves model at specific version; `GET /api/models/{id}/diff?from=&to=` computes structural diff (added/removed/changed by entity type); `domain/service/model_diff.go` with `Diff()` and `DiffEntities`; memory store stubs for all new interface methods; PG store queries `model_versions` table for real multi-version history; 14B.8 eviction scoped to memory-only (log note added for postgres path); handler tests and contract tests for all new endpoints/methods. 14B.1–14B.5, 14B.8 complete. (2026-04-03)

- [x] **feat(phase-14a): PostgreSQL Foundation** — Repository interfaces (`usecase.ModelRepository`, `usecase.ChangesetRepository`) extracted to usecase package; `StoredModel`/`StoredChangeset` DTOs moved to usecase; all handlers updated to depend on interfaces not concrete types; `PGModelStore` and `PGChangesetStore` implemented (pgx v5, golang-migrate, soft delete, system user/workspace bootstrap); initial SQL migration for 8 tables (users, organizations, workspaces, models, model_versions, changesets, memberships); `StorageConfig` added to entity config; `docker-compose.yml` updated with postgres:16-alpine service; `main.go` wired to select memory or postgres driver at startup; contract test suite runs same tests against both implementations (28 tests, all pass including against real postgres). Items 14A.1–14A.10 complete. (2026-04-03)

- [x] **refactor(phase-13): 13.6.1-13.6.4, 13.7.1-13.7.4, 13.5.3, stale comment cleanup** — Removed `capability.RealizedBy` derived field and `AddRealizedBy()` from domain entity; all analyzers migrated to `GetServicesForCapability()`; stale `realizedBy` comments purged from yaml_parser, complexity, fragmentation; created `changeset_dto.go` DTO layer in handler, removed all `json:` tags from `entity.ChangeAction` / `entity.Changeset`; extracted `modelHandler`, `changesetHandler`, `viewHandler`, `aiHandler` sub-structs from monolithic Handler; added analyzer golden fixtures for bottleneck/coupling/complexity/fragmentation/gap/cognitive-load; added frontend trust badges (amber=analyzer, blue=AI) and `Explanation` text to all signal rows in SignalsView; added WhatIfPage and AdvisorPage smoke tests (78 total frontend tests pass). All 15 backend packages pass. (2026-04-03)
- [x] **refactor(phase-13): 13.1, 13.3, 13.4, 13.5 (backend)** — Split `view_enriched.go` (1281 lines) into 7 focused files: `view_helpers.go`, `view_need.go`, `view_capability.go`, `view_team.go`, `view_ownership.go`, `view_realization.go`, `view_map.go`; deleted `view_enriched.go`; removed backward-compat `corsMiddleware` with `"*"`; added `Explanation` field to `Signal` and `SuggestedSignal` with human-readable threshold explanations in all 5 analyzer rules; defined `SourceType` enum (`model_fact`, `analyzer_finding`, `ai_interpretation`) in `domain/valueobject`; added `Source SourceType` to `Signal` entity and `SourceTag` to `SuggestedSignal` with `SourceAnalyzerFinding` tagging; all 14 packages pass. (2026-04-03)
- [x] **test(phase-12): 12.6.1-12.6.4** — Frontend tests in CI: added `npm run test` step to `.github/workflows/ci.yml`; smoke tests for NeedView, TeamTopologyView, OwnershipView, RealizationView, CognitiveLoadView, DashboardPage, UploadPage, WhatIfPage, AdvisorPage (78 tests total across 18 test files, all passing). (2026-04-02/03)
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

## Phase 14: Persistence, Auth & Multi-Tenancy

**Architecture documents:**
- Backend: `docs/ARCHITECTURE_EVOLUTION.md` — data hierarchy, tenancy,
  auth, API routes, database choice
- Frontend: `docs/FRONTEND_EVOLUTION.md` — navigation layers, tab
  structure, new views, platform chrome

Read both documents first. Each phase contains backend **and** frontend
items together. See Execution Order at the bottom for dependencies.

### 14A — Repository Interfaces & PostgreSQL Foundation

- [x] **14A.1** — Define repository interfaces in usecase package:
      `ModelRepository`, `ChangesetRepository`, `VersionRepository` with
      operations scoped by workspace context.
      _File: `usecase/repository.go`_ (#backend)
- [x] **14A.2** — Add PostgreSQL driver (`pgx`) and migration tool
      (`golang-migrate`). _File: `go.mod`_ (#backend)
- [x] **14A.3** — Create initial migration: `users`, `organizations`,
      `org_memberships`, `workspaces`, `workspace_memberships`, `models`,
      `model_versions`, `changesets` tables.
      Schema defined in `docs/ARCHITECTURE_EVOLUTION.md` §3.
      _File: `infrastructure/persistence/migrations/001_initial.up.sql`_ (#backend)
- [x] **14A.4** — Implement PostgreSQL-backed `ModelRepository`.
      All queries scoped through workspace → org chain.
      _File: `infrastructure/persistence/pg_model_store.go`_ (#backend)
- [x] **14A.5** — Implement PostgreSQL-backed `ChangesetRepository`.
      _File: `infrastructure/persistence/pg_changeset_store.go`_ (#backend)
- [x] **14A.6** — Implement PostgreSQL-backed `VersionRepository`.
      _File: `infrastructure/persistence/pg_version_store.go`_ (#backend)
- [x] **14A.7** — Config: `storage.driver: postgres|memory`,
      `storage.database_url`, `storage.migrate_on_startup`.
      _Files: `entity/config.go`, `config/base.yaml`_ (#backend)
- [x] **14A.8** — Wire in `main.go`: construct PG stores when driver=postgres,
      run migrations on startup. Keep memory stores for driver=memory.
      _File: `cmd/server/main.go`_ (#backend)
- [x] **14A.9** — Tests: repository interface tests that run against both
      memory and postgres implementations.
      _File: `persistence/*_test.go`_ (#backend)
- [x] **14A.10** — Docker Compose: add PostgreSQL service for local dev.
      _File: `docker-compose.yml`_ (#infra)

### 14B — Model History & Multi-Model

- [x] **14B.1** — On changeset commit, create a new `ModelVersion` record
      with raw content and commit message. (2026-04-03)
      _File: `handler/changeset.go`_ (#backend)
- [x] **14B.2** — API: `GET .../models` — list models in workspace. (2026-04-03)
      _File: `handler/model.go`_ (#backend)
- [x] **14B.3** — API: `GET .../models/{id}/history` — list versions. (2026-04-03)
      _File: `handler/model.go`_ (#backend)
- [x] **14B.4** — API: `GET .../models/{id}/versions/{v}` — get version. (2026-04-03)
      _File: `handler/model.go`_ (#backend)
- [x] **14B.5** — API: `GET .../models/{id}/diff?from={v1}&to={v2}` —
      structured diff between two model versions. Returns added/removed/
      changed entities grouped by type. (2026-04-03)
      _File: `handler/model.go`, `domain/service/model_diff.go`_ (#backend)
- [x] **14B.6** — Eviction scoped to memory-only; log note added when
      driver=postgres and SessionTTL > 0. (2026-04-03)
      _File: `cmd/server/main.go`_ (#backend)

### 14C — Model List & History UI (#frontend, depends on 14B APIs)

- [x] **14C.1** — Model list page: card grid with name, created_at,
      version count badge, Load button. Empty/loading states. (2026-04-03)
      _File: `pages/ModelsPage.tsx`_ (#frontend)
- [x] **14C.2** — Version history page: timeline of model versions with
      commit message, date, and compare actions. (2026-04-03)
      _File: `pages/ModelHistoryPage.tsx`_ (#frontend)
- [x] **14C.3** — Inline diff viewer: compare two model versions,
      shows added/removed/changed entities grouped by type with
      green/red/amber color coding. (2026-04-03)
      _File: `components/model/DiffViewer.tsx`_ (#frontend)

### 14D — PostgreSQL Data Lifecycle (#backend)

Background purge for soft-deleted rows and test data isolation.
Must be done before Phase 15 — once real users/orgs exist, stale
test data and soft-deleted rows must not leak into tenant queries.

**Design decisions:**
- Mirror the memory store's eviction pattern (ticker goroutine in
  `StartEviction`, stop channel in `StopEviction`).
- **5-minute interval** — same as memory store; lightweight since the
  query hits indexed `deleted_at` columns and typically deletes 0 rows.
- **7-day retention** — gives time to recover accidental deletes before
  hard-delete. Configurable per deployment.
- **Delete order matters** — FK constraints require children first:
  `changesets → model_versions → models → workspaces`.
- CI doesn't need special TTL — CI uses ephemeral postgres containers.
  Test isolation is for **local dev** where the same DB persists.

- [ ] **14D.1** — Add `PurgeRetention` (default `168h` = 7d) and
      `PurgeInterval` (default `5m`) to `StorageConfig`. Set defaults in
      `DefaultConfig()`.
      _File: `entity/config.go`_ (#backend)
      ```go
      PurgeRetention time.Duration `koanf:"purge_retention"` // 7d
      PurgeInterval  time.Duration `koanf:"purge_interval"`  // 5m
      ```
- [ ] **14D.2** — Implement background purge in `PGModelStore.StartEviction`:
      ticker-based goroutine that hard-deletes rows where
      `deleted_at < NOW() - $1`. SQL (executed in order):
      _File: `persistence/pg_model_store.go`_ (#backend)
      ```sql
      DELETE FROM changesets     WHERE deleted_at < NOW() - $1;
      DELETE FROM model_versions WHERE deleted_at < NOW() - $1;
      DELETE FROM models         WHERE deleted_at < NOW() - $1;
      DELETE FROM workspaces     WHERE deleted_at < NOW() - $1;
      ```
      Use the existing `stopCh` pattern from memory store. Log deletions.
- [ ] **14D.3** — Wire purge config in `main.go`: call `StartEviction`
      with `cfg.Storage.PurgeRetention` and `cfg.Storage.PurgeInterval`
      when `driver=postgres`. Remove the "ignored for postgres" log line.
      _File: `cmd/server/main.go`_ (#backend)
- [ ] **14D.4** — Test isolation: add `setupTestOrgWorkspace(t, db)` helper
      in contract tests. Creates a unique org (`test-{random}`) + workspace
      per test run, returns IDs for store constructors. `t.Cleanup`
      hard-deletes all data in that org via `DELETE FROM organizations
      WHERE id = $1` (CASCADE handles children). Pass org/workspace IDs
      to PGModelStore/PGChangesetStore so test data is scoped.
      _File: `persistence/repository_contract_test.go`_ (#backend)
- [ ] **14D.5** — Verify: run contract tests twice against same DB,
      confirm `GET /api/models` returns no orphaned test data. Write a
      unit test that soft-deletes a model, sets retention to 0, triggers
      purge, and confirms the row is gone.
      _File: `persistence/*_test.go`_ (#backend)

---

## Phase F: Frontend Restructure (no backend dependency)

**Architecture document: `docs/FRONTEND_EVOLUTION.md`**
These items are frontend-only. They can run in parallel with any backend
phase. Do them early — the tab infrastructure is needed before F5/F6 UI
can be built on top.

### FA — View Regrouping & Tabs

Move existing views into grouped pages with horizontal tabs.

- [x] **FA.1** — Create shared `TabBar` and `TabbedPage` components.
      `TabBar` syncs with URL `?tab=` for deep linking.
      _Files: `components/ui/url-tab-bar.tsx`_ (#frontend) (2026-04-03)
- [x] **FA.2** — Create `NeedsPage` with tabs: Overview (NeedView),
      Traceability (Realization ValueChain).
      _File: `pages/NeedsPage.tsx`_ (#frontend) (2026-04-03)
- [x] **FA.3** — Create `CapabilitiesPage` with tabs: Hierarchy
      (CapabilityView), Services (Realization ServiceTable).
      _File: `pages/CapabilitiesPage.tsx`_ (#frontend) (2026-04-03)
- [x] **FA.4** — Create `TeamsPage` with tabs: Topology
      (TeamTopologyView), Ownership (OwnershipView), Cognitive Load
      (CognitiveLoadView).
      _File: `pages/TeamsPage.tsx`_ (#frontend) (2026-04-03)
- [x] **FA.5** — Update sidebar: Architecture section becomes UNM Map,
      Needs, Capabilities, Teams, Signals (8 items total, down from 12).
      _Files: `components/layout/Sidebar.tsx`, `App.tsx`_ (#frontend) (2026-04-03)
- [x] **FA.6** — Delete standalone RealizationView and CognitiveLoadView
      route entries. Keep components as tab content.
      _File: `App.tsx`_ (#frontend) (2026-04-03)

### FB — Interaction Consistency

Standardize click/edit patterns across all views.

- [x] **FB.1** — Replace Ownership service popover with `SlidePanel`.
      Remove `openSvcPopover` and custom `getBoundingClientRect()` logic.
      _File: `pages/views/OwnershipView.tsx`_ (#frontend) (2026-04-03)
- [ ] **FB.2** — Create unified `EntityDetailPanel` that adapts by
      entity type (team, service, capability). Replace `AntiPatternPanel`
      and feature-specific detail panels.
      _File: `components/detail/EntityDetailPanel.tsx`_ (#frontend)
- [x] **FB.3** — Standardize QuickAction pencil opacity (0.7+ default,
      not 0.35). Add QuickAction to Team Topology table rows.
      _Files: `features/team-topology/GraphView.tsx`,
      `features/team-topology/TableView.tsx`_ (#frontend) (2026-04-03)
- [ ] **FB.4** — Add AI insights to Ownership detail panel.
      _File: `features/ownership/TeamLane.tsx`_ (#frontend)
- [ ] **FB.5** — Add cross-view entity navigation: clicking a team name
      in Capability → navigates to Teams page; clicking a service in
      Need View → navigates to Capabilities/Services tab.
      _Files: multiple view components_ (#frontend)

### FC — New Analyzer Views (backend + frontend)

Expose analyzer data that exists in the backend but has no frontend view.
Backend items can start anytime; frontend items need FA (tabs) first.

- [ ] **FC.1** — Backend: `GET /views/gaps` endpoint. Wrap `GapAnalyzer`
      output. Include `orphan_services` (currently omitted from HTTP).
      _File: `handler/view_gaps.go`, `usecase/analysis_runner.go`_ (#backend)
- [ ] **FC.2** — Backend: `GET /views/dependencies` endpoint. Wrap
      `DependencyAnalyzer`, include `Service.DependsOn` graph.
      _File: `handler/view_dependencies.go`_ (#backend)
- [ ] **FC.3** — Backend: `GET /views/interactions` endpoint. Wrap
      `InteractionDiversityAnalyzer` output.
      _File: `handler/view_interactions.go`_ (#backend)
- [x] **FC.4** — Frontend: Gaps tab in NeedsPage. Show unmapped needs,
      unrealized caps, unowned services in categorized lists.
      _File: `features/needs/GapsTab.tsx`_ (#frontend)
- [x] **FC.5** — Frontend: Dependencies tab in CapabilitiesPage. Graph
      visualization of service dependencies with cycle highlighting.
      _File: `features/capabilities/DependenciesTab.tsx`_ (#frontend)
- [x] **FC.6** — Frontend: Interactions tab in TeamsPage. Mode
      distribution chart, isolated/over-reliant team indicators.
      _File: `features/teams/InteractionsTab.tsx`_ (#frontend)

---

## Phase 15: Authentication & Multi-Tenancy

**Architecture document: `docs/ARCHITECTURE_EVOLUTION.md` §4–6**

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
- [ ] **15B.6** — Frontend: Workspace dashboard page — model list,
      members, quick actions (import, create). Replaces Upload as home.
      _File: `pages/WorkspaceDashboardPage.tsx`_ (#frontend)
- [ ] **15B.7** — Frontend: Update API client with org/workspace path
      scoping. _File: `services/api/client.ts`_ (#frontend)
- [ ] **15B.8** — Frontend: Breadcrumb (Org > Workspace > Model) in TopBar.
      _File: `components/layout/Breadcrumb.tsx`_ (#frontend)
- [ ] **15B.9** — Frontend: URL structure migration
      (`/:orgSlug/:wsSlug/models/:id/...`).
      _Files: `App.tsx`, `services/api/client.ts`_ (#frontend)
- [ ] **15B.10** — Frontend: Sidebar context switching — workspace-level
      nav vs model-level nav.
      _File: `components/layout/Sidebar.tsx`_ (#frontend)
- [ ] **15B.11** — Frontend: Settings pages (org settings, workspace
      settings, API keys).
      _Files: `pages/settings/OrgSettingsPage.tsx`,
      `pages/settings/WorkspaceSettingsPage.tsx`_ (#frontend)

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
- [ ] **16.1.3** — Frontend: author attribution on changesets and
      model versions.
      _File: `components/changeset/ChangesetList.tsx`_ (#frontend)
- [ ] **16.1.4** — Frontend: changeset list with review status badges
      and author avatars. Filter by status (draft, in_review, approved).
      _File: `features/changes/ChangesetListTab.tsx`_ (#frontend)

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

- [ ] **BUG: Version diff returns empty on PostgreSQL** — `GET .../diff?from=1&to=2`
      returns all-null added/removed/changed even after a `move_service` commit.
      Likely cause: `GetVersion` deserializes the stored YAML but the round-trip
      loses derived ownership mappings (team→service), so both versions look
      identical to the diff logic. The diff needs to compare the fully-resolved
      model (with ownership graph rebuilt), not just the raw parsed output.
      _File: `persistence/pg_model_store.go`, `domain/service/model_diff.go`_ (#backend)
- [ ] **BUG: No cleanup process for soft-deleted / stale PostgreSQL data** —
      Soft-deleted models, versions, and changesets accumulate forever. Contract
      tests leave nameless models in the DB. Need: (1) periodic background purge
      of soft-deleted rows older than N days, (2) tests must clean up or use a
      separate database, (3) consider hard-deleting nameless/empty models on
      startup or via an admin endpoint.
      _File: `persistence/pg_model_store.go`, `cmd/server/main.go`_ (#backend)
- [x] **BUG: Parse endpoint requires manual `?format=dsl`** — Fixed: `handleParse`
      and `handleValidate` now sniff the first 64 bytes of the body. Content
      starting with `system ` or `system"` is auto-detected as DSL; everything
      else falls back to YAML. Explicit `?format=dsl` / `?format=yaml` still
      take precedence. (2026-04-03)
      _File: `handler/model.go`_ (#backend)

---

## Execution Order

```
Phases 10–14A               ─── DONE
    │
    ├── ARCHITECTURE_EVOLUTION.md ─── APPROVED
    ├── FRONTEND_EVOLUTION.md     ─── APPROVED
    │
Phase 14B (History + Multi)    ─── DONE (backend)
    │
Phase 14C (Model List + History UI) ─── DONE (frontend)
    │
Phase 14D (Data Lifecycle)         ─── PG purge + test isolation
    │                                    (must complete before Phase 15)
    │
Phase F (Frontend Restructure) ─── no backend dependency
    │   FA (View Regrouping + Tabs)
    │   FB (Interaction Consistency)
    │   FC (New Analyzer Views)     ← backend endpoints + frontend tabs
    │
Phase 15A (Auth)               ─── backend + frontend (login, auth context)
    │
Phase 15B (Orgs + Workspaces)  ─── backend + frontend (platform chrome,
    │                                breadcrumb, workspace dashboard, settings,
    │                                URL migration, sidebar context switching)
    │
Phase 15C (Authorization)      ─── backend only (role checks + permissions)
    │
Phase 16 (Collaboration)      ─── backend + frontend (comments, review,
    │                                changeset list, approve/reject, API keys)
    │
Phase 17 (Hardening)          ─── ongoing, export + calibration + completeness
    │
Phase 18 (Ecosystem)          ─── git integration, onboarding, import, notifications
```

**Architecture documents:**
- `docs/ARCHITECTURE_EVOLUTION.md` — backend: data hierarchy, tenancy,
  auth, API routes (APPROVED, all 6 decisions resolved)
- `docs/FRONTEND_EVOLUTION.md` — frontend: navigation layers, tab
  structure, new views, platform chrome (APPROVED)

Each phase now contains **both** backend and frontend work. No separate
frontend section at the end — the AI engineer picks up all items in a
phase together, backend APIs first, frontend consuming them after.

Phase F (frontend restructure) is the exception: it has zero backend
dependency and can start immediately, in parallel with any other work.
