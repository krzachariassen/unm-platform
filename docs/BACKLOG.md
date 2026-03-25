# UNM Platform — Product Backlog

## Philosophy

This backlog follows **incremental delivery**. Each phase produces a working, testable artifact. No phase depends on future phases. Each phase is fully covered by TDD and follows Clean Architecture from day one.

**Guiding rule**: At the end of every phase, you can demo something real.

**Tech stack**: Go backend (Clean Architecture) + React frontend (Vite + Tailwind + shadcn/ui).

---

## Phase 1: Domain Model & YAML Parser (Go)

**Goal**: Define the core UNM domain model in Go and parse YAML-based model files into validated in-memory representations.

**Why first**: The model is the foundation. Everything else — parsing, rendering, analysis, AI — is a projection of or operation on this model. Get the model right and everything downstream is simpler.

**Deliverable**: Go CLI tool that reads a `.unm.yaml` file, parses it, validates it, and outputs a summary to stdout.

### Backlog Items

#### 1.1 — Go Project Scaffold
- **Description**: Initialize Go module, directory structure following Clean Architecture (`cmd/`, `internal/domain/`, `internal/usecase/`, `internal/adapter/`, `internal/infrastructure/`), Makefile, `.gitignore`.
- **Acceptance**: `go build ./...` succeeds. `go test ./...` runs (even if no tests yet).

#### 1.2 — Value Objects
- **Description**: Implement all value objects as Go types with constructors and validation.
- **Types**: `EntityID`, `TeamType` (enum), `InteractionMode` (enum), `Confidence` (score + evidence), `MappingStatus` (enum), `Severity` (enum)
- **TDD**: Each value object has table-driven tests for valid input, invalid input, edge cases, equality.
- **Acceptance**: All value object tests green. Zero external dependencies.

#### 1.3 — Core Domain Entities
- **Description**: Implement all core domain entities as Go structs with constructors and domain methods.
- **Entities**: `Actor`, `Scenario`, `Need`, `Outcome`, `Capability`, `Service`, `Team`, `Platform`, `Interaction`, `Signal`, `DataAsset`, `ExternalDependency`, `InferredMapping`
- **Relationships**: All relationships between entities are modeled as `Relationship` structs with `Target` (entity ID), optional `Description` (string), and optional `Role` (primary/supporting/consuming). This applies to: `dependsOn`, `realizedBy`, `supportedBy`, `supports`, `owns`, `provides`. The description is the semantic label rendered on edges in visualization. The role distinguishes primary implementers from supporting/consuming services.
- **New entity types**:
  - `Service`: now has a `Type` field (core-service, worker-service, specialized-service, infrastructure, skeleton)
  - `Signal`: categorized finding (bottleneck, fragmentation, cognitive-load, coupling, gap) with severity, evidence, and affected entities
  - `DataAsset`: storage/messaging infrastructure (database, cache, event-stream, blob-storage, search-index) with service usage relationships
  - `ExternalDependency`: system outside the modeled boundary with service usage
- **Key behaviors**:
  - `Capability`: hierarchical nesting (`AddChild`), fragmentation detection (`IsFragmented`), depth calculation
  - `Team`: cognitive load detection (`IsOverloaded`), capability count
  - `Service`: orphan detection (`IsOrphan`), skeleton detection (`IsSkeleton`)
  - `Need`: capability mapping check
- **TDD**: Each entity has tests for construction, validation, domain methods. Test that relationships carry descriptions and roles correctly.
- **Acceptance**: All entity tests green. Domain package has zero imports outside stdlib.

#### 1.4 — UNM Model Aggregate
- **Description**: Implement `UNMModel` as the root aggregate holding all entities and relationships. Supports adding entities (with duplicate detection), querying by type/ID, relationship traversal.
- **Query methods**: `GetCapabilitiesForTeam`, `GetServicesForCapability`, `GetTeamsForCapability`, `GetNeedsForActor`, `GetOrphanServices`, `GetFragmentedCapabilities`, `GetOverloadedTeams`, `Summary`
- **TDD**: Build test models programmatically and verify all query methods.
- **Acceptance**: Can construct a full model in Go and query all relationships.

#### 1.5 — YAML Parser
- **Description**: Implement a YAML parser using `gopkg.in/yaml.v3` that reads `.unm.yaml` files and produces `UNMModel` instances. Define YAML schema structs, transform to domain entities, handle errors with line context.
- **Relationship parsing**: Must support both short form (`"service-name"`) and long form (`{target: "service-name", description: "..."}`) for all relationship lists (`dependsOn`, `realizedBy`, `supportedBy`, `supports`, `owns`, `provides`). Both forms can be mixed in the same list.
- **TDD**: Parser tests with YAML fixture files in `testdata/` → expected model output. Tests for valid files, malformed YAML, missing required fields, unknown fields. Specific tests for short-form-only, long-form-only, and mixed relationship lists.
- **Acceptance**: Can parse the `examples/inca.unm.yaml` file (which uses described relationships) into a valid `UNMModel` with all relationship descriptions preserved.

#### 1.6 — Validation Engine
- **Description**: Implement `ValidationEngine` that checks a `UNMModel` for structural integrity. Returns typed `ValidationResult` with errors and warnings.
- **Mandatory rules**:
  - Every need must reference at least one capability
  - Every leaf capability must be realized by at least one service OR decompose to sub-capabilities
  - Every service must have an owner team
  - Interaction must reference valid teams and mode
  - Confidence must be between 0.0 and 1.0
- **Warning rules**:
  - Capability owned by more than 2 teams → fragmentation warning
  - Team owning more than 6 capabilities → cognitive load warning
  - Service without any capability support → orphan warning
  - Circular dependencies between capabilities → cycle warning
  - Inferred mapping with confidence < 0.5 → low confidence warning
  - Leaf capability not in any `need.supportedBy` → unlinked capability warning
  - All team interactions use the same mode → interaction diversity warning
  - Capability without `visibility` → missing visibility warning
- **TDD**: Each rule has positive and negative test cases.
- **Acceptance**: Validator catches all defined anti-patterns and returns structured results.

#### 1.7 — CLI: Parse & Validate Command
- **Description**: Build CLI entrypoint (`cmd/cli/main.go`) using stdlib `flag` or cobra. Commands: `parse <file>` (parse + validate + print summary), `validate <file>` (validate only, exit code reflects result).
- **TDD**: Integration tests running CLI against fixture files, checking stdout and exit codes.
- **Acceptance**: `go run ./cmd/cli/ parse examples/inca.unm.yaml` outputs a clean summary with entity counts and any validation warnings.

#### 1.8 — Example Model: INCA
- **Description**: The `examples/inca.unm.yaml` file already exists. Verify it parses and validates cleanly. Fix any issues found during integration.
- **Acceptance**: Full round-trip: YAML → parse → validate → summary with zero errors.

#### 1.9 — UNM Framework Compliance & Schema Refinement

This step brings the model from ~60% UNM framework coverage to 100%, eliminates structural anti-patterns, and aligns the DSL with how UNM is actually defined by Rich Allen ([userneedsmapping.com](https://www.userneedsmapping.com/)) and the [Team Topologies integration](https://teamtopologies.com/key-concepts-content/exploring-team-and-service-boundaries-with-user-needs-mapping).

##### 1.9.1 — Drop `Scenario` Entity

- **Description**: Remove `Scenario` as a first-class entity from the domain model, YAML parser, and validation engine. UNM does not have a "scenario" concept — the process is **Actor → Need → Capability**. The `scenario` entity added an unnecessary indirection (Actor → Scenario → Need → Capability) where the scenario's description duplicated the need's description and the actor reference was redundant since needs already carry an `actor` field. If a user wants to group needs contextually, the `need.description` field is sufficient.
- **Changes**:
  - Remove `Scenario` struct from `entity/` package
  - Remove `scenarios` section from YAML parser schema
  - Remove `need.scenario` field from `Need` entity
  - Remove "Need without a scenario → context warning" validation rule
  - Update `UNMModel` aggregate to remove scenario-related query methods
  - Update test fixtures to remove scenario data
- **TDD**: Tests verify that YAML files without `scenarios` parse correctly. Tests verify that YAML files *with* a `scenarios` section produce a parse warning ("scenarios is deprecated, ignored") rather than a hard error, for backward compatibility.
- **Acceptance**: All existing tests pass without scenarios. The `Need` entity directly references its `Actor` without intermediary. Parser gracefully ignores legacy `scenarios` sections.

##### 1.9.2 — Add Visibility / Layer Field to Capabilities

- **Description**: Add a `visibility` field to `Capability` representing its position in the UNM vertical value chain. This is the defining characteristic of UNM — the vertical axis represents how visible a capability is to the user. Without it, the model is a dependency graph, not a UNM map.
- **Values**: Integer layer (0 = user-facing, higher = deeper infrastructure), or named levels:
  - `user-facing` (layer 0) — Capabilities the user directly interacts with (e.g., Menu Serving, Feed Ingestion API surface)
  - `domain` (layer 1) — Core business processing (e.g., Publishing, Indexing, Query)
  - `foundational` (layer 2) — Internal infrastructure capabilities (e.g., Entity CRUD, Registry, Backup)
  - `infrastructure` (layer 3) — Deep infrastructure (e.g., Data Assets, External Dependencies)
- **Changes**:
  - Add `Visibility` field (string enum or int) to `Capability` entity
  - Add `visibility` to YAML parser schema for capabilities
  - Add validation: warn if capability has no visibility set
  - Add `GetCapabilitiesByLayer()` query to `UNMModel` aggregate
- **TDD**: Capabilities with and without visibility. Query by layer returns correct groupings. Visibility values validate against allowed set.
- **Acceptance**: Parsed model exposes visibility/layer for each capability. The INCA example model can be rendered as a proper vertical UNM value chain.

##### 1.9.3 — Enforce Hierarchical Capability Decomposition

- **Description**: The DSL spec already supports `children` on capabilities but the model treats all capabilities as flat siblings. This step enforces and exercises hierarchical grouping. UNM is iterative — you start with high-level capability groups and drill into sub-chains. This enables L1/L2/L3 zoom levels.
- **Changes**:
  - Ensure `Capability.Children` is fully wired in parser and model
  - Add `GetRootCapabilities()` (returns only top-level capabilities)
  - Add `GetCapabilityDepth()` and `GetCapabilityPath()` methods
  - Add `FlattenCapabilities()` that returns all capabilities across all nesting levels
  - Parent capabilities should automatically inherit visibility from their most user-facing child if not explicitly set
  - Validate: a parent capability should not have `realizedBy` — only leaf capabilities have services
- **TDD**: Models with nested capabilities. Root query returns only parents. Depth and path traversal. Validation that parent capabilities without `realizedBy` don't trigger orphan warnings. Flattening produces complete list.
- **Acceptance**: The INCA example can be restructured with top-level groups (e.g., "Catalog Management", "Ingestion & Sync", "Publication Pipeline", "Consumer Serving", "Operations & Quality") containing the 15 leaf capabilities as children.

##### 1.9.4 — Eliminate Bidirectional Relationship Redundancy

- **Description**: The current schema models the same relationship in both directions, creating data consistency risk and file bloat. Three redundant patterns exist:
  1. `capability.realizedBy: [service]` ↔ `service.supports: [capability]` — same link stated twice
  2. `data_asset.usedBy: [service]` ↔ `service.dataAssets: [data_asset]` — same link stated twice
  3. `external_dependency.usedBy: [service]` ↔ `service.externalDependsOn: [external]` — same link stated twice

  **Resolution**: Keep the **top-down / source-of-truth direction** only and derive the reverse at query time:
  - **Keep** `capability.realizedBy` → **Drop** `service.supports` (derive via `GetCapabilitiesForService()`)
  - **Keep** `data_asset.usedBy` and `data_asset.producedBy/consumedBy` → **Drop** `service.dataAssets` (derive via `GetDataAssetsForService()`)
  - **Keep** `external_dependency.usedBy` → **Drop** `service.externalDependsOn` (derive via `GetExternalDepsForService()`)
- **Changes**:
  - Remove `supports` field from `Service` entity
  - Remove `dataAssets` field from `Service` entity
  - Remove `externalDependsOn` field from `Service` entity
  - Add derived query methods to `UNMModel`: `GetCapabilitiesForService(serviceID)`, `GetDataAssetsForService(serviceID)`, `GetExternalDepsForService(serviceID)`
  - Update YAML parser to ignore `supports`, `dataAssets`, `externalDependsOn` on services (with deprecation warning)
  - Update all test fixtures
- **TDD**: Tests that derived queries return correct results. Tests that models with only top-down declarations work. Tests that legacy YAML with bidirectional declarations parses with deprecation warnings but doesn't error. Round-trip test: derived reverse lookups match what was previously declared explicitly.
- **Acceptance**: Model file size drops significantly (from ~2,300 lines to ~1,400 for the INCA extended example). Zero data consistency risk from dual declarations. All queries still work via derived methods.

##### 1.9.5 — Add `platform` Entity Support

- **Description**: The DSL spec defines `platform` as a first-class entity for grouping platform teams that provide shared capabilities. The YAML parser and model should support it.
- **Changes**:
  - Add `Platform` entity to domain model (name, description, teams list, provides list)
  - Add `platforms` section to YAML parser
  - Add `GetPlatformForTeam()` query
  - Validate: platform teams should have `type: platform`
- **TDD**: Platform with teams. Query platform for team. Validation that non-platform-typed teams in a platform group trigger warnings.
- **Acceptance**: The INCA model can group `inca-core-dev` and `inca-dev` into an "INCA Platform" entity that provides shared capabilities.

##### 1.9.6 — Update Example Models

- **Description**: Update `examples/inca.unm.yaml` and `examples/inca.unm.extended.yaml` to reflect all schema changes from 1.9.1–1.9.5:
  - Remove `scenarios` section
  - Remove `scenario` field from all needs
  - Add `visibility` to all capabilities
  - Group capabilities hierarchically using `children`
  - Remove `supports`, `dataAssets`, `externalDependsOn` from all services (keep only on capabilities, data_assets, external_dependencies)
  - Add `platforms` section
- **TDD**: Updated examples parse and validate cleanly.
- **Acceptance**: Full round-trip for both example files. `inca.unm.extended.yaml` shrinks from ~2,300 to ~1,400 lines. Both examples render a proper UNM vertical value chain with hierarchy.

##### 1.9.7 — UNM Value Chain View Data

- **Description**: Add a new query/projection to the model that produces the data needed to render a proper UNM value chain visualization: user at top → needs → capability groups (by visibility layer) → leaf capabilities → services at bottom. This is the canonical UNM view.
- **Changes**:
  - Add `BuildValueChain()` method to `UNMModel` that returns a layered structure: `[]ValueChainLayer` where each layer has a visibility level and its capabilities, ordered by dependency depth
  - Add `BuildUNMMap(actorFilter)` that produces the full actor → need → capability → service chain for a given actor, respecting visibility ordering
- **TDD**: Value chain from test models produces correct layer ordering. Actor-filtered map returns only relevant paths.
- **Acceptance**: The INCA model produces a value chain where Menu Serving (user-facing) is at the top, Entity CRUD (foundational) is near the bottom, and the dependency arrows flow downward. This is a renderable UNM map.

#### 1.10 — Drop `type` from Service Entity

- **Description**: Remove the `type` field (core-service, worker-service, specialized-service, infrastructure, skeleton) from the `Service` entity. Service types are an operational/infrastructure classification that is not part of the UNM framework. UNM treats services as dependencies in the value chain that realize capabilities — whether a service is a "worker" or "core-service" doesn't affect the Actor→Need→Capability chain, team boundaries, or visibility layers. Keeping non-UNM concerns in the model adds complexity that hurts usability.
- **Guiding principle**: Keep the UNM model clean and simple. Complexity that doesn't serve UNM's purpose should live outside the model (e.g., in a service catalog, operational metadata, or a separate infrastructure model).
- **Changes**:
  - Remove `Type` field from `Service` entity
  - Remove `ServiceType` enum from value objects
  - Remove `type` from YAML parser schema for services
  - Remove `IsSkeleton()` domain method from Service (signals can reference services by name instead)
  - Update YAML parser to ignore `type` on services with deprecation warning
  - Update all test fixtures
  - Update both example YAML files to remove `type` from all services
- **TDD**: Services without `type` parse correctly. Legacy YAML with `type` parses with deprecation warning. All existing queries and analysis still work.
- **Acceptance**: Service entity is simplified to: `name`, `description`, `ownedBy`, `dependsOn`. Example files are leaner. The model contains only UNM-relevant concepts.

---

## Phase 2: Model Querying & Analysis (Go)

**Goal**: Enable programmatic querying and automated analysis of UNM models to surface architectural and organizational insights.

**Why second**: Once you have a valid model, the immediate value is asking questions — where is fragmentation? where is cognitive overload? what capabilities are orphaned? This is the analytical core.

**Deliverable**: CLI commands that output analysis reports (fragmentation, cognitive load, dependencies, gaps) from a parsed model.

### Backlog Items

#### 2.1 — Query Engine
- **Description**: Implement `QueryEngine` as a use case that supports structured queries against a `UNMModel`.
- **Queries**:
  - Find all capabilities for an actor (actor → needs → capabilities)
  - Find all services realizing a capability (direct + transitive)
  - Find all teams owning a capability
  - Find all dependencies of a capability (transitive closure)
  - Find all capabilities for a team
  - Find orphan services (no capability mapping)
  - Find unmapped needs (no capability)
  - Find all interactions for a team
- **TDD**: Each query type tested with fixture models.
- **Acceptance**: All queries return correct results for the INCA example model.

#### 2.2 — Fragmentation Analyzer
- **Description**: Detect capabilities split across multiple teams, and teams co-owning overlapping capabilities. Produce a `FragmentationReport` listing each fragmented capability with its teams.
- **TDD**: Models with known fragmentation patterns → expected detection.
- **Acceptance**: Correctly identifies all fragmentation in test models and the INCA example.

#### 2.3 — Cognitive Load Analyzer
- **Description**: Calculate per-team cognitive load metrics: capability count, service count, dependency count, interaction count. Produce ranked report with threshold-based warnings.
- **TDD**: Models with known load patterns → expected scores and rankings.
- **Acceptance**: Correctly ranks teams and flags overloaded ones.

#### 2.4 — Dependency Analyzer
- **Description**: Analyze service and capability dependency graphs: depth, breadth, cycles, critical paths. Uses graph traversal (DFS/BFS).
- **TDD**: Models with known dependency graphs including cycles and deep chains.
- **Acceptance**: Detects cycles, reports max depth, identifies critical dependency chains.

#### 2.5 — Gap Analyzer
- **Description**: Find needs without capabilities, capabilities without services, services without teams, capabilities without needs.
- **TDD**: Models with deliberate gaps at each layer.
- **Acceptance**: All gap types detected and reported.

#### 2.6 — CLI: Analysis Commands
- **Description**: Add CLI commands for each analysis type: `analyze fragmentation <file>`, `analyze cognitive-load <file>`, `analyze dependencies <file>`, `analyze gaps <file>`, `analyze all <file>`.
- **TDD**: Integration tests with fixture models.
- **Acceptance**: Each command produces correct, readable output. `analyze all` runs everything.

---

## Phase 2.5: Advanced Analysis — Bottleneck, Coupling & Complexity (Go)

**Goal**: Close the gap between what our analyzers detect and what the AI report (see `examples/inca-as-is-unm-map.md`) reveals. The existing Phase 2 analyzers cover fragmentation, basic cognitive load, dependency depth/cycles, and gaps. This phase adds fan-in/fan-out bottleneck detection, data-asset-mediated coupling analysis, and service-level complexity scoring — the patterns the AI report flags as B1-B4, CP1-CP3, and CL1-CL2.

**Why now**: Running `analyze all` against `examples/inca.unm.extended.yaml` surfaces valid findings but misses the most critical signals from the AI report: core having 24 dependents, Kafka-mediated coupling between publisher and indexer, and publisher's 20-workflow complexity. The model already contains the data needed for these analyses (service dependencies, data assets, signals). We just need analyzers that leverage it. Additionally, real-world DSL validation revealed gaps: capabilities not linked to needs go unreported, uniform interaction modes are not flagged, and analysis findings are not translated into actionable signal suggestions.

**Deliverable**: Extended CLI `analyze` output that detects bottleneck services, coupling hotspots, per-service complexity, unlinked capabilities, interaction diversity issues, and auto-generated signal suggestions. Running `analyze all` produces a report that aligns with the AI report's signals and surfaces structural issues that even the AI report missed.

### Backlog Items

#### 2.7 — Bottleneck Analyzer (Fan-In / Fan-Out)

- **Description**: Detect services and capabilities that are depended upon by a disproportionate number of other entities (fan-in bottlenecks) or that depend on a disproportionate number (fan-out risk). Produces a `BottleneckReport` with ranked services by inbound dependency count and outbound dependency count.
- **Metrics**:
  - Fan-in: count of services that list this service in their `dependsOn` (how many things break if this service goes down?)
  - Fan-out: count of services this service depends on (how exposed is this service to upstream failures?)
  - Also compute fan-in for capabilities via `realizedBy` reverse lookups
- **Thresholds**: Flag services with fan-in > 5 as bottleneck warnings. Flag services with fan-in > 10 as critical bottlenecks.
- **Expected INCA results**: `core` and `registry` should both surface as critical bottlenecks with 10+ dependents, matching AI report signals B1 and B2. `async` should surface as a fan-out bottleneck (B4).
- **TDD**: Models with known fan-in patterns → expected detection. Test that self-loops don't inflate counts. Test threshold boundaries.
- **Acceptance**: `analyze bottleneck` correctly identifies `core` and `registry` as the top bottleneck services. Output ranks services by fan-in count.

#### 2.8 — Coupling Analyzer (Data-Asset-Mediated)

- **Description**: Detect implicit coupling between services that share data assets (Kafka topics, databases, caches). Two services that don't directly depend on each other but read/write the same Kafka topic or database table are implicitly coupled — schema changes, lag, or outages in the shared asset affect both. The model's `data_assets` section with `usedBy`, `producedBy`, and `consumedBy` provides this information.
- **Metrics**:
  - For each data asset: list all services that use it, grouped by access pattern (producer/consumer/read-write)
  - Flag data assets used by services owned by different teams (cross-team coupling)
  - Flag producer/consumer pairs where lag or schema drift could cause data staleness
- **Expected INCA results**: Should detect publisher → indexer coupling via shared Kafka topics (CP1). Should detect `inca_catalog_v2` Docstore table shared by registry, publisher, and admin. Should flag cross-team data asset coupling.
- **TDD**: Models with shared data assets across teams → expected coupling detection. Models with single-team data assets → no coupling flagged.
- **Acceptance**: `analyze coupling` surfaces the publisher-indexer Kafka coupling and shared Docstore table coupling, matching AI report signals CP1 and the coupling findings in CAP-03, CAP-06, CAP-07.

#### 2.9 — Service Complexity Scorer

- **Description**: Compute per-service complexity metrics that go beyond team-level breadth. The current Cognitive Load Analyzer (2.3) measures team-level breadth (cap count, svc count). This analyzer measures individual service complexity using data already in the model: dependency count (fan-in + fan-out), number of capabilities it realizes, number of data assets it touches, and signal references.
- **Metrics per service**:
  - `dependencyScore`: count of direct `dependsOn` entries (outbound) + count of services depending on it (inbound)
  - `capabilityScore`: count of capabilities this service appears in via `realizedBy`
  - `dataAssetScore`: count of data assets this service is listed in via `usedBy`
  - `signalScore`: count of signals that reference this service in `affects`
  - `totalComplexity`: weighted sum of the above
- **Expected INCA results**: `publisher` and `menu-cache-worker` should score highest, matching AI report CL1 and CL2. `core` should also score high due to its bottleneck position.
- **TDD**: Models with services of varying complexity → expected scoring and ranking.
- **Acceptance**: `analyze complexity` produces a ranked table of services by complexity score. Top services match the AI report's cognitive load hotspot findings.

#### 2.10 — Self-Dependency Warning

- **Description**: Fix the inconsistency between the dependency analyzer (silently filters self-loops) and the validator (reports capability self-loops as circular deps). Add a dedicated `WarnSelfDependency` validation warning for any entity that lists itself in `dependsOn`. Self-dependencies are modeling errors, not real cycles — they should be reported clearly and filtered from the graph.
- **Changes**:
  - Add `WarnSelfDependency` warning code to the validator
  - Validator emits `WarnSelfDependency` for both services and capabilities that depend on themselves
  - Dependency analyzer no longer needs its own self-loop filtering — the model should be clean after validation
  - Or: strip self-loops in the parser after emitting a parse warning, so all downstream code receives clean data
- **TDD**: Services and capabilities with self-dependencies → expected warnings. Self-loops not reported as cycles. Self-loops not inflating fan-in/fan-out counts.
- **Acceptance**: Self-dependencies produce a clear, distinct warning. No inconsistency between validator and analyzer.

#### 2.11 — Skeleton Service Detection

- **Description**: Distinguish "orphan" services (missing a `realizedBy` link — a modeling gap) from "skeleton" services (deployed but non-functional — an architectural gap). Skeleton detection uses signals: if a signal of category `gap` references a service, and the service is also an orphan, flag it as a skeleton rather than a regular orphan.
- **Expected INCA results**: `serving-gateway`, `status`, and `async-entity-writer` should be flagged as skeletons (matching AI report G1), while `hive-ingester` and `pkg/ (shared)` should remain as regular orphans.
- **TDD**: Models with orphan services that have gap signals → skeleton detection. Orphan services without gap signals → regular orphan.
- **Acceptance**: `analyze gaps` output distinguishes skeletons from orphans. Matches AI report's G1 signal.

#### 2.12 — CLI: Extended Analysis Commands

- **Description**: Add CLI commands for the new analysis types: `analyze bottleneck <file>`, `analyze coupling <file>`, `analyze complexity <file>`. Update `analyze all` to include all new analyzers.
- **TDD**: Integration tests with the INCA extended model.
- **Acceptance**: `analyze all` output now covers fragmentation, cognitive-load, dependencies, gaps, bottleneck, coupling, and complexity. Running against `examples/inca.unm.extended.yaml` produces findings that align with the AI report signals.

#### 2.13 — Unlinked Capability Detector

- **Description**: Detect leaf capabilities that are not referenced by any `need.supportedBy`. These are capabilities that exist in the model but no user need drives them. In UNM, the value chain flows from Actor → Need → Capability — a capability without a need is either internal plumbing that should have `infrastructure` visibility, or indicates a gap in need articulation. The analyzer should distinguish between capabilities with `infrastructure` visibility (expected to be unlinked) and higher-visibility capabilities that lack needs (likely a modeling gap).
- **Metrics**:
  - Count of unlinked leaf capabilities (total, and broken down by visibility level)
  - List of unlinked capabilities with their visibility level
  - Percentage of leaf capabilities linked to needs (coverage metric)
- **Expected INCA results**: Capabilities like `Async Entity Writing`, `Platform Health Monitoring` (infrastructure) are expected to be unlinked. Capabilities like `Catalog Indexing`, `Registry & Configuration Management` (foundational) being unlinked suggests missing internal-actor needs.
- **TDD**: Models with linked and unlinked capabilities at various visibility levels → expected detection and classification.
- **Acceptance**: `analyze gaps` output includes an "unlinked capabilities" section showing which capabilities lack needs and whether this is expected (infrastructure) or suspicious (domain/foundational).

#### 2.14 — Interaction Mode Diversity Analyzer

- **Description**: Analyze team interaction patterns and flag when all interactions use the same mode. Real Team Topologies models typically have a mix of `x-as-a-service`, `collaboration`, and `facilitating` modes. Uniform mode suggests the model is incomplete or that team relationships haven't been thoughtfully classified. Also flag teams that have zero declared interactions (isolated teams).
- **Metrics**:
  - Interaction mode distribution: count of each mode (x-as-a-service, collaboration, facilitating)
  - Teams with no declared interactions (isolation warning)
  - Teams with 4+ interactions of the same type (over-reliance warning)
- **Expected INCA results**: All 12 interactions are `x-as-a-service` → flag for review. `catalog-quality-eng` and `dotcom-eng` should surface as candidates for `collaboration` or `facilitating` modes given their cross-org nature.
- **TDD**: Models with diverse/uniform interaction modes → expected warnings. Models with isolated teams → expected detection.
- **Acceptance**: `analyze cognitive-load` (or a new `analyze interactions` command) surfaces interaction diversity findings.

#### 2.15 — Auto-Generated Signal Suggestions

- **Description**: Automatically suggest signals based on analysis findings. The current analyzers detect patterns (bottlenecks, fragmentation, high complexity) but only report them as analysis output. This step generates candidate `signal` entries that can be added to the model, bridging the gap between analysis results and the declarative signal model. Suggestions are emitted as structured output alongside analysis results, not automatically injected into the model.
- **Signal generation rules**:
  - Fan-in > 10 → suggest `bottleneck` signal (critical)
  - Fan-in 5-10 → suggest `bottleneck` signal (high)
  - Team cognitive load score > 20 → suggest `cognitive-load` signal (high)
  - Capability realized by services from 3+ teams → suggest `fragmentation` signal
  - Dependency chain depth > 4 → suggest `coupling` signal on deepest path
  - Unlinked domain/foundational capability → suggest `gap` signal
- **Expected INCA results**: Should auto-suggest a `cognitive-load` signal for `inca-publisher-dev` (load score 25) and `inca-dev` (load score 22), which the current model's signals section lacks.
- **TDD**: Models with known patterns → expected signal suggestions. Verify suggestions include actionable evidence and correct severity.
- **Acceptance**: `analyze all` output includes a "suggested signals" section with structured signal entries that could be copy-pasted into the YAML model.

---

## Phase 3: REST API Server (Go)

**Goal**: Expose model operations as a REST API that the frontend can consume. This is the bridge between backend and frontend.

**Why third**: The frontend needs an API to talk to. Building the API before the frontend ensures a clean contract and lets us test the API independently.

**Deliverable**: Running Go HTTP server with endpoints for parsing, validation, analysis, and model querying.

### Backlog Items

#### 3.1 — HTTP Server Scaffold
- **Description**: Set up HTTP server in `cmd/server/main.go` with router, middleware (CORS, JSON content-type, logging, error handling), graceful shutdown.
- **TDD**: Server starts, responds to health check.
- **Acceptance**: `go run ./cmd/server/` starts server on configurable port.

#### 3.2 — Model Upload & Parse Endpoint
- **Description**: `POST /api/models/parse` — accepts YAML body or file upload, returns parsed model as JSON.
- **Response shape**: Full model with all entities, relationships, and validation results.
- **TDD**: HTTP tests with test YAML payloads.
- **Acceptance**: Upload INCA YAML, get back full JSON model.

#### 3.3 — Validation Endpoint
- **Description**: `POST /api/models/validate` — validates a model and returns structured errors/warnings.
- **TDD**: Valid and invalid models → expected responses.
- **Acceptance**: Returns clear validation results with error/warning categorization.

#### 3.4 — Analysis Endpoints
- **Description**: `POST /api/models/analyze/{type}` — runs analysis on a submitted model. `POST /api/models/analyze/all` runs all.
- **Types**: `fragmentation`, `cognitive-load`, `dependencies`, `gaps` (Phase 2), `bottleneck`, `coupling`, `complexity`, `interactions` (Phase 2.5)
- **TDD**: Known models → expected analysis JSON.
- **Acceptance**: All analysis types (Phase 2 + 2.5) accessible via API.

#### 3.5 — Query Endpoints
- **Description**: `GET /api/models/{id}/capabilities`, `GET /api/models/{id}/teams`, `GET /api/models/{id}/needs`, etc. — query endpoints for exploring a loaded model.
- **TDD**: Load model, query endpoints, verify results.
- **Acceptance**: All query types from Phase 2 accessible via API.

#### 3.6 — View Data Endpoints
- **Description**: `GET /api/models/{id}/views/{viewType}` — returns pre-computed view data optimized for frontend rendering (nodes + edges + metadata for each view type).
- **View types**: `need`, `capability`, `realization`, `ownership`, `team-topology`, `cognitive-load`
- **TDD**: Models → expected view data shapes.
- **Acceptance**: Each view type returns renderable graph data.

---

## Phase 4: Interactive Web Frontend (React)

**Goal**: Build a beautiful, modern interactive web application for exploring UNM models with multiple views, drill-down, filtering, and anti-pattern highlighting.

**Why fourth**: Backend is solid with API. Now we build the visual layer that makes UNM models tangible and explorable.

**Deliverable**: React application with model upload, multiple interactive views, and analysis dashboards.

**Layout**: The layout is crucial. We must support mulutple layouts but ONE of them must be render as the UNM layout.

**Stack**: Vite + React 19 + TypeScript + Tailwind CSS v4 + shadcn/ui + React Flow (for graph visualization) + Lucide icons.

### Backlog Items

#### 4.1 — Frontend Project Setup
- **Description**: Initialize Vite + React + TypeScript project. Install and configure Tailwind CSS v4, shadcn/ui components, React Router, React Flow, Lucide. Set up path aliases (`@/`). Configure API proxy to Go backend.
- **Acceptance**: `npm run dev` serves the app. Tailwind styles work. shadcn/ui button renders.

#### 4.2 — App Shell & Navigation
- **Description**: Build the application shell: sidebar navigation, top bar with model name, main content area. Use shadcn/ui layout patterns. Routes for each view type + dashboard + upload.
- **Design**: Dark theme. Clean, minimal. Sidebar shows view types as navigation items with icons.
- **Acceptance**: Navigation between views works. Layout is responsive.

#### 4.3 — Model Upload Page
- **Description**: Upload `.unm.yaml` files via drag-and-drop or file picker. Send to backend parse endpoint. Show validation results (errors in red, warnings in amber). On success, navigate to dashboard.
- **Components**: File dropzone, validation result cards, error/warning lists.
- **Acceptance**: Upload INCA YAML → see validation results → navigate to dashboard.

#### 4.4 — Dashboard / Overview Page
- **Description**: Model summary dashboard showing entity counts (cards), key metrics (fragmentation score, cognitive load distribution), and quick-access buttons to each view.
- **Components**: Stat cards, mini charts, navigation cards for each view.
- **Acceptance**: Dashboard shows accurate counts and metrics from backend.

#### 4.5 — Need View
- **Description**: Interactive graph showing actor → need → capability chains. Actors on the left, needs in the middle, capabilities on the right. Color-coded by actor. Click a need to see its supporting capabilities highlighted.
- **Rendering**: React Flow with custom node types for actor, need, capability.
- **Acceptance**: INCA model renders with all actors, needs, and capability connections.

#### 4.6 — Capability View
- **Description**: Hierarchical capability tree with dependency edges. Expandable/collapsible nodes. Shows decomposition levels. Dependency arrows between capabilities.
- **Rendering**: React Flow with hierarchical layout. Collapsible groups for parent capabilities.
- **Acceptance**: INCA capabilities render with hierarchy and dependencies visible.

#### 4.7 — Realization View
- **Description**: Capability → service → team mapping. Three-column layout showing how capabilities are implemented. Click a capability to highlight its services and owning teams.
- **Rendering**: React Flow or Sankey-style diagram.
- **Acceptance**: Full realization chain visible for INCA model.

#### 4.8 — Ownership View
- **Description**: Team-centric view showing what each team owns. Highlight fragmented capabilities (owned by multiple teams) in red. Show team type with color coding (stream-aligned=blue, platform=purple, enabling=green, complicated-subsystem=amber).
- **Rendering**: Grouped layout with teams as containers, capabilities as cards inside.
- **Acceptance**: Fragmentation and team types clearly visible.

#### 4.9 — Team Topologies View
- **Description**: Teams rendered by type with interaction edges. Shows collaboration, x-as-a-service, and facilitating relationships with different edge styles. Platform boundaries drawn as grouped regions.
- **Rendering**: React Flow with custom edge types (solid=collaboration, dashed=x-as-a-service, dotted=facilitating).
- **Acceptance**: All team types and interaction modes visible for INCA model.

#### 4.10 — Cognitive Load Dashboard
- **Description**: Per-team dashboard with bar charts showing capability count, service count, dependency count, interaction count. Threshold lines for warning/critical. Ranked table.
- **Components**: shadcn/ui table + charts (recharts or chart.js).
- **Acceptance**: Teams ranked by cognitive load. Overloaded teams flagged.

#### 4.11 — Anti-Pattern Highlighting
- **Description**: Across all views, visually highlight anti-patterns: fragmented capabilities (red border), overloaded teams (orange badge), orphan services (gray/dashed), missing mappings (dotted placeholder). Click for details panel.
- **Acceptance**: Anti-patterns visible in at least 3 view types.

#### 4.12 — Filter & Search
- **Description**: Global search across all entities. Per-view filters: by actor, by team, by team type, by capability. Filter controls in sidebar or top bar.
- **Acceptance**: Filtering reduces the displayed graph correctly. Search finds entities across types.

---

## Phase 4.5: View API Enrichment — Move Logic to Backend (Go + React)

**Goal**: Eliminate business logic, graph traversal, and data aggregation from the frontend. The React views currently rebuild relationships, compute counts, detect anti-patterns, and derive signals client-side. All of this belongs in the backend. The frontend should receive pre-computed, render-ready data and do nothing but display it.

**Why now**: The frontend (Phase 4) was built iteratively and accumulated logic that violates the clean separation between backend (compute) and frontend (render). Before adding more views or features, we must establish the correct contract: **the API computes, the UI renders**.

**Principle**: A view component should contain zero `for` loops over model relationships, zero threshold checks, zero aggregation `reduce()` calls. If it needs data, the API must provide it in the exact shape the component needs.

**Deliverable**: Enriched view API endpoints that return pre-grouped, pre-aggregated, render-ready structures. Simplified frontend components that map API response → JSX with no intermediate transformation.

### Backlog Items

#### 4.5.1 — Enrich Need View API

- **Description**: The frontend `NeedView.tsx` contains `buildGroups()` which traverses `has need` and `supportedBy` edges to reconstruct actor → need → capability groupings. It also computes `totalNeeds` and `unmappedCount` via reduce. Move all of this to `buildNeedView()` in `view.go`.
- **Current frontend logic**: `buildGroups()` (node filtering, edge traversal, grouping), `totalNeeds` (reduce), `unmappedCount` (reduce)
- **API should return**: `{ groups: [{ actor, needs: [{ need, capabilities }] }], total_needs, unmapped_count }` — or equivalent structured nodes/edges where the grouping is pre-built.
- **Frontend becomes**: Iterate over `groups`, render each actor header and its needs table. Zero graph traversal.
- **TDD**: Test that the view endpoint returns correct groupings, counts for the INCA model.
- **Acceptance**: `NeedView.tsx` has no `buildGroups`, no `reduce`, no edge traversal.

#### 4.5.2 — Enrich Capability View API

- **Description**: The frontend `CapabilityView.tsx` contains `buildCapInfos()` which merges capability and realization views, maps `realizedBy` and `ownedBy` edges, and builds cap → services → teams structures. It also computes `svcCapCount` (service-to-capability count), `highSpanServices` (services realizing 3+ capabilities), and the `fragmented` list. Move all to the backend.
- **Current frontend logic**: `buildCapInfos()` (multi-view merge, edge traversal), `svcCapCount` (aggregation), `highSpanServices` (threshold filter + sort), `fragmented` (anti-pattern detection), leaf count (filter)
- **API should return**: Per-capability info with services, teams, dependencies, children, cross-team flag. Plus `high_span_services`, `fragmented_capabilities`, `leaf_capability_count` as top-level fields.
- **Frontend becomes**: Iterate over visibility bands, render capability cards. Zero graph building.
- **TDD**: Test enriched view response shape with known models.
- **Acceptance**: `CapabilityView.tsx` has no `buildCapInfos`, no `svcCapCount`, no `highSpanServices` computation.

#### 4.5.3 — Enrich Ownership View API

- **Description**: The frontend `OwnershipView.tsx` contains `buildData()` — the most complex client-side function (~90 lines) — which builds team lanes, unowned services, service rows, service-capability counts, and cross-team detection from ownership + realization views. Move entirely to the backend.
- **Current frontend logic**: `buildData()` (multi-view merge, lane building, cross-team detection), `crossTeamCaps` (flatMap + filter), `highSpanSvcs` (filter + sort), `overloadedTeams` (filter), `noCapCount` / `multiCapCount` (aggregation)
- **API should return**: `{ lanes: [{ team, caps: [{ cap, services, crossTeam }] }], unowned_services, service_rows, cross_team_capabilities, high_span_services, overloaded_teams, service_gap_counts }`
- **Frontend becomes**: Toggle between "By Team" (render lanes) and "By Service" (render service_rows). Zero data building.
- **TDD**: Test lanes, cross-team detection, service rows for INCA model.
- **Acceptance**: `OwnershipView.tsx` has no `buildData`, no cross-team computation.

#### 4.5.4 — Enrich Team Topology View API

- **Description**: The frontend `TeamTopologyView.tsx` fetches both team-topology and ownership views, then computes per-team capability counts from ownership edges and builds per-team interaction lists by traversing edges in both directions. Move to backend.
- **Current frontend logic**: `capCounts` (edge traversal for ownership), `teamInteractions` (bidirectional edge aggregation)
- **API should return**: Per-team node data with `capability_count`, `service_count`, `interaction_count`, and `interactions` list (mode, target team, via capability, description). Single view endpoint, no multi-view merge.
- **Frontend becomes**: Group teams by type, render interaction badges. Zero edge traversal.
- **TDD**: Test team node enrichment for INCA model.
- **Acceptance**: `TeamTopologyView.tsx` fetches one endpoint, does no edge traversal.

#### 4.5.5 — Enrich Cognitive Load View API

- **Description**: The cognitive-load view endpoint was partially fixed (now returns `service_count`, `dependency_count`, `interaction_count`, `total_load`). But the frontend still computes sorting, max values, percentages, and threshold checks. Move remaining logic to backend.
- **Current frontend logic**: `sort()` by capability count, `maxCaps` (Math.max), `pct` (percentage calculation), `isOver` (threshold check)
- **API should return**: Pre-sorted team list with `load_percentage`, `status` ("ok" | "warning" | "critical"). Teams already ranked by `total_load` descending.
- **Frontend becomes**: Render the sorted list with bars and status badges. Zero sorting, zero percentage math.
- **TDD**: Test sort order, percentages, status flags.
- **Acceptance**: `CognitiveLoadView.tsx` has no `sort`, no `Math.max`, no threshold logic.

#### 4.5.6 — Enrich UNM Map View API

- **Description**: The frontend `UNMMapView.tsx` is the most logic-heavy component. It merges 4 view endpoints (need, ownership, realization, capability), builds `computeChain()` (BFS traversal for chain highlighting), derives `capDepEdges` from service-level dependencies, and constructs the full node/edge graph with cross-team detection. The data merge and relationship derivation must move to the backend. Layout computation (x/y positioning) can stay client-side as it's a rendering concern.
- **Current frontend logic**: 4-endpoint merge, `computeChain()` (BFS), `capDepEdges` derivation, `capToTeam` / `capToSvcs` maps, cross-team detection
- **API should return**: Single `unm-map` view endpoint with all nodes (actors, needs, capabilities with services/teams embedded, connections) and edges (need chains, capability deps). Chain highlighting could be a separate `GET /models/:id/chain?node=X` endpoint or stay client-side.
- **Frontend becomes**: Receives merged graph data, computes layout positions (x/y), renders React Flow nodes. No model traversal.
- **TDD**: Test merged view data for INCA model.
- **Acceptance**: `UNMMapView.tsx` fetches one endpoint (plus optionally chain), does no multi-view merge.

#### 4.5.7 — Enrich Realization View API

- **Description**: The frontend `RealizationView.tsx` contains `buildRows()` which builds service rows from realization view edges, maps services to teams and capabilities, and sorts. It also computes `noCapCount` and `multiCapCount`. Move to backend.
- **Current frontend logic**: `buildRows()` (edge traversal, grouping, sorting), gap counts
- **API should return**: Pre-built `service_rows` with team, capabilities, and gap counts.
- **Frontend becomes**: Render sorted table. Zero graph building.
- **TDD**: Test service rows for INCA model.
- **Acceptance**: `RealizationView.tsx` has no `buildRows`, no gap counting.

#### 4.5.8 — Anti-Pattern Labels from API

- **Description**: The frontend `AntiPatternPanel.tsx` interprets boolean flags (`is_fragmented`, `is_overloaded`, `is_mapped`, `is_leaf`) into human-readable anti-pattern messages. This interpretation logic belongs in the backend — the API should return structured anti-pattern objects with message, severity, and category.
- **Current frontend logic**: Flag → message mapping, conditional rendering logic
- **API should return**: Per-node `anti_patterns: [{ code, message, severity }]` in view node data.
- **Frontend becomes**: Renders `anti_patterns` array as badges/tooltips. Zero interpretation.
- **TDD**: Test anti-pattern generation for nodes with various flag combinations.
- **Acceptance**: `AntiPatternPanel.tsx` renders pre-built messages, does no flag interpretation.

---

## Phase 4.6: Interactive Cognitive Load & Team Topology Views (Go + React)

**Goal**: Transform the Cognitive Load and Team Topology views from static read-only displays into interactive, information-dense dashboards that let engineers reason about team structure and system health at a glance.

**Why now**: Phase 4.5 gave us enriched API data. Phase 4.6 spends that investment — surfaces team descriptions, interaction details, and load breakdowns in genuinely useful UIs rather than flat tables and card lists.

**Deliverable**: Cognitive Load becomes a stacked-bar breakdown dashboard with clickable drill-down panels. Team Topology becomes an interactive graph with directed, colored edges showing who-talks-to-whom and how.

### Backlog Items

#### 4.6.1 — Enrich Team Topology & Cognitive Load APIs with Named Resource Lists

- **Description**: Both views currently return only *counts* of services and capabilities per team. To power drill-down panels in 4.6.2 and 4.6.3, the backend must return named lists.
- **Changes in `view_enriched.go`**:
  - `enrichedTeamTopologyTeam`: add `Services []string \`json:"services"\`` and `Capabilities []string \`json:"capabilities"\`` (name/label lists, not IDs)
  - `teamLoad` struct: add `Services []string \`json:"services"\`` and `Capabilities []string \`json:"capabilities"\``
  - Populate both from the model: `GetCapabilitiesForTeam`, `GetServicesForTeam` (or equivalent traversal)
- **TDD**: Test that the team-topology and cognitive-load view endpoints return services and capabilities arrays populated for INCA model.
- **Acceptance**: `GET /api/models/:id/views/team-topology` returns `teams[*].services` and `teams[*].capabilities` as string arrays. Same for `cognitive-load`. No counts-only fields removed (backward compatible).
- **File**: `backend/internal/adapter/handler/view_enriched.go` only.

#### 4.6.2 — Cognitive Load View: Stacked Bar + Drill-Down Panel

- **Description**: Replace the single-color load bar with a 4-segment stacked bar (caps, services, deps, interactions). Add threshold markers at 75% and 90%. Make rows clickable to open a drill-down side panel per team. Add sort and filter controls.
- **Frontend changes (`CognitiveLoadView.tsx`)**:
  - **Stacked bar**: 4 colored segments, each segment width = its contribution to total_load as % of max. Colors: caps=purple (#7c3aed), services=blue (#1d4ed8), deps=orange (#b45309), interactions=green (#15803d). Each segment has a hover tooltip showing its value and weight ("5 caps × 2 = 10").
  - **Threshold markers**: Hairline vertical lines at 75% and 90% of the bar track, with tiny "warn" and "crit" labels.
  - **Formula tooltip**: Hover the load number → popover: "Capabilities ×2 + Services + Dependencies + Interactions ×2"
  - **Sort controls**: Buttons — "By Load" (default, desc), "By Name", "By Type". Active button highlighted.
  - **Filter controls**: Dropdown or toggle chips for status (ok/warning/critical) and team type (stream-aligned/platform/enabling/complicated-subsystem).
  - **Clickable rows**: Clicking any row opens a right-side panel (300px, slides in from right) showing:
    - Team name + type badge + description
    - Load breakdown table: each component with its raw value, weight, and contribution
    - "What's driving load" plain-English summary (e.g., "High dependency count (10) suggests this team is tightly coupled to many others.")
    - Owned capabilities (bulleted list from `capabilities[]`)
    - Owned services (bulleted list from `services[]`)
    - Interactions list: who calls them + who they call
    - Anti-patterns (if any)
    - Close button (×)
  - Keep team description visible as a subtitle under the team name in the row (truncated, 1 line).
- **No backend changes** (4.6.1 covers the API additions needed here).
- **Acceptance**: Stacked bar shows 4 segments. Threshold lines visible. Clicking a row opens the panel. Panel shows description, breakdown, services, capabilities. Panel closes on ×. Sort/filter works.
- **File**: `frontend/src/pages/views/CognitiveLoadView.tsx` only.

#### 4.6.3 — Team Topology View: Interactive Graph with Focus Mode

- **Description**: Replace swim-lane cards with an interactive SVG-based graph. Keep the existing table as a "Table" mode toggle. The graph shows teams as nodes and interactions as directed colored arrows.
- **Frontend changes (`TeamTopologyView.tsx`)**:
  - **View toggle**: Header buttons "Graph | Table". Graph is default. Table restores current swim-lane + interaction table layout.
  - **Graph layout** (pure SVG + absolute-positioned divs, no library):
    - Compute positions based on team type column: platform teams in column 1 (left), stream-aligned in column 2 (center), complicated-subsystem + enabling in column 3 (right). Within each column, distribute teams vertically with equal spacing.
    - Canvas: scrollable div with `position: relative`, min-height fits all nodes.
    - Team nodes: rounded cards (180×90px), left border color by team type (existing color scheme), team name + type badge. Show anti-pattern warning dot (orange ●) if `anti_patterns` has any items. Show cognitive load indicator as a thin colored bottom bar (green/amber/red based on load — derive from `is_overloaded` and capability count vs threshold).
    - Fan-in badge: on each node, show "← N" (number of teams that interact with it as target). Platform teams will show high fan-in — this is architecturally significant.
  - **Edges** (SVG `<svg>` overlay with `position: absolute; top:0; left:0; width:100%; height:100%; pointer-events:none`):
    - Bezier curve paths between node centers with arrowhead markers (SVG `<marker>` + `<path>`).
    - Color by mode: `x-as-a-service` = #7c3aed (purple), `collaboration` = #1d4ed8 (blue), `facilitating` = #15803d (green).
    - Default stroke-width: 1.5px. Hover on node: connected edges become 3px.
    - Edge pointer-events enabled for hover: hover an edge → tooltip shows `via` label + description.
  - **Focus mode**: Click a team node:
    - Dim all non-connected nodes (opacity 0.2) and their edges.
    - Highlight the selected node (ring border) and its direct edges (full opacity, thicker).
    - Open right-side detail panel (300px) showing: team name, type, description, owned services list, owned capabilities list, all interactions (in + out) with full via + description, anti-patterns.
    - Click background (or ×) to dismiss focus and panel.
  - **Edge filter bar**: Below header, 3 toggle chips — one per interaction mode. Toggling a chip shows/hides all edges of that mode. Default: all visible.
  - **Interaction mode legend**: Small legend row showing the 3 colors with labels. Already exists in the current header — keep it.
  - Anti-pattern panel: existing `<AntiPatternPanel>` component still used for node click detail (reuse or adapt).
- **No backend changes** (4.6.1 covers services/capabilities lists needed for the panel).
- **Acceptance**: Graph renders with colored directed edges. Clicking a node focuses it and opens panel. Edge hover shows tooltip. Mode filter chips work. "Table" toggle restores old view. Fan-in badges visible on platform nodes.
- **File**: `frontend/src/pages/views/TeamTopologyView.tsx` only.

### Parallelization

After 4.6.1 (backend, must go first):
- Teammate "cogload-ui": implements 4.6.2 — owns `CognitiveLoadView.tsx`
- Teammate "topology-graph": implements 4.6.3 — owns `TeamTopologyView.tsx`

Both frontend items are independent — they touch different files and different views.

---

## Phase 4.7: Deep Capability & Ownership Views (Go + React)

**Goal**: Transform the Capability and Ownership views from flat read-only lists into interactive, hierarchical, insight-driven views. Capability view gains domain-area grouping, dependency graph context, and a detail panel. Ownership view gains collapsible lanes, a metrics bar, and a "By Domain Area" tab that shows cross-team ownership patterns.

**Why now**: Phase 4.5 and 4.6 established rich API data and interactive patterns. Phase 4.7 applies those lessons to the two views most used for architectural reasoning — capability ownership and capability structure.

**Deliverable**: Capability view shows hierarchy, reverse dependencies, and a full detail panel per capability. Ownership view shows collapsible lanes, a structural health bar, and a new "By Domain Area" grouping.

### Backlog Items

#### 4.7.1 — Backend: Reverse Dependency Map + Parent Grouping for Capability View

- **Description**: Add two new computed fields to `buildEnrichedCapabilityView` so the frontend can show fan-in and domain-area hierarchy without doing graph traversal.
- **Changes in `view_enriched.go`** (this file only):
  - `enrichedCapability` struct: add `depended_on_by_count int \`json:"depended_on_by_count"\`` — how many other capabilities list this cap in their `depends_on`.
  - `enrichedCapabilityResponse` struct: add `parent_groups []capParentGroup \`json:"parent_groups"\`` — a flat list of parent (non-leaf) capabilities, each with their children's IDs so the frontend can group leaf cards under domain headers.
  - New struct: `capParentGroup { ID string; Label string; Children []string }` (children = slice of child capability IDs)
  - Compute `depended_on_by_count` by building a reverse map: for each cap, count how many OTHER caps have this cap's ID in their `DependsOn` list.
  - Compute `parent_groups` by iterating `m.Capabilities` and collecting non-leaf capabilities (those with `len(cap.Children) > 0`) with their child IDs.
- **TDD**: Test that `depended_on_by_count` is correct for INCA model (e.g., "Entity CRUD & Lifecycle" should have the highest count). Test that `parent_groups` returns the correct parent→children mapping.
- **File**: `backend/internal/adapter/handler/view_enriched.go` only.
- **Acceptance**: `GET /api/models/:id/views/capability` returns `depended_on_by_count` per capability and `parent_groups` array. Build passes. Tests green.

#### 4.7.2 — Capability View: Hierarchy Mode + Dependency Context + Detail Panel

- **Description**: Redesign `CapabilityView.tsx` to add a second view mode, richer dependency chips, and a clickable detail side panel.
- **File**: `frontend/src/pages/views/CapabilityView.tsx` only.
- **Changes**:
  - **View mode toggle** (top-right of header): "By Visibility" (default, existing grid) | "By Domain". Active button highlighted.
  - **"By Domain" mode**: Renders `parent_groups` from the API as collapsible section headers. Each parent shows: name, child count, expand/collapse chevron. Children render as cards (same card design as existing) beneath the header. Capabilities not under any parent group appear in an "Uncategorized" section. Default: all parents expanded.
  - **Existing "By Visibility" mode**: Keep current band→grid layout. No changes to that mode except for the card improvements below.
  - **Card improvements (apply to both modes)**:
    - **Reverse dep badge**: If `depended_on_by_count > 0`, show `"↙ N"` badge top-right in foundational green (`#065f46` bg `#d1fae5`). High counts (≥5) use amber. Tooltip: "N capabilities depend on this".
    - **Team chips with type color**: Replace uniform gray team chips with colored chips using the existing TEAM_TYPE_CONFIG accent colors. Each team chip: `background: typeConfig.bg, color: typeConfig.accent, border: typeConfig.border`.
    - **`depends_on` as typed chips**: Remove the "needs: ..." plain text. Replace with chips colored by the TARGET capability's `visibility` band color. Use VIS_BANDS color scheme. Label = target cap name. If the depends_on cap isn't in the capabilities list, fall back to a gray chip.
    - **Anti-pattern badge**: If `anti_patterns && anti_patterns.length > 0`, show orange ⚠ count badge top-right next to the fragmented badge.
    - **Clickable card**: `onClick` opens the detail panel (see below). Add `cursor-pointer` and `hover:shadow-md` transition.
  - **Detail side panel** (right-anchored, 340px, fixed, backdrop):
    - Header: cap label + visibility badge + close ×
    - Description (full text, not truncated)
    - **"Depended on by N capabilities"** — if count > 0, list cap names that depend on THIS cap (compute client-side: filter all `viewData.capabilities` where `cap.depends_on.some(d => d.id === selected.id)`)
    - **Depends on** section: list of caps this cap depends on (with their visibility badges)
    - **Services** section: service chips with team ownership and team type dot
    - **Teams** section: team chips with type color + description if available (not in current API — skip if unavailable)
    - **Children** section (if not leaf): list of child capability names
    - **Anti-patterns** section: if any, show as warning items
    - Close on × or backdrop click
- **TypeScript interface**: Add `depended_on_by_count: number` to capability interface. Add `parent_groups: Array<{ id: string; label: string; children: string[] }>` to response interface.
- **Acceptance**: "By Domain" mode shows parent group headers with children nested. Reverse dep badge shows on "Entity CRUD & Lifecycle" (highest fan-in). Clicking a card opens the detail panel. `tsc --noEmit` clean.

#### 4.7.3 — Ownership View: Collapsible Lanes + Metrics Bar + "By Domain Area" Tab

- **Description**: Redesign `OwnershipView.tsx` to add collapsible lanes, a structural health metrics bar, and replace the "By Service" tab with "By Domain Area".
- **File**: `frontend/src/pages/views/OwnershipView.tsx` only.
- **Changes**:
  - **Tab rename**: "By Service" → "By Domain Area". Keep the "By Team" tab. Remove or keep the old service table — replace with new domain area grouping (see below).
  - **Metrics bar** (always visible, below header, above tabs):
    ```
    [9 teams] [27 capabilities] [33 services] [2 cross-team ⚠] [0 unowned]
    ```
    Each metric is a small pill. Cross-team count is amber if > 0. Unowned count is red if > 0. These replace the yellow "Structural Signals" box (remove it — the data is in the metrics bar now).
  - **"By Team" tab — collapsible lanes**:
    - Each team lane is now collapsible. State: `useState<Set<string>>` of expanded team IDs. Default: all expanded (or all collapsed for models with >5 teams — use `viewData.lanes.length > 5 ? new Set() : new Set(all ids)` as initial state).
    - **Collapsed lane header**: shows team name + type badge + first 2 lines of description + cap count badge + service count badge + cross-team indicator if any of their caps are cross-team.
    - **Expanded**: existing cap rows with services (no change to the expanded content).
    - **"Expand all" / "Collapse all"** button pair top-right of the lanes section.
    - Clicking the team header row toggles expand/collapse (currently it opens AntiPatternPanel — move that to a dedicated ⓘ info icon button instead, keep the click for toggle).
  - **Capability row tooltip**: On hover of a capability name in a team lane, show a `title` attribute with the capability's description. This requires either passing descriptions through the lane data (backend change in 4.7.4) or a client-side lookup if capability descriptions are already available somewhere. Since the ownership view doesn't currently include capability descriptions in `cap.data`, use `title={cg.cap.label}` as fallback — descriptions will be added in 4.7.4.
  - **"By Domain Area" tab**:
    - Uses `parent_groups` from the capability API. Requires fetching capability view alongside ownership view (`Promise.all`). Or: compute the grouping client-side from the ownership lane data using the `children[]` relationship (not currently in ownership lane data — use the capability API fetch).
    - Fetch both `ownership` and `capability` view data in parallel on mount.
    - Render: each parent group as a section (e.g., "Feed & Ingestion Management"), with a sub-list of its child capabilities. For each child capability, show: capability name + visibility badge + owning team chip(s). Cross-team capabilities highlighted in amber. This answers: "Is each functional area cleanly owned?"
    - If no parent groups exist (flat model), show a message: "No capability hierarchy defined — all capabilities are top-level."
  - **Service chip popover** (on click): clicking a service chip in a team lane opens a small popover (positioned near the chip) showing: service name, owning team + type, all capabilities it realizes. Use `useState<{svc, position} | null>` and render as an absolutely-positioned card. Click outside to dismiss.
  - Keep `AntiPatternPanel` for the ⓘ team info button and capability name click.
- **Acceptance**: Lanes collapse/expand on header click. Metrics bar shows correct counts. "By Domain Area" tab groups capabilities under their parent groups with team ownership. Service chip click shows popover. `tsc --noEmit` clean.

#### 4.7.4 — Backend: Add Capability Descriptions to Ownership Lane Data

- **Description**: The ownership lane data (`enrichedCapRef` struct) currently has `id`, `label`, and a `data` map with `visibility` and `is_leaf`. Add `description string` so the "By Team" lane can show capability descriptions on hover without a second API call.
- **File**: `backend/internal/adapter/handler/view_enriched.go` only.
- **Change**: In the `capRef` (or `enrichedCapRef`) struct used in `buildEnrichedOwnershipView`, add `Description string \`json:"description"\``. Populate from `cap.Description` during lane building.
- **Frontend use**: In `OwnershipView.tsx`, use `cg.cap.data.description` as the `title` attribute on the capability name button.
- **TDD**: Test that capability descriptions appear in the ownership view lane data.
- **Acceptance**: `GET /api/models/:id/views/ownership` returns `description` in each lane capability. Build + tests green.

### Parallelization

```
Wave 1 (parallel — independent files):
  Teammate "cap-backend":    4.7.1 — view_enriched.go (capability additions only)
  Teammate "own-backend":    4.7.4 — view_enriched.go (ownership lane description)
    → CONFLICT: both touch view_enriched.go. Run sequentially:
      cap-backend (4.7.1) first, then own-backend (4.7.4).

Wave 2 (parallel — after both backend items done):
  Teammate "cap-ui":   4.7.2 — CapabilityView.tsx only
  Teammate "own-ui":   4.7.3 — OwnershipView.tsx only
```

Both 4.7.1 and 4.7.4 touch `view_enriched.go` — run them sequentially with the same agent to avoid conflicts. Both 4.7.2 and 4.7.3 touch different frontend files — run them in parallel.

---

## Phase 4.8: Ownership View UX Redesign (React)

**Goal**: Transform the Ownership View from a dense, flat list of team lanes into a clear, scannable, and visually informative ownership dashboard. The current view is hard to read — capabilities blend together, services lack context, team health is invisible, and there's no way to focus on problems.

**Why now**: The Ownership View is the primary tool for answering "who owns what?" — the most common question in any architecture review. If this view is unclear, the entire platform loses credibility. Phase 4.5–4.7 delivered the data; this phase makes it usable.

**Principle**: Every row should communicate ownership health at a glance without clicking. Color, proportion, and spatial grouping replace text-heavy density.

**Deliverable**: A polished Ownership View where teams, capabilities, and services are visually distinct, problem areas are immediately obvious, and the "By Domain Area" tab is genuinely useful for cross-team governance.

### Backlog Items

#### 4.8.1 — Ownership View: Visual Hierarchy & Team Health Cards

- **Description**: Replace the current flat team lane headers with richer team health cards. The header row currently shows team name, type text, and truncated description — all in the same visual weight. Redesign the team lane header to communicate team health at a glance.
- **File**: `frontend/src/pages/views/OwnershipView.tsx`
- **Changes**:
  - **Team header redesign**: Replace the flat `px-4 py-2.5` header with a two-row header card:
    - Row 1: Team name (bold) + type badge (colored, not plain text) + overloaded/cross-team badges + expand chevron
    - Row 2 (always visible, even when collapsed): Description text (up to 2 lines, truncated with `…`), plus a mini metric strip: `[N caps · M services · K interactions]` in muted text
  - **Left accent bar by health**: Instead of a static team-type color accent, use a health-based accent: green (healthy), amber (cross-team or warning signals), red (overloaded). Team type is already shown as a badge, so the accent can carry health semantics.
  - **Collapsed state improvements**: When collapsed, show the first 3 capability names inline as gray chips after the metrics strip, giving a preview without expanding.
- **TDD**: N/A (frontend visual change only).
- **Acceptance**: Team lanes show description, metrics, and preview chips when collapsed. Overloaded teams have red accent. `tsc --noEmit` clean.

#### 4.8.2 — Ownership View: Capability Rows with Visibility Bands

- **Description**: The current capability rows inside team lanes are undifferentiated — all capabilities look the same regardless of whether they're user-facing or infrastructure. Add visibility band coloring and structural indicators to each capability row.
- **File**: `frontend/src/pages/views/OwnershipView.tsx`
- **Changes**:
  - **Visibility indicator**: Add a small colored dot or thin left border to each capability row using the VIS_BANDS color scheme (blue=user-facing, purple=domain, green=foundational, gray=infrastructure). This gives instant visual grouping within a team lane.
  - **Group capabilities by visibility within each lane**: Sort capabilities within a team lane by visibility level (user-facing first, infrastructure last) with subtle divider lines between groups. Add tiny visibility label headers (e.g., "user-facing · 2") between groups.
  - **Service chip improvements**: Show service count per capability as a small badge `(×N)` when a capability has 3+ services. Color service chips by their owning team type (not just gray/amber). Cross-team services get a distinct border treatment.
  - **Remove `maxWidth: 220` truncation on capability names**: Let names wrap naturally or increase to 320px. Current truncation hides important context.
- **Acceptance**: Capabilities within a team lane are visually grouped by visibility. Service chips have team-type coloring. `tsc --noEmit` clean.

#### 4.8.3 — Ownership View: Problem-Focused Filtering & Signals Bar

- **Description**: The current view has no way to filter to just problem areas. Add filter controls that let users focus on ownership anti-patterns: cross-team capabilities, unowned capabilities, overloaded teams, and teams with no services.
- **File**: `frontend/src/pages/views/OwnershipView.tsx`
- **Changes**:
  - **Metrics bar upgrade**: Replace the plain-text metrics bar with clickable signal pills that act as toggle filters. When a pill is active, only relevant lanes/capabilities are shown:
    - `"2 cross-team"` → filters to show only lanes containing cross-team capabilities
    - `"6 unowned"` → shows the unowned capabilities section prominently
    - `"1 overloaded"` → filters to only overloaded team lanes
  - **Team type filter chips**: Add toggle chips for team types (stream-aligned, platform, enabling, complicated-subsystem) similar to the Cognitive Load view. Active chips filter visible lanes.
  - **"Show problems only" toggle**: Single toggle button that hides all clean, healthy team lanes and shows only lanes with at least one signal (cross-team, overloaded, or containing capabilities that are also unowned elsewhere).
- **Acceptance**: Clicking signal pills filters the view. Team type chips filter lanes. "Show problems only" shows a focused problem view. `tsc --noEmit` clean.

#### 4.8.4 — Ownership View: "By Domain Area" Tab Redesign

- **Description**: The current "By Domain Area" tab is a flat list of domain groups with capability names and team badges. It doesn't answer the key question: "Is each functional area cleanly owned by one team?" Redesign it to make ownership health per domain immediately visible.
- **File**: `frontend/src/pages/views/OwnershipView.tsx`
- **Changes**:
  - **Domain group header with ownership summary**: Each domain group header should show: group name + total capabilities + primary owning team (the team that owns the majority) + "N teams" count if multiple teams own capabilities in this domain. Color the header accent by ownership clarity: green if 1 team owns all, amber if 2 teams, red if 3+ teams.
  - **Ownership matrix mini-view per domain**: Below the header, show a small horizontal bar chart or dot grid showing what percentage each team owns of this domain. Example: `[inca-core-dev ████████░░ 80%] [inca-ingestion-dev ██░░░░░░░░ 20%]`.
  - **Capability rows with team alignment**: Within each domain group, indent capability rows and show the owning team chip aligned to a consistent column. Cross-team capabilities are highlighted with amber background and show all owning teams.
  - **Unowned capabilities integrated**: Instead of showing unowned capabilities in a separate section at the bottom, show them inline within their parent domain group (if they have one) with a red "unowned" badge. Only truly orphaned capabilities (no parent group) go to a separate section.
- **Acceptance**: Each domain area header shows ownership distribution. Cross-domain ownership fragmentation is visible at a glance. `tsc --noEmit` clean.

#### 4.8.5 — Ownership View: Service Detail Popover Enhancement

- **Description**: The current service chip popover shows only service name and owning team — not enough to be useful. Enhance it to show a complete service context card.
- **File**: `frontend/src/pages/views/OwnershipView.tsx`
- **Changes**:
  - **Enriched popover content**: When clicking a service chip, the popover should show:
    - Service name (mono font, bold)
    - Owning team + team type badge
    - Capabilities this service realizes (list of capability names with visibility dots)
    - If the service realizes 3+ capabilities: show `"high-span service"` warning
    - If the service is from a different team than the lane: show `"cross-team dependency"` indicator
  - **Backend prerequisite**: The ownership view API may need to include per-service capability lists. If not available, fetch from the capability view endpoint (already loaded for "By Domain Area" tab).
  - **Popover positioning**: Ensure the popover doesn't overflow viewport edges. Use smart positioning (above if near bottom, left if near right edge).
- **Acceptance**: Service popover shows capabilities, team type, and cross-team indicators. Popover positions correctly at viewport edges. `tsc --noEmit` clean.

### Parallelization

```
Wave 1 (parallel — independent concerns):
  Teammate "health":   4.8.1 — Team health cards (lane headers only)
  Teammate "filter":   4.8.3 — Filtering & signals bar (controls + filter logic)

Wave 2 (depends on Wave 1):
  Teammate "caps":     4.8.2 — Capability rows with visibility bands
  Teammate "domain":   4.8.4 — "By Domain Area" redesign
  Teammate "popover":  4.8.5 — Service popover enhancement
```

Wave 1 items touch different parts of OwnershipView.tsx (header vs controls bar). Wave 2 items depend on Wave 1 lane structure and can run in parallel since they touch different sections of the component.

---

## Phase 4.9: External Dependencies in Views (Go + React)

**Goal**: Surface `external_dependencies` as first-class visible entities in the platform. Currently, external deps are parsed and stored but never rendered anywhere. This is a gap — cross-org coupling and blast radius are key architectural concerns.

**Why now**: External dependencies are structurally equivalent to services for ownership and coupling purposes. They should appear wherever services appear in the context of "what does this team depend on?"

**Deliverable**: External dependencies visible in Ownership View (per-team section), Realization View (extra column), and Dashboard summary stats.

### Backlog Items

#### 4.9.1 — Backend: Add External Dependencies to Ownership View API

- **Description**: Extend `buildEnrichedOwnershipView` to include external dependencies per team. For each team lane, add an `external_deps` array of the external dependencies that team's services depend on (via the `ExternalDependency.UsedBy` relationship).
- **File**: `backend/internal/adapter/handler/view_enriched.go`
- **Changes**:
  - In the `enrichedOwnershipLane` struct, add `ExternalDeps []extDepRef \`json:"external_deps"\`` where `extDepRef { ID, Label, Description string; ServiceCount int }`.
  - For each lane (team), scan `ExternalDependency.UsedBy` in the model: if any `usedBy.Target` is a service owned by this team, include that external dep in the lane's `external_deps`.
  - Also add a top-level `external_dependency_count int` to the response struct.
- **TDD**: Test that INCA extended model correctly assigns external deps to their teams' lanes.
- **Acceptance**: `GET /api/models/:id/views/ownership` returns `external_deps` in each lane and `external_dependency_count` at top level. Build + tests green.

#### 4.9.2 — Backend: Add External Dependencies to Realization View API

- **Description**: Extend `buildEnrichedRealizationView` to include external dependency usage per service.
- **File**: `backend/internal/adapter/handler/view_enriched.go`
- **Changes**:
  - In the `serviceRow` struct used by the realization view, add `ExternalDeps []string \`json:"external_deps"\`` — names of external systems this service depends on.
  - Populate by scanning `ExternalDependency.UsedBy` for each service.
- **TDD**: Test external deps appear in service rows.
- **Acceptance**: `GET /api/models/:id/views/realization` returns `external_deps` per service row. Build + tests green.

#### 4.9.3 — Ownership View: External Deps Section per Team Lane

- **Description**: Add an "External Dependencies" collapsible section at the bottom of each team lane in the Ownership View, showing which external systems this team's services depend on.
- **File**: `frontend/src/pages/views/OwnershipView.tsx`
- **Changes**:
  - After the existing capabilities + services content in each team lane, add: if `lane.external_deps.length > 0`, render a row: `[External] ext-dep-name · ext-dep-name ...` using gray pill chips with an external link icon (→) prefix.
  - Each external dep chip shows: name + `(N services)` count if > 1.
  - The metrics bar should show `external_dependency_count` if > 0 as a neutral-gray pill: `"3 external deps"`.
  - Update TypeScript interface to include `external_deps` in lane type.
- **Acceptance**: External deps appear in team lanes. Metrics bar shows count. `tsc --noEmit` clean.

#### 4.9.4 — Realization View: External Deps Column

- **Description**: Add an "External Deps" column to the Realization View table showing which external systems each service depends on.
- **File**: `frontend/src/pages/views/RealizationView.tsx`
- **Changes**:
  - Add a column header "External Deps" after the Capabilities column.
  - For each service row, show gray pill chips for each external dep name. If empty, show `—`.
  - Update TypeScript interface to include `external_deps: string[]` in service row type.
- **Acceptance**: Realization View shows external deps per service. `tsc --noEmit` clean.

#### 4.9.5 — Dashboard: External Dependency Count in Summary Stats

- **Description**: Add `external_dependencies` to the Model Statistics grid on the Dashboard, so users can immediately see the cross-org dependency count.
- **File**: `frontend/src/pages/DashboardPage.tsx`
- **Changes**:
  - Add `external_dependencies: 'External Deps'` to `SUMMARY_LABELS`.
  - The parse summary already returns this key from the backend — just add it to the label map.
- **Acceptance**: Dashboard shows "External Deps" stat card. `tsc --noEmit` clean.

### Parallelization

```
Wave 1 (parallel — independent backend files):
  Teammate "own-backend":   4.9.1 — view_enriched.go (ownership section only)
  Teammate "real-backend":  4.9.2 — view_enriched.go (realization section only)
    → CONFLICT: both touch view_enriched.go. Run sequentially: 4.9.1 then 4.9.2.

Wave 2 (parallel — frontend, after backend done):
  Teammate "own-ui":   4.9.3 — OwnershipView.tsx
  Teammate "real-ui":  4.9.4 — RealizationView.tsx
  Teammate "dash-ui":  4.9.5 — DashboardPage.tsx
```

---

## Phase 4.10: UNM Value Chain Risk Surfacing & Signals View (Go + React)

**Goal**: Surface risks that propagate through the UNM value chain (Actor → Need → Capability → Service → Team). The platform currently surfaces team-layer and structural risks well (cognitive load, bottlenecks, fragmentation), but it's blind to the user-facing end of the chain. This phase fills that gap and creates a structured health view organized by UNM layers — the input feed for Phase 6 AI.

**Why now**: Every analyzer the platform has operates at the team/service layer. But UNM's differentiator is that it starts from user needs. If a single need requires 3+ teams to deliver, that's a coordination risk. If a user-facing capability has cross-team service ownership, that's a reliability risk. These are the insights that make UNM valuable — and they must be visible before AI can reason about them.

**Critical connection to Phase 6 AI**: The Signals View built here becomes the structured data feed that AI prompts consume. When a user asks the AI "What are the top risks?", the AI reads the value-chain-organized signals, not raw analyzer dumps. The value stream analysis (4.10.5) is the foundation for AI recommendations like "Team A is not a real value stream — recompose it around needs X and Y by moving services A and B."

**Deliverable**: Value chain risk analysis API, UNM Signals View (three-layer health dashboard), Need View team-span enrichment, Capability View user-facing-at-risk section, and value stream coherence analysis.

### Backlog Items

#### 4.10.1 — Backend: Value Chain Traversal Analyzer (Go)

- **Description**: New analyzer that traverses the full UNM value chain for each need: Need → `supportedBy` → Capabilities → `realizedBy` → Services → `ownedBy` → Teams. Computes per-need:
  - `team_span`: count of distinct teams involved in delivering this need
  - `capability_count`: how many capabilities support this need
  - `service_count`: total services across the delivery chain
  - `teams`: list of team IDs involved
  - `cross_team`: boolean (team_span > 1)
  - `at_risk`: boolean (team_span > 2, or any team in chain has high structural load)
- **Layer**: `infrastructure/analyzer/value_chain.go`
- **TDD**: Known model with needs spanning 1, 2, 3+ teams → expected team_span and risk flags. Need with capability backed by overloaded team → at_risk=true. Need with no capabilities → flagged as unbacked.
- **Acceptance**: Analyzer produces correct traversal for all needs. Self-contained, no UI dependency.

#### 4.10.2 — Backend: Signals API Endpoint (Go)

- **Description**: New `GET /api/models/{id}/views/signals` endpoint that aggregates findings from ALL existing analyzers and the new value chain analyzer, organized by the three UNM layers:
  ```json
  {
    "health": { "ux_risk": "red", "architecture_risk": "amber", "org_risk": "red" },
    "user_experience_layer": {
      "needs_requiring_3plus_teams": [...],
      "needs_with_no_capability_backing": [...],
      "needs_fully_owned_by_high_load_team": [...]
    },
    "architecture_layer": {
      "user_facing_caps_with_cross_team_services": [...],
      "capabilities_not_connected_to_any_need": [...],
      "capabilities_fragmented_across_teams": [...]
    },
    "organization_layer": {
      "top_teams_by_structural_load": [...],
      "critical_bottleneck_services": [...],
      "teams_with_no_stream_coherence": [...]
    }
  }
  ```
  Health indicators: red (3+ findings), amber (1-2 findings), green (0 findings) per layer.
- **Layer**: `adapter/handler/signals.go`
- **TDD**: HTTP test with INCA model → expected signal counts per layer. Empty model → all green. Model with known issues → correct risk levels.
- **Acceptance**: Endpoint returns structured three-layer findings. Replaces the frontend's current `analyzeAll + extractSignals()` patchwork on the Dashboard.

#### 4.10.3 — Frontend: Signals View (React)

- **Description**: New `SignalsView.tsx` — the most important missing view. Lives in navigation as "Health" (or "Signals"). Organized by UNM value chain layer with a three-column health header for board-level read in 5 seconds.
- **Layout**:
  - **Top row**: Three cards — "UX Risk" | "Architecture Risk" | "Org Risk" — each with red/amber/green indicator and finding count
  - **User Experience Layer section**: Expandable list of needs with team-span issues, unbacked needs, needs owned by overloaded teams
  - **Architecture Layer section**: User-facing capabilities at cross-team risk, unlinked capabilities, fragmented capabilities
  - **Organization Layer section**: Top 3 teams by structural load severity (link to Cognitive Load view), critical bottleneck services count (link to analysis)
  - Each finding row: entity name + specific risk description + severity badge
- **Layer**: `frontend/src/pages/views/SignalsView.tsx`
- **Acceptance**: View renders three layers with correct data from signals API. Health indicators match finding counts. Clicking findings navigates to relevant detail views. `tsc --noEmit` clean.

#### 4.10.4 — Need View: Team Span Enrichment (Go + React)

- **Description**: Enrich the Need View with delivery chain visibility. Each need card shows a team-span badge: how many distinct teams are involved in delivering that need. When span > 2, the card gets a red left border and the badge turns red.
- **Backend**: Enrich the Need View API response to include `team_span` (distinct owning teams across the need's delivery chain) and `teams` (list of team IDs). Uses the value chain analyzer from 4.10.1.
- **Frontend**: `NeedView.tsx` — add team-span badge per need row. Red when > 2. Show team chips on hover/click.
- **TDD**: Backend: need with 1 team → span=1, need with 3 teams → span=3, need with no capabilities → span=0 (unbacked).
- **Acceptance**: Need View shows team span per need. High-span needs visually distinguished. `tsc --noEmit` clean.

#### 4.10.5 — Backend: Value Stream Coherence Analyzer (Go)

- **Description**: The missing link between Team Topologies and UNM. Teams declare themselves as `stream-aligned`, but nothing validates whether the needs they serve actually form a coherent value stream. This analyzer answers: "Is team X really a value stream, or is it a grab-bag of unrelated needs?"
- **Analysis per stream-aligned team**:
  - **Needs served**: Which user needs does this team's delivery chain touch?
  - **Need clustering**: Are these needs related (shared capabilities, shared actors) or scattered?
  - **Stream coherence score**: Ratio of shared-capability connections to total need count. High coherence = genuine value stream. Low coherence = organizational accident.
  - **Stream recommendation**: If coherence is low, which needs form natural clusters that could be split into coherent streams?
- **Layer**: `infrastructure/analyzer/value_stream.go`
- **TDD**: Team serving related needs (shared capabilities) → high coherence. Team serving unrelated needs → low coherence. Non-stream-aligned team → skipped. Team with only 1 need → coherence=1.0 (trivially coherent).
- **Acceptance**: Each stream-aligned team gets a coherence assessment. Low-coherence teams flagged in signals view (4.10.2 updated to include). This data feeds directly into Phase 6 AI prompts for value stream recommendations.

#### 4.10.6 — Capability View: User-Facing At-Risk Section (React)

- **Description**: User-facing capabilities (visibility=user-facing) whose `realizedBy` services come from 2+ different teams are the highest-priority architectural risk. Add a dedicated "User-facing at risk" section at the top of the Capability View before the full list.
- **Layer**: `frontend/src/pages/views/CapabilityView.tsx`
- **Changes**: Filter capabilities where `visibility === 'user-facing' && teams.length >= 2`. Show in a red-accented section at top with team chips and service list. The API already returns visibility, is_fragmented, and teams — no backend change needed.
- **Acceptance**: User-facing cross-team capabilities appear in prominent top section. `tsc --noEmit` clean.

#### 4.10.7 — Dashboard: Replace extractSignals() with Signals API

- **Description**: The Dashboard currently has a `extractSignals()` function that manually parses the `analyzeAll` response client-side. Replace this with a single call to the new signals API endpoint from 4.10.2. The Dashboard becomes a welcome/summary screen that shows the three-layer health indicators and links to the Signals view for details.
- **Layer**: `frontend/src/pages/DashboardPage.tsx`
- **Changes**: Remove `extractSignals()` and its `analyzeAll` fetch. Replace with `api.getViewSignals(modelId)`. Show health indicators as the primary visual. Add "View Details →" link to SignalsView.
- **Acceptance**: Dashboard loads faster (one API call instead of analyzeAll). Health indicators match Signals View. `tsc --noEmit` clean.

### Parallelization

```
Wave 1 (parallel — independent backend analyzers):
  Teammate "value-chain":   4.10.1 — Value Chain Traversal Analyzer
  Teammate "stream":        4.10.5 — Value Stream Coherence Analyzer

Wave 2 (after Wave 1 — needs analyzer outputs):
  Teammate "signals-api":   4.10.2 — Signals API Endpoint

Wave 3 (parallel — frontend, after API exists):
  Teammate "signals-ui":    4.10.3 — Signals View
  Teammate "need-ui":       4.10.4 — Need View team span (backend + frontend)
  Teammate "cap-ui":        4.10.6 — Capability View user-facing at-risk section
  Teammate "dash-fix":      4.10.7 — Dashboard signals API integration
```

---

## Phase 5: Custom DSL Parser (Go)
Won't do!
---

## Phase 6: AI-Powered Interactive Platform

**Goal**: Transform the UNM platform from a read-only visualization tool into an interactive advisor that understands architecture, recommends improvements, and lets users explore what-if scenarios — both manually and with AI assistance.

**Why sixth**: Model, analysis, visualization, and UNM value chain risk surfacing (Phase 4.10) are stable. The platform computes fragmentation, cognitive load, bottlenecks, coupling, gaps, value stream coherence, and need delivery risks. The Signals View organizes all of this by UNM layer. AI can now reason over ALL of this structured data to answer organizational design questions that no deterministic analyzer can handle. This is the strategic differentiator — no other tool does this.

**What Phase 6 is NOT**: It is not about AI generating YAML from codebases. That is already solved by `docs/CODE_TO_DSL_AGENT.md` as a pre-step. Embedding code scanning into the platform would change its boundary from architecture modeling to code analysis, which is a different product. AI model generation stays external; AI model understanding and advising is the platform's superpower.

**Deliverable**: Changeset engine for what-if exploration, prompt library for AI-assisted analysis, OpenAI integration, chat-style advisor UI, and 100+ tested AI question/answer pairs covering UNM and Team Topologies.

### Architecture Decision: How What-If Works

The platform maintains the concept of a **Changeset** — a set of proposed modifications to the loaded model. This enables both human and AI-driven what-if without modifying the source YAML.

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────────┐
│  Source Model     │ +   │  Changeset       │  →  │  Projected Model     │
│  (original YAML)  │     │  (move svc,      │     │  (source + changes   │
│                   │     │   split team,     │     │   applied in-memory) │
│                   │     │   add capability) │     │                      │
└──────────────────┘     └──────────────────┘     └──────────────────────┘
        ↓                                                  ↓
  Run ALL analyzers                                 Run ALL analyzers
        ↓                                                  ↓
  Original analysis                                Projected analysis
        ↓                                                  ↓
        └──────────────── DIFF ────────────────────────────┘
                           ↓
                    Impact Report
           "Moving service X reduces Team A
            cognitive load from HIGH to MEDIUM,
            but creates fragmentation in Cap Z"
```

This is the foundation for everything in Phase 6:
- **Human what-if**: User creates changesets in the UI → sees structural impact
- **AI what-if**: AI proposes changesets → same engine → same impact assessment
- **AI advisor**: AI receives original model + analysis (+ optionally projected model + analysis) and reasons about recommendations in natural language

The Changeset engine is deterministic Go code. The AI layer adds natural language reasoning on top. This separation means the platform works without AI (pure what-if), and AI makes it dramatically more useful (explains WHY and suggests WHAT).

### Architecture: Clean Architecture Layers

```
domain/entity/         → Changeset, ChangeAction, ImpactReport
domain/service/        → ChangesetApplier (applies changeset to model copy)
infrastructure/ai/     → PromptLibrary, OpenAIClient, PromptRenderer
adapter/handler/       → /api/models/{id}/changesets, /api/models/{id}/ask
```

### Architecture: Prompt Library

AI capabilities are driven by a **prompt library** — a set of versioned prompt templates stored as Go embedded files. Each prompt is a text template that receives structured data (model summary, analysis results, specific question context) and produces a focused query to OpenAI.

```
prompts/
  advisor/
    structural-load.tmpl      → "Given this team's load dimensions..."
    service-placement.tmpl    → "Given these dependency patterns..."
    team-boundary.tmpl        → "Given these interaction patterns..."
    fragmentation.tmpl        → "Given these fragmented capabilities..."
    bottleneck.tmpl           → "Given these bottleneck services..."
    value-stream.tmpl         → "Given this team's stream coherence and needs..."
    need-delivery-risk.tmpl   → "Given this need's team span and delivery chain..."
  whatif/
    impact-assessment.tmpl    → "Compare original vs projected analysis..."
    transition-plan.tmpl      → "Suggest steps to reach this target state..."
  query/
    natural-language.tmpl     → "Answer this question about the model..."
    model-summary.tmpl        → "Summarize this architecture model..."
    health-summary.tmpl       → "Summarize value chain risks by UNM layer..."
```

Prompts are iterable — improving a prompt immediately improves every answer without code changes. Each prompt includes UNM and Team Topologies framework context so the AI reasons within the correct domain.

### Backlog Items

#### 6.1 — Changeset Domain Model (Go)
- **Description**: Define `Changeset` and `ChangeAction` domain entities. A changeset is an ordered list of actions: `MoveService(svc, fromTeam, toTeam)`, `SplitTeam(team, newTeamA, newTeamB, serviceAssignments)`, `MergeTeams(teamA, teamB, newTeam)`, `AddCapability(...)`, `RemoveCapability(...)`, `ReassignCapability(cap, fromTeam, toTeam)`, `AddInteraction(...)`, `RemoveInteraction(...)`, `UpdateTeamSize(team, size)`.
- **Layer**: `domain/entity/changeset.go`
- **TDD**: Action creation, serialization, validation (e.g., can't move a service to a non-existent team).
- **Acceptance**: All action types constructible, serializable to JSON, validatable against a model.

#### 6.2 — Changeset Applier (Go)
- **Description**: Domain service that takes a `UNMModel` + `Changeset` and produces a new `UNMModel` (deep copy with changes applied). The original model is never mutated.
- **Layer**: `domain/service/changeset_applier.go`
- **TDD**: Each action type applied to a known model → expected output model. Invalid actions → error. Composite changesets (multiple actions) → expected cumulative result.
- **Acceptance**: All action types produce correct projected models. Original model untouched.

#### 6.3 — Impact Analyzer (Go)
- **Description**: Runs ALL existing analyzers (fragmentation, cognitive load, dependencies, gaps, bottleneck, coupling, complexity) on both original and projected models, then diffs the results. Produces an `ImpactReport` with per-dimension before/after comparisons.
- **Layer**: `infrastructure/analyzer/impact.go`
- **TDD**: Known model + known changeset → expected impact deltas (e.g., "team X cognitive load: domain_spread high→medium").
- **Acceptance**: Impact report correctly identifies all changed dimensions. No false positives on unchanged dimensions.

#### 6.4 — Changeset REST API (Go)
- **Description**: API endpoints for changeset management:
  - `POST /api/models/{id}/changesets` — create a new changeset (JSON body with actions)
  - `GET /api/models/{id}/changesets/{csId}` — get changeset details
  - `GET /api/models/{id}/changesets/{csId}/projected` — get the projected model (parse response format)
  - `GET /api/models/{id}/changesets/{csId}/impact` — get the impact report (all analyzer diffs)
  - `POST /api/models/{id}/changesets/{csId}/apply` — export projected model as YAML (download)
- **Layer**: `adapter/handler/changeset.go`, `adapter/repository/changeset_store.go`
- **TDD**: HTTP tests for each endpoint. Changeset stored in memory (same as model store).
- **Acceptance**: Full CRUD for changesets. Impact endpoint returns structured before/after diffs.

#### 6.5 — Prompt Library Infrastructure (Go)
- **Description**: Build the prompt library system: `PromptLibrary` loads embedded `.tmpl` files, `PromptRenderer` injects model/analysis data into templates, `OpenAIClient` sends rendered prompts and returns structured responses.
- **Layer**: `infrastructure/ai/prompt_library.go`, `infrastructure/ai/openai_client.go`, `infrastructure/ai/renderer.go`
- **TDD**: Template rendering with known data → expected prompt text. OpenAI client tested with mock HTTP server.
- **Acceptance**: Prompts render correctly. Client handles errors, rate limits, and timeouts gracefully.
- **Config**: OpenAI API key via `ai.api_key_env` in config. Model and reasoning effort from `ai.model` and `ai.reasoning.*` (see Phase 6.5).

#### 6.6 — AI Advisor Prompts (Prompt Templates)
- **Description**: Write the core prompt templates for the AI advisor. Each prompt includes UNM framework context, Team Topologies principles, and the relevant analysis data serialized as structured text. Prompts are iterable — no code changes needed to improve them.
- **Prompts to write**:
  1. `advisor/structural-load.tmpl` — Cognitive load reduction recommendations
  2. `advisor/service-placement.tmpl` — Where services should live based on value flow
  3. `advisor/team-boundary.tmpl` — Team split/merge/restructure recommendations
  4. `advisor/fragmentation.tmpl` — How to resolve fragmented capability ownership
  5. `advisor/bottleneck.tmpl` — How to reduce bottleneck risk
  6. `advisor/coupling.tmpl` — How to reduce coupling through data assets
  7. `advisor/interaction-mode.tmpl` — Interaction mode recommendations (collab vs xaas vs facilitating)
  8. `advisor/value-stream.tmpl` — Value stream coherence assessment and recomposition recommendations ("Team X is not a real value stream — recompose around needs Y, Z by moving services A, B")
  9. `advisor/need-delivery-risk.tmpl` — Need-level delivery chain risk analysis (team span, overloaded teams in chain)
  10. `advisor/general.tmpl` — Open-ended questions about the architecture
  11. `whatif/impact-assessment.tmpl` — Explain impact of a changeset in natural language
  12. `whatif/transition-plan.tmpl` — Suggest step-by-step transition from current to target state
  13. `query/natural-language.tmpl` — Answer factual questions about the model
  14. `query/model-summary.tmpl` — Produce an executive summary of the architecture
  15. `query/health-summary.tmpl` — Summarize value chain risks organized by UNM layer (consumes Signals API data)
- **Layer**: `infrastructure/ai/prompts/` (embedded via `//go:embed`)
- **TDD**: Each template renders with sample data without errors. Output contains expected framework context sections.
- **Acceptance**: All 15 templates written, renderable, and producing coherent AI prompts when filled with real model data.

#### 6.7 — AI Advisor REST API (Go)
- **Description**: API endpoint for AI-assisted analysis:
  - `POST /api/models/{id}/ask` — accepts `{ "question": "...", "category": "structural-load|service-placement|..." }`. Selects appropriate prompt template, renders with model + analysis data, calls OpenAI, returns structured response.
  - `POST /api/models/{id}/changesets/{csId}/explain` — AI explains the impact of a changeset in natural language using the impact report as context.
- **Layer**: `adapter/handler/ai.go`
- **TDD**: Mock OpenAI responses. Verify correct prompt template selection. Verify model/analysis data correctly injected. Verify graceful fallback when API key not configured.
- **Acceptance**: Questions routed to correct prompts. Responses include reasoning and actionable recommendations. Works without API key (returns "AI not configured" message).

#### 6.8 — What-If Explorer UI (React)
- **Description**: Frontend interface for creating and exploring changesets:
  - **Action palette**: buttons/forms for "Move service", "Split team", "Merge teams", "Reassign capability", "Change team size"
  - **Changeset sidebar**: shows list of pending actions, with remove/undo
  - **Impact panel**: side-by-side before/after for each analysis dimension (cognitive load, fragmentation, etc.), color-coded (green = improvement, red = regression, gray = no change)
  - **Export button**: download projected model as YAML
- **Layer**: `frontend/src/pages/WhatIfPage.tsx`, `frontend/src/components/changeset/`
- **Acceptance**: Users can compose changesets, see structural impact instantly, and export the result. No AI required — pure deterministic analysis.

#### 6.9 — AI Advisor Chat UI (React)
- **Description**: Chat-style interface for asking questions about the architecture:
  - Text input for natural language questions
  - Quick-action buttons for common questions ("How do I reduce cognitive load?", "Where should this service live?", "Summarize this architecture")
  - Response panel with markdown rendering for AI answers
  - Context indicator showing which model/changeset the AI is reasoning about
  - Optional: AI can propose changesets that the user can preview in the What-If Explorer
- **Layer**: `frontend/src/pages/AdvisorPage.tsx`, `frontend/src/components/advisor/`
- **Acceptance**: Users can ask questions and receive contextual recommendations. Common questions accessible via one click. AI-proposed changesets linkable to What-If Explorer.

#### 6.10 — AI Test Suite: 115 Questions & Expected Behaviors
- **Description**: A comprehensive test suite of 115 questions across all advisor categories (including value stream and need delivery), tested against the INCA model (or a purpose-built test model). Each test specifies:
  - The question
  - The category (structural-load, service-placement, etc.)
  - Expected behaviors: key phrases/entities that MUST appear in the answer, assertions that MUST NOT appear (e.g., recommending an action that makes things worse)
  - Whether the answer should reference specific teams, services, or capabilities from the model
- **Layer**: `backend/internal/infrastructure/ai/ai_test.go`, `testdata/ai_questions.json`
- **Test approach**: Run against a live OpenAI API in CI (with `UNM_OPENAI_API_KEY` set), or skip gracefully. Tests validate that answers contain expected entities and don't contradict the model data.
- **Categories and example questions** (see full list below).
- **Acceptance**: 115 tests defined. At least 85% pass rate against the current prompt library. Failures drive prompt improvements.

### AI Test Suite: 115 Questions

The questions are grouped by category. Each is tested against the INCA model where the AI receives the full model + analysis data including the Signals View output (Phase 4.10) for value chain context.

**Category: Structural Load (15 questions)**
1. "Which team has the highest structural load and why?"
2. "How can we reduce inca-core-dev's cognitive load?"
3. "What would happen if we moved inca-admin to a new team?"
4. "Is inca-core-dev's high interaction load a problem?"
5. "Which dimension contributes most to inca-publisher-dev's load?"
6. "Should we split inca-core-dev into two teams? How?"
7. "What is a healthy number of capabilities for a platform team?"
8. "Why is inca-observability-dev at medium instead of high?"
9. "If we add 3 more services to inca-serving, what happens to their load?"
10. "Which teams are at risk of becoming overloaded if the system grows?"
11. "Rank all teams by urgency of needing a restructure."
12. "What Team Topologies principle is inca-core-dev violating?"
13. "How many people should inca-ingestion-dev have to keep service load low?"
14. "Is inca-catalog-experience-dev's low load a sign they're underutilized?"
15. "Compare inca-core-dev's load profile to what Team Topologies recommends for a platform team."

**Category: Service Placement (15 questions)**
16. "Which services are owned by the wrong team?"
17. "Should inca-indexer be owned by inca-core-dev or catalog-quality-eng?"
18. "Where should inca-async-entity-writer live based on its dependency pattern?"
19. "Which services are candidates for extraction into a new team?"
20. "Does inca-serverless belong in inca-publisher-dev?"
21. "What services could be consolidated?"
22. "If we create a new 'inca-data-pipeline' team, which services should it own?"
23. "Which service has the most cross-team dependencies?"
24. "Are there services that only exist to coordinate between other services?"
25. "Which services are isolated enough to be owned by an external team?"
26. "Should inca-blob be its own platform service?"
27. "What services form a natural cluster that should be co-owned?"
28. "Which services are candidates for decommissioning?"
29. "If inca-core is split into read and write services, where should each go?"
30. "What value flow is inca-model-conversion-worker aligned to?"

**Category: Team Boundaries (10 questions)**
31. "How many teams should this system have ideally?"
32. "Which teams should be merged?"
33. "Is dotcom-eng a good team boundary for inca-ig-ingester?"
34. "Should inca-observability-dev be a platform team instead of complicated-subsystem?"
35. "What would a stream-aligned decomposition look like for this system?"
36. "Are there teams that should switch from x-as-a-service to collaboration?"
37. "Which team boundaries violate the inverse Conway maneuver?"
38. "If we reorganize around 4 teams instead of 9, what should they be?"
39. "Which teams have overlapping responsibilities?"
40. "Is catalog-quality-eng correctly classified as complicated-subsystem?"

**Category: Interaction Patterns (10 questions)**
41. "Are our interaction modes diverse enough?"
42. "Which interactions should change from x-as-a-service to collaboration?"
43. "Are there missing interactions that should exist?"
44. "Which team has the most interaction overhead?"
45. "Should inca-core-dev's self-interaction be a concern?"
46. "What facilitating relationships are missing?"
47. "Are there teams that never interact but should?"
48. "Which interaction patterns suggest tight coupling?"
49. "How would adding a facilitating relationship between X and Y help?"
50. "What does the interaction graph say about organizational silos?"

**Category: Fragmentation & Ownership (10 questions)**
51. "Which capabilities have fragmented ownership?"
52. "How should we resolve fragmentation in Data Transformation & Conversion?"
53. "Are there capabilities that should be split to reduce fragmentation?"
54. "Which team should be the single owner of Catalog Indexing?"
55. "Are there orphan services that support no capability?"
56. "Which capabilities are only realized by a single service (SPOF)?"
57. "Which needs have the weakest capability support?"
58. "Are there capabilities without any user need driving them?"
59. "Is our visibility hierarchy correct for all capabilities?"
60. "Which user-facing capabilities have the longest dependency chain?"

**Category: Dependency & Coupling (10 questions)**
61. "What is the critical dependency path in this system?"
62. "Which service is the biggest bottleneck?"
63. "How can we reduce inca-core's fan-in?"
64. "Are there circular dependencies?"
65. "Which data assets create hidden coupling?"
66. "What would break if inca-core went down?"
67. "Which services have the highest fan-out?"
68. "Are there unnecessary transitive dependencies?"
69. "Which dependencies are architectural vs runtime-only?"
70. "How can we decouple inca-publisher from inca-core?"

**Category: What-If Scenarios (15 questions)**
71. "What if we move inca-admin and inca-blob to a new 'inca-ops' team?"
72. "What if we merge inca-ingestion-dev and inca-publisher-dev?"
73. "What if inca-core-dev grows from 5 to 10 people?"
74. "What if we extract all Cadence workers into a dedicated team?"
75. "What if we split Catalog Entity Management into read-path and write-path capabilities?"
76. "What if dotcom-eng stops owning inca-ig-ingester?"
77. "What if we add a new 'inca-gateway' team for all serving?"
78. "What if inca-observability-dev takes over inca-regression-detector AND inca-realtime-stats as a platform?"
79. "What if we consolidate inca-blob into inca-core?"
80. "What if we introduce a facilitating relationship from inca-core-dev to help inca-ingestion-dev adopt entity patterns?"
81. "What if we double the team size of inca-core-dev?"
82. "What if inca-serving takes over inca-query from catalog-quality-eng?"
83. "What if we remove all self-loop interactions?"
84. "What if we create a shared 'data-pipeline' platform team?"
85. "What if we align teams strictly to the capability visibility hierarchy?"

**Category: Value Stream & Need Delivery (15 questions)**
86. "Which needs require the most teams to deliver?"
87. "Is inca-core-dev a real value stream or an organizational accident?"
88. "What needs form natural clusters that should define stream-aligned teams?"
89. "Which stream-aligned teams have low stream coherence?"
90. "How would you recompose teams around coherent value streams?"
91. "Which needs have no capability backing?"
92. "Which needs are at risk because their delivery chain passes through an overloaded team?"
93. "If we defined value streams around user needs, how many would we have?"
94. "Which user-facing needs have the longest delivery chain?"
95. "What would the team structure look like if every need was delivered by exactly one team?"
96. "Are there needs that share capabilities but are served by different teams?"
97. "Which needs would benefit most from a dedicated stream-aligned team?"
98. "What is the delivery risk for the 'Content Discovery & Search' need?"
99. "How do our value streams compare to what UNM recommends?"
100. "Show me the full delivery chain for each need with team assignments."

**Category: Strategic & Executive (15 questions)**
101. "Summarize this architecture for a VP of Engineering."
102. "What are the top 3 organizational risks in this system?"
103. "What would you prioritize restructuring first and why?"
104. "How well does this system follow Team Topologies principles?"
105. "What is the biggest impediment to fast flow in this organization?"
106. "If we could make only one change, what should it be?"
107. "How does this system's cognitive load compare to Team Topologies recommendations?"
108. "What is the blast radius if the core team loses a senior engineer?"
109. "Are we organized for innovation or stability?"
110. "What would a 12-month organizational evolution roadmap look like?"
111. "Which teams are best positioned to adopt platform-as-a-product?"
112. "What metrics should we track to measure organizational health?"
113. "How does our interaction pattern compare to successful Team Topologies case studies?"
114. "What Conway's Law implications does our current structure have?"
115. "Write an RFC proposing the single most impactful team restructure for this system."

### Parallelization

```
Wave 1 (parallel — no dependencies between items):
  Teammate "changeset":   6.1 + 6.2 (domain model + applier)
  Teammate "prompts":     6.5 + 6.6 (prompt infrastructure + templates)

Wave 2 (after Wave 1):
  Teammate "impact":      6.3 (impact analyzer — needs changeset applier)
  Teammate "ai-api":      6.7 (AI REST API — needs prompt library)

Wave 3 (after Wave 2):
  Teammate "changeset-api": 6.4 (changeset REST API — needs impact analyzer)

Wave 4 (parallel — frontend, after APIs exist):
  Teammate "whatif-ui":   6.8 (What-If Explorer)
  Teammate "advisor-ui":  6.9 (AI Advisor Chat)

Wave 5 (after everything):
  Teammate "ai-tests":    6.10 (100-question test suite — needs working AI API)
```

---

## Phase 6.5: Platform Configuration System (Go + React + DevOps)

**Goal**: Replace every hardcoded value, scattered environment variable, and magic constant with a unified, layered configuration system — inspired by .NET Core's `IConfiguration` pattern. Config loads once at startup with a deterministic precedence order. The rest of the code receives a typed struct and knows nothing about files, env vars, or loading logic.

**Why now**: A codebase audit found **40+ hardcoded config values** spread across Go files and React code: ports, timeouts, CORS origins, AI models, analysis thresholds, fan-in limits, team size defaults, feature flags, and more. Every one of these is a deployment risk. Before Phase 7+ adds more complexity and before any production deployment, the config story must be clean. This also enables per-prompt-category `reasoning.effort` for GPT-5.4 — one model everywhere, configurable depth per environment.

**Principle**: Configuration follows the .NET Core `IConfiguration` pattern adapted for Go:
1. **One load, once, at startup** — the config system resolves all sources into a single typed struct
2. **Deterministic precedence** — defaults → base file → env file → env vars → CLI flags (each overrides the previous)
3. **Zero config awareness in business code** — analyzers, handlers, and AI client receive typed config structs via constructor injection, they never call `os.Getenv()` or read files
4. **Secrets never in config files** — config files reference env var names (`api_key_env: "UNM_OPENAI_API_KEY"`), resolved at load time

**Library**: Use [`koanf`](https://github.com/knadh/koanf) — Go's cleanest layered config library. Lightweight (~1MB), supports YAML + env vars + CLI flags, merges sources in order, unmarshals directly into typed structs via `koanf:"tag"`. No global state (unlike Viper's singleton).

**Deliverable**: A `config/` directory with environment-specific YAML files, a koanf-based config loader, typed config structs injected into all components, frontend build-time config, and zero hardcoded values in application code.

### Architecture: Layered Config Loading (Like .NET Core)

```
Precedence (highest wins):

  5. CLI flags               --port=9090                    (only for CLI binary)
  4. Environment variables   UNM_SERVER_PORT=9090           (override anything)
  3. Environment config file config/production.yaml         (env-specific overrides)
  2. Base config file        config/base.yaml               (shared defaults)
  1. Code defaults           struct field defaults           (absolute fallback)

Loading happens ONCE at startup in cmd/server/main.go and cmd/cli/main.go.
The result is a single Config struct passed to all constructors.
No component ever calls os.Getenv() or reads a file — they receive typed config.
```

This is the same pattern as .NET Core's `WebApplicationBuilder.Configuration`:
- `base.yaml` = `appsettings.json`
- `{env}.yaml` = `appsettings.{Environment}.json`
- Env vars = `IConfiguration` env var provider
- CLI flags = command-line args provider

### Architecture: Config File Structure

```
config/
  base.yaml          ← shared defaults (always loaded first)
  local.yaml         ← development overrides (loaded when UNM_ENV=local or unset)
  production.yaml   ← production overrides (loaded when UNM_ENV=production)
  test.yaml         ← test overrides (loaded when UNM_ENV=test, AI disabled)
```

`base.yaml` contains ALL keys with sensible defaults. Environment files only override what differs. This eliminates duplication — production.yaml is ~15 lines, not a full copy.

### Full Config Schema: `base.yaml`

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  cors_origins: ["http://localhost:5173"]
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

frontend:
  api_base_url: "/api"
  dev_proxy_target: "http://localhost:8080"
  dev_server_port: 5173

ai:
  enabled: true
  provider: openai
  api_key_env: "UNM_OPENAI_API_KEY"
  base_url: "https://api.openai.com/v1"
  model: "gpt-5.4"
  request_timeout: 60s
  reasoning:
    default: "low"
    advisor: "medium"
    advisor_deep: "high"
    whatif: "high"
    query: "none"
    summary: "low"

analysis:
  default_team_size: 5
  cognitive_load:
    domain_spread_thresholds: [4, 6]    # [medium_at, high_at]
    service_load_thresholds: [2.0, 3.0] # services÷team_size
    interaction_load_thresholds: [4, 7] # weighted score
    dependency_load_thresholds: [5, 9]  # outbound deps
  interaction_weights:
    collaboration: 3
    facilitating: 2
    x-as-a-service: 1
  bottleneck:
    fan_in_warning: 5
    fan_in_critical: 10
  signals:
    need_team_span_warning: 2
    need_team_span_critical: 3
    high_span_service_threshold: 3      # caps per service before warning
    interaction_over_reliance: 4        # same-mode interactions before warning
    depth_chain_threshold: 4            # dependency chain depth
  value_chain:
    at_risk_team_span: 3                # needs with 3+ teams = at risk
  overloaded_capability_threshold: 6    # entity.Team.IsOverloaded()

features:
  debug_routes: true                    # POST /api/debug/load-example
  debug_example_paths:                  # candidate paths for load-example
    - "../examples/inca.unm.extended.yaml"
    - "../../examples/inca.unm.extended.yaml"
    - "examples/inca.unm.extended.yaml"

logging:
  level: "info"                         # debug, info, warn, error
  format: "text"                        # text, json
```

### Environment Override: `local.yaml`

```yaml
# local.yaml — development environment (default)
# Only overrides what differs from base.yaml. Most base defaults are already
# tuned for local development, so this file is minimal.

ai:
  reasoning:
    advisor: "medium"
    advisor_deep: "high"
    whatif: "high"

features:
  debug_routes: true
```

### Environment Override: `production.yaml`

```yaml
# production.yaml — production deployment
server:
  cors_origins: ["https://unm.internal.company.com"]
  read_timeout: 60s
  write_timeout: 60s
  shutdown_timeout: 30s

ai:
  request_timeout: 120s
  reasoning:
    default: "medium"
    advisor: "high"
    advisor_deep: "xhigh"
    whatif: "high"

features:
  debug_routes: false

logging:
  level: "warn"
  format: "json"
```

### Environment Override: `test.yaml`

```yaml
# test.yaml — fast test execution
ai:
  enabled: false

features:
  debug_routes: false

logging:
  level: "error"
  format: "text"
```

### Architecture: AI Reasoning Effort Selection

One model (`gpt-5.4`) everywhere. The only thing that changes between environments and prompt categories is `reasoning.effort` — how hard the model thinks before answering.

```
Prompt Template Category           Config Lookup               Reasoning Effort
──────────────────────────────────────────────────────────────────────────────────
                                                          local        production
advisor/structural-load.tmpl   →   ai.reasoning.advisor   →   medium       high
advisor/value-stream.tmpl      →   ai.reasoning.advisor   →   medium       high
advisor/team-boundary.tmpl     →   ai.reasoning.advisor_deep → high        xhigh
whatif/impact-assessment.tmpl  →   ai.reasoning.whatif     →   high         high
query/natural-language.tmpl    →   ai.reasoning.query      →   none         none
query/model-summary.tmpl       →   ai.reasoning.summary   →   low          low
(unknown category)             →   ai.reasoning.default    →   low          medium
```

Each prompt template declares its `category`. The AI handler looks up `ai.reasoning.{category}` from the loaded config, falls back to `ai.reasoning.default` if the category is not defined. The model is always `ai.model` — one model, configurable depth.

### Architecture: Env Var Override Convention

Any config key can be overridden by an environment variable using the `UNM_` prefix and `_` as separator:

```
server.port           → UNM_SERVER_PORT=9090
ai.model              → UNM_AI_MODEL=gpt-5.4
ai.reasoning.advisor  → UNM_AI_REASONING_ADVISOR=xhigh
logging.level         → UNM_LOGGING_LEVEL=debug
```

This is automatic via koanf's env provider with prefix `UNM_` and delimiter `_`. No code changes needed to support new overrides.

### Architecture: Clean Architecture Layers

```
domain/entity/config.go          → Config struct hierarchy (value object, validated)
infrastructure/config/loader.go  → LoadConfig(env string) → Config
                                    uses koanf: base.yaml → {env}.yaml → env vars
cmd/server/main.go               → cfg := config.LoadConfig(os.Getenv("UNM_ENV"))
                                    passes cfg.Server, cfg.AI, cfg.Analysis to constructors
cmd/cli/main.go                  → same LoadConfig, same Config struct
```

The `LoadConfig` function is the ONLY place that knows about files, env vars, and koanf. Everything downstream receives typed Go structs.

### Hardcoded Values Audit — What Moves to Config

Full audit of all hardcoded values currently in the codebase:

**Server & Network (5 values)**
| Current location | Hardcoded value | Config key |
|---|---|---|
| `cmd/server/main.go:23` | `"8080"` | `server.port` |
| `cmd/server/main.go:74` | `10*time.Second` | `server.shutdown_timeout` |
| `handler/middleware.go:21` | `"*"` (CORS origin) | `server.cors_origins` |
| `vite.config.ts:19` | `http://localhost:8080` | `frontend.dev_proxy_target` |
| `frontend/src/lib/api.ts:1` | `'/api'` | `frontend.api_base_url` |

**AI Client (4 values)**
| Current location | Hardcoded value | Config key |
|---|---|---|
| `openai_client.go:13` | `https://api.openai.com/v1` | `ai.base_url` |
| `openai_client.go:14` | `"gpt-4o"` | `ai.model` |
| `handler/ai.go:106` | `60*time.Second` | `ai.request_timeout` |
| `openai_client.go:43` | `"UNM_OPENAI_API_KEY"` | `ai.api_key_env` |

**Analysis Thresholds (18 values)**
| Current location | Hardcoded value | Config key |
|---|---|---|
| `cognitive_load.go:167-206` | `4, 6` (domain spread) | `analysis.cognitive_load.domain_spread_thresholds` |
| `cognitive_load.go:167-206` | `2.0, 3.0` (svc load) | `analysis.cognitive_load.service_load_thresholds` |
| `cognitive_load.go:167-206` | `4, 7` (interaction) | `analysis.cognitive_load.interaction_load_thresholds` |
| `cognitive_load.go:167-206` | `5, 9` (dep fan-out) | `analysis.cognitive_load.dependency_load_thresholds` |
| `cognitive_load.go:24-34` | `3, 2, 1` (interaction weights) | `analysis.interaction_weights.*` |
| `entity/team.go:44` | `5` (default team size) | `analysis.default_team_size` |
| `entity/team.go:71` | `> 6` (IsOverloaded) | `analysis.overloaded_capability_threshold` |
| `bottleneck.go:73-74` | `5, 10` (fan-in) | `analysis.bottleneck.fan_in_warning/critical` |
| `interaction_diversity.go:80` | `>= 4` (over-reliance) | `analysis.signals.interaction_over_reliance` |
| `signal_suggestions.go:15` | `4` (chain depth) | `analysis.signals.depth_chain_threshold` |
| `value_chain.go:98` | `> 2` (at-risk span) | `analysis.value_chain.at_risk_team_span` |
| `view_enriched.go:337,694` | `>= 3` (high-span svc) | `analysis.signals.high_span_service_threshold` |
| `signals.go:104` | `>= 3` (need team span) | `analysis.signals.need_team_span_critical` |

**Frontend Thresholds (duplicated from backend — 4 values)**
| Current location | Hardcoded value | Config key (served via API) |
|---|---|---|
| `CapabilityView.tsx:154` | `>= 3` | Should read from backend config API |
| `OwnershipView.tsx:219,486` | `>= 3` | Should read from backend config API |
| `RealizationView.tsx:96` | `>= 3` | Should read from backend config API |

**Features & Debug (3 values)**
| Current location | Hardcoded value | Config key |
|---|---|---|
| `handler/debug.go:22-26` | file paths | `features.debug_example_paths` |
| `handler/debug.go` | always registered | `features.debug_routes` |
| `UploadPage.tsx:24` | `600ms` delay | (UX constant, keep in code) |

### Backlog Items

#### 6.5.1 — Config Structs & Validation (Go)

- **Description**: Define the `Config` struct hierarchy with `koanf:"tag"` annotations for direct unmarshaling. This is a value object — no business logic, just shape + validation. Validation runs once after loading and fails fast with clear errors.
- **Structs**: `Config` (root), `ServerConfig`, `FrontendConfig`, `AIConfig` (with `Reasoning map[string]string`), `AnalysisConfig`, `CognitiveLoadConfig`, `InteractionWeightConfig`, `BottleneckConfig`, `SignalsConfig`, `ValueChainConfig`, `FeaturesConfig`, `LoggingConfig`.
- **Validation**: Port in 1-65535, reasoning effort in `{none,low,medium,high,xhigh}`, timeout > 0, cors_origins non-empty, fan_in_warning < fan_in_critical.
- **Layer**: `domain/entity/config.go`
- **TDD**: Valid config → passes. Invalid port → error with field path. Unknown reasoning effort → error listing valid values. Missing required field → error.
- **Acceptance**: All config structs defined with koanf tags and validation. `go vet` clean.

#### 6.5.2 — Config Loader with koanf (Go)

- **Description**: Single `LoadConfig(env string) (*Config, error)` function that implements the layered loading pattern:
  1. Load `config/base.yaml` (defaults)
  2. Load `config/{env}.yaml` (overrides, env defaults to `"local"`)
  3. Load environment variables with prefix `UNM_` and delimiter `_` (highest priority)
  4. Resolve secrets: read `cfg.AI.APIKeyEnv` env var name → actual API key
  5. Unmarshal into `Config` struct
  6. Run validation
  7. Return typed, validated `*Config`
- **Dependencies**: `github.com/knadh/koanf/v2`, `koanf/providers/file`, `koanf/providers/env`, `koanf/parsers/yaml`
- **Layer**: `infrastructure/config/loader.go`
- **TDD**: base.yaml only → full defaults. base + production → production overrides only changed keys. `UNM_SERVER_PORT=9090` → overrides file value. Missing base.yaml → clear error. Missing env file → warning, uses base only. `UNM_ENV=test` → AI disabled.
- **Acceptance**: Layered loading works with correct precedence. No secrets in config files. Clear errors for misconfiguration. Single function call, single return value.

#### 6.5.3 — Config Files: base, local, production, test

- **Description**: Create the config files:
  - `config/base.yaml` — complete config with ALL keys and sensible defaults (the full schema above)
  - `config/local.yaml` — minimal overrides for development (mostly empty, base is already tuned for local)
  - `config/production.yaml` — production overrides: restricted CORS, higher timeouts, higher reasoning effort, JSON logging, debug routes disabled
  - `config/test.yaml` — AI disabled, error-only logging
  - Update `.gitignore`: config files are committed (no secrets). `ai.env` stays ignored.
- **Layer**: `config/` directory at project root
- **Acceptance**: All four files valid YAML. `base.yaml` has every key. Env files only override what differs. Git-tracked.

#### 6.5.4 — Wire Config into Server Startup (Go)

- **Description**: Refactor `cmd/server/main.go` to load config ONCE at startup, then inject it into all components:
  ```go
  func main() {
      cfg, err := config.LoadConfig(os.Getenv("UNM_ENV"))
      // cfg.Server → http.Server
      // cfg.AI → ai.NewOpenAIClient(cfg.AI)
      // cfg.Analysis → analyzer constructors
      // cfg.Server.CORSOrigins → middleware
      // cfg.Features → debug route registration
      // cfg.Logging → log setup
  }
  ```
  Remove ALL `os.Getenv()` calls from main.go. Remove all hardcoded values. Components receive typed config via constructors — they never know about files or env vars.
- **Layer**: `cmd/server/main.go`, `adapter/handler/handler.go`, `adapter/handler/middleware.go`
- **Changes**:
  - `handler.New()` gains a `Config` parameter (or sub-configs)
  - Middleware reads CORS origins from `cfg.Server.CORSOrigins` (not hardcoded `"*"`)
  - Debug routes registered only when `cfg.Features.DebugRoutes == true`
  - Server timeout from `cfg.Server.ShutdownTimeout`
- **TDD**: Server starts with local config. Server starts with test config (AI disabled). Server with `features.debug_routes: false` → no debug endpoints.
- **Acceptance**: Zero `os.Getenv()` in main.go. Zero hardcoded values. `UNM_ENV=local go run ./cmd/server/` works with no other flags.

#### 6.5.5 — Wire Config into AI Client & Prompt Library (Go)

- **Description**: Refactor `OpenAIClient` and prompt library to receive config:
  - Remove `const defaultModel` and `const defaultBaseURL`
  - `NewOpenAIClient(cfg AIConfig)` — reads model, base_url, api_key from config
  - `Complete(ctx, systemPrompt, userMessage, reasoningEffort string)` — adds `reasoning.effort` to request body
  - Prompt library's `Render(category, data)` resolves `cfg.AI.Reasoning[category]` → effort, falls back to `cfg.AI.Reasoning["default"]`
  - `ai.request_timeout` from config → `context.WithTimeout` in handler
- **Layer**: `infrastructure/ai/openai_client.go`, `infrastructure/ai/prompt_library.go`, `adapter/handler/ai.go`
- **TDD**: Advisor category → request body has `reasoning.effort: "high"`. Query category → `"none"`. Unknown category → default. Mock HTTP server validates request body.
- **Acceptance**: Different prompt categories use different reasoning efforts from config. One model everywhere.

#### 6.5.6 — Wire Config into Analyzers (Go)

- **Description**: Refactor ALL analyzers to accept thresholds from config instead of hardcoded constants. Analyzer constructors gain a config parameter:
  - `NewCognitiveLoadAnalyzer(cfg CognitiveLoadConfig)` — thresholds from config
  - `NewBottleneckAnalyzer(cfg BottleneckConfig)` — fan-in thresholds from config
  - `NewSignalSuggestionGenerator(cfg SignalsConfig)` — depth chain, over-reliance thresholds
  - `NewValueChainAnalyzer(cfg ValueChainConfig)` — at-risk team span
  - `NewInteractionDiversityAnalyzer(cfg SignalsConfig)` — over-reliance count
  - `entity.Team.IsOverloaded()` → uses `cfg.Analysis.OverloadedCapabilityThreshold`
  - `InteractionWeight()` → reads from `cfg.Analysis.InteractionWeights`
  - Enriched view handlers: high-span service threshold from config
- **Layer**: All `infrastructure/analyzer/*.go` files, `domain/entity/team.go`, `adapter/handler/view_enriched.go`, `adapter/handler/signals.go`
- **TDD**: Each analyzer tested with custom thresholds. Changing a threshold in config → different analysis results. Default thresholds produce same results as current hardcoded behavior.
- **Acceptance**: Zero hardcoded thresholds in analyzer code. All thresholds configurable. Existing test results unchanged with default config values.

#### 6.5.7 — Wire Config into CLI (Go)

- **Description**: `cmd/cli/main.go` loads config the same way the server does — `config.LoadConfig(os.Getenv("UNM_ENV"))`. CLI flag `--env` can override `UNM_ENV`. CLI passes config to analyzers.
- **Layer**: `cmd/cli/main.go`
- **TDD**: CLI with default env → local config. CLI with `--env=production` → production thresholds. CLI respects `UNM_SERVER_PORT` override.
- **Acceptance**: CLI and server use identical config loading. No hardcoded thresholds in CLI.

#### 6.5.8 — Frontend Config (React + Vite)

- **Description**: Two-part frontend config:
  1. **Build-time config** via Vite `.env` files:
     - `frontend/.env` — defaults: `VITE_API_BASE_URL=/api`, `VITE_APP_TITLE=UNM Platform`, `VITE_ENVIRONMENT=local`
     - `frontend/.env.production` — overrides: `VITE_APP_TITLE=UNM Platform`, `VITE_ENVIRONMENT=production`
     - `frontend/src/lib/config.ts` — typed config object reading from `import.meta.env`
     - `vite.config.ts` — proxy target from `process.env.VITE_BACKEND_URL || 'http://localhost:8080'`
  2. **Runtime config API** — new `GET /api/config` endpoint that returns analysis thresholds the frontend needs to stay in sync with the backend (high-span threshold, need team-span threshold, etc.). Frontend fetches once on app init, stores in React context.
- **Layer**: `frontend/.env`, `frontend/.env.production`, `frontend/src/lib/config.ts`, `frontend/vite.config.ts`, `adapter/handler/config.go`
- **Changes**:
  - `api.ts`: `const API_BASE` reads from config
  - `CapabilityView.tsx`, `OwnershipView.tsx`, `RealizationView.tsx`: thresholds from runtime config (not hardcoded `>= 3`)
  - `index.html`, `Sidebar.tsx`: title from config
  - `UploadPage.tsx`: debug button shown only when `VITE_ENVIRONMENT !== 'production'`
- **TDD**: Build with production env → no debug controls. Dev build → debug controls visible. Frontend thresholds match backend config API response.
- **Acceptance**: Zero hardcoded URLs, ports, or thresholds in frontend code. Frontend thresholds always in sync with backend.

#### 6.5.9 — Config Documentation & README Update

- **Description**: Document the configuration system:
  - `docs/CONFIGURATION.md`:
    - Loading precedence (base → env file → env vars)
    - Full config schema reference with descriptions and defaults
    - How to add a new environment (create one YAML file)
    - How to add a new prompt category reasoning effort
    - Env var override convention (`UNM_` prefix, `_` separator)
    - Secret management (ai.env pattern)
    - Reasoning effort levels and cost implications
  - Update `README.md` quickstart to reference config
  - Update `ai.env.example` with all supported env vars
- **Layer**: Documentation
- **Acceptance**: New developer can set up the platform by following README. Adding a new environment = one YAML file, zero code changes.

### Parallelization

```
Wave 1 (parallel — no dependencies):
  Teammate "structs":       6.5.1 — Config structs + validation
  Teammate "files":         6.5.3 — Config YAML files (base, local, production, test)

Wave 2 (after Wave 1):
  Teammate "loader":        6.5.2 — koanf config loader

Wave 3 (parallel — wiring, all need the loader):
  Teammate "server":        6.5.4 — Wire into server startup + middleware
  Teammate "ai":            6.5.5 — Wire into AI client + prompt library
  Teammate "analyzers":     6.5.6 — Wire into all analyzers (biggest item)
  Teammate "cli":           6.5.7 — Wire into CLI

Wave 4 (after Wave 3 — needs backend config API from 6.5.4):
  Teammate "frontend":      6.5.8 — Frontend config (build-time + runtime)
  Teammate "docs":          6.5.9 — Documentation
```

---

## Phase 6.9: AI-Powered Per-Page Insights

**Goal**: Replace all hardcoded template explanation strings across every view page with real AI-generated insights, delivered via a single async batch call per page domain. Signals, cognitive load, needs, capabilities, ownership, topology, and dashboard pages all become truly intelligent — rows expand to show AI-reasoned plain-English explanations and concrete improvement suggestions.

**Why now**: Phase 6 added the AI advisor (ask-anything chat). Phase 6.9 makes AI ambient — every visual finding on every page silently enriches itself with contextual reasoning on load, without requiring the user to ask. This is the "AI intelligence everywhere" milestone.

**Deliverable**: `GET /api/models/{id}/insights/{domain}` endpoint returning structured JSON keyed by finding ID → all 7 view pages display real AI explanations and suggestions in expandable rows, falling back gracefully to template strings when AI is not configured or the call is in flight.

### Architecture

**Finding ID scheme** (stable string keys agreed by backend and frontend):
- Signals domain: `need-cross-team:{needId}`, `need-unbacked:{needId}`, `cap-fragmented:{capId}`, `cap-disconnected:{capId}`, `bottleneck:{serviceId}`, `team-load:{teamId}`, `team-coherence:{teamId}`
- Cognitive load domain: `team-load:{teamId}` (structural load score), `team-coherence:{teamId}` (coherence metric)
- Needs domain: `actor:{actorId}`, `need:{needId}` (cross-team risk, unbacked status)
- Capabilities domain: `cap:{capId}` (fragmentation, hierarchy issues)
- Ownership domain: `team:{teamId}` (boundary clarity, overload)
- Topology domain: `interaction:{teamA}:{teamB}` (interaction mode appropriateness)
- Dashboard domain: `summary` (overall architecture health narrative)

**Backend flow**:
1. Handler receives `GET /api/models/{id}/insights/{domain}`
2. Loads model from store
3. Builds domain-specific context using existing analyzers/view builders (same data the view endpoints return)
4. Renders prompt template from `prompts/insights/{domain}.tmpl`
5. Calls `h.aiClient.CompleteJSON(ctx, systemPrompt, userMessage)` which sets `response_format: {"type":"json_object"}`
6. Returns `{"domain": "signals", "insights": {"bottleneck:feed-api": {"explanation": "...", "suggestion": "..."}, ...}}`

**Frontend flow**:
1. `InsightsContext` caches insights by `modelId:domain` key — no re-calls on nav back
2. `usePageInsights(domain)` hook: fires on mount, checks cache first, sets loading state
3. Each view passes `insights[findingId]?.explanation ?? hardcodedFallback` to `ExpandableRow`

### Backlog Items

#### 6.9.1 — `CompleteJSON` method on OpenAI client

Add `ResponseFormat *ResponseFormatOption` field to `chatCompletionRequest`. Implement `CompleteJSON(ctx, system, user string) (string, error)` on `OpenAIClient` — identical to `Complete` but always sets `response_format: {"type": "json_object"}` and omits `reasoning_effort`. Add unit test using httptest server asserting the request body has `response_format.type == "json_object"`.

File: `backend/internal/infrastructure/ai/openai_client.go` + `openai_client_test.go`

#### 6.9.2 — Insight prompt templates (7 domains)

Create `backend/internal/infrastructure/ai/prompts/insights/` directory with one `.tmpl` file per domain. Each template receives the full domain context as Go template data and instructs the model to return a JSON object with finding IDs as keys and `{explanation, suggestion}` objects as values.

Templates:
- `signals.tmpl` — receives signals view data (needs, caps, teams, services with risk scores)
- `cognitive-load.tmpl` — receives teams with structural load scores and coherence metrics
- `needs.tmpl` — receives actor→need→capability chains, cross-team counts
- `capabilities.tmpl` — receives capability hierarchy, team ownership, fragmentation data
- `ownership.tmpl` — receives team→capability→service matrix, overload indicators
- `topology.tmpl` — receives team interactions, coupling metrics
- `dashboard.tmpl` — receives summary stats, health signals for overall narrative

Each prompt instructs: "Return ONLY valid JSON. Keys are finding IDs. Values are objects with 'explanation' (plain English, 1-2 sentences, no jargon) and 'suggestion' (concrete action, imperative, 1 sentence)."

File: `backend/internal/infrastructure/ai/prompts/insights/*.tmpl`

#### 6.9.3 — `GET /api/models/{id}/insights/{domain}` handler

New file `backend/internal/adapter/handler/insights.go`. Follows same pattern as `ai.go`:
- Parse `modelId` and `domain` from URL
- Load model from `h.store`
- Switch on domain to build domain-specific context struct (reuse existing analyzer/presenter calls)
- Look up prompt template from `h.promptLibrary`
- Call `h.aiClient.CompleteJSON(ctx, systemPrompt, contextJSON)`
- Parse and validate returned JSON
- Return `InsightsResponse{Domain: domain, Insights: map[string]InsightItem}` as JSON

If AI not configured: return `{"domain": "...", "insights": {}, "ai_configured": false}` with 200.
If model not found: 404.
If template missing: 500.

Register route in `cmd/server/main.go`: `router.Get("/api/models/{id}/insights/{domain}", h.handleGetInsights)`

File: `backend/internal/adapter/handler/insights.go`

#### 6.9.4 — Frontend API client + InsightsContext

Add to `frontend/src/lib/api.ts`:
```ts
export interface InsightItem { explanation: string; suggestion: string }
export interface InsightsResponse { domain: string; insights: Record<string, InsightItem>; ai_configured: boolean }
// api.getInsights(modelId, domain) → GET /api/models/{id}/insights/{domain}
```

Create `frontend/src/lib/InsightsContext.tsx`:
- Cache: `Map<string, InsightsResponse>` keyed by `${modelId}:${domain}`
- `getInsights(domain)`: async, checks cache → fetches if miss → stores in cache → returns
- Invalidates cache when `modelId` changes
- Exposes `{ getInsights, prefetch }` via context

Create `frontend/src/hooks/usePageInsights.ts`:
- Takes `domain: string`
- On mount: calls `getInsights(domain)`, sets `insights` state and `loading` state
- Returns `{ insights: Record<string, InsightItem>, loading: boolean }`

Files: `frontend/src/lib/api.ts`, `frontend/src/lib/InsightsContext.tsx`, `frontend/src/hooks/usePageInsights.ts`

#### 6.9.5 — Wire InsightsContext into AppShell

Wrap `<Outlet />` (or the whole app) with `<InsightsProvider>` in `AppShell.tsx`. Ensure context is available to all view pages.

File: `frontend/src/components/layout/AppShell.tsx`

#### 6.9.6 — Update SignalsView to use AI insights

Replace all hardcoded explanation/suggestion strings in `SignalsView.tsx` with:
```tsx
const { insights, loading: insightsLoading } = usePageInsights('signals')
// Per row:
explanation={insights[`bottleneck:${svc.service_id}`]?.explanation ?? hardcodedFallback}
suggestion={insights[`bottleneck:${svc.service_id}`]?.suggestion ?? hardcodedSuggestion}
```

Add subtle loading indicator (skeleton shimmer or faded text) on expandable rows while `insightsLoading`.

File: `frontend/src/pages/views/SignalsView.tsx`

#### 6.9.7 — Update CognitiveLoadView to use AI insights

Same pattern as 6.9.6. Domain: `cognitive-load`. Finding IDs: `team-load:{teamId}`, `team-coherence:{teamId}`.

File: `frontend/src/pages/views/CognitiveLoadView.tsx`

#### 6.9.8 — Update NeedView to use AI insights

Domain: `needs`. Finding IDs: `actor:{actorId}`, `need:{needId}`.

File: `frontend/src/pages/views/NeedView.tsx`

#### 6.9.9 — Update CapabilityView to use AI insights

Domain: `capabilities`. Finding IDs: `cap:{capId}`.

File: `frontend/src/pages/views/CapabilityView.tsx`

#### 6.9.10 — Update OwnershipView to use AI insights

Domain: `ownership`. Finding IDs: `team:{teamId}`.

File: `frontend/src/pages/views/OwnershipView.tsx`

#### 6.9.11 — Update TeamTopologyView to use AI insights

Domain: `topology`. Finding IDs: `interaction:{teamA}:{teamB}`.

File: `frontend/src/pages/views/TeamTopologyView.tsx`

#### 6.9.12 — Update DashboardPage to use AI insights

Domain: `dashboard`. Finding ID: `summary`. Replace static health narrative text with AI-generated overall assessment shown below the health cards.

File: `frontend/src/pages/DashboardPage.tsx`

### Parallelization Plan

```
Lead: Write Phase 6.9 backlog, then spawn team.

Wave 1 (parallel):
  Teammate "backend-insights": 6.9.1 + 6.9.2 + 6.9.3 — backend changes only
    Files: openai_client.go, openai_client_test.go, prompts/insights/*.tmpl,
           handler/insights.go, cmd/server/main.go

  Teammate "frontend-insights": 6.9.4 + 6.9.5 + 6.9.6 + 6.9.7 + 6.9.8 +
                                 6.9.9 + 6.9.10 + 6.9.11 + 6.9.12 — frontend only
    Files: api.ts, InsightsContext.tsx, usePageInsights.ts, AppShell.tsx,
           SignalsView.tsx, CognitiveLoadView.tsx, NeedView.tsx,
           CapabilityView.tsx, OwnershipView.tsx, TeamTopologyView.tsx, DashboardPage.tsx

No file conflicts — backend and frontend are completely separate.
```

---

## Phase 6.10: External Dependencies in Views + Platform Quality Hardening

**Goal**: Two-pronged phase: (A) make external dependencies visible across all views, and (B) fix every quality gap discovered during the firm E2E review of all completed phases. External dependencies like Cadence (used by 12 INCA services) are invisible single points of failure. The quality hardening items address real bugs, dead code, missing tests, and UX issues that undermine platform credibility.

**Why now**: External dependencies represent hidden concentration risk. The quality issues were discovered during a comprehensive review of every completed phase (1 through 6.9) and a full browser E2E test of every page, click, and interactive element. Fixing them now prevents compounding technical debt before Phase 7 (Transitions) begins.

**Deliverable**: External dependency nodes visible in UNM Map, Capability View, and Signals View with fan-in risk indicators. All quality findings resolved: model state persistence, dead code removal, error handling consistency, missing handler tests, frontend type safety, and impact analyzer config alignment.

### Part A: External Dependencies in Views

#### Key Findings from Research

- `ExternalDependency` entity has: `ID`, `Name`, `Description`, `UsedBy []ExternalUsage`
- Ownership View already shows external deps per team lane (partial implementation exists)
- Realization View already shows external deps per service as strings
- Bottleneck analyzer only measures service-to-service fan-in — completely ignores external dependency fan-in
- No signals exist for "Cadence is used by 12 services" — critical blind spot

#### Backlog Items

##### 6.10.1 — Extend bottleneck analyzer: external dependency fan-in

In `backend/internal/infrastructure/analyzer/bottleneck.go`, add detection of external dependency concentration:
- For each external dependency, count how many services use it (`len(dep.UsedBy)`)
- Flag as `IsCritical` when ≥ 5 services share the same external dependency (configurable threshold)
- Flag as `IsWarning` when ≥ 3 services

Add `ExternalDependencyBottlenecks []ExternalDepBottleneck` to `BottleneckResult`:
```go
type ExternalDepBottleneck struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    ServiceCount int     `json:"service_count"`
    Services    []string `json:"services"`
    IsCritical  bool     `json:"is_critical"`
}
```

##### 6.10.2 — Add external dependency signals to signals view

In `backend/internal/adapter/handler/signals.go` (or view builder), add a new field to the organization layer:
```go
type OrganizationLayer struct {
    // existing fields...
    CriticalExternalDeps []SignalsExtDepItem `json:"critical_external_deps"`
}
type SignalsExtDepItem struct {
    DepName      string   `json:"dep_name"`
    ServiceCount int      `json:"service_count"`
    Services     []string `json:"services"`
    IsCritical   bool     `json:"is_critical"`
}
```

Add to `api.ts` frontend type definitions.

##### 6.10.3 — Add external dependencies to Capability View backend

In `backend/internal/adapter/handler/view_enriched.go`, extend `buildEnrichedCapabilityView` to include external dependencies used by each capability's services:
- For each capability's services, look up `model.GetExternalDepsForService(serviceName)`
- Aggregate unique external deps across all services in the capability
- Add `ExternalDeps []string` (or `[]ExtDepRef`) to the capability view item

##### 6.10.4 — Add external dependency nodes to UNM Map backend

In `backend/internal/adapter/handler/view_enriched.go`, extend `buildUNMMapView` to include external dependency nodes:
- Add `ExternalDeps []UNMMapExtDep` to the UNM Map response
- Each includes: `id`, `name`, `description`, `service_count`, `services []string`
- Position them as "infrastructure" layer nodes (below services)

##### 6.10.5 — Render external dependencies in SignalsView

In `frontend/src/pages/views/SignalsView.tsx`, add a new section in the Organization layer card for critical external dependencies:
- Section title: "External Dependency Concentration"
- Each item: dep name, service count badge, list of dependent services (expandable)
- Color code: red if ≥ 5 services, amber if ≥ 3

##### 6.10.6 — Render external dependencies in CapabilityView

In `frontend/src/pages/views/CapabilityView.tsx`, add external dependencies to the capability detail panel:
- Show which external systems this capability's services rely on
- Include service count per external dep

##### 6.10.7 — Render external dependencies in UNMMapView

In `frontend/src/pages/views/UNMMapView.tsx`, render external dependency nodes at the bottom of the map:
- Separate row/section labeled "External Dependencies"
- Each node shows name + service count badge
- Fan-in ≥ 5: red badge; fan-in ≥ 3: amber badge; otherwise grey

### Part B: Platform Quality Hardening (E2E Review Findings)

The following findings come from a comprehensive review covering:
- **Full browser E2E test** of every page (Upload, Dashboard, Signals, UNM Map, Need, Capability, Ownership, Team Topology, Cognitive Load, Realization, What-If, Advisor)
- **Backend code quality audit** of all handlers, analyzers, AI integration, and middleware
- **Frontend code quality audit** of all views, API types, and state management
- **Test coverage gap analysis** across the full codebase

#### CRITICAL Findings

##### 6.10.8 — Model state persistence across page refresh (CRITICAL)

**Bug**: Model ID and parse result are stored only in React `useState` (`frontend/src/lib/model-context.tsx`). Any page refresh, direct URL navigation, or browser back/forward loses all data and redirects to Upload. Every view page checks `if (!modelId) navigate('/')`.

**Impact**: Users lose their entire session on any page refresh. This is the most impactful UX bug in the platform.

**Fix**: Persist `modelId` and `parseResult` to `localStorage` on `setModel`. On `ModelProvider` mount, check `localStorage` for a stored model ID, verify it still exists on the backend (`GET /api/models/{id}`), and restore the context. Clear on `clearModel`. This enables deep-linking, refresh survival, and browser back/forward.

**Files**: `frontend/src/lib/model-context.tsx`
**TDD**: Test that refreshing on `/dashboard` restores the model. Test that a stale localStorage ID (model deleted from backend) correctly redirects to Upload.

##### 6.10.9 — Insights endpoint returns HTTP 200 on internal errors (CRITICAL)

**Bug**: `backend/internal/adapter/handler/insights.go` (lines 141-198) returns HTTP 200 with `ai_configured: true` and empty `insights` for every failure: JSON marshal errors, template render errors, AI call errors, and AI response parse errors. The frontend cannot distinguish "no findings" from "AI is down" or "prompt failed."

**Impact**: Silent failures are impossible to debug. Users see empty insight rows and assume the AI found nothing, when in reality the AI may have failed.

**Fix**: Return distinguishable responses:
- AI errors: HTTP 200 with `{"ai_configured": true, "insights": {}, "error": "ai_unavailable"}` (transient)
- Template errors: HTTP 500 with `{"error": "insight template error"}`
- Marshal errors: HTTP 500 with `{"error": "internal error"}`
- Frontend should show a subtle "AI insights unavailable" message instead of silently hiding rows.

**Files**: `backend/internal/adapter/handler/insights.go`, `frontend/src/hooks/usePageInsights.ts`

##### 6.10.10 — Impact analyzer uses hardcoded default config for cognitive load (CRITICAL)

**Bug**: `backend/internal/infrastructure/analyzer/impact.go` line 95 creates a `NewCognitiveLoadAnalyzer` with `entity.DefaultConfig().Analysis` instead of the server's configured thresholds. If thresholds are customized (via config files), the What-If impact deltas will not match the live dashboard analysis.

**Impact**: What-If Explorer shows inconsistent cognitive load changes compared to the actual Cognitive Load View.

**Fix**: `ImpactAnalyzer` must accept and store the `AnalysisConfig` at construction time (injected in `cmd/server/main.go`). `diffCognitiveLoad` uses the injected config instead of `DefaultConfig()`.

**Files**: `backend/internal/infrastructure/analyzer/impact.go`, `backend/cmd/server/main.go`
**TDD**: Test that `ImpactAnalyzer` with custom thresholds produces different results than default thresholds.

#### MAJOR Findings

##### 6.10.11 — AI Advisor page not discoverable (no sidebar link)

**Bug**: The `/advisor` route exists in `App.tsx` and `AdvisorPage` renders correctly, but there is no entry in the sidebar `navItems` array (`frontend/src/components/layout/Sidebar.tsx`). Users can only reach it via the floating "Ask AI" button.

**Fix**: Add `{ to: '/advisor', label: 'AI Advisor', icon: Bot, always: false }` to the sidebar `navItems` array, positioned after What-If Explorer.

**Files**: `frontend/src/components/layout/Sidebar.tsx`, `frontend/src/App.tsx`

##### 6.10.12 — Dead legacy view builder functions (~300 lines)

**Bug**: `backend/internal/adapter/handler/view.go` lines 74-395 contain six `build*View` functions (`buildNeedView`, `buildCapabilityView`, `buildRealizationView`, `buildOwnershipView`, `buildTeamTopologyView`, `buildCognitiveLoadView`) that are never called. The `handleView` switch (lines 53-71) routes exclusively to `buildEnriched*View` functions from `view_enriched.go`. The old types (`viewNode`, `viewEdge`, `viewResponse`) are also unused.

**Impact**: ~300 lines of dead code confusing readers and inflating the codebase.

**Fix**: Delete the six unused functions and the three unused types. Keep `handleView` and `registerViewRoutes`.

**Files**: `backend/internal/adapter/handler/view.go`

##### 6.10.13 — Dashboard silently hides signals on API failure

**Bug**: `frontend/src/pages/DashboardPage.tsx` — when the signals fetch fails, `signals` is set to `null` and the entire "Health Signals" section is silently omitted. No error message is shown. Compare with `SignalsView.tsx` which properly shows an error state.

**Fix**: Show a fallback message like "Health signals unavailable" in the dashboard card when the fetch fails, matching the pattern used in other views.

**Files**: `frontend/src/pages/DashboardPage.tsx`

##### 6.10.14 — Frontend `api.ts` type definitions incomplete / mismatched

**Bug**: `frontend/src/lib/api.ts` defines `CapabilityViewResponse` (lines 138-155) that omits fields the `CapabilityView.tsx` component actually uses (`parent_groups`, `depended_on_by_count`, etc., defined as local interfaces in the page). Every view caller uses `as unknown as SpecificResponse` to cast `getView` results, bypassing TypeScript's type checking. This means API/schema drift is invisible.

**Fix**: Complete the shared type definitions in `api.ts` to match actual backend responses. Add typed view fetch helpers (e.g., `getNeedView(id)`, `getCapabilityView(id)`) that return the correct types without casting.

**Files**: `frontend/src/lib/api.ts`, all view pages

##### 6.10.15 — AI client ignores config-resolved API key

**Bug**: `backend/cmd/server/main.go` lines 30-31 check `cfg.AI.APIKey != ""` (resolved by config loader from `api_key_env`), but then calls `ai.NewOpenAIClient()` which independently reads `UNM_OPENAI_API_KEY` from the environment (hardcoded in `openai_client.go` line 44). If the config specifies a different env var name, the API key resolution is inconsistent.

**Fix**: Pass `cfg.AI.APIKey` to a new `ai.NewOpenAIClientWithKey(apiKey string)` constructor, or modify `NewOpenAIClient` to accept the resolved key. The client should not independently read environment variables.

**Files**: `backend/internal/infrastructure/ai/openai_client.go`, `backend/cmd/server/main.go`

##### 6.10.16 — No panic recovery middleware

**Bug**: `backend/internal/adapter/handler/middleware.go` has no panic recovery middleware. A panic in any handler will crash the entire server process.

**Fix**: Add a `recoveryMiddleware` that catches panics, logs the stack trace, and returns HTTP 500. Wire it as the outermost middleware in `NewRouter`.

**Files**: `backend/internal/adapter/handler/middleware.go`, `backend/internal/adapter/handler/router.go`

#### MINOR Findings

##### 6.10.17 — Missing handler tests for 5 endpoints

**Gap**: No dedicated test files exist for `health.go`, `signals.go`, `insights.go`, `middleware.go`, `debug.go`. These handlers have zero test coverage.

**Fix**: Add `health_test.go`, `signals_test.go`, `insights_test.go`, `middleware_test.go`, `debug_test.go` with basic happy-path and error-path tests.

**Files**: `backend/internal/adapter/handler/`

##### 6.10.18 — No frontend tests at all

**Gap**: Zero test files under `frontend/src/`. No `vitest`/`jest` dependency. No component or integration tests.

**Fix**: Add `vitest` to dev dependencies. Create at minimum:
- `model-context.test.tsx` — test state persistence and restoration
- `api.test.ts` — test error extraction and type helpers
- Smoke tests for at least 3 major views (Dashboard, Signals, Ownership)

**Files**: `frontend/package.json`, `frontend/src/`

##### 6.10.19 — Empty AI question validation

**Bug**: `backend/internal/adapter/handler/ai.go` accepts empty or whitespace-only questions and sends them to the AI model (wasted API cost, confusing UX).

**Fix**: Add `strings.TrimSpace(req.Question)` validation. Return 400 if empty.

**Files**: `backend/internal/adapter/handler/ai.go`

##### 6.10.20 — Inconsistent error handling in API client

**Bug**: Frontend API functions use two different error patterns:
- `parseModel`, `createChangeset`, `askAdvisor` use `extractError` to parse backend error bodies
- `getCapabilities`, `getTeams`, `getAnalysis`, `analyzeAll` use generic `Failed to fetch…` without body parsing

**Fix**: Standardize all API functions to use `extractError` for consistent, informative error messages.

**Files**: `frontend/src/lib/api.ts`

##### 6.10.21 — Config handler does not use shared `writeJSON`

**Bug**: `backend/internal/adapter/handler/config_handler.go` (lines 61-62) duplicates header setting + JSON encoding instead of using the shared `writeJSON` helper.

**Fix**: Replace with `writeJSON(w, http.StatusOK, resp)`.

**Files**: `backend/internal/adapter/handler/config_handler.go`

##### 6.10.22 — `@import` PostCSS warning in frontend build

**Bug**: Vite build shows `@import must precede all other statements` warning for the Google Fonts import in `index.css`. This is a CSS ordering issue.

**Fix**: Move the `@import url('...')` line to the very top of `frontend/src/index.css`, before any `@tailwind` directives or other rules.

**Files**: `frontend/src/index.css`

##### 6.10.23 — `http.Client` in OpenAI client has no default timeout

**Bug**: `backend/internal/infrastructure/ai/openai_client.go` creates `&http.Client{}` with no `Timeout` field. While individual requests use context deadlines, any call without a context deadline could hang indefinitely.

**Fix**: Set a default `Timeout` on the `http.Client` (e.g., 120 seconds as a safety net).

**Files**: `backend/internal/infrastructure/ai/openai_client.go`

### Parallelization Plan

```
Wave 1 (parallel — independent work):
  Teammate "backend-extdeps":     6.10.1 + 6.10.2 + 6.10.3 + 6.10.4
    Files: bottleneck.go, signals.go/view handler, view_enriched.go

  Teammate "frontend-extdeps":    6.10.5 + 6.10.6 + 6.10.7
    Files: SignalsView.tsx, CapabilityView.tsx, UNMMapView.tsx, api.ts (types only)

  Teammate "critical-bugs":       6.10.8 + 6.10.9 + 6.10.10
    Files: model-context.tsx, insights.go, impact.go, main.go

  Teammate "major-fixes":         6.10.11 + 6.10.12 + 6.10.13 + 6.10.14 + 6.10.15 + 6.10.16
    Files: Sidebar.tsx, view.go, DashboardPage.tsx, api.ts, openai_client.go, middleware.go

Wave 2 (after Wave 1 — polish):
  Teammate "tests-quality":       6.10.17 + 6.10.18 + 6.10.19 + 6.10.20 + 6.10.21 + 6.10.22 + 6.10.23
    Files: handler test files, frontend test setup, ai.go, api.ts, config_handler.go, index.css, openai_client.go
```

---

## Phase 6.12: Architecture Refactoring & Code Quality

**Goal**: Systematically address architectural debt, Clean Architecture violations, SOLID non-compliance, dead code, duplication, and type safety issues identified in the comprehensive platform review. This phase transforms the codebase from "working but messy" to "clean, maintainable, and principle-compliant."

**Why now**: The platform has grown through 6 phases of feature development. Each phase added value but also accumulated structural debt. The handler layer has become a god object (18 dependencies, ~32 methods). Business logic leaked into the adapter layer. Frontend views duplicate ~200 lines of boilerplate each. Fixing this before Phase 7 (Transitions) prevents the debt from compounding further.

**Full findings document**: [`docs/ARCHITECTURE_REVIEW.md`](./ARCHITECTURE_REVIEW.md) — contains the complete analysis with file paths, line numbers, dependency maps, and architecture health scoring.

**Architecture Health Score**: 7/10 — Clean domain boundaries but monolithic handler, missing use case services, and widespread frontend duplication.

### Backlog Items

#### 6.12.1 — Extract use case services from handler layer (SOLID: S, Clean Architecture)

The adapter/handler layer currently contains analysis orchestration, signal health classification, AI context building, and anti-pattern detection — all of which are application/domain logic.

**Extract these new use case / domain services:**

| New Service | From (handler file) | Lines Moved | Responsibility |
|---|---|---|---|
| `usecase/signals_service.go` | `signals.go` L85-279 | ~195 | Run 6 analyzers, merge results, classify health levels, apply threshold rules |
| `usecase/analysis_runner.go` | `analysis.go` L64-108 | ~45 | Dispatch to correct analyzer by type key, merge results |
| `usecase/ai_context_builder.go` | `ai.go` L297-425 | ~129 | Build AI prompt data from model + analyzer results |
| `usecase/changeset_explainer.go` | `ai.go` L182-241 | ~60 | Chain impact analysis → prompt render → AI call |
| `domain/service/anti_pattern_detector.go` | `view_enriched.go` L20-53 | ~34 | `detectTeamAntiPatterns`, `detectCapabilityAntiPatterns` etc. |

Handlers become thin: parse HTTP → call use case → write JSON.

**TDD**: Each new service gets its own `_test.go` testing behavior independently of HTTP.
**Files**: All handler files above + new use case / domain service files.

#### 6.12.2 — Registry pattern for analyzers and views (SOLID: O)

Replace closed `switch` statements with table-driven registries:

```go
type AnalyzerRegistry map[string]func(*entity.UNMModel) any
type ViewBuilderRegistry map[string]func(*entity.UNMModel, entity.AnalysisConfig) any
```

Adding a new analysis type or view type becomes **one line in a registry** instead of editing 4 files.

**Impact**: `validAnalysisType`, `runAnalysis`, `handleView` switches all collapse to map lookups.
**TDD**: Test that registry contains expected keys, test unknown key returns 400.
**Files**: `handler/analysis.go`, `handler/view.go`, `cmd/server/main.go`

#### 6.12.3 — `HandlerDeps` struct to replace 18-parameter constructor (KISS, ISP)

Replace:
```go
func New(cfg, pv, frag, cogLoad, dep, gap, bottle, coup, compl, interact, unlink, sigSug, valChain, valStr, csStore, impact, ai, store) *Handler
```

With:
```go
type HandlerDeps struct {
    Config          entity.Config
    ParseAndValidate *usecase.ParseAndValidate
    Analyzers       AnalyzerSet
    AI              AIClient      // interface
    Stores          StoreSet
}
```

**Benefit**: Adding a new dependency doesn't break every test file. Tests only construct what they need.
**TDD**: Existing handler tests continue to pass with the new struct.
**Files**: `handler/handler.go`, `cmd/server/main.go`, all handler test files

#### 6.12.4 — Inject `CognitiveLoadAnalyzer` into `ValueChainAnalyzer` (DIP, SRP)

Fix the architectural violation where `ValueChainAnalyzer.Analyze()` constructs a `CognitiveLoadAnalyzer` with `DefaultConfig()` internally.

**Before**: `value_chain.go` L48-50 — `NewCognitiveLoadAnalyzer(defaults.CognitiveLoad, ...)`
**After**: `ValueChainAnalyzer` accepts a `CognitiveLoadProvider` interface at construction. Server wires the same configured instance.

**TDD**: Test that custom thresholds propagate through value chain analysis.
**Files**: `infrastructure/analyzer/value_chain.go`, `cmd/server/main.go`

#### 6.12.5 — Delete backend dead code (~465 lines)

| What | Location | Lines |
|---|---|---|
| 6 legacy `build*View` functions + unused types | `handler/view.go` L74-395 | ~320 |
| `QueryEngine` (unused in production) | `usecase/query_engine.go` | ~145 |

**Decision needed**: Is `QueryEngine` intended for future use? If yes, document it and skip deletion. If no, remove it and its tests.

**TDD**: All existing tests continue to pass after deletion.
**Files**: `handler/view.go`, `usecase/query_engine.go`, `usecase/query_engine_test.go`

#### 6.12.6 — Delete frontend dead code (~200 lines)

| What | Location |
|---|---|
| `ViewPage.tsx` | Never imported, no route uses it |
| `types/model.ts` | Zero imports from any file |
| `lib/runtimeConfig.ts` | Zero references |
| 7 unused API methods in `api.ts` | `getCapabilities`, `getTeams`, `getNeeds`, `getServices`, `getActors`, `getAnalysis`, `analyzeAll` |
| Unused typed view methods | `getNeedView`, `getCapabilityView` (views use `getView` instead) |
| `CognitiveLoadAnalysis` interface | Not imported |
| `Badge` / `badgeVariants` from `components/ui/badge.tsx` | Not imported |

**TDD**: Frontend builds without errors after deletion.
**Files**: All above

#### 6.12.7 — Consolidate frontend duplication into shared modules

Create shared utilities to eliminate ~120 lines of duplication per view:

| New Module | Replaces | Copies Eliminated |
|---|---|---|
| `lib/slug.ts` | Inline `slug()` in 6 views | 6 |
| `lib/team-type-styles.ts` | `TEAM_TYPE_BADGE` in 3+ views + unused `types/model.ts` | 4 |
| `lib/visibility-styles.ts` | `VIS_BADGE` in 3 views | 3 |
| `hooks/useModelView.ts` | Duplicated fetch+loading+error+redirect in 8+ views | 8+ |
| Shared `ViewState` component | Duplicated loading/error fallback JSX | 8+ |
| Use shared `ExpandableRow` in SignalsView | Local duplicate in `SignalsView.tsx` L79+ | 1 |

**TDD**: All views render identically after refactor.
**Files**: New shared modules + all view files

#### 6.12.8 — Fix frontend type safety — eliminate `as unknown as` casts

Replace `api.getView()` → `as unknown as XResponse` pattern with properly typed view fetch methods:

```typescript
export async function getNeedView(modelId: string): Promise<NeedViewResponse> { ... }
export async function getCapabilityView(modelId: string): Promise<CapabilityViewResponse> { ... }
// etc.
```

Move all local view response interfaces into `api.ts` as the single source of truth.

**TDD**: TypeScript strict mode passes. Any backend API drift causes a compile error instead of a silent runtime issue.
**Files**: `lib/api.ts`, all 8 view files

#### 6.12.9 — Migrate inline hex colors to CSS design tokens

`index.css` already defines HSL design tokens (`--color-foreground`, `--color-muted-foreground`, `--color-border`, etc.) but views hardcode hex equivalents (`#111827`, `#6b7280`, `#e5e7eb`).

Map the ~100 hardcoded hex values to Tailwind semantic classes:
- `#111827` → `text-foreground`
- `#6b7280` → `text-muted-foreground`
- `#9ca3af` → `text-muted-foreground/70`
- `#e5e7eb` → `border-border`
- `#f3f4f6` → `bg-muted`
- `#f9fafb` → `bg-background`

Remove imperative `onMouseEnter`/`onMouseLeave` in `Sidebar.tsx` — replace with Tailwind `hover:` classes.

**Files**: All view files, `Sidebar.tsx`, `TopBar.tsx`

#### 6.12.10 — Extract oversized components (>300 lines)

| File | Extract Into |
|---|---|
| `OwnershipView.tsx` (~855 lines) | `TeamLane`, `ServicePopover`, `AntiPatternPanel`, `FilterBar` |
| `TeamTopologyView.tsx` (~745 lines) | `TopologyGraph`, `TopologyTable`, `TeamDetailPanel` |
| `UNMMapView.tsx` (~654 lines) | `MapNode`, `MapLegend`, `MapDetailPanel` |
| `CapabilityView.tsx` (~537 lines) | `CapabilityCard`, `CapabilityDetailPanel`, `GroupingSelector` |
| `SignalsView.tsx` (~449 lines) | `SignalSection`, `SignalRow` |
| `CognitiveLoadView.tsx` (~394 lines) | `LoadCard`, `LoadSidePanel`, `LoadLegend` |
| `AdvisorPanel.tsx` (~391 lines) | `ChatMessage`, `QuickActions`, `PageConfigResolver` |

Each extracted component gets its own file in a colocated directory (e.g., `pages/views/ownership/`).

**Files**: All oversized view files

#### 6.12.11 — Shared test helper factory for backend

Replace duplicated `newTestHandler()` in 5+ handler test files with a single factory:

```go
// internal/adapter/handler/testutil_test.go
func newTestHandler(opts ...TestOption) *Handler { ... }
```

Uses functional options so each test only specifies the dependencies it cares about. Defaults cover the rest.

**Benefit**: Adding a new analyzer to Handler doesn't break every test file.
**Files**: All handler `*_test.go` files, new `testutil_test.go`

#### 6.12.12 — Backend deduplication: coalesce helpers + cognitive load runs

1. **Generic coalesce**: Replace 5 identical `coalesceXItems` functions in `signals.go` with one generic:
   ```go
   func coalesce[T any](items []T, max int) []T
   ```

2. **Eliminate duplicate cognitive load analysis**: `handleSignals` currently runs cognitive load twice (once directly, once via `ValueChainAnalyzer`). After 6.12.4, the value chain analyzer reuses the injected instance.

3. **Merge `countHighLoadTeams` / `anyHighLoad`**: Same concept in `impact.go` and `signals.go` — unify into `analyzer` package.

**Files**: `signals.go`, `impact.go`, `value_chain.go`

### Parallelization Plan

```
Wave 1 (parallel — no file conflicts):
  Teammate "use-case-extraction":   6.12.1 + 6.12.4
    Files: New usecase/*.go, handler/signals.go, handler/ai.go, handler/analysis.go,
           infrastructure/analyzer/value_chain.go, domain/service/anti_pattern_detector.go

  Teammate "dead-code-cleanup":     6.12.5 + 6.12.6
    Files: handler/view.go, usecase/query_engine.go, frontend dead files

  Teammate "frontend-dedup":        6.12.7 + 6.12.8
    Files: New shared modules, all view files, api.ts

Wave 2 (after Wave 1 — depends on handler changes):
  Teammate "handler-refactor":      6.12.2 + 6.12.3 + 6.12.11
    Files: handler/handler.go, handler/analysis.go, handler/view.go,
           all handler test files, cmd/server/main.go

  Teammate "frontend-polish":       6.12.9 + 6.12.10
    Files: All view files, Sidebar.tsx, new component directories

Wave 3 (after Wave 2):
  Teammate "backend-dedup":         6.12.12
    Files: signals.go, impact.go, value_chain.go
```

---

## Phase 7: Transformation & Transition Planning

**Goal**: Support modeling current vs target organizational and architectural states, with step-by-step transition plans.

**Why seventh**: With model, analysis, visualization, and AI in place, transformation planning is the strategic capstone — using all prior capabilities to plan intentional evolution.

**Deliverable**: Transition modeling, delta analysis, and step-by-step migration planning through CLI, API, and frontend.

### Backlog Items

#### 7.1 — State Snapshots
- **Description**: Support named snapshots of the model (e.g., "current", "target-q3"). Stored as separate model files.
- **TDD**: Snapshot creation, loading, listing.
- **Acceptance**: Multiple snapshots coexist and are queryable.

#### 7.2 — Delta Analysis
- **Description**: Compare two model snapshots and produce a structured diff: added, removed, moved, merged, split entities and relationships.
- **TDD**: Known model pairs → expected deltas.
- **Acceptance**: All change types correctly detected.

#### 7.3 — Transition Step Modeling
- **Description**: Define transition steps with actions (move, merge, split, extract) and expected outcomes.
- **TDD**: Steps applied to model produce expected intermediate states.
- **Acceptance**: Step-by-step application transforms current to target.

#### 7.4 — Transition Validation
- **Description**: Validate that a transition plan actually transforms the current state into the target state without gaps.
- **TDD**: Valid and invalid transition plans → expected validation results.
- **Acceptance**: Incomplete or contradictory transitions are flagged.

#### 7.5 — Transition View (Frontend)
- **Description**: Side-by-side or overlay view of current vs target state with highlighted deltas. Step-through navigation.
- **Acceptance**: Transitions visually clear with step-by-step walkthrough.

#### 7.6 — Impact Analysis
- **Description**: For a proposed change (move capability, merge teams, split service), show downstream impacts across capability, service, team, and dependency dimensions.
- **TDD**: Known changes → expected impact cascades.
- **Acceptance**: Impact analysis comprehensive and accurate.

---

## Phase 8: Platform Maturity & Ecosystem

**Goal**: Production-grade features for real organizational adoption.

**Why last**: Core value is proven. Now harden for multi-user, multi-model, and CI/CD integration.

### Backlog Items

#### 8.1 — Model Persistence (Database)
- **Description**: Store models in SQLite (expandable to PostgreSQL). Multi-model, multi-snapshot support.
- **Acceptance**: Models persist across server restarts. Multiple models manageable.

#### 8.2 — CI/CD Integration
- **Description**: GitHub Action / CLI integration for validating `.unm` files in PRs, tracking model drift.
- **Acceptance**: PR check passes/fails based on model validation.

#### 8.3 — Export Formats
- **Description**: Export models to JSON, Mermaid, PlantUML, Structurizr DSL.
- **Acceptance**: Exported formats render correctly in their respective tools.

#### 8.4 — ADR & Documentation Links
- **Description**: Link model entities to ADRs, runbooks, documentation URLs.
- **Acceptance**: Entities have clickable documentation references in the frontend.

#### 8.5 — Metrics & SLA Integration
- **Description**: Attach operational metrics (latency, error rate, throughput) to services and capabilities.
- **Acceptance**: Metrics visible in service and capability views.

#### 8.6 — Multi-Model Federation
- **Description**: Federate multiple UNM models (from different teams/orgs) into an enterprise L1 view.
- **Acceptance**: Enterprise view composed from team-level models.

#### 8.7 — Plugin Architecture
- **Description**: Support custom analysis rules, view types, and data sources via plugins.
- **Acceptance**: Third-party plugins can extend major extension points.

#### 8.8 — Authentication & Multi-User
- **Description**: User authentication, model ownership, role-based access.
- **Acceptance**: Multiple users can own and share models.

---

## Cross-Cutting Concerns (All Phases)

### Testing
- **Backend**: Go standard `testing` package + testify assertions. Table-driven tests. 90%+ coverage target.
- **Frontend**: Vitest for unit tests. Playwright for E2E. Visual regression for rendered views (Phase 4+).
- Every phase maintains its coverage targets before moving to the next.

### Documentation
- Each phase updates the README with current capabilities.
- API documentation generated from handler code.
- Example models updated as features are added.

### CI
- All tests run on every commit.
- Lint + type-check on every commit.
- `go vet` and `golangci-lint` for backend.
- ESLint + TypeScript strict mode for frontend.

---

## Phase Dependency Graph

```
Phase 1 (Domain + YAML Parser)
  ↓
Phase 2 (Query + Basic Analysis)
  ↓
Phase 2.5 (Advanced Analysis — Bottleneck, Coupling, Complexity)
  ↓
Phase 3 (REST API Server)
  ↓
Phase 4 (Interactive Web Frontend)
  ↓
Phase 4.5 (View API Enrichment — Move Logic to Backend)
  ↓
Phase 4.6 (Interactive Cognitive Load & Team Topology Views)
  ↓
Phase 4.8 (Ownership View UX Redesign)
  ↓
Phase 4.9 (External Dependencies in Views)
  ↓
Phase 4.10 (UNM Value Chain Risks + Signals)  ← the AI input feed
  ↓
Phase 5 (Custom DSL Parser)                   ← can start after Phase 3
  ↓
Phase 6 (AI-Powered Interactive Platform)     ← requires Phase 4.10 (signals as AI input)
  6.1–6.4: Changeset Engine (Go, no AI)       ← pure backend, no AI dependency
  6.5–6.7: AI Infrastructure + API            ← needs OpenAI key, can start after 6.1
  6.8–6.9: Frontend (What-If + Advisor)       ← needs 6.4 + 6.7 APIs
  6.10: AI Test Suite                         ← needs 6.7 working
  ↓
Phase 6.10 (Ext Deps + Quality Hardening)    ← can start after Phase 6.9
  6.10.1–7: External deps in views            ← backend + frontend parallel
  6.10.8–10: Critical bugs (state, insights)  ← highest priority
  6.10.11–16: Major fixes (dead code, tests)  ← after critical bugs
  6.10.17–23: Minor polish                    ← last wave
  ↓
Phase 6.5 (Platform Configuration System)     ← can start after Phase 6 (AI client exists)
  6.5.1–6.5.3: Config model + files           ← no dependencies, can start anytime
  6.5.4–6.5.7: Wiring (server, AI, CLI, FE)   ← needs config loader
  6.5.8–6.5.9: Migration sweep + docs         ← needs wiring done
  ↓
Phase 6.12 (Architecture Refactoring)         ← can start after Phase 6.10 (quality hardening)
  6.12.1–6.12.4: Core architecture fixes      ← use case extraction, DIP, registry, HandlerDeps
  6.12.5–6.12.8: Dead code + type safety      ← cleanup, consolidation
  6.12.9–6.12.12: Polish + dedup              ← styling, component extraction, test helpers
  See: docs/ARCHITECTURE_REVIEW.md            ← full findings with scores and dependency maps
  ↓
Phase 7 (Transitions)                         ← benefits from Phase 6 changeset engine
  ↓
Phase 8 (Platform Maturity)                   ← requires all above
```

---

## Estimated Effort (Rough)

| Phase | Items | Estimated Duration |
|-------|-------|-------------------|
| Phase 1 — Domain + YAML + UNM Compliance | 8 + 1.9 (7 sub-items) + 1.10 | 3-4 weeks |
| Phase 2 — Basic Analysis | 6 | 2 weeks |
| Phase 2.5 — Advanced Analysis | 9 (2.7–2.15) | 2-3 weeks |
| Phase 3 — REST API | 6 | 1-2 weeks |
| Phase 4 — Frontend | 12 | 4-6 weeks |
| Phase 4.5 — View API Enrichment | 8 (4.5.1–4.5.8) | 1-2 weeks |
| Phase 4.8 — Ownership View UX Redesign | 5 (4.8.1–4.8.5) | 1-2 weeks |
| Phase 4.9 — External Dependencies in Views | 5 (4.9.1–4.9.5) | 1 week |
| Phase 4.10 — UNM Value Chain Risks + Signals | 7 (4.10.1–4.10.7) | 2-3 weeks |
| Phase 5 — Custom DSL | 9 | 3-4 weeks |
| Phase 6 — AI-Powered Interactive Platform | 10 (6.1–6.10) | 6-8 weeks |
| Phase 6.10 — Ext Deps + Quality Hardening | 23 (6.10.1–6.10.23) | 2-3 weeks |
| Phase 6.12 — Architecture Refactoring | 12 (6.12.1–6.12.12) | 3-4 weeks |
| Phase 6.5 — Platform Configuration System | 9 (6.5.1–6.5.9) | 2-3 weeks |
| Phase 7 — Transitions | 6 | 3-4 weeks |
| Phase 8 — Platform Maturity | 8 | 6-8 weeks |

**Total**: ~36-54 weeks for full platform.

**MVP (Phases 1-4.5)**: ~14-19 weeks to a working visual platform with YAML input, advanced analysis, signal suggestions, clean API-driven views, and proper UNM value chain rendering.

**Full Analysis MVP (Phases 1-4.10)**: ~17-24 weeks to a platform with complete UNM value chain risk surfacing, Signals View, and value stream coherence analysis. This is the data foundation the AI needs.

**Interactive MVP (Phases 1-6.4)**: ~25-34 weeks to a platform with what-if exploration (no AI required). Users can compose changesets and see structural impact.

**AI-Powered MVP (Phases 1-6.9)**: ~31-42 weeks to the full AI-powered interactive platform with advisor, what-if, and changeset management.
