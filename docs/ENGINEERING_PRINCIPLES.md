# Engineering Principles

## Core Philosophy

UNM Platform is built on the principle that **architecture models should be as rigorous as code**. We apply the same engineering discipline to building this tool that we expect it to bring to architecture practices.

**Stack**: Go backend (Clean Architecture) + React frontend (Vite + Tailwind CSS + shadcn/ui).

---

## 1. Test-Driven Development (TDD)

### The Cycle

1. **Red**: Write a failing test that describes the behavior you want
2. **Green**: Write the minimum code to make the test pass
3. **Refactor**: Clean up without changing behavior (tests stay green)

### Rules

- No production code exists without a corresponding test
- Tests are written *before* implementation, not after
- Each test tests one behavior, not one method
- Test files are co-located with source: `actor.go` → `actor_test.go`

### Backend (Go)

- Use Go standard `testing` package
- Table-driven tests for value objects and parsing
- Subtests with `t.Run()` for grouping related cases
- Use testify for assertions where it improves readability
- Test naming: `TestNewCapability/fragmentation_detection`
- Test data in `testdata/` directory

### Frontend (React/TypeScript)

- Use Vitest for unit and component tests
- React Testing Library for component behavior tests
- Playwright for E2E tests
- Test naming: describe blocks matching component/function names

### Test Layers

| Layer | What | How | Speed |
|-------|------|-----|-------|
| Unit | Domain entities, value objects, domain services | Pure functions, no I/O | < 1ms each |
| Integration | Use cases with real parsers, HTTP handlers | Real implementations | < 100ms each |
| E2E | Full parse → validate → API → render pipeline | HTTP + browser | < 10s each |

### Coverage Targets

- Domain layer: 95%+ (Go: `go test -cover`)
- Use cases: 90%+
- Handlers/adapters: 85%+
- Frontend components: 80%+

---

## 2. Clean Architecture

### Layer Rules (Backend)

```
Domain ← Use Cases ← Adapters ← Infrastructure
```

Dependencies point **inward only**. Inner layers never know about outer layers.

### Domain Layer (innermost)

- Pure Go, zero external dependencies (stdlib only)
- Contains: entities, value objects, domain services, repository interfaces
- Package: `internal/domain/entity`, `internal/domain/valueobject`, `internal/domain/service`
- Tested with pure unit tests

### Use Case Layer

- Orchestrates domain objects to fulfill application scenarios
- Depends only on domain layer (via interfaces)
- Each use case is a struct with one public method (e.g., `Execute`)
- Input/output through DTOs, not domain entities directly
- Package: `internal/usecase`

### Adapter Layer

- Adapts external input/output to use case format
- HTTP handlers, presenters, repository implementations
- Packages: `internal/adapter/handler`, `internal/adapter/presenter`, `internal/adapter/repository`

### Infrastructure Layer (outermost)

- Concrete implementations: YAML parser, DSL parser, file system, database
- Implements interfaces defined in inner layers
- Package: `internal/infrastructure/parser`, `internal/infrastructure/analyzer`

### Dependency Inversion

Inner layers define interfaces. Outer layers implement them.

```go
// Domain defines the contract
type ModelRepository interface {
    FindByID(ctx context.Context, id string) (*UNMModel, error)
    Save(ctx context.Context, model *UNMModel) error
}

// Infrastructure implements it
type FileModelRepository struct { /* ... */ }
func (r *FileModelRepository) FindByID(ctx context.Context, id string) (*UNMModel, error) { /* ... */ }
```

### Frontend Architecture

- Components are pure and receive data via props
- API calls happen in hooks (`useModel`, `useAnalysis`)
- Types match backend API contracts
- No business logic in components — delegate to hooks and utilities

---

## 3. SOLID Principles

### Single Responsibility (S)

Each package/struct has one reason to change.

- `parser` parses input into models — it does not validate semantics
- `validator` validates model constraints — it does not parse
- `handler` handles HTTP — it does not run analysis
- `analyzer` runs analysis — it does not format output

### Open/Closed (O)

Open for extension, closed for modification.

- New view types added by implementing `ViewProjection` interface, not by modifying existing views
- New analysis types added as new analyzer implementations
- New validation rules added as new `Rule` implementations

### Liskov Substitution (L)

Subtypes substitutable for base types.

- All parsers (YAML, DSL) implement `Parser` interface and produce the same `UNMModel`
- All analyzers implement `Analyzer` interface and produce typed reports

### Interface Segregation (I)

Clients depend only on interfaces they use.

- `ModelReader` vs `ModelWriter` interfaces (not one fat `ModelRepository`)
- `Parseable` vs `Validatable` vs `Analyzable`

### Dependency Inversion (D)

High-level modules don't depend on low-level modules.

- Use cases depend on `Parser` interface, not `YAMLParser` struct
- Handlers depend on use case interfaces, not concrete implementations

---

## 4. KISS (Keep It Simple)

### Guidelines

- Start with the simplest solution that meets the current requirement
- Do not build for hypothetical future needs (YAGNI)
- Prefer stdlib over external dependencies where reasonable
- If a function is longer than 30 lines, it probably does too much
- If a struct has more than 5 dependencies, it probably has too many responsibilities

### Applied to This Project

- Phase 1 uses YAML before building a custom DSL parser — adoption first
- Start with file-based model storage before adding a database
- Start with chi/stdlib router, not a framework
- Start with React Flow before building custom D3 renderers
- Frontend components use shadcn/ui patterns — don't reinvent

---

## 5. DRY (Don't Repeat Yourself)

### The Rule of Three

- First time: just write it
- Second time: note the duplication
- Third time: extract and abstract

### Caveat

Premature abstraction is worse than duplication. Two pieces of code that *look* the same but *change for different reasons* should stay separate.

---

## 6. Domain-Driven Design (DDD) Practices

### Ubiquitous Language

Use the UNM/Team Topologies vocabulary consistently across code, tests, docs, and conversations:

- "Capability", not "feature" or "module"
- "Need", not "requirement" or "user story"
- "Realizes", not "implements" (for service → capability)
- "Stream-aligned", not "product team" or "feature team"
- "Interaction mode", not "communication pattern"

### Aggregates

- `UNMModel` is the root aggregate — all entity access goes through it
- `Capability` contains sub-capabilities (composite pattern)
- `Team` has owned capabilities and interaction definitions

### Value Objects

- `EntityID` — identifier with validation
- `Confidence` — score (0.0–1.0) with evidence string
- `TeamType` — enum: stream-aligned, platform, enabling, complicated-subsystem
- `InteractionMode` — enum: collaboration, x-as-a-service, facilitating
- `MappingStatus` — enum: asserted, inferred, candidate, deprecated
- `Severity` — enum: low, medium, high, critical

---

## 7. Code Quality Standards

### Go Backend

- `go vet` must pass
- `golangci-lint` recommended
- No exported types without documentation comments
- Error handling: return errors, don't panic. Use `fmt.Errorf` with `%w` for wrapping.
- No `interface{}` / `any` in domain layer
- Table-driven tests for all value types
- Context propagation for all I/O operations

### React Frontend

- TypeScript strict mode
- No `any` types
- ESLint + Prettier
- Components: functional only, no class components
- State: React hooks, no external state management initially
- Styling: Tailwind utility classes only, no custom CSS files (except index.css for theme)

### Naming

- Go: PascalCase exported, camelCase unexported, packages lowercase
- TypeScript: PascalCase components, camelCase functions/hooks, SCREAMING_SNAKE constants
- Files: Go uses snake_case, TypeScript uses kebab-case for components
- Interfaces (Go): No `I` prefix. Name for the behavior: `Parser`, `Repository`, `Analyzer`

### Error Handling

- Domain errors are typed: `ValidationError`, `ParseError`, `ModelIntegrityError`
- Go: return `error` from all fallible functions, wrap with context
- Frontend: API errors mapped to user-friendly messages
- Never swallow errors silently

---

## 8. Git Workflow

### Branch Strategy

- `main` — always deployable
- `feat/phase-N/<description>` — feature branches per backlog item
- `fix/<description>` — bug fixes

### Commit Discipline

- Atomic commits: one logical change per commit
- Passing tests on every commit
- Conventional commits: `feat(domain): add Capability entity with decomposition support`

### PR Requirements

- All tests pass
- Coverage thresholds met
- No lint errors
- Clear description of what changed and why
