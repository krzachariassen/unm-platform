# UNM Platform — Full UI/UX & Architecture Review

**Date:** 2026-03-17  
**Scope:** All frontend pages, layout, navigation, components, styling, API layer, state management, data flow, and backend API design  
**Goal:** Identify inconsistencies and propose a unified, minimalist design system with a clean architecture  
**Approach:** Big-bang rewrite on branch `refactor/ui-unification` (no users yet — safe to rewrite in place)

---

## Executive Summary

The frontend has grown organically and now contains **14 distinct routes** rendered by **13 components** totalling **~8,200 lines of TSX**. While each page individually looks polished, the overall experience feels fragmented — as if built by different people at different times (because it was). This review identifies concrete problems and proposes a unified design direction.

**Key findings:**

1. No shared page layout — every page reinvents its own header, stat cards, and spacing
2. Navigation is sidebar-only with 14 flat items — too many for a flat list, no grouping
3. Style tokens are redefined per file (5+ copies of `gradientTitle`, `cardShell`, `CARD_SHELL`)
4. Mixed styling: inline `style={{}}` objects vs Tailwind classes vs CSS variables — often in the same component
5. Giant monolith components (5 files > 800 lines, 3 files > 1,000 lines)
6. No horizontal/tab navigation pattern despite several pages being logically grouped

---

## 1. Navigation & Information Architecture

### Current state

The sidebar (`Sidebar.tsx`) renders a flat list of 14 items:

```
Upload  |  Dashboard  |  Signals  |  UNM Map  |  Need View  |  Capability View
Ownership View  |  Team Topology  |  Cognitive Load  |  Realization View
Edit Model  |  What-If Explorer  |  AI Recommendations  |  AI Advisor
```

**Problems:**

- **No grouping.** 14 flat items forces users to scan everything. There are clear groups: (a) model management, (b) analysis views, (c) team/org views, (d) AI tools.
- **"Edit Model"** is a dead route that just redirects to `/unm-map`. It wastes a sidebar slot.
- **No breadcrumbs or context.** Once inside a view, there is zero indication of where you are in the conceptual hierarchy.
- **Sidebar width (w-56 = 224px)** consumes significant horizontal space on every page, even though the user only uses it to navigate occasionally.

### Recommendation

Adopt a two-tier navigation:

| Layer | Purpose | Implementation |
|-------|---------|----------------|
| **Left sidebar** (collapsible, ~56px icons-only by default) | Top-level sections: Home, Explore, Teams, AI | Slim icon rail, expands on hover or toggle |
| **Horizontal tabs** (per section) | Sub-views within a section | Persistent tab bar below the TopBar |

Proposed grouping:

```
HOME
  ├─ Upload / Dashboard

EXPLORE (horizontal tabs when active)
  ├─ UNM Map │ Needs │ Capabilities │ Realization │ Signals

TEAMS (horizontal tabs when active)
  ├─ Ownership │ Team Topology │ Cognitive Load

AI TOOLS (horizontal tabs when active)
  ├─ What-If │ Recommendations │ Advisor

EDIT (integrated into views, not a separate page)
```

This reduces sidebar items from 14 → 4, and each section gets at most 5 tabs.

---

## 2. Page Layout Inconsistency

### Current state

Every page constructs its own layout from scratch. There is no shared `PageLayout`, `PageHeader`, or `PageContent` component.

| Page | Title style | Content width | Has stat cards | Stat card implementation |
|------|-------------|---------------|----------------|------------------------|
| DashboardPage | Gradient h1, inline style | `maxWidth: 1200` | Yes | Inline div with `style={{}}` |
| UploadPage | Tailwind `text-2xl font-bold` | `max-w-2xl mx-auto` | No | — |
| NeedView | Inline `gradientTitle` const | No max-width | Yes | Custom `StatCard` component |
| CapabilityView | Inline `H1_GRADIENT` const | No max-width | Yes | Different `StatCard` |
| OwnershipView | Inline `H1_GRADIENT` const | No max-width | No | — |
| SignalsView | Inline `gradientTitle` const | No max-width | Yes | Inline styled divs |
| TeamTopologyView | `text-2xl font-bold tracking-tight` (Tailwind) | No max-width | Yes | Yet another `StatCard` |
| CognitiveLoadView | Inline gradient style | No max-width | Yes | Inline styled divs |
| RealizationView | `text-2xl font-bold` (Tailwind) | No max-width | No | — |
| WhatIfPage | `text-2xl font-semibold tracking-tight` | `max-w-6xl` | No | — |
| AdvisorPage | `text-lg font-semibold tracking-tight` | `max-w-3xl` | No | — |
| RecommendationsPage | Inline `gradientH1Style` const | `max-w-4xl` | No | — |

**Problems:**

- **5 different ways to render a page title.** Some use the gradient text effect, some don't. The gradient is copy-pasted across 6 files.
- **Content width is inconsistent.** Some pages constrain width (`max-w-2xl`, `max-w-3xl`, `max-w-4xl`, `max-w-6xl`, `maxWidth: 1200`), some go full-width.
- **Stat cards are reimplemented from scratch** in NeedView, CapabilityView, DashboardPage, CognitiveLoadView, and SignalsView — each with different padding, radius, gradients, and font sizes.

### Recommendation

Create a `PageLayout` component:

```tsx
<PageLayout
  title="Need View"
  subtitle="Actor → Need → Capability delivery chains"
  stats={[...]}         // optional stat cards
  tabs={[...]}          // optional horizontal tab bar
  maxWidth="xl"         // 'sm' | 'md' | 'lg' | 'xl' | 'full'
>
  {children}
</PageLayout>
```

This single component would replace ~200 duplicated lines across the codebase.

---

## 3. Styling Chaos

### Current state

Three styling systems are used simultaneously, often in the same file:

| System | Example | Files using it |
|--------|---------|---------------|
| **Inline `style={{}}`** | `style={{ fontSize: 30, fontWeight: 800, ... }}` | All 13 page files |
| **Tailwind classes** | `className="text-2xl font-bold"` | All 13 page files |
| **CSS custom properties** | `hsl(var(--color-border))` | `index.css`, `button.tsx`, `card.tsx` |

**Most egregious examples:**

- `Sidebar.tsx` uses inline `onMouseEnter`/`onMouseLeave` handlers to manually set `background` and `color` on hover, instead of Tailwind's `hover:` utilities.
- The gradient title style is copy-pasted as a const in **6 files** (`gradientTitle`, `gradientH1Style`, `H1_GRADIENT`) — same values each time.
- `DashboardPage.tsx` (562 lines) has **zero** Tailwind classes in its entire return JSX — everything is inline `style={{}}`.
- `card.tsx` and `button.tsx` use the design token system (`cn()`, `cva()`, CSS variables), but almost no page actually uses these components.

**Duplicated style constants across files:**

| Constant name | Files it appears in | Values |
|---------------|---------------------|--------|
| `gradientTitle` / `H1_GRADIENT` / `gradientH1Style` | NeedView, CapabilityView, OwnershipView, SignalsView, CognitiveLoadView, RecommendationsPage | All identical: `fontSize: 30, fontWeight: 800, background: linear-gradient(...)` |
| `cardShell` / `CARD_SHELL` / `reportCardStyle` | NeedView, CapabilityView, OwnershipView, SignalsView, DashboardPage, RecommendationsPage | All identical: `borderRadius: 20, background: linear-gradient(...)` |
| `SECTION_LABEL` | CapabilityView, OwnershipView | Identical uppercase label style |
| `SUBTITLE` / `subtitleStyle` | CapabilityView, OwnershipView, RecommendationsPage | Identical: `fontSize: 14, color: '#64748b'` |
| `TEAM_TYPE_BADGE` | team-type-styles.ts + redefined in CognitiveLoadView | Same colors, different names |

### Recommendation

1. **Pick one styling system — Tailwind — and commit to it.** Remove all inline `style={{}}` except for truly dynamic values (positions, calculated dimensions in the map view).
2. **Extract shared design tokens** into a single file (`lib/design-tokens.ts` or Tailwind config):
   - Page title: `text-[30px] font-extrabold tracking-tight bg-gradient-to-br from-slate-800 to-slate-500 bg-clip-text text-transparent`
   - Card shell: `rounded-[20px] bg-gradient-to-br from-white to-slate-50 border border-slate-200 shadow-sm`
   - Section label: `text-[11px] font-semibold text-slate-500 uppercase tracking-wider`
3. **Create a shared `StatCard` component** used by all views.

---

## 4. Component Architecture

### Current state — file sizes

| File | Lines | Notes |
|------|-------|-------|
| UNMMapView.tsx | 1,121 | Custom SVG rendering, full layout engine |
| CapabilityView.tsx | 1,045 | Multiple sub-sections, inline components |
| OwnershipView.tsx | 1,025 | Matrix grid, popovers, all inline |
| TeamTopologyView.tsx | 989 | Complex card grid, modal-like details |
| SignalsView.tsx | 857 | 5 signal sections, collapsible cards |
| RealizationView.tsx | 598 | Two-tab view, tables |
| DashboardPage.tsx | 562 | Stats, health rings, signal bars, view cards |
| CognitiveLoadView.tsx | 513 | Speedometer SVG, team cards |
| UploadPage.tsx | 388 | Upload flow, step tracking |
| NeedView.tsx | 375 | Actor-need chains |
| WhatIfPage.tsx | 322 | Two tabs (AI + Manual), changeset forms |
| RecommendationsPage.tsx | 315 | AI-generated report display |
| AdvisorPage.tsx | 153 | Chat interface |

**Problems:**

- **5 files exceed 800 lines** — each is a monolith containing layout, data fetching, state management, business logic, and presentation all mixed together.
- **Inline sub-components.** Most files define helper components (`StatCard`, `HealthRing`, `AnimatedNumber`, `SignalBar`, `Speedometer`) inside the same file instead of extracting them.
- **No data-fetching pattern.** Some views use the custom `useModelView` hook, others use raw `useEffect` + `useState` + `api.getX()`. Two different patterns in the same codebase:

```
// Pattern A (NeedView, CapabilityView, TeamTopologyView, CognitiveLoadView)
const { data, loading, error } = useModelView('view-name', api.getXxxView)

// Pattern B (RealizationView, OwnershipView, SignalsView, DashboardPage)
const [data, setData] = useState(null)
const [loading, setLoading] = useState(true)
useEffect(() => { api.getX(modelId).then(setData)... }, [modelId])
```

### Recommendation

1. **Extract reusable components** from page files into `components/`:
   - `StatCard` — used in 5+ views
   - `HealthRing` / `Speedometer` — used in dashboard and cognitive load
   - `FilterBar` / `SortBar` — most views have some sort/filter controls
   - `CollapsibleSection` — at least 4 views have expand/collapse patterns
   - `InsightsBanner` — the AI insights pattern appears in 6 views
2. **Standardize data fetching** — use `useModelView` everywhere, remove manual `useEffect` patterns.
3. **Break up monolith pages** — each view should be ≤300 lines. Sub-sections become child components.

---

## 5. Specific Page Issues

### 5.1 Sidebar

- **Hover handling via `onMouseEnter`/`onMouseLeave`** manually sets `style.background` and `style.color` — should use Tailwind `hover:bg-gray-100 hover:text-gray-900`.
- **Hardcoded colors** (`#f9fafb`, `#e5e7eb`, `#111827`, `#9ca3af`, `#6b7280`) — should use Tailwind palette classes.
- **No collapse/expand** — always takes 224px.

### 5.2 TopBar

- Good: clean, minimal
- Issue: The edit-mode button and search are the only items. The bar is underutilized and could house horizontal tab navigation.
- Issue: Hardcoded colors instead of Tailwind classes.

### 5.3 DashboardPage (562 lines)

- **100% inline styles, 0 Tailwind.** This is the most visually polished page but the hardest to maintain.
- `HealthRing`, `AnimatedNumber`, `SignalBar` are all defined inline — should be extracted.
- `VIEW_CARDS` navigation cards duplicate the sidebar's route list.

### 5.4 UNMMapView (1,121 lines)

- Implements a full SVG-based layout engine with manual coordinate calculation. This is inherently complex and justified.
- However, the node rendering, edge rendering, and tooltip logic should each be separate files.
- The `useChangeset` integration adds edit-mode complexity that could be a separate layer.

### 5.5 WhatIfPage (322 lines)

- Has its own tab bar implementation (`AI Scenarios | Manual Mode`) — this is exactly the pattern that should be standardized.
- The `AIWhatIfTab` and `ManualWhatIfTab` are defined in the same file but are fully independent — they should be separate files.

### 5.6 RealizationView (598 lines)

- Has its own tab bar (`chain | service`) — another tab implementation to standardize.
- Uses `useEffect` + `Promise.all` for data fetching instead of `useModelView`.

### 5.7 SignalsView (857 lines)

- Has complex filtering (URL-based via `useSearchParams`) that's well-implemented but unique to this page.
- Defines risk styling objects (`RISK`) that partially overlap with styles in DashboardPage.

---

## 6. Design Direction: Minimalist Horizontal-First

Based on the analysis, here is the proposed unified design system:

### 6.1 Layout structure

```
┌──────────────────────────────────────────────────────────┐
│ [≡]  UNM Platform    [Search...]    [Edit] [User]       │  ← TopBar (h-14)
├──┬───────────────────────────────────────────────────────┤
│  │  Explore │ Needs │ Capabilities │ Realization │ Signals│  ← Section tabs
│  ├───────────────────────────────────────────────────────┤
│  │                                                       │
│  │                  Page Content                         │
│H │                                                       │
│o │                                                       │
│m │                                                       │
│e │                                                       │
│  │                                                       │
│  │                                                       │
│  │                                                       │
│  │                                                       │
├──┴───────────────────────────────────────────────────────┤
│ [PendingChangesBar — when edit mode is active]           │
└──────────────────────────────────────────────────────────┘
```

- **Sidebar:** Collapsed by default to a 56px icon rail with 4 section icons. Expands on hover.
- **Horizontal tabs:** Each section has persistent tab navigation below the TopBar.
- **Content area:** Gets maximum horizontal space.

### 6.2 Component hierarchy

```
App
└─ AppShell
   ├─ Sidebar (icon rail, 56px collapsed / 224px expanded)
   ├─ TopBar (search, edit mode toggle, model status)
   ├─ SectionTabs (horizontal tabs per active section)
   ├─ PageLayout
   │   ├─ PageHeader (title, subtitle, stat cards)
   │   └─ PageContent (scrollable)
   └─ PendingChangesBar (sticky bottom, edit mode only)
```

### 6.3 Shared components to extract

| Component | Replaces | Used by |
|-----------|----------|---------|
| `PageHeader` | 13 different title/subtitle patterns | All pages |
| `StatCard` | 5 different stat card implementations | Dashboard, NeedView, CapabilityView, CognitiveLoadView, SignalsView |
| `TabBar` | 3 inline tab implementations | WhatIfPage, RealizationView, section navigation |
| `FilterBar` | Various inline filter buttons | SignalsView, RealizationView, CognitiveLoadView, OwnershipView |
| `CollapsibleCard` | Various expand/collapse patterns | NeedView, SignalsView, CognitiveLoadView, TeamTopologyView |
| `InsightBanner` | Inline AI insight rendering | NeedView, CapabilityView, OwnershipView, TeamTopologyView, CognitiveLoadView, SignalsView |
| `RiskBadge` | Various risk/level badge styles | SignalsView, CognitiveLoadView, DashboardPage |
| `EmptyState` | Various "no data" placeholders | AdvisorPage, WhatIfPage, SignalsView |
| `DataTable` | Various inline table renderings | OwnershipView, RealizationView |

### 6.4 Design tokens

Standardize on a single set of tokens (in Tailwind config or a shared module):

```
Colors:
  text-primary:     #111827 (slate-900)
  text-secondary:   #64748b (slate-500)
  text-muted:       #94a3b8 (slate-400)
  bg-card:          white → slate-50 gradient
  bg-surface:       #f8fafc (slate-50)
  border-default:   #e2e8f0 (slate-200)
  accent-blue:      #2563eb
  accent-violet:    #7c3aed

Spacing:
  page-padding:     24px (p-6)
  card-padding:     24px (p-6)
  card-radius:      16px (rounded-2xl) — standardize from the current mix of 8/12/14/16/20px
  gap-cards:        16px (gap-4)

Typography:
  page-title:       text-2xl font-bold tracking-tight (drop the gradient effect for cleanliness)
  section-title:    text-base font-semibold
  card-title:       text-sm font-semibold
  body:             text-sm
  caption:          text-xs text-muted-foreground
  stat-number:      text-2xl font-extrabold tabular-nums
```

---

## 7. Implementation Plan — Big-Bang Rewrite

> **Context:** No users are on the platform yet. We are doing this as a clean rewrite
> on a dedicated branch (`refactor/ui-unification`), not an incremental migration.
> All changes land in one shot.

### Step 1: Shared component library

Create the foundation components that every page will use. Build these first, in isolation,
before touching any existing page.

**New files to create:**

| File | Purpose |
|------|---------|
| `components/layout/SectionTabs.tsx` | Horizontal tab bar — renders below TopBar, driven by route |
| `components/ui/PageHeader.tsx` | Unified page title + subtitle + optional stat cards row |
| `components/ui/StatCard.tsx` | Single stat card (value, label, icon, color) |
| `components/ui/TabBar.tsx` | Generic horizontal tab selector (used by WhatIfPage, RealizationView, etc.) |
| `components/ui/FilterBar.tsx` | Row of filter/sort chips with active state |
| `components/ui/CollapsibleCard.tsx` | Card with expand/collapse toggle |
| `components/ui/InsightBanner.tsx` | AI insight callout (tip icon + markdown content) |
| `components/ui/RiskBadge.tsx` | Risk/level badge (low/medium/high with consistent colors) |
| `components/ui/EmptyState.tsx` | Empty-state placeholder (icon + message + optional CTA) |
| `components/ui/DataTable.tsx` | Simple table with header styling and optional sort |
| `components/ui/HealthRing.tsx` | SVG ring chart (extracted from DashboardPage) |
| `components/ui/Speedometer.tsx` | SVG gauge (extracted from CognitiveLoadView) |

**Rules for all new components:**
- Tailwind classes only — zero inline `style={{}}` except for truly dynamic computed values
- Props interface exported alongside the component
- No API calls — receive data as props
- Must handle missing/null data gracefully

### Step 2: Rewrite layout shell

Rewrite `AppShell.tsx`, `Sidebar.tsx`, and `TopBar.tsx` together as one unit.

**Sidebar changes:**
- Default state: collapsed 56px icon rail showing 4 section icons (Home, Explore, Teams, AI)
- Hover or click: expands to ~200px showing labels
- Remove the flat 14-item list entirely
- Remove dead "Edit Model" entry

**TopBar changes:**
- Keep: model name + validation badge + search + edit-mode toggle
- Add: slot for `SectionTabs` component (renders the horizontal tabs for the active section)
- Convert all hardcoded hex colors to Tailwind classes
- Remove `onMouseEnter`/`onMouseLeave` manual style manipulation

**SectionTabs (new):**
- Reads current route to determine active section and active tab
- Renders horizontal tabs for the section
- Tab definitions:

```
SECTION: Home (no tabs — just Upload and Dashboard)
  /           → Upload
  /dashboard  → Dashboard

SECTION: Explore
  /unm-map      → UNM Map
  /need         → Needs
  /capability   → Capabilities
  /realization  → Realization
  /signals      → Signals

SECTION: Teams
  /ownership      → Ownership
  /team-topology  → Team Topology
  /cognitive-load → Cognitive Load

SECTION: AI
  /what-if          → What-If
  /recommendations  → Recommendations
  /advisor          → Advisor
```

**App.tsx changes:**
- Remove legacy `/views/*` redirect routes
- Remove `/edit` redirect route
- Keep all actual routes unchanged (paths stay the same)

### Step 3: Rewrite every page to use shared components

Rewrite all 13 page/view files. For each file:

1. Replace the custom title/header block with `<PageHeader>`
2. Replace custom stat cards with `<StatCard>` from shared library
3. Replace custom tab bars with `<TabBar>`
4. Replace inline expand/collapse with `<CollapsibleCard>`
5. Replace inline AI insight rendering with `<InsightBanner>`
6. Convert all inline `style={{}}` to Tailwind classes
7. Migrate data fetching to TanStack Query (`useQuery`) — no manual `useEffect` patterns
8. Extract any remaining large inline sub-components into separate files
9. Delete all duplicated style constants (`gradientTitle`, `cardShell`, `H1_GRADIENT`, etc.)

**Per-page specifics:**

| Page | Key changes |
|------|-------------|
| **DashboardPage** | Full rewrite — currently 0% Tailwind. Extract `HealthRing`, `AnimatedNumber`, `SignalBar`. Remove `VIEW_CARDS` duplication. Use TanStack Query for data fetching. |
| **NeedView** | Replace inline `StatCard`, `gradientTitle`, `cardShell`, `pill`. Use shared `InsightBanner`. Migrate `useModelView` to `useQuery`. |
| **CapabilityView** | Replace inline `StatCard`, `H1_GRADIENT`, `CARD_SHELL`, `SECTION_LABEL`, `SUBTITLE`. Extract band-grouped capability rendering into sub-component. Migrate to `useQuery`. |
| **OwnershipView** | Replace `H1_GRADIENT`, `CARD_SHELL`, `SECTION_LABEL`, `SUBTITLE`. Extract matrix grid into sub-component. Migrate to `useQuery`. Request backend to include parent-group data in ownership endpoint. |
| **TeamTopologyView** | Replace inline `TEAM_TYPES` styling with shared tokens. Extract zone-grouped team cards into sub-component. Migrate to `useQuery`. |
| **CognitiveLoadView** | Extract `Speedometer` to shared component. Replace inline `TEAM_TYPE_BADGE`, `LEVEL` configs. Migrate to `useQuery`. |
| **SignalsView** | Replace `gradientTitle`, `cardShell`, `RISK` config. Migrate to `useQuery`. Keep the `useSearchParams` filtering. |
| **RealizationView** | Replace inline tab bar with shared `TabBar`. Migrate to `useQuery`. Request backend to include need chains in realization endpoint. |
| **UNMMapView** | **Full React Flow rewrite** (see Step 4 in Section 8). Do NOT keep the hand-rolled SVG engine. |
| **WhatIfPage** | Replace inline tab bar with shared `TabBar`. Move tabs to `features/whatif/`. |
| **AdvisorPage** | Smallest page — straightforward Tailwind conversion + `useQuery`. |
| **RecommendationsPage** | Replace `gradientH1Style`, `reportCardStyle`. Tailwind + `useQuery`. |
| **UploadPage** | Already mostly Tailwind — minor cleanup of remaining inline styles. |

### Step 4: Delete dead code and validate

1. Delete `EditModelPage.tsx` (dead redirect — route removed in step 2)
2. Delete all duplicate style constants from `lib/` that are now unused
3. Verify no file imports deleted constants
4. Run `cd frontend && npm run build` — must pass with zero errors
5. Manual smoke test: Upload → Dashboard → each view → AI tools

### Step 5: Update agent docs

Update the following files to reflect the new architecture:

| File | Changes |
|------|---------|
| `.claude/agents/frontend-engineer/AGENT.md` | Update "Standard View Page Pattern" to show `PageHeader` + `useModelView` + shared components. Update component checklists. |
| `.claude/agents/frontend-engineer/anti-patterns.md` | Add anti-pattern for inline `style={{}}` when Tailwind equivalent exists. Add anti-pattern for reimplementing `StatCard`/`TabBar`/etc. |
| `.claude/agents/frontend-engineer/MEMORY.md` | Add entry documenting the UI unification rewrite. |
| `.claude/agents/common/architecture.md` | Update frontend file tree to show new `components/ui/` components and `layout/SectionTabs.tsx`. |

---

## 8. Engineer Handoff — Final Execution Plan

> **NOTE:** This is the authoritative execution plan. It supersedes the per-section
> recommendations in Parts 1–4. All architecture decisions (TanStack Query, React Flow,
> layered file structure, expanded shadcn/ui) are incorporated here.

**Branch:** `refactor/ui-unification` (create from `main`)

**Prerequisite:** The agent doc updates (`.claude/rules/frontend-architecture.md`,
updated `AGENT.md`, `anti-patterns.md`, `react-conventions.md`, `stack.md`) must be
committed BEFORE the engineer starts. These files define the rules the engineer follows.
See "Agent Doc Updates" at the end of this section.

**Strict ordering — each step depends on the previous:**

```
STEP 0  Architecture foundation (no visual changes)
        ├─ npm install @tanstack/react-query @tanstack/react-query-devtools
        ├─ Create types/ layer
        │   ├─ types/model.ts        — ParseResponse, ValidationItem
        │   ├─ types/views.ts        — all *ViewResponse interfaces
        │   ├─ types/changeset.ts    — ChangeAction (DISCRIMINATED UNION), ImpactDelta, etc.
        │   ├─ types/insights.ts     — InsightsResponse, InsightItem
        │   └─ types/common.ts       — RiskLevel, VisibilityLevel, TeamType enums
        ├─ Create services/api/ layer
        │   ├─ services/api/client.ts   — fetch wrapper (error handling, AbortSignal, base URL)
        │   ├─ services/api/models.ts   — parseModel, exportModel, loadExample
        │   ├─ services/api/views.ts    — typed view fetchers (NO generic getView)
        │   ├─ services/api/changesets.ts
        │   ├─ services/api/insights.ts
        │   ├─ services/api/advisor.ts
        │   └─ services/api/index.ts    — re-exports
        ├─ Create hooks/ updates
        │   ├─ Replace useModelView → useViewQuery (TanStack Query wrapper)
        │   ├─ Replace InsightsContext → useQuery with refetchInterval
        │   ├─ Replace useAIEnabled → useQuery with staleTime: Infinity
        │   ├─ Clean SearchContext (remove unused teamFilter, actorFilter, teamTypeFilter)
        │   └─ Delete useRequireModel (replace all imports with useModel)
        ├─ Add QueryClientProvider to app providers
        ├─ Delete old lib/api.ts (after all imports migrated)
        ├─ Delete lib/InsightsContext.tsx (replaced by TanStack Query)
        └─ Verify: npm run build + npm run test pass, zero visual changes

STEP 1  Shared component library
        ├─ Install missing shadcn/ui components: Tabs, Select, Dialog, Tooltip, Table, Sheet
        ├─ components/ui/page-header.tsx
        ├─ components/ui/stat-card.tsx
        ├─ components/ui/tab-bar.tsx        (or use shadcn Tabs directly)
        ├─ components/ui/filter-bar.tsx
        ├─ components/ui/collapsible-card.tsx
        ├─ components/ui/insight-banner.tsx
        ├─ components/ui/risk-badge.tsx
        ├─ components/ui/empty-state.tsx
        ├─ components/ui/data-table.tsx
        ├─ components/ui/health-ring.tsx    (extracted from DashboardPage)
        └─ components/ui/speedometer.tsx    (extracted from CognitiveLoadView)
        → npm run build must pass

STEP 2  Rewrite layout shell
        ├─ Sidebar.tsx → collapsible icon rail with 4 sections
        ├─ TopBar.tsx → Tailwind-only, add SectionTabs slot
        ├─ SectionTabs.tsx (new) → horizontal tabs per section
        └─ App.tsx → remove legacy /views/* redirects, remove EditModelPage
        → npm run build must pass

STEP 3  Rewrite pages (simplest to most complex)
        Each page rewrite must:
        - Use TanStack Query (useQuery) for data fetching
        - Use PageHeader, StatCard, TabBar etc. from shared components
        - Use Tailwind only (zero inline style={{}} except dynamic computed values)
        - Stay under 300 lines (extract sub-components to features/)
        - Use typed imports from types/ (zero `as unknown as` casts)

        Order:
        1.  AdvisorPage.tsx          (153 lines → ~100)
        2.  RecommendationsPage.tsx  (315 lines → ~120)
        3.  WhatIfPage.tsx           (322 lines → ~80, extract tabs to features/whatif/)
        4.  NeedView.tsx             (375 lines → ~150)
        5.  UploadPage.tsx           (388 lines → ~200)
        6.  CognitiveLoadView.tsx    (513 lines → ~200, Speedometer → components/ui/)
        7.  DashboardPage.tsx        (562 lines → ~250, extract HealthRing etc.)
        8.  RealizationView.tsx      (598 lines → ~200)
        9.  SignalsView.tsx          (857 lines → ~250)
        10. TeamTopologyView.tsx     (989 lines → ~250)
        11. OwnershipView.tsx        (1,025 lines → ~250)
        12. CapabilityView.tsx       (1,045 lines → ~250)
        → npm run build after each page

STEP 4  UNMMapView — full React Flow rewrite
        This is separate from Step 3 because it's a different kind of work.

        4a. Create services/transforms/map-transform.ts
            Transform UNMMapViewResponse → React Flow nodes + edges
            Include unit tests (map-transform.test.ts)

        4b. Create features/unm-map/
            ├─ layout.ts              — visibility band positioning (pure, tested)
            ├─ layout.test.ts
            ├─ chain.ts               — BFS highlight logic (pure, tested)
            ├─ chain.test.ts
            ├─ types.ts               — UNM map-specific types
            ├─ constants.ts           — colors, dimensions
            └─ panels.ts             — panel builder functions (pure, tested)

        4c. Create components/unm-map/
            ├─ ActorNode.tsx           — custom React Flow node
            ├─ NeedNode.tsx
            ├─ CapabilityNode.tsx
            ├─ ExtDepNode.tsx
            ├─ MapToolbar.tsx          — legend + zoom
            ├─ DetailDrawer.tsx        — slide-in panel (positioned relative, not fixed)
            └─ CapabilityEditForm.tsx

        4d. Rewrite UNMMapView.tsx as thin orchestrator using React Flow
            ├─ <ReactFlow> with custom node types
            ├─ Built-in zoom/pan (remove hand-rolled mouse handlers)
            ├─ Built-in edge rendering (remove 6 SVG path blocks)
            ├─ TanStack Query for data fetching
            ├─ queryClient.invalidateQueries() on changeset commit (replaces refreshKey)
            └─ Target: ~150 lines

        → npm run build + npm run test must pass
        → Visual verification: all node types render, edges connect correctly,
          zoom/pan works, click-to-highlight works, edit form stages changes

STEP 5  Delete dead code + final validation
        ├─ Delete EditModelPage.tsx
        ├─ Delete all orphaned style constants
        ├─ Delete old lib/api.ts if not already deleted
        ├─ Delete lib/InsightsContext.tsx if not already deleted
        ├─ npm run build must pass
        ├─ npm run test must pass
        └─ Full smoke test: Upload → Dashboard → every view → AI tools → edit mode
```

**Key constraints:**
- Every page must preserve its existing functionality exactly
- This is a presentation-layer + architecture rewrite, not a feature change
- The backend is not touched at all
- All data fetching uses TanStack Query (no manual useEffect + useState patterns)
- All styling uses Tailwind (no inline style={{}} except React Flow node positions)
- UNMMapView uses React Flow (no hand-rolled SVG layout engine)
- Every pure function (layout, chain, transforms) has unit tests

**Quality gates (check after every step):**
- `cd frontend && npm run build` passes
- `cd frontend && npm run test` passes
- No `as unknown as` casts in new code
- No inline `style={{}}` except computed dynamic values
- No duplicate style constants across files
- Every page ≤ 300 lines (except UNMMapView canvas which is React Flow)
- Every page handles loading, error, and empty states

---

## Appendix A: File-by-file line counts

| File | Lines |
|------|-------|
| `views/UNMMapView.tsx` | 1,121 |
| `views/CapabilityView.tsx` | 1,045 |
| `views/OwnershipView.tsx` | 1,025 |
| `views/TeamTopologyView.tsx` | 989 |
| `views/SignalsView.tsx` | 857 |
| `views/RealizationView.tsx` | 598 |
| `DashboardPage.tsx` | 562 |
| `views/CognitiveLoadView.tsx` | 513 |
| `UploadPage.tsx` | 388 |
| `views/NeedView.tsx` | 375 |
| `WhatIfPage.tsx` | 322 |
| `RecommendationsPage.tsx` | 315 |
| `AdvisorPage.tsx` | 153 |
| **Total** | **~8,267** |

## Appendix B: Style system usage by file

| File | Inline `style={{}}` | Tailwind `className` | CSS vars | Button/Card from `ui/` |
|------|---------------------|---------------------|----------|----------------------|
| DashboardPage | **Heavy** | Minimal | No | No |
| UploadPage | Moderate | **Primary** | No | Yes (Button) |
| NeedView | **Heavy** | Moderate | No | No |
| CapabilityView | **Heavy** | Moderate | No | No |
| OwnershipView | **Heavy** | Moderate | No | No |
| SignalsView | **Heavy** | Moderate | No | No |
| TeamTopologyView | **Heavy** | Moderate | No | No |
| CognitiveLoadView | **Heavy** | Moderate | No | No |
| RealizationView | Moderate | **Primary** | No | No |
| WhatIfPage | Moderate | **Primary** | No | Yes (Card, Button) |
| AdvisorPage | Moderate | **Primary** | No | No |
| RecommendationsPage | Moderate | **Primary** | No | Yes (Button) |
| UNMMapView | **Heavy** (SVG) | Minimal | No | No |

## Appendix C: Data fetching patterns

| File | Pattern | Hook |
|------|---------|------|
| NeedView | `useModelView` | ✅ Standardized |
| CapabilityView | `useModelView` | ✅ Standardized |
| TeamTopologyView | `useModelView` | ✅ Standardized |
| CognitiveLoadView | `useModelView` | ✅ Standardized |
| OwnershipView | Manual `useEffect` | ❌ Should migrate |
| SignalsView | Manual `useEffect` | ❌ Should migrate |
| RealizationView | Manual `useEffect` + `Promise.all` | ❌ Should migrate |
| DashboardPage | Manual `useEffect` (2 calls) | ❌ Should migrate |
| UNMMapView | Manual `useEffect` | ❌ Should migrate |

---

# Part 2: Deep Dive — UNMMapView.tsx (1,157 lines)

This component is the most complex in the codebase and the hardest to work on.
Every AI-generated change to it risks breaking something because the component
has no separation of concerns — data fetching, graph theory, coordinate math,
SVG rendering, HTML rendering, edit forms, detail panels, drag handling, zoom,
and changeset integration are all in one file with tight coupling throughout.

---

## UNM-1. Structural Anatomy — One Function Does Everything

The file breaks down as follows:

| Lines | Concern | Description |
|-------|---------|-------------|
| 1–10 | Imports | 9 imports |
| 11–36 | Constants | 21 layout constants + 1 color map + 1 team color hasher |
| 38–65 | Types | 5 interfaces (`SvcInfo`, `PNode`, `Conn`, `ActorGroup`, `BandInfo`, `ChainData`) |
| 67–151 | Graph traversal | `computeChain()` — BFS highlight logic (85 lines) |
| 153–174 | More types + helpers | `PanelItem` interface + `capCardHeight()` |
| 176–415 | Layout engine | `buildLayout()` — the 240-line coordinate calculator |
| 418–1156 | **The component** | **738 lines** — everything else |

The component itself (lines 418–1156) contains:
- **17 `useState` hooks** (lines 421–473)
- **3 `useRef` hooks** (lines 474, 486–488)
- **2 `useMemo` hooks** for layout + chain data (lines 433–462)
- **9 `useCallback` hooks** (mouse handlers, loadMap, openNodePanel, etc.)
- **5 `useEffect` hooks** (data loading, team loading, refresh, wheel zoom)
- **Return JSX: ~440 lines** of mixed SVG + HTML rendering

**Why this is a problem for AI (and humans):**

When an AI agent needs to change e.g. "how capabilities are rendered", it must
understand the full file because:
1. The capability's position comes from `buildLayout()` (line 176–415)
2. The capability's data shape is `PNode` (line 41–49)
3. The capability's click handler is in `openNodePanel()` (line 593–690)
4. The capability's highlight comes from `computeChain()` (line 67–151)
5. The capability's edit form is in the detail panel JSX (line 1052–1126)
6. The capability's visual rendering is in the node layer JSX (line 926–1003)
7. The capability's edge connections are in both `conns` and `depConns` SVG rendering
8. The capability's pending/staged state involves `pendingCapNames`, `stagedCaps`, `isVirtualPending`

That's 8 different locations across 1,157 lines. Touching any one can break the others.

---

## UNM-2. The 240-Line Layout Engine Has No Tests

`buildLayout()` (lines 176–415) is a pure function that takes API data and returns
positioned nodes, connections, and band information. It's the most algorithmically
complex code in the frontend. It:

1. Groups needs under actors and assigns X positions
2. Computes capability centroid positions based on linked needs
3. Groups capabilities into visibility bands with dynamic heights
4. Places capabilities within bands with collision avoidance
5. Places external dependencies below the last band
6. Computes edge connections between all node types
7. Builds service-to-capability cross-reference maps

**This function has zero tests.** Any change to it requires running the full app
and visually inspecting the result. It's also 240 lines of dense coordinate math
where a single off-by-one pixel error creates a broken layout.

**Recommendation:** Extract to its own file (`lib/unm-map-layout.ts`) with
comprehensive unit tests. The function is already pure — it takes data in and
returns a result with no side effects. This is the single most impactful change
for making the map maintainable.

---

## UNM-3. Data Flow is Spaghetti

The data flow through this component is:

```
API (getUNMMapView)
  ↓
rawMapData (useState) — stores raw nodes/edges from API
  ↓
layout (useMemo) — combines rawMapData + pending actions, calls buildLayout()
  ↓
chainData (useMemo) — derives highlight maps from layout + rawMapData
  ↓
pnodes, conns, depConns, extDepConns — destructured from layout
  ↓
JSX rendering — iterates over all arrays

Separately:
API (getTeams) → teams (useState) — for edit form dropdown options
ChangesetContext → actions, refreshKey, isEditMode — for edit mode
usePageInsights('dashboard') → insights — for AI insight badges
```

**Problems:**

1. **Two separate API calls.** `getUNMMapView()` gets nodes and edges, but
   `getTeams()` is called separately just to populate a dropdown. The team
   names are already in the node data (`cap.data.team_label`), so this
   second call is partially redundant.

2. **`actions` from ChangesetContext leak into the layout.** The `useMemo`
   for `layout` (lines 433–449) creates fake `ViewNode` objects for pending
   `add_capability` actions and merges them into the real data. This means
   the layout engine must handle "ghost nodes" — nodes that don't actually
   exist yet. This phantom-node injection uses `as unknown as Record<string, unknown>`
   casts (line 438) because `ChangeAction` doesn't have typed fields per action type.

3. **`refreshKey` coupling.** Line 555–557: when `refreshKey > 0`, `loadMap()`
   re-runs. But `refreshKey` starts at 0 and only bumps on `exitEditMode()`.
   If the user commits a changeset while in edit mode (which is the normal flow),
   the refresh happens via `PendingChangesBar` calling commit, then `exitEditMode()`
   bumps `refreshKey`, then this `useEffect` fires. This is a fragile Rube Goldberg
   machine — a simple `refetch()` callback would be clearer.

---

## UNM-4. The `openNodePanel` Function is a 100-Line Monster

`openNodePanel` (lines 593–690) is a single callback that handles click events for
4 different node types (`ext-dep`, `actor`, `need`, `capability`). For each type,
it builds a completely different `PanelItem` with different fields.

**For capabilities alone** (lines 648–690), it:
1. Looks up visibility config
2. Builds service text
3. Builds a slug for AI insight lookup
4. Checks 3 insight key patterns (`cap:`, `cap-fragmented:`, `cap-disconnected:`)
5. Builds AI insight fields
6. Computes unique teams
7. Checks if node is a pending ghost
8. Sets panel state with 8+ conditional fields
9. Initializes edit state with 8 fields tracking original values

All of this is a single function body. If an AI needs to change "what happens when
you click a capability", it needs to understand all 100 lines of this function
plus the edit state it initializes.

**Recommendation:** Split into `buildActorPanel()`, `buildNeedPanel()`,
`buildCapabilityPanel()`, `buildExtDepPanel()` — each a pure function that
takes a node and returns a `PanelItem`.

---

## UNM-5. The Render JSX is 440 Lines of Mixed SVG + HTML

The return block (lines 719–1154) contains:

| Section | Lines | Content |
|---------|-------|---------|
| Legend bar | 723–756 | Mixed Tailwind + inline styles, SVG within HTML |
| SVG layer | 775–870 | Arrow markers, band backgrounds, actor outlines, 3 edge types × 2 (visual + hit area) |
| HTML node layer | 873–1005 | 4 node types with deeply nested inline styles |
| Detail drawer | 1014–1151 | Fixed-position panel with edit form + info fields |

**The node rendering alone** (lines 873–1005) contains 4 `if/return` branches
for actor, need, ext-dep, and capability. The capability branch (lines 926–1003)
is 77 lines of deeply nested inline style objects with conditional logic for
`isPending`, `isVirtualPending`, `justStaged`, `isFragmented`, and `crossTeam`.

**The edge rendering** is done 6 times — visual path + hit area for each of
3 edge types (demand, supply, ext-dep). Each pair is nearly identical:

```typescript
// Visual path (line 822–828)
<path d={d} fill="none" stroke={c.color} strokeWidth={1.2}
  opacity={connOpacity(...)} markerEnd={marker} style={{ pointerEvents: 'none' }} />

// Hit area (line 848–853) — same path, transparent, wider
<path d={d} fill="none" stroke="transparent" strokeWidth={14}
  style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
  onClick={e => { e.stopPropagation(); openConnPanel(c) }} />
```

This pattern is repeated 3 times with minor variations. A single `<EdgePath>`
component would replace all 6 blocks.

---

## UNM-6. The Detail Drawer Panel is Hardcoded at Position `top: 56`

```typescript
style={{
  position: 'fixed', right: 0, top: 56, bottom: 0, width: 320,
  ...
}}
```

The `top: 56` assumes the TopBar is exactly 56px tall. If the TopBar height
changes (which it will in the UI rewrite), this panel breaks. It also doesn't
account for the SectionTabs that will be added.

The panel should be positioned relative to its container, not the viewport.

---

## UNM-7. Edit State Tracking is Fragile

The edit form (lines 1052–1126) tracks original values to compute diffs:

```typescript
const [editState, setEditState] = useState<{
  capLabel: string; description: string; visibility: string; teamName: string
  origDescription: string; origVisibility: string; origTeam: string
  svcs: SvcInfo[]
} | null>(null)
```

`handleSaveEdit` (lines 559–578) compares current values to `orig*` values
to build change actions:

```typescript
if (editState.description !== editState.origDescription)
  actions.push({ type: 'update_description', ... })
if (editState.visibility !== editState.origVisibility)
  actions.push({ type: 'update_capability_visibility', ... })
if (editState.teamName !== editState.origTeam)
  actions.push({ type: 'reassign_capability', ... })
```

The `stagedCaps` visual feedback (line 573–574) uses a `setTimeout` to remove
the green checkmark after 2 seconds — a fire-and-forget timeout with no cleanup
on unmount:

```typescript
setStagedCaps(prev => new Set([...prev, capName]))
setTimeout(() => setStagedCaps(prev => { ... }), 2000)
```

If the component unmounts during those 2 seconds, React will warn about
updating state on an unmounted component.

---

## UNM-8. The `as unknown as` Cast Epidemic

The file contains multiple unsafe type casts:

```typescript
// Line 438 — casting changeset action to extract fields
const ac = a as unknown as Record<string, unknown>

// Line 481 — same pattern again
const an = a as unknown as Record<string, unknown>

// Line 236-243 — casting node.data fields
cap.data.team_label as string
cap.data.team_type as string
cap.data.services as Array<{ id: string; label: string; team_name?: string }>
cap.data.visibility as string
```

The `ViewNode.data` type is `Record<string, unknown>`, which forces every
access to cast. A typed view-specific node interface would eliminate all of
these casts.

---

## UNM-9. Proposed Decomposition

The file should be broken into these modules:

```
pages/views/UNMMapView.tsx            → ~150 lines (orchestrator)
lib/unm-map/
  ├─ layout.ts                        → buildLayout() + capCardHeight() (testable, pure)
  ├─ layout.test.ts                   → unit tests for layout computation
  ├─ chain.ts                         → computeChain() (testable, pure)
  ├─ chain.test.ts                    → unit tests for BFS highlighting
  ├─ types.ts                         → PNode, Conn, BandInfo, ActorGroup, ChainData, etc.
  ├─ constants.ts                     → all layout constants + VIS color map
  └─ panels.ts                        → buildActorPanel, buildNeedPanel, etc. (pure)

components/unm-map/
  ├─ MapCanvas.tsx                    → SVG + HTML rendering (scrollable, zoomable container)
  ├─ MapToolbar.tsx                   → Legend, zoom controls, clear-highlight button
  ├─ ActorNode.tsx                    → Actor node rendering
  ├─ NeedNode.tsx                     → Need node rendering
  ├─ CapabilityNode.tsx               → Capability node rendering (handles pending/staged states)
  ├─ ExtDepNode.tsx                   → External dependency node rendering
  ├─ EdgePath.tsx                     → Single edge component (visual + hit area)
  ├─ BandBackground.tsx               → Visibility band SVG rendering
  ├─ DetailDrawer.tsx                 → Slide-in detail panel
  └─ CapabilityEditForm.tsx           → Edit form for capabilities
```

**What `UNMMapView.tsx` becomes:**

```tsx
export function UNMMapView() {
  const { data, loading, error, refetch } = useViewData(api.getUNMMapView)
  const layout = useMemo(() => data ? buildLayout(data, pendingActions) : null, [data, pendingActions])

  if (loading) return <LoadingState />
  if (error) return <ErrorState message={error} />
  if (!layout) return null

  return (
    <ModelRequired>
      <MapToolbar zoom={zoom} onZoomChange={setZoom} highlight={highlight} onClearHighlight={...} />
      <MapCanvas layout={layout} zoom={zoom} onNodeClick={...} onEdgeClick={...}>
        {layout.bands.map(b => <BandBackground key={b.vis} band={b} />)}
        {layout.conns.map(c => <EdgePath key={c.id} conn={c} />)}
        {layout.pnodes.map(n => <MapNode key={n.id} node={n} />)}
      </MapCanvas>
      <DetailDrawer panel={panel} editState={editState} onClose={...} onSave={...} />
    </ModelRequired>
  )
}
```

~150 lines. Each child component is independently testable, readable, and
changeable without understanding the full 1,157-line context.

---

## UNM-10. Implementation Priority for the Rewrite

In the big-bang rewrite plan (Section 8, Step 3), UNMMapView is listed last
(item 13) because it's the most complex. Here's the specific sub-step order:

```
UNMMapView rewrite sub-steps:

1. Extract types    → lib/unm-map/types.ts
2. Extract constants → lib/unm-map/constants.ts
3. Extract layout   → lib/unm-map/layout.ts + layout.test.ts
4. Extract chain    → lib/unm-map/chain.ts + chain.test.ts
5. Extract panels   → lib/unm-map/panels.ts
6. Create node components → components/unm-map/{ActorNode,NeedNode,CapabilityNode,ExtDepNode}.tsx
7. Create EdgePath  → components/unm-map/EdgePath.tsx
8. Create BandBackground → components/unm-map/BandBackground.tsx
9. Create MapToolbar → components/unm-map/MapToolbar.tsx
10. Create MapCanvas → components/unm-map/MapCanvas.tsx (scroll, zoom, drag)
11. Create DetailDrawer → components/unm-map/DetailDrawer.tsx
12. Create CapabilityEditForm → components/unm-map/CapabilityEditForm.tsx
13. Rewrite UNMMapView.tsx as thin orchestrator
14. npm run build + visual verification

Each sub-step should keep the app working. Steps 1–5 are pure extractions
(no visual changes). Steps 6–12 create components that are initially unused.
Step 13 switches to the new components.
```

Quality gates for UNMMapView specifically:
- Layout engine has unit tests covering: single actor, multiple actors, all visibility bands, empty bands, external dependencies, pending capabilities
- Chain computation has unit tests covering: actor click, need click, capability click, ext-dep click, dependency traversal
- Each node component renders correctly in isolation
- Zoom, drag-to-pan, and click-to-highlight all work
- Edit form stages changes to changeset context
- Detail drawer slides in/out without overlapping TopBar or SectionTabs

---

# Part 3: Frontend Architecture Principles Gap Analysis

This section goes deeper than the visual/UX issues covered in Part 1, analyzing the actual
code architecture: API client design, state management, data flow, type safety, and
backend API contract alignment.

---

## 9. API Client Layer (`lib/api.ts`) — 601 lines of problems

### 9.1 Massive monolith file with no structure

`api.ts` is a single 601-line file containing:
- **27 type/interface definitions** (lines 16–424)
- **22 API methods** on a single `api` object (lines 426–600)
- No grouping, no namespacing, no separation of concerns

The types and the fetcher logic are mixed together in one file. Types should live separately
so components can import types without pulling in the fetch layer.

### 9.2 Duplicate endpoints — two ways to call the same backend route

The file has **two paths to the same backend view data**:

| Generic method | Typed method | Backend route |
|----------------|-------------|---------------|
| `api.getView(id, 'need')` | `api.getNeedView(id)` | `GET /api/models/{id}/views/need` |
| `api.getView(id, 'capability')` | `api.getCapabilityView(id)` | `GET /api/models/{id}/views/capability` |
| `api.getView(id, 'cognitive-load')` | `api.getCognitiveLoadView(id)` | `GET /api/models/{id}/views/cognitive-load` |
| `api.getView(id, 'signals')` | `api.getSignals(id)` | `GET /api/models/{id}/views/signals` |

`DashboardPage` actually uses `api.getView(modelId, 'cognitive-load')` and then casts it
with `as unknown as CognitiveLoadViewResponse` — a type-safety hole. Meanwhile, the typed
method `api.getCognitiveLoadView()` exists and is used by `CognitiveLoadView`. This is the
kind of duplication that causes bugs.

**Recommendation:** Remove `getView()` entirely. Keep only the typed methods. If a component
needs cognitive-load data, it should call `api.getCognitiveLoadView()`.

### 9.3 Query endpoints (getCapabilities, getTeams, etc.) — dead weight?

There are 5 "query" endpoints:

```
api.getCapabilities(id) → GET /api/models/{id}/capabilities
api.getTeams(id)        → GET /api/models/{id}/teams
api.getNeeds(id)        → GET /api/models/{id}/needs
api.getServices(id)     → GET /api/models/{id}/services
api.getActors(id)       → GET /api/models/{id}/actors
```

These return **flat lists** (just names and basic fields). They are only used in two places:

1. `ActionForm.tsx` — to populate dropdown options for changeset actions
2. `UNMMapView.tsx` — to get team names
3. `model-context.tsx` — to verify a stored model still exists (calls `getActors` as a health check)

The view endpoints already return all this data (and more) as part of their enriched responses.
These query endpoints are a **separate backend surface** that returns a subset of data that's
already available. For the ActionForm use case, the entity names could come from the
`ParseResponse.summary` or from a dedicated lightweight endpoint.

**Recommendation:** Evaluate whether these can be consolidated into the model's `ParseResponse`
(which the frontend already has in memory via `ModelContext`). The model-exists check should
use `GET /health` or a dedicated `HEAD /api/models/{id}` endpoint, not a full actors fetch.

### 9.4 No request cancellation

No API call uses `AbortController`. When users navigate between views quickly, stale responses
from previous pages can arrive and update state for a view the user has already left.
`useModelView` doesn't cancel its fetch when deps change or unmount. The manual `useEffect`
patterns in other views don't either.

### 9.5 No response caching

Every view fetches fresh data on every mount. For the same model ID, navigating from
NeedView to CapabilityView and back causes 3 redundant fetches. `InsightsContext` caches
insights, but view data has no caching at all.

### 9.6 Inconsistent `encodeURIComponent` usage

Some methods encode the model ID:
```typescript
api.getInsights(modelId: string)     → encodeURIComponent(modelId)
api.getNeedView(modelId: string)     → encodeURIComponent(modelId)
```

Others don't:
```typescript
api.getCapabilities(id: string)      → raw `${id}`
api.getTeams(id: string)             → raw `${id}`
api.getSignals(id: string)           → raw `${id}`
```

Model IDs are hex strings so this doesn't break today, but it's sloppy.

---

## 10. State Management — 4 Contexts, 3 Patterns, 0 Coordination

### 10.1 Context overview

| Context | Purpose | Provider location | State stored |
|---------|---------|-------------------|-------------|
| `ModelContext` | Current model ID + parse result | Wraps `<Routes>` | modelId, parseResult, loadedAt |
| `SearchContext` | Global search query + filters | Wraps `<Routes>` | query, teamFilter, actorFilter, teamTypeFilter |
| `ChangesetContext` | Edit mode + pending actions | Wraps `<Routes>` (outermost) | isEditMode, actions[], description, refreshKey |
| `InsightsContext` | AI insights cache | Wraps `<AppShell>` only when AI enabled | cache Map, inflight Map |

### 10.2 `useRequireModel` is a no-op

```typescript
export function useRequireModel() {
  const { modelId, parseResult, loadedAt, isHydrating, setModel, clearModel } = useModel()
  return { modelId, parseResult, loadedAt, isHydrating, setModel, clearModel }
}
```

This function destructures `useModel()` and re-returns the exact same fields. It adds
zero logic. It was presumably intended to enforce a guard (throw/redirect if no model),
but `ModelRequired` handles that at the JSX level. Six views import `useRequireModel`
while others import `useModel` directly — the inconsistency suggests unclear intent.

**Recommendation:** Delete `useRequireModel`. Use `useModel()` everywhere. The guard is
handled by `<ModelRequired>`.

### 10.3 `SearchContext` has unused fields

`SearchContext` exposes `teamFilter`, `actorFilter`, and `teamTypeFilter`, but:

- `teamFilter` is only used in `OwnershipView` (reads it)
- `actorFilter` is never read by any component
- `teamTypeFilter` is never read by any component
- No component **sets** `actorFilter` or `teamTypeFilter`

These filters are defined but have no UI and no consumers.

**Recommendation:** Remove unused fields. If per-page filtering is needed, handle it with
local component state (as most views already do with their own filter state).

### 10.4 No coordination between contexts

`ChangesetContext` tracks edit mode and pending actions, but:
- When a changeset is committed, `PendingChangesBar` manually calls `setModel()` to update
  `ModelContext` with the new parse result. This cross-context update is done via a callback
  chain, not a coordinated state update.
- `refreshKey` (a counter bumped on exit-edit-mode) is consumed by `UNMMapView` to trigger
  a reload — a manual pub/sub pattern that only works for one view.
- If the user is on `NeedView` when a changeset is committed, NeedView won't reload because
  it doesn't watch `refreshKey`.

### 10.5 `InsightsContext` placement creates conditional provider nesting

```typescript
// AppShell.tsx
const shell = <div>...<Outlet />...</div>
return aiEnabled ? <InsightsProvider>{shell}</InsightsProvider> : shell
```

When AI is disabled, `InsightsContext` doesn't exist. `useInsightsContext()` handles this
with a fallback noop, but this conditional provider nesting is an unusual pattern that makes
the component tree inconsistent between AI-enabled and AI-disabled modes.

---

## 11. Data Flow — The Spaghetti Map

### 11.1 How data reaches components (6 different patterns)

```
Pattern A: useModelView hook (standardized)
  Component → useModelView(api.getXxxView) → { data, loading, error }
  Used by: NeedView, CapabilityView, TeamTopologyView, CognitiveLoadView

Pattern B: Manual useEffect + single fetch
  Component → useEffect → api.getX(modelId).then(setData) → { data, loading, error }
  Used by: SignalsView, UNMMapView

Pattern C: Manual useEffect + Promise.all (multiple endpoints)
  Component → useEffect → Promise.all([api.getX(), api.getY()]).then(...)
  Used by: OwnershipView (ownership + capability), RealizationView (realization + need)

Pattern D: Manual useEffect + multiple sequential fetches
  Component → useEffect → api.getSignals() + api.getView('cognitive-load')
  Used by: DashboardPage

Pattern E: ParseResponse from context (no API call)
  Component → useModel() → parseResult.summary
  Used by: WhatIfPage (for suggestion chips)

Pattern F: Direct API call in callback (not on mount)
  Component → user action → api.askAdvisor() / api.createChangeset()
  Used by: AdvisorPage, WhatIfPage, PendingChangesBar
```

### 11.2 Cross-view data fetching

Two views fetch data from **other** view endpoints to supplement their own:

| View | Fetches its own endpoint | Also fetches |
|------|--------------------------|-------------|
| OwnershipView | `getOwnershipView()` | `getCapabilityView()` — to get parent groups |
| RealizationView | `getRealizationView()` | `getNeedView()` — to build need→capability chains |
| DashboardPage | `getSignals()` | `getView('cognitive-load')` — to show team load preview |

This reveals a **backend API design issue**: the view endpoints don't return self-contained
data. OwnershipView needs capability parent-groups but doesn't get them from its own endpoint.
RealizationView needs actor→need chains but gets them from the need endpoint instead.

**Recommendation:** Either:
1. Enrich each view endpoint to return everything the frontend needs (self-contained), or
2. Create a data layer that can compose multiple API responses before passing to components.

### 11.3 No data invalidation strategy

When a changeset is committed:
1. `PendingChangesBar` calls `api.commitChangeset()`, gets back a `CommitResponse`
2. It manually reconstructs a `ParseResponse` by merging `CommitResponse.summary` with the existing `parseResult`
3. It calls `setModel()` to update `ModelContext`
4. It bumps `refreshKey` so `UNMMapView` reloads

But every other view still has stale data. If the user is on `NeedView`, navigates to
`UNMMapView` to commit, then goes back to `NeedView`, they see the old pre-commit data
until they navigate away and back (triggering a remount and fresh fetch).

---

## 12. Type Safety Issues

### 12.1 `as unknown as` casts

```typescript
// DashboardPage.tsx — casting generic ViewResponse to CognitiveLoadViewResponse
api.getView(modelId, 'cognitive-load')
  .then(data => setTeamLoads((data as unknown as CognitiveLoadViewResponse).team_loads ?? []))

// OwnershipView.tsx — casting capability view response to a local subset type
setCapViewData(capData as unknown as CapViewData)
```

These double-casts bypass TypeScript's type system entirely. If the backend response shape
changes, these won't catch it at compile time.

### 12.2 `ChangeAction` — a 27-type union with 25 optional flat fields

```typescript
export interface ChangeAction {
  type: 'move_service' | 'split_team' | 'merge_teams' | ... 25 more ...
  service_name?: string
  from_team_name?: string
  to_team_name?: string
  ... 22 more optional fields ...
}
```

Every field except `type` is optional because different action types use different fields.
This means TypeScript can't enforce that `move_service` requires `service_name` and
`from_team_name` — you can pass `{ type: 'move_service' }` with zero fields and it compiles.

**Recommendation:** Use discriminated unions:

```typescript
type ChangeAction =
  | { type: 'move_service'; service_name: string; from_team_name: string; to_team_name: string }
  | { type: 'split_team'; original_team_name: string; new_team_a_name: string; ... }
  | ...
```

### 12.3 Backend response field naming inconsistencies

| Field | Used in | Name |
|-------|---------|------|
| Validation valid flag | `ParseResponse` | `is_valid` |
| Validation valid flag | `CommitResponse` | `valid` |
| Summary field | `ParseResponse` | `summary` (typed: `{ actors, needs, capabilities, services, teams }`) |
| Summary field | `CommitResponse` | `summary` (typed: `Record<string, number>`) |

The same concept (`is the model valid?`) has two different field names depending on
which endpoint returns it. The summary has two different types for what is logically
the same data.

---

## 13. Backend API Design Issues

### 13.1 View endpoints are not self-contained

As noted in 11.2, views that should be self-contained require fetching from multiple endpoints.
The backend already has all the data when building a view — it should include related data
in the response rather than forcing the frontend to call multiple endpoints.

### 13.2 Dead/broken routes

`POST /api/models/analyze/changesets` and `POST /api/models/analyze/ask` are registered
in `changeset.go` and route to `handleAnalyze`, which reads `r.PathValue("type")` — but
these literal paths **don't define** a `{type}` path variable. `type` will be empty,
`ValidAnalysisType` will fail, and these endpoints return 400. They appear unused from
the frontend.

### 13.3 Inconsistent error response format

| Endpoint | Error format |
|----------|-------------|
| Most endpoints | `{"error": "message"}` |
| Export | Raw text (when not JSON) |
| Health | Manual JSON write (not using `writeJSON`) |
| Panic recovery | Plain-text `http.Error` |
| Commit (409) | `CommitResponse` with `validation.valid: false` |
| Insights | Rich payload with `error` and `status` fields |

### 13.4 No API versioning

All routes are under `/api/` with no version prefix. When the response shapes change
(and they will as views become self-contained), there's no way to support old and new
clients simultaneously.

### 13.5 Model store is in-memory only

Both `ModelStore` and `ChangesetStore` are in-memory maps. Server restart loses all
models. This is fine for the current prototype phase but should be called out — any
future persistence strategy will need to be retrofitted into the entire handler layer.

---

## 14. Hooks Architecture

### 14.1 `useModelView` — good pattern, incomplete adoption

`useModelView` is the right abstraction. It handles hydration, loading, error, and data
states in one place. But only 4 out of 9 views use it. The other 5 can't use it because:

1. They need **multiple API calls** (OwnershipView, RealizationView, DashboardPage)
2. They need **reload on external trigger** (UNMMapView via `refreshKey`)
3. They manage **URL-based filter state** alongside data (SignalsView via `useSearchParams`)

**Recommendation:** Extend `useModelView` (or create a more capable hook) to support:
- Multiple fetchers (returns composed data)
- Manual refetch trigger (for post-commit reload)
- AbortController for cancellation

### 14.2 `useAIEnabled` — module-level singleton cache

```typescript
let resolved: boolean | null = null
export function useAIEnabled(): boolean {
  const [enabled, setEnabled] = useState(resolved ?? false)
  useEffect(() => {
    if (resolved !== null) return
    getRuntimeConfig().then(cfg => { resolved = cfg.ai?.enabled ?? false; setEnabled(resolved) })
  }, [])
  return enabled
}
```

This uses a **module-level mutable variable** (`resolved`) to cache the result. This works
but is unusual — it means the first component to mount with `useAIEnabled()` triggers the
fetch, and all subsequent components immediately see the cached value. The problem: if the
runtime config changes, there's no way to invalidate without a full page reload.

### 14.3 `usePageInsights` — polling with timeouts

Well-implemented polling hook with cancellation, timeout, and state transitions. This is
actually the best-architectured hook in the codebase. However, it's tightly coupled to
`InsightsContext` which is conditionally mounted (see 10.5).

---

## 15. Architecture Recommendations for the Rewrite

These recommendations should be incorporated into the Step 1 (shared component library)
and Step 3 (page rewrites) of the implementation plan.

### 15.1 Split `api.ts` into modules

```
lib/api/
  ├─ client.ts          — fetch wrapper with error handling, AbortController, base URL
  ├─ types.ts           — all response/request types (imported by components)
  ├─ models.ts          — parseModel, exportModel, loadExample
  ├─ views.ts           — getNeedView, getCapabilityView, etc. (typed, no generic getView)
  ├─ changesets.ts      — createChangeset, commitChangeset, getImpact
  ├─ insights.ts        — getInsights, getInsightsStatus
  ├─ advisor.ts         — askAdvisor
  └─ index.ts           — re-exports everything as `api`
```

### 15.2 Add a `useViewData` hook that replaces both `useModelView` and manual patterns

```typescript
function useViewData<T>(
  fetcher: (id: string, signal: AbortSignal) => Promise<T>,
  options?: { refreshOn?: unknown[] }
): { data: T | null; loading: boolean; error: string | null; refetch: () => void }
```

Features over current `useModelView`:
- **AbortController** — cancels in-flight request on unmount or dep change
- **`refetch()`** — manual trigger for post-commit reload (replaces `refreshKey`)
- **`refreshOn`** — optional dependencies that trigger re-fetch (e.g., `[refreshKey]`)
- Support for multi-fetch via a composition helper

### 15.3 Make `ChangeAction` a discriminated union

Type-safe per action type. See recommendation in 12.2.

### 15.4 Delete `useRequireModel`

It's a passthrough. Use `useModel()` everywhere.

### 15.5 Clean up `SearchContext`

Remove `actorFilter`, `teamFilter`, `teamTypeFilter` — they're unused. Keep only `query`
and `setQuery`. Each view manages its own filter state locally (as most already do).

### 15.6 Add data invalidation on changeset commit

When a changeset is committed, the model has changed. Either:
- **Option A:** Bump a model version counter in `ModelContext` that all `useViewData` hooks
  observe, causing all mounted views to refetch.
- **Option B:** Invalidate a client-side cache keyed by `modelId + version`.

### 15.7 Request backend to make views self-contained

Ask backend to include parent-group data in the ownership view response and actor→need
chain data in the realization view response, eliminating the need for cross-view fetching.

---

## 16. Implementation Plan

> **All implementation details are consolidated in Section 8: "Engineer Handoff — Final Execution Plan".**
> Section 8 incorporates all architecture decisions from Parts 1–4 (TanStack Query, React Flow,
> layered file structure, expanded shadcn/ui) into a single, authoritative step-by-step plan.

---

# Part 4: Do We Have the Right Frontend Architecture and Principles?

## The Honest Answer: No

The backend is built on a rigorous Clean Architecture foundation with strict layer
dependencies, 75 test files for 76 production files (99% test coverage by file),
TDD protocol, and clear conventions. The frontend has none of this. It's a collection
of loosely written guidelines that are routinely violated by the codebase itself.

---

## 17. Backend vs Frontend — The Discipline Gap

### 17.1 Architecture rules

| Dimension | Backend (Go) | Frontend (React) |
|-----------|-------------|-----------------|
| **Architecture pattern** | Clean Architecture with 4 layers, strict dependency direction, violation detection rules | None — just "pages, components, hooks, lib" flat folders |
| **Layer dependency rules** | "Domain MUST NOT import from outer layer" — explicit, enforceable | None — components import from anywhere |
| **File organization** | By architectural layer (`domain/entity/`, `usecase/`, `adapter/handler/`) | By feature type (`pages/`, `components/`, `hooks/`, `lib/`) with no sublayering |
| **Single responsibility** | One file per entity, handlers are thin, business logic in domain services | 5 files > 800 lines, components do data fetching + business logic + rendering |
| **Conventions doc** | `go-conventions.md` (32 lines) — naming, file structure, error handling, testing, HTTP handlers | `react-conventions.md` (36 lines) — but it's vague: "use hooks", "use Tailwind", "handle states" |

### 17.2 Testing discipline

| Dimension | Backend | Frontend |
|-----------|---------|----------|
| **Test files** | 75 test files for 76 production files | **3 test files** for 51 production files |
| **Test ratio** | ~1:1 | **1:17** |
| **TDD protocol** | "No production code without a failing test first" | Build validation only (`npm run build`) |
| **What's tested** | Domain logic, parsers, analyzers, handlers, stores, AI integration | `api.test.ts` (error extraction), `model-context.test.tsx` (localStorage), `SignalsView.test.tsx` (one view) |
| **What's NOT tested** | Nothing significant — nearly everything has tests | Layout engines, chain computation, data transformations, all hooks, all views, all rendering |

### 17.3 Clean Architecture enforcement

| Backend has | Frontend equivalent |
|-------------|-------------------|
| `domain/entity/` — pure domain types, zero imports | Types are scattered in `api.ts` (601 lines), view files, and component files |
| `domain/service/` — business logic | Business logic lives inside React components |
| `usecase/` — orchestration | No orchestration layer — components call API directly |
| `adapter/handler/` — thin HTTP layer | No adapter concept — data is consumed raw from JSON |
| `adapter/presenter/` — view model transformers | View models are built inside 1,000-line components with `useMemo` |
| `adapter/repository/` — storage abstraction | `localStorage` calls hardcoded in `model-context.tsx` |
| Layer violation detection rules in `clean-architecture.md` | Nothing |

---

## 18. Are We Using the Right Technology?

### 18.1 What we're building

The UNM Platform frontend is fundamentally an **interactive data visualization tool**
with these characteristics:

1. **Graph/map visualization** — the UNM Map is a custom SVG canvas with positioned
   nodes, directional edges, visibility bands, zoom, pan, and interactive highlighting
2. **Multiple analytical views** — 8 distinct data views (needs, capabilities, ownership,
   team topology, cognitive load, realization, signals, dashboard) each with filtering,
   sorting, and drill-down
3. **Architecture editing** — batch changeset creation with impact preview and commit
4. **AI-assisted analysis** — chat interfaces, recommendations, what-if scenarios
5. **Single-page app** — model loaded once, views navigate between different perspectives
   on the same data

### 18.2 Technology stack assessment

| Technology | Used? | Verdict |
|------------|-------|---------|
| **React 19** | Yes | **Right choice.** SPA with complex interactive views. React is appropriate. |
| **TypeScript** | Yes | **Right choice.** Complex data shapes need type safety. But it's undermined by `as unknown as` casts everywhere. |
| **Vite 6** | Yes | **Right choice.** Fast dev server, good HMR, standard tooling. |
| **Tailwind CSS v4** | Declared | **Right choice, but not actually used.** The majority of the codebase uses inline `style={{}}` instead. |
| **shadcn/ui** | 3 components | **Right choice, but barely adopted.** Only `Button`, `Card`, and `badge` exist. No `Tabs`, `Select`, `Dialog`, `Dropdown`, `Sheet`, `Table`, `Tooltip`, etc. |
| **React Router** | Yes | **Right choice.** Standard, works well for what we need. |
| **React Flow (`@xyflow/react`)** | Installed (12.10.1) | **Installed but NOT USED.** Zero imports in the entire codebase. UNMMapView hand-rolls a custom SVG layout engine instead. This is the single biggest technology misuse. |
| **D3.js** | Listed in AGENT.md as "already installed" | **Not installed, not used.** Not in `package.json`. The AGENT.md says "For data viz: use D3.js (already installed)" — this is incorrect documentation. |
| **Lucide React** | Yes | **Right choice.** Consistent icon library. |
| **react-markdown + remark-gfm** | Yes | **Right choice** for the AI chat/recommendation rendering. |

### 18.3 The React Flow situation — this is critical

`@xyflow/react` (React Flow) version 12.10.1 is installed in `package.json` and
was explicitly chosen for graph/diagram visualization. It provides:

- Node positioning with layout algorithms
- Edge routing with path computation
- Built-in zoom, pan, and minimap
- Node click/drag/selection handling
- TypeScript-first API with typed nodes and edges
- Performance optimization for large graphs

**The UNM Map ignores all of this** and instead hand-builds a 1,157-line SVG layout
engine that:
- Manually computes X/Y coordinates for every node
- Manually draws Bézier curves for every edge
- Manually implements zoom with CSS transforms
- Manually implements drag-to-pan with mouse event handlers
- Manually implements hit areas with invisible wider SVG paths
- Has zero tests

React Flow would replace ~800 of those 1,157 lines. The layout engine, edge
rendering, zoom/pan handling, and hit detection would all be handled by the library.
What remains would be custom node components (~50 lines each) and the data
transformation from API response to React Flow node/edge format (~100 lines).

**Recommendation:** The UNMMapView rewrite should be built on React Flow, not
another hand-rolled SVG engine. The visibility bands can be implemented as
React Flow background panels. The detail drawer stays as a separate component.

### 18.4 Missing technologies we should add

| Need | Current approach | Recommended |
|------|------------------|-------------|
| **State management for cached API data** | Raw `useState` + `useEffect` per view | **TanStack Query (React Query)** — handles caching, deduplication, background refetch, stale-while-revalidate, error retry, request cancellation. Replaces all manual fetch patterns AND the custom `InsightsContext` cache. |
| **Form handling** | Raw `useState` per field + manual diff in `handleSaveEdit` | **React Hook Form** or at minimum a `useForm` hook — handles dirty tracking, validation, submission. |
| **Tab/dialog/dropdown components** | Hand-rolled or missing | **shadcn/ui** (already listed as stack) — adopt `Tabs`, `Select`, `Dialog`, `DropdownMenu`, `Sheet`, `Tooltip`, `Table` components |
| **Data table** | Hand-rolled HTML tables with inline styles | **shadcn/ui Table** or **TanStack Table** for sortable/filterable tables |

---

## 19. What the Frontend Architecture Should Look Like

The backend follows Clean Architecture. The frontend should follow an equivalent
pattern adapted for React — not the same layers verbatim, but the same rigor.

### 19.1 Proposed frontend layer architecture

```
┌─────────────────────────────────────────────────┐
│  Pages (Route-level orchestrators)              │  ← Thin. Compose hooks + components.
├─────────────────────────────────────────────────┤
│  Features (Domain-specific components)          │  ← NeedView, CapabilityCards, MapCanvas
├─────────────────────────────────────────────────┤
│  Components (Shared UI primitives)              │  ← PageHeader, StatCard, TabBar, Button
├─────────────────────────────────────────────────┤
│  Hooks (State + data access)                    │  ← useViewData, useChangeset, useSearch
├─────────────────────────────────────────────────┤
│  Services (API client + data transforms)        │  ← api/, transformers, type guards
├─────────────────────────────────────────────────┤
│  Types (Pure domain types, zero dependencies)   │  ← models, enums, interfaces
└─────────────────────────────────────────────────┘
```

**Dependency rule (same as backend):** Dependencies point downward only.
- Pages import Features, Components, Hooks
- Features import Components, Hooks, Services
- Components import nothing above them (receive data as props)
- Hooks import Services, Types
- Services import Types only
- Types import nothing

### 19.2 Proposed file structure

```
frontend/src/
├── app/
│   ├── App.tsx                       # Router + providers
│   ├── routes.ts                     # Route definitions (data, not JSX)
│   └── providers.tsx                 # Context provider composition
│
├── pages/                            # Route-level orchestrators (THIN)
│   ├── UploadPage.tsx
│   ├── DashboardPage.tsx
│   ├── ExplorePage.tsx               # Hosts: UNMMap, Needs, Capabilities, etc. via tabs
│   ├── TeamsPage.tsx                 # Hosts: Ownership, Topology, CognitiveLoad via tabs
│   └── AIPage.tsx                    # Hosts: WhatIf, Recommendations, Advisor via tabs
│
├── features/                         # Domain-specific feature modules
│   ├── unm-map/
│   │   ├── UNMMapCanvas.tsx          # React Flow canvas
│   │   ├── nodes/                    # Custom React Flow node components
│   │   ├── layout.ts                 # Layout algorithm (pure, tested)
│   │   ├── layout.test.ts
│   │   └── types.ts
│   ├── needs/
│   │   ├── NeedView.tsx
│   │   ├── NeedCard.tsx
│   │   └── need-utils.ts
│   ├── capabilities/
│   ├── ownership/
│   ├── teams/
│   ├── cognitive-load/
│   ├── signals/
│   ├── realization/
│   ├── changeset/
│   └── advisor/
│
├── components/                       # Shared UI (domain-agnostic)
│   ├── layout/
│   │   ├── AppShell.tsx
│   │   ├── Sidebar.tsx
│   │   ├── TopBar.tsx
│   │   └── SectionTabs.tsx
│   └── ui/                           # shadcn/ui primitives
│       ├── button.tsx
│       ├── card.tsx
│       ├── tabs.tsx
│       ├── select.tsx
│       ├── dialog.tsx
│       ├── tooltip.tsx
│       ├── table.tsx
│       ├── stat-card.tsx
│       ├── page-header.tsx
│       ├── filter-bar.tsx
│       ├── risk-badge.tsx
│       ├── empty-state.tsx
│       └── collapsible-card.tsx
│
├── hooks/                            # Shared hooks
│   ├── useViewData.ts                # Data fetching with cache + cancel + refetch
│   ├── useChangeset.ts
│   ├── useSearch.ts
│   └── useAIEnabled.ts
│
├── services/                         # API client + data transforms
│   ├── api/
│   │   ├── client.ts                 # Fetch wrapper (error handling, base URL)
│   │   ├── models.ts                 # parseModel, exportModel
│   │   ├── views.ts                  # getNeedView, getCapabilityView, etc.
│   │   ├── changesets.ts             # createChangeset, commit, impact
│   │   ├── insights.ts
│   │   ├── advisor.ts
│   │   └── index.ts                  # Re-export
│   └── transforms/                   # API response → view model
│       ├── need-transform.ts
│       ├── capability-transform.ts
│       └── map-transform.ts          # API data → React Flow nodes/edges
│
├── types/                            # Pure domain types (ZERO imports)
│   ├── model.ts                      # ParseResponse, ValidationItem
│   ├── views.ts                      # NeedViewResponse, CapabilityViewResponse, etc.
│   ├── changeset.ts                  # ChangeAction (discriminated union), ImpactDelta
│   ├── insights.ts                   # InsightsResponse, InsightItem
│   └── common.ts                     # Shared types (RiskLevel, VisibilityLevel, TeamType)
│
└── lib/                              # Pure utilities (zero React)
    ├── utils.ts                      # cn(), slug()
    ├── config.ts
    └── runtime-config.ts
```

### 19.3 Rules that must be enforced (equivalent to `clean-architecture.md`)

These should be written into a new `.claude/rules/frontend-architecture.md`:

```
# Frontend Architecture Rules

## Layer Dependencies (downward only)

Pages → Features → Components → Hooks → Services → Types → (nothing)

## Violations That Must Be Caught

1. A component in `components/ui/` importing from `features/` or `pages/`
2. A type file in `types/` importing from any other layer
3. A service in `services/` importing React or any React hook
4. A page doing data fetching directly (must use hooks)
5. A feature component > 300 lines (must be decomposed)
6. An API response type cast with `as unknown as` (must use typed response)
7. Inline `style={{}}` when a Tailwind class exists for the same purpose
8. Reimplementing a component that exists in `components/ui/`
9. A component without loading/error/empty state handling for async data
10. A `useState` + `useEffect` manual fetch pattern (must use useViewData or TanStack Query)

## File Size Limits

- Pages: ≤150 lines (orchestrators only)
- Feature components: ≤300 lines
- Shared components: ≤100 lines
- Hooks: ≤100 lines
- Service modules: ≤150 lines

## Testing Requirements

- All pure functions in `lib/`, `services/transforms/`, and `features/*/` must have tests
- Layout algorithms must have comprehensive unit tests
- Data transformation functions must have snapshot tests
- Test files co-located: `foo.ts` → `foo.test.ts`
```

---

## 20. Technology Decision: TanStack Query vs Custom Hooks

The single most impactful technology addition would be **TanStack Query** (React Query).
Here's why:

### What it replaces

| Current pattern | Lines of code | TanStack Query equivalent |
|----------------|---------------|--------------------------|
| `useModelView` hook | 47 lines | `useQuery({ queryKey: ['view', modelId, viewType], queryFn: ... })` |
| Manual `useEffect` fetch patterns (5 views) | ~100 lines total | Same `useQuery` call |
| `InsightsContext` cache + polling | 60 lines | `useQuery` with `refetchInterval` option |
| `useAIEnabled` singleton cache | 19 lines | `useQuery` with `staleTime: Infinity` |
| `refreshKey` hack for post-commit reload | Cross-context coupling | `queryClient.invalidateQueries({ queryKey: ['view', modelId] })` |
| No request cancellation | Multiple bugs | Built-in `AbortController` per query |
| No stale data detection | Stale views after commit | Automatic background refetch |

### What it gives us for free

- **Automatic caching:** Navigate away and back — data is instant from cache
- **Deduplication:** Two components requesting the same view → one network call
- **Background refetch:** Cached data shown immediately, fresh data fetched in background
- **Request cancellation:** Unmounting a view cancels its in-flight request
- **Cache invalidation:** After changeset commit, one call invalidates all stale views
- **Retry logic:** Failed requests automatically retry with backoff
- **DevTools:** Visual inspector for cache state during development

This would replace `InsightsContext`, `useModelView`, all manual `useEffect` fetch
patterns, the `refreshKey` hack, and the `useAIEnabled` singleton — approximately
**300 lines of custom code** replaced by a battle-tested library.

---

## 21. Summary: What Must Change

> **For the step-by-step execution plan, see Section 8.**

| Area | Current State | Target State |
|------|---------------|-------------|
| **Architecture rules** | 36-line `react-conventions.md` — vague guidelines | Full `frontend-architecture.md` with layer deps, violations, file limits |
| **Architecture pattern** | None — flat folders | Layered: Pages → Features → Components → Hooks → Services → Types |
| **Test discipline** | 3 tests for 51 files (1:17 ratio) | Test all pure functions, transforms, and layout algorithms (target 1:3 ratio minimum) |
| **API layer** | 601-line monolith `api.ts` | Modular `services/api/` with typed client |
| **Data fetching** | 6 different manual patterns | TanStack Query everywhere |
| **Type safety** | `as unknown as` casts, flat optional unions | Discriminated unions, typed view responses, zero casts |
| **State management** | 4 contexts with no coordination | TanStack Query cache + minimal context (changeset only) |
| **Map visualization** | 1,157-line hand-rolled SVG engine | React Flow (already installed, never used) |
| **Component library** | 3 shadcn/ui components, everything else hand-rolled | Full shadcn/ui adoption (Tabs, Select, Dialog, Table, etc.) |
| **Styling** | Mixed inline styles + Tailwind + CSS vars | Tailwind only (enforced by architecture rules) |
| **File sizes** | 5 files > 800 lines | Max 300 lines per feature component, 150 per page |
| **Agent docs** | ~~References D3 (not installed), describes manual fetch pattern~~ | **DONE** — Updated to reflect new architecture, hooks, and library choices |
