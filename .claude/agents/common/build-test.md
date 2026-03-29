# Build & Test Commands

## Backend (Go)

```bash
cd backend && go test ./...              # Run all tests
cd backend && go test ./... -v           # Verbose
cd backend && go test ./... -cover       # With coverage
cd backend && go build ./cmd/server/     # Build API server
cd backend && go build ./cmd/cli/        # Build CLI
cd backend && go run ./cmd/server/       # Run server (port 8080)
```

## AI Tests (require API key)

```bash
source ai.env && cd backend && go test ./internal/infrastructure/ai/... -v -timeout 10m
```

AI tests skip gracefully when UNM_OPENAI_API_KEY is not set.

## Frontend (React/TypeScript)

```bash
cd frontend && npm install               # Install deps
cd frontend && npm run build             # Production build: tsc -b && vite build (MUST pass; matches CI)
cd frontend && npm run dev               # Dev server (port 5173)
```

## Important

- Use `npm run build` for frontend validation — same script as Dockerfile/CI (`tsc -b && vite build`).
- Backend AI tests skip gracefully when UNM_OPENAI_API_KEY is not set.
