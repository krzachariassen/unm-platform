# Agent ownership, curation, and quality

This document defines who owns agent briefing materials, how memory and anti-patterns are curated, how human overrides are tracked, and what metrics we use to judge agent quality.

---

## 1. Agent ownership

| Agent | Owner | Memory curation | Anti-patterns curation | Cadence |
|-------|-------|-----------------|-------------------------|---------|
| backend-engineer | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly skim of `AGENT.md` + shared common docs for drift |
| frontend-engineer | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly skim of `AGENT.md` + shared common docs for drift |
| fullstack-engineer | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly skim of `AGENT.md` + shared common docs for drift |
| code-reviewer | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly skim of review checklists and severity rubric |
| ui-reviewer | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly skim of UX heuristics and scope |
| documentation-writer | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly check against DSL/spec and example models |
| code-to-dsl | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly alignment with parser/domain changes |
| backlog-manager | Agent author | Monthly — agent appends, engineer curates | As needed — PR-based, any engineer | Monthly memory review; quarterly check against `docs/BACKLOG.md` workflow |

**Owner** means accountability for keeping `AGENT.md`, `MEMORY.md`, and optional extras accurate—not exclusive authorship of every line.

---

## 2. Memory curation policy

### Limits and rhythm

- **Hard cap:** 30 entries per agent in `MEMORY.md`.
- **Cadence:** Owning engineer runs a **monthly** pass; agents may **append** anytime between passes.

### Per-entry outcomes

Each entry is classified into exactly one outcome:

| Outcome | Meaning |
|---------|---------|
| **Promote** | Move the substance into `anti-patterns.md` (or another durable doc) and remove or shorten the memory entry to a pointer. |
| **Keep** | Leave in `MEMORY.md`; still relevant and within cap. |
| **Prune** | Delete from `MEMORY.md`; not worth permanent retention. |

### Promotion criteria

Promote when the learning was **referenced or clearly useful** in a later task (same or related agent)—i.e. it prevented a repeat mistake or sped correct work.

### Pruning criteria

Prune when the entry is **older than two months** and was **never reused** (no subsequent task referenced it and it did not change behavior in review).

### Entry format (required)

Every `MEMORY.md` entry must include:

1. **Date** (ISO or explicit calendar date).
2. **Context / service area** (e.g. parser, API handler, specific view).
3. **Clear learning** (one concrete rule or fact, not vague advice).

### Header requirement

Each agent’s `MEMORY.md` **must** include a short header block that points to this file, for example:

```markdown
<!-- Memory policy: see .claude/agents/AGENT_OWNERSHIP.md §2 -->
```

---

## 3. Human override tracking

When a human **modifies or rejects** AI-produced work during PR review, record it in an **append-only** log so patterns feed back into agents.

**Log path:** `.claude/agent-overrides.md`

**Row format (append-only):**

```text
| Date | Agent | Task | Override Category | Description |
```

Use one table in the file; append new rows to the bottom. Keep descriptions short and actionable (what was wrong, what was done instead).

### Override categories

| Category | Use when |
|----------|----------|
| `missed-edge-case` | Boundary condition or scenario was not handled. |
| `wrong-pattern` | Incorrect or outdated approach vs. codebase norms. |
| `violated-constraint` | Architectural rule, layer boundary, or project convention was broken. |
| `hallucinated` | Non-existent imports, APIs, files, or test data were introduced. |
| `scope-creep` | Work went beyond the requested task without agreement. |
| `insufficient-tests` | Tests missing, too shallow, or not aligned with the change. |
| `style` | Minor style or convention mismatch without functional error. |

### Promotion rule

When the **same** `(Agent, Override Category)` pair appears **three or more times** in `.claude/agent-overrides.md`, the owning engineer (or any engineer in a PR) **must** add a corresponding bullet to that agent’s `anti-patterns.md` (or create the file if missing), referencing the log dates if helpful. Reset counting after promotion by treating the anti-pattern as the source of truth going forward.

---

## 4. Metrics

### Definitions

| Metric | Definition | How measured | Target |
|--------|------------|--------------|--------|
| Validation pass rate | Share of agent-led task completions where full `/validate` (backend + frontend checks per project convention) passes on the **first** run after the agent declares completion. | From `.claude/agent-log.md`: count rows where `Validation` = pass on first run, divided by total completions for that agent in the window. | Trend up; team sets numeric target per quarter (e.g. ≥80% once logging is stable). |
| Task completion time | Wall-clock time from documented task start to PR creation (or equivalent “done” event). | Log start/end timestamps in `.claude/agent-log.md` or derive from PR timeline; aggregate median/p90 per agent. | Baseline first; then reduce p90 without sacrificing validation pass rate. |
| Human override rate | Share of agent-touched PRs where a human materially changed agent-written code or rejected a substantive chunk. | Count PRs with ≥1 row in `.claude/agent-overrides.md` for that PR’s agent / task, divided by agent-attributed PRs in the same window. | Trend down; overrides should become **specific** and **rare** as anti-patterns improve. |
| Memory quality | Share of memory entries that were **promoted** to `anti-patterns.md` (or equivalent) during the monthly curation. | During monthly pass: `promoted / (promoted + pruned + kept)` for entries touched that month, per agent. | Non-zero promotion rate when memory is healthy; avoid 100% keep (stale memory). |
| Anti-pattern coverage | Share of agents that have a non-empty, maintained `anti-patterns.md`. | Count agents under `.claude/agents/` with `anti-patterns.md` present vs. total agents (8). | **100%** |

### Completion log

**Path:** `.claude/agent-log.md`

**Purpose:** One row per agent-attributed task completion for validation rate, timing, and scope.

**Row format (append-only):**

```text
| Date | Agent | Task Summary | Validation | Gates | Files |
```

**Column guidance:**

- **Validation:** `pass-first` | `pass-retry` | `fail` (and optional short note).
- **Gates:** e.g. `tests`, `vet`, `tsc`, `build` — what was run; use project-standard names.
- **Files:** Rough scope (e.g. `backend/internal/...` or PR link) for later audits.

---

## Related paths

| Artifact | Location |
|----------|----------|
| Override log | `.claude/agent-overrides.md` |
| Completion / validation log | `.claude/agent-log.md` |
| Agent definitions | `.claude/agents/<agent>/` |
| This policy | `.claude/agents/AGENT_OWNERSHIP.md` |
