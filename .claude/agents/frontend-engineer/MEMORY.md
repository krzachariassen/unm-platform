# Frontend Engineer — Operational Memory

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
TSC alone misses esbuild JSX transformation issues. ALWAYS run both
`npx tsc --noEmit` AND `npx vite build` to catch all errors.

### 2025-03 — Adjacent JSX in map callbacks
Multiple JSX elements returned from `.map()` must be wrapped in a fragment or
container element. Omitting this causes Vite build failures even if TSC passes.

## Known Gotchas

- Tailwind v4 uses CSS variables, not the old `tailwind.config.js` theme system
- shadcn/ui components are copy-pasted into `components/ui/`, not npm-installed
- Sidebar navigation items must be disabled when no model is loaded
- Team type values from API are kebab-case (`complicated-subsystem`) — format to
  Title Case for display: `"complicated-subsystem"` → `"Complicated Subsystem"`
