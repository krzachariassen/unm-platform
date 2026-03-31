# UNM DSL Specification v0.1

## Overview

The UNM DSL (User Needs Mapping Domain-Specific Language) is a textual modeling language for expressing user needs, capabilities, services, teams, and their relationships. It is designed to be:

- **Human-readable**: Approachable for engineers and architects
- **Machine-parseable**: Enables automated validation, analysis, and visualization
- **Versionable**: Text-based, suitable for git workflows
- **Hierarchical**: Supports drill-down decomposition like C4's zoom levels

## Design Principles

1. **Model first** — The DSL defines a semantic model, not a visual layout
2. **Explicit relationships** — All connections are declared, not implied
3. **Layered abstraction** — Support for L1 (enterprise) through L5 (team/transition) views
4. **AI-assisted** — The platform analyzes the data layer and generates findings; the DSL file is the authoritative data source
5. **Composable** — Models can import and reference other models

---

## 1. Meta-Model

### Core Entities

| Entity | Description | Required Fields |
|--------|-------------|-----------------|
| `system` | Top-level boundary (product, platform, org unit) | name, description |
| `actor` | A person or external system with needs | name, description |
| `need` | What an actor is trying to achieve (includes `outcome` as a field) | name, actor, outcome |
| `capability` | What the system/org must be able to do | name, description, visibility |
| `service` | Concrete implementation realizing capabilities | name, description |
| `team` | Organizational unit owning services/capabilities | name, type |
| `platform` | Grouping of platform teams | name, teams |
| `interaction` | How two teams work together | from, to, mode |
| `data_asset` | Storage/messaging infrastructure shared by services | name, type, usedBy |
| `external_dependency` | System outside the modeled boundary | name, description |

### Relationship Types

Every relationship in the model can carry a **description** — a human-readable label that explains *what* the relationship means, not just *that* it exists. Descriptions are rendered as edge labels in visualizations.

| Relationship | From | To | Semantics | Description Example |
|-------------|------|-----|-----------|---------------------|
| `hasNeed` | Actor | Need | Actor has this need | — |
| `supportedBy` | Need | Capability | Need is fulfilled by capability | "Ingestion enables initial catalog upload" |
| `realizedBy` | Capability | Service | Capability is implemented by service | "Handles feed parsing and normalization" |
| `ownedBy` | Service/Capability | Team | Team is responsible for this | "Primary maintainer since Q2 2024" |
| `dependsOn` | Capability/Service | Capability/Service | Requires this to function | "Calls validation API before persisting" |
| `decomposesTo` | Capability | Capability[] | Breaks down into sub-capabilities | — |
| `provides` | Platform/Team | Capability | Makes this capability available | "Exposes as self-service API" |
| `interactsWith` | Team | Team | Teams have interaction | "Weekly sync on schema changes" |

### Described Relationship

All relationships support two forms:

**Short form** (just the reference):
```yaml
dependsOn:
  - "catalog-parser"
```

**Long form** (with description, optional role, and metadata):
```yaml
dependsOn:
  - target: "catalog-parser"
    description: "Sends raw feed data for format-specific parsing"
  - target: "schema-validator"
    description: "Validates parsed output against merchant schema rules"
```

**With role** (for service-to-capability and team-to-capability):
```yaml
realizedBy:
  - target: "core"
    role: "primary"
    description: "Authoritative CRUD store for all entity types"
  - target: "registry"
    role: "supporting"
    description: "Provides catalog context for entity operations"
  - target: "publisher"
    role: "consuming"
    description: "Reads entities during publishing pipeline"
```

Roles distinguish how a service or team relates to a capability:
- **primary** — this is the main implementation or owner
- **supporting** — contributes to but does not own the capability
- **consuming** — uses the capability as a client

Both forms can be mixed in the same list. The parser treats a plain string as short form and an object as long form.

This applies to all list-type relationships: `dependsOn`, `realizedBy`, `supportedBy`, `owns`, `provides`, and `usedBy`.

### Value Types

```
TeamType         = "stream-aligned" | "platform" | "enabling" | "complicated-subsystem"
InteractionMode  = "collaboration" | "x-as-a-service" | "facilitating"
RelationshipRole = "primary" | "supporting" | "consuming"
DataAssetType    = "database" | "cache" | "event-stream" | "blob-storage" | "search-index"
Visibility       = "user-facing" | "domain" | "foundational" | "infrastructure"
```

#### Visibility (UNM Value Chain Position)

The `visibility` field on capabilities represents position in the UNM vertical value chain — the defining characteristic of a User Needs Map. The user sits at the top; capabilities are positioned by how visible they are to the user. This enables rendering a proper UNM map where dependency arrows flow downward from visible to invisible.

| Level | Meaning | Examples |
|-------|---------|----------|
| `user-facing` | Directly experienced by end users | Order Submission, Product Search (API surface) |
| `domain` | Core business processing, not directly visible | Publishing, Indexing, Validation, Bulk Actions |
| `foundational` | Internal capabilities that underpin domain logic | Entity CRUD, Registry, Backup & Restore |
| `infrastructure` | Deep infrastructure, fully invisible | Data storage, external systems, shared packages |

---

## 2. DSL Syntax

### 2.1 System Declaration

```unm
system "BookShelf" {
  description "Online bookstore platform for browsing, purchasing, and managing books"
}
```

The system block also supports version metadata fields, managed automatically
by the platform when models are edited through the API:

```unm
system "BookShelf" {
  description "Online bookstore platform"
  version "5"
  lastModified "2026-03-30T15:37:34Z"
  author "kristian"
}
```

### 2.2 Actors

```unm
actor "Reader" {
  description "End user browsing and purchasing books"
}

actor "Author" {
  description "Book author managing listings and tracking sales"
}
```

### 2.3 Needs

```unm
need "Find and purchase books easily" {
  actor "Reader"
  outcome "I can search for books and complete a purchase in under a minute"
  supportedBy "Search & Discovery"
  supportedBy "Order Processing"
}
```

Multiple actors can share a need using comma-separated values:

```unm
need "View catalog data" {
  actor "Reader", "Author"
  outcome "Users can browse the full book catalog"
  supportedBy "Catalog Management"
}
```

> **Note**: Earlier versions of the DSL included a `scenario` entity as an intermediary between actors and needs. This has been removed. UNM follows a direct **Actor → Need → Capability** chain. Contextual information belongs in the need's `description` field. YAML files containing a `scenarios` section will parse with a deprecation warning.

### 2.4 Capabilities (with visibility and hierarchical decomposition)

Capabilities carry a `visibility` field representing their position in the UNM vertical value chain. Parent capabilities group children and inherit visibility from their most user-facing child if not explicitly set. Only leaf capabilities have `realizedBy` service references — parent capabilities are realized through their children.

All relationships within capabilities support an optional description using `:` syntax.

```unm
capability "Catalog & Inventory" {
  description "Full lifecycle management of the book catalog"
  visibility "domain"

  capability "Catalog CRUD" {
    description "Create, read, update, delete book records"
    visibility "foundational"
    realizedBy "catalog-api" : "Primary entity persistence"
  }

  capability "Catalog Search" {
    description "Full-text search across the catalog"
    visibility "user-facing"
    realizedBy "search-service" : "Indexes and queries book data"
  }
}
```

Relationships without `:` are still valid — the description is optional.

Alternatively, use flat `parent` references instead of nesting:

```unm
capability "Catalog & Inventory" {
  description "Full lifecycle management of the book catalog"
  visibility "domain"
}

capability "Catalog CRUD" {
  description "Create, read, update, delete book records"
  visibility "foundational"
  parent "Catalog & Inventory"
  realizedBy "catalog-api"
}
```

### 2.5 Services

Services are concrete implementations that realize capabilities. A service has a name, description, team owner, and service-to-service dependencies.

```unm
service "catalog-api" {
  description "Authoritative book catalog CRUD service"
  ownedBy "Storefront"
  dependsOn "order-service" : "Validates stock availability on order placement"
}

service "search-service" {
  description "Full-text search engine for the book catalog"
  ownedBy "Discovery"
  dependsOn "catalog-api" : "Indexes book data from the catalog"
}

service "notification-service" {
  description "Multi-channel notification dispatch service"
  ownedBy "Platform"
}
```

> **Note**: Earlier versions included `type` (service classification) and `supports` (service→capability) fields on services. These are removed. The parser derives reverse lookups at query time. YAML files containing these deprecated fields will parse with warnings.

### 2.6 Teams (with Team Topologies types)

```unm
team "Storefront" {
  type "stream-aligned"
  description "Owns the catalog browsing experience and author tools"
  size 6
  owns "Catalog Management"
  owns "Author Dashboard"
}

team "Platform" {
  type "platform"
  description "Provides shared infrastructure services"
  size 3
  owns "Notification Delivery"
}
```

Teams can declare inline interactions:

```unm
team "Discovery" {
  type "stream-aligned"
  description "Owns search and recommendations"
  size 4
  owns "Search & Discovery"
  interacts "Storefront" mode "x-as-a-service" via "Catalog Management" description "Consumes catalog APIs"
}
```

### 2.7 Team Interactions

Interactions use arrow syntax to connect two teams:

```unm
interaction "Discovery" -> "Storefront" {
  mode "x-as-a-service"
  via "Catalog Management"
  description "Discovery team consumes catalog APIs for indexing"
}

interaction "Fulfillment" -> "Platform" {
  mode "x-as-a-service"
  via "Notification Delivery"
  description "Fulfillment triggers order notifications via platform APIs"
}
```

**Modes:** `x-as-a-service`, `collaboration`, `facilitating`

### 2.8 Platform Groupings

```unm
platform "Infrastructure Platform" {
  description "Shared infrastructure services and tooling"
  teams ["Platform", "SRE"]
}
```

### 2.9 Data Assets

Data assets model the storage and messaging infrastructure that services share. Shared data assets reveal implicit coupling that service dependency graphs miss.

```unm
data_asset "books_db" {
  type "database"
  description "Primary relational database storing the book catalog"
  usedBy "catalog-api"
  usedBy "search-service"
}

data_asset "book_change_events" {
  type "event-stream"
  description "Events emitted when book records are created or updated"
  usedBy "search-service"
  usedBy "recommendation-engine"
}
```

Data asset types: `database`, `cache`, `event-stream`, `blob-storage`, `search-index`

### 2.10 External Dependencies

External dependencies model systems outside the modeled boundary. They show blast radius and cross-team coupling.

```unm
external_dependency "Payment Gateway" {
  description "Third-party payment processing provider"
  usedBy "order-service" : "Processes credit card and wallet payments"
}

external_dependency "Email Provider" {
  description "Transactional email delivery service"
  usedBy "notification-service" : "Sends order confirmations and author notifications"
}
```

The `usedBy` field supports an optional description using `:` syntax.

### 2.11 Transition Modeling

```unm
transition "Consolidate catalog ownership" {
  description "Move from fragmented to single-team catalog ownership"

  current {
    capability "Catalog Management" ownedBy team "Team A"
    capability "Catalog Management" ownedBy team "Team B"
  }

  target {
    capability "Catalog Management" ownedBy team "Storefront"
  }

  step 1 "Merge teams" {
    action merge team "Team A" team "Team B" into team "Storefront"
    expected_outcome "Single team owns all catalog operations"
  }

  step 2 "Extract platform capabilities" {
    action extract capability "Data Persistence" to team "Platform"
    expected_outcome "Storage becomes x-as-a-service"
  }
}
```

### 2.12 Imports and Composition

```unm
import "catalog.unm"
import authors from "authors.unm"
```

Simple imports include all entities from the referenced file. Named imports
assign an alias for qualified references.

---

## 3. Abstraction Levels

Like C4, the model supports multiple zoom levels:

| Level | Name | Contains | Purpose |
|-------|------|----------|---------|
| L1 | Enterprise | Systems, high-level capabilities | Executive/strategy view |
| L2 | System | Capabilities, actors, needs | Product/architecture view |
| L3 | Capability | Sub-capabilities, services | Engineering view |
| L4 | Realization | Services, APIs, dependencies | Implementation view |
| L5 | Organization | Teams, interactions, platforms | Org design view |

A capability at L2 can expand into a full L3 model. A system at L1 can expand into a full L2 model.

---

## 4. View Projections

The same model generates multiple views:

### UNM Value Chain View (Primary)
- Shows: actor (top) → needs → capabilities ordered by visibility layer → services (bottom)
- Purpose: The canonical UNM map — the vertical value chain anchored by the user
- Layout: User at top. Capabilities positioned vertically by `visibility` level. Dependencies flow downward. Team Topologies shapes overlaid.
- Filters by: actor, visibility layer

### Need View
- Shows: actor → need → capability
- Purpose: Strategy and product conversations
- Filters by: actor

### Capability View  
- Shows: capability hierarchy + dependencies
- Purpose: Architecture and org design
- Filters by: system, level

### Realization View
- Shows: capability → services → teams
- Purpose: Engineering reality mapping
- Filters by: capability, team

### Ownership View
- Shows: teams → capabilities, highlighting fragmentation
- Purpose: Org design and execution friction analysis
- Filters by: team type, interaction mode

### Transition View
- Shows: current state vs target state, delta
- Purpose: Transformation planning
- Highlights: fragmented, duplicated, missing, consolidation candidates

### Team Topologies View
- Shows: team types, interaction modes, platform boundaries
- Purpose: Organizational optimization
- Highlights: cognitive load, interaction overhead

### Cognitive Load View
- Shows: per-team capability count, dependency count, interaction count
- Purpose: Identifying overloaded teams
- Thresholds: configurable warning/critical levels

---

## 5. Validation Rules

The parser enforces structural rules:

### Mandatory
1. Every `need` must reference at least one `capability` via `supportedBy`
2. Every leaf `capability` must be `realizedBy` at least one `service` OR `decomposesTo` sub-capabilities
3. A parent capability must NOT have `realizedBy` — only leaf capabilities have services
4. A `service` cannot `ownedBy` zero teams
5. A `service` cannot own a `team` (reversed relationship)
6. `interaction` must reference valid teams and a valid mode

### Warnings
1. Capability owned by more than 2 teams → fragmentation warning
2. Team owning more than 6 capabilities → cognitive load warning
3. Service not referenced by any capability's `realizedBy` → orphan warning
4. Circular dependencies between capabilities → cycle warning
5. Capability without `visibility` → missing visibility warning
6. `scenarios` section present in YAML → deprecation warning (ignored, not parsed)
7. `supports`, `dataAssets`, or `externalDependsOn` on services → deprecation warning (ignored, derived from source-of-truth entities)
8. `signals`, `pain_points`, or `inferred` sections present → deprecation warning (ignored — these are platform-computed outputs, not user-authored data)
9. Leaf capability not referenced by any `need.supportedBy` → unlinked capability warning (capability exists but no user need drives it — may indicate internal plumbing that should be `infrastructure` visibility, or a missing need)
10. All team interactions use the same `mode` → interaction diversity warning (real Team Topologies models typically mix x-as-a-service, collaboration, and facilitating; uniform mode may indicate the model is incomplete or the teams have not been classified carefully)
11. Critical dependency chain depth > 4 → deep dependency chain warning (long chains increase blast radius and latency; report the full chain path)
12. Team with cognitive load score > 20 (capabilities × 2 + services × 1 + outbound deps × 1 + interactions × 2) → cognitive load threshold warning

### Configurable
- Max capabilities per team (default: 6)
- Max teams per capability (default: 2)
- Max dependency chain depth before warning (default: 4)
- Cognitive load score threshold (default: 20)
- Required fields per entity type

---

## 6. File Format

- Extension: `.unm`
- Encoding: UTF-8
- One file per system or composable module
- Import paths relative to project root

---

## 7. YAML Format

Models can also be expressed in YAML (`.unm.yaml`). Both formats produce the same internal model. See [YAML_GUIDE.md](YAML_GUIDE.md) for a full tutorial.

```yaml
system:
  name: "BookShelf"
  description: "Online bookstore platform"

actors:
  - name: "Reader"
    description: "End user browsing and purchasing books"

needs:
  - name: "Find and purchase books easily"
    actor: "Reader"
    outcome: "I can search for books and complete a purchase quickly"
    supportedBy:
      - "Search & Discovery"                             # short form
      - target: "Order Processing"                       # long form
        description: "Handles the purchase lifecycle"

capabilities:
  - name: "Search & Discovery"
    description: "Help readers find books"
    visibility: "user-facing"
    realizedBy:
      - target: "search-service"
        role: "primary"
        description: "Full-text search across the catalog"
    dependsOn:
      - "Catalog Management"

  - name: "Order Processing"
    description: "Handle the purchase lifecycle"
    visibility: "user-facing"
    realizedBy:
      - "order-service"

  - name: "Catalog Management"
    description: "Maintain the authoritative book catalog"
    visibility: "domain"
    realizedBy:
      - "catalog-api"

services:
  - name: "catalog-api"
    description: "Authoritative book catalog CRUD service"
    ownedBy: "Storefront"

  - name: "search-service"
    description: "Full-text search engine for the book catalog"
    ownedBy: "Discovery"
    dependsOn:
      - target: "catalog-api"
        description: "Indexes book data from the catalog"

  - name: "order-service"
    description: "Manages cart, checkout, and payment"
    ownedBy: "Fulfillment"

teams:
  - name: "Storefront"
    type: "stream-aligned"
    owns:
      - "Catalog Management"

  - name: "Discovery"
    type: "stream-aligned"
    owns:
      - "Search & Discovery"

  - name: "Fulfillment"
    type: "stream-aligned"
    owns:
      - "Order Processing"

interactions:
  - from: "Discovery"
    to: "Storefront"
    mode: "x-as-a-service"
    via: "Catalog Management"
    description: "Discovery team consumes catalog APIs for indexing"

external_dependencies:
  - name: "Payment Gateway"
    description: "Third-party payment processing provider"
    usedBy:
      - target: "order-service"
        description: "Processes credit card and wallet payments"
```

**Key schema rules**:
- **Unidirectional relationships**: Capabilities declare `realizedBy` services (source of truth). Services do NOT declare `supports`. Data assets declare `usedBy` services. External dependencies declare `usedBy` services. Reverse lookups are derived at query time.
- **Visibility is required**: Every capability should have a `visibility` level for proper UNM value chain rendering.
- **Hierarchy**: Parent capabilities group children via `children` field. Only leaf capabilities have `realizedBy`.
- **Relationship forms**: short string and long object (`target` + `description` + optional `role`) can be mixed freely in any list.
