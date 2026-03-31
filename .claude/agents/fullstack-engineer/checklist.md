# Fullstack Engineer Checklist

Use this for vertical slices that touch **both** `backend/` and `frontend/`. For narrative detail, see `.claude/agents/fullstack-engineer/AGENT.md` and `examples/`.

> **Note:** Full validation matches the **validation pipeline** in `.claude/commands/validate.md` (backend build/vet/test + frontend `npm run build`).

## API Contract

- [ ] Go response struct(s) written with `json` tags — **snake_case** keys
- [ ] TypeScript interface(s) in `frontend/src/types/` mirror the Go shape exactly
- [ ] Empty slices initialized in Go where the UI expects arrays (`[]` not `null` / omitted)
- [ ] Enums represented as **strings** in JSON unless there is a strong reason not to
- [ ] No derived fields left for the frontend when the backend already has the data

## Backend Implementation

- [ ] Domain changes (if any): no outer-layer imports; constructor/tests per entity rules
- [ ] Use case or domain service invoked from handler — handler stays thin
- [ ] Presenter used when response shaping is non-trivial (`internal/adapter/presenter/`)
- [ ] Route registered in server setup; error body `{"error": "..."}` on failures
- [ ] `cd backend && go test ./... && go vet ./...` passes

## Frontend Implementation

- [ ] Types in `types/` and API function in `services/api/` (no ad-hoc `fetch` elsewhere)
- [ ] Page or component implements loading, error, and empty states
- [ ] New page: route added in `App.tsx` and nav link in `Sidebar.tsx` if applicable
- [ ] Page that needs a loaded model wrapped with `<ModelRequired>` (or project equivalent)
- [ ] `cd frontend && npm run build` passes

## Integration

- [ ] Both suites green: `go test ./...`, `go vet ./...`, `cd frontend && npm run build`
- [ ] JSON keys in actual responses match TypeScript property names (snake_case)
- [ ] No duplicate business logic in React that belongs in presenter/use case
