# Code Reviewer -- Operational Memory

> **Policy**: 30-entry cap · Monthly curation (Promote / Keep / Prune)
> See `.claude/agents/AGENT_OWNERSHIP.md` §2 for full curation rules.

## Common Issues Found

### Architecture Violations
- View-specific logic leaking into domain entities
- HTTP handlers doing computation instead of delegating to use cases
- Frontend computing derived values that should come from the backend presenter

### Test Gaps
- AI tests accidentally mocked (must use real OpenAI)
- Missing edge case tests for 0-element collections
- Changeset actions tested for happy path but not for invalid inputs

### Frontend Patterns
- TSC passes but Vite build fails (adjacent JSX in .map callbacks)
- opacity:0 used for action buttons (should be opacity:0.35 minimum)
- Warning icons rendered without explanation text

## Review Flags (auto-flag these)
- Any import from `internal/domain/` that references outer layers
- Any `httptest.NewServer` in AI test files (mocking violation)
- Any `fetch()` call in frontend that doesn't go through `services/api/`
- Any `opacity-0` or `opacity: 0` on interactive elements
- Any floating panel without a backdrop dismiss handler
- Any `useEffect` + `useState` pattern for data fetching (should use TanStack Query)
- Any `as unknown as` type cast (should fix the type at the source)
- Any inline `style={{}}` when Tailwind classes exist for the same purpose
- Any page/view file exceeding 300 lines
- Any component file exceeding 200 lines
- Any types defined in component files instead of `types/`
- Any hand-rolled SVG graph rendering instead of React Flow

## Confirmed Safe Patterns (do not re-flag)
- `httptest.NewServer` in `openai_client_test.go` is used for HTTP error-handling tests (non-200, empty choices, request inspection), NOT for mocking AI responses. These are labeled with comments explaining the distinction and are acceptable per the "no mocking real AI responses" rule. The real-API test is gated on `UNM_AI_TESTS=true`.
- `fetch(` in `model-context.tsx` is a deliberate exception: the comment "Use a direct fetch (not api.ts) to avoid circular import issues" at line 52 is architecturally necessary for the hydration verification ping.
- `fetch(` in `runtimeConfig.ts` is used for a one-shot config endpoint `/api/config` before the API module is available — also an acceptable exception.
- `ChangeAction` in `internal/domain/entity/changeset.go` carries JSON tags (`json:"..."`) because it is deserialized from HTTP request bodies via the changeset handler. This is an intentional crossing of the adapter concern into the domain entity for the changeset action struct; it is a known trade-off in this design.
- CORS middleware defaults to `"*"` when no origins are configured; this is safe because in production `config/production.yaml` should set `server.cors_origins` to restrict origins. The `DefaultConfig()` already seeds `["http://localhost:5173"]`.

## Patterns Discovered in 2026-03 Full Audit
- [2026-03-29] api.ts URL inconsistency: `getNeedView` and `getCapabilityView` use hardcoded `/api/models/...` instead of `${API_BASE}`. All other methods use `API_BASE`. This will break if backend is deployed at a non-root path.
- [2026-03-29] `time.Sleep` in `model_store_test.go` (5ms) is used to test that `LastAccessedAt` is updated — this is the only reliable way to verify monotonic time update and is a narrow acceptable exception.
- [2026-03-29] `time.Sleep(100ms)` in `ai_advisor_test.go` is inside a real-API test gate (`UNM_AI_TESTS=true`) and is a rate-limit courtesy delay, not a synchronization sleep. Acceptable.
- [2026-03-29] No presenter layer exists (architecture doc mentions `adapter/presenter/` but no such package is present). View model construction lives in the handler package in `view_*.go` files. This is an informal presenter pattern inside the adapter layer — technically correct layering, but not following the declared architecture layout.
- [2026-03-29] `insightEntry` type assertion `raw.(insightEntry)` in `insights.go` is unsafe — if the cache is ever written with a different type, this will panic. A checked assertion `raw, ok := ...; if !ok { ... }` would be safer.
- [2026-03-29] `w.Write(data)` in `model.go:160` ignores the error return from Write. For binary/YAML responses after `WriteHeader(200)` this is a common minor issue.
- [2026-03-29] `DebugRoutes: true` in `DefaultConfig()` means debug routes are ON by default in all environments that don't have a config file override. Production config must explicitly set `features.debug_routes: false` or it will expose the debug endpoints.
