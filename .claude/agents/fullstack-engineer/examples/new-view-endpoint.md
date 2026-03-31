# Example: Adding a New View Endpoint

## Goal

Add a "Team Health" view showing cognitive load signals per team.

## Phase 1: Define the API Contract

Define the exact shapes before writing any code.

**Go response struct** (`backend/internal/adapter/handler/`):

```go
type TeamHealthResponse struct {
    Teams []TeamHealthEntry `json:"teams"`
}

type TeamHealthEntry struct {
    Name           string  `json:"name"`
    Type           string  `json:"type"`
    ServiceCount   int     `json:"service_count"`
    CognitiveLoad  string  `json:"cognitive_load"`
    Score          float64 `json:"score"`
}
```

**TypeScript mirror** (`frontend/src/types/views.ts`):

```typescript
interface TeamHealthResponse {
    teams: TeamHealthEntry[]
}

interface TeamHealthEntry {
    name: string
    type: string
    service_count: number
    cognitive_load: string
    score: number
}
```

## Phase 2: Backend (TDD)

1. Write failing test for the handler
2. Register route: `mux.HandleFunc("GET /api/v1/models/{id}/views/team-health", h.HandleTeamHealthView)`
3. Implement handler (thin — calls presenter)
4. Implement presenter logic to compute team health from model
5. Run `cd backend && go test ./... && go vet ./...`

Key: initialize slices with `make([]TeamHealthEntry, 0)` to avoid `null` JSON.

## Phase 3: Frontend

1. Add types in `frontend/src/types/views.ts`
2. Add API function in `frontend/src/services/api/views.ts`
3. Create `frontend/src/pages/views/TeamHealthView.tsx` using TanStack Query (`useQuery`)
4. Add route in `App.tsx`
5. Add nav item in `Sidebar.tsx`
6. Wrap page with `<ModelRequired>`
7. Handle loading, error, and empty states via shared components

## Phase 4: Integration Validation

```bash
cd backend && go test ./... && go vet ./...
cd frontend && npx tsc --noEmit && npx vite build
```

All four commands must pass with zero errors.

## Key Rules

- `snake_case` JSON keys always
- Empty slices, not null (`make([]T, 0)`)
- No frontend re-computation — if backend has the data, return it
- Backend tests before frontend work
