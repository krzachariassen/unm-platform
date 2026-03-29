# Pre-Submission Safety Checklist

Before declaring any task complete, verify ALL of the following:

## Backend Changes

- [ ] `cd backend && go build ./cmd/server/` succeeds
- [ ] `cd backend && go build ./cmd/cli/` succeeds
- [ ] `cd backend && go test ./...` passes (all tests)
- [ ] No new `go vet` warnings
- [ ] Tests exist for new/changed code (TDD — tests written first)

## Frontend Changes

- [ ] `cd frontend && npm run build` passes (zero errors; runs `tsc -b` + `vite build`)
- [ ] No linter errors in edited files

## Git Flow

- [ ] NOT on `main` — work is on a feature branch (`git branch --show-current`)
- [ ] Branch name follows convention: `<type>/<short-description>`
- [ ] Commits are on the feature branch, not main

## Both

- [ ] No hardcoded secrets or API keys in committed code
- [ ] Changes are minimal and focused on the task
- [ ] Commit message follows: `<type>(<scope>): <description>`
  Types: feat, fix, refactor, test, docs, chore
  Scopes: domain, parser, validator, api, frontend, ai, changeset
- [ ] Branch pushed: `git push -u origin HEAD`
