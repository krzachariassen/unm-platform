# How to Write a UNM Model in YAML

This guide walks you through writing a `.unm.yaml` model file from scratch.
It is practical and tutorial-style. For the full reference spec, see
[UNM_DSL_SPECIFICATION.md](UNM_DSL_SPECIFICATION.md).

> **Tip:** The `.unm` DSL format is more concise and git-friendly. If you're
> starting fresh, consider using the [DSL Guide](DSL_GUIDE.md) instead. Both
> formats produce the same internal model and are fully supported.

---

## What You're Building

A UNM model captures the vertical value chain from user needs down to the
services and teams that fulfill them. The end result is a structured,
queryable, versionable artifact — not a diagram, but the data that generates
diagrams automatically.

The chain looks like this:

```
Actor → Need → Capability → Service → Team
```

You work top-down: start with who the users are and what they need, then
describe what your system must be able to do (capabilities), then show which
services and teams make that possible.

---

## Minimal Working Example

A valid model has at least one actor, one need, one capability, one service,
and one team. Here is the smallest model that passes validation:

```yaml
system:
  name: "Order Platform"
  description: "Manages the order lifecycle from placement to fulfillment"

actors:
  - name: "Customer"
    description: "End user placing and tracking orders"

needs:
  - name: "Place an order"
    actor: "Customer"
    outcome: "My order is confirmed and a restaurant is notified"
    supportedBy:
      - "Order Submission"

capabilities:
  - name: "Order Submission"
    description: "Accept and validate incoming orders"
    visibility: "user-facing"
    realizedBy:
      - target: "order-api"
        role: "primary"
        description: "REST API that accepts and validates order payloads"

services:
  - name: "order-api"
    description: "Public-facing order intake service"
    ownedBy: "order-team"

teams:
  - name: "order-team"
    type: "stream-aligned"
    description: "Owns end-to-end order placement and confirmation"
```

That's a complete model. The sections below explain how to grow it.

---

## Section by Section

### `system`

The top-level boundary. Describes the scope of what you're modeling.

```yaml
system:
  name: "Catalog Platform"
  description: "Manages product catalog data across ingestion, storage, and serving"
```

---

### `actors`

People or external systems that have needs. Be specific — each actor should
represent a meaningfully different user type with different goals.

```yaml
actors:
  - name: "Merchant"
    description: "Restaurant owner managing menu data"

  - name: "Eater"
    description: "End user browsing and ordering from restaurants"

  - name: "Downstream Platform Team"
    description: "Internal teams consuming catalog data via API or events"
```

---

### `needs`

What each actor is trying to accomplish. A need has an `actor`, a short
`name`, and an `outcome` — a first-person statement of what success looks like
for the actor.

The `supportedBy` field links needs to capabilities. Use short form (just the
capability name) or long form (with a description of how this capability helps).

```yaml
needs:
  - name: "Publish menu changes within SLA"
    actor: "Merchant"
    outcome: >
      My menu updates are live for customers within minutes of submission,
      without manual intervention.
    supportedBy:
      - "Feed Ingestion"
      - target: "Entity Publishing"
        description: "Validates and publishes entities to the serving layer"
```

**Rules:**
- An actor must exist in the `actors` list.
- Every need must reference at least one capability via `supportedBy`.
- The capability names in `supportedBy` must match a capability defined in
  `capabilities` (or a child capability).

---

### `capabilities`

What the system must be able to do. Capabilities are the core of a UNM model —
they represent durable system abilities, not services or teams.

Every capability needs a `visibility` level that places it in the UNM value
chain:

| Visibility | Meaning |
|------------|---------|
| `user-facing` | Directly experienced by end users |
| `domain` | Core business processing, not directly user-visible |
| `foundational` | Internal capabilities underpinning domain logic |
| `infrastructure` | Deep infrastructure, fully invisible to users |

#### Flat capability (no children)

```yaml
capabilities:
  - name: "Feed Ingestion"
    description: "Accept catalog data from merchant feeds"
    visibility: "user-facing"
    realizedBy:
      - target: "feed-worker"
        role: "primary"
        description: "Workflow worker that processes merchant feeds"
```

#### Hierarchical capabilities (parent groups children)

Group related capabilities under a parent. **Only leaf capabilities have
`realizedBy`.** Parent capabilities are realized through their children.

```yaml
capabilities:
  - name: "Catalog Management"
    description: "Full lifecycle management of catalog entities"
    visibility: "domain"
    children:
      - name: "Entity CRUD"
        description: "Create, read, update, delete on catalog entities"
        visibility: "foundational"
        realizedBy:
          - target: "entity-store"
            role: "primary"
            description: "Authoritative CRUD store"
      - name: "Bulk Editing"
        description: "High-throughput edits across many entities"
        visibility: "domain"
        realizedBy:
          - target: "bulk-worker"
            role: "primary"
            description: "Async bulk operation worker"
```

#### `realizedBy` — short form vs long form

Both forms can be mixed freely in the same list:

```yaml
realizedBy:
  - "entity-store"                       # short form — just the service name
  - target: "registry"                   # long form — adds description and role
    role: "supporting"
    description: "Provides catalog context for entity lookups"
```

**Roles:**
- `primary` — main implementation or owner of this capability
- `supporting` — contributes to but does not own
- `consuming` — uses the capability as a client

---

### `services`

Concrete implementations. A service declares its owner and its dependencies on
other services. **It does not declare which capabilities it supports** — that
relationship is defined on the capability side via `realizedBy`.

```yaml
services:
  - name: "entity-store"
    description: "Authoritative CRUD store for all catalog entities"
    ownedBy: "catalog-core"
    dependsOn:
      - target: "registry"
        description: "Catalog context lookup on every entity operation"
      - target: "workflow-engine"       # short form also valid
```

**What NOT to include on services:**
- `supports` — declare this on capabilities via `realizedBy`
- `dataAssets` — declare this on `data_assets` via `usedBy`
- `externalDependsOn` — declare this on `external_dependencies` via `usedBy`

---

### `teams`

Organizational units that own capabilities and services. Use Team Topologies
types to classify each team.

**Types:** `stream-aligned`, `platform`, `enabling`, `complicated-subsystem`

```yaml
teams:
  - name: "catalog-core"
    type: "platform"
    description: "Provides shared catalog infrastructure to stream-aligned teams"
    owns:
      - "Entity CRUD"
      - "Catalog Registry"

  - name: "ingestion-team"
    type: "stream-aligned"
    description: "Owns feed ingestion and processing end-to-end"
    owns:
      - "Feed Ingestion"
```

---

### `interactions`

How two teams work together. Interactions model the Team Topologies
collaboration modes.

**Modes:** `x-as-a-service`, `collaboration`, `facilitating`

```yaml
interactions:
  - from: "ingestion-team"
    to: "catalog-core"
    mode: "x-as-a-service"
    via: "Entity CRUD"
    description: "Ingestion team writes entities via the entity-store API without collaboration overhead"

  - from: "ingestion-team"
    to: "enabling-team"
    mode: "facilitating"
    description: "Enabling team helps ingestion team adopt new validation framework"
```

---

### `platforms`

Optional grouping of platform teams that together provide shared capabilities.

```yaml
platforms:
  - name: "Catalog Platform Group"
    description: "Catalog infrastructure services and tooling"
    teams:
      - "catalog-core"
      - "catalog-dev"
    provides:
      - "Entity CRUD"
      - "Catalog Registry"
```

---

### `data_assets`

Shared storage and messaging infrastructure. Model data assets when multiple
services share the same database, cache, or event stream — this reveals
implicit coupling that service dependency graphs miss.

```yaml
data_assets:
  - name: "entity_store_db"
    type: "database"
    description: "Primary entity store for all catalog entities"
    usedBy:
      - target: "entity-store"
        access: "read-write"
      - target: "publisher"
        access: "read"
      - target: "backup"
        access: "read"

  - name: "entity_change_events"
    type: "event-stream"
    description: "Entity change events published on catalog mutations"
    producedBy: "entity-store"
    consumedBy:
      - "publisher"
      - "cache-worker"
```

**Types:** `database`, `cache`, `event-stream`, `blob-storage`, `search-index`

---

### `external_dependencies`

Systems outside the modeled boundary that your services call. Modeling these
reveals blast radius and cross-team coupling.

```yaml
external_dependencies:
  - name: "workflow-engine"
    description: "Workflow orchestration service"
    usedBy:
      - target: "feed-worker"
        description: "Schedules and tracks feed ingestion workflows"
      - target: "publisher-worker"
        description: "Orchestrates publishing pipelines"
```

---

## Relationship Forms — Quick Reference

All list-type relationships (`supportedBy`, `realizedBy`, `dependsOn`, `owns`,
`usedBy`, `provides`) support two forms that can be mixed freely:

```yaml
# Short form — just the name
dependsOn:
  - "registry"

# Long form — adds description and optional role
dependsOn:
  - target: "registry"
    description: "Catalog routing config on every entity operation"

# Mixed — both in the same list
realizedBy:
  - "entity-store"
  - target: "registry"
    role: "supporting"
    description: "Provides context for entity operations"
```

---

## Required vs Optional Fields

| Entity | Required | Optional |
|--------|----------|----------|
| `system` | `name`, `description` | — |
| `actor` | `name`, `description` | — |
| `need` | `name`, `actor`, `outcome`, `supportedBy` (≥1) | `description` |
| `capability` | `name`, `description`, `visibility` | `children`, `realizedBy`, `dependsOn` |
| `service` | `name`, `description`, `ownedBy` | `dependsOn` |
| `team` | `name`, `type` | `description`, `owns`, `provides` |
| `interaction` | `from`, `to`, `mode` | `via`, `description` |
| `platform` | `name` | `description`, `teams`, `provides` |
| `data_asset` | `name`, `type`, `description` | `usedBy`, `producedBy`, `consumedBy` |
| `external_dependency` | `name`, `description` | `usedBy` |

---

## Common Mistakes

**1. Putting `realizedBy` on a parent capability**

Parent capabilities are realized through their children. Only leaf
capabilities (no `children`) have `realizedBy`.

```yaml
# Wrong
capabilities:
  - name: "Catalog Management"
    visibility: "domain"
    realizedBy:         # ← ERROR: this is a parent
      - "entity-store"
    children:
      - name: "Entity CRUD"
        ...
```

**2. Declaring `supports` on a service**

Services do not declare which capabilities they support. That relationship
belongs on the capability via `realizedBy`.

```yaml
# Wrong
services:
  - name: "entity-store"
    supports:          # ← DEPRECATED: not parsed
      - "Entity CRUD"
```

**3. Missing `supportedBy` on a need**

Every need must link to at least one capability.

```yaml
# Wrong
needs:
  - name: "Place an order"
    actor: "Customer"
    outcome: "My order is confirmed"
    # ← ERROR: missing supportedBy
```

**4. Using an undefined actor in a need**

The actor name must match exactly (case-sensitive) a name in the `actors` list.

**5. Leaf capability with no `realizedBy`**

A leaf capability (no children) must have at least one service in `realizedBy`,
or it must `decomposesTo` sub-capabilities.

---

## Validation Summary

The platform validates your model on load and reports:

**Errors (model rejected):**
- Need with no `supportedBy`
- Leaf capability with no `realizedBy` and no `decomposesTo`
- Parent capability with `realizedBy`
- Service with no `ownedBy`
- Invalid interaction mode or team type value

**Warnings (model loads, issues flagged):**
- Capability owned by more than 2 teams → fragmentation risk
- Team owning more than 6 capabilities → cognitive load risk
- Service not referenced by any capability's `realizedBy` → orphan service
- Capability not referenced by any need → unlinked capability
- All team interactions use the same mode → incomplete interaction modeling
- Dependency chain depth > 4 → deep chain blast radius risk
- Missing `visibility` on a capability

---

## Full Example

See [`examples/minimal.unm.yaml`](../examples/minimal.unm.yaml) for a compact
example demonstrating all YAML format features. For a larger model, see
[`examples/nexus.unm.yaml`](../examples/nexus.unm.yaml).

For the same concepts expressed in the more concise DSL format, see
[`examples/bookshelf.unm`](../examples/bookshelf.unm).
