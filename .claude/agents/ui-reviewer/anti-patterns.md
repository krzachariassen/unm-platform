# UI Review Anti-Patterns — Do NOT Do These

## 1. Reporting Subjective Style Preferences as Critical
```md
WRONG — severity inflation
**Critical:** I prefer rounded-lg over rounded-md on cards.

CORRECT — separate taste from breakage
**Suggestion (non-blocking):** Consider rounded-lg for consistency with Settings.
**Critical:** Primary action is missing on mobile; users cannot complete the flow.
```

## 2. Missing Cross-View Consistency Checks
```md
WRONG — only reviewing one screen in isolation
The Teams tab shows "12 teams" and the summary looks fine.

CORRECT — compare the same fact across surfaces
Teams tab: 12 teams. Dashboard summary: 11 teams. Capability detail chips: 12.
Flag the mismatch and identify which count is authoritative (or fix the stale view).
```

## 3. Not Testing No-Model State
```md
WRONG — review assumes a model is always loaded
Verified list sorting and filters with examples/inca.unm.yaml loaded.

CORRECT — include empty and error paths
- Load app with no model selected: empty state copy, CTA to import/open
- Invalid or corrupt model: error message is understandable; user can recover
- Loading state: no layout jump or double-fetch flicker
```
