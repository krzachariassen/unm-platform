# Frontend Engineer — Operational Memory

> **Policy**: 30-entry cap · Monthly curation (Promote / Keep / Prune)
> See `.claude/agents/AGENT_OWNERSHIP.md` §2 for full curation rules.

## Learnings

### 2025-03 — Edit Panel UX
The edit workflow uses `EditPanel` component that slides in from the right on
UNMMapView. When EditPanel is open, DetailPanel must be hidden to prevent overlap.
After commit, remount EditPanel via key increment and reload map data.

### 2025-03 — Smart Dropdowns
`ActionForm` uses `useModelEntities()` hook to fetch teams, services, capabilities,
needs, and actors for dropdown population. Entity fields use `type: 'entity'` with
`source` property. The `update_description` action dynamically adjusts entity_name
options based on entity_type selection.

### 2025-03 — Build validation
`cd frontend && npm run build` runs `tsc -b && vite build`, matching CI/Dockerfile.
It catches TypeScript project-mode issues (e.g. unused imports), JSX/transform errors,
and the production bundle in one command.

### 2025-03 — Adjacent JSX in map callbacks
Multiple JSX elements returned from `.map()` must be wrapped in a fragment or
container element. Omitting this causes Vite build failures even if TSC passes.

### 2026-03 — Shared utility modules (Phase 6.12)
These shared modules exist — use them, do not inline:
- `frontend/src/lib/slug.ts` — `toSlug(name)` for URL/id slugification
- `frontend/src/lib/team-type-styles.ts` — `getTeamTypeStyle(type)` for team type badge colors
- `frontend/src/lib/visibility-styles.ts` — `getVisibilityStyle(v)` for capability visibility badges
- `frontend/src/components/ViewState.tsx` — unified loading/error/empty state renderer

### 2026-03 — Architecture rewrite (refactor/ui-unification)
Major architecture overhaul. Key changes:
- **Data fetching**: `useModelView` and manual `useEffect` patterns replaced by **TanStack Query** (`useQuery`)
- **API client**: monolith `lib/api.ts` split into `services/api/` modules (client, models, views, changesets, insights, advisor)
- **Types**: moved from `lib/api.ts` to `types/` directory (model.ts, views.ts, changeset.ts, insights.ts, common.ts)
- **Graph viz**: UNMMapView rewritten with **React Flow** (`@xyflow/react`) — no more hand-rolled SVG layout engine
- **Shared components**: PageHeader, StatCard, TabBar, InsightBanner etc. in `components/ui/`
- **File limits**: Pages ≤ 300 lines, Components ≤ 200 lines, Hooks ≤ 100 lines
- **Features directory**: `features/<name>/` for domain-specific logic (e.g., `features/unm-map/`)
- **D3.js is NOT installed** — previous docs were wrong about this
- See `.claude/rules/frontend-architecture.md` for the full layered architecture rules

## Known Gotchas

- Tailwind v4 uses CSS variables, not the old `tailwind.config.js` theme system
- shadcn/ui components are copy-pasted into `components/ui/`, not npm-installed
- Sidebar navigation items must be disabled when no model is loaded
- Team type values from API are kebab-case (`complicated-subsystem`) — format to
  Title Case for display: `"complicated-subsystem"` → `"Complicated Subsystem"`
- **Deleted files must be staged explicitly** — `git add <file list>` silently skips
  deletions. Always run `git status --short` before committing; if any ` D` lines
  appear, run `git add -A` or `git rm` before committing. Unstaged deletions stay
  in the git index and break Docker builds in CI even when local `npm run build` passes.
