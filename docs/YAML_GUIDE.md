# How to Write a UNM Model in YAML

This guide walks you through writing a `.unm.yaml` model file from scratch.
It is practical and tutorial-style. For the full reference spec, see
[UNM_DSL_SPECIFICATION.md](UNM_DSL_SPECIFICATION.md). For the concise DSL
format, see [DSL_GUIDE.md](DSL_GUIDE.md).

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

A valid model requires at least one actor, one need, one capability, one service,
and one team:

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

services:
  - name: "order-api"
    description: "Public-facing order intake service"
    ownedBy: "order-team"
    realizes:
      - "Order Submission"

teams:
  - name: "order-team"
    type: "stream-aligned"
    description: "Owns end-to-end order placement and confirmation"
```

That's a complete model. The sections below explain each section.

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

People or external systems that have needs. Each actor should represent a
meaningfully different user type with different goals.

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

What each actor is trying to accomplish. A need has an `actor`, a `name`,
and an `outcome` — a first-person statement of what success looks like.

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

**Multi-actor needs** — when multiple actors share a need, use a list:

```yaml
needs:
  - name: "View catalog data"
    actor: ["Merchant", "Eater"]
    outcome: "Users can browse the full catalog"
    supportedBy:
      - "Catalog Management"
```

**Rules:**
- An actor must exist in the `actors` list.
- Every need must reference at least one capability via `supportedBy`.

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
```

Services declare which capabilities they realize — see the `services` section.

#### Hierarchical capabilities — flat form (recommended)

Use `parent` to group related capabilities. This form produces the cleanest YAML:

```yaml
capabilities:
  - name: "Catalog Management"
    description: "Full lifecycle management of catalog entities"
    visibility: "domain"

  - name: "Entity CRUD"
    description: "Create, read, update, delete on catalog entities"
    visibility: "foundational"
    parent: "Catalog Management"

  - name: "Bulk Editing"
    description: "High-throughput edits across many entities"
    visibility: "domain"
    parent: "Catalog Management"
```

#### Hierarchical capabilities — nested form (alternative)

You can also nest capabilities inside their parent using `children`:

```yaml
capabilities:
  - name: "Catalog Management"
    description: "Full lifecycle management of catalog entities"
    visibility: "domain"
    children:
      - name: "Entity CRUD"
        description: "Create, read, update, delete on catalog entities"
        visibility: "foundational"
      - name: "Bulk Editing"
        description: "High-throughput edits across many entities"
        visibility: "domain"
```

Both forms produce the same model. Use whichever reads better for your structure.

**Visibility inheritance:** Child capabilities with no `visibility` inherit from
their parent. Set `visibility` explicitly on a child to override.

#### Capability dependencies

Capabilities can depend on other capabilities:

```yaml
capabilities:
  - name: "Search & Discovery"
    description: "Help users find products"
    visibility: "user-facing"
    dependsOn:
      - "Catalog Management"
      - target: "Data Persistence"
        description: "Needs storage for search indexes"
```

---

### `services`

Concrete implementations. A service declares:
- `ownedBy` — which team owns it
- `realizes` — which capabilities it implements (source of truth)
- `externalDeps` — which external systems it depends on
- `dependsOn` — which other services it depends on

```yaml
services:
  - name: "search-service"
    description: "Full-text search engine for the catalog"
    ownedBy: "search-team"
    realizes:
      - target: "Search & Discovery"
        role: "primary"
        description: "Full-text search across title and description"
    externalDeps:
      - "Elasticsearch"
    dependsOn:
      - target: "catalog-api"
        description: "Indexes product data from the catalog"

  - name: "catalog-api"
    description: "Authoritative catalog CRUD service"
    ownedBy: "catalog-core"
    realizes:
      - "Entity CRUD"
      - target: "Bulk Editing"
        role: "supporting"
        description: "Handles bulk updates via batch API"
    dependsOn:
      - "registry"              # short form also valid

  - name: "feed-worker"
    description: "Processes merchant feed submissions"
    ownedBy: "ingestion-team"
    realizes:
      - "Feed Ingestion"
    externalDeps:
      - "workflow-engine"
```

#### `realizes` — short form vs long form

Both forms can be mixed freely:

```yaml
realizes:
  - "Entity CRUD"                      # short form — just the capability name
  - target: "Bulk Editing"             # long form — adds description and role
    role: "supporting"
    description: "Handles batch updates"
```

**Roles:**
- `primary` — main implementation
- `supporting` — contributes to but does not own
- `consuming` — uses the capability as a client

---

### `teams`

Organizational units that own capabilities. Use Team Topologies types to
classify each team.

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

#### Inline interactions

Teams declare how they interact with other teams using `interacts`. This is the
primary authoring path for team interactions in YAML:

```yaml
teams:
  - name: "ingestion-team"
    type: "stream-aligned"
    owns:
      - "Feed Ingestion"
    interacts:
      - target: "catalog-core"
        mode: "x-as-a-service"
        via: "Entity CRUD"
        description: "Ingestion team writes entities through the catalog-core API"

      - target: "enabling-team"
        mode: "facilitating"
        description: "Enabling team helps ingestion team adopt validation framework"
```

**Interaction modes:** `x-as-a-service`, `collaboration`, `facilitating`

> **Note:** There is no top-level `interactions:` section in YAML. All interactions
> are declared on teams via `interacts`.

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
      - "catalog-api"
      - "publisher"
      - "backup"

  - name: "entity_change_events"
    type: "event-stream"
    description: "Entity change events published on catalog mutations"
    usedBy:
      - "publisher"
      - "cache-worker"
```

**Types:** `database`, `cache`, `event-stream`, `blob-storage`, `search-index`

---

### `external_dependencies`

Systems outside the modeled boundary. These are **definition-only** — you
declare their name and description here. The edges showing which services use
them come from `service.externalDeps` declarations.

```yaml
external_dependencies:
  - name: "Elasticsearch"
    description: "Search engine cluster for full-text indexing"

  - name: "workflow-engine"
    description: "Workflow orchestration service for async pipelines"

  - name: "Payment Gateway"
    description: "Third-party payment processing provider"
```

Services declare their use of these systems:

```yaml
services:
  - name: "search-service"
    ownedBy: "search-team"
    realizes:
      - "Search & Discovery"
    externalDeps:
      - "Elasticsearch"

  - name: "feed-worker"
    ownedBy: "ingestion-team"
    realizes:
      - "Feed Ingestion"
    externalDeps:
      - "workflow-engine"
```

---

## Relationship Forms — Quick Reference

All list-type relationships (`supportedBy`, `realizes`, `dependsOn`, `owns`)
support two forms that can be mixed freely:

```yaml
# Short form — just the name
dependsOn:
  - "registry"

# Long form — adds description and optional role
dependsOn:
  - target: "registry"
    description: "Catalog routing config on every entity operation"

# Mixed — both in the same list
realizes:
  - "Entity CRUD"
  - target: "Bulk Editing"
    role: "supporting"
    description: "Handles batch updates"
```

---

## Required vs Optional Fields

| Entity | Required | Optional |
|--------|----------|----------|
| `system` | `name`, `description` | — |
| `actor` | `name`, `description` | — |
| `need` | `name`, `actor`, `outcome`, `supportedBy` (≥1) | — |
| `capability` | `name`, `description`, `visibility` | `parent`, `children`, `dependsOn` |
| `service` | `name`, `ownedBy` | `description`, `realizes`, `externalDeps`, `dependsOn` |
| `team` | `name`, `type` | `description`, `size`, `owns`, `interacts` |
| `platform` | `name` | `description`, `teams` |
| `data_asset` | `name`, `type`, `description` | `usedBy` |
| `external_dependency` | `name`, `description` | — |

---

## Common Mistakes

**1. Putting `realizedBy` on a capability**

Services declare which capabilities they realize — not the other way around.

```yaml
# Wrong — realizedBy is not a valid capability field
capabilities:
  - name: "Search & Discovery"
    visibility: "user-facing"
    realizedBy:          # ERROR: removed in v2
      - "search-service"

# Correct — service declares realizes
services:
  - name: "search-service"
    ownedBy: "search-team"
    realizes:
      - "Search & Discovery"
```

**2. Using top-level `interactions:` section**

Top-level `interactions:` has been removed. Declare interactions on teams.

```yaml
# Wrong — top-level interactions: is not parsed
interactions:
  - from: "ingestion-team"
    to: "catalog-core"
    mode: "x-as-a-service"

# Correct — declare on the team
teams:
  - name: "ingestion-team"
    interacts:
      - target: "catalog-core"
        mode: "x-as-a-service"
```

**3. Using `usedBy` on `external_dependencies`**

External dependency definitions do not carry `usedBy`. Services declare which
external systems they use.

```yaml
# Wrong — usedBy is not a valid external_dependency field
external_dependencies:
  - name: "Elasticsearch"
    description: "Search engine"
    usedBy:              # ERROR: removed in v2
      - "search-service"

# Correct — declare on the service
services:
  - name: "search-service"
    ownedBy: "search-team"
    externalDeps:
      - "Elasticsearch"
```

**4. Missing `supportedBy` on a need**

Every need must link to at least one capability.

```yaml
# Wrong
needs:
  - name: "Place an order"
    actor: "Customer"
    outcome: "My order is confirmed"
    # ERROR: missing supportedBy

# Correct
needs:
  - name: "Place an order"
    actor: "Customer"
    outcome: "My order is confirmed"
    supportedBy:
      - "Order Submission"
```

**5. Referencing an undefined actor**

The actor name in a need must match exactly (case-sensitive) a name in the
`actors` list.

---

## Validation Summary

The platform validates your model on load and reports:

**Errors (model rejected):**
- Need with no `supportedBy`
- Service with no `ownedBy`
- Invalid interaction mode or team type value
- Capability `parent` references a non-existent capability
- Circular `parent` references

**Warnings (model loads, issues flagged):**
- Capability owned by more than 2 teams → fragmentation risk
- Team owning more than 6 capabilities → cognitive load risk
- Service not referenced by any realized capability → orphan service
- Capability not referenced by any need → unlinked capability
- All team interactions use the same mode → incomplete interaction modeling
- Dependency chain depth > 4 → deep chain blast radius risk
- Missing `visibility` on a capability
- `service.externalDeps` references an undeclared external dependency

---

## Full Example

See [`examples/nexus.unm.yaml`](../examples/nexus.unm.yaml) for a large
real-world model demonstrating all YAML format features.

For the same concepts in the more concise DSL format, see
[`examples/bookshelf.unm`](../examples/bookshelf.unm).
