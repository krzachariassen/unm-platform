# Architecture, Principles & Code Quality Review

**Date**: 2026-03-20  
**Scope**: Full codebase review covering Go backend + React frontend  
**Methodology**: Static analysis of all packages, dependency mapping, SOLID/KISS/Clean Architecture compliance, dead code detection, duplication analysis, test coverage audit

---

## 1. Architecture Health — Score: 7/10

### What's Working Well

- **Domain purity**: `domain/entity`, `domain/service`, `domain/valueobject` have zero imports from adapter or infrastructure layers. This is the hardest rule to maintain and it's clean.
- **Use case ports**: `usecase.ParseAndValidate` defines `ModelParser` and `ModelValidator` interfaces — proper inward-facing ports.
- **Composition root**: `cmd/server/main.go` is the single place where all dependencies are wired. Explicit and traceable.
- **Test coverage**: 76 test files for 65 source files. Domain entities, analyzers, parsers, AI, and config all have dedicated tests.
- **Clean import graph**: No forbidden edges. Domain never imports infrastructure. Use cases never import adapters.

### What Needs Improvement

- **Monolithic Handler**: 18 dependencies, ~32 receiver methods across all handler files. Every new feature adds another constructor parameter.
- **Business logic in adapter layer**: `view_enriched.go`, `signals.go`, and `ai.go` contain analysis orchestration, anti-pattern detection, and health classification that belong in use cases or domain services.
- **No domain ports for analyzers**: All 12 analyzers are concrete types passed directly to Handler. No interfaces, no dependency inversion beyond parse/validate.
- **Presenter layer missing**: `CLAUDE.md` describes a presenter layer but it doesn't exist as a package. Presentation logic lives inside handlers.

---

## 2. Clean Architecture Violations

### 2.1 — Handler as Application Service (MAJOR)

The handler layer is doing the work of missing use case / application services:

| Handler File | Misplaced Logic | Should Be |
|---|---|---|
| `signals.go` L85-279 | Runs 6 analyzers, merges results, classifies health levels, applies threshold rules | `usecase/signals_service.go` |
| `view_enriched.go` L20-53 | `detectTeamAntiPatterns`, `detectCapabilityAntiPatterns` — domain policy decisions | `domain/service/anti_pattern_detector.go` |
| `view_enriched.go` L154-161 | Constructs `ValueChainAnalyzer` inside view builder, runs analysis inline | Should use injected analyzer |
| `ai.go` L297-425 | `buildAIPromptData` runs `cognitiveLoad.Analyze(m)` and `valueChain.Analyze(m)` — orchestration | `usecase/ai_context_builder.go` |
| `ai.go` L182-241 | `handleExplainChangeset` — chains impact analysis + prompt render + AI call | `usecase/changeset_explainer.go` |
| `analysis.go` L64-108 | `runAnalysis` switch dispatches to analyzers and merges results | `usecase/analysis_runner.go` |

**Impact**: Handler files are 800-1200 lines, hard to test in isolation, and every new analysis dimension requires touching the adapter layer.

### 2.2 — Missing Domain Interfaces / Ports (MAJOR)

The domain layer defines **zero interfaces**. All analysis capabilities are wired as concrete infrastructure types:

```
Handler
  ├── *analyzer.FragmentationAnalyzer       (concrete)
  ├── *analyzer.CognitiveLoadAnalyzer       (concrete)
  ├── *analyzer.DependencyAnalyzer          (concrete)
  ├── *analyzer.GapAnalyzer                 (concrete)
  ├── *analyzer.BottleneckAnalyzer          (concrete)
  ├── *analyzer.CouplingAnalyzer            (concrete)
  ├── *analyzer.ComplexityAnalyzer          (concrete)
  ├── *analyzer.InteractionDiversityAnalyzer(concrete)
  ├── *analyzer.UnlinkedCapabilityAnalyzer  (concrete)
  ├── *analyzer.SignalSuggestionGenerator   (concrete)
  ├── *analyzer.ValueChainAnalyzer          (concrete)
  ├── *analyzer.ValueStreamAnalyzer         (concrete)
  ├── *analyzer.ImpactAnalyzer              (concrete)
  ├── *ai.OpenAIClient                      (concrete)
  ├── *repository.ModelStore                (concrete)
  └── *repository.ChangesetStore            (concrete)
```

This means: no mocking individual analyzers in handler tests, no swapping implementations, no interface segregation.

### 2.3 — Analyzer Constructs Another Analyzer (MAJOR)

`ValueChainAnalyzer.Analyze()` (line 48-50) creates a new `CognitiveLoadAnalyzer` with `entity.DefaultConfig()`:

```go
defaults := entity.DefaultConfig().Analysis
ca := NewCognitiveLoadAnalyzer(defaults.CognitiveLoad, defaults.InteractionWeights)
bReport := ca.Analyze(m)
```

This violates Dependency Inversion (infrastructure constructing infrastructure with hardcoded defaults) and creates an inconsistency: the server's configured thresholds are ignored inside value chain analysis.

---

## 3. SOLID Compliance

### 3.1 — Single Responsibility: PARTIAL

| Component | Verdict | Issue |
|---|---|---|
| Domain entities | PASS | Each entity file has one entity |
| Domain services | PASS | `Validator` validates, `ChangesetApplier` applies |
| Analyzers | PASS (mostly) | One analysis per type, except `ValueChainAnalyzer` embedding cognitive load |
| `Handler` struct | FAIL | 18 deps, ~32 methods, handles everything from health to AI insights |
| `view_enriched.go` | FAIL | Anti-pattern detection + view assembly + analyzer orchestration in one 1262-line file |
| `signals.go` | FAIL | 6-analyzer orchestration + health classification + response assembly in one 329-line handler |
| `analysis.go` | PARTIAL | Dispatch is OK, but `build*Response` presentation logic is mixed with orchestration |

### 3.2 — Open/Closed: FAIL

Adding a new analysis type requires editing **4 files**:
1. `analysis.go` — `validAnalysisType` map + `runAnalysis` switch
2. `handler.go` — new field on Handler struct
3. `handler.go` — new constructor parameter
4. `cmd/server/main.go` — new analyzer instantiation

Adding a new view type requires editing **2 files**:
1. `view.go` — `handleView` switch
2. `cmd/server/main.go` (if new dependency needed)

**Fix**: Registry/map pattern. Each analyzer registers itself with a type key.

### 3.3 — Liskov Substitution: PASS

No interface stubs, no `panic("not implemented")`, no partial implementations found.

### 3.4 — Interface Segregation: FAIL

`Handler` requires all 18 dependencies even for endpoints that use zero of them (e.g., `/api/health`, `/api/config`). Every handler test must construct the full 18-argument object.

### 3.5 — Dependency Inversion: PARTIAL

- **PASS**: `usecase.ParseAndValidate` depends on `ModelParser` and `ModelValidator` interfaces.
- **FAIL**: Everything else is concrete. Handlers depend on `*analyzer.X` not `analyzer.XPort`. AI client is concrete. Stores are concrete.

---

## 4. Dead Code

### 4.1 — Backend Dead Code

| Location | What | Lines | Evidence |
|---|---|---|---|
| `handler/view.go` L74-395 | Six legacy `build*View` functions + types (`viewNode`, `viewEdge`, `viewResponse`) | ~320 lines | Never called. `handleView` routes to `buildEnriched*View` exclusively. |
| `usecase/query_engine.go` | `QueryEngine` — 8 exported methods | ~145 lines | `NewQueryEngine()` only called in `query_engine_test.go`. HTTP query handlers iterate `stored.Model` directly. Entire type is unused in production. |
| `middleware.go` L49-53 | `corsMiddleware` | 5 lines | Only used in tests. Production uses `makeCORSMiddleware`. Intentional compat wrapper but effectively dead. |

### 4.2 — Frontend Dead Code

| Location | What | Evidence |
|---|---|---|
| `pages/ViewPage.tsx` | Entire component | Never imported. `App.tsx` wires views directly, no `/views/:viewType` route. |
| `types/model.ts` | `TEAM_TYPE_COLORS`, `VIEW_TYPES`, re-exported types | Zero imports from any other file in `frontend/src/`. |
| `lib/runtimeConfig.ts` | `RuntimeConfig`, `getRuntimeConfig` | Zero references. Backend `GET /api/config` exists but frontend never calls this helper. |
| `lib/api.ts` L344-399 | `getCapabilities`, `getTeams`, `getNeeds`, `getServices`, `getActors`, `getAnalysis`, `analyzeAll` | Views use `api.getView(modelId, '...')` instead. These typed methods are never called. |
| `lib/api.ts` L454-463 | `getNeedView`, `getCapabilityView` | Exist but views use `getView` + `as unknown as` instead. |
| `lib/api.ts` L114-117 | `CognitiveLoadAnalysis` interface | Not imported anywhere. |
| `components/ui/badge.tsx` | `Badge`, `badgeVariants` | Not imported. Views define local badge components. |

---

## 5. Code Duplication

### 5.1 — Backend Duplication

| Pattern | Locations | Impact |
|---|---|---|
| Cognitive load analyzed multiple times per request | `signals.go` L95 runs `h.cognitiveLoad.Analyze(m)`, then `h.valueChain.Analyze(m)` internally runs it again with `DefaultConfig()` | Same model analyzed 2x per signals request with potentially different thresholds |
| `countHighLoadTeams` / `anyHighLoad` | `impact.go` L209-217 and `signals.go` L291-297 | Same concept, different helpers |
| `coalesceNeedRisks` / `coalesceCapItems` / `coalesceTeamItems` / `coalesceSvcItems` / `coalesceExtDepItems` | `signals.go` L301-329 | Five nearly identical "truncate list if > 12" functions — single generic would suffice |
| CLI `print*` functions mirror `build*Response` handlers | `cmd/cli/main.go` vs `handler/analysis.go` | Two maintenance paths for identical report logic |
| Test wiring | `newTestHandler` in 5+ test files reconstructs full 18-arg Handler | Single test helper factory needed |

### 5.2 — Frontend Duplication

| Pattern | Locations | Impact |
|---|---|---|
| `slug()` helper for insight keys | Copied identically in NeedView, CapabilityView, SignalsView, CognitiveLoadView, TeamTopologyView, OwnershipView | 6 copies of the same function |
| `TEAM_TYPE_BADGE` color maps | CognitiveLoadView, OwnershipView, RealizationView (+ unused `TEAM_TYPE_COLORS` in `types/model.ts`) | 4 copies of team-type → color mapping |
| Fetch + loading + error + redirect pattern | Every view: `useEffect` + `api.getView` + `as unknown as` + `navigate('/')` | ~15 lines duplicated per view (8+ views) |
| `VIS_BADGE` / visibility color maps | NeedView, CapabilityView, UNMMapView | 3 copies |
| Inline hex colors | `#111827`, `#6b7280`, `#9ca3af`, `#f3f4f6` etc. repeated across all views | 100+ hardcoded hex values vs CSS variables defined in `index.css` but unused |
| Loading/error fallback JSX | Every view repeats `flex items-center justify-center h-full` + gray/red text | Shared `ViewState` component needed |
| `ExpandableRow` | SignalsView defines its own local copy (L79+); `components/ui/ExpandableRow.tsx` also exists and is used by NeedView | Two implementations |

---

## 6. Oversized Components

### Backend (>50 lines)

| File | Function | Lines | Recommendation |
|---|---|---|---|
| `view_enriched.go` | `buildEnrichedOwnershipView` | ~256 lines | Extract lane builder, popover builder |
| `view_enriched.go` | `buildEnrichedCapabilityView` | ~234 lines | Extract grouping logic, service chip builder |
| `view_enriched.go` | `buildEnrichedNeedView` | ~150 lines | Extract actor grouping, risk badge builder |
| `signals.go` | `handleSignals` | ~195 lines | Move to `usecase/signals_service.go` |
| `view_enriched.go` | `buildUNMMapView` | ~136 lines | Separate file: `unm_map_builder.go` |
| `ai.go` | `buildAIPromptData` | ~129 lines | Move to `usecase/ai_context_builder.go` |
| `changeset_applier.go` | `deepCopyModel` | ~140 lines | Extract per-entity-type copy helpers |

### Frontend (>300 lines)

| File | Lines | Recommendation |
|---|---|---|
| `OwnershipView.tsx` | ~855 | Extract `TeamLane`, `ServicePopover`, `AntiPatternPanel`, `FilterBar` |
| `TeamTopologyView.tsx` | ~745 | Extract `TopologyGraph`, `TopologyTable`, `TeamDetailPanel` |
| `UNMMapView.tsx` | ~654 | Extract `MapNode`, `MapLegend`, `MapPanel` |
| `CapabilityView.tsx` | ~537 | Extract `CapabilityCard`, `CapabilityDetailPanel`, `GroupingSelector` |
| `SignalsView.tsx` | ~449 | Extract `SignalSection`, `SignalRow`, consolidate `ExpandableRow` |
| `CognitiveLoadView.tsx` | ~394 | Extract `LoadCard`, `LoadSidePanel`, `LoadLegend` |
| `AdvisorPanel.tsx` | ~391 | Extract `ChatMessage`, `QuickActions`, `PageConfigResolver` |

---

## 7. Type Safety Issues (Frontend)

### `as unknown as` Casts — 7 Locations

Every enriched view fetches via `api.getView(modelId, 'viewType')` which returns `ViewResponse` (nodes/edges graph format from Phase 3), but enriched views return completely different shapes. Each view casts:

```typescript
const data = await api.getView(modelId, 'need') as unknown as NeedViewResponse
```

Found in: `NeedView.tsx`, `CapabilityView.tsx`, `OwnershipView.tsx`, `TeamTopologyView.tsx`, `CognitiveLoadView.tsx`, `RealizationView.tsx`, `UNMMapView.tsx`.

Each view also defines its own local response interface that partially duplicates or diverges from types in `api.ts`.

**Fix**: Use the existing typed methods (`getNeedView`, `getCapabilityView`) or add new ones, return correct types, delete casts.

---

## 8. Styling Inconsistency (Frontend)

The codebase has **three competing styling approaches**:

1. **Tailwind classes** (`className="flex items-center gap-2"`) — used for layout
2. **Inline `style={{}}` objects** — used for colors, borders, specific sizing
3. **CSS variables** (defined in `index.css` `@theme` block) — defined but largely unused

`index.css` defines HSL design tokens (`--color-foreground`, `--color-muted-foreground`, `--color-border`, etc.) but views hardcode hex equivalents (`#111827`, `#6b7280`, `#e5e7eb`).

The `Sidebar.tsx` uses imperative `onMouseEnter`/`onMouseLeave` DOM style mutation instead of CSS `:hover` or Tailwind `hover:`.

---

## 9. Test Coverage Gaps

### Missing Test Files

| Source File | Test Coverage |
|---|---|
| `handler/health.go` | No test |
| `handler/signals.go` | No test |
| `handler/insights.go` | No test |
| `handler/middleware.go` | No test |
| `handler/debug.go` | No test |
| `handler/view_enriched.go` | No dedicated test (covered indirectly via `view_test.go`) |
| `handler/router.go` | No test |
| `handler/respond.go` | No test |
| `repository/changeset_store.go` | No test |
| `cmd/runquestions/main.go` | No test |
| **Frontend**: entire `src/` | Zero test files, no vitest/jest dependency |

### Test Anti-Pattern: Full Wiring in Every Test

Every handler test file duplicates the full 18-argument `newTestHandler()` construction. If a new analyzer is added, all test files break. Solution: single `testutil.NewTestHandler()` factory.

---

## 10. KISS Violations

| Issue | Where | Simpler Alternative |
|---|---|---|
| 18-parameter constructor | `handler.New()` | `HandlerDeps` struct or functional options |
| Switch-based dispatch for analyze/view | `analysis.go`, `view.go` | Table-driven registry |
| Analyzer constructed inside view builder | `view_enriched.go` L914 | Injected dependency |
| `ValueChainAnalyzer` builds `CognitiveLoadAnalyzer` | `value_chain.go` L48-50 | Constructor injection |
| 5 identical truncation helpers | `signals.go` L301-329 | Generic `coalesce[T](items []T, max int) []T` |
| Duplicated test wiring | 5+ files | Single `testutil` package |

---

## 11. TODO Comments in Codebase

| File | Line | Comment |
|---|---|---|
| `parser/yaml_parser.go` | L120 | `// TODO: emit deprecation warnings in a future version.` |
| `parser/yaml_parser.go` | L344 | `// TODO: emit deprecation warnings in a future version.` |

No `FIXME`, `HACK`, or `XXX` found.

---

## Summary of Findings by Priority

### Must Fix (Architectural / Structural)

1. **Extract business logic from handlers** into use case services (`SignalsService`, `AIContextBuilder`, `AnalysisRunner`)
2. **Fix `ValueChainAnalyzer` default config** — inject `CognitiveLoadAnalyzer` instead of constructing with defaults
3. **Remove dead code** — ~465 lines backend, ~200 lines frontend
4. **Fix type safety** — eliminate `as unknown as` casts, consolidate view types

### Should Fix (Maintainability)

5. **Registry pattern** for analyzers and views — Open/Closed compliance
6. **`HandlerDeps` struct** — replace 18-parameter constructor
7. **Shared test helper factory** — eliminate duplicated wiring
8. **Frontend hooks**: `useModelView<T>` to replace duplicated fetch patterns
9. **Consolidate duplicated code** — slug helper, team type badges, visibility colors

### Nice to Have (Polish)

10. **Migrate inline hex to CSS variables** — use existing `@theme` tokens
11. **Extract oversized components** — files > 300 lines into sub-components
12. **Standardize styling approach** — pick Tailwind + CSS vars, stop inline styles
13. **Add missing handler tests** — health, signals, insights, middleware, debug
