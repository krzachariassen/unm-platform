# Code Reviewer -- Operational Memory

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
- Any `fetch()` call in frontend that doesn't go through `api.ts`
- Any `opacity-0` or `opacity: 0` on interactive elements
- Any floating panel without a backdrop dismiss handler
