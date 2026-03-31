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
  realizedBy "task-api"
}

service "task-api" {
  description "REST API for task CRUD operations"
  ownedBy "Product Team"
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
  realizedBy "search-service" : "Full-text search across title and author"
  realizedBy "recommendation-engine" : "Personalized suggestions"
  dependsOn "Catalog Management"
}
```

#### Relationships on capabilities

**`realizedBy`** links a capability to the services that implement it.
Three forms are supported:

```unm
// Simple — just the service name
realizedBy "search-service"

// With description — colon shorthand
realizedBy "search-service" : "Full-text search implementation"

// With role and description — block form
realizedBy "search-service" {
  role "primary"
  description "Main search implementation"
}
```

**`dependsOn`** links to other capabilities this one requires:

```unm
dependsOn "Catalog Management"
dependsOn "Data Persistence" : "Needs reliable storage for search indexes"
```

#### Nested capabilities

Capabilities can nest to form a hierarchy. Only leaf capabilities (no children)
have `realizedBy`:

```unm
capability "Catalog & Inventory" {
  description "Full lifecycle management of book catalog"
  visibility "domain"

  capability "Catalog CRUD" {
    description "Create, read, update, delete book records"
    visibility "foundational"
    realizedBy "catalog-api"
  }

  capability "Inventory Tracking" {
    description "Track stock levels across warehouses"
    visibility "domain"
    realizedBy "inventory-service"
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
  realizedBy "catalog-api"
}
```

---

### `service`

Services are concrete implementations. They declare an owner and
service-to-service dependencies:

```unm
service "catalog-api" {
  description "Authoritative book catalog CRUD service"
  ownedBy "Storefront"
  dependsOn "order-service" : "Validates stock on order placement"
}
```

The `dependsOn` field supports the same three forms as capability relationships
(simple, colon shorthand, and block form).

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

---

### `interaction`

Top-level interaction blocks describe how teams work together:

**Modes:** `x-as-a-service`, `collaboration`, `facilitating`

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

Systems outside the modeled boundary. The `usedBy` field supports an optional
description using colon syntax:

```unm
external_dependency "Payment Gateway" {
  description "Third-party payment processing provider"
  usedBy "order-service" : "Processes credit card and wallet payments"
}

external_dependency "Email Provider" {
  description "Transactional email delivery service"
  usedBy "notification-service" : "Sends order confirmations"
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

All list-type relationships (`supportedBy`, `realizedBy`, `dependsOn`) support
three forms that can be mixed freely:

```unm
// 1. Simple — just the name
realizedBy "catalog-api"

// 2. Colon shorthand — adds a description
realizedBy "catalog-api" : "Primary CRUD implementation"

// 3. Block form — adds role and description
realizedBy "catalog-api" {
  role "primary"
  description "Primary CRUD implementation"
}
```

---

## Common Mistakes

**1. Missing `supportedBy` on a need**

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

**2. Putting `realizedBy` on a parent capability**

Only leaf capabilities (no nested children) have `realizedBy`:

```unm
// Wrong — parent has realizedBy
capability "Catalog" {
  visibility "domain"
  realizedBy "catalog-api"

  capability "Search" {
    visibility "user-facing"
    realizedBy "search-service"
  }
}

// Correct — only the leaf has realizedBy
capability "Catalog" {
  visibility "domain"

  capability "Search" {
    visibility "user-facing"
    realizedBy "search-service"
  }
}
```

**3. Forgetting quotes on multi-word names**

```unm
// Wrong — unquoted multi-word name
actor Search & Discovery { ... }

// Correct
actor "Search & Discovery" { ... }
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
with 3 actors, 4 needs, 6 capabilities, 6 services, 4 teams, data assets, and
external dependencies.

Parse it yourself:

```bash
cd backend && go run ./cmd/cli/ parse ../examples/bookshelf.unm
```
