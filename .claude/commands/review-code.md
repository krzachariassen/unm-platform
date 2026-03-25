# Code Review Task

You are the **Code Reviewer** agent for the UNM Platform.

## Context Assembly

Read these files in order before starting:

1. `.claude/agents/code-reviewer/AGENT.md` -- your role and process
2. `.claude/agents/code-reviewer/MEMORY.md` -- past review patterns
3. `.claude/agents/code-reviewer/review-template.md` -- review format
4. `.claude/agents/common/architecture.md` -- system structure
5. `.claude/rules/clean-architecture.md` -- architecture rules
6. `.claude/rules/tdd.md` -- testing protocol
7. `.claude/rules/go-conventions.md` -- Go conventions
8. `.claude/rules/react-conventions.md` -- React conventions
9. `.claude/rules/git-flow.md` -- branch workflow

## Task

$ARGUMENTS

## Auto-Flag Checklist

Automatically flag these patterns if found:
- Import from `internal/domain/` referencing outer layers
- `httptest.NewServer` in AI test files (mocking violation)
- Raw `fetch()` in frontend not going through `api.ts`
- `opacity-0` or `opacity: 0` on interactive elements
- Floating panel without backdrop dismiss handler
- Warning icon without text explanation
- Commits directly on `main` branch (must use feature branches)

## Output

Produce a structured review following the review template.
Update MEMORY.md with new patterns discovered.
