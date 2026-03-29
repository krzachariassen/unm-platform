# Example: API Contract Design

## The Contract Pattern

Define Go response struct AND TypeScript interface BEFORE writing implementation.

```go
// Go — backend/internal/adapter/handler/
type FeatureResponse struct {
    Items []FeatureItem `json:"items"`
    Total int           `json:"total"`
}

type FeatureItem struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Score float64 `json:"score"`
}
```

```typescript
// TypeScript — frontend/src/lib/api.ts
interface FeatureResponse {
    items: FeatureItem[]
    total: number
}

interface FeatureItem {
    id: string
    name: string
    score: number
}
```

## Common Pitfalls

### 1. omitempty Causes null

```go
// WRONG — empty slice becomes null in JSON
type Response struct {
    Items []Item `json:"items,omitempty"`
}

// CORRECT — initialize the slice
items := make([]Item, 0)
```

TypeScript receives `null` instead of `[]`, causing `.map()` to crash.

### 2. Integer Enums

```go
// WRONG — TypeScript gets opaque numbers
type TeamType int
const StreamAligned TeamType = 1

// CORRECT — TypeScript gets readable strings
type TeamType string
const StreamAligned TeamType = "stream-aligned"
```

### 3. Inconsistent Casing

```go
// WRONG — Go field is PascalCase, JSON is snake_case, but TS uses camelCase
type Entry struct {
    ServiceCount int `json:"serviceCount"` // camelCase in JSON
}

// CORRECT — snake_case everywhere in JSON
type Entry struct {
    ServiceCount int `json:"service_count"` // snake_case in JSON
}
```

TypeScript property must match exactly: `service_count`, not `serviceCount`.

### 4. Frontend Re-computation

```typescript
// WRONG — computing ratio in frontend
const ratio = team.services / team.size

// CORRECT — backend returns the ratio
const ratio = team.service_ratio  // already computed by presenter
```

If the backend has the data, return it. The frontend should display, not derive.
