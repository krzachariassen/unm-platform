# BACKLOG-ACTIVE.md — AI-Managed Active Backlog

_10-20 items. AI updates this file. Human owns `docs/BACKLOG.md` (strategic roadmap)._
_Last updated: 2026-03-29_

---

## Recently Completed

- [x] **6.10.1** — Extend bottleneck analyzer: external dependency fan-in detection (2026-03-29)
- [x] **6.10.2** — Add external dependency signals to signals view API + type cleanup in api.ts (2026-03-29)
- [x] **6.10.3** — Add external deps to Capability View backend (2026-03-29)
- [x] **6.10.4** — Add external dependency nodes to UNM Map backend response (2026-03-29)
- [x] **6.10.5** — Render external deps in SignalsView (2026-03-29)
- [x] **6.10.6** — Render external deps in CapabilityView detail panel (2026-03-29)
- [x] **6.10.7** — Render external dependency nodes in UNMMapView (2026-03-29)
- [x] **6.10.8** — Model state persistence across page refresh (localStorage) (2026-03-29)
- [x] **6.10.9** — Insights endpoint returns HTTP 200 on internal errors + regression test (2026-03-29)
- [x] **6.10.10** — Impact analyzer uses injected config instead of hardcoded DefaultConfig (2026-03-29)
- [x] **6.10.11** — AI Advisor page in sidebar nav (2026-03-29)
- [x] **6.10.12** — Delete ~300 lines of dead legacy view builder functions (2026-03-29)
- [x] **6.10.13** — Dashboard error handling for signals API failure (2026-03-29)
- [x] **6.10.14** — Frontend api.ts type definitions complete — typed view fetch methods (2026-03-29)
- [x] **6.10.15** — AI client reads config-resolved API key correctly (2026-03-29)
- [x] **6.10.16** — Panic recovery middleware added (2026-03-29)

---

## Up Next

- [ ] **6.10.17** — Missing handler tests for health, signals, insights, advisor endpoints (#backend)
- [ ] **6.10.18** — No frontend tests. Add vitest + smoke tests for model-context, api.ts, and 3 major views (#frontend)
- [ ] **6.10.19** — Empty AI question validation. Return 400 on empty/whitespace — `ai.go` (#backend)
- [ ] **6.10.20** — Inconsistent error handling in frontend API client. Standardize all functions to use `extractError` — `api.ts` (#frontend)
- [ ] **6.10.21** — Config handler duplicates JSON encoding instead of using `writeJSON` helper — `config_handler.go` (#backend)
- [ ] **6.10.22** — `@import` PostCSS warning. Move Google Fonts import above @tailwind directives — `index.css` (#frontend)
- [ ] **6.10.23** — http.Client in OpenAI client has no default timeout. Add 120s safety net — `openai_client.go` (#backend)
- [ ] **7.1** — State snapshots: named model snapshots (current, target-q3) (#backend)
