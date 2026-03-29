# Backend Engineer — Operational Memory

> **Policy**: 30-entry cap · Monthly curation (Promote / Keep / Prune)
> See `.claude/agents/AGENT_OWNERSHIP.md` §2 for full curation rules.

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

### 2026-03 — HandlerDeps struct
`adapter/handler/handler.go` uses a `HandlerDeps` named-field struct instead of a positional
constructor. When adding new handler dependencies, add a field to `HandlerDeps` and update
`cmd/server/main.go` (the only call site). All handler test files use `newTestHandler(t, deps)`
helper that constructs `HandlerDeps` — update that helper too.

### 2026-03 — View registry
`adapter/handler/view.go` dispatches views via `map[string]viewBuilder` (`viewRegistry`), not a
switch statement. To add a new view type, add an entry to the map — no switch arms to modify.

### 2026-03 — Use case extraction (Phase 6.12)
5 use case services live in `internal/usecase/`: SignalsService, AnalysisRunner, AIContextBuilder,
ChangesetExplainer. AntiPatternDetector lives in `internal/domain/service/`. Handlers are now
thin: they call the use case, not the analyzer directly. CognitiveLoadAnalyzer is injected into
ValueChainAnalyzer via `NewValueChainAnalyzerWithCogLoad()` using a `CognitiveLoadProvider`
interface — do not construct it internally with `DefaultConfig()`.

### 2026-03 — Wave dependency in parallel agent tasks
When two waves of parallel agents share a branch: Wave 2 agents must start AFTER Wave 1 agents
have committed and their changes are on the branch. Wave 2 (HandlerDeps refactor) needed to see
Wave 1's use-case interfaces before it could define HandlerDeps correctly. Always commit Wave 1
work before spawning Wave 2.

## Known Gotchas

- Unused variables in Go cause compilation failures — clean up test helpers
- `go test ./...` runs all packages including CLI — make sure CLI compiles
- View presenter builds view-specific DTOs — don't pollute domain entities with view concerns
