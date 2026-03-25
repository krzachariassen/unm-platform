# Clean Architecture Rules

## Layer Dependencies (inward only)

```
Infrastructure → Adapter → UseCase → Domain
```

Domain MUST NOT import from any outer layer.
Interfaces are defined where they are USED, not where they are implemented.

## Package Rules

- **domain/entity/**: One file per entity (actor.go, capability.go, etc.)
- **domain/valueobject/**: Enums and typed constants (TeamType, InteractionMode)
- **domain/service/**: Domain logic that spans entities (validator, changeset applier)
- **usecase/**: Application orchestration. May call domain services and infrastructure via interfaces.
- **adapter/handler/**: HTTP handlers. Thin — delegate to use cases.
- **adapter/presenter/**: Transform domain models into view-specific JSON.
- **adapter/repository/**: In-memory stores implementing domain interfaces.
- **infrastructure/**: Parser, analyzers, serializers, AI integration.

## Violations That Must Be Caught

- Importing `internal/adapter/` from `internal/domain/`
- Importing `internal/infrastructure/` from `internal/usecase/`
- Business logic in HTTP handlers (handler should call usecase, not compute)
- Domain entities with JSON tags (JSON is an adapter concern)
