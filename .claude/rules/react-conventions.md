# React & Frontend Conventions

## Architecture

See `frontend-architecture.md` for the full layered architecture, file size limits,
and dependency rules. This file covers coding conventions only.

## Component Patterns

- Functional components only (no class components)
- Components in PascalCase files: `NeedView.tsx`, `DetailPanel.tsx`
- Hooks prefixed with `use`: `useModel.ts`, `useAIEnabled.ts`
- Use React 19 features where appropriate
- Props interfaces defined and exported alongside the component
- Prefer composition over configuration — small components composed together

## Styling

- Tailwind CSS v4 utility classes (primary styling method)
- CSS variables for theme values (`text-muted-foreground`, `bg-background`, `border-border`)
- `cn()` utility from `@/lib/utils` for conditional class merging
- shadcn/ui copy-paste pattern for standard UI primitives
- Lucide React for icons — do NOT add other icon libraries
- No inline `style={{}}` except dynamic computed values (React Flow positions, chart offsets)
- No style constant objects — use Tailwind classes

## State Management

- **Server state**: TanStack Query (`useQuery`, `useMutation`, `queryClient.invalidateQueries`)
- **Global client state**: React Context (`ModelContext`, `ChangesetContext`, `SearchContext`)
- **Local UI state**: `useState`
- **URL state**: `useSearchParams`
- No manual `useEffect` + `useState` for data fetching — use TanStack Query
- `useCallback` and `useMemo` only when there is a measured performance need

## API Integration

- All API functions in `services/api/` — never raw `fetch` in components or hooks
- Use `ModelContext` for model ID propagation
- Handle loading, error, and empty states for all data fetches
- Use `enabled: !!modelId` in queries to prevent fetching without a model

## Type Safety

- All API response types in `types/` — not scattered across files
- Discriminated unions for actions/events (not flat optional fields)
- No `as unknown as` casts — fix the type at the source
- No `any` — use `unknown` with runtime type guards

## Routing

- React Router with route definitions in `App.tsx`
- Views as pages under `pages/views/`
- Section tabs for grouped views (not sidebar items)

## Build Validation

- `cd frontend && npm run build` MUST pass (runs `tsc -b` + `vite build`, matching CI)
- `cd frontend && npm run test` MUST pass (runs `vitest`)
