# UNM DSL Specification v2

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
5. **Source-of-truth direction** — Services declare what they realize and which external dependencies they use; reverse lookups are computed by the system

> **For the complete field reference** — including which fields are authored vs. derived, relationship directions, and the v2 freeze removal table — see [`META_MODEL_REFERENCE.md`](META_MODEL_REFERENCE.md).

---

## 1. Meta-Model

### Core Entities

| Entity | Description | Required Fields |
|--------|-------------|-----------------|
| `system` | Top-level boundary (product, platform, org unit) | name, description |
| `actor` | A person or external system with needs | name, description |
| `need` | What an actor is trying to achieve | name, actor, outcome |
| `capability` | What the system/org must be able to do | name, description, visibility |
| `service` | Concrete implementation realizing capabilities | name, ownedBy |
| `team` | Organizational unit owning capabilities | name, type |
| `platform` | Grouping of platform teams | name |
| `interaction` | How two teams work together | from, to, mode |
| `data_asset` | Storage/messaging infrastructure shared by services | name, type |
| `external_dependency` | System outside the modeled boundary | name, description |

### Relationship Types (v2 — source-of-truth direction)

| Relationship | Authored On | Semantics |
|-------------|-------------|-----------|
| `need.supportedBy` | Need | Need is fulfilled by this capability |
| `service.realizes` | Service | Service implements this capability (source of truth) |
| `service.externalDeps` | Service | Service depends on this external dependency (source of truth) |
| `service.dependsOn` | Service | Service calls this other service |
| `service.ownedBy` | Service | Team responsible for this service |
| `team.owns` | Team | Team owns this capability |
| `team.interacts` | Team | Team interacts with another team (primary authoring path) |
| `capability.dependsOn` | Capability | Capability requires another capability |
| `capability.parent` | Capability | Flat-form parent reference for hierarchy |
| `data_asset.usedBy` | Data Asset | Which services use this data asset |

> **Derived (not authored):** `capability.realizedBy` is computed by the system from `service.realizes` declarations. Do not author `realizedBy` in model files.

### Value Types

```
TeamType         = "stream-aligned" | "platform" | "enabling" | "complicated-subsystem"
InteractionMode  = "collaboration" | "x-as-a-service" | "facilitating"
RelationshipRole = "primary" | "supporting" | "consuming"
DataAssetType    = "database" | "cache" | "event-stream" | "blob-storage" | "search-index"
Visibility       = "user-facing" | "domain" | "foundational" | "infrastructure"
```

#### Visibility (UNM Value Chain Position)

The `visibility` field on capabilities represents position in the UNM vertical value chain. The user sits at the top; capabilities are positioned by how visible they are to the user. This enables rendering a proper UNM map where dependency arrows flow downward from visible to invisible.

| Level | Meaning | Examples |
|-------|---------|----------|
| `user-facing` | Directly experienced by end users | Order Submission, Product Search |
| `domain` | Core business processing, not directly visible | Publishing, Indexing, Validation |
| `foundational` | Internal capabilities that underpin domain logic | Entity CRUD, Registry, Backup |
| `infrastructure` | Deep infrastructure, fully invisible | Data storage, external systems |

**Visibility inheritance:** When a child capability has no `visibility` set, it inherits the parent's visibility. Set `visibility` explicitly on a child to override.

### Described Relationships

All list-type relationships support two forms that can be mixed freely:

**Short form** (just the reference):
```yaml
dependsOn:
  - "catalog-parser"
```

**Long form** (with description and optional role):
```yaml
dependsOn:
  - target: "catalog-parser"
    description: "Sends raw feed data for format-specific parsing"
```

**With role** (for `service.realizes`):
```yaml
realizes:
  - target: "Search"
    role: "primary"
    description: "Full-text search implementation"
  - target: "Catalog"
    role: "supporting"
    description: "Reads catalog data for indexing"
```

Roles distinguish how a service relates to a capability:
- **primary** — main implementation
- **supporting** — contributes to but does not own
- **consuming** — uses the capability as a client

---

## 2. DSL Syntax

### 2.1 System Declaration

```unm
system "BookShelf" {
  description "Online bookstore platform for browsing, purchasing, and managing books"
}
```

The system block also supports version metadata managed automatically by the platform:

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

**Multi-actor needs** — multiple actors can share a need:

```unm
need "View catalog data" {
  actor "Reader", "Author"
  outcome "Users can browse the full book catalog"
  supportedBy "Catalog Management"
}
```

In YAML, use either a single string or a list:

```yaml
needs:
  - name: "View catalog data"
    actor: "Reader"                  # single actor

  - name: "Manage inventory"
    actor: ["Reader", "Author"]      # multi-actor list
```

### 2.4 Capabilities (with visibility and hierarchical decomposition)

Capabilities carry a `visibility` field representing their position in the UNM vertical value chain. Services declare which capabilities they realize — capabilities do not declare their own `realizedBy`.

```unm
capability "Catalog & Inventory" {
  description "Full lifecycle management of the book catalog"
  visibility "domain"

  capability "Catalog CRUD" {
    description "Create, read, update, delete book records"
    visibility "foundational"
  }

  capability "Catalog Search" {
    description "Full-text search across the catalog"
    visibility "user-facing"
  }
}
```

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
}
```

Capabilities can also declare dependencies on other capabilities:

```unm
capability "Search & Discovery" {
  description "Help readers find books"
  visibility "user-facing"
  dependsOn "Catalog Management"
}
```

**Visibility inheritance:** Child capabilities with no `visibility` inherit from their parent. Set `visibility` on a child to override.

### 2.5 Services

Services are concrete implementations that realize capabilities. A service declares its owner, the capabilities it realizes, any external dependencies it uses, and its service-to-service dependencies.

```unm
service "search-service" {
  description "Full-text search engine for the book catalog"
  ownedBy "Discovery"
  realizes "Catalog Search" : "Indexes and queries book data"
  externalDeps "Elasticsearch"
  dependsOn "catalog-api" : "Indexes book data from the catalog"
}

service "catalog-api" {
  description "Authoritative book catalog CRUD service"
  ownedBy "Storefront"
  realizes "Catalog CRUD"
}

service "notification-service" {
  description "Multi-channel notification dispatch service"
  ownedBy "Platform"
  realizes "Notification Delivery"
  externalDeps "Email Provider"
}
```

`realizes` supports the same short/long/role forms as all other relationships:

```unm
// Simple
realizes "Search & Discovery"

// With description
realizes "Catalog Search" : "Full-text search implementation"

// With role and description (block form)
realizes "Catalog Search" {
  role "primary"
  description "Full-text search implementation"
}
```

### 2.6 Teams (with Team Topologies types)

```unm
team "Storefront" {
  type "stream-aligned"
  description "Owns the catalog browsing experience and author tools"
  size 6
  owns "Catalog CRUD"
  owns "Author Dashboard"
}

team "Platform" {
  type "platform"
  description "Provides shared infrastructure services"
  size 3
  owns "Notification Delivery"
}
```

Teams declare inline interactions using `interacts` — this is the primary authoring path for team interactions:

```unm
team "Discovery" {
  type "stream-aligned"
  description "Owns search and recommendations"
  size 4
  owns "Catalog Search"
  interacts "Storefront" mode "x-as-a-service" via "Catalog CRUD" description "Consumes catalog APIs"
}
```

In YAML, `interacts` uses an object form:

```yaml
teams:
  - name: "Discovery"
    type: "stream-aligned"
    interacts:
      - target: "Storefront"
        mode: "x-as-a-service"
        via: "Catalog CRUD"
        description: "Consumes catalog APIs"
```

### 2.7 Team Interactions (standalone form — DSL only)

In the DSL, interactions can also be declared as standalone top-level blocks using arrow syntax. This is ergonomic for expressing interactions outside of team blocks:

```unm
interaction "Discovery" -> "Storefront" {
  mode "x-as-a-service"
  via "Catalog CRUD"
  description "Discovery team consumes catalog APIs for indexing"
}

interaction "Fulfillment" -> "Platform" {
  mode "x-as-a-service"
  via "Notification Delivery"
  description "Fulfillment triggers order notifications via platform APIs"
}
```

**Modes:** `x-as-a-service`, `collaboration`, `facilitating`

> **Note:** The top-level `interactions:` section has been removed from YAML. In YAML, declare all interactions via `team.interacts`. Standalone arrow-syntax blocks remain available in the DSL format only.

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

External dependencies define systems outside the modeled boundary. They are **definition-only** (name + description). The edges that show which services use them come from `service.externalDeps` declarations.

```unm
external_dependency "Payment Gateway" {
  description "Third-party payment processing provider"
}

external_dependency "Elasticsearch" {
  description "Search engine cluster for full-text indexing"
}

external_dependency "Email Provider" {
  description "Transactional email delivery service"
}
```

Services declare their use of external dependencies:

```unm
service "order-service" {
  ownedBy "Fulfillment"
  realizes "Order Processing"
  externalDeps "Payment Gateway"
}

service "search-service" {
  ownedBy "Discovery"
  realizes "Catalog Search"
  externalDeps "Elasticsearch"
}
```

In YAML:

```yaml
external_dependencies:
  - name: "Payment Gateway"
    description: "Third-party payment processing provider"

services:
  - name: "order-service"
    ownedBy: "Fulfillment"
    realizes:
      - "Order Processing"
    externalDeps:
      - "Payment Gateway"
```

### 2.11 Transition Modeling

```unm
transition "Consolidate catalog ownership" {
  description "Move from fragmented to single-team catalog ownership"

  current {
    capability "Catalog CRUD" ownedBy team "Team A"
    capability "Catalog CRUD" ownedBy team "Team B"
  }

  target {
    capability "Catalog CRUD" ownedBy team "Storefront"
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
2. A `service` cannot have an empty `ownedBy`
3. `team.interacts` must reference a valid interaction mode
4. `capability.parent` must reference a capability that exists in the model
5. Circular `parent` references are rejected

### Warnings
1. Capability owned by more than 2 teams → fragmentation warning
2. Team owning more than 6 capabilities → cognitive load warning
3. Service not referenced by any capability's realized-by list → orphan warning
4. Circular dependencies between capabilities → cycle warning
5. Capability without `visibility` → missing visibility warning
6. Leaf capability not referenced by any `need.supportedBy` → unlinked capability warning
7. All team interactions use the same `mode` → interaction diversity warning
8. Critical dependency chain depth > 4 → deep dependency chain warning
9. Team with cognitive load score > 20 → cognitive load threshold warning
10. `service.externalDeps` references an external dependency not declared in `external_dependencies` → unresolved reference warning

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
      - "Search & Discovery"
      - target: "Order Processing"
        description: "Handles the purchase lifecycle"

capabilities:
  - name: "Search & Discovery"
    description: "Help readers find books"
    visibility: "user-facing"
    dependsOn:
      - "Catalog Management"

  - name: "Order Processing"
    description: "Handle the purchase lifecycle"
    visibility: "user-facing"

  - name: "Catalog CRUD"
    description: "Create, read, update, delete book records"
    visibility: "foundational"
    parent: "Catalog Management"

  - name: "Catalog Management"
    description: "Maintain the authoritative book catalog"
    visibility: "domain"

services:
  - name: "search-service"
    description: "Full-text search engine for the book catalog"
    ownedBy: "Discovery"
    realizes:
      - target: "Search & Discovery"
        role: "primary"
        description: "Full-text search across the catalog"
    externalDeps:
      - "Elasticsearch"
    dependsOn:
      - target: "catalog-api"
        description: "Indexes book data from the catalog"

  - name: "catalog-api"
    description: "Authoritative book catalog CRUD service"
    ownedBy: "Storefront"
    realizes:
      - "Catalog CRUD"

  - name: "order-service"
    description: "Manages cart, checkout, and payment"
    ownedBy: "Fulfillment"
    realizes:
      - "Order Processing"
    externalDeps:
      - "Payment Gateway"

teams:
  - name: "Storefront"
    type: "stream-aligned"
    owns:
      - "Catalog Management"

  - name: "Discovery"
    type: "stream-aligned"
    owns:
      - "Search & Discovery"
    interacts:
      - target: "Storefront"
        mode: "x-as-a-service"
        via: "Catalog Management"
        description: "Discovery team consumes catalog APIs for indexing"

  - name: "Fulfillment"
    type: "stream-aligned"
    owns:
      - "Order Processing"

external_dependencies:
  - name: "Elasticsearch"
    description: "Search engine cluster for full-text indexing"

  - name: "Payment Gateway"
    description: "Third-party payment processing provider"
```

**Key schema rules:**
- **Service-side authoring:** Services declare `realizes` (which capabilities they implement) and `externalDeps` (which external systems they use). Capabilities do NOT declare `realizedBy`.
- **Team-side interactions:** Teams declare `interacts` inline. There is no top-level `interactions:` section in YAML.
- **Visibility is required:** Every capability should have a `visibility` level for proper UNM value chain rendering. Children inherit from parents if not set.
- **Hierarchy:** Use `parent` (flat form) or `children` (nested form) to group capabilities. Both produce the same model.
- **Relationship forms:** Short string and long object (`target` + `description` + optional `role`) can be mixed freely in any list.
