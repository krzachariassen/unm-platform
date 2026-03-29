# Documentation Writer Anti-Patterns — Do NOT Do These

## 1. Documenting Features That Do Not Exist Yet (Aspirational Docs)
```md
WRONG — reads as shipped behavior
## Real-time collaboration
Multiple editors see live cursors and conflict resolution...

CORRECT — label intent and status
## Roadmap: real-time collaboration (not implemented)
Planned behavior: ... See BACKLOG.md for status.
```

## 2. Including YAML Examples Without Verifying They Parse
```md
WRONG — hand-written snippet that fails the real parser
capabilities:
  - id: bad example
    name: Unclosed quote

CORRECT — copy from examples/*.unm.yaml or run parser after edits
capabilities:
  - id: example-cap
    name: "Verified against unm-platform parser"
# Confirm: unm-platform parser or CI snippet check passes
```

## 3. Using Internal Project Names as Examples in Public Docs
```md
WRONG — leaks org-specific codenames in user-facing README
Configure your model like the UberX Dispatch Mesh `hivemind-prod` graph.

CORRECT — neutral, fictional examples
Configure your model like `payments-api` serving the `checkout` capability.
```
