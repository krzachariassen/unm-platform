# Validation Task

Run the full validation suite for the UNM Platform.

## Backend Validation

```bash
cd backend && go vet ./...
cd backend && go test ./...
cd backend && go build ./cmd/server/
cd backend && go build ./cmd/cli/
```

## Frontend Validation

```bash
cd frontend && npx tsc --noEmit
cd frontend && npx vite build
```

## Integration Validation

If servers can be started:
```bash
cd backend && go run ./cmd/server/ &
cd frontend && npm run dev &
# Wait for both to start, then:
curl -s http://localhost:8080/api/v1/health
curl -s http://localhost:5173
```

## Report

For each validation step, report:
- PASS or FAIL
- If FAIL: exact error message and file:line
- Summary: total checks, passed, failed

$ARGUMENTS
