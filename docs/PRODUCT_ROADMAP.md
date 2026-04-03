# UNM Platform — Product Roadmap

_What we ship to users, organized by the value it delivers._
_For engineering tasks and implementation details, see `docs/BACKLOG.md`._

**Vision:** Turn User Needs Mapping from a workshop artifact into a living
engineering tool — versioned, analyzable, continuously updated as the
system and organization evolve.

**Positioning:** We are not building a better way to do UNM. We are
building a way for UNM to actually live and matter over time.

---

## What exists today (v0 — Solo Explorer)

A single-user architecture explorer. Upload a model, get instant analysis.

| Capability | What users can do |
|-----------|-------------------|
| **Model authoring** | Write architecture models in `.unm` DSL — actors, needs, capabilities, services, teams, external deps, team interactions |
| **Upload & parse** | Upload `.unm` or `.unm.yaml` files; instant validation with errors and warnings |
| **Architecture visualization** | 8 interactive views: UNM Map (full graph), Need View, Capability View, Ownership View, Team Topology, Cognitive Load, Realization View, Signals View |
| **Health signals** | 24+ automated findings across UX, architecture, and organization layers — fragmented capabilities, cross-team needs, bottleneck services, cognitive overload |
| **AI advisor** | Ask natural-language questions about your architecture; get context-aware answers |
| **What-If explorer** | Propose structural changes (move a service, add a team) and see the impact before committing |
| **AI recommendations** | Generate a full restructuring report with prioritized actions |
| **Model export** | Download your model as `.unm` or `.unm.yaml` |

### Limitations

- In-memory only — models are lost on server restart
- Single user — no login, no teams, no sharing
- Manual upload only — no integration with source control
- No model history — can't see how your architecture evolved

---

## Milestone 1 — Persistent Platform

_Your models survive a restart. Multiple models in one place._

| Feature | User value |
|---------|-----------|
| **Durable storage** | Models are saved to a database — come back tomorrow and your work is still there |
| **Multi-model** | Manage multiple architecture models, not just one at a time |
| **Model history** | Every committed changeset creates a version — see how your architecture evolved |
| **Version diff** | Compare two versions side-by-side to see what changed |

---

## Milestone 2 — Team Platform

_Your whole team uses it. Real identity, shared workspaces._

| Feature | User value |
|---------|-----------|
| **Sign in with Google** | Secure login via Google OAuth — no passwords to manage |
| **Organizations** | Group your team under an org. First login auto-creates a personal org |
| **Workspaces** | Separate models by domain, team, or project. Default "General" workspace on signup |
| **Roles & permissions** | Owner, Admin, Editor, Viewer — control who can change what |
| **User attribution** | See who uploaded a model, proposed a change, or committed a version |

---

## Milestone 3 — Collaborative Architecture

_Propose, discuss, review, and approve changes — like PRs for your architecture._

| Feature | User value |
|---------|-----------|
| **Changeset comments** | Discuss proposed changes with your team — threaded comments on any changeset |
| **Review workflow** | Changesets go through draft → in review → approved/rejected, just like code review |
| **Approval gates** | Only workspace admins or designated reviewers can approve structural changes |
| **Activity feed** | See what changed, who changed it, and when — across your workspace |

---

## Milestone 4 — Engineering Integration

_Your architecture model lives in your engineering workflow, not beside it._

| Feature | User value |
|---------|-----------|
| **API keys** | Programmatic access for CI/CD pipelines — authenticate without a browser |
| **GitHub Action** | Validate `.unm` files in pull requests — CI fails when the architecture model is invalid |
| **Git import** | Point the platform at a repo; it finds and imports `.unm` files automatically |
| **Webhook sync** | Push to main → model auto-updates in the platform |
| **CLI validate** | Run `unm validate model.unm` locally or in CI, get machine-readable JSON output |

---

## Milestone 5 — Adoption & Onboarding

_Going from zero to first insight in minutes, not hours._

| Feature | User value |
|---------|-----------|
| **AI model generator** | Describe your system in plain text → AI generates a starter `.unm` model |
| **Getting-started wizard** | Answer questions about your teams, services, and users → build a model step by step |
| **Template library** | Start from a pre-built pattern: microservices org, platform team, startup, monolith migration |
| **Import from Backstage** | Already have a `catalog-info.yaml`? Import services and teams automatically |
| **Import from Structurizr** | Have a C4 model? Map it to UNM and keep going |

---

## Milestone 6 — Product Hardening

_Trustworthy analysis. Flexible output. Reliable editing._

| Feature | User value |
|---------|-----------|
| **Model completeness score** | Know how much of your architecture is modeled — analysis confidence tied to coverage |
| **Configurable thresholds** | Tune what "overloaded" or "bottleneck" means for your org |
| **Mermaid & JSON export** | Share architecture diagrams in docs, wikis, or other tools |
| **Changeset undo** | Made a mistake? Revert to a previous version |
| **Slack notifications** | Get notified when someone proposes or commits a change |

---

## Future (exploring)

These are ideas, not commitments. They'll be shaped by user feedback.

- **Real-time collaborative editing** — multiple users editing the same model simultaneously (deferred — changeset-as-PR is the v1 pattern)
- **Transition planning** — model a future-state architecture and generate a migration plan with sequenced steps
- **Model marketplace** — share and discover architecture patterns across organizations
- **Rich Allen integration** — import models from Rich's collaborative workshop tool; UNM Platform becomes the "living system" layer downstream

---

## How milestones map to engineering

| Milestone | Backlog phases | Key dependency |
|-----------|---------------|----------------|
| Persistent Platform | Phase 14A, 14B | — |
| Team Platform | Phase 15A, 15B, 15C | Persistence |
| Collaborative Architecture | Phase 16 | Auth + Tenancy |
| Engineering Integration | Phase 16.3, 18.1 | Auth |
| Adoption & Onboarding | Phase 18.2, 18.3 | Core platform |
| Product Hardening | Phase 17 | Ongoing |

_For detailed engineering tasks, see `docs/BACKLOG.md`._
