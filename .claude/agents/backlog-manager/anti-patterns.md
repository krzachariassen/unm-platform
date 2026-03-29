# Backlog Manager Anti-Patterns — Do NOT Do These

## 1. Inventing New Backlog Items Not Traced to the Roadmap
```md
WRONG — AI adds arbitrary work not aligned with existing docs/BACKLOG.md intent
- [ ] Add dark mode toggle
- [ ] Rewrite parser in Rust

CORRECT — restate or pull from what humans already captured; propose additions in prose until approved
- [ ] (from agreed scope) DSL: support X as specified in section Y
```

## 2. Reordering Items (Human Decides Priority)
```md
WRONG — reshuffling items to match model preference
Moved "Graph export" above "API versioning" because it seems faster.

CORRECT — preserve human order; note suggestions in prose if needed
Keep order from docs/BACKLOG.md / human edit; do not reorder priorities.
```

## 3. Unbounded Growth Without Human Triage
```md
WRONG — docs/BACKLOG.md accumulates many ad-hoc unchecked lines without approval

CORRECT — keep the file structured; trim Recently Completed; add new work only after human approval
```
