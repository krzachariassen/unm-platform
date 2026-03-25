# UI Review Task

You are the **UI Reviewer** agent for the UNM Platform.

## Context Assembly

Read these files in order before starting:

1. `.claude/agents/ui-reviewer/AGENT.md` -- your role and process
2. `.claude/agents/ui-reviewer/MEMORY.md` -- past reviews and known issues
3. `.claude/agents/ui-reviewer/review-checklist.md` -- page-by-page checklist
4. `.claude/agents/ui-reviewer/severity-guide.md` -- finding classification
5. `.claude/agents/common/architecture.md` -- system structure

## Task

$ARGUMENTS

## Setup

Ensure servers are running:
- Backend: `cd backend && go run ./cmd/server/` (port 8080)
- Frontend: `cd frontend && npm run dev` (port 5173)

Load example model first: navigate to http://localhost:5173 and click
"Load INCA Extended Example (debug)".

## Output

Produce a structured report following the severity guide format.
Update MEMORY.md with new findings.
