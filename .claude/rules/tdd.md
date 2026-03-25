# TDD Protocol

## Cycle

1. **Red** — Write a failing test that defines the expected behavior
2. **Green** — Write the minimum code to make it pass
3. **Refactor** — Clean up while keeping tests green

## Rules

- No production code without a failing test first
- Run `go test ./...` (backend) or `npx tsc --noEmit && npx vite build` (frontend) after every change
- Tests co-located with code: `foo.go` → `foo_test.go`
- Use table-driven tests for Go when testing multiple cases
- Use testify assertions (`assert`, `require`) for clarity

## AI Tests (special rule)

AI tests MUST use the real OpenAI API. No mocking.
Tests skip gracefully with `t.Skip` when `UNM_OPENAI_API_KEY` is not set.
Source `ai.env` before running AI tests.
