# UNM Meta-Model Reference

This document is the authoritative field reference for the UNM v2 meta-model.
It lists every entity, every field, whether the field is **authored** (written
by the user in a model file) or **derived** (computed by the system at parse or
query time), and the canonical relationship direction.

For authoring syntax, see:
- [YAML_GUIDE.md](YAML_GUIDE.md) — YAML format tutorial
- [DSL_GUIDE.md](DSL_GUIDE.md) — DSL format tutorial
- [UNM_DSL_SPECIFICATION.md](UNM_DSL_SPECIFICATION.md) — Full spec

---

## Actor

Represents a person or external system that has needs.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `description` | string | Authored | Required. |

---

## Need

Captures what one or more actors are trying to achieve.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `actor` | string or string[] | Authored | Required. Single actor name or list of actor names. |
| `outcome` | string | Authored | Required. First-person statement of success. |
| `supportedBy` | Relationship[] | Authored | Required (≥1). References to capabilities that fulfill this need. |

**Relationship direction:** Need → Capability (need declares which capabilities support it).

---

## Capability

Represents what the system must be able to do. Capabilities are organized in
a hierarchy and positioned in the UNM value chain by visibility.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `description` | string | Authored | Required. |
| `visibility` | string | Authored | Required for correct UNM rendering. Valid values: `user-facing`, `domain`, `foundational`, `infrastructure`. Inherited from parent if not set. |
| `parent` | string | Authored | Optional. Flat-form parent reference. Name of the parent capability. Mutually exclusive with `children` as an authoring pattern (both resolve to the same hierarchy). |
| `children` | Capability[] | Authored | Optional. Nested child capabilities. Same hierarchy as `parent` references — choose whichever form reads better. |
| `dependsOn` | Relationship[] | Authored | Optional. Other capabilities this capability requires. |
| `realizedBy` | Relationship[] | **Derived** | NOT authored. Computed from `service.realizes` declarations. Represents which services implement this capability. |

**Relationship direction:** Service → Capability (`service.realizes` is the source of truth; `capability.realizedBy` is derived).

---

## Service

Concrete implementation that realizes one or more capabilities.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `description` | string | Authored | Optional. |
| `ownedBy` | string | Authored | Required. Name of the team that owns this service. |
| `realizes` | Relationship[] | Authored | Optional. Which capabilities this service implements. Use role (`primary`, `supporting`, `consuming`) to qualify the relationship. |
| `externalDeps` | string[] | Authored | Optional. Names of external dependencies this service depends on. Must match entries in `external_dependencies`. |
| `dependsOn` | Relationship[] | Authored | Optional. Other services this service calls. |

**Relationship direction:**
- Service → Capability (`realizes`) — service is the source of truth
- Service → ExternalDependency (`externalDeps`) — service is the source of truth
- Service → Service (`dependsOn`)

---

## Team

Organizational unit that owns capabilities. Classified by Team Topologies type.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `type` | string | Authored | Required. Valid values: `stream-aligned`, `platform`, `enabling`, `complicated-subsystem`. |
| `description` | string | Authored | Optional. |
| `size` | int | Authored | Optional. Number of people. Defaults to 5 if not set. |
| `owns` | Relationship[] | Authored | Optional. Capabilities owned by this team. |
| `interacts` | TeamInteraction[] | Authored | Optional. Inline interaction declarations with other teams. |

**TeamInteraction fields:**

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `target` | string | Authored | Required. Name of the other team. |
| `mode` | string | Authored | Required. Valid values: `x-as-a-service`, `collaboration`, `facilitating`. |
| `via` | string | Authored | Optional. Capability mediating the interaction. |
| `description` | string | Authored | Optional. Human-readable label. |

---

## Interaction

A directed relationship between two teams describing how they work together.
In YAML, authored via `team.interacts`. In DSL, also available as standalone
arrow-syntax blocks.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `from` | string | Authored | Required. Name of the initiating team. |
| `to` | string | Authored | Required. Name of the receiving team. |
| `mode` | string | Authored | Required. Valid values: `x-as-a-service`, `collaboration`, `facilitating`. |
| `via` | string | Authored | Optional. Capability mediating the interaction. |
| `description` | string | Authored | Optional. Human-readable label. |

---

## Platform

Groups platform teams that together provide shared capabilities.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `description` | string | Authored | Optional. |
| `teams` | string[] | Authored | Optional. Names of teams in this platform group. |

---

## DataAsset

Shared storage or messaging infrastructure used by multiple services.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `type` | string | Authored | Required. Valid values: `database`, `cache`, `event-stream`, `blob-storage`, `search-index`. |
| `description` | string | Authored | Required. |
| `usedBy` | string[] | Authored | Optional. Names of services that use this data asset. |

**Relationship direction:** DataAsset → Service (`usedBy`). Declared on the data asset.

---

## ExternalDependency

System outside the modeled boundary that internal services depend on.
Declared as definitions only — edges come from `service.externalDeps`.

| Field | Type | Authored / Derived | Notes |
|-------|------|--------------------|-------|
| `name` | string | Authored | Required. Unique within the model. |
| `description` | string | Authored | Required. |
| `usedBy` (internal) | ExternalUsage[] | **Derived** | NOT authored in model files. Populated by the system from `service.externalDeps` declarations. |

**Relationship direction:** Service → ExternalDependency (`service.externalDeps` is the source of truth).

---

## Relationship (shared value type)

Used by `need.supportedBy`, `capability.dependsOn`, `service.realizes`, `service.dependsOn`, `team.owns`.

| Field | Type | Notes |
|-------|------|-------|
| `target` | string | Required. Name of the referenced entity. |
| `description` | string | Optional. Human-readable edge label. |
| `role` | string | Optional. Valid values: `primary`, `supporting`, `consuming`. |

**Short form:** just the name string. **Long form:** object with `target`, `description`, `role`. Both can be mixed in the same list.

---

## Derived Queries (not authored)

The following data is available at query time but is NOT declared in model files:

| Query | Derived From | Notes |
|-------|-------------|-------|
| `capability.realizedBy` | `service.realizes` | Which services implement this capability |
| `team's external deps` | `service.externalDeps` for owned services | External systems reachable through a team |
| `actor's capabilities` | `need.supportedBy` + `need.actor` | Capabilities relevant to an actor |
| `service's capabilities` | `service.realizes` | Capabilities realized by a service |
| `external dep's consumers` | `service.externalDeps` | Services that use this external dependency |

---

## Removed Fields (v2 freeze)

These fields existed in earlier versions of the schema and have been **removed**.
Model files containing these fields will be rejected or silently ignored depending
on the field.

| Field | Removed From | Replacement | Notes |
|-------|-------------|-------------|-------|
| `capability.realizedBy` | `yamlCapability` struct | `service.realizes` | Services now declare what they realize, not vice versa. |
| `capability.ownedBy` | `yamlCapability` struct | `team.owns` | Teams declare what they own. |
| `external_dependency.usedBy` | `yamlExternalDependency` struct | `service.externalDeps` | Services declare which external systems they use. |
| `interactions:` (top-level) | `yamlDocument` struct | `team.interacts` | Inline interactions on teams are the YAML authoring path. DSL arrow syntax kept. |
| `service.type` | `yamlService` struct | (removed — not used) | Service classification was unused. |
| `service.supports` | `yamlService` struct | `service.realizes` | Renamed and direction changed. |
| `service.dataAssets` | `yamlService` struct | `data_asset.usedBy` | Data assets declare their users. |
| `service.externalDependsOn` | `yamlService` struct | `service.externalDeps` | Renamed to `externalDeps`. |
| `scenarios:` (top-level) | `yamlDocument` struct | (removed) | Scenarios section is not parsed or supported. |
| `need.scenario` | `yamlNeed` struct | (removed) | Contextual information belongs in `need.outcome`. |
| `signals:` (top-level) | `yamlDocument` struct | (removed — platform computed) | Signals are analysis outputs, not user-authored data. |
| `pain_points:` (top-level) | `yamlDocument` struct | (removed — platform computed) | Analysis outputs, not user-authored data. |
| `inferred:` (top-level) | `yamlDocument` struct | (removed — platform computed) | Analysis outputs, not user-authored data. |
