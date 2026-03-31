# Frontend Anti-Patterns — Do NOT Do These

## 1. Manual Data Fetching (useEffect + useState)
```tsx
// WRONG — manual fetch with state management
const [data, setData] = useState(null)
const [loading, setLoading] = useState(true)
const [error, setError] = useState(null)
useEffect(() => {
  api.getData(id).then(setData).catch(setError).finally(() => setLoading(false))
}, [id])

// CORRECT — TanStack Query
const { data, isLoading, error } = useQuery({
  queryKey: ['myData', id],
  queryFn: () => viewsApi.getData(id),
  enabled: !!id,
})
```

## 2. Inline Style Objects
```tsx
// WRONG — inline styles
<div style={{ background: 'linear-gradient(...)', borderRadius: 12, padding: 24 }}>

// WRONG — style constant objects
const CARD_SHELL = { background: '#1a1a2e', borderRadius: '12px' }
<div style={CARD_SHELL}>

// CORRECT — Tailwind classes
<div className="bg-card rounded-lg p-6">
```

## 3. `as unknown as` Type Casts
```tsx
// WRONG — casting hides bugs
const data = response as unknown as MyType

// CORRECT — type the API response properly
const data: MyType = await viewsApi.getTypedData(id)
```

## 4. Monolith Files (> 300 lines for pages, > 200 for components)
```tsx
// WRONG — 1000+ line view file with inline sub-components
export function BigView() {
  const InlineCard = () => { /* 80 lines */ }
  const InlineTable = () => { /* 120 lines */ }
  // ... 800 more lines
}

// CORRECT — extract to separate files
// features/big-view/BigViewCard.tsx
// features/big-view/BigViewTable.tsx
// pages/views/BigView.tsx (thin orchestrator, ~200 lines)
```

## 5. Raw Fetch Instead of API Client
```tsx
// WRONG — raw fetch in component
const res = await fetch(`/api/v1/models/${id}/teams`)
const data = await res.json()

// CORRECT — use typed API service
import { viewsApi } from '@/services/api'
const data = await viewsApi.getTeams(id)
```

## 6. Warning Icons Without Explanation
```tsx
// WRONG — icon without context
<AlertTriangle className="text-amber-500" />

// CORRECT — icon with explanation text
<AlertTriangle className="text-amber-500" size={14} />
<span className="text-xs text-amber-600">Fragmented across 3 teams</span>
```

## 7. Invisible Interactive Elements
```tsx
// WRONG — hidden until hover (not discoverable)
<button className="opacity-0 group-hover:opacity-100">Edit</button>

// CORRECT — subtle but visible
<button className="opacity-35 hover:opacity-100 transition-opacity">Edit</button>
```

## 8. Floating Panels Without Dismiss
```tsx
// WRONG — panel stays open forever
{showPanel && <div className="fixed top-4 right-4"><Panel /></div>}

// CORRECT — backdrop for click-outside dismiss
{showPanel && (
  <>
    <div className="fixed inset-0 z-30" onClick={() => setShowPanel(false)} />
    <div className="fixed top-4 right-4 z-40"><Panel /></div>
  </>
)}
```

## 9. Missing Empty/Error States
```tsx
// WRONG — assumes data always exists
return <div>{items.map(i => <Card key={i.id} {...i} />)}</div>

// CORRECT — handles all states
if (isLoading) return <LoadingState />
if (error) return <ErrorState message={error.message} />
if (!data?.items.length) return <EmptyState message="No items found" />
return <div>{data.items.map(i => <Card key={i.id} {...i} />)}</div>
```

## 10. Hand-Rolling What React Flow Provides
```tsx
// WRONG — custom SVG layout engine with manual zoom/pan
<svg onMouseDown={handlePan} onWheel={handleZoom}>
  <path d={computeEdgePath(source, target)} />
  <g transform={`translate(${x},${y})`}>{/* node */}</g>
</svg>

// CORRECT — React Flow handles zoom, pan, edges, node positioning
<ReactFlow
  nodes={nodes}
  edges={edges}
  nodeTypes={customNodeTypes}
  fitView
/>
```

## 11. Duplicating Shared Components
```tsx
// WRONG — every view reimplements stat cards
const StatCard = ({ label, value }) => (
  <div style={{ background: '#1e1e3a', borderRadius: 12, padding: 16 }}>
    <span style={{ fontSize: 12 }}>{label}</span>
    <span style={{ fontSize: 28 }}>{value}</span>
  </div>
)

// CORRECT — use the shared component
import { StatCard } from '@/components/ui/stat-card'
<StatCard label={label} value={value} trend={trend} />
```

## 12. Creating New Context for Server Data
```tsx
// WRONG — wrapping API data in a Context provider
const InsightsContext = createContext<InsightsData | null>(null)
export function InsightsProvider({ children }) {
  const [insights, setInsights] = useState(null)
  useEffect(() => { /* fetch and poll */ }, [])
  return <InsightsContext.Provider value={insights}>{children}</InsightsContext.Provider>
}

// CORRECT — TanStack Query handles caching, refetching, and sharing
const { data: insights } = useQuery({
  queryKey: ['insights', modelId, page],
  queryFn: () => insightsApi.getInsights(modelId, page),
  refetchInterval: 30_000,
})
```
