# UI Review Severity Guide

## Critical (must fix before release)

- Data displayed incorrectly (wrong numbers, wrong labels)
- Interactive elements that don't work (buttons, links, filters)
- Panels/modals that can't be closed
- JavaScript errors visible in console
- Navigation that leads to blank pages
- Data inconsistency across views

## UX Issues (should fix)

- Warning icons without explanation text
- Missing loading/error/empty states
- Poor discoverability (invisible buttons, hidden actions)
- Panels that overlap critical content
- Click targets too small to hit reliably
- Inconsistent formatting (some values Title Case, others kebab-case)

## Minor (nice to fix)

- Spacing inconsistencies
- Animation jankiness
- Tooltip wording improvements
- Color contrast edge cases
- Alignment off by a few pixels

## Report Format

```
## Page: [Name] (`/route`)

### Critical Issues
1. **[Element]** — [Description]
   - Expected: [what should happen]
   - Actual: [what actually happens]

### UX Issues
1. **[Element]** — [Description]
   - Why it matters: [user impact]

### Missing Information
1. **[Element]** — [What's missing and why it matters]

### Minor Issues
1. **[Element]** — [Description]
```
