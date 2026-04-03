# UNM Platform — Architecture Evolution

_Design document for persistence, multi-tenancy, collaboration, and auth.
These are interdependent concerns that must be designed together before
any implementation begins._

**Status:** Approved
**Owner:** Kristian Zachariassen
**Created:** 2026-04-01

---
yes
## 1. Why This Document Exists

The current platform is single-user, in-memory, and stateless across
restarts. Two independent reviews identified persistence, auth, and
collaboration as the critical path to product readiness. These are not
independent features — they are a single architectural evolution:

- **Persistence** without multi-tenancy means rebuilding the schema later
- **Auth** without an authorization model means bolting on permissions later
- **Multi-model** without workspaces means inventing organization later
- **Collaboration** without team isolation means retrofitting access control

This document defines the target data model, tenancy boundaries, auth
model, API structure, and collaboration patterns **before** writing the
first migration. The goal: build once, correctly.

---

## 2. Data Hierarchy

```
Organization (tenant boundary)
├── Members (users with org-level roles)
├── Workspaces (isolation units within an org)
│   ├── Members (users with workspace-level roles)
│   └── Models
│       ├── Versions (immutable snapshots)
│       ├── Changesets (proposed changes)
│       │   └── Reviews / Comments
│       └── Analysis Cache
└── Settings (org-level config, AI keys, etc.)
```

### Why this hierarchy?

**Organization** is the tenant boundary. Data never leaks across orgs. This
is the billing unit, the SSO boundary, and the top-level isolation guarantee.

**Workspace** is the collaboration unit within an org. A workspace groups
related models. Examples:
- "Platform Team" workspace with the platform UNM model
- "Eats Engineering" workspace with the delivery org model
- "Architecture Review" workspace shared across teams for org-wide modeling

Workspaces solve the "team isolation + selective sharing" problem: teams
own their workspaces, but an org admin can create cross-team workspaces.

**Model** is a single UNM model (one `.unm` file = one model). Models live
in exactly one workspace.

**Version** is an immutable snapshot of a model at a point in time. Created
on changeset commit or explicit "save version." Enables history, diff, and
rollback.

---

## 3. Entity Definitions

### Organization

```
Organization
  id              UUID
  name            string          "Acme Engineering"
  slug            string          "acme-eng" (URL-safe, unique)
  created_at      timestamp
  updated_at      timestamp
```

### User

Users are external identities (Google OAuth). They exist independently of
organizations and can belong to multiple orgs.

```
User
  id              UUID
  email           string          unique, from OAuth provider
  name            string          display name
  avatar_url      string          from OAuth provider
  created_at      timestamp
  last_login_at   timestamp
```

### Organization Membership

```
OrgMembership
  org_id          UUID            FK → Organization
  user_id         UUID            FK → User
  role            enum            owner | admin | member
  joined_at       timestamp

  UNIQUE(org_id, user_id)
```

Roles:
- **owner**: full control, billing, can delete org. At least one per org.
- **admin**: manage members, create/delete workspaces, manage settings.
- **member**: access workspaces they're added to.

### Workspace

```
Workspace
  id              UUID
  org_id          UUID            FK → Organization
  name            string          "Platform Team"
  slug            string          URL-safe, unique within org
  visibility      enum            private | org-visible
  created_by      UUID            FK → User
  created_at      timestamp
  updated_at      timestamp
```

Visibility:
- **private**: only workspace members can access
- **org-visible**: any org member can view (read-only); only workspace
  members can edit

### Workspace Membership

```
WorkspaceMembership
  workspace_id    UUID            FK → Workspace
  user_id         UUID            FK → User
  role            enum            admin | editor | viewer
  joined_at       timestamp

  UNIQUE(workspace_id, user_id)
```

Roles:
- **admin**: manage workspace members, delete models
- **editor**: create/edit models, create/commit changesets
- **viewer**: read-only access to models and analysis

### Model

```
Model
  id              UUID
  workspace_id    UUID            FK → Workspace
  name            string          derived from system.name in the UNM file
  description     string          derived from system.description
  current_version int             points to latest committed version
  source_format   enum            dsl | yaml
  created_by      UUID            FK → User
  created_at      timestamp
  updated_at      timestamp
```

### Model Version

Immutable. Once created, never modified.

```
ModelVersion
  id              UUID
  model_id        UUID            FK → Model
  version         int             monotonically increasing
  raw_content     text            the .unm or .unm.yaml source
  source_format   enum            dsl | yaml
  commit_message  string          optional, human or changeset-generated
  committed_by    UUID            FK → User
  committed_at    timestamp
```

### Changeset

```
Changeset
  id              UUID
  model_id        UUID            FK → Model
  base_version    int             which model version this was created against
  actions_json    jsonb           serialized []ChangeAction
  status          enum            draft | committed | rejected
  title           string          optional
  created_by      UUID            FK → User
  created_at      timestamp
  updated_at      timestamp
```

### Changeset Comment (future — collaboration)

```
ChangesetComment
  id              UUID
  changeset_id    UUID            FK → Changeset
  user_id         UUID            FK → User
  body            text
  created_at      timestamp
```

---

## 4. Tenancy & Isolation Model

### Data isolation

**Hard boundary:** Organization. All queries include `org_id` in the WHERE
clause (or join through workspace → org). No cross-org data access, ever.

**Soft boundary:** Workspace. Within an org, workspaces provide access
control but not data isolation — an org admin can see all workspaces.

### Implementation strategy

**Row-level tenancy** (not schema-per-tenant). All orgs share the same
database and tables. Every query that touches user data is scoped by org_id
through the workspace → org chain.

Why not schema-per-tenant: we're not building for thousands of tenants with
strict compliance requirements. Row-level is simpler, cheaper, and sufficient
for a product serving engineering teams.

### Query scoping pattern

Every repository method that returns user data takes a context carrying the
authenticated user's org_id. The repository layer enforces scoping:

```go
type ModelRepository interface {
    Store(ctx context.Context, workspaceID uuid.UUID, model *entity.UNMModel) (uuid.UUID, error)
    Get(ctx context.Context, modelID uuid.UUID) (*StoredModel, error)
    List(ctx context.Context, workspaceID uuid.UUID) ([]*StoredModel, error)
    // ...
}
```

The `Get` implementation joins through workspace to verify the model belongs
to the caller's org. This is enforced at the repository layer, not the
handler layer — defense in depth.

---

## 5. Authorization Model

### Decision: middleware + context, not per-handler checks

Auth is enforced in two layers:

1. **Auth middleware** (request level): verifies session, loads user + org
   membership into request context. Rejects unauthenticated requests with 401.
2. **Authorization middleware** (resource level): checks workspace membership
   and role before allowing access. Returns 403 on insufficient permissions.

Handlers never check permissions directly. They receive a context that is
already authorized. This prevents the common bug where one handler forgets
to check.

### Permission matrix

| Action | Org Owner/Admin | Workspace Admin | Editor | Viewer |
|--------|:-:|:-:|:-:|:-:|
| Create workspace | Yes | — | — | — |
| Delete workspace | Yes | Yes | — | — |
| Add workspace member | Yes | Yes | — | — |
| Create model | Yes | Yes | Yes | — |
| Edit model / commit changeset | Yes | Yes | Yes | — |
| View model / analysis | Yes | Yes | Yes | Yes |
| Delete model | Yes | Yes | — | — |
| View model history | Yes | Yes | Yes | Yes |
| Manage org members | Yes | — | — | — |
| Manage org settings | Yes | — | — | — |

### API key access (future)

For CI/CD integration (GitHub Action validating `.unm` files), the platform
will need API key auth in addition to OAuth sessions. API keys are scoped to
a workspace with a specific role (typically `editor` for CI writes, `viewer`
for CI reads). This is a Phase 16+ concern but the auth middleware should be
designed to accept both session cookies and API key headers from the start.

---

## 6. API Route Evolution

### Current routes (single-user, no scoping)

```
POST   /api/models/parse
GET    /api/models/{id}/views/{viewType}
POST   /api/models/{id}/changesets
...
```

### Target routes (org + workspace scoping)

```
# Auth
GET    /auth/google
GET    /auth/callback
POST   /auth/logout
GET    /api/me

# Organizations
GET    /api/orgs                              list user's orgs
POST   /api/orgs                              create org
GET    /api/orgs/{orgSlug}                    get org details
PUT    /api/orgs/{orgSlug}/members            manage members

# Workspaces
GET    /api/orgs/{orgSlug}/workspaces         list workspaces
POST   /api/orgs/{orgSlug}/workspaces         create workspace
GET    /api/orgs/{orgSlug}/ws/{wsSlug}        get workspace details
PUT    /api/orgs/{orgSlug}/ws/{wsSlug}/members manage members

# Models (scoped to workspace)
POST   /api/orgs/{orgSlug}/ws/{wsSlug}/models/parse
GET    /api/orgs/{orgSlug}/ws/{wsSlug}/models
GET    /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}
GET    /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/views/{viewType}
GET    /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/history
GET    /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/versions/{v}
POST   /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/changesets
...

# Config / Health (unscoped)
GET    /health
GET    /api/config
```

### Migration path

The existing handler logic (parse, view, analyze, changeset) stays the same.
What changes:

1. Routes get workspace prefix
2. Handlers receive workspace context from middleware
3. Repository calls include workspace scoping
4. Model IDs are UUIDs, not random hex tokens

The existing `Handler` struct and its methods do not need rewriting — they
need wrapping with context injection.

### Local development mode

When `auth.enabled: false` (local dev), the platform creates a default
org ("local"), default workspace ("default"), and default user ("local").
All existing routes work without the org/workspace prefix — middleware
injects the default context. This preserves the current development
experience.

---

## 7. Collaboration Model

### What "collaboration" means for UNM

UNM models are architectural artifacts, not real-time documents. The
collaboration pattern is closer to **git** (propose, review, merge) than
to **Google Docs** (concurrent cursor editing).

### Changeset-as-PR workflow

The changeset system already implements propose → preview → commit. Adding
multi-user collaboration means:

1. **Multiple users can view the same model** (already works — models are
   shared within a workspace)
2. **Any editor can create a changeset** (proposed change to the model)
3. **Other editors/admins can comment on the changeset** (ChangesetComment)
4. **Commit requires workspace admin or changeset creator** (authorization)
5. **Committed changeset creates a new version** (immutable history)

This is sufficient for v1 collaboration. Real-time concurrent editing is
explicitly out of scope — it requires conflict resolution, operational
transforms, and a fundamentally different architecture. The value of UNM
Platform is analysis and evolution, not real-time co-authoring.

### What we explicitly defer

- Real-time concurrent editing (OT/CRDT)
- Cursor presence ("who is looking at this model")
- Notifications (Slack/email on changeset creation)
- Comments on specific entities within a model (only on changesets)

---

## 8. Database Choice

### Decision: PostgreSQL

SQLite is tempting for single-binary simplicity but fails on:
- Concurrent write access from multiple users
- Row-level locking for changeset operations
- JSON/JSONB queries for changeset actions
- Future: full-text search on model content
- Future: hosted deployment with connection pooling

PostgreSQL supports all of the above and is the standard choice for
multi-tenant web applications.

### Single-binary deployment option

For local dev and small deployments, embed PostgreSQL via Docker Compose
(as today for the app itself). The production deployment adds a managed
PostgreSQL instance. The application binary stays single — it connects to
PostgreSQL via `DATABASE_URL`.

### Schema migration

Use `golang-migrate` with embedded SQL migration files. Migrations run
automatically on startup (with a flag to disable for production
blue/green deploys).

### Config

```yaml
storage:
  driver: postgres                 # postgres | memory
  database_url: ""                 # set via UNM_STORAGE__DATABASE_URL env var
  max_connections: 20
  migrate_on_startup: true         # false for production
```

When `driver: memory`, the existing in-memory stores are used (for tests
and quick local experiments). This preserves backward compatibility with
the current development workflow.

---

## 9. Domain Impact

### What changes in the domain layer

**Nothing.** The domain layer (`entity/`, `service/`, `valueobject/`) stays
pure. `UNMModel`, `Changeset`, `ChangesetApplier`, `ValidationEngine` — none
of these know about persistence, users, or organizations. This is the Clean
Architecture payoff.

### What changes in the adapter/infrastructure layers

| Layer | Change |
|-------|--------|
| **Repository** | New interface definitions in `usecase/`. New PostgreSQL implementations in `infrastructure/persistence/`. |
| **Handler** | New auth routes. Existing routes wrapped with workspace context. Model IDs become UUIDs. |
| **Middleware** | Auth middleware (session verification). Authorization middleware (workspace role check). |
| **Config** | New `auth` and `storage` config sections. |
| **Main** | Wire PostgreSQL stores. Run migrations. Construct auth handlers. |

### What changes in the frontend

| Area | Change |
|------|--------|
| **Auth** | Login page, auth context, protected route wrapper |
| **Routing** | Routes include org/workspace slugs |
| **API client** | Base URL includes org/workspace path |
| **Model context** | Models loaded from persistent store, not localStorage |
| **New pages** | Org settings, workspace management, model list, model history |
| **Changeset UI** | Author attribution, comment thread (future) |

---

## 10. Implementation Phases (revised)

The backlog phases should reflect this architecture:

### Phase 14A: Repository Interfaces & PostgreSQL Foundation
Define interfaces. Set up PostgreSQL with migrations. Implement model
and changeset stores. Wire into main.go with `storage.driver` config.
Verify existing tests pass with both memory and postgres drivers.

### Phase 14B: Model History & Multi-Model
Model versions on commit. List models. Version history API. Frontend
model selector.

### Phase 15A: Auth Foundation
Google OAuth flow. Session management. Auth middleware. User table and
creation on first login. Frontend login page and auth context.
Local dev mode with auth disabled.

### Phase 15B: Organizations & Workspaces
Org and workspace tables. Membership tables. Management APIs. Frontend
org/workspace UI. Route migration to scoped paths. Default org/workspace
for local dev.

### Phase 15C: Authorization
Workspace role checks. Permission matrix enforcement. Changeset ownership.

### Phase 16+: Collaboration
Changeset comments. Author attribution in UI. API key auth for CI.
Notification hooks.

---

## 11. Architecture Decisions (Resolved)

_All decisions finalized 2026-04-01._

### AD-1: Org creation flow → Auto-create on first login

When a user signs in for the first time, the system automatically creates
a personal organization named after the user (e.g., "Kristian's Org").
No manual "create org" step. The user lands in their org immediately with
zero setup friction. They can rename it, invite others, or create
additional orgs later.

**Schema impact:** `POST /auth/callback` creates User row + Organization
row + OrgMembership (role: owner) in a single transaction.

### AD-2: Default workspace → Yes, auto-created "General" workspace

Every new org gets a "General" workspace with `visibility: org-visible`.
All org members can see models in it. This means a first-time user can
upload a model within seconds of signing in — no workspace creation step.

Private team workspaces can be created later for team isolation.

**Schema impact:** Org creation transaction also creates the default
workspace + WorkspaceMembership for the owner.

### AD-3: Model transfer → Within org only

Models can move between workspaces **within the same org** (admin action).
Version history is preserved. Cross-org transfer is not supported — users
export and re-import. This preserves the org as a hard data isolation
boundary and avoids audit confusion.

**API:** `POST /api/orgs/{slug}/ws/{wsSlug}/models/{id}/move`
with `{target_workspace_slug: "..."}`.

### AD-4: Deletion semantics → Soft delete with 30-day retention

All deletions are soft: a `deleted_at` timestamp is set, the entity is
hidden from all queries, and a background job purges records older than
30 days. This prevents accidental loss of months of version history.

Workspace/org deletion cascades soft-delete to all contained resources.
Admin can restore within the 30-day window.

**Schema impact:** Add `deleted_at TIMESTAMPTZ NULL` to organizations,
workspaces, models, model_versions, changesets. All queries add
`WHERE deleted_at IS NULL`.

### AD-5: Single-player mode → Personal org IS single-player

No separate "solo mode." The auto-created personal org is the single-user
experience. Same code path, same schema, same API. One user in an org is
functionally identical to single-player — the user doesn't see "org"
language until they invite someone. This avoids maintaining two code paths.

The existing `auth.enabled: false` local dev mode (Phase 15A.7) creates
a default local org/workspace/user for development — that's the only
special case.

### AD-6: AI key management → BYOK (bring your own key) per org

Org admins set their OpenAI API key in org settings. The platform never
provides or bills for AI usage in v1. This avoids the entire billing/
metering/margin question until the business model is validated.

**Schema impact:** Add `ai_api_key_encrypted TEXT` to organizations table.
Encrypted at rest with a platform key.

**Config impact:** The global `ai.api_key_env` config remains as a
fallback for local dev and self-hosted deployments. In multi-tenant mode,
per-org keys take precedence.

---

## 12. What This Document Does NOT Cover

- **Billing and subscription management** — deferred until pricing model
  is validated
- **SSO / SAML** — enterprise feature, post-MVP
- **Audit logging** — important for enterprise, but not v1
- **Rate limiting** — needed for hosted deployment, not for initial release
- **Backup and disaster recovery** — operational concern, not architecture
- **Multi-region deployment** — premature optimization
