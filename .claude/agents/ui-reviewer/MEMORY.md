# UI Reviewer — Operational Memory

## Reviews Completed

### 2025-03 — Full Platform Review (post-edit-panel)

Issues found and fixed:
- Signal explanations showing "undefined" for team_span — added fallback to "multiple"
- Team type filter labels using raw kebab-case — added Title Case formatting
- Need "at risk" reason unclear for 0-team edge case — improved explanation
- Capability visibility badges missing for unknown values — added fallback
- QuickAction buttons invisible (opacity:0) — changed to opacity:0.35
- Team topology detail drawer not toggling off — fixed toggle logic
- Realization capability chips not navigating with context — fixed deep-linking
- AntiPatternPanel lacking click-outside dismiss — added invisible backdrop

## Recurring Issues (watch for these)

- Warning icons without text explanation
- Popovers positioned off-screen on narrow viewports
- Filter combinations producing empty views with no "empty state" message
- Inconsistent data counts between dashboard and detail views
- Team type `complicated-subsystem` displayed raw instead of formatted
