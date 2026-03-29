# Code Review Anti-Patterns — Do NOT Do These

## 1. View-Specific Logic in Domain Entities
```go
// WRONG — domain knows about presentation
type Capability struct {
    ID          string
    BadgeColor  string // ← view concern
    ShowWarning bool   // ← view concern
}

// CORRECT — domain holds facts; presenters map to UI
type Capability struct {
    ID string
    // ...
}
// Color/warning derived in application layer or API DTO
```

## 2. HTTP Handlers Doing Computation Instead of Delegating
```go
// WRONG — orchestration and business rules in handler
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
    model := store.Get(id)
    var out []Signal
    for _, s := range model.Services {
        if len(s.DependsOn) > 4 {
            out = append(out, Signal{Type: "complexity", Target: s.ID})
        }
    }
    json.NewEncoder(w).Encode(out)
}

// CORRECT — inject use case / service
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
    result := analyzer.Analyze(ctx, id)
    json.NewEncoder(w).Encode(result)
}
```

## 3. Frontend Computing Derived Values That Belong in the Backend
```ts
// WRONG — business rule only in client
const riskScore = teams.length * capabilities.filter((c) => c.critical).length

// CORRECT — API returns computed projections
const { riskScore, summary } = await api.getModelProjection(modelId)
```

## 4. AI Tests Accidentally Mocked
```go
// WRONG — httptest fake OpenAI violates project rule
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(fakeChatCompletion)
}))
t.Cleanup(server.Close)

// CORRECT — real client, skip if no key
if os.Getenv("UNM_OPENAI_API_KEY") == "" {
    t.Skip("UNM_OPENAI_API_KEY not set")
}
client := ai.NewOpenAIClient()
```

## 5. TSC Passes but Vite Build Fails (Adjacent JSX in .map)
```tsx
// WRONG — implicit array of adjacent JSX nodes inside map callback
{items.map((item) => (
  <div key={item.id}>{item.name}</div>
  <span className="text-muted">{item.hint}</span>
))}

// CORRECT — single parent per iteration
{items.map((item) => (
  <div key={item.id}>
    <div>{item.name}</div>
    <span className="text-muted">{item.hint}</span>
  </div>
))}
```

## 6. opacity: 0 on Action Buttons
```tsx
// WRONG — not discoverable, hurts accessibility
<button className="opacity-0 group-hover:opacity-100">Edit</button>

// CORRECT — subtle but always visible
<button className="opacity-35 hover:opacity-100 transition-opacity">Edit</button>
```

## 7. Warning Icons Without Explanation Text
```tsx
// WRONG
<AlertTriangle className="text-amber-500" />

// CORRECT
<AlertTriangle className="text-amber-500" size={14} />
<span className="text-xs text-amber-600">Fragmented across 3 teams</span>
```

## 8. Rubber-Stamping Reviews
```md
WRONG — no verification
"Looks good to me!"

CORRECT — tie comments to evidence
- Checked handler delegates to AnalyzerService (handler_test.go:42)
- Ran `go test ./...` and `npm run build` on this branch
- Open question: should empty `dependsOn` be validated? Not blocking.
```
