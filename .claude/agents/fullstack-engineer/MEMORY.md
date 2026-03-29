# Fullstack Engineer — Operational Memory

> **Policy**: 30-entry cap · Monthly curation (Promote / Keep / Prune)
> See `.claude/agents/AGENT_OWNERSHIP.md` §2 for full curation rules.

## Learnings

### 2025-03 — View API Enrichment Pattern
When frontend views need computed data (counts, cross-team flags, risk assessments),
add the computation to the backend view presenter (`adapter/presenter/view.go`)
and return it as part of the view API response. This keeps the frontend thin.

Pattern:
1. Add fields to the Go view model struct
2. Compute in the presenter `build*View()` function
3. Return via the existing `/api/v1/models/{id}/views/{viewType}` endpoint
4. Update the TypeScript type in `api.ts`
5. Simplify the frontend component to just render what the API provides

### 2025-03 — Changeset Commit Flow
Full stack: frontend `commitChangeset()` → backend `handleCommitChangeset` →
applies changeset → validates → replaces model → clears caches → returns result.
Frontend then remounts EditPanel and reloads current view data.

## Known Gotchas

- Backend view responses use camelCase JSON tags (`json:"teamSpan"`)
- Frontend TypeScript types must match exact JSON keys
- The model ID is propagated via ModelContext on the frontend
- View endpoints return full view-specific DTOs, not raw domain entities
