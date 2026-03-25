# Backend Anti-Patterns — Do NOT Do These

## 1. Domain Layer Contamination
```go
// WRONG — domain importing infrastructure
package entity
import "gopkg.in/yaml.v3"  // ← NEVER

// CORRECT — domain stays pure
package entity
// No external imports in domain entities
```

## 2. Business Logic in Handlers
```go
// WRONG — handler computing results
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
    model := store.Get(id)
    signals := []Signal{}
    for _, svc := range model.Services {
        if len(svc.DependsOn) > 5 { // ← business logic in handler!
            signals = append(signals, Signal{...})
        }
    }
}

// CORRECT — handler delegates
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
    model := store.Get(id)
    result := analyzer.Analyze(model)
    json.NewEncoder(w).Encode(result)
}
```

## 3. Mocking OpenAI
```go
// WRONG
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(mockResponse)
}))
client := openai.NewClient(server.URL)

// CORRECT
client := ai.NewOpenAIClient() // uses real API key
if os.Getenv("UNM_OPENAI_API_KEY") == "" {
    t.Skip("UNM_OPENAI_API_KEY not set")
}
```

## 4. God Packages
```go
// WRONG — everything in one package
package model
type Actor struct { ... }
type Capability struct { ... }
type Validator struct { ... }
type YAMLParser struct { ... }

// CORRECT — one concept per file, layered packages
// domain/entity/actor.go
// domain/entity/capability.go
// domain/service/validator.go
// infrastructure/parser/yaml_parser.go
```
