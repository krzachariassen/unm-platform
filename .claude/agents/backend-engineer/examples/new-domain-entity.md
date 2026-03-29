# Example: Adding a New Domain Entity

## Goal

Add a `Platform` entity that groups platform teams.

## File Location

`backend/internal/domain/entity/platform.go`

## Rules

- No JSON tags (JSON is an adapter concern)
- No imports from outer layers (adapter, infrastructure, net/http)
- Constructor validates invariants and returns error on invalid state

## Phase 1: Red — Write the Failing Test

`backend/internal/domain/entity/platform_test.go`:

```go
func TestNewPlatform_Valid(t *testing.T) {
    p, err := entity.NewPlatform("data-platform", []string{"data-team", "infra-team"})
    require.NoError(t, err)
    assert.Equal(t, "data-platform", p.Name)
    assert.Len(t, p.Teams, 2)
}

func TestNewPlatform_EmptyName(t *testing.T) {
    _, err := entity.NewPlatform("", []string{"team-a"})
    require.Error(t, err)
    assert.Contains(t, err.Error(), "name")
}

func TestNewPlatform_NoTeams(t *testing.T) {
    _, err := entity.NewPlatform("platform", nil)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "teams")
}
```

## Phase 2: Green — Minimum Implementation

```go
package entity

import "fmt"

type Platform struct {
    Name  string
    Teams []string
}

func NewPlatform(name string, teams []string) (*Platform, error) {
    if name == "" {
        return nil, fmt.Errorf("platform name must not be empty")
    }
    if len(teams) == 0 {
        return nil, fmt.Errorf("platform must have at least one team")
    }
    return &Platform{Name: name, Teams: teams}, nil
}
```

## Verification Checklist

- [ ] File in `internal/domain/entity/` — one file per entity
- [ ] No JSON tags on the struct
- [ ] No imports from outer layers
- [ ] Constructor validates invariants
- [ ] Tests cover: valid, empty name, no teams
