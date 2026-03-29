# Example: Extending the YAML Parser

## Goal

Add support for parsing an `outcome` field on Needs.

## File Location

`backend/internal/infrastructure/parser/yaml_parser.go`

## Phase 1: Red — Write the Failing Test

`backend/internal/infrastructure/parser/yaml_parser_test.go`:

```go
func TestYAMLParser_NeedOutcome(t *testing.T) {
    input := `
system:
  name: test
actors:
  - name: User
needs:
  - name: Browse catalog
    actor: User
    outcome: Find relevant products quickly
    supportedBy:
      - Search
capabilities:
  - name: Search
    visibility: user-facing
    realizedBy:
      - search-svc
services:
  - name: search-svc
    ownedBy: search-team
teams:
  - name: search-team
    type: stream-aligned
`
    model, err := ParseYAML([]byte(input))
    require.NoError(t, err)
    require.Contains(t, model.Needs, "Browse catalog")
    assert.Equal(t, "Find relevant products quickly", model.Needs["Browse catalog"].Outcome)
}
```

## Phase 2: Green — Minimum Implementation

In `yaml_parser.go`, find the `yamlNeed` struct and add the field:

```go
type yamlNeed struct {
    Name        string   `yaml:"name"`
    Actor       string   `yaml:"actor"`
    Outcome     string   `yaml:"outcome"`      // NEW
    SupportedBy []string `yaml:"supportedBy"`
}
```

In the transformation logic, map it:

```go
need := &entity.Need{
    Name:        yn.Name,
    Actor:       yn.Actor,
    Outcome:     yn.Outcome,  // NEW — optional, may be empty
    SupportedBy: yn.SupportedBy,
}
```

## Key Rules

- Optional fields: always check for nil/empty before dereferencing
- Add test fixture in `testdata/` if the feature is complex
- Deprecated fields are silently ignored (never error)
- Both `yaml_parser.go` and `yaml_parser_test.go` are edited together
