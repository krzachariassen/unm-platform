# Example: Adding a New Analyzer

## Goal

Add an analyzer that detects teams owning too many services (service sprawl).

## File Location

`backend/internal/infrastructure/analyzer/service_sprawl.go`

## Phase 1: Red — Write the Failing Test

`backend/internal/infrastructure/analyzer/service_sprawl_test.go`:

```go
func TestServiceSprawlAnalyzer_DetectsOverloadedTeams(t *testing.T) {
    model := testutil.LoadModel(t, "testdata/sprawl.unm.yaml")

    analyzer := NewServiceSprawlAnalyzer()
    signals := analyzer.Analyze(model)

    require.Len(t, signals, 1)
    assert.Equal(t, "service-sprawl", signals[0].ID)
    assert.Equal(t, entity.SeverityWarning, signals[0].Severity)
    assert.Contains(t, signals[0].Evidence, "platform-team")
}

func TestServiceSprawlAnalyzer_NoFalsePositives(t *testing.T) {
    model := testutil.LoadModel(t, "testdata/simple.unm.yaml")

    analyzer := NewServiceSprawlAnalyzer()
    signals := analyzer.Analyze(model)

    assert.Empty(t, signals)
}
```

Run `go test ./internal/infrastructure/analyzer/...` — should fail (function
doesn't exist yet).

## Phase 2: Green — Minimum Implementation

```go
type ServiceSprawlAnalyzer struct {
    threshold int
}

func NewServiceSprawlAnalyzer() *ServiceSprawlAnalyzer {
    return &ServiceSprawlAnalyzer{threshold: 8}
}

func (a *ServiceSprawlAnalyzer) Analyze(model *entity.UNMModel) []entity.Signal {
    teamServiceCount := make(map[string]int)
    for _, svc := range model.Services {
        teamServiceCount[svc.OwnedBy]++
    }

    var signals []entity.Signal
    for team, count := range teamServiceCount {
        if count > a.threshold {
            signals = append(signals, entity.Signal{
                ID:          "service-sprawl",
                Category:    "team-health",
                Severity:    entity.SeverityWarning,
                Title:       fmt.Sprintf("Team %q owns %d services", team, count),
                Description: "Teams owning too many services face high cognitive load.",
                Evidence:    []string{team},
            })
        }
    }
    return signals
}
```

## Phase 3: Refactor

- Extract threshold to config if needed
- Ensure signal ID is unique per team (append team name)

## Key Rules

- Analyzer lives in `infrastructure/analyzer/` — never in domain
- Returns `[]entity.Signal` using the standard Signal struct
- Test uses real model fixtures from `testdata/`, never fabricated structs
- Both positive (signals found) and negative (no false positives) tests
