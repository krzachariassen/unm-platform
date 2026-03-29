# Backend Engineer Checklists

These checklists are **extracted from** `.claude/agents/backend-engineer/AGENT.md` for quick reference and for tooling that points agents at standalone files.

> **Note:** They align with the repository **validation pipeline** (see `.claude/commands/validate.md`): backend gates include `go build`, `go vet`, and `go test ./...`.

## New Domain Entity

- [ ] File in `internal/domain/entity/` — one file per entity
- [ ] No JSON tags on the struct (JSON is adapter concern)
- [ ] No imports from outer layers (adapter, infrastructure, net/http, etc.)
- [ ] Constructor function validates invariants and returns error on invalid state
- [ ] Test covers: valid construction, boundary conditions, invalid inputs
- [ ] All exported fields and methods documented

## New API Endpoint

- [ ] Handler in `internal/adapter/handler/` — thin, delegates to use case
- [ ] Route registered in server setup with Go 1.22+ method routing: `mux.HandleFunc("GET /api/v1/resource", h.Handler)`
- [ ] Request parsing: check content type, decode body, validate required fields
- [ ] Response struct defined in handler or presenter, with proper json tags
- [ ] Error responses use consistent structure: `{"error": "message"}`
- [ ] Handler does NOT call domain services directly — routes through use case
- [ ] Test file covers: happy path, missing model ID, invalid input, not found

## New Analyzer

- [ ] Lives in `internal/infrastructure/analyzer/`
- [ ] Implements the `Analyzer` interface (or relevant interface from use case layer)
- [ ] Returns `[]Signal` (see domain/entity/signal.go for Signal structure)
- [ ] Each signal has: ID, Type, Severity, Title, Description, Evidence (concrete entity names)
- [ ] Test uses realistic model fixture from `testdata/`
- [ ] Test verifies both: signals found when expected, no false positives on clean model

## New Parser Extension

- [ ] YAML parser in `internal/infrastructure/parser/yaml_parser.go`
- [ ] DSL parser in `internal/infrastructure/parser/dsl_parser.go` (PEG-based)
- [ ] Add test case in `yaml_parser_test.go` covering the new syntax
- [ ] Add/update test fixture in `testdata/` if needed
- [ ] Gracefully handle optional fields (nil check before dereferencing)
- [ ] Deprecated fields silently ignored (not errored)

## Error Handling

- [ ] Return errors, never panic (except truly unrecoverable program errors)
- [ ] Wrap errors with context: `fmt.Errorf("parsing actor %q: %w", name, err)`
- [ ] Sentinel errors for domain-level failures: `var ErrActorNotFound = errors.New("actor not found")`
- [ ] HTTP handlers translate domain errors to appropriate HTTP status codes

## Test Quality

- [ ] Tests use `testify/require` for fatal assertions (stops test on failure)
- [ ] Tests use `testify/assert` for non-fatal assertions (collects all failures)
- [ ] Table-driven tests for functions with multiple input variants
- [ ] Test names describe behavior: `TestYAMLParser_MissingSystemName`, not `TestParse1`
- [ ] No `time.Sleep` in tests — use deterministic inputs
