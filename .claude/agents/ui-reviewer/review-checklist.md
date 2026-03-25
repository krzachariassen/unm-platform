# UI Review Checklist

## Setup

The app runs on:
- Backend: http://localhost:8080
- Frontend: http://localhost:5173

Load the example model: click "Load INCA Extended Example (debug)" on the upload page.

## Per-Page Review Protocol

For each page:
1. Screenshot the full page (scroll if needed)
2. Click every interactive element
3. Hover over elements for tooltips and hover states
4. Open every panel, popup, modal, drawer and inspect content
5. Test edge cases: empty states, error states
6. Check responsive behavior

---

## Pages

### Upload Page (`/`)
- File drop zone affordance and drag hover state
- "Load INCA Extended Example" button progress
- Step indicator transitions (idle -> active -> done)
- Error handling on upload failure
- Smooth navigation after successful load

### Dashboard (`/dashboard`)
- Stats cards: correct numbers and labels
- Health summary rings/gauges rendering
- Signal summary bars: proportional, correct colors
- Validation section: errors vs warnings distinct
- "Explore Views" grid: correct links, hover states
- Cognitive Load teaser: meaningful data

### Signals View (`/signals`)
- Health indicator cards: correct data
- Layer sections: correct labels and groupings
- Signal row expansion: useful explanations
- AI Recommendation panel/modal: loads, closes correctly
- Alert icons MUST have text explanation (not icon-only)

### UNM Map View (`/unm-map`)
- Full rendering: all actors, needs, capabilities
- Visibility bands: correct colors and labels
- Legend: complete and accurate
- Zoom/pan: controls work, scroll wheel works
- Click every node type: detail panel opens with complete info
- Edit button: panel slides in, dropdowns populated, commit reloads map
- Detail panel hidden when edit panel open

### Need View (`/need`)
- Stats bar: total needs, unmapped count, percentages
- Per-actor sections: correct grouping
- Need expansion: capabilities, team info, anti-patterns
- "At risk" badges: reason on hover/click
- Search filtering

### Capability View (`/capability`)
- Toggle: Visibility vs Domain view
- Stat cards accuracy
- Capability card detail panel: complete info
- Fragmented capability indicators with explanation
- Visibility pills explained

### Ownership View (`/ownership`)
- Tab: Team view vs Domain view
- Filter toggles: cross-team, overloaded, unowned, problems-only
- Service pill popovers: rich content, correct position, dismiss on click-outside
- AntiPatternPanel: correct position, dismiss works

### Team Topology (`/team-topology`)
- Graph vs Table view toggle
- Graph: team nodes, interaction edges, detail drawer
- Drawer: close via X and backdrop
- Interaction mode checkboxes filtering
- Table: rows with QuickAction

### Cognitive Load View (`/cognitive-load`)
- Summary cards: high/medium/low counts
- Distribution bar: proportional, hover shows names
- Sort pills and team type filters
- Team card expansion: load dimensions, services, capabilities
- Gauge rendering at extremes

### Realization View (`/realization`)
- Tab: Value Chain vs By Service
- Filter pills: all, cross-team, unbacked
- Value chain expansion: capabilities and services
- Cross-team warnings with team names
- Capability chips clickable with info

### Edit Model Page (`/edit`)
- All dropdowns populated with model entities
- Action type grouping logical
- Each action type shows correct fields with dropdowns
- Add/remove/clear actions
- Preview impact and commit flow
- Export YAML

### What-If Explorer (`/what-if`)
- AI tab: chat interface, suggestion chips, markdown
- Manual tab: same smart dropdowns as Edit page
- Tab switching preserves state

---

## Cross-Cutting Checks

- Sidebar highlights current page correctly
- Sidebar items disabled when no model loaded
- TopBar search filters content on current page
- TopBar validation badges accurate and clickable
- Browser back/forward works (no blank pages)
- "Ask AI" button visible on all pages (when AI enabled)
- No orphaned loading spinners
- Consistent spacing, fonts, colors
- No cards showing only a name with no context
- No colored badges without legend or tooltip
