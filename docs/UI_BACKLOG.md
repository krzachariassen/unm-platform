# UNM Platform — UI Fix Backlog

Generated from the full critical UI review (2026-03-24). Every finding is captured here as an actionable backlog item with file references and acceptance criteria. Items are grouped by priority then by page.

---

## How to Execute This Backlog

**Use an agent team.** Most items below are independent (different files/pages) and can run in parallel. Suggested parallelization:

```
Wave 1 (highest priority, independent):
  Teammate "fix-unm-map-edges"       → UI-01 (UNM Map edges)
  Teammate "fix-edit-forms"          → UI-02 (Edit Model empty forms)
  Teammate "fix-warning-icons"       → UI-06 (⚠ icons global pattern)
  Teammate "fix-ai-insight-stuck"    → UI-09 (stuck AI insight loader)

Wave 2 (after Wave 1, independent):
  Teammate "fix-upload-dashboard"    → UI-11 through UI-22 (Upload + Dashboard)
  Teammate "fix-unm-map-detail"      → UI-23 through UI-28 (UNM Map UX)
  Teammate "fix-need-capability"     → UI-29 through UI-39 (Need + Capability views)
  Teammate "fix-ownership-topology"  → UI-40 through UI-52 (Ownership + Team Topology)

Wave 3 (after Wave 2):
  Teammate "fix-cogload-realization" → UI-53 through UI-64 (Cognitive Load + Realization)
  Teammate "fix-edit-whatif-ai"      → UI-65 through UI-77 (Edit + What-If + AI pages)
  Teammate "fix-crosscutting"        → UI-78 through UI-84 (Cross-cutting polish)
```

Each teammate must:
1. Read the relevant source files before editing
2. Make the minimal change that fixes the issue
3. Run `cd frontend && npm run build` to verify no TypeScript errors after each change
4. Never edit files owned by another teammate (check file ownership above)

---

## P0 — Showstoppers

---

### UI-01 — UNM Map: Draw edges connecting nodes

**File**: `frontend/src/pages/views/UNMMapView.tsx`

**Problem**: The SVG overlay draws zero edges between nodes. There are no lines connecting actor→need, need→capability, or capability→capability. The entire point of the UNM Map value chain visualization is missing.

**Root cause to investigate**: The component receives `edges` from the API response (`UNMMapViewResponse.edges`). Check whether:
1. The API is returning edges (verify via `fetch('/api/models/{id}/views/unm-map')`)
2. If edges exist in the response, find where they should be rendered in the SVG and why they are not being drawn
3. If edges are absent from the API response, the fix is in the backend presenter (`internal/adapter/presenter/view.go` — `buildUNMMapView`)

**Fix**:
- Inspect the `edges` array from the API response in `UNMMapViewResponse`
- For each edge, draw an SVG path/line from the source node's center to the target node's center
- Edge types to distinguish visually:
  - Actor→Need: solid gray arrow (`#9ca3af`)
  - Need→Capability (`supportedBy`): solid blue arrow (`#3b82f6`)
  - Capability→Capability (`dependsOn`): dashed purple arrow (`#7c3aed`)
  - Capability→Service (`realizedBy`): already shown as service pills inside capability nodes — no separate edge needed
- Use quadratic bezier curves (`<path d="M... Q... ..."/>`) for edges that would otherwise overlap, straight lines for clearly separated nodes
- Arrowheads: add an SVG `<defs><marker>` arrowhead and reference it via `marker-end`
- Render edges in the SVG **before** nodes so nodes appear on top

**Acceptance**:
- Opening `/unm-map` with the INCA example shows visible arrows from each actor to its needs, from each need to its supporting capabilities, and between capability nodes that have `dependsOn` relationships
- Edges do not pass through unrelated nodes (basic routing or bezier offset is sufficient)
- Clicking an edge or hovering shows a tooltip with the relationship description if one exists

---

### UI-02 — Edit Model: Fix empty forms for 10+ broken action types

**File**: `frontend/src/components/changeset/ActionForm.tsx`

**Problem**: Selecting the following action types renders a completely empty form with no input fields:
`change_type`, `change_size`, `split_team`, `merge_teams`, `reassign_capability`, `change_visibility`, `link_to_service`, `add_need`, `link_need_cap`, `add_actor`

**Fix**:
- Open `ActionForm.tsx` and find the switch/if-else block that renders fields per action type
- For each broken action type, add the appropriate fields:

| Action type | Fields needed |
|---|---|
| `change_type` | Entity type selector (Team/Service) + entity name dropdown + new type dropdown (TeamType or ServiceType enum) |
| `change_size` | Team dropdown + new size input (number or Small/Medium/Large select) |
| `split_team` | Source team dropdown + new team name input + services to move (multi-select) |
| `merge_teams` | Team A dropdown + Team B dropdown + merged team name input |
| `reassign_capability` | Capability dropdown + from-team dropdown + to-team dropdown |
| `change_visibility` | Capability dropdown + new visibility select (user-facing/domain/foundational/infrastructure) |
| `link_to_service` | Capability dropdown + service dropdown |
| `add_need` | Actor dropdown + need name input + outcome input |
| `link_need_cap` | Need dropdown + capability dropdown |
| `add_actor` | Actor name input + description input |

- All entity dropdowns must be populated from the loaded model data (available via `useModel()` → `parseResult`)
- All enum selects must use the correct option values matching the backend API

**Acceptance**:
- Selecting any of the 10 previously-broken action types shows a non-empty, correctly-labelled form
- All entity fields use `<select>` dropdowns populated with real model entities, not free-text inputs
- Filling in a form and clicking Add appends the action to the Pending Changes list

---

### UI-03 — Edit Model: Fix "Add action" not appending to pending list

**File**: `frontend/src/components/changeset/ActionForm.tsx`, `frontend/src/components/changeset/ActionList.tsx`

**Problem**: After filling in fields for `move_service` (or any action), clicking Add does not append the action to the Pending Changes list. "No actions added yet" persists.

**Fix**:
- Trace the Add button's `onClick` handler in `ActionForm.tsx`
- Verify the action object being constructed matches the shape expected by the parent state
- Check that the callback prop (`onAdd` or equivalent) is being called with a valid, non-empty action object
- Verify `ActionList.tsx` correctly renders the list when it has items
- Add basic form validation: if required fields are empty, show inline error messages and do NOT call `onAdd`

**Acceptance**:
- Adding any action type appends it to the Pending Changes list immediately
- The list persists across action type changes
- Individual actions can be removed from the list

---

### UI-04 — Sidebar: Fix disabled nav items — truly disable or truly enable

**File**: `frontend/src/components/layout/Sidebar.tsx`

**Problem** (from reviewer): Sidebar items appear visually grayed out when no model is loaded but may be clickable. The current code uses `pointer-events-none` on the NavLink when `disabled` is true (line 47), which is correct. However, the NavLink `to` prop still resolves, and React Router may still process the link in some edge cases.

**Fix**:
- Confirm `pointer-events-none` is correctly applied by also adding `tabIndex={-1}` and `aria-disabled="true"` to disabled links
- Add a tooltip on hover for disabled links: `"Load a model first to access this view"`
- Use `<span>` instead of `<NavLink>` when disabled so there is zero chance of navigation

```tsx
// When disabled, render a <span> instead of <NavLink>
if (disabled) {
  return (
    <span key={to} title="Load a model first" aria-disabled="true"
      className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm opacity-40 cursor-not-allowed">
      <Icon size={15} style={{ flexShrink: 0 }} />
      {label}
    </span>
  )
}
```

**Acceptance**:
- When no model is loaded, clicking any grayed nav item does nothing and shows a "Load a model first" tooltip
- When a model is loaded, all items are clickable normally

---

### UI-05 — Dashboard: Fix stats card data desync (race condition)

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: Stats cards (actors, needs, capabilities, services, teams) show different numbers than the signals section, suggesting they pull from different state sources or have a race condition during model loading.

**Fix**:
- Audit `DashboardPage.tsx` — identify all data-fetching calls and confirm they all use the same `modelId` from `useModel()`
- Ensure no fetch fires before the model is fully loaded (check that `modelId` is non-null before any API call)
- If stats and signals call different endpoints, verify both use the same `modelId`
- Add a loading state that shows a skeleton UI until ALL data fetches for the page have resolved — do not show partial data

**Acceptance**:
- Stats card counts (actors, needs, etc.) match the actual model entity counts as returned by the parse response
- Signal counts on the dashboard match the counts shown in `/signals`
- Refreshing the page shows consistent data across all sections

---

### UI-06 — Global: Add explanations to all ⚠ warning icons

**Files**:
- `frontend/src/pages/views/UNMMapView.tsx`
- `frontend/src/pages/views/OwnershipView.tsx`
- `frontend/src/pages/views/NeedView.tsx`
- `frontend/src/pages/views/RealizationView.tsx`
- Any other view that renders a warning/alert icon

**Problem**: Warning icons (⚠, triangle-exclamation, alert icons) appear throughout the app with no tooltip, aria-label, or adjacent text explaining what the warning is about. Users cannot determine the meaning.

**Fix**:
- Search all view files for warning icon renders: `⚠`, `AlertTriangle`, `AlertCircle`, `warning`
- For every warning icon, one of the following must be true:
  1. A `title` attribute or `aria-label` is set on the icon with the exact warning text
  2. A Tooltip component wraps the icon with explanatory text
  3. Adjacent visible text explains the warning inline
- Example fix for UNM Map capability warning:
  ```tsx
  // Before
  <span>⚠</span>
  // After
  <span title={`Fragmented capability: owned by ${teams.join(', ')}`}
        aria-label={`Warning: fragmented`} style={{cursor:'help'}}>⚠</span>
  ```
- In the Ownership View, capability warning icons should say something like: `"This capability is realized by services from multiple teams: [team A, team B]"`
- In the Need View, at-risk indicators should explain WHY the need is at risk in their tooltip (e.g., "At risk: capability 'X' is fragmented across 3 teams")

**Acceptance**:
- Hovering any ⚠ icon in the app shows a non-empty tooltip explaining the specific issue
- All warning icons have `aria-label` set for screen readers
- No bare ⚠ symbol or AlertTriangle icon appears anywhere without explanatory text

---

### UI-07 — Need View: Fix "Loading AI insight..." stuck indefinitely

**File**: `frontend/src/pages/views/NeedView.tsx`, `frontend/src/hooks/usePageInsights.ts`

**Problem**: When expanding a need row, the expanded content shows "Loading AI insight..." that never resolves — no timeout, no fallback, no error state.

**Fix**:
- In the insight loading logic, add a timeout (suggest 10 seconds): if the insight hasn't loaded after 10 seconds, replace the loading state with a fallback message
- If AI is disabled/unavailable, do not show the loading placeholder at all — suppress it entirely
- Add an error state: if the fetch fails, show "AI insight unavailable" instead of a forever-spinner
- Check `usePageInsights.ts` or `InsightsContext.tsx` for where the loading state is set and add the timeout there

```tsx
// Suggested states: 'loading' | 'loaded' | 'error' | 'unavailable'
// When AI is disabled: show nothing (no loading, no error)
// When loading > 10s: transition to 'error' state
```

**Acceptance**:
- When AI is disabled, no loading indicator appears in expanded need rows
- When AI is enabled but slow/failing, the spinner resolves to an error message within 10 seconds
- When AI loads successfully, the insight text appears correctly

---

### UI-08 — Cognitive Load: Expand cards to show readable detail with thresholds

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx`

**Problem**:
1. Team card expansion shows raw abbreviated numbers (`7 Dom 7 Svc 8 Ixn 5Deps`) with no context, thresholds, or severity
2. The dimension reference legend is buried at the bottom of the page, far from where the data appears
3. There is no expand/collapse — all data is shown inline in a single compressed row

**Fix**:
- Implement a proper expand/collapse for team cards:
  - Collapsed state: show team name, team type, overall load severity (High/Medium/Low with colored badge), and service/capability counts
  - Expanded state: show a table or card grid with each dimension:
    ```
    | Dimension        | Value | Threshold    | Severity |
    |------------------|-------|--------------|----------|
    | Domain Spread    |   7   | 6+ = High    |  🔴 High |
    | Service Count    |   7   | 6+ = High    |  🔴 High |
    | Interaction Load |   8   | 7+ = High    |  🔴 High |
    | Dependencies     |   5   | 5+ = Medium  |  🟡 Med  |
    ```
  - Also show in expanded state: list of owned services, list of owned capabilities, and the AI insight if available
- Move the dimension reference legend to be an expandable info panel at the TOP of the page (collapsible, defaulting to collapsed)
- Only one card should be expanded at a time: expanding a new card collapses the previously open one

**Acceptance**:
- Clicking a team card expands it to show readable dimension breakdown with threshold context
- Opening a second card closes the first
- Reference legend is accessible at the top without scrolling to the bottom

---

### UI-09 — Cognitive Load: Remove hardcoded "5 ppl" placeholder

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx`

**Problem**: Every team shows "5 ppl" as team size. This is a hardcoded placeholder that shows fake data and erodes user trust.

**Fix**:
- If the backend API response includes a team size field, use it
- If the backend does not provide team size data, remove the "ppl" metric entirely from the UI
- Do not show "5 ppl" or any other hardcoded value

**Acceptance**:
- No team shows "5 ppl" unless the backend API actually returns a people count of 5
- If no people count data is available, the metric is absent from the UI

---

## P1 — Critical Issues Per Page

---

### UI-10 — Upload: Add loading indicator and progress for example load

**File**: `frontend/src/pages/UploadPage.tsx`

**Problem**: Clicking "Load INCA Extended Example" shows no spinner, no disabled state, no progress feedback. Users don't know if their click registered.

**Fix**:
- Add a loading state variable (`isLoading`)
- When loading:
  - Disable the button and show a spinner icon inside it
  - Show step progress below the button: "Reading file... → Parsing & validating... → Analysing architecture... → Done!"
  - Each step should be visible for at least 300ms so users can read it
- After loading completes, keep the "Done!" state visible for 1 second before navigating to `/dashboard`

**Acceptance**:
- Clicking the example load button immediately shows a spinner and disabled button
- Steps transition visibly during loading
- Users see confirmation before navigation

---

### UI-11 — Upload: Show active model state on the upload page

**File**: `frontend/src/pages/UploadPage.tsx`

**Problem**: When a model is already loaded, `/` shows the exact same upload form with no indication that an active model exists.

**Fix**:
- If `modelId` is non-null (model is loaded), show a banner or card at the top of the page:
  ```
  ✓ Model loaded: [system name]
  [Go to Dashboard] [Upload a different model]
  ```
- The upload form should remain accessible below for replacing the model
- This makes the page's dual purpose (initial load + model replacement) clear

**Acceptance**:
- When a model is loaded, the page prominently shows the loaded model name and a link to the dashboard
- The upload form is still accessible for replacing the model

---

### UI-12 — Dashboard: Fix "1 warnings" grammar

**File**: `frontend/src/components/layout/TopBar.tsx`

**Problem**: TopBar shows "1 warnings" (plural) when there is exactly 1 warning. All pages are affected.

**Fix**:
```tsx
// Before
`${count} warnings`
// After
`${count} ${count === 1 ? 'warning' : 'warnings'}`
```

**Acceptance**: "1 warning" (singular), "2 warnings" (plural) — correct in all cases.

---

### UI-13 — Dashboard: Platform Health gauges need explanatory tooltips

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: The three health rings (UX, Arch, Org) show "Critical" or "Warning" with no explanation of what criteria determine the rating.

**Fix**:
- Wrap each health ring in a Tooltip that lists the contributing signals when hovered
- Example tooltip content: `"Critical: 6 needs at risk, 2 fragmented capabilities"`
- The tooltip text should be derived from the actual signal data, not hardcoded

**Acceptance**: Hovering any health ring shows a tooltip listing the specific issues contributing to that rating.

---

### UI-14 — Dashboard: Fix non-deterministic "Top 5 teams" ordering

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: The "Top 5 teams by structural cognitive load" list changes order on each page load.

**Fix**:
- Sort teams deterministically: primary sort by cognitive load score (descending), secondary sort alphabetically by team name for ties
- If the sort happens on the backend, verify the backend returns a stable sort

**Acceptance**: The top 5 list is always in the same order for the same model data.

---

### UI-15 — Dashboard: Make signal category rows clickable

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: The architecture signals summary (Needs at risk: 6, Fragmented capabilities: 2, etc.) shows counts but rows are not clickable. Users cannot drill into specifics.

**Fix**:
- Make each signal category row a link that navigates to `/signals` with a filter pre-applied for that category
- Style as a hoverable row with a right-arrow indicator
- Example: clicking "Fragmented capabilities: 2" navigates to `/signals?filter=fragmentation`
- The Signals view must then apply the filter on load if a `filter` query param is present

**Acceptance**: Clicking a signal category row on the dashboard navigates to the Signals view filtered to that category.

---

### UI-16 — Dashboard: Explain colored semicircles on stats cards

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: Stats cards have colored semicircles with no legend or tooltip explaining their meaning.

**Fix**:
- Option A: Add a tooltip on the semicircle explaining what the color represents (e.g., "Color indicates change from last analysis")
- Option B: If the semicircles are purely decorative, remove them
- Do NOT leave unexplained colored UI elements

**Acceptance**: Every colored element on stats cards either has an explanatory tooltip or is removed.

---

### UI-17 — Dashboard: Format validation warning as user-facing copy

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: Validation warnings display as raw internal text: `team "inca-core-dev" owns 7 capabilities (cognitive load)`

**Fix**:
- Parse the validation warning and format it as a user-facing message with an icon, formatted team name, and a link to the relevant view:
  ```
  ⚠ Team inca-core-dev owns 7 capabilities — this may cause high cognitive load.
  [View in Ownership →]
  ```

**Acceptance**: Validation warnings are readable, formatted, and link to the relevant view.

---

### UI-18 — Dashboard: Add data freshness indicator

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: No timestamp showing when the model was loaded or last analyzed.

**Fix**:
- Store the timestamp when `setModel()` is called in `model-context.tsx`
- Show "Last loaded: 3 minutes ago" or "Last loaded: today at 10:42" in the TopBar or Dashboard header

**Acceptance**: Dashboard shows when the model was last loaded.

---

### UI-19 — UNM Map: Fix external dependency detail panel showing "Warning" as type

**File**: `frontend/src/pages/views/UNMMapView.tsx`

**Problem**: Clicking an external dependency node shows "Warning" as the entity type label instead of "External Dependency".

**Fix**:
- Find the detail panel rendering logic for `ext-dep` node type
- Change the type label from "Warning" to "External Dependency"

**Acceptance**: External dependency detail panel header shows "External Dependency" not "Warning".

---

### UI-20 — UNM Map: Move detail panel to a fixed side drawer

**File**: `frontend/src/pages/views/UNMMapView.tsx`

**Problem**: Clicking a node appends detail content at the very bottom of the page, requiring scrolling. On a large map users don't see any response to their click.

**Fix**:
- Implement a fixed-position right side drawer for node detail:
  ```
  position: fixed; right: 0; top: 56px; bottom: 0; width: 320px;
  background: white; border-left: 1px solid #e5e7eb; overflow-y: auto;
  transform: translateX(100%); transition: transform 0.2s;
  ```
- When a node is clicked, slide the drawer in (`translateX(0)`)
- Show a close button (✕) in the drawer header
- The drawer should show: entity type badge, name, description, and all relationship fields
- When the edit panel is open, the detail drawer should close (they share the right side)

**Acceptance**:
- Clicking any node slides a detail drawer in from the right without requiring scroll
- The drawer shows complete entity information
- Close button dismisses the drawer

---

### UI-21 — UNM Map: Add needs list to actor detail panel

**File**: `frontend/src/pages/views/UNMMapView.tsx`

**Problem**: Clicking an actor shows only name and description. No list of the actor's needs.

**Fix**:
- In the actor detail panel (now the right drawer per UI-20), add a "Needs" section listing all needs for that actor
- Each need should show: name, status (mapped/unmapped), and a brief capability list
- This data is available in the `parseResult` from `useModel()`

**Acceptance**: Actor detail drawer shows the actor's full list of needs with their mapping status.

---

### UI-22 — UNM Map: Fix "Clear highlight" to use button element

**File**: `frontend/src/pages/views/UNMMapView.tsx`

**Problem**: "Clear highlight" is rendered as a `<span>` with no cursor or button affordance.

**Fix**:
```tsx
// Before: <span onClick={...}>✕ Clear highlight</span>
// After:
<button onClick={clearHighlight}
  style={{cursor:'pointer', color:'#6b7280', fontSize:12, background:'none', border:'none', padding:'2px 6px'}}
  className="hover:text-gray-900">
  ✕ Clear highlight
</button>
```

**Acceptance**: "Clear highlight" shows pointer cursor on hover and is keyboard-focusable.

---

### UI-23 — Need View: Add percentage to stats bar

**File**: `frontend/src/pages/views/NeedView.tsx`

**Problem**: Stats bar shows raw counts only. "11 TOTAL NEEDS / 0 UNMAPPED / 6 AT RISK" — no percentages.

**Fix**:
- Add percentage below or next to each count:
  - "AT RISK: 6 (55%)"
  - "UNMAPPED: 0 (0%)"
- Percentage = count / total × 100, rounded to nearest integer

**Acceptance**: Stats bar shows both raw count and percentage for at-risk and unmapped metrics.

---

### UI-24 — Need View: Add counts to visibility filter pills

**File**: `frontend/src/pages/views/NeedView.tsx`

**Problem**: Visibility filter pills (User Facing, Domain, etc.) show no counts.

**Fix**:
- Compute how many needs fall under each visibility level (based on their supporting capabilities' visibility)
- Show count in the pill: "User Facing (3)", "Domain (5)", etc.

**Acceptance**: Each filter pill shows the count of needs in that visibility category.

---

### UI-25 — Need View: Add outcome to collapsed need row

**File**: `frontend/src/pages/views/NeedView.tsx`

**Problem**: Outcome is only visible after expanding a need row, but it's core to understanding what the need means.

**Fix**:
- In the collapsed need row, show the outcome text truncated to 1 line (with CSS `text-overflow: ellipsis`) below the need name
- On hover, show the full outcome in a tooltip

**Acceptance**: Collapsed need rows show a truncated outcome with full text on hover.

---

### UI-26 — Capability View: Replace all-zero stat cards with positive message

**File**: `frontend/src/pages/views/CapabilityView.tsx`

**Problem**: When all issue stats are 0 (FRAGMENTED: 0, UNOWNED: 0, HIGH-SPAN: 0, AT RISK: 0), the four zero cards look like broken/non-functional analysis.

**Fix**:
- If ALL four stat values are 0, replace the four stat cards with a single green "No architecture issues detected" message card
- If at least one stat is non-zero, show all four cards normally

**Acceptance**: When no issues exist, the capability view shows a positive confirmation instead of four zero cards.

---

### UI-27 — Capability View: Add tooltip to `↙16` dependency annotation

**File**: `frontend/src/pages/views/CapabilityView.tsx`

**Problem**: The `↙16` annotation on high-dependency capabilities has no explanation.

**Fix**:
```tsx
<span title="16 other capabilities depend on this capability"
      style={{cursor:'help', fontSize:11, color:'#6b7280'}}>
  ↙{dependentCount}
</span>
```

**Acceptance**: Hovering `↙N` shows tooltip "N other capabilities depend on this capability".

---

### UI-28 — Capability View: Standardize detail panel fields with UNM Map

**File**: `frontend/src/pages/views/CapabilityView.tsx`

**Problem**: Capability detail panel in Capability View shows different field labels and structure than the capability detail in UNM Map.

**Fix**:
- Extract a shared `CapabilityDetailPanel` component used by both views
- Standard fields: Description, Visibility, Teams (with type), Services (with role), Depends On, Depended On By, External Dependencies, Anti-patterns (if any)
- Both views use this component so the detail is always consistent

**Acceptance**: Capability detail panel shows the same fields in the same order in both Capability View and UNM Map.

---

### UI-29 — Ownership View: Add tooltip to ⚠ icons on capabilities

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: Warning icons on capabilities in the ownership view have no tooltip.

**Fix**:
- For cross-team capabilities: `title="This capability is realized by services from multiple teams: [team A, team B]"`
- For other warnings: provide the specific reason

**Acceptance**: All ⚠ icons in Ownership View have non-empty, specific tooltip text.

---

### UI-30 — Ownership View: Add search/filter

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: No inline search for the ownership view. With 9 teams and 28 capabilities, finding a specific entity requires scrolling.

**Fix**:
- Add a search input above the team cards: `placeholder="Filter by team, service, or capability..."`
- Filter logic: show only team cards where team name, any service name, or any capability name matches the search term (case-insensitive substring)
- If the search term does not match any team's content, show "No results for '[term]'"

**Acceptance**: Typing in the search box filters team cards in real time. Clearing the search restores all cards.

---

### UI-31 — Ownership View: Fix team type filter chip active state

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: Team type filter chips have no visual active/selected state.

**Fix**:
- When a filter chip is active, apply a distinct style:
  ```
  background: #1d4ed8; color: white; border-color: #1d4ed8;
  ```
- When inactive:
  ```
  background: white; color: #374151; border: 1px solid #d1d5db;
  ```

**Acceptance**: Active filter chip is visually distinct from inactive chips.

---

### UI-32 — Ownership View: Show "no results" message for empty filter

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: Clicking a filter chip that matches no teams shows a blank page with no feedback.

**Fix**:
- If the filtered result is empty, show:
  ```
  No [team-type] teams in this model.
  [Clear filter]
  ```

**Acceptance**: Empty filter state shows an explanatory message and a way to clear the filter.

---

### UI-33 — Ownership View: Clarify "Show problems only" button scope

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: "Show problems only" is ambiguous — users don't know what "problems" includes.

**Fix**:
- Add a tooltip: `title="Show teams with: overloaded capabilities, cross-team fragmentation, or unowned services"`
- Alternatively, rename the button to "Problems only" with a `(?)` icon that shows the scope on hover

**Acceptance**: Users can understand what "Show problems only" filters before clicking it.

---

### UI-34 — Ownership View: Make "By Domain Area" view interactive

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: Switching to "By Domain Area" renders read-only text. Capability names and team names are not clickable, unlike the Team view.

**Fix**:
- In the Domain Area view, make capability names clickable buttons that open the capability detail drawer/panel
- Make team names clickable links that navigate to that team's section in Team view or open a team detail panel

**Acceptance**: Capability and team names in Domain Area view are clickable and open relevant detail panels.

---

### UI-35 — Ownership View: Make external dependency service counts expandable

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: External dependency shows "(2 svcs)" count but no way to see which services.

**Fix**:
- Change "(2 svcs)" to a clickable element that expands inline or shows a tooltip listing the service names:
  ```
  cadence (2 svcs: inca-core, inca-async)
  ```

**Acceptance**: Hovering or clicking the service count on an external dependency shows which services use it.

---

### UI-36 — Team Topology: Implement team detail drawer

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: Clicking a team node in the graph view does nothing. No detail drawer is implemented.

**Fix**:
- Implement a right-side drawer that slides in when a team node is clicked (graph view) or a table row is clicked (table view)
- Drawer content:
  - Team name (h2)
  - Team type badge with explanation tooltip (see UI-38)
  - Description
  - Services list (with count)
  - Capabilities list (with count)
  - Interactions list: "→ Team B (X-as-a-Service): via [capability]"
- Drawer close: ✕ button in header + clicking the backdrop
- In graph view, highlight the clicked team node while the drawer is open

**Acceptance**:
- Clicking a team node/row opens the detail drawer from the right
- Drawer shows complete team information
- Close button and backdrop click both dismiss the drawer

---

### UI-37 — Team Topology: Add interaction mode filter

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: No way to filter edges by interaction mode (X-as-a-Service, Collaboration, Facilitating).

**Fix**:
- Add a filter bar above the graph with three toggleable chips:
  - X-as-a-Service (default: on)
  - Collaboration (default: on)
  - Facilitating (default: on)
- When a mode is toggled off, hide the corresponding edges in the graph view and rows in the table view
- Active state uses the same chip styling as other filter chips in the app

**Acceptance**: Toggling an interaction mode chip hides/shows the corresponding edges/rows.

---

### UI-38 — Team Topology: Add search and team type filter

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: No search or team type filter on the Team Topology page.

**Fix**:
- Add a search input: filters teams by name in both graph and table views
- Add team type filter chips: stream-aligned, platform, enabling, complicated-subsystem
- In graph view, non-matching teams are dimmed; in table view, non-matching rows are hidden

**Acceptance**: Search and team type filter work in both graph and table views.

---

### UI-39 — Team Topology: Default to graph view and explain team types

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**:
1. Default view is table; should be graph
2. Team type terms are not explained anywhere

**Fix**:
- Set initial active tab to "graph" not "table"
- Add team type explanations as tooltips on the team type badge anywhere it appears in this view:
  - `stream-aligned`: "Aligned to a flow of work from a business domain segment"
  - `platform`: "Provides internal services to reduce cognitive load of other teams"
  - `enabling`: "Helps other teams adopt new practices or technologies"
  - `complicated-subsystem`: "Owns a subsystem requiring deep specialist knowledge"

**Acceptance**: Page opens on graph view by default. Team type badges have explanatory tooltips.

---

### UI-40 — Team Topology: Add team types to table rows

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: Interaction table shows team names but not team types, which are essential for Team Topologies analysis.

**Fix**:
- Add team type badges next to team names in the From and To columns:
  ```
  inca-core-dev [platform]   →   inca-ingestion-dev [stream-aligned]
  ```
- Use consistent color coding: stream-aligned=blue, platform=purple, enabling=green, complicated-subsystem=orange

**Acceptance**: Team type badges appear next to team names in the interaction table.

---

### UI-41 — Team Topology: Make "1 overloaded" badge interactive

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: The "1 overloaded" badge in the subtitle is not interactive.

**Fix**:
- Make the badge a button that filters/highlights the overloaded team(s)
- On click: in graph view, dim all non-overloaded teams; in table view, filter to rows involving overloaded teams
- Add a "clear" option when the filter is active

**Acceptance**: Clicking "1 overloaded" badge highlights/filters to show the overloaded team.

---

### UI-42 — Team Topology: Expand truncated interaction descriptions

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: Description column in the table is truncated at ~50 characters with no way to see the full text.

**Fix**:
- Add `title={fullDescription}` to the description cell so hovering shows the full text
- Alternatively, make rows expandable to show the full description

**Acceptance**: Full interaction description is accessible on hover or expand.

---

### UI-43 — Realization View: Fix page title to match sidebar label

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: Page title says "Value Chain Traceability" but sidebar calls it "Realization View".

**Fix**:
- Change the page `<h1>` to "Realization View" with subtitle "End-to-end value chain traceability"
- OR change the sidebar label to "Value Chain" if that name is preferred — but they must match

**Acceptance**: Page title and sidebar label refer to the same view with the same name.

---

### UI-44 — Realization View: Show team names inline for cross-team warnings

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: "2 TEAMS ⚠" shows a count without naming the teams. Users must expand rows or read tooltip.

**Fix**:
- In the collapsed need row, show the first two team names inline: "inca-core-dev, inca-ingestion-dev ⚠" with a tooltip listing all teams if there are more than 2
- The ⚠ icon should have `title="Served by multiple teams — this may indicate fragmentation"`

**Acceptance**: Cross-team need rows show team names inline without requiring expansion.

---

### UI-45 — Realization View: Fix collapsed actor group expand affordance

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: Some actor groups appear collapsed with no visible chevron or expand control.

**Fix**:
- All actor groups must have a visible `▶`/`▼` chevron at the left of the group header
- The entire group header row should be clickable to toggle expand/collapse
- Collapsed group shows need count; expanded shows all needs with their capability chains

**Acceptance**: All actor groups have visible expand/collapse controls and toggle correctly.

---

### UI-46 — Realization View: Expand "+N more" services inline

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: "+2 more" services truncation has no hover/click to reveal.

**Fix**:
- Show all services on a single line separated by commas, OR
- Make "+2 more" a clickable element that expands inline to show all services
- As a minimum: `title="Also: inca-publisher, inca-serving"` tooltip on "+2 more"

**Acceptance**: All services for a need-capability pair are accessible without expanding the row.

---

### UI-47 — Realization View: Make capability chips interactive in By Service tab

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: Capability chips in the By Service tab are plain text — not clickable.

**Fix**:
- Render capability names as `<button>` elements styled as chips
- On click: open the capability detail panel (reuse `CapabilityDetailPanel` from UI-28) or navigate to `/capability` with that capability highlighted

**Acceptance**: Capability chips in the By Service tab are clickable and show capability details.

---

### UI-48 — Realization View: Fix semantic color for "0 UNBACKED" stat card

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: "0 UNBACKED" uses a red background, but 0 unbacked capabilities is GOOD — red implies a problem.

**Fix**:
- When unbacked count is 0: use green background (`#f0fdf4` / `#16a34a`)
- When unbacked count is > 0: use red background (`#fef2f2` / `#dc2626`)
- Same logic for cross-team: many cross-team is a warning (amber), 0 cross-team could be neutral

**Acceptance**: Stat card colors reflect the actual severity of the metric value (0 unbacked = green, not red).

---

### UI-49 — Edit Model: Group action types into an accordion or categorized UI

**File**: `frontend/src/components/changeset/ActionForm.tsx`

**Problem**: 27 action types in a flat dropdown is overwhelming. Optgroup labels exist but the list is still too long.

**Fix**:
- Replace the flat `<select>` with a categorized button grid:
  - Group 1: Services (move_service, add_service, remove_service, link_to_service)
  - Group 2: Teams (add_team, remove_team, split_team, merge_teams, change_type, change_size)
  - Group 3: Capabilities (reassign_capability, change_visibility, add_capability, link_need_cap)
  - Group 4: Needs & Actors (add_need, remove_need, add_actor)
  - Group 5: Interactions & Other (add_interaction, remove_interaction, update_description)
- Each group is a collapsible section. Actions within each group are shown as labeled buttons.
- Clicking an action button selects it and shows its form fields below.
- Add a brief description under each action button: e.g., "Move a service from one team to another"

**Acceptance**: Action types are displayed in categorized groups with descriptions. Users can select any action without scrolling a long list.

---

### UI-50 — Edit Model: Add action type descriptions and tooltips

**File**: `frontend/src/components/changeset/ActionForm.tsx`

**Problem**: Action types like "Split Team", "Change Size", "Merge Teams" have no explanation.

**Fix**:
- Add a `description` field to each action type definition:
  | Action | Description |
  |--------|-------------|
  | `split_team` | "Create a new team and move selected services to it" |
  | `merge_teams` | "Merge two teams into one, combining all services" |
  | `change_type` | "Change the Team Topologies type of a team" |
  | `change_size` | "Update the cognitive size classification of a team" |
  | `reassign_capability` | "Move a capability's ownership to a different team" |
  | `change_visibility` | "Change a capability's visibility level in the value chain" |
- Show the description as subtitle text below the action name in the form header

**Acceptance**: Every action type shows a one-sentence description of what it does.

---

### UI-51 — Edit Model: Reposition "Export YAML" button

**File**: `frontend/src/pages/EditModelPage.tsx`

**Problem**: "Export YAML" button appears at the top before any changes, implying it exports the unchanged model.

**Fix**:
- Move "Export YAML" to appear in the Pending Changes section or after the Commit phase
- Label it clearly: "Export Current Model as YAML" (before any changes) or "Export Updated Model as YAML" (after commit)

**Acceptance**: Export YAML button position and label clearly indicate whether it exports the original or modified model.

---

### UI-52 — Edit Model: Add form validation for required fields

**File**: `frontend/src/components/changeset/ActionForm.tsx`

**Problem**: Clicking Add with empty required fields shows no error — the action either silently fails or adds an invalid entry.

**Fix**:
- Validate all required fields before calling `onAdd`
- Show inline red error messages under each empty required field: "This field is required"
- The Add button should be disabled OR show error messages — not silently fail

**Acceptance**: Submitting an incomplete action form shows specific error messages for each empty required field.

---

### UI-53 — What-If: Make suggestion chips context-aware

**File**: `frontend/src/pages/WhatIfPage.tsx`

**Problem**: Suggestion chips are static/hardcoded and don't reference the actual model's problems.

**Fix**:
- Generate suggestion chips from the model's signals/analysis results:
  - If there's an overloaded team: "How would you split [overloaded team name]?"
  - If there's a fragmented capability: "What if we consolidated [fragmented capability]?"
  - If there's a bottleneck service: "What's the impact of isolating [bottleneck service]?"
- Fall back to generic chips if no signals are available
- Limit to 4 chips maximum

**Acceptance**: At least one suggestion chip references an actual entity from the loaded model (e.g., actual team name, actual capability name).

---

### UI-54 — What-If: Show AI availability status

**File**: `frontend/src/pages/WhatIfPage.tsx`

**Problem**: No indication of whether AI is configured before the user tries to interact.

**Fix**:
- Use the existing `useAIEnabled()` hook
- If AI is disabled, show a banner at the top of the AI tab:
  ```
  ⚠ AI features are not configured. The AI Advisor requires an API key.
  [Learn how to configure AI →]
  ```
- Disable the Ask button and input when AI is not available

**Acceptance**: When AI is disabled, the AI tab clearly communicates this before the user tries to interact.

---

### UI-55 — AI Recommendations: Add cancel and progress to generation

**File**: `frontend/src/pages/RecommendationsPage.tsx`

**Problem**: 2-3 minute generation with no cancel option and no progress indicator.

**Fix**:
- Add a "Cancel" button next to the loading message that aborts the in-flight fetch
- Add a simple step progress indicator: "Analyzing team structure... → Checking capability coverage... → Generating report..."
- Steps can rotate every 15-20 seconds to give the appearance of progress
- Cache the last generated report in component state: show it immediately on page load, with a "Regenerate" button to get a fresh one

**Acceptance**:
- Cancel button is present during generation and aborts the request
- Progress text updates during the wait
- Returning to the page shows the cached report, not a blank loading state

---

### UI-56 — AI Advisor: Add AI status indicator

**File**: `frontend/src/pages/AdvisorPage.tsx`

**Problem**: No indication of whether AI is configured before interacting.

**Fix**:
- Same pattern as UI-54: use `useAIEnabled()` and show a banner when AI is disabled
- When AI is disabled, quick action buttons and the input should be disabled with a clear message

**Acceptance**: AI status is communicated before the user tries to interact.

---

### UI-57 — AI Advisor: Reduce category selector complexity

**File**: `frontend/src/pages/AdvisorPage.tsx`

**Problem**: 13 categories with no grouping or explanation overwhelm users.

**Fix**:
- Group categories:
  - Overview: General, Model Summary, Health Summary
  - Teams: Structural Load, Team Boundary, Interaction Mode
  - Architecture: Service Placement, Fragmentation, Bottleneck, Coupling
  - Delivery: Value Stream, Need Delivery Risk
  - Other: Free-form (rename from "Natural Language")
- Show grouped categories in a `<select>` with optgroups or in a collapsible section
- Add a brief description to each category on hover/focus

**Acceptance**: Categories are grouped logically. "Natural Language" is renamed to "Free-form".

---

## P2 — Cross-cutting Polish

---

### UI-58 — Global: Fix CSS keyframes leaking into rendered text

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx` (or wherever the spinner CSS is defined)

**Problem**: `@keyframes spin { to { transform: rotate(360deg) } }` is being rendered as visible text content somewhere in the Cognitive Load view.

**Fix**:
- Find the component that renders this CSS as text (likely a `<style>` tag content being included in a text node by mistake)
- Ensure CSS is in a proper `<style>` tag, a `.css` file, or a CSS-in-JS template — never concatenated into visible text content

**Acceptance**: No CSS code appears as visible text on any page.

---

### UI-59 — Cognitive Load: Fix "Healthy" vs "Low" terminology inconsistency

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx`

**Problem**: Summary says "5 Healthy" but team rows show "Low" load level. Two different words for the same thing.

**Fix**:
- Pick one term and use it consistently: recommend "Low" for the metric value (High/Medium/Low), and "Healthy" only as a summary label if both are needed
- Summary can say: "5 Low load (healthy)"

**Acceptance**: The same state is described with the same word throughout the Cognitive Load view.

---

### UI-60 — Cognitive Load: Add hover interaction to distribution bar

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx`

**Problem**: Distribution bar segments have no hover state showing which teams are in each category.

**Fix**:
- On hover over a bar segment, show a tooltip listing the team names in that load category:
  ```
  High load (3): inca-core-dev, inca-ingestion-dev, inca-publisher-dev
  ```

**Acceptance**: Hovering each distribution bar segment shows the names of teams in that category.

---

### UI-61 — Cognitive Load: Add active state to sort pills

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx`

**Problem**: No visual indication of which sort is currently active or its direction.

**Fix**:
- Active sort pill: solid background color, white text
- Add an arrow indicator (↑/↓) for sort direction
- Clicking the same active pill toggles asc/desc

**Acceptance**: Active sort is visually distinct. Direction arrow is shown.

---

### UI-62 — Cognitive Load: Visually separate sort from filter controls

**File**: `frontend/src/pages/views/CognitiveLoadView.tsx`

**Problem**: Sort pills (Severity, Name, Type) and filter chips (stream-aligned, platform, etc.) look visually identical.

**Fix**:
- Add a small "Sort:" label before sort controls and a "Filter:" label before filter chips
- Use a subtle visual separator (or different background colors) between the two groups

**Acceptance**: Users can distinguish sorting controls from filtering controls at a glance.

---

### UI-63 — Team Topology: Make "Via" capability names link to Capability View

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: Capability names in the "Via" column are plain text.

**Fix**:
- Render capability names in the Via column as links/buttons that navigate to `/capability` with that capability highlighted or selected

**Acceptance**: Clicking a capability name in the Via column navigates to the Capability View showing that capability.

---

### UI-64 — Upload: Remove "(debug)" from example button label

**File**: `frontend/src/pages/UploadPage.tsx`

**Problem**: "Load INCA Extended Example (debug)" looks like a dev artifact.

**Fix**:
- Change button text to "Load INCA Extended Example"

**Acceptance**: Button label has no "(debug)" suffix.

---

### UI-65 — Upload: Add helpful content below the upload form

**File**: `frontend/src/pages/UploadPage.tsx`

**Problem**: Large empty space below the upload form and example button (~40% of viewport unused).

**Fix**:
- Add a "Getting Started" section with:
  - Brief description of what UNM Platform does (1-2 sentences)
  - Accepted file formats (`.unm.yaml`, `.unm`)
  - Link to example model or documentation
  - "What is UNM?" expandable section with a 3-sentence explanation

**Acceptance**: The upload page's empty space is filled with useful onboarding content.

---

### UI-66 — Dashboard: Add "View in Signals" link to signal summary section

**File**: `frontend/src/pages/DashboardPage.tsx`

**Problem**: The signal summary section has no navigation path to the full Signals view (beyond a separate "Details" button).

**Fix**:
- Make the entire Architecture Signals card header a link to `/signals`
- The existing "Details ->" button is fine to keep; also make the card title clickable

**Acceptance**: Clicking the Architecture Signals card header navigates to `/signals`.

---

### UI-67 — Ownership View: Remove or hide quick edit buttons behind hover

**File**: `frontend/src/pages/views/OwnershipView.tsx`

**Problem**: Every entity has a quick edit pencil button causing visual clutter, especially on dense team cards.

**Fix**:
- Hide edit buttons by default
- Show them only when hovering over the entity row/card they belong to
- CSS: `.edit-btn { opacity: 0 } .entity-row:hover .edit-btn { opacity: 1 }`

**Acceptance**: Edit buttons are not visible by default; they appear on entity hover.

---

### UI-68 — Team Topology: Self-interaction row explanation

**File**: `frontend/src/pages/views/TeamTopologyView.tsx`

**Problem**: "inca-core-dev → X-as-a-Service → inca-core-dev" looks like a data error.

**Fix**:
- Add a small "(self)" badge next to rows where From and To are the same team
- Add a tooltip: "This team provides a service it also consumes internally"

**Acceptance**: Self-interaction rows have a "(self)" badge and explanatory tooltip.

---

### UI-69 — Realization View: Add "mapped %" stat card

**File**: `frontend/src/pages/views/RealizationView.tsx`

**Problem**: The stats section is missing a "mapped %" metric (all needs having at least one capability).

**Fix**:
- Add a "MAPPED" stat card showing: `{mappedCount}/{totalNeeds} ({percentage}%)`
- A need is "mapped" if it has at least one supporting capability

**Acceptance**: Stats section shows a "mapped" percentage metric.

---

### UI-70 — Global: Consistent "Ask AI" FAB positioning

**File**: `frontend/src/components/advisor/AdvisorPanel.tsx`

**Problem**: The floating "Ask AI" button may overlap bottom-right content on smaller viewports or content-dense pages.

**Fix**:
- Ensure the FAB is positioned at a fixed offset from the viewport bottom-right: `bottom: 24px; right: 24px`
- On pages where the FAB may cover content (Cognitive Load, Ownership with many team cards), verify by testing at 1280px width
- If content is covered, add `padding-bottom: 80px` to the main scroll container so content isn't hidden behind the FAB

**Acceptance**: The Ask AI FAB does not cover any interactive content at 1280px viewport width.

---

### UI-71 — Edit Model: Sort service and team dropdowns alphabetically

**File**: `frontend/src/components/changeset/ActionForm.tsx`

**Problem**: Service and team dropdowns are in non-alphabetical order.

**Fix**:
- Sort all entity dropdown options alphabetically by label before rendering

**Acceptance**: All entity dropdowns show options in alphabetical order.

---

## Signals Page (needs focused review and fixes)

The Signals page (`/signals`) was not fully tested due to browser tab contention during the parallel review. A focused review is needed. However, based on indirect observations:

---

### UI-72 — Signals View: Verify signal row expand/collapse works

**File**: `frontend/src/pages/views/SignalsView.tsx`

**Fix**:
- Manually test each signal row expand/collapse
- Verify expanded content shows: signal explanation text (non-empty), actionable suggestion, affected entities list
- Verify collapse animation has no glitches
- Fix any empty expanded rows

**Acceptance**: All signal rows expand to show meaningful explanation and suggestion text.

---

### UI-73 — Signals View: Health indicator cards need tooltips

**File**: `frontend/src/pages/views/SignalsView.tsx`

**Fix**:
- Add a tooltip to each health indicator card explaining how the score is calculated
- Example: "Score based on 3 critical and 1 medium signal in this layer"

**Acceptance**: Hovering a health indicator card shows what contributed to its score.

---

### UI-74 — Signals View: Verify AI Recommendation modal behavior

**File**: `frontend/src/pages/views/SignalsView.tsx`

**Fix**:
- Verify the AI Recommendation modal:
  - Opens when the button is clicked
  - Closes with Escape key
  - Closes when clicking outside the modal
  - Has scroll lock on body while open
  - Is not positioned off-screen
- Fix any failing behavior

**Acceptance**: AI Recommendation modal opens, closes, and positions correctly.

---

## Summary Table

| ID | Page | Priority | File(s) |
|----|------|----------|---------|
| UI-01 | UNM Map | P0 | `UNMMapView.tsx` |
| UI-02 | Edit Model | P0 | `ActionForm.tsx` |
| UI-03 | Edit Model | P0 | `ActionForm.tsx`, `ActionList.tsx` |
| UI-04 | Sidebar | P0 | `Sidebar.tsx` |
| UI-05 | Dashboard | P0 | `DashboardPage.tsx` |
| UI-06 | Global | P0 | Multiple view files |
| UI-07 | Need View | P0 | `NeedView.tsx`, `usePageInsights.ts` |
| UI-08 | Cognitive Load | P0 | `CognitiveLoadView.tsx` |
| UI-09 | Cognitive Load | P0 | `CognitiveLoadView.tsx` |
| UI-10 | Upload | P1 | `UploadPage.tsx` |
| UI-11 | Upload | P1 | `UploadPage.tsx` |
| UI-12 | Global | P1 | `TopBar.tsx` |
| UI-13 | Dashboard | P1 | `DashboardPage.tsx` |
| UI-14 | Dashboard | P1 | `DashboardPage.tsx` |
| UI-15 | Dashboard | P1 | `DashboardPage.tsx`, `SignalsView.tsx` |
| UI-16 | Dashboard | P1 | `DashboardPage.tsx` |
| UI-17 | Dashboard | P1 | `DashboardPage.tsx` |
| UI-18 | Dashboard | P1 | `DashboardPage.tsx`, `model-context.tsx` |
| UI-19 | UNM Map | P1 | `UNMMapView.tsx` |
| UI-20 | UNM Map | P1 | `UNMMapView.tsx` |
| UI-21 | UNM Map | P1 | `UNMMapView.tsx` |
| UI-22 | UNM Map | P1 | `UNMMapView.tsx` |
| UI-23 | Need View | P1 | `NeedView.tsx` |
| UI-24 | Need View | P1 | `NeedView.tsx` |
| UI-25 | Need View | P1 | `NeedView.tsx` |
| UI-26 | Capability View | P1 | `CapabilityView.tsx` |
| UI-27 | Capability View | P1 | `CapabilityView.tsx` |
| UI-28 | Capability View | P1 | `CapabilityView.tsx`, `UNMMapView.tsx` |
| UI-29 | Ownership | P1 | `OwnershipView.tsx` |
| UI-30 | Ownership | P1 | `OwnershipView.tsx` |
| UI-31 | Ownership | P1 | `OwnershipView.tsx` |
| UI-32 | Ownership | P1 | `OwnershipView.tsx` |
| UI-33 | Ownership | P1 | `OwnershipView.tsx` |
| UI-34 | Ownership | P1 | `OwnershipView.tsx` |
| UI-35 | Ownership | P1 | `OwnershipView.tsx` |
| UI-36 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-37 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-38 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-39 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-40 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-41 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-42 | Team Topology | P1 | `TeamTopologyView.tsx` |
| UI-43 | Realization | P1 | `RealizationView.tsx` |
| UI-44 | Realization | P1 | `RealizationView.tsx` |
| UI-45 | Realization | P1 | `RealizationView.tsx` |
| UI-46 | Realization | P1 | `RealizationView.tsx` |
| UI-47 | Realization | P1 | `RealizationView.tsx` |
| UI-48 | Realization | P1 | `RealizationView.tsx` |
| UI-49 | Edit Model | P1 | `ActionForm.tsx` |
| UI-50 | Edit Model | P1 | `ActionForm.tsx` |
| UI-51 | Edit Model | P1 | `EditModelPage.tsx` |
| UI-52 | Edit Model | P1 | `ActionForm.tsx` |
| UI-53 | What-If | P1 | `WhatIfPage.tsx` |
| UI-54 | What-If | P1 | `WhatIfPage.tsx` |
| UI-55 | AI Recommendations | P1 | `RecommendationsPage.tsx` |
| UI-56 | AI Advisor | P1 | `AdvisorPage.tsx` |
| UI-57 | AI Advisor | P1 | `AdvisorPage.tsx` |
| UI-58 | Cognitive Load | P2 | `CognitiveLoadView.tsx` |
| UI-59 | Cognitive Load | P2 | `CognitiveLoadView.tsx` |
| UI-60 | Cognitive Load | P2 | `CognitiveLoadView.tsx` |
| UI-61 | Cognitive Load | P2 | `CognitiveLoadView.tsx` |
| UI-62 | Cognitive Load | P2 | `CognitiveLoadView.tsx` |
| UI-63 | Team Topology | P2 | `TeamTopologyView.tsx` |
| UI-64 | Upload | P2 | `UploadPage.tsx` |
| UI-65 | Upload | P2 | `UploadPage.tsx` |
| UI-66 | Dashboard | P2 | `DashboardPage.tsx` |
| UI-67 | Ownership | P2 | `OwnershipView.tsx` |
| UI-68 | Team Topology | P2 | `TeamTopologyView.tsx` |
| UI-69 | Realization | P2 | `RealizationView.tsx` |
| UI-70 | Global | P2 | `AdvisorPanel.tsx` |
| UI-71 | Edit Model | P2 | `ActionForm.tsx` |
| UI-72 | Signals | P1 | `SignalsView.tsx` |
| UI-73 | Signals | P1 | `SignalsView.tsx` |
| UI-74 | Signals | P1 | `SignalsView.tsx` |
