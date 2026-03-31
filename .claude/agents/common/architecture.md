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
├── pages/           # Route pages (UploadPage, DashboardPage)
│   └── views/       # View components (UNMMapView, NeedView, CapabilityView, etc.)
├── features/        # Domain-specific logic scoped to one feature (unm-map/, whatif/)
├── components/      # Reusable UI
│   ├── ui/          # shadcn/ui + shared components (PageHeader, StatCard, etc.)
│   ├── layout/      # AppShell, Sidebar, TopBar, SectionTabs
│   └── <domain>/    # Domain-specific components (changeset/, advisor/, unm-map/)
├── hooks/           # Shared custom hooks (TanStack Query wrappers)
├── services/api/    # API client functions split by domain (models, views, changesets)
├── types/           # Shared TypeScript interfaces, enums
└── lib/             # model-context.tsx, changeset-context.tsx, search-context.tsx, utils.ts
```

Dependencies flow: Pages → Features → Components → Hooks → Services → Types.
See `.claude/rules/frontend-architecture.md` for full rules.

## Dependency Rule

Dependencies point inward: Domain ← UseCases ← Adapters ← Infrastructure.
Domain has ZERO external imports. Never import adapter or infrastructure from domain.
