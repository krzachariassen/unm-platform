# Backend Engineering Task

You are the **Backend Engineer** agent for the UNM Platform.

## Context Assembly

Read these files in order before starting:

1. `.claude/agents/backend-engineer/AGENT.md` -- your role and process
2. `.claude/agents/backend-engineer/MEMORY.md` -- past learnings (check for relevant history)
3. `.claude/agents/common/architecture.md` -- system structure
4. `.claude/agents/common/domain-model.md` -- domain concepts
5. `.claude/agents/common/build-test.md` -- build and test commands
6. `.claude/agents/common/stack.md` -- technology stack
7. `.claude/rules/clean-architecture.md` -- architecture rules
8. `.claude/rules/tdd.md` -- testing protocol
9. `.claude/rules/go-conventions.md` -- Go conventions
10. `.claude/agents/backend-engineer/anti-patterns.md` -- what NOT to do

## Task

$ARGUMENTS

## Completion Criteria

1. All tests pass: `cd backend && go test ./...`
2. No vet warnings: `cd backend && go vet ./...`
3. Code follows Clean Architecture (no layer violations)
4. Tests written first (TDD)
5. MEMORY.md updated if new learnings discovered
