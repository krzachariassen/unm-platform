# Frontend Anti-Patterns — Do NOT Do These

## 1. Warning Icons Without Explanation
```tsx
// WRONG — icon without context
<AlertTriangle className="text-amber-500" />

// CORRECT — icon with explanation text
<AlertTriangle className="text-amber-500" size={14} />
<span className="text-xs text-amber-600">Fragmented across 3 teams</span>
```

## 2. Invisible Interactive Elements
```tsx
// WRONG — hidden until hover (not discoverable)
<button className="opacity-0 group-hover:opacity-100">Edit</button>

// CORRECT — subtle but visible
<button className="opacity-35 hover:opacity-100 transition-opacity">Edit</button>
```

## 3. Floating Panels Without Dismiss
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

## 4. Raw Fetch Instead of API Client
```tsx
// WRONG
const res = await fetch(`/api/v1/models/${id}/teams`)
const data = await res.json()

// CORRECT
import * as api from '@/lib/api'
const data = await api.getTeams(modelId)
```

## 5. Missing Empty/Error States
```tsx
// WRONG — assumes data always exists
return <div>{items.map(i => <Card key={i.id} {...i} />)}</div>

// CORRECT — handles all states
if (loading) return <Spinner />
if (error) return <ErrorMessage error={error} />
if (items.length === 0) return <EmptyState message="No items found" />
return <div>{items.map(i => <Card key={i.id} {...i} />)}</div>
```
