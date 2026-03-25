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
system "Uber Delivery" {
  description "End-to-end food delivery platform"
  
  // contents: actors, capabilities, services, teams, etc.
}
```

### 2.2 Actors

```unm
actor "Merchant" {
  description "Restaurant or store owner using the platform"
}

actor "Eater" {
  description "End consumer ordering food"
}
```

### 2.3 Needs

```unm
need "Publish catalog data" {
  actor "Merchant"
  outcome "My catalog is available for customers to browse and order from"
  
  supportedBy capability "Catalog ingestion"
  supportedBy capability "Catalog validation"
  supportedBy capability "Catalog publication"
}
```

> **Note**: Earlier versions of the DSL included a `scenario` entity as an intermediary between actors and needs. This has been removed. UNM follows a direct **Actor → Need → Capability** chain. Contextual information belongs in the need's `description` field. YAML files containing a `scenarios` section will parse with a deprecation warning.

### 2.4 Capabilities (with visibility and hierarchical decomposition)

Capabilities carry a `visibility` field representing their position in the UNM vertical value chain. Parent capabilities group children and inherit visibility from their most user-facing child if not explicitly set. Only leaf capabilities have `realizedBy` service references — parent capabilities are realized through their children.

All relationships within capabilities support an optional description using `:` syntax.

```unm
capability "Catalog & Inventory" {
  description "Enable merchants to provide, maintain, and protect catalog data"
  visibility domain
  
  capability "Catalog ingestion" {
    description "Accept catalog data from various sources"
    visibility user-facing
    realizedBy service "feed-ingestion-service" : "Receives and normalizes merchant feeds"
    realizedBy service "catalog-parser" : "Handles format-specific parsing (CSV, JSON, XML)"
  }
  
  capability "Catalog validation" {
    description "Ensure catalog data meets quality standards"
    visibility domain
    realizedBy service "schema-validator" : "Validates structure and field types"
    realizedBy service "quality-gate" : "Enforces business quality thresholds"
  }
  
  capability "Catalog storage" {
    description "Persist and retrieve catalog data"
    visibility foundational
    realizedBy service "catalog-store" : "Primary entity persistence"
  }
}
```

Relationships without `:` are still valid — the description is optional.

### 2.5 Services

Services are dependencies in the value chain that realize capabilities. A service has a name, description, team owner, and service-to-service dependencies. **Services do not declare which capabilities they support** — that relationship is defined on the capability side via `realizedBy` and derived at query time. Similarly, data asset and external dependency usage is declared on those entities, not on services.

```unm
service "entity-store" {
  description "Authoritative CRUD store for all catalog entities"
  ownedBy team "catalog-core"

  dependsOn service "registry" : "Catalog context lookup on every entity operation"
}

service "feed-worker" {
  description "Workflow worker executing feed ingestion, sync, and transformation"
  ownedBy team "ingestion-team"

  dependsOn service "entity-store" : "Entity writes during feed processing"
  dependsOn service "registry" : "Catalog config lookup"
}

service "serving-gateway" {
  description "Gateway layer for serving catalog data to consumers"
  ownedBy team "serving-team"
}
```

> **Eliminated fields**: Earlier versions included `type` (service classification), `supports` (service→capability), `dataAssets` (service→data_asset), and `externalDependsOn` (service→external) on services. These are all removed:
> - `type` is an operational concern, not a UNM concept — service classification belongs in a service catalog, not the UNM model
> - `supports`, `dataAssets`, `externalDependsOn` duplicate relationships already declared on capabilities, data assets, and external dependencies respectively
>
> The parser derives reverse lookups at query time. YAML files containing these deprecated fields will parse with warnings.

### 2.6 Teams (with Team Topologies types)

```unm
team "Ingestion Team" {
  type stream-aligned
  description "Handles merchant catalog feed ingestion and processing"

  owns capability "Catalog ingestion"
  owns service "feed-ingestion-service"
  owns service "catalog-parser"
}

team "Catalog Platform" {
  type platform
  description "Provides shared catalog infrastructure"

  provides capability "Catalog storage"
  provides capability "Backup and restore"
}
```

### 2.7 Team Interactions

```unm
interaction {
  from team "Ingestion Team"
  to team "Catalog Platform"
  mode x-as-a-service
  via capability "Catalog storage"
  description "Ingestion team consumes storage APIs without collaboration overhead"
}

interaction {
  from team "Ingestion Team"
  to team "Merchant Experience"
  mode collaboration
  via capability "Catalog publication"
  description "Close collaboration during publication workflow redesign"
}
```

### 2.8 Platform Groupings

```unm
platform "Catalog Platform Group" {
  description "Catalog and inventory infrastructure platform"

  includes team "Catalog Platform"
  includes team "Storage Platform"
  includes team "Observability Platform"

  provides capability "Catalog storage"
  provides capability "Observability"
  provides capability "Backup and restore"
}
```

### 2.9 Data Assets

Data assets model the storage and messaging infrastructure that services share. Shared data assets reveal implicit coupling that service dependency graphs miss.

```unm
data_asset "entity_store" {
  type database
  description "Primary entity store for all catalog entities"
  usedBy service "entity-store" role "read-write"
  usedBy service "publisher" role "read"
  usedBy service "backup" role "read"
}

data_asset "entity_change_events" {
  type event-stream
  description "Entity change events published on catalog mutations"
  producedBy service "entity-store"
  consumedBy service "event-consumer"
  consumedBy service "publisher"
  consumedBy service "cache-worker"
}

data_asset "config_cache" {
  type cache
  description "Cache for catalog configuration and routing metadata"
  usedBy service "registry" role "read-write"
}
```

Data asset types: `database`, `cache`, `event-stream`, `blob-storage`, `search-index`

### 2.10 External Dependencies

External dependencies model systems outside the modeled boundary. They show blast radius and cross-team coupling.

```unm
external "store-info-service" {
  description "Store and location information service outside system boundary"
  usedBy service "entity-store" : "Store info for entity context"
  usedBy service "registry" : "Store metadata for catalog routing"
  usedBy service "feed-worker" : "Store details during feed processing"
}

external "workflow-engine" {
  description "Workflow orchestration engine"
  usedBy service "publisher-worker" : "Orchestrates publishing workflows"
  usedBy service "feed-worker" : "Feed ingestion and sync workflows"
  usedBy service "backup" : "Backup and restore orchestration"
}
```

### 2.11 Transition Modeling

```unm
transition "Consolidate catalog ownership" {
  description "Move from fragmented to stream-aligned catalog ownership"
  
  current {
    capability "Catalog publication" ownedBy team "Team A"
    capability "Catalog publication" ownedBy team "Team B"
    capability "Catalog publication" ownedBy team "Team C"
  }
  
  target {
    capability "Catalog publication" ownedBy team "Catalog Stream"
  }
  
  step 1 "Align Team A and Team B" {
    action merge team "Team A" team "Team B" into team "Catalog Stream"
    expected_outcome "Single team owns ingestion and validation"
  }
  
  step 2 "Absorb Team C scope" {
    action move capability "Catalog publication" from team "Team C" to team "Catalog Stream"
    expected_outcome "Full publication pipeline under one team"
  }
  
  step 3 "Extract platform capabilities" {
    action extract capability "Catalog storage" to team "Catalog Platform"
    expected_outcome "Storage becomes x-as-a-service"
  }
}
```

### 2.12 Imports and Composition

```unm
import "catalog.unm" as catalog
import "merchant-experience.unm" as mx

system "Delivery Platform" {
  includes catalog."Catalog Platform Group"
  includes mx."Merchant Experience"

  // cross-system relationships
  dependency {
    from mx."Menu Display"
    to catalog."Catalog publication"
    description "Menu display depends on published catalog data"
  }
}
```

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

Models are expressed in YAML. This is the primary authoring format for Phase 1 and remains supported alongside the custom DSL.

```yaml
system:
  name: "Catalog Platform"
  description: "Catalog and inventory management platform"

actors:
  - name: "Merchant"
    description: "Business owner managing product catalog"

needs:
  - name: "Publish catalog data"
    actor: "Merchant"
    outcome: "My catalog is available for customers"
    supportedBy:
      - "Catalog ingestion"                              # short form
      - target: "Catalog validation"                     # long form
        description: "Validates schema before publishing"

capabilities:
  - name: "Catalog Management"
    description: "Enable merchants to manage catalog data"
    visibility: "domain"
    children:
      - name: "Catalog Entity CRUD"
        description: "Foundational CRUD on all entity types"
        visibility: "foundational"
        realizedBy:
          - target: "entity-store"
            role: "primary"
            description: "Primary entity service and internal CRUD API"
          - target: "registry"
            role: "supporting"
            description: "Catalog context for entity operations"
          - target: "publisher"
            role: "consuming"
            description: "Reads entities during publishing"
      - name: "Entity Editing API"
        description: "High-level editing operations with business rules"
        visibility: "domain"
        realizedBy:
          - target: "entity-store"
            role: "primary"
            description: "EntityEditingService and validation pipeline"
  - name: "Consumer Serving"
    description: "Serve catalog data to end consumers"
    visibility: "user-facing"
    children:
      - name: "Menu Item Serving"
        description: "Fast-read endpoint for menu items"
        visibility: "user-facing"
        realizedBy:
          - target: "serving"
            role: "primary"
            description: "ServeItems RPC handler"

# Services declare owner and service-to-service dependencies only.
# Capability support is declared on capabilities via realizedBy (source of truth).
# Data asset and external dependency usage is declared on those entities.
services:
  - name: "entity-store"
    description: "Authoritative CRUD store for all catalog entities"
    ownedBy: "catalog-core"
    dependsOn:
      - target: "registry"
        description: "Catalog context lookup on every entity operation"

  - name: "serving"
    description: "Consumer-facing fast-read endpoint for menu items"
    ownedBy: "serving-team"
    dependsOn:
      - target: "serving-ingestor"
        description: "Item records"

teams:
  - name: "catalog-core"
    type: "platform"
    owns:
      - "Catalog Entity CRUD"
      - "Entity Editing API"

platforms:
  - name: "Catalog Platform Group"
    description: "Catalog and inventory infrastructure platform"
    teams:
      - "catalog-core"
      - "catalog-dev"
    provides:
      - "Catalog Entity CRUD"
      - "Catalog Registry & Configuration"

data_assets:
  - name: "entity_store"
    type: "database"
    description: "Primary entity store"
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
      - "event-consumer"
      - "publisher"
      - "cache-worker"

external_dependencies:
  - name: "store-info-service"
    description: "Store information service outside system boundary"
    usedBy:
      - target: "entity-store"
        description: "Store info for entity context"
      - target: "registry"
        description: "Store metadata for catalog routing"

  - name: "workflow-engine"
    description: "Workflow orchestration engine"
    usedBy:
      - target: "publisher-worker"
        description: "Orchestrates publishing workflows"
      - target: "feed-worker"
        description: "Feed ingestion and sync workflows"

interactions:
  - from: "ingestion-team"
    to: "catalog-core"
    mode: "x-as-a-service"
    via: "Catalog Entity CRUD"
    description: "Ingestion team writes entities via entity-store's editing service"
```

**Key schema rules**:
- **Unidirectional relationships**: Capabilities declare `realizedBy` services (source of truth). Services do NOT declare `supports`. Data assets declare `usedBy` services. External dependencies declare `usedBy` services. Reverse lookups are derived at query time.
- **Visibility is required**: Every capability should have a `visibility` level for proper UNM value chain rendering.
- **Hierarchy**: Parent capabilities group children. Only leaf capabilities have `realizedBy`.
- **Relationship forms**: short string and long object (`target` + `description` + optional `role`) can be mixed freely in any list. The parser treats plain strings as short form.
