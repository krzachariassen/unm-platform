# Pre-Submission Safety Checklist

Before declaring any task complete, verify ALL of the following:

## Backend Changes

- [ ] `cd backend && go build ./cmd/server/` succeeds
- [ ] `cd backend && go build ./cmd/cli/` succeeds
- [ ] `cd backend && go test ./...` passes (all tests)
- [ ] No new `go vet` warnings
- [ ] Tests exist for new/changed code (TDD — tests written first)

## Frontend Changes

- [ ] `cd frontend && npx tsc --noEmit` passes (zero errors)
- [ ] `cd frontend && npx vite build` succeeds (zero errors)
- [ ] No linter errors in edited files

## Both

- [ ] No hardcoded secrets or API keys in committed code
- [ ] Changes are minimal and focused on the task
- [ ] Commit message follows: `<type>(<scope>): <description>`
  Types: feat, fix, refactor, test, docs, chore
  Scopes: domain, parser, validator, api, frontend, ai, changeset
