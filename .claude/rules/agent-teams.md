# Agent Team Rules

## When to Parallelize

| Situation | Use | Why |
|-----------|-----|-----|
| 3+ items touching different files | Agent team (3-5 teammates) | True parallel implementation |
| Cross-layer work within one stack (e.g. entity + parser + tests) | Single specialist agent | Shared context prevents drift — but still use the specialist agent, not the orchestrator |
| Quick focused tasks (run tests, lint) | Subagent | Lower cost, fast result |
| Research before implementation | Subagent (Explore) | Read-only, cheap |
| Frontend + Backend simultaneously | Agent team | Independent stacks |

> **Critical distinction**: "Single session" means *don't parallelize* — it does NOT mean implement directly as the orchestrator. Always route through the appropriate specialist agent (`/backend`, `/frontend`, etc.). The orchestrator's job is routing and coordination, not implementation.

## Orchestrator vs. Specialist Agents

The orchestrator MUST NOT implement code directly. Its responsibilities are:
1. Classify the task (backend / frontend / fullstack / docs)
2. Route to the correct specialist agent
3. Coordinate multiple agents for multi-stack work
4. Integrate and validate after agents finish

If the orchestrator finds itself editing source files, writing tests, or running builds — stop and delegate to the right specialist agent instead.

## File Ownership

- Two teammates MUST NOT edit the same file
- Break work so each teammate owns distinct files
- If two items need the same file, run sequentially or assign one owner

## Teammate Prompts

Include in every spawn prompt:
1. Which backlog items to implement
2. Which files/packages they own
3. Interfaces or types they depend on (paste or reference)
4. The TDD protocol: write failing tests first

## Integration

After parallel teammates finish, the lead MUST:
1. Run `go test ./...` / `npx tsc --noEmit && npx vite build`
2. Run `go vet ./...`
3. Fix integration issues (import cycles, interface mismatches)
4. Verify phase deliverable works end-to-end
