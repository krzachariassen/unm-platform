# Agent Ownership & Memory Curation

_Who owns what, and how agent knowledge stays useful over time._

---

## 1. Agent Roster

| Agent | Focus | Key Files |
|-------|-------|-----------|
| backend-engineer | Go backend — domain, handlers, analyzers, parsers | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |
| frontend-engineer | React frontend — pages, components, hooks, styling | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |
| fullstack-engineer | Cross-stack features touching both backend and frontend | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |
| code-reviewer | Code quality audits, architecture compliance | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |
| ui-reviewer | UX review, visual bugs, interaction quality | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |
| documentation-writer | Docs, specs, README, examples, tutorials | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |
| backlog-manager | Backlog maintenance, completion tracking | `AGENT.md`, `MEMORY.md`, `anti-patterns.md` |

Each agent lives in `.claude/agents/<name>/`. The three files serve
distinct purposes:

- **AGENT.md** — Role, workflow, constraints. Stable. Changes rarely.
- **MEMORY.md** — Operational learnings. Changes often. Subject to curation.
- **anti-patterns.md** — Proven mistakes to avoid. Graduates from MEMORY.md.

---

## 2. Memory Curation

### Rules

- **30-entry hard cap** per agent in `MEMORY.md`.
- **Reusable knowledge only** — entries must be about platform behavior,
  not current task progress.
- **Structured format** — every entry needs:
  1. Date (ISO or calendar date)
  2. Context / service area
  3. One concrete learning (not vague advice)

### Curation Outcomes

When an agent's memory approaches the cap, or during periodic review:

| Action | When |
|--------|------|
| **Promote** | The learning prevented a repeat mistake or was referenced in a later task. Move to `anti-patterns.md`, shorten the memory entry to a pointer. |
| **Keep** | Still relevant, not yet validated by reuse. |
| **Prune** | Older than 2 months with no reuse. Delete. |

### Header Requirement

Each `MEMORY.md` must include:

```markdown
> **Policy**: 30-entry cap · Curation: Promote / Keep / Prune
> See `.claude/agents/AGENT_OWNERSHIP.md` §2
```

---

## 3. Anti-Patterns

`anti-patterns.md` is the durable record of proven mistakes. Entries
come from two sources:

1. **Promoted memory entries** — a MEMORY.md learning that was validated
   by reuse in a later task.
2. **Direct observation** — when a human overrides agent work during
   review, the fix should be captured directly in the agent's
   `anti-patterns.md` so it doesn't repeat.

Anti-patterns are permanent. They are not subject to pruning.

---

## 4. Keeping Agents Current

Agent files drift when the codebase changes. Watch for:

- **AGENT.md drift** — workflow references files that were renamed,
  split, or deleted. Skim quarterly or after major refactors.
- **MEMORY.md staleness** — entries reference patterns that no longer
  exist (e.g., removed v1 parser). Prune after migrations.
- **Shared context drift** — files in `.claude/agents/common/` must
  stay aligned with the actual codebase (architecture.md, stack.md,
  domain-model.md, build-test.md).

---

## Related Files

| What | Where |
|------|-------|
| Agent definitions | `.claude/agents/<agent>/` |
| Shared context | `.claude/agents/common/` |
| Backlog (work items) | `docs/BACKLOG.md` |
| Product roadmap | `docs/PRODUCT_ROADMAP.md` |
| This policy | `.claude/agents/AGENT_OWNERSHIP.md` |
