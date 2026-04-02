# Writing UNM Models in the DSL Format

This guide teaches you how to write `.unm` files — the recommended format for
User Needs Map models. The DSL is concise, git-friendly, and designed to be
read and written by humans.

For the full reference specification, see
[UNM_DSL_SPECIFICATION.md](UNM_DSL_SPECIFICATION.md). For the YAML format,
see [YAML_GUIDE.md](YAML_GUIDE.md).

---

## Why `.unm` over YAML?

| | `.unm` (DSL) | `.unm.yaml` (YAML) |
|---|---|---|
| **Readability** | Compact, minimal boilerplate | Verbose, indentation-heavy |
| **Git diffs** | Clean, line-oriented changes | Noisy diffs from indentation |
| **Typing speed** | Less punctuation to type | More colons, dashes, quotes |
| **Tooling** | Parsed by the UNM Platform | Parsed by the UNM Platform |
| **Export** | Supported (Export .unm button) | Supported (Export .yaml button) |

Both formats produce the same internal model. Use whichever you prefer.

---

## Minimal Working Example

The smallest valid `.unm` file:

```unm
system "Task Tracker" {
  description "Simple task management app"
}

actor "User" {
  description "Person managing their tasks"
}

need "Create and track tasks" {
  actor "User"
  outcome "I can create tasks and see their status at a glance"
  supportedBy "Task Management"
}

capability "Task Management" {
  description "Create, update, and organize tasks"
  visibility "user-facing"
}

service "task-api" {
  description "REST API for task CRUD operations"
  ownedBy "Product Team"
  realizes "Task Management"
}

team "Product Team" {
  type "stream-aligned"
  description "Owns the task management experience"
}
```

Save this as `tracker.unm` and parse it:

```bash
cd backend && go run ./cmd/cli/ parse ../tracker.unm
```

That's a complete model. The sections below show how to grow it.

---

## DSL Syntax Rules

1. **Blocks** use `keyword "name" { ... }` syntax
2. **Strings** are double-quoted: `"like this"`
3. **Bare keywords** (like `visibility`, `type`, `mode`) don't need quotes for
   simple values, but quoting is always valid
4. **Comments** use `//` or `#` — both work
5. **No commas** between fields — newlines separate them
6. **Whitespace** is flexible — indent however you like

---

## Section by Section

### `system`

Every `.unm` file starts with a system block defining the scope:

```unm
system "BookShelf" {
  description "Online bookstore platform for browsing, purchasing, and managing books"
}
```

The system block also stores version metadata (set automatically when you
edit models through the platform):

```unm
system "BookShelf" {
  description "Online bookstore platform"
  version "3"
  lastModified "2026-03-30T15:37:34Z"
  author "kristian"
}
```

---

### `actor`

Actors are people or systems with needs. Each gets its own block:

```unm
actor "Reader" {
  description "End user browsing and purchasing books"
}

actor "Author" {
  description "Book author managing listings and tracking sales"
}
```

---

### `need`

Needs capture what an actor wants to achieve. Every need links to at least one
capability via `supportedBy`:

```unm
need "Find and purchase books easily" {
  actor "Reader"
  outcome "I can search for books and complete a purchase in under a minute"
  supportedBy "Search & Discovery"
  supportedBy "Order Processing"
}
```

**Multiple actors** can share a need using comma-separated values:

```unm
need "View catalog data" {
  actor "Reader", "Author"
  outcome "Users can browse the full book catalog"
  supportedBy "Catalog Management"
}
```

---

### `capability`

Capabilities define what the system must be able to do. Each has a `visibility`
level that positions it in the UNM value chain:

| Visibility | Meaning |
|---|---|
| `user-facing` | Directly experienced by end users |
| `domain` | Core business processing, not directly visible |
| `foundational` | Internal capabilities underpinning domain logic |
| `infrastructure` | Deep infrastructure, fully invisible |

```unm
capability "Search & Discovery" {
  description "Help readers find books through search and recommendations"
  visibility "user-facing"
  dependsOn "Catalog Management"
}
```

Services declare which capabilities they realize — not the capability itself.

#### Capability dependencies

```unm
capability "Search & Discovery" {
  description "Help readers find books"
  visibility "user-facing"
  dependsOn "Catalog Management"
  dependsOn "Data Persistence" : "Needs reliable storage for search indexes"
}
```

#### Nested capabilities

Capabilities can nest to form a hierarchy. Only leaf capabilities (no children)
are realized by services:

```unm
capability "Catalog & Inventory" {
  description "Full lifecycle management of book catalog"
  visibility "domain"

  capability "Catalog CRUD" {
    description "Create, read, update, delete book records"
    visibility "foundational"
  }

  capability "Inventory Tracking" {
    description "Track stock levels across warehouses"
    visibility "domain"
  }
}
```

Alternatively, use flat `parent` references instead of nesting:

```unm
capability "Catalog & Inventory" {
  description "Full lifecycle management of book catalog"
  visibility "domain"
}

capability "Catalog CRUD" {
  description "Create, read, update, delete book records"
  visibility "foundational"
  parent "Catalog & Inventory"
}
```

**Visibility inheritance:** A child capability with no `visibility` inherits
from its parent. Set `visibility` explicitly on a child to override.

---

### `service`

Services are concrete implementations. They declare:
- `ownedBy` — which team owns this service
- `realizes` — which capabilities this service implements
- `externalDeps` — which external systems this service depends on
- `dependsOn` — which other services this service calls

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

service "order-service" {
  description "Manages cart, checkout, and payment"
  ownedBy "Fulfillment"
  realizes "Order Processing"
  externalDeps "Payment Gateway"
}
```

#### `realizes` — relationship forms

Three forms are supported and can be mixed freely:

```unm
// Simple — just the capability name
realizes "Catalog CRUD"

// With description — colon shorthand
realizes "Catalog Search" : "Full-text search implementation"

// With role and description — block form
realizes "Catalog Search" {
  role "primary"
  description "Full-text search implementation"
}
```

Roles:
- `primary` — main implementation
- `supporting` — contributes to but does not own
- `consuming` — uses the capability as a client

#### `dependsOn` — relationship forms

The same three forms apply:

```unm
dependsOn "catalog-api"
dependsOn "catalog-api" : "Indexes book data"
dependsOn "catalog-api" {
  description "Indexes book data"
}
```

---

### `team`

Teams use Team Topologies types to classify their purpose:

**Types:** `stream-aligned`, `platform`, `enabling`, `complicated-subsystem`

```unm
team "Storefront" {
  type "stream-aligned"
  description "Owns the catalog browsing experience"
  size 6
  owns "Catalog Management"
  owns "Search & Discovery"
}

team "Platform" {
  type "platform"
  description "Provides shared infrastructure services"
  size 3
  owns "Notification Delivery"
}
```

#### Inline interactions

Teams declare how they interact with other teams using `interacts` — the
primary authoring path for team interactions:

```unm
team "Discovery" {
  type "stream-aligned"
  description "Owns search and recommendations"
  size 4
  owns "Search & Discovery"
  interacts "Storefront" mode "x-as-a-service" via "Catalog CRUD" description "Consumes catalog APIs"
}
```

**Interaction modes:** `x-as-a-service`, `collaboration`, `facilitating`

---

### `interaction` (standalone form — DSL only)

Top-level interaction blocks describe how teams work together using arrow syntax.
This is ergonomic for expressing interactions at the file level rather than
inside team blocks:

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

> Note: Standalone `interaction` blocks exist only in the DSL format. In YAML,
> use `team.interacts` instead.

---

### `platform`

Group platform teams that together provide shared capabilities:

```unm
platform "Infrastructure Platform" {
  description "Shared infrastructure services and tooling"
  teams ["Platform", "SRE"]
}
```

---

### `data_asset`

Model shared storage and messaging infrastructure to reveal implicit coupling:

**Types:** `database`, `cache`, `event-stream`, `blob-storage`, `search-index`

```unm
data_asset "books_db" {
  type "database"
  description "Primary relational database storing the book catalog"
  usedBy "catalog-api"
  usedBy "search-service"
}

data_asset "book_change_events" {
  type "event-stream"
  description "Events emitted when book records change"
  usedBy "search-service"
  usedBy "recommendation-engine"
}
```

---

### `external_dependency`

Systems outside the modeled boundary. External dependencies are
**definition-only** — you declare their name and description here. Services
declare which external systems they use via `externalDeps`.

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

Services declare their use:

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

---

## Advanced Features

### Transition Modeling

Plan organizational changes with before/after states and concrete steps:

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
}
```

### Imports

Compose models from multiple files:

```unm
import "catalog.unm"
import authors from "authors.unm"
```

---

## Relationship Forms — Quick Reference

All list-type relationships (`supportedBy`, `realizes`, `dependsOn`) support
three forms that can be mixed freely:

```unm
// 1. Simple — just the name
realizes "Catalog CRUD"

// 2. Colon shorthand — adds a description
realizes "Catalog CRUD" : "Primary CRUD implementation"

// 3. Block form — adds role and description
realizes "Catalog CRUD" {
  role "primary"
  description "Primary CRUD implementation"
}
```

---

## Common Mistakes

**1. Putting `realizes`/`realizedBy` on a capability**

Services declare which capabilities they realize — not the capability itself.

```unm
// Wrong — realizedBy is not a capability field in v2
capability "Search & Discovery" {
  visibility "user-facing"
  realizedBy "search-service"    // ERROR: removed in v2
}

// Correct — declare on the service
service "search-service" {
  ownedBy "Discovery"
  realizes "Search & Discovery"
}
```

**2. Missing `supportedBy` on a need**

Every need must link to at least one capability:

```unm
// Wrong — no supportedBy
need "Buy a book" {
  actor "Reader"
  outcome "I purchased a book"
}

// Correct
need "Buy a book" {
  actor "Reader"
  outcome "I purchased a book"
  supportedBy "Order Processing"
}
```

**3. Forgetting quotes on multi-word names**

```unm
// Wrong — unquoted multi-word name
actor Search & Discovery { ... }

// Correct
actor "Search & Discovery" { ... }
```

**4. Declaring external dependency edges on the dependency**

External dependency definitions do not carry usage edges. Services declare
which external systems they use.

```unm
// Wrong — usedBy is not valid on external_dependency in v2
external_dependency "Elasticsearch" {
  description "Search engine"
  usedBy "search-service"    // ERROR: removed in v2
}

// Correct — declare on the service
service "search-service" {
  ownedBy "Discovery"
  externalDeps "Elasticsearch"
}
```

---

## Exporting and Versioning

The UNM Platform supports exporting models in both `.unm` and `.yaml` formats.
Use the **Export .unm** button on the Edit Model page to download the current
model as a `.unm` file for git versioning.

The platform automatically tracks version metadata:
- **version**: incremented on every model change
- **lastModified**: UTC timestamp of the last change

These are stored in the `system` block and exported with the file.

---

## Full Example

See [`examples/bookshelf.unm`](../examples/bookshelf.unm) for a complete model
demonstrating all DSL format features, including capabilities, services with
`realizes` and `externalDeps`, teams with inline interactions, and external
dependencies.

Parse it yourself:

```bash
cd backend && go run ./cmd/cli/ parse ../examples/bookshelf.unm
```

For a large real-world example in YAML format, see
[`examples/nexus.unm.yaml`](../examples/nexus.unm.yaml).
