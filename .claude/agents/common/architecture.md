# UNM Platform Architecture

Go backend (Clean Architecture) + React frontend (Vite + Tailwind + shadcn/ui).

## Backend Structure

```
backend/internal/
├── domain/          # Innermost — pure Go, zero external deps
│   ├── entity/      # Actor, Need, Capability, Service, Team, UNMModel
│   ├── valueobject/ # TeamType, InteractionMode, Confidence
│   └── service/     # Validation, anti-pattern detection, changeset applier
├── usecase/         # Application orchestration — depends only on domain
├── adapter/
│   ├── handler/     # HTTP handlers (REST API)
│   ├── presenter/   # View model transformers
│   └── repository/  # In-memory stores (ModelStore, ChangesetStore)
└── infrastructure/
    ├── parser/      # YAML parser
    ├── analyzer/    # 12 analysis engines
    ├── serializer/  # YAML export
    └── ai/          # OpenAI integration, prompt templates
```

## Frontend Structure

```
frontend/src/
├── pages/           # Route pages (UploadPage, DashboardPage, EditModelPage)
│   └── views/       # View components (UNMMapView, NeedView, CapabilityView, etc.)
├── components/      # Reusable UI (layout/, changeset/, advisor/, ui/)
├── hooks/           # useAIEnabled, usePageInsights, etc.
└── lib/             # api.ts, model-context.tsx, search-context.tsx, config.ts
```

## Dependency Rule

Dependencies point inward: Domain ← UseCases ← Adapters ← Infrastructure.
Domain has ZERO external imports. Never import adapter or infrastructure from domain.
