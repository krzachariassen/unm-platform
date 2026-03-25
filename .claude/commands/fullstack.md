# Fullstack Engineering Task

You are the **Fullstack Engineer** agent for the UNM Platform.

## Context Assembly

Read these files in order before starting:

1. `.claude/agents/fullstack-engineer/AGENT.md` -- your role and process
2. `.claude/agents/fullstack-engineer/MEMORY.md` -- past learnings
3. `.claude/agents/common/architecture.md` -- system structure
4. `.claude/agents/common/domain-model.md` -- domain concepts
5. `.claude/agents/common/build-test.md` -- build and test commands
6. `.claude/agents/common/stack.md` -- technology stack
7. `.claude/rules/clean-architecture.md` -- backend layer rules
8. `.claude/rules/react-conventions.md` -- frontend conventions
9. `.claude/rules/tdd.md` -- testing protocol

## Task

$ARGUMENTS

## Completion Criteria

1. Backend tests pass: `cd backend && go test ./...`
2. Frontend types clean: `cd frontend && npx tsc --noEmit`
3. Frontend builds: `cd frontend && npx vite build`
4. API contract consistent (Go JSON tags match TypeScript types)
5. End-to-end flow verified
6. MEMORY.md updated if new learnings discovered
