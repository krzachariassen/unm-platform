# Frontend Architecture Rules

These rules apply to all code under `frontend/src/`. They mirror the discipline
of the backend's Clean Architecture and TDD conventions.

## Layered Architecture

```
Pages  →  Features  →  Components  →  Hooks  →  Services  →  Types
```

**Dependency direction:** left may import right; right NEVER imports left.

| Layer | Path | Responsibility |
|-------|------|----------------|
| **Pages** | `pages/`, `pages/views/` | Route entry points. Compose features and components. ≤ 300 lines. |
| **Features** | `features/<name>/` | Domain-specific logic for one feature (e.g., `features/unm-map/`, `features/whatif/`). May contain sub-components, hooks, and utils scoped to that feature. |
| **Components** | `components/ui/`, `components/layout/`, `components/<domain>/` | Reusable presentation components. No API calls. Receive data via props. |
| **Hooks** | `hooks/` | Shared custom hooks. Data fetching hooks wrap TanStack Query. |
| **Services** | `services/api/` | API client functions. Pure fetch wrappers — no React, no hooks, no state. |
| **Types** | `types/` | Shared TypeScript interfaces, enums, and type utilities. No runtime code. |

## File Size Limits

| File type | Hard limit | Target |
|-----------|-----------|--------|
| Page / view | 300 lines | 150–250 |
| Component | 200 lines | 80–150 |
| Hook | 100 lines | 40–80 |
| Service file | 150 lines | 60–100 |
| Type file | 200 lines | 50–100 |

If a file exceeds its hard limit, extract sub-components or utilities before merging.

## Data Fetching — TanStack Query

All server state uses TanStack Query. No manual `useEffect` + `useState` + `setLoading` patterns.

```typescript
// CORRECT — TanStack Query
import { useQuery } from '@tanstack/react-query'
import { viewsApi } from '@/services/api'

export function NeedView() {
  const { modelId } = useModel()
  const { data, isLoading, error } = useQuery({
    queryKey: ['needs', modelId],
    queryFn: () => viewsApi.getNeeds(modelId!),
    enabled: !!modelId,
  })
  // ...
}

// WRONG — manual fetch + state
const [data, setData] = useState(null)
const [loading, setLoading] = useState(true)
useEffect(() => {
  api.getNeeds(modelId).then(setData).finally(() => setLoading(false))
}, [modelId])
```

After changeset commits, invalidate affected queries:
```typescript
queryClient.invalidateQueries({ queryKey: ['needs', modelId] })
```

## Styling — Tailwind Only

- Tailwind CSS v4 utility classes for all styling
- CSS variables (`text-muted-foreground`, `bg-background`) for theme tokens
- `cn()` utility from `@/lib/utils` for conditional classes
- shadcn/ui components for standard UI patterns
- **NEVER** use inline `style={{}}` except for dynamic computed values (e.g., React Flow node positions, chart pixel offsets)
- **NEVER** create style constant objects (`const CARD_SHELL = { background: '...' }`)
- **NEVER** add `.css` files — Tailwind handles everything

## Graph Visualization — React Flow

UNM Map and any future graph/diagram views use `@xyflow/react` (React Flow):
- Custom node types for domain entities (Actor, Need, Capability, etc.)
- Built-in zoom/pan — do NOT hand-roll mouse event handlers
- Built-in edge rendering — do NOT manually compute SVG paths
- Layout logic in pure functions with unit tests (`features/<name>/layout.ts`)

## Type Safety

- All API response types defined in `types/` — no inline type definitions in components
- Discriminated unions for action types (not flat optionals)
- **NEVER** use `as unknown as` casts — fix the type at the source
- **NEVER** use `any` — use `unknown` with type guards if the shape is truly dynamic
- Prefer `satisfies` over `as` for type narrowing

## API Client — services/api/

The API client is split by domain:
```
services/api/
├── client.ts       — fetch wrapper (base URL, error handling, AbortSignal)
├── models.ts       — parse, export, load example
├── views.ts        — typed view fetchers
├── changesets.ts   — changeset operations
├── insights.ts     — AI insights
├── advisor.ts      — AI advisor
└── index.ts        — re-exports
```

Rules:
- Each function returns a typed response — no generic `getView(name)` that returns `any`
- Error responses throw typed errors (not silent failures)
- All functions accept `AbortSignal` for cancellation
- No React imports — these are pure async functions

## State Management

| State type | Solution |
|------------|----------|
| Server state (API data) | TanStack Query |
| Global client state (model ID, edit mode) | React Context |
| Local UI state (open/closed, selected tab) | `useState` |
| URL state (filters, search) | `useSearchParams` |

Do NOT create new Context providers for data that comes from the server.

## Testing Requirements

- Every pure function (layout algorithms, data transforms, chain traversals) must have unit tests
- Test files co-located: `layout.ts` → `layout.test.ts`
- Use Vitest (already configured)
- Minimum: all files in `features/*/` and `services/api/` must have tests

## Component Patterns

### Shared components receive data via props — no API calls inside
```typescript
// CORRECT
export function StatCard({ label, value, trend }: StatCardProps) { ... }

// WRONG — fetches its own data
export function StatCard({ modelId }: { modelId: string }) {
  const { data } = useQuery(...)
  ...
}
```

### Pages compose features and shared components
```typescript
export function NeedView() {
  const { modelId } = useModel()
  const { data, isLoading, error } = useQuery({ ... })

  if (isLoading) return <LoadingState />
  if (error) return <ErrorState message={error.message} />
  if (!data) return <EmptyState />

  return (
    <ModelRequired>
      <PageHeader title="Needs" description="..." />
      <div className="grid gap-4">
        <StatCard label="Total" value={data.total} />
        {/* ... */}
      </div>
    </ModelRequired>
  )
}
```

## Build Validation

Every change must pass:
```bash
cd frontend && npm run build    # tsc -b + vite build
cd frontend && npm run test     # vitest
```
