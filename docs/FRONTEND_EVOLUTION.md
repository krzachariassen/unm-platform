# UNM Platform — Frontend Evolution

_Design document for the frontend architecture as the platform evolves from
a single-user explorer to a multi-tenant collaborative tool. This is the
frontend counterpart to `docs/ARCHITECTURE_EVOLUTION.md`._

**Status:** Approved
**Owner:** Kristian Zachariassen
**Created:** 2026-04-03

---

## 1. Why This Document Exists

The frontend was built as a single-model, single-user explorer with 12
standalone pages. Three forces require a structural redesign:

1. **Navigation doesn't scale.** 7 items under "Views" with overlapping
   data. Users can't predict what clicking something will do — four
   different interaction patterns exist across pages.
2. **Hidden backend data.** Five analyzers produce insights no frontend
   view consumes (gap analysis, dependency graphs, interaction diversity,
   data coupling, service complexity).
3. **New platform layers.** The multi-tenant hierarchy (Organization →
   Workspace → Model → Version) from `ARCHITECTURE_EVOLUTION.md` needs
   entirely new pages: login, org selector, workspace dashboard, model
   list, version history, settings.

These are not independent features. The tab restructure, new views, and
platform chrome must be designed together.

---

## 2. Two Navigation Layers

The app serves two contexts that the current sidebar conflates:

### Platform Layer (new)

Manages the multi-tenant hierarchy. Users land here on login and navigate
to a specific model.

```
Login → Org Selector → Workspace Dashboard → Model List → [enter model]
```

Pages:
- **Login** — Google OAuth sign-in
- **Org Selector** — pick an org (shown only if user has multiple orgs)
- **Workspace Dashboard** — models, members, settings, import actions.
  Replaces current Upload page as the "home"
- **Model List** — card grid per workspace: name, last updated, version
  count, health score summary
- **Settings** — org settings (members, AI key) and workspace settings
  (members, API keys, visibility)

### Model Layer (redesigned)

Once inside a model, the sidebar shows grouped pages with horizontal
tabs. This is where architecture exploration and editing happens.

```
┌──────────────────────────────────────────────────┐
│ TopBar: [Org] > [Workspace] > [Model]    [U] [?]│
├──────┬───────────────────────────────────────────┤
│      │  [Tab1]  [Tab2]  [Tab3]                   │
│ Side │───────────────────────────────────────────│
│ bar  │                                           │
│      │           Tab Content                     │
│      │                                           │
└──────┴───────────────────────────────────────────┘
```

---

## 3. Sidebar Structure (Model Layer)

```
Overview
  Dashboard

Architecture
  UNM Map
  Needs              tabs: [Overview | Traceability | Gaps]
  Capabilities       tabs: [Hierarchy | Services | Dependencies]
  Teams              tabs: [Topology | Ownership | Cognitive Load | Interactions]
  Signals

Changes
  What-If            tabs: [AI Scenarios | Manual]
  Changesets         tabs: [Active | History]

AI
  Recommendations
  Advisor
```

Sidebar items: **8** (down from 12). All architecture views grouped
into 3 entries with tabs. The sidebar adapts based on context:
- At workspace level → shows workspace nav (models, members, settings)
- At model level → shows the structure above

---

## 4. Tab Inventory — What Goes Where

### 4.1 Needs (3 tabs)

| Tab | Source | Content | Status |
|-----|--------|---------|--------|
| **Overview** | `NeedView.tsx` | Actor-grouped need cards, capability badges, risk flags, AI insights | Exists |
| **Traceability** | `RealizationView.tsx` Value Chain tab | End-to-end Need → Capability → Service → Team chain | Move from Realization |
| **Gaps** | `GapAnalyzer` output | Unmapped needs, unrealized capabilities, unowned services, unneeded capabilities | **New view** |

**Gaps data available in backend** (`GapReport`): `unmapped_needs[]`,
`unrealized_capabilities[]`, `unowned_services[]`, `unneeded_capabilities[]`,
`orphan_services[]` (orphans currently omitted from HTTP builder — fix needed).
Currently only accessible via `POST /analyze/gaps` — no view endpoint.

**Backend work needed:**
- New view endpoint: `GET /api/models/{id}/views/gaps`
- Expose `orphan_services` in gap analysis HTTP builder

### 4.2 Capabilities (3 tabs)

| Tab | Source | Content | Status |
|-----|--------|---------|--------|
| **Hierarchy** | `CapabilityView.tsx` | Parent/child tree, visibility bands, detail panel, QuickAction edits | Exists |
| **Services** | `RealizationView.tsx` Service Table tab | Service-first table: capabilities, ext deps, team ownership | Move from Realization |
| **Dependencies** | `DependencyAnalyzer` output | Service dependency graph, cycle detection, critical path | **New view** |

**Dependencies data available in backend** (`DependencyReport`): `cycles[]`,
`max_depths`, `critical_path`. Plus `Service.DependsOn[]` in domain model.
Not rendered in any view today.

**Backend work needed:**
- New view endpoint: `GET /api/models/{id}/views/dependencies`
- Include `Service.DependsOn` graph data in response

### 4.3 Teams (4 tabs)

| Tab | Source | Content | Status |
|-----|--------|---------|--------|
| **Topology** | `TeamTopologyView.tsx` | Graph/table of teams, interaction lines, detail panel | Exists |
| **Ownership** | `OwnershipView.tsx` | Team lanes with capabilities and services | Exists (needs interaction fix) |
| **Cognitive Load** | `CognitiveLoadView.tsx` | Dimensional cognitive load per team | Move |
| **Interactions** | `InteractionDiversityAnalyzer` output | Mode distribution, isolated teams, over-reliant teams | **New view** |

**Interactions data available in backend** (`InteractionDiversityReport`):
`mode_distribution`, `isolated_teams[]`, `over_reliant_teams[]`,
`all_modes_same`. Analyzed but never shown to users.

**Backend work needed:**
- New view endpoint: `GET /api/models/{id}/views/interactions`

### 4.4 Changes (2 tabs)

| Tab | Source | Content | Status |
|-----|--------|---------|--------|
| **Active** | Phase 16 backlog | Changeset list with review status, comments, author | **New** (Phase 16) |
| **History** | Phase 14B backlog | Model version timeline with inline diff viewer | **New** (Phase 14B) |

What-If stays as a separate sidebar item — it's the editing workspace, not
a change tracking view.

---

## 5. Interaction Pattern Standardization

The current frontend has four different patterns for click behavior. The
target is two:

### Detail Panel pattern (editable views)

Click an entity → `SlidePanel` opens on the right with full entity
details, AI insight, anti-patterns, and QuickAction edit buttons.

Used by: UNM Map, Capability, Team Topology.

**Must adopt:** Ownership (currently uses a custom fixed-position popover
for services — replace with SlidePanel).

### Expand/collapse pattern (analytical views)

Click a card → expands inline to show more detail.

Used by: Need View, Cognitive Load, Signals, Realization.

Appropriate for read-focused diagnostic views.

### What to fix

| Issue | Current | Target |
|-------|---------|--------|
| Ownership service popover | Custom `getBoundingClientRect()` fixed div + backdrop | `SlidePanel` (same as Capability) |
| Ownership team detail | `AntiPatternPanel` (anti-patterns only) | Rich entity panel (metadata, services, capabilities, interactions, AI insight) |
| Team Topology pencil visibility | `opacity: 0.35` default, easy to miss | `opacity: 0.7+` default |
| Team Topology table edit | No QuickAction on table rows | Add QuickAction to table rows |
| AI insights in Ownership | Not present | Add AI insight section to team/service panels |
| Cross-view navigation | Only Realization → Capability link exists | Entity names clickable to navigate to relevant view |

---

## 6. Platform Chrome — Multi-Tenant UI

### 6.1 TopBar evolution

Current: `[UNM Platform]  [Search]`

Target:

```
[Org Logo] Acme Eng  >  Platform Team  >  nexus-model   [Search] [3 pending] [KZ ▼]
           ^org          ^workspace       ^model          ^global  ^changeset ^user
```

Each breadcrumb segment is clickable. Click org → org selector. Click
workspace → model list. Click model → model dashboard. User avatar →
settings, logout.

### 6.2 URL structure

Current: `/dashboard`, `/need`, `/ownership`

Target:

```
/login
/orgs                                              # org selector
/:orgSlug                                          # workspace dashboard
/:orgSlug/:wsSlug                                  # model list
/:orgSlug/:wsSlug/models/:id                       # model dashboard
/:orgSlug/:wsSlug/models/:id/needs?tab=overview    # grouped page with tab
/:orgSlug/:wsSlug/models/:id/teams?tab=topology
/:orgSlug/:wsSlug/models/:id/map
/:orgSlug/:wsSlug/settings                         # workspace settings
/settings/org                                      # org settings
```

### 6.3 Local dev mode

When `auth.enabled: false`, the platform auto-creates org "local" and
workspace "default". URLs simplify:

```
/                         → redirect to /local/default
/local/default            → model list (workspace dashboard)
/local/default/models/:id → model views
```

No login page shown. No org selector. Default user injected by middleware.

---

## 7. Shared Components

### TabBar

```
components/ui/tab-bar.tsx
```

Horizontal tab strip used by every grouped page. Syncs with URL `?tab=`
param for deep linking. Preserves filter/search state across tab switches
within the same page.

### EntityDetailPanel

```
components/detail/EntityDetailPanel.tsx
```

Replaces the fragmented panels (`AntiPatternPanel`, capability
`DetailPanel`, team-topology `DetailPanel`). Adapts content based on
entity type (team, service, capability, need). All pages use the same
panel for the same entity type.

### Breadcrumb

```
components/layout/Breadcrumb.tsx
```

Org → Workspace → Model navigation. Rendered in TopBar. Drives the
platform-level navigation.

---

## 8. Implementation Phases

These phases are designed to interleave with the backend phases in the
backlog. Each can be worked on independently when its backend dependency
is met.

### Phase F1 — View Regrouping & Tabs (no backend dependency)

Frontend-only. Move existing view components into grouped pages with
horizontal tabs.

1. Create `TabBar` and `TabbedPage` shared components
2. Create `NeedsPage` with tabs: Overview (NeedView), Traceability
   (RealizationView Value Chain)
3. Create `CapabilitiesPage` with tabs: Hierarchy (CapabilityView),
   Services (RealizationView Service Table)
4. Create `TeamsPage` with tabs: Topology (TeamTopologyView), Ownership
   (OwnershipView), Cognitive Load (CognitiveLoadView)
5. Create `ChangesPage` with tabs: What-If (WhatIfPage)
6. Update sidebar and router
7. Delete standalone RealizationView and CognitiveLoadView route entries

**Can start immediately** — no backend changes required.

### Phase F2 — Interaction Consistency (no backend dependency)

Frontend-only. Fix the interaction pattern inconsistencies.

1. Replace Ownership service popover with `SlidePanel`
2. Create unified `EntityDetailPanel`
3. Standardize QuickAction pencil visibility
4. Add QuickAction to Team Topology table view
5. Add AI insights to Ownership detail panel
6. Add cross-view entity navigation links

**Can start immediately** — no backend changes required.

### Phase F3 — New Analyzer Views (backend + frontend)

Expose hidden analyzer data. Requires new backend view endpoints.

1. Backend: `GET /views/gaps` endpoint (wrap `GapAnalyzer` output,
   include orphan services)
2. Backend: `GET /views/dependencies` endpoint (wrap `DependencyAnalyzer`,
   include `Service.DependsOn` graph)
3. Backend: `GET /views/interactions` endpoint (wrap
   `InteractionDiversityAnalyzer` output)
4. Frontend: Gaps tab in NeedsPage
5. Frontend: Dependencies tab in CapabilitiesPage (graph visualization)
6. Frontend: Interactions tab in TeamsPage

**Can start after** Phase F1 (tabs exist to put the new views in).

### Phase F4 — Model List & History UI (depends on Phase 14B)

Frontend for multi-model and version history.

1. Model list page (card grid, health score per model)
2. Version history tab in ChangesPage
3. Inline diff viewer component
4. Model upload becomes "Add model" in workspace context

**Depends on** Phase 14B backend APIs (model list, history, diff).

### Phase F5 — Platform Chrome (depends on Phase 15)

Full multi-tenant UI. Requires auth and org/workspace APIs.

1. Login page
2. Auth context (session check, 401 redirect)
3. Breadcrumb component (org > workspace > model)
4. TopBar redesign with breadcrumb and user avatar
5. Org selector page
6. Workspace dashboard (replaces Upload as home)
7. Settings pages (org settings, workspace settings)
8. URL structure migration (`/:orgSlug/:wsSlug/models/:id/...`)
9. API client update with org/workspace path scoping
10. Protected route wrapper

**Depends on** Phase 15A (auth), 15B (org/workspace APIs).

### Phase F6 — Collaboration UI (depends on Phase 16)

1. Changeset list with review status badges
2. Comment thread on changeset review dialog
3. Author attribution on changesets and versions
4. Approve/reject actions on changeset dialog

**Depends on** Phase 16 backend APIs.

---

## 9. Mapping to Backlog Phases

| Frontend Phase | Backend Dependency | When to Execute |
|---------------|-------------------|-----------------|
| **F1** — View regrouping & tabs | None | Anytime (can start now) |
| **F2** — Interaction consistency | None | Anytime (can start now) |
| **F3** — New analyzer views | New view endpoints (gap, dependency, interaction) | After F1 |
| **F4** — Model list & history UI | Phase 14B (model list, history, diff APIs) | After 14B backend |
| **F5** — Platform chrome | Phase 15A + 15B (auth, org, workspace APIs) | After 15 backend |
| **F6** — Collaboration UI | Phase 16 (changeset comments, review, API keys) | After 16 backend |

F1 and F2 have zero backend dependency and can run in parallel with the
current Phase 14B backend work.

---

## 10. What This Document Does NOT Cover

- Backend API design — see `docs/ARCHITECTURE_EVOLUTION.md`
- Real-time collaborative editing — explicitly deferred
- Mobile/responsive layout — not a priority for v1
- Design system / component library — use shadcn/ui + Tailwind as today
- Performance optimization — address when it becomes a problem

---

## 11. Related Documents

| Document | Purpose |
|----------|---------|
| `docs/ARCHITECTURE_EVOLUTION.md` | Backend architecture: data hierarchy, tenancy, auth, API routes |
| `docs/BACKLOG.md` | Engineering task list with phased execution |
| `docs/PRODUCT_ROADMAP.md` | User-facing milestones and capabilities |
