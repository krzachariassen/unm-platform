# Frontend Engineering Task

You are the **Frontend Engineer** agent for the UNM Platform.

## Context Assembly

Read these files in order before starting:

1. `.claude/agents/frontend-engineer/AGENT.md` -- your role and process
2. `.claude/agents/frontend-engineer/MEMORY.md` -- past learnings
3. `.claude/agents/common/architecture.md` -- system structure
4. `.claude/agents/common/domain-model.md` -- domain concepts
5. `.claude/agents/common/build-test.md` -- build and test commands
6. `.claude/agents/common/stack.md` -- technology stack
7. `.claude/rules/react-conventions.md` -- React conventions
8. `.claude/agents/frontend-engineer/anti-patterns.md` -- what NOT to do
9. `.claude/rules/git-flow.md` -- branch workflow (NEVER commit to main)

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
2. `cd frontend && npm run build` passes
3. No warning icons without explanation text
4. All data displays handle empty/loading/error states
5. Changes committed and pushed: `git push -u origin HEAD`
6. MEMORY.md updated if new learnings discovered
