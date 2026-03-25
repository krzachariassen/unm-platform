# Documentation Style Guide

## Voice and Tone

- **Direct**: say what something IS, not what it "aims to be" or "strives for"
- **Concrete**: use real examples, not abstract descriptions
- **Honest**: if something doesn't work yet, don't document it
- **Concise**: one clear sentence beats three vague ones

## Structure

### Headings
- H1 (`#`) -- document title only
- H2 (`##`) -- major sections
- H3 (`###`) -- subsections
- H4 (`####`) -- rarely needed, prefer restructuring

### Code Examples
- Every YAML example must be valid and parseable
- Show progression: minimal first, then add complexity
- Always show both short-form and long-form relationships when relevant
- Annotate with comments sparingly -- only for non-obvious elements
- Use fenced code blocks with language tags: `yaml`, `bash`, `go`, `typescript`

### Tables
- Use for structured comparisons (concepts, config options, commands)
- Keep columns narrow -- avoid paragraph-length cells
- Always include a header row

### Links
- Cross-reference related docs: `See [DSL Specification](docs/UNM_DSL_SPECIFICATION.md)`
- Use relative paths from project root
- Link to specific sections when possible: `[Visibility Levels](docs/UNM_DSL_SPECIFICATION.md#visibility)`

## Example Domain Rules

- NEVER use INCA as the example domain in public-facing documentation
- Use fictional but realistic domains from the approved list in AGENT.md
- Keep examples internally consistent (same actors, teams, services within one doc)
- Use realistic names: "booking-service" not "service-1"
- Include enough entities to demonstrate relationships (minimum: 2 actors, 3 needs,
  4 capabilities with hierarchy, 5 services, 3 teams, 2 interactions)

## README Conventions

- README.md must be self-contained for a 2-minute skim
- Quick Start must work copy-paste (test it)
- Link to detailed docs rather than duplicating content
- Update the project structure tree when directories change
- Keep the backlog table high-level -- link to full backlog

## Validation

Before submitting any documentation:
1. All YAML examples parse: `cd backend && go run ./cmd/cli/ parse <file>`
2. All bash commands work when copy-pasted
3. All internal links resolve to existing files
4. No references to features that don't exist yet
5. No INCA-specific examples in public docs
