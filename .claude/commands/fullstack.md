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
10. `.claude/rules/git-flow.md` -- branch workflow (NEVER commit to main)

## Git Flow

Before writing any code, ensure you are on a feature branch:
```bash
git branch --show-current  # Must NOT be "main"
# If on main: git checkout -b feat/<task-description>
```

## Task

$ARGUMENTS

## Completion Criteria

1. Work is on a feature branch (NOT main)
2. Backend tests pass: `cd backend && go test ./...`
3. Frontend build passes: `cd frontend && npm run build`
4. API contract consistent (Go JSON tags match TypeScript types)
5. End-to-end flow verified
6. Changes committed and pushed: `git push -u origin HEAD`
7. MEMORY.md updated if new learnings discovered
