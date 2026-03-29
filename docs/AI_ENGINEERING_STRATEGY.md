# INCA Platform: AI Engineering Strategy

**Status**: Updated Draft — Incorporates UNM Platform Validation Findings
**Owner**: Kristian Zachariassen
**Target**: Inventory & Catalog Platform
**Original Date**: March 2026
**Last Updated**: March 2026

---

## Executive Summary

The problem: INCA engineers spend significant time on KTLO work, bug fixes,
dependency updates, flaky tests, and operational issues. Our current AI tooling
is a single shared CLAUDE.md file (381 lines) that gives every AI interaction
the same context regardless of task type. This does not scale, cannot enforce
quality gates, and has no feedback mechanism to improve over time.

The proposal: Build a specialized agent workforce — purpose-built AI agents for
specific task types (starting with KTLO), each briefed with focused context,
validated through automated gates, and improved through a structured feedback
loop. Agents are stateless; knowledge is version-controlled in the repo.

**Validation update**: The agent architecture described in this strategy has
been validated on a separate project — the UNM Platform — confirming that the
core patterns (specialized agents, auto-routing orchestration, memory-based
learning, parallel agent execution) work in practice. This updated strategy
incorporates those findings and the architectural improvements discovered
during validation. Key innovations include: intelligent auto-routing
orchestration, parallel agent execution with git worktrees, single backlog file
management, IDE-agnostic rule syncing, and the principle that skills are tools
agents wield — not scripts that replace agent judgment.

The ask: Approve Phase 1 investment (4 weeks, 1 engineer full-time + 1 senior
engineer part-time) to build and validate the first agent: the KTLO Engineer.

Phase 1 success criteria:
- KTLO Engineer agent operational with common context modules and orchestrator
- Validation pipeline (build, test, coverage, lint, security) integrated
- Tested on at least 5 real KTLO tickets with >60% first-pass success rate

Expected 6-month outcome: 50% reduction in KTLO time-to-resolution, >50% of
eligible KTLO tickets handled with AI assistance, zero increase in production
incidents from AI-assisted changes.

---

## 1. Why Are We Doing This?

### 1.1 The Platform

INCA (Inventory & Catalog Platform) is a large-scale distributed platform
consisting of approximately 40 microservices and worker services. It handles
catalog management, feed processing, indexing, publishing, and serving for
Uber's delivery business. The codebase follows MVCS architecture with strict
layer conventions, uses Cadence for workflow orchestration, Docstore/MySQL for
storage, and YARPC/Protobuf for communication.

### 1.2 Where We Are with AI Today

Our current AI integration is a first-generation, monolithic context approach:

| Asset | Description |
|-------|-------------|
| 1 shared CLAUDE.md (381 lines) | Single file covering architecture, build commands, conventions |
| 6 rule files in claude/rules/ | Testing, error handling, observability, Cadence, code quality, Go style |
| 6 shared slash commands | diff-coverage, PR create/update, coverage-check, buildprune, ctxprop |
| 1 MCP server (inca-mcp) | 3 tools: readentity, getcatalogconfig, simulatepublishing |

This is the "employee handbook" model — every AI interaction receives the same
381-line preamble regardless of purpose.

### 1.3 The Challenges

1. **Context Dilution**: Every interaction gets the full CLAUDE.md regardless
   of task type. The more we add, the less the model focuses on any part.
2. **No Task Specialization**: A bug fix gets the same context as a new feature
   or a code review. No adaptation of behavior or validation.
3. **No Workflow Guardrails**: Nothing prevents skipping observability, Flipr
   gating, workflow versioning, or coverage checks.
4. **No Feedback Loop**: No measurement of success, override rates, or
   systematic learning from mistakes.
5. **High KTLO Burden**: Maintenance work consumes significant engineering
   time on well-defined, repetitive tasks.
6. **No Path to Scale**: The single CLAUDE.md approach cannot accommodate
   40+ services without becoming unmanageably large.

---

## 2. What Are We Doing?

### 2.1 Vision

Treat AI as a specialized workforce — not a single general-purpose assistant,
but a team of purpose-built agents, each briefed with exactly the context they
need, and validated through automated quality gates before any human reviews
their work.

### 2.2 Design Principles

**Agents are disposable; knowledge is durable.** Every agent invocation starts
fresh. Knowledge lives in version-controlled briefing materials plus curated
operational memory. Auditable, reviewable, mergeable.

**Validation over trust.** Do not try to make the AI "smart enough" to never
make mistakes. Make the validation pipeline fast enough that catching mistakes
is cheap.

**Specialize narrowly, compose broadly.** A KTLO agent that deeply understands
Docstore error patterns outperforms a general agent that knows a little about
everything. Compose agents in sequence or parallel.

**Human-in-the-loop is a feature, not a limitation.** The goal is to shift
engineers from "write the code" to "review and approve pre-validated packages."

**Intelligence orchestrates, determinism enforces.** The orchestrator uses LLM
reasoning to classify tasks, route to agents, and adapt to novel situations.
Validation gates use deterministic scripts to enforce quality bars. Skills and
tools are invoked by agent judgment, not hardcoded into pipelines. This
preserves adaptability while guaranteeing safety.

_This principle emerged from UNM validation. A fully deterministic orchestrator
(Python script) handles only cases its author anticipated. A fully intelligent
orchestrator (LLM) might skip steps. The solution: intelligence at the
decision layer, determinism at the enforcement layer. The agent reasons about
what to do; the environment guarantees that validation runs._

**Backend computes, frontend displays.** All derived data is computed in the
backend. The frontend only visualizes what the API provides. No re-computation
of signals, thresholds, or ratios on the client side.

_Learned from UNM: a threshold mismatch bug was caused by splitting
computation between backend and frontend. The fix — moving all computation to
the backend — is now an architectural invariant._

**Anti-patterns over abstract rules.** Showing what NOT to do with concrete
code examples is more effective than abstract guidance. Each agent maintains an
`anti-patterns.md` with real examples of mistakes to avoid.

**Measure the pipeline, not the model.** The question is not "how good is the
LLM." It is "what percentage of tickets go from creation to merged diff with
less than 30 minutes of engineer time?"

**Start with the boring stuff.** KTLO is the starting point: well-defined,
low-risk, high-volume.

---

## 3. Architecture: The Agent Workforce

### 3.1 System Overview

```
┌─────────────────────────────────────────────┐
│             Engineer Intent                  │
│  (Ticket, bug report, feature request)       │
└──────────────────┬──────────────────────────┘
                   │
                   v
┌─────────────────────────────────────────────┐
│        Layer 1: ORCHESTRATOR                 │
│                                              │
│  - Classifies task intent from description   │
│  - Selects agent(s) and execution strategy   │
│  - Assembles briefing packet per agent       │
│  - Injects service-specific context          │
│  - Determines: single, sequential, parallel  │
└──────────────────┬──────────────────────────┘
                   │
                   v
┌─────────────────────────────────────────────┐
│        Layer 2: SPECIALIZED AGENT(S)         │
│                                              │
│  Clean-slate LLM invocation with:            │
│  - Agent role definition (AGENT.md)          │
│  - Operational memory (MEMORY.md)            │
│  - Anti-patterns and examples                │
│  - Service-specific context (injected)       │
│  - Domain knowledge via MCP server           │
│  - Available skills as tools                 │
│  - Constraints and checklists                │
└──────────────────┬──────────────────────────┘
                   │
                   v
┌─────────────────────────────────────────────┐
│       Layer 3: VALIDATION PIPELINE           │
│                                              │
│  Automated Gates (hard blocks):              │
│  - Build passes (bazel/go build)             │
│  - Tests pass (bazel/go test)                │
│  - Diff coverage >= 90%                      │
│  - Lint compliance                           │
│  - Security scan (no secrets, no PII)        │
│                                              │
│  AI Review Gate (soft blocks, Phase 2+):     │
│  - Code Reviewer agent reviews output        │
│  - Checks observability, gating, patterns    │
│                                              │
│  Self-correction:                            │
│  - On failure: one structured retry with     │
│    failure details before escalating         │
└──────────────────┬──────────────────────────┘
                   │
                   v
┌─────────────────────────────────────────────┐
│          Layer 4: HUMAN REVIEW               │
│                                              │
│  Engineer receives pre-validated package.    │
│  Role: verify and approve, not build.        │
└──────────────────┬──────────────────────────┘
                   │
                   v
┌─────────────────────────────────────────────┐
│      Layer 5: FEEDBACK & LEARNING            │
│                                              │
│  - Agent appends learnings to MEMORY.md      │
│  - Human overrides captured and categorized  │
│  - Repeated patterns promoted to anti-patterns│
│  - Monthly curation of memory files          │
│  - Backlog updated via completion hook       │
└─────────────────────────────────────────────┘
```

### 3.2 Layer 1: The Orchestrator

The orchestrator is the missing layer. Its job: assemble the right context for
the right agent so the LLM receives a focused briefing instead of a 381-line
preamble.

#### What It Actually Is

The orchestrator is a Claude Code custom command — a markdown file in
`.claude/commands/` that becomes a /slash command. No new infrastructure,
no separate service, no build step.

#### Auto-Routing vs. Per-Agent Commands

The original Phase 1 design uses per-agent-type commands: the engineer types
`/ktlo` to invoke the KTLO Engineer. This is sufficient for Phase 1 validation.

**UNM validation finding**: An auto-routing orchestrator (`/run <task>`)
that classifies intent from the task description and routes to the correct
agent(s) is more powerful and more aligned with the "AI Engineer" vision:

| Approach | Pros | Cons |
|----------|------|------|
| Per-agent commands (`/ktlo`, `/review`) | Simple, explicit, no classification errors | Human must know which agent to invoke |
| Auto-routing orchestrator (`/run`) | Human describes the task; AI determines the approach | Classification can be wrong; mitigated by memory and explicit routing rules |

Recommendation: Start with per-agent commands in Phase 1 for predictability.
Introduce auto-routing in Phase 2 when the agent catalog is larger and routing
rules are validated. The auto-routing orchestrator should include a decision
tree that the LLM follows, not free-form classification.

#### Execution Strategies

_Discovered during UNM validation — the original strategy only addresses
sequential agent chaining. Real-world tasks require four distinct strategies._

**Strategy A: Single Agent, Single Branch**
One agent, one task, one branch. The simplest case.

**Strategy B: Multi-Agent, Single Branch**
Multiple agents contribute to one feature. They share a branch because their
work is one logical change. Agents execute sequentially (dependent work) or as
teammates with file ownership (independent work on the same branch).

**Strategy C: Parallel Tasks via Git Worktrees**
Multiple independent tasks each get their own worktree, branch, and PR. True
parallel execution — each agent works in an isolated working directory.

**Strategy D: Hybrid (Sequential Core + Parallel Periphery)**
Do shared/dependent work first on the main branch, then fan out to worktrees
for independent follow-up work.

```
Is it ONE feature or MULTIPLE independent tasks?
├── ONE feature
│   ├── Touches only one layer? → Strategy A (single agent)
│   └── Touches multiple layers?
│       ├── Layers are dependent? → Strategy B (sequential)
│       └── Layers are independent? → Strategy B with agent team
└── MULTIPLE independent tasks
    └── Strategy C (worktrees, each task gets own branch + PR)
```

#### Agent Teams: Parallel Execution Rules

_Innovation from UNM validation. Not in the original strategy._

When multiple agents work in parallel:

- **File ownership is mandatory**: Two agents MUST NOT edit the same file.
  Break work so each agent owns distinct files.
- **Worktree isolation**: Each parallel agent gets its own git worktree
  (own working directory, own branch, own commit history).
- **Integration validation**: After parallel agents finish, the orchestrator
  runs cross-cutting validation (full build + test suite).
- **Teammate spawn prompts**: Include file ownership, interfaces/types
  depended on, TDD protocol, and validation commands.

#### Context Assembly

For each agent being invoked, the orchestrator assembles a briefing packet:

**Always read:**
- The agent's `AGENT.md` — role, process, constraints
- The agent's `MEMORY.md` — past operational learnings
- `common/architecture.md` — system structure
- `common/domain-model.md` — domain concepts (if applicable)

**For code-writing agents, also read:**
- `common/build-test.md` — build and test commands
- `common/safety-checklist.md` — pre-submission checks
- `common/security.md` — secrets, PII, access control
- Relevant rules from `.claude/rules/`
- The agent's `anti-patterns.md` (if it exists)
- Service-specific context file (if it exists)

### 3.3 Layer 2: Specialized Agent Profiles

#### Why KTLO First

| Factor | Rationale |
|--------|-----------|
| Highest burden | Largest consumer of engineering time |
| Most repetitive | Bug fixes, dependency updates follow well-defined patterns |
| Lowest risk | Small, scoped, conservative changes |
| Easiest to measure | Clear start/end timestamps, well-defined "done" criteria |
| Proves the framework | One agent end-to-end validates the entire architecture |

#### The KTLO Engineer Agent

The KTLO Engineer is specialized by **operational posture**, not technology.
It uses the same technical capabilities as a backend or frontend engineer but
wraps them in a conservative, maintenance-focused risk profile.

_UNM validation insight: The UNM project built agents specialized by
technology layer (backend-engineer, frontend-engineer, fullstack-engineer).
This is the right decomposition for feature development. The KTLO Engineer
adds a different axis — risk tolerance. Same technical skills, different
judgment framework: always conservative, always scoped tightly, always
regression-tested._

Key constraints that distinguish KTLO from feature development:
- NEVER change public API contracts without explicit approval
- ALWAYS Flipr-gate behavioral changes
- ALWAYS maintain or improve test coverage (minimum 90% diff coverage)
- NEVER modify Cadence workflow execution paths without `workflow.GetVersion()`
- NEVER introduce new external service dependencies without approval
- When choosing between a conservative fix and an elegant refactor, choose
  the conservative fix
- Scope changes tightly to the issue at hand

#### Agent Catalog

| Agent | Persona | What It Does | Phase |
|-------|---------|-------------|-------|
| KTLO Engineer | L3/L4 maintenance engineer | Fixes bugs, tech debt, dependency updates, operational issues. Conservative posture. | Phase 1 |
| Test Writer | QA-focused engineer | Writes tests, improves coverage, generates mocks. | Phase 2 |
| Code Reviewer | Senior engineer | Reviews diffs against standards, catches missing observability/gating. | Phase 2 |
| Bug Investigator | SRE / incident debugger | Investigates production issues using MCP tools. | Phase 3 |
| Feature Builder | L5 engineer | Builds features following patterns with observability and rollout planning. | Phase 3 |
| Migration Specialist | Infrastructure engineer | Handles schema changes, data migrations, backward-compatible rollouts. | Phase 3 |

**UNM-validated additional agents** (proven useful in the UNM Platform project):

| Agent | What It Does | Applicability to INCA |
|-------|-------------|----------------------|
| Backend Engineer | Go specialist with 7-phase TDD workflow. Technology-layer specialist. | Direct — INCA's backend is also Go |
| Frontend Engineer | React/TypeScript specialist with 4-phase workflow. | Applicable if INCA builds frontend tooling |
| Fullstack Engineer | Vertical-slice owner designing API contracts before implementation. | Applicable for new features |
| UI Reviewer | Systematic UX testing with cross-view consistency checks. | Applicable for any user-facing tooling |
| Documentation Writer | Verifies claims against code before writing. Approved example domains. | Direct — documentation discipline applies |
| Backlog Manager | Single `docs/BACKLOG.md` maintenance with completion hooks. | Direct — solves a universal AI workflow gap |

### 3.4 Agent Profile Structure

```
claude/agents/
├── common/                          # Shared context modules
│   ├── architecture.md              # System structure (~30 lines)
│   ├── domain-model.md              # Domain concepts and relationships
│   ├── build-test.md                # Build/test commands (~20 lines)
│   ├── safety-checklist.md          # Universal pre-submission checks
│   ├── security.md                  # Secrets, PII, access control
│   └── stack.md                     # Technology stack reference
│
├── ktlo-engineer/                   # Phase 1
│   ├── AGENT.md                     # Role, scope, constraints, process
│   ├── MEMORY.md                    # Operational learnings (agent-maintained)
│   ├── anti-patterns.md             # Known mistakes to avoid (team-curated)
│   ├── checklist.md                 # Validation steps
│   └── examples/                    # Gold-standard past fixes
│
├── [additional agents per roadmap]
│
└── backlog-manager/                 # Backlog management (UNM-validated)
    └── AGENT.md                     # Single backlog rules and workflows
```

#### Service-Specific Context (Injected, Not Baked)

Service-specific context lives alongside the services and is injected by the
orchestrator based on the working directory:

```
core/claude-context.md
feed-worker/claude-context.md
publisher/claude-context.md
indexer/claude-context.md
```

The agent's role and constraints stay consistent while its domain knowledge
adapts to wherever it is working.

### 3.5 Layer 3: Validation Pipeline

Every agent output passes through automated validation before reaching a
human. This is the environmental enforcement layer — it runs deterministically
regardless of whether the agent "remembered" to validate.

```
Agent Output
    │
    ├──► Hard Gates (automated, blocking)
    │    ├── Build passes
    │    ├── Tests pass
    │    ├── Diff coverage >= 90%
    │    ├── Lint compliance
    │    ├── No hardcoded secrets
    │    ├── No PII in log statements
    │    └── Flipr gate present for behavioral changes
    │
    ├──► Self-Correction (one retry)
    │    ├── If hard gate fails: agent receives structured failure
    │    │   details and gets one attempt to fix
    │    └── If retry fails: escalate to human with failure report
    │
    ├──► AI Review Gate (Phase 2+)
    │    ├── Code Reviewer agent reviews the output
    │    ├── Checks observability, error handling, gating
    │    └── Generates structured review for human
    │
    └──► Human Review (always required)
         ├── Engineer receives pre-validated package
         ├── AI-generated review summary highlights decisions
         └── Approve, request changes, or reject
```

_UNM validation finding: Prompt-based validation is unreliable — the agent
sometimes skips checks. The validation pipeline must be a tool the agent
invokes (a script or command), not instructions the agent might follow.
Environmental enforcement (like pre-commit hooks) is even stronger because
it runs regardless of agent behavior._

### 3.6 Skills as Agent Tools

_New section — emerged from analysis of the Claude Skills ecosystem vs.
the agent framework architecture._

Skills (installable packages like `/mega-review`, `/code-review-expert`) are
powerful execution tools. They should be used **by agents**, not **instead of
agents**.

The distinction:
- **Skills** are deterministic, reliable, parallelized execution tools. A
  skill like `/mega-review` launches 10-11 review agents in parallel with
  consensus weighting. It does one thing extremely well.
- **Agents** are intelligent, context-aware, adaptive reasoners. An agent
  reads its MEMORY.md, understands the project's architectural constraints,
  and decides when a deep review is warranted vs. a quick check.

The correct architecture: the agent decides which skills to invoke based on
judgment, context, and memory. A Code Reviewer agent might invoke
`/mega-review` for high-risk auth changes and `/dual-review` for a simple
refactor. The skill is a tool; the agent is the engineer who wields it.

A deterministic script orchestrator (Python routing to hardcoded agents)
handles only cases its author anticipated. An intelligent orchestrator
handles novel situations. The "AI Engineer" must be intelligent at the
decision layer and deterministic at the enforcement layer — not the reverse.

### 3.7 Agent Memory: Operational Learning

Each agent maintains a `MEMORY.md` file containing reusable operational
learnings. Memory is different from curated `anti-patterns.md` — think of
it as the difference between a team wiki (authoritative) and an engineer's
field notes (practical tips).

#### What Goes in MEMORY.md

Only reusable platform knowledge that will help on a future, different task.
The test: "would this help me if I were working on a completely different
ticket in this service area six weeks from now?"

**Good entries** (reusable platform knowledge):
```markdown
### 2026-03-15 | feed-worker | Cadence workflow timing
Several feed-worker workflows use workflow.Sleep() with long durations.
Always check for sleeps that affect testing. Use workflow.Now(ctx) for
time comparisons, never time.Now().
```

**Bad entries** (task-specific notes):
```markdown
### 2026-03-15 | core | Current task progress
Fixed the nil pointer in entity.go line 42, still need to update tests.
```

#### Memory Guardrails

- **Reusable knowledge only**: Must be about platform behavior, not current
  task progress
- **Structured entries**: Date, service context, clear learning
- **Size-bounded**: Hard cap at 30 entries per agent
- **Proven-useful promotion**: During monthly curation, entries referenced
  in subsequent tasks are promoted to `anti-patterns.md`
- **Monthly curation**: Engineer reviews with three outcomes per entry:
  promote, keep, or prune
- **Transparent and auditable**: Markdown file in the repo, reviewable in diffs

### 3.8 Feedback Loop & Ownership

| Responsibility | Owner | Cadence |
|----------------|-------|---------|
| Agent definitions (AGENT.md) | Agent author | Updated when process changes; reviewed quarterly |
| Anti-patterns (anti-patterns.md) | Any engineer (PR-based) | Updated when a pattern repeats 3+ times |
| Examples (examples/) | Senior engineer | Added when a gold-standard fix is identified |
| Operational memory (MEMORY.md) | Agent (automated) + reviewing engineer | Agent appends per-session; engineer curates monthly |
| Service context (claude-context.md) | Service owner | Updated when service patterns change; quarterly review |
| Metrics review | Team lead | Weekly (15 min review) |

**Staleness Prevention:**
- Each briefing file includes a `Last reviewed: YYYY-MM-DD` header
- Monthly: engineer reviews MEMORY.md, promotes valuable entries, prunes noise
- Quarterly: team reviews all agent definitions for accuracy
- If a service context file has not been updated in 6 months, flag it

### 3.9 Agent Chaining (Phase 2+)

Sequential agent composition for complex tasks:

```
Ticket: "Fix NPE in entity lookup when external_id is empty"

  Step 1: KTLO Engineer
          Input:  Ticket + affected code + service context + memory
          Output: Modified source files

  Step 2: Test Writer
          Input:  KTLO Engineer's changes + existing tests + coverage
          Output: New/updated test files

  Step 3: Validation Pipeline
          Input:  Combined changeset
          Output: Pass/fail for all hard gates

  Step 4: Code Reviewer
          Input:  Combined diff + gate results + memory
          Output: Structured review document

  Step 5: Human Engineer
          Input:  Diff + tests + review + gate results
          Action: Review and approve
```

### 3.10 Parallel Agent Execution (UNM-Validated)

_New section — not in the original strategy. Validated on UNM Platform._

For multiple independent tasks, agents execute in parallel using git worktrees:

```bash
# Task 1: Fix flaky test
git worktree add .worktrees/fix-flaky-test -b fix/flaky-test main

# Task 2: Update deprecated dependency
git worktree add .worktrees/chore-dep-update -b chore/dep-update main

# Each agent works in its own worktree with its own branch
# Each agent: commit, push, create PR from their worktree
# After all done: clean up worktrees
```

**Rules:**
- One task = one worktree = one branch = one PR
- Two agents MUST NOT edit the same file
- The lead/orchestrator stays in the main worktree
- Integration validation runs after all agents complete
- Worktrees cleaned up after tasks complete

### 3.11 Backlog Management (UNM-Validated)

_New section — not in the original strategy. Validated on UNM Platform._

AI agents are unreliable at maintaining updated backlogs unless given explicit
structure. The solution: a single backlog file (`docs/BACKLOG.md`) with a
dedicated agent.

**`docs/BACKLOG.md`**:
- Human-owned structure, phases, and priorities
- Phased checklists with scope tags and a "Recently Completed" section
- Updated after every task via orchestrator completion hook (checkboxes, dates, pruning)

**Backlog Manager Agent**:
- Invoked after every task completion (via orchestrator hook)
- Marks completed items, maintains Recently Completed
- Suggests next unchecked items when near-term work is thin
- Never invents items, never reorders, never rewrites roadmap narrative without explicit instruction

### 3.12 IDE-Agnostic Rule Syncing (UNM-Validated)

_New section — not in the original strategy._

The agent framework must work across development tools. The `.claude/` directory
is native to Claude Code. For other IDEs (Cursor, VS Code with Copilot, etc.),
rules must be synced to the IDE's native format.

Pattern validated on UNM Platform:

| IDE | Rule Location | Sync With |
|-----|--------------|-----------|
| Claude Code | `.claude/rules/`, `.claude/agents/`, `.claude/commands/` | Source of truth |
| Cursor | `.cursor/rules/*.mdc` | Mirror of `.claude/` rules with IDE-specific frontmatter |

Each Cursor rule file includes:
- `alwaysApply: true` for universal rules (git-flow, routing)
- `globs: ["backend/**"]` for layer-specific rules
- Content extracted from the corresponding `.claude/` files

When updating agent rules in `.claude/`, keep `.cursor/rules/` in sync.

---

## 4. MCP Server Extensions

The inca-mcp server is the agent's read-only interface to the platform.

### Existing Tools

| Tool | Purpose |
|------|---------|
| readentity | Read entity from INCA catalog |
| getcatalogconfig | Retrieve catalog configuration |
| simulatepublishing | Simulate publishing without persisting |

### Proposed New Tools

| Tool | Purpose | Phase |
|------|---------|-------|
| getcoverage | Current test coverage for a package | Phase 1 |
| getservicehealth | Recent error rates, latency, SLA status | Phase 2 |
| getrecentchanges | Git log for a service area | Phase 2 |
| getfliprstate | Active Flipr gates for a service | Phase 2 |
| getservicedependencies | Upstream/downstream dependency graph | Phase 3 |
| getdocstoreschema | Schema definition for a Docstore table | Phase 3 |
| getrecentincidents | Recent operational incidents | Phase 3 |

---

## 5. Implementation Roadmap

### Phase 1: KTLO Engineer (Weeks 1-4)

Goal: Build the KTLO Engineer agent end-to-end and validate on real tickets.

**5.1** Create common context modules (`common/architecture.md`,
`build-test.md`, `safety-checklist.md`, `security.md`)

**5.2** Build the KTLO Engineer agent (`AGENT.md`, `MEMORY.md`,
`anti-patterns.md`, `checklist.md`, `examples/`)

**5.3** Build orchestrator command (`/ktlo` slash command with context
assembly and service-specific context injection)

**5.4** Build validation pipeline (`/validate` command running all hard gates
with structured pass/fail output)

**5.5** Create initial service context for 1-2 highest-KTLO-volume services

**Phase 1 Exit Criteria:**
- KTLO Engineer operational with common context and orchestrator
- Validation pipeline catches issues that would fail in CI
- Tested on at least 5 real KTLO tickets
- First-pass success rate >60%
- MEMORY.md has accumulated at least 5 operational learnings
- Adding a second agent requires only writing briefing files

### Phase 2: Test Writer, Code Reviewer & Chaining (Weeks 5-10)

**5.6** Build Test Writer Agent
**5.7** Build Code Reviewer Agent
**5.8** Implement agent chaining (KTLO → Test Writer → Validation → Reviewer)
**5.9** Create service-specific context for top 5 services
**5.10** Begin measurement
**5.11** Introduce auto-routing orchestrator (`/run`)
**5.12** Implement backlog management agent and single `docs/BACKLOG.md` workflow

**Phase 2 Exit Criteria:**
- Three agents operational and chained
- Service-specific context for top 5 services
- At least 2 weeks of measurement data
- First-pass success rate >50% for the full chain

### Phase 3: Expansion & Feedback (Weeks 11-16)

**5.13** Add Bug Investigator, Feature Builder, Migration Specialist
**5.14** Establish feedback loop with explicit ownership
**5.15** Build measurement dashboard
**5.16** Implement parallel agent execution with worktrees
**5.17** Add IDE-agnostic rule syncing (Cursor rules)

**Phase 3 Exit Criteria:**
- 6+ agents operational
- Feedback loop producing anti-pattern updates
- Measurement dashboard live with 4+ weeks of data
- Measurable reduction in KTLO time-to-resolution

### Phase 4: Maturity & Proactive Agents (Months 5-6+)

**5.18** Proactive agents (Flipr Gate Cleaner, Dependency Updater, Coverage
Gap Finder)
**5.19** Extend to all INCA services
**5.20** Cross-team sharing and packaging

---

## 6. Expected Outcomes & Measurement

### 6.1 Key Metrics

| Metric | Definition | Target (6 months) |
|--------|-----------|-------------------|
| First-pass success rate | Agent outputs passing all gates without retry | > 60% |
| Human override rate | Agent outputs modified or rejected at review | < 30% |
| KTLO time-to-resolution | Median time from ticket to merged diff | 50% reduction |
| KTLO AI-assistance rate | Eligible tickets where an agent was invoked | > 50% |
| Coverage delta | Coverage change on files touched by agents | Net positive |
| Production incident rate | Incidents attributed to AI-assisted changes | 0 increase |
| Memory quality | Entries promoted to anti-patterns at monthly review | > 20% |

### 6.2 What Changes

| Dimension | Today | After Strategy |
|-----------|-------|---------------|
| Context delivery | 1 shared CLAUDE.md | Agent-specific briefing packets |
| Task specialization | None | Dedicated agents with constraints |
| Quality enforcement | Rules as suggestions | Automated validation gates |
| Review burden | Review from scratch | Verify pre-validated packages |
| Measurement | None | Per-agent success rates |
| Feedback | None | Memory + anti-patterns + curation |
| Execution model | Sequential only | Sequential + parallel via worktrees |
| Backlog management | Manual | Single file with AI agent |
| IDE support | Claude Code only | Claude Code + Cursor (synced rules) |

---

## 7. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| AI introduces subtle bugs not caught by tests | 90%+ diff coverage gate, Flipr-gate behavioral changes, Code Reviewer as second check |
| Engineers rubber-stamp AI reviews | Track override rate; investigate if below 15% |
| Briefing materials become stale | Monthly memory curation, quarterly definition review, staleness tracking |
| Over-investment in tooling vs. engineering | Phase 1 is 4 weeks, 1 agent; measure ROI before Phase 2 |
| AI-generated code leaks secrets/PII | Hard gate in validation; security rules in every agent |
| AI hallucination | Validation pipeline catches build/test failures; Code Reviewer checks for hallucination |
| MEMORY.md accumulates noise | 30-entry cap, reusable-knowledge-only rule, monthly curation |
| Prompt-based validation skipped by agent | Environmental enforcement — validation is a tool/script, not an instruction |

---

## 8. Governance & Security

### 8.1 Autonomy Policy

Human review is the final gate for all AI-assisted changes. Changes require:
- Sustained >80% first-pass rate and <20% override rate over 4+ weeks
- Leadership approval at a scheduled decision point
- Kill switch: any engineer can disable agent autonomy by removing
  `claude-context.md`
- Revocation: any P1/P2 incident from AI-assisted change → full human
  review reinstated

### 8.2 Security Considerations

- **MCP access**: Read-only. No credentials, tokens, or writable endpoints
- **Prompt injection**: Dynamic inputs (tickets, errors) treated as untrusted
- **Audit logging**: All invocations logged (metadata only, no code content)
- **Existing gates**: Shift-left security checks apply to all agent output

---

## 9. Resource Requirements

### Phase 1 (Weeks 1-4)
- 1 senior engineer (part-time, ~40%): Define agent profile, seed anti-patterns
- 1 engineer (full-time): Build context modules, orchestrator, validation

### Phase 2 (Weeks 5-10)
- 1 senior engineer (part-time, ~30%): Review chaining quality, service context
- 1 engineer (full-time): Build agents, implement chaining, begin measurement

### Phase 3 (Weeks 11-16)
- 1 engineer (part-time, ~50%): Build remaining agents, dashboard, feedback loop

### Ongoing (Phase 4+)
- ~10% of one engineer's time for quality upkeep and memory curation
- 1 hour/week for metrics review

---

## 10. Decision Points

| When | Decision | Go Criteria | No-Go Action |
|------|----------|------------|-------------|
| After Phase 1 | Proceed to Phase 2? | >60% first-pass on 5+ tickets | Iterate or reassess |
| After Phase 2 | Expand catalog? | Chained pipeline works; >50% first-pass for full chain | Continue iterating on core 3 |
| After Phase 3 | Consider autonomy changes? | Dashboard live; feedback producing updates; measurable KTLO reduction | Continue with full human review |

---

## Appendix A: UNM Platform Validation Summary

The agent architecture was validated on the UNM Platform project (Go backend +
React frontend), a greenfield product for executable architecture modeling.

### What Was Validated

| Component | Implementation | Finding |
|-----------|---------------|---------|
| Auto-routing orchestrator | `/run` command with 4 execution strategies | Works — LLM reliably classifies tasks and routes to correct agents |
| 8 specialized agents | backend, frontend, fullstack, code-reviewer, ui-reviewer, documentation-writer, code-to-dsl, backlog-manager | Agent pattern works for feature development, not only KTLO |
| Agent profile structure | AGENT.md + MEMORY.md + anti-patterns.md | Effective — agents produce higher-quality output with focused context |
| Parallel execution | Git worktrees with file ownership | Works — enables true parallelism for independent tasks |
| Memory system | MEMORY.md per agent with structured entries | Useful but needs guardrails (cap, curation, promotion) |
| Single backlog | `docs/BACKLOG.md` with completion hooks | Solves the "AI forgets to update backlog" problem |
| IDE syncing | .cursor/rules/ mirroring .claude/ | Essential for tool-agnostic agent frameworks |
| Prompt-based validation | Safety checklist as instructions | Unreliable — agents sometimes skip checks. Must be environmental. |

### Key Architectural Insight

Intelligence orchestrates; determinism enforces. The orchestrator reasons about
what to do (which agents, what order, parallel vs. sequential). The validation
pipeline enforces quality deterministically (build, test, lint — always runs).
Skills are tools agents invoke by judgment, not scripts that replace agent
judgment.

---

## Appendix B: Glossary

| Term | Definition |
|------|-----------|
| Agent | Context package (role + rules + constraints + memory) shaping a clean-slate LLM invocation |
| Briefing Packet | Assembled context provided to an agent at invocation |
| Orchestrator | System that classifies tasks, selects agents, assembles briefing packets |
| Hard Gate | Automated validation check that blocks on failure |
| Soft Gate | AI-powered review that provides advisory feedback |
| Anti-pattern | Documented mistake pattern in team-curated briefing materials |
| Operational Memory | Per-agent MEMORY.md containing practical learnings |
| Agent Chaining | Sequential invocation where each output feeds the next |
| Agent Team | Parallel execution of multiple agents with file ownership |
| Worktree | Git worktree providing an isolated working directory for parallel agents |
| Skill | Installable package providing deterministic execution tools that agents invoke by judgment |
| Single backlog | One file (`docs/BACKLOG.md`): human-owned structure; AI updates completion state via backlog-manager |
