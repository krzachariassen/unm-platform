# Documentation Writer -- Operational Memory

> **Policy**: 30-entry cap · Monthly curation (Promote / Keep / Prune)
> See `.claude/agents/AGENT_OWNERSHIP.md` §2 for full curation rules.

## Learnings

### 2025-03 -- Current documentation state
- `README.md` exists with project overview, tech stack, structure, backlog, quick start
- `docs/` has: BACKLOG, UNM_DSL_SPECIFICATION, CONFIGURATION, YAML_GUIDE, AI strategy/vision docs
- Engineering principles consolidated into `.claude/agents/common/domain-model.md` and agent rules
- `examples/` exists at project root with `inca.unm.yaml` and `inca.unm.v2.yaml`

### 2025-03 -- Example models in testdata
- `backend/testdata/simple.unm.yaml` -- minimal valid model (1 actor, 1 need, 1 capability, 1 service, 1 team)
- `backend/testdata/relationships.unm.yaml` -- demonstrates short/long form relationships
- `backend/testdata/invalid.unm.yaml` -- invalid model for parser error testing
- These all use generic names (User, Customer, Admin) -- good for tests but not for docs

### 2025-03 -- DSL specification
- `docs/UNM_DSL_SPECIFICATION.md` is comprehensive (639 lines) covering meta-model,
  syntax, validation rules, view types
- It uses INCA-flavored examples throughout (Merchant, catalog, feed-ingestion)
- The YAML format is the primary/canonical format; custom `.unm` DSL is secondary

### 2025-03 -- Broken "Load INCA Extended Example" button
- The debug handler at `POST /api/debug/load-example` looks for `inca.unm.extended.yaml`
  (tries several relative paths) — this file does NOT exist in the repo
- The frontend button "Load INCA Extended Example" in UploadPage.tsx calls this endpoint
- Do NOT document or instruct users to click this button — it will fail
- Instead, direct users to upload `examples/inca.unm.yaml` manually via the Upload page

### 2026-03 -- README structure update
- Added `backend/pkg/unmmodel/` (shared model types)
- Added `frontend/src/components/graph/` (React Flow graph components)
- Added `scripts/` top-level directory
- Moved `config/` to top-level (separate from backend/) with actual file list
- Removed "(7 agents)" count from .claude agents comment (may change)

### 2026-03 -- DSL documentation overhaul
- Created `docs/DSL_GUIDE.md` as the primary tutorial for writing `.unm` files
- Created `examples/bookshelf.unm` using the approved BookShelf fictional domain
- Updated `README.md`: added DSL quick start snippet, fixed project structure tree, updated examples listing
- Updated `YAML_GUIDE.md`: added tip pointing to DSL Guide, fixed stale example references
- Updated `UNM_DSL_SPECIFICATION.md`: replaced all INCA/Uber examples with BookShelf domain, fixed interaction syntax (uses `->` not `from/to` keywords), added version/lastModified/author metadata fields, added `parent` flat reference for capabilities, added inline `interacts` on teams, fixed `external` keyword to `external_dependency`
- Old `inca.unm.yaml` and `inca.unm.v2.yaml` examples no longer exist; replaced by `minimal.unm.yaml`, `bookshelf.unm`, `nexus.unm`, `nexus.unm.yaml`
- The DSL parser supports both `//` and `#` comments
- The DSL interaction syntax is `interaction "from" -> "to" { ... }` not `interaction { from ... to ... }`
- The `supportedBy` field in DSL uses `supportedBy "name"` not `supportedBy capability "name"`

## Known Gaps

- `inca.unm.extended.yaml` referenced by the debug handler does not exist — the "Load Example" button is broken
- No tutorial for the What-If Explorer feature
- No architecture decision records (ADR) documentation
