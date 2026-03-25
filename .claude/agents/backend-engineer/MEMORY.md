# Backend Engineer — Operational Memory

## Learnings

### 2025-03 — AI tests must use real OpenAI
All AI tests must call the real OpenAI API. Do NOT mock responses.
Use `t.Skip("UNM_OPENAI_API_KEY not set")` when the key is unavailable.
Source `ai.env` before running AI tests.

### 2025-03 — Changeset system
The changeset system uses `ChangeActionType` constants in `domain/entity/changeset.go`.
`ChangesetApplier` in `domain/service/changeset_applier.go` applies actions to the in-memory model.
`ModelStore.Replace()` swaps the model after commit. Always clear caches after replace.

### 2025-03 — YAML serializer
`infrastructure/serializer/yaml_serializer.go` handles export to DSL-compliant YAML.
Must handle both short-form and long-form relationship representations.

## Known Gotchas

- Unused variables in Go cause compilation failures — clean up test helpers
- `go test ./...` runs all packages including CLI — make sure CLI compiles
- View presenter builds view-specific DTOs — don't pollute domain entities with view concerns
