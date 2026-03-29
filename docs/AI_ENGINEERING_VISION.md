# INCA Platform: 1-Year Vision for AI Engineering

**Status**: Updated Draft — Incorporates UNM Platform Validation Findings
**Owner**: Kristian Zachariassen
**Target**: Inventory & Catalog Platform (INCA)
**Original Date**: March 2026
**Last Updated**: March 2026

---

## Executive Summary

INCA is entering a new phase of engineering.

The industry is moving beyond AI as chat, autocomplete, or a one-off
assistant. The leading platforms now describe AI as an agentic teammate:
software that can take a real task, reason across multiple steps, edit files,
run tests, use tools, and return work for human approval. The major developer
tooling companies are converging on the same direction: AI that executes work,
not just discusses it (see Appendix A).

INCA should not treat this as a tooling trend. We should treat it as a new
layer of engineering capacity.

Over the next 12 months, our goal is to create the first AI Engineer for INCA:
not a chatbot, not a single model, and not a vague assistant, but an
orchestrated system of specialized agents that can, within clearly bounded
scope, operate at L4-equivalent output quality on defined categories of
engineering work.

This is not a replacement strategy. It is a leverage strategy.

We are not trying to use AI instead of excellent engineers. We are trying to
ensure that every excellent engineer at INCA can operate with more focus, more
speed, and more impact — especially in a period of significant team growth
where we have real capacity needs and meaningful headcount to fill.

**Validation update**: The agent architecture described in this vision has
been partially validated on a separate project — the UNM Platform. Eight
specialized agents, an auto-routing orchestrator, parallel agent execution,
memory-based learning, and a structured backlog system have been built and
tested. The core patterns work. This updated vision incorporates findings
from that validation, including a new operating principle ("Intelligence
orchestrates, determinism enforces") and evidence that the agent model
applies to both greenfield feature development and operational maintenance.

---

## 1. Why This Matters Now

We have a hiring challenge and a leverage challenge at the same time.

INCA has meaningful platform scope, operational burden, and cross-service
complexity. We also have significant incoming headcount to fill in a market
where strong engineers are expensive, scarce, and slow to ramp. If we solve
this only with hiring, we will improve capacity gradually. If we solve this
with hiring plus AI Engineering, we can improve capacity immediately in the
parts of the workload that are repetitive, bounded, and validation-friendly.

The market direction makes this moment especially important. The major
platforms — GitHub, Google, AWS, Microsoft — are all converging on the same
model: AI that takes real engineering tasks, executes multi-step work,
validates its own output, and returns results for human review. This is not
speculative; it is the current product direction of every major developer
tooling company (see Appendix A).

The conclusion is simple: the next wave is not AI as assistance. It is AI as
delegated execution, with human oversight and strong controls.

INCA has an opportunity to lead rather than follow.

---

## 2. Our North Star

By the end of the next 12 months, every INCA engineer should have access to
an AI Engineer teammate.

That AI Engineer will not be a single model. It will be a coordinated
engineering system made up of specialized agents, shared memory, platform
context, read-only operational awareness, and hard validation gates. It will
be able to take on meaningful portions of KTLO and other bounded engineering
work, produce high-quality diffs, explain its decisions, and hand work back
to humans for final review.

Our north star is not "more prompts." Our north star is not "AI in the IDE."
Our north star is not "faster autocomplete."

Our north star is this:

> INCA will create a new unit of engineering capacity — the AI Engineer —
> that works beside our human engineers, scales on demand for bounded and
> validation-friendly work, improves over time, and gives human engineers
> back the work that matters most.

This is the future we are building:

- **Human engineers** own architecture, product judgment, reliability
  strategy, leadership, and the hardest problem-solving.
- **AI Engineers** absorb the bounded execution layer: KTLO, repetitive
  fixes, test generation, review preparation, dependency updates, low-risk
  cleanup, and other work that benefits from speed, consistency, and
  constant availability.
- **The result** is not fewer engineers. The result is more engineering per
  engineer.

---

## 3. What We Mean by "AI Engineer"

In this vision, "AI Engineer" has two meanings. Both matter.

### A. AI Engineer as a Digital Engineering Capability

This is the system we are building.

It is a collection of specialized agents — KTLO, testing, review,
investigation, and more over time — connected by an orchestrator that routes
work, assembles the right context, invokes the right specialist, and proves
output quality through validation.

The components:

| Component | Role |
|-----------|------|
| **Orchestrator** | Classifies task intent, selects agents, determines execution strategy (single, sequential, parallel), and assembles briefing packets. In early phases, implemented through lightweight, repo-native workflows. Intelligence at the decision layer — the orchestrator reasons about what to do, not just follows a script. |
| **Specialized agents** | Purpose-built context packages that shape LLM behavior for specific task types (KTLO, testing, review) or technology layers (backend, frontend, fullstack). Each has a role definition, operational memory, anti-patterns, and examples. |
| **Context and memory** | Institutional knowledge encoded in version-controlled files: conventions, anti-patterns, past mistakes, service-specific patterns. Two-tier backlog system keeps project state current across sessions. Available on demand, improving over time. |
| **Skills as tools** | Deterministic execution packages (code review skills, coverage analysis, dependency scanning) that agents invoke by judgment when the task warrants them. Skills are tools in the agent's toolkit — not scripts that replace agent reasoning. |
| **MCP tools** | Read-only access to live platform state (catalog config, entity data, service health, coverage metrics) so agents work with real information, not stale assumptions. |
| **Validation gates** | Environmental quality enforcement: build, test, coverage, linting, security scans. Runs deterministically — the same quality bar we hold for human engineers, enforced automatically before any work reaches human review. |
| **Human review** | The final gate. Always. No AI agent merges code without human approval. |

The system is deliberately simple in its early phases. The sophistication is
in the specialization and validation, not in the infrastructure.

**UNM validation finding**: The agent pattern has been proven on the UNM
Platform project with eight specialized agents covering implementation,
review, documentation, and backlog management. Key validation insights:

- Auto-routing orchestration (LLM classifies task and routes to agents) is
  more powerful than per-task-type commands (human selects the agent)
- Parallel agent execution via git worktrees with file ownership rules
  enables true concurrent engineering work
- Memory-based learning (MEMORY.md) works but requires guardrails:
  30-entry cap, monthly curation, proven-useful promotion to anti-patterns
- Two-tier backlog management (AI-managed active + human-owned strategic)
  solves the chronic "AI fails to maintain updated backlogs" problem
- Prompt-based validation is unreliable — agents sometimes skip checks.
  Validation must be environmental (a script the agent invokes as a tool),
  not instructional (a prompt the agent might follow)

### B. AI Engineer as a Human Role

If the AI Engineering capability proves measurable value through Phase 2
milestones (by Q3 2026), INCA should plan to formalize a dedicated AI
Engineer role in H2 2026 to own this capability as a real platform, not a
side project.

This role would be responsible for:
- Designing agent roles and boundaries
- Building and refining context engineering
- Owning evals, telemetry, and success metrics
- Defining tool access, safety controls, and governance
- Partnering with service owners on adoption
- Ensuring the system becomes easier to use than manual alternatives

This matters because the market is formalizing AI engineering as a real
discipline. Major technology companies are listing dedicated AI Engineer
roles, signaling that AI engineering is becoming a first-class function rather
than a side assignment (see Appendix A).

This is a real ambition with an evidence gate. The hire is planned, not
assumed. The decision is gated on the same evidence-based approach that
governs everything in this strategy: prove value, then invest to scale.

---

## 4. What We Give Back to Human Engineers

This vision is ultimately about what we return to the people building INCA.

### We give back focus

Our best engineers should spend less time on repetitive mechanics and more
time on architecture, performance, reliability, scalability, and product
judgment.

*Today*: A senior engineer spends 3 hours investigating why entity-lookup
returns nil for external entities with empty external_id, writes the fix,
adds tests, validates coverage, creates the diff.

*Tomorrow*: The KTLO agent investigates, writes the fix, adds tests, passes
all validation gates, and returns a pre-validated diff for 15 minutes of
review.

### We give back craft

Instead of burning senior attention on flaky tests, dependency bumps,
boilerplate, repetitive KTLO, and review preparation, we let AI prepare
validated work and let humans apply their judgment where it counts.

*Today*: A dependency deprecation affecting 6 services is addressed
sequentially by one engineer over a week.

*Tomorrow*: Multiple agent invocations address each service in parallel
(using git worktrees for isolation), with the engineer reviewing validated
diffs rather than writing each one.

### We give back speed

AI Engineers can operate in parallel, asynchronously, and without waiting for
context switching. That means more issues can move forward at once, especially
low-risk and repetitive work.

*Today*: KTLO tickets queue up during sprint work, addressed in batches when
someone has a gap.

*Tomorrow*: An engineer invokes /ktlo on a ticket between meetings, and the
agent returns a validated diff before the next meeting ends.

### We give back learning

By encoding conventions, anti-patterns, examples, memory, and service context
into the system, we make institutional knowledge available on demand. New
engineers ramp faster. Existing engineers spend less time rediscovering known
patterns.

*Today*: A new engineer joins the feed-worker team and spends weeks absorbing
Cadence patterns, Docstore conventions, and error handling norms from code
review feedback.

*Tomorrow*: The same engineer has agent-curated service context, documented
anti-patterns, and working examples available from day one — the same
knowledge the agents themselves use.

### We give back ambition

When teams are less trapped by maintenance gravity, they have more room for
platform improvements, simplification, quality investments, and stronger
technical leadership.

This is the human promise of AI Engineering: not less engineering, but better
engineering lives.

---

## 5. What "L4-Equivalent Output" Means

We should be disciplined here.

We are not claiming that AI will become a full human L4 engineer in every
sense. We are not claiming broad product judgment, organizational influence,
or open-ended architecture leadership.

We are defining L4-equivalent in a bounded, engineering-operational way.

By the end of the year, an AI Engineer at INCA should be able to, within
defined scope:

1. Take a well-scoped engineering task
2. Identify the relevant code paths and service context
3. Choose a conservative implementation path
4. Follow INCA patterns and conventions
5. Write or update tests to meet coverage requirements
6. Run validation loops (build, test, lint, coverage, security)
7. Produce a diff, summary, and rationale
8. Respect rollout and safety constraints (Flipr gating, Cadence versioning)
9. Hand the result to a human reviewer in a state that is materially easier
   to approve than to recreate manually

That is how we define L4-equivalent output: by execution quality, safety, and
reviewability inside bounded scope. Not by autonomy, not by judgment breadth,
not by organizational impact.

**UNM validation note**: Several of these criteria have been demonstrated
in practice on the UNM Platform for feature development tasks. The
fullstack-engineer agent designs API contracts, implements backend with TDD,
builds frontend, validates both, and produces diffs with summaries — meeting
criteria 1-7 and 9. Criteria 8 (Flipr gating, Cadence versioning) is
INCA-specific and will be validated in Phase 1.

---

## 6. How This Helps with Hiring and Scale

We should be explicit: this is not a workaround for hiring fewer people.

It is a way to make every hire stronger, every current engineer more
effective, and every unit of engineering time more valuable.

With meaningful team growth ahead, we need a way to increase capacity
immediately while preserving quality and accelerating ramp time. We face
two realities:

1. It will take time to hire them well
2. Even after hiring, it will take time for new engineers to absorb INCA's
   patterns, architecture, and conventions

AI Engineering helps on both fronts.

It gives us a way to scale execution capacity before every role is filled,
and it shortens the distance between "new engineer joined" and "new engineer
is productive." Every future engineer joins a team with a stronger support
system, encoded knowledge, and more leverage on day one.

This also becomes a recruiting advantage. The best engineers increasingly
want environments where they can operate at high leverage. A team that uses
AI to remove friction, speed up feedback loops, and preserve human ownership
of hard decisions is more attractive than a team that still expects senior
engineers to spend large portions of their week on preventable mechanics.

---

## 7. How This Vision Sits on Top of the Strategy

The AI Engineering Strategy remains the execution plan. It already makes the
right foundational bets:

- Specialized agents instead of one generic assistant
- Orchestrated briefing packets instead of monolithic context
- Hard validation gates instead of trust
- MCP as controlled context access
- Feedback loops and anti-pattern capture instead of static docs
- Phase-based rollout starting with KTLO
- Explicit go/no-go criteria at each phase boundary

That strategic direction is well aligned with where the market is going. The
major platforms are converging on specialized agents, context engineering,
validation gates, and human-in-the-loop review — exactly the components our
strategy builds on (see Appendix A). The strategy is directionally right.

Where this vision adds value is by making the ambition clearer:

- We are not building a tool
- We are not just improving prompts
- We are building a new layer of engineering capacity for INCA

The strategy says how. This vision says why it matters.

**Validation update**: The UNM Platform project has served as a proving
ground for the agent architecture. It validated the core patterns —
specialized agents, context assembly, MEMORY.md learning, auto-routing
orchestration, parallel execution — in a greenfield context before applying
them to INCA's operational workload. This strengthens the evidence base for
the strategy's approach: the patterns work, and the areas that need
strengthening (validation pipeline enforcement, memory curation, measurement)
are known and addressable.

The updated strategy incorporates several innovations discovered during
validation:

| Innovation | Description | Strategy Section |
|-----------|-------------|-----------------|
| Auto-routing orchestrator | LLM classifies task intent and routes to agents | 3.2 |
| Parallel execution with worktrees | True concurrent agent work with file ownership | 3.10 |
| Two-tier backlog management | AI-managed active list + human-owned roadmap | 3.11 |
| Skills as agent tools | Deterministic skills invoked by agent judgment | 3.6 |
| IDE-agnostic rule syncing | .cursor/rules/ mirroring .claude/ for tool independence | 3.12 |
| Environmental validation | Validation as a tool/script, not a prompt instruction | 3.5 |

---

## 8. Operating Principles for the Next 12 Months

This vision is guided by non-negotiable principles.

### Human judgment stays final

The future is not autonomous merges. The future is human engineers reviewing
much stronger packages of work. Any change to the human-review requirement
requires leadership approval, sustained evidence of quality, and explicit
decision-point authorization (see Strategy, Section 8.1).

### Validation proves, slogans do not

Benchmarks are useful signals, but they do not replace engineering proof.
Even benchmark maintainers have had to revise how coding benchmarks are
interpreted: OpenAI introduced SWE-bench Verified to improve reliability,
then later stated it no longer measures frontier coding capability well due
to contamination. That reinforces the right principle for INCA: benchmarks
inspire; gates prove. Our validation pipeline is the proof, not model release
announcements.

### Security is first-class

As agents gain tool access, the risk surface expands. GitHub's public guidance
is explicit: MCP-connected agents should be constrained with allowlisted,
ideally read-only tools. INCA's MCP server provides read-only platform state.
Agents do not have production credentials, cannot deploy, and cannot modify
live systems. Each new MCP tool requires security review before deployment.

### The system must become easier than the workaround

If AI Engineering feels slow, brittle, or ceremonial, engineers will route
around it. Adoption comes from usefulness, not mandate. The agent workflow
must be faster and more reliable than the manual alternative, or it will be
abandoned. This is a design constraint, not an aspiration.

### We measure pipeline outcomes, not model novelty

The right question is not "Which model is smartest?" The right question is
"Did engineering capacity increase safely and measurably?"

We measure: first-pass success rate, human override rate, KTLO
time-to-resolution, production incident attribution. Not: model benchmarks,
token throughput, or prompt engineering sophistication.

### Intelligence orchestrates, determinism enforces

_New principle — emerged from UNM Platform validation._

The orchestrator uses LLM reasoning to classify tasks, adapt to novel
situations, and decide which agents and tools to use. Validation gates use
deterministic scripts to enforce quality bars. Skills are tools that agents
invoke by judgment — not pipelines that replace agent reasoning.

A fully deterministic orchestrator (Python script routing to hardcoded agents)
handles only cases its author anticipated — it is sophisticated automation,
not an AI Engineer. A fully intelligent orchestrator with no enforcement might
skip critical steps. The solution is intelligence at the decision layer and
determinism at the enforcement layer. The agent reasons about what to do; the
environment guarantees that validation runs.

This principle resolves the tension between reliability and adaptability: the
agent is free to reason about novel tasks, but the validation pipeline catches
mistakes regardless of whether the agent "remembered" to check.

---

## 9. Year-One Milestones

The first 6 months follow the phased rollout defined in the AI Engineering
Strategy, with explicit go/no-go criteria at each phase boundary. Months 7-12
are contingent on Phase 3 results.

### Q2 2026: Prove the First Capability

Ship the first KTLO Engineer workflow end-to-end. Establish trust through
conservative scope, strong validation, and visible wins.

**Success**: First-pass success rate >60% on 5+ real KTLO tickets. Framework
validated — adding new agents requires only briefing files, not infrastructure
changes.

**If it doesn't work**: Iterate on context quality and validation gates. If
fundamentally blocked after 4 additional weeks, reassess and publish learnings.

### Q3 2026: Expand from Single Agent to Coordinated System

Add Test Writer and Code Reviewer agents. Implement agent chaining. Deepen
service-specific context. Begin measuring pipeline outcomes. Introduce
auto-routing orchestrator and parallel execution.

**Success**: Chained pipeline works end-to-end. 2+ weeks of measurement data.
First-pass rate >50% for the full chain. If measurable value demonstrated,
begin planning the dedicated AI Engineer role for H2.

**If it doesn't work**: Continue with KTLO alone. Investigate whether
chaining or individual agent quality is the bottleneck.

### Q4 2026: Operationalize the Platform

Move from promising prototype to dependable workflow. Establish usage
patterns, weekly metrics, safety review, and clear ownership. Add Bug
Investigator, Feature Builder, and Migration Specialist agents.

**Success**: Measurable reduction in KTLO time-to-resolution. >50% of
eligible KTLO tickets handled with AI assistance. Feedback loop producing
anti-pattern updates. Dashboard live.

**If it doesn't work**: Narrow the catalog to proven agents. Focus on context
quality and validation reliability rather than breadth.

### Q1 2027: Lock In the Model

If milestones have been met, transition AI Engineering from initiative to
platform. If the hire was approved, the AI Engineer owns and scales the
capability. Establish cross-team sharing patterns.

**Success**: INCA leadership can describe AI Engineering as an operating
capability with measurable output, known controls, and clear owners — not an
experiment.

---

## 10. What Success Looks Like in 12 Months

By March 2027, success should look like this:

- INCA has a functioning AI Engineer capability that handles a meaningful
  share of eligible KTLO and adjacent engineering work through a governed,
  validated workflow.
- Human engineers experience a visible reduction in repetitive execution
  burden and a visible increase in time spent on design, reliability,
  platform improvements, and technical leadership.
- New engineers ramp faster because service conventions, examples, and
  institutional knowledge are available through the system instead of living
  only in people's heads.
- Quality has not degraded. Zero increase in production incidents attributed
  to AI-assisted changes. Coverage remains stable or improves. The validation
  pipeline has proven that AI output meets the same bar as human output.
- INCA leadership can talk about AI Engineering not as experimentation, but as
  an operating capability with measurable output, known controls, and clear
  owners.
- The agent framework is IDE-agnostic — it works in Claude Code, Cursor, and
  any future tools through synced rule systems.
- Institutional knowledge is encoded, curated, and continuously improving
  through the memory/anti-pattern/feedback loop cycle.

And if the phased milestones supported the decision, INCA has hired or is
hiring its first AI Engineer to own and scale the capability for the long term.

---

## 11. Execution Strategy

The AI Engineering Strategy is the execution strategy for this vision. It
translates the ambition set here into a phased plan for delivery, with clear
architecture, controls, milestones, and decision points.

Where this vision defines the future we want to create for INCA and its
engineers, the strategy defines how we will build toward it: through
specialized agents, orchestrated context, validation-first workflows,
controlled platform access, feedback loops, and measurable rollout phases.

Together, the two documents serve distinct but aligned purposes:

- This vision defines the direction, intent, and desired outcomes.
- The AI Engineering Strategy defines the execution path, operating model,
  and evidence gates required to realize that direction responsibly.

---

## 12. Leadership Ask

Approve this as the one-year north star for INCA.

Specifically:

1. **Treat the AI Engineering Strategy as the build plan.** It defines the
   architecture, phased rollout, resource requirements, and go/no-go criteria.
2. **Treat AI Engineer as a new engineering capability, not just a developer
   tool.** This is a new unit of capacity, not a feature of the IDE.
3. **Plan to formalize a dedicated AI Engineer role in H2 2026** if Phase 2
   milestones demonstrate measurable value by Q3. A real ambition with an
   evidence gate — not a soft maybe.
4. **Evaluate success based on outcomes**: reclaimed human engineering
   capacity, reduced KTLO drag, improved onboarding leverage, and preserved
   quality. Not on model benchmarks or demo impressiveness.

This is the shift we should aim for: From AI as assistant → to AI as
teammate → to AI as managed engineering capacity.

That is the next chapter of engineering for INCA.

---

## Appendix A: Market Validation

This appendix provides specific external evidence supporting the vision's
direction.

### Platform Direction

| Company | Signal | Relevance to INCA |
|---------|--------|-------------------|
| GitHub | Copilot moving from "pair programmer" to "peer programmer." Agents solve multi-step problems, test their own work, return results for review. Productizing custom agents, file-based specialization, and context engineering. | Directly validates our architecture: specialized agents, orchestrated context, validation-first design. |
| Google | Jules operates as an asynchronous coding agent in a secure cloud VM. Fixes bugs, writes tests, returns plan plus diff for review. | Validates asynchronous engineering teammate model and "hand back validated work" pattern. |
| AWS | Amazon Q Developer's agentic experience moves beyond suggestions into file changes, diffs, and command execution. | Validates the action layer: AI that modifies code and runs tooling, not just AI that suggests. |
| Microsoft | Copilot Studio supports multi-agent orchestration, agent identity, and MCP integration. 230,000+ organizations use it. | Validates orchestration at enterprise scale and multi-agent coordination. |

### The AI Engineer as a Human Role

Major technology companies are increasingly formalizing AI Engineer roles,
signaling that AI engineering is emerging as a recognized discipline rather
than a temporary side assignment.

### Security Posture

The more agentic the system becomes, the more important it is to constrain
tool use, preserve human review, and build attribution and governance into the
design from day one. GitHub's public guidance is explicit: MCP-connected
agents should be constrained with allowlisted, ideally read-only tools,
designed to maximize interpretability, minimize autonomy, and reduce anomalous
behavior. This is exactly the posture INCA's strategy adopts.

---

## Appendix B: How This Vision Maps to the AI Engineering Strategy

| Vision Section | Strategy Section | Relationship |
|---------------|-----------------|-------------|
| North Star (Section 2) | Vision (Section 2.1) | Vision provides aspirational framing; strategy provides concrete definition |
| AI Engineer Capability (Section 3A) | Architecture (Section 3) | Vision describes capability model; strategy defines orchestrator, agents, validation, and feedback in detail |
| AI Engineer Role (Section 3B) | Not covered | Vision introduces the human role; strategy focuses on the system |
| What We Give Back (Section 4) | Not covered | Vision provides human-impact narrative; strategy focuses on technical execution |
| L4-Equivalent Output (Section 5) | KTLO Agent Definition (Section 3.3) | Vision frames quality standard; strategy defines the agent that delivers it |
| Hiring & Scale (Section 6) | Not covered | Vision provides business context; strategy focuses on technical delivery |
| Operating Principles (Section 8) | Design Principles (Section 2.2), Governance (Section 8) | Aligned; vision states leadership commitments, strategy operationalizes them |
| Year-One Milestones (Section 9) | Roadmap (Section 5), Decision Points (Section 10) | Vision provides quarterly framing; strategy provides week-level detail |
| Leadership Ask (Section 12) | Resource Requirements (Section 9) | Vision asks for strategic approval; strategy asks for specific resources |

### UNM Platform Validation Coverage

| Vision Component | UNM Status | Notes |
|-----------------|-----------|-------|
| Specialized agents | Validated (8 agents built) | Works for feature dev, not only KTLO |
| Orchestrator (auto-routing) | Validated | `/run` command classifies and routes successfully |
| Context and memory | Validated (needs guardrails) | Memory works but requires cap/curation policy |
| Skills as tools | Analyzed (not yet integrated) | Architecture defined; integration is Phase 2+ |
| MCP tools | Not applicable | UNM has no MCP server; INCA already has one |
| Validation gates | Partially validated | Prompt-based; needs environmental enforcement |
| Human review | Validated (PR-based) | Git-flow rules enforce PR-based workflow |
| Parallel execution | Validated | Git worktrees with file ownership |
| Two-tier backlog | Validated | Backlog-manager agent with completion hook |
| IDE syncing | Validated | .cursor/rules/ mirrors .claude/ |
| L4-equivalent output | Partially demonstrated | Criteria 1-7, 9 demonstrated for feature dev |
