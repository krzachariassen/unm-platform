# Fullstack Anti-Patterns — Do NOT Do These

## 1. API Contract Mismatch (Go JSON vs TypeScript)
```go
// WRONG — Go field serializes as camelCase mismatch or wrong name
type TeamDTO struct {
    TeamName string `json:"team_name"` // frontend expects teamName
}
```
```ts
// WRONG — TypeScript does not match encoded JSON
interface TeamDTO {
  teamName: string // API actually sends team_name
}

// CORRECT — align json tags with frontend types (or shared OpenAPI)
// Go: `json:"teamName"`
// TS: teamName: string
```

## 2. Frontend Computing Derived Data the Backend Already Has
```ts
// WRONG — duplicating server logic in the UI
const fragmented = capabilities.filter(
  (c) => new Set(c.teams).size > 2
)

// CORRECT — backend exposes the derived field; UI displays it
interface CapabilityView {
  isFragmented: boolean
  // ...
}
```

## 3. Starting Frontend Before Backend API Is Verified
```md
WRONG — build UI against assumed shapes/endpoints
- Add React route and mock data
- Assume POST /api/v1/models/:id/analyze returns { signals: [...] }
- Discover at integration time that the handler is missing or shape differs

CORRECT — verify contract first, then wire UI
- Run handler test or curl against real route
- Confirm JSON matches TypeScript types (or generate types from OpenAPI)
- Then implement components that call the verified api.ts methods
```
