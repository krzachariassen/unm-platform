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
These shared modules now exist — use them, do not inline:
- `frontend/src/lib/slug.ts` — `toSlug(name)` for URL/id slugification
- `frontend/src/lib/team-type-styles.ts` — `getTeamTypeStyle(type)` for team type badge colors
- `frontend/src/lib/visibility-styles.ts` — `getVisibilityStyle(v)` for capability visibility badges
- `frontend/src/hooks/useModelView.ts` — `useModelView(fetchFn)` shared loading/error/data hook
- `frontend/src/components/ViewState.tsx` — unified loading/error/empty state renderer

### 2026-03 — api.ts parallel agent conflict risk
When two agents both need to modify `frontend/src/lib/api.ts` (e.g., one deleting old methods,
another adding new ones), they will produce a merge conflict. Assign api.ts ownership to ONE agent
per wave. If both need it, sequence them rather than parallelizing.

## Known Gotchas

- Tailwind v4 uses CSS variables, not the old `tailwind.config.js` theme system
- shadcn/ui components are copy-pasted into `components/ui/`, not npm-installed
- Sidebar navigation items must be disabled when no model is loaded
- Team type values from API are kebab-case (`complicated-subsystem`) — format to
  Title Case for display: `"complicated-subsystem"` → `"Complicated Subsystem"`
