# React & Frontend Conventions

## Component Patterns

- Functional components only (no class components)
- Components in PascalCase files: `NeedView.tsx`, `DetailPanel.tsx`
- Hooks prefixed with `use`: `useModel.ts`, `useAIEnabled.ts`
- Use React 19 features where appropriate

## Styling

- Tailwind CSS v4 utility classes (primary styling method)
- CSS variables for theme values
- shadcn/ui copy-paste pattern (not npm installed)
- Lucide React for icons

## State Management

- React hooks: `useState`, `useEffect`, `useCallback`, `useMemo`
- Context for shared state: `ModelContext`, `SearchContext`
- No external state library needed at current scale

## API Integration

- All API calls via `@/lib/api.ts`
- Use `ModelContext` for model ID propagation
- Handle loading, error, and empty states for all data fetches

## Routing

- React Router with route definitions in `App.tsx`
- Views as pages under `pages/views/`

## Build Validation

- `npx tsc --noEmit` MUST pass (type safety)
- `npx vite build` MUST pass (JSX transformation)
- Both checks are mandatory — TSC alone misses build errors
