# UNM Domain Model

## Core Entities

| Entity | Description |
|--------|-------------|
| Actor | Person or system with needs (Merchant, Eater, Operator) |
| Need | What an actor is trying to achieve, linked to one actor |
| Capability | What the system must do to support a need. Has `visibility`: user-facing, domain, foundational, infrastructure |
| Service | Concrete implementation realizing capabilities. Declares owner and service-to-service dependencies |
| Team | Organizational unit owning services. Has Team Topologies type: stream-aligned, platform, enabling, complicated-subsystem |
| Platform | Grouping of platform teams |
| Interaction | How two teams work together: collaboration, x-as-a-service, facilitating |
| Data Asset | Storage/messaging infrastructure (database, cache, event-stream). Declares which services use it (source of truth). |
| External Dependency | System outside the modeled boundary. Declares which services depend on it (source of truth). |
| Signal | Categorized architectural finding with severity and evidence |

## Value Chain (top-down)

```
Actor → hasNeed → Need → supportedBy → Capability → realizedBy → Service → ownedBy → Team
```

Additional relationships:
- capability → dependsOn → capability
- capability → decomposesTo → capability (hierarchy via children)
- service → dependsOn → service
- team → interactsWith → team (with mode)
- data_asset → usedBy → service
- external_dependency → usedBy → service

## Unidirectional Principle

Relationships are declared in ONE direction only. The parser derives reverse
lookups at query time. Services do NOT declare `supports` — capabilities
declare `realizedBy`. Data assets and external dependencies declare `usedBy`.

## YAML Relationship Forms

- Short: `"service-name"`
- Long: `{target: "service-name", description: "Calls feed-api for fetching"}`

Both can be mixed in the same list.
