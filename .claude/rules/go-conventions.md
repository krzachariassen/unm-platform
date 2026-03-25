# Go Conventions

## Naming

- Exported types: PascalCase (`TeamType`, `UNMModel`)
- Internal: camelCase
- Interfaces named by behavior: `Validator`, `Parser`, `Analyzer`
- No `I` prefix on interfaces

## File Structure

- One file per entity/type
- Tests co-located: `actor_test.go` next to `actor.go`
- Packages named by Clean Architecture layer, not by feature

## Error Handling

- Return errors, don't panic
- Wrap errors with context: `fmt.Errorf("parsing actor %s: %w", name, err)`
- Use sentinel errors for domain-level expected failures

## Testing

- Use `testing` + `testify/assert` + `testify/require`
- Table-driven tests for multiple cases
- Test file names: `*_test.go` in same package

## HTTP Handlers

- stdlib `net/http` with Go 1.22+ method routing: `mux.HandleFunc("GET /api/v1/resource", handler)`
- Handlers are thin: parse request, call usecase, write response
- Middleware for CORS, logging, recovery
