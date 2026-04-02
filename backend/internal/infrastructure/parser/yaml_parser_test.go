package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
)

func TestYAMLParser_SimpleModel(t *testing.T) {
	p := parser.NewYAMLParser()
	model, err := parser.ParseFile("../../../testdata/simple.unm.yaml")
	require.NoError(t, err)
	require.NotNil(t, model)

	_ = p // parser is used via ParseFile

	assert.Equal(t, "Simple System", model.System.Name)
	assert.Equal(t, "A minimal test system", model.System.Description)

	// Actor
	require.Len(t, model.Actors, 1)
	actor, ok := model.Actors["User"]
	require.True(t, ok, "actor 'User' should be present")
	assert.Equal(t, "User", actor.Name)
	assert.Equal(t, "A basic user", actor.Description)

	// Need
	require.Len(t, model.Needs, 1)
	need, ok := model.Needs["Do something"]
	require.True(t, ok, "need 'Do something' should be present")
	assert.Equal(t, "User", need.ActorNames[0])
	assert.Equal(t, "Something is done", need.Outcome)
	require.Len(t, need.SupportedBy, 1)
	assert.Equal(t, "Core Capability", need.SupportedBy[0].TargetID.String())

	// Capability — realizedBy is now wired via service.realizes
	require.Len(t, model.Capabilities, 1)
	cap, ok := model.Capabilities["Core Capability"]
	require.True(t, ok, "capability 'Core Capability' should be present")
	assert.Equal(t, "Core Capability", cap.Name)
	require.Len(t, cap.RealizedBy, 1)
	assert.Equal(t, "core-service", cap.RealizedBy[0].TargetID.String())
	assert.Equal(t, "Main implementation", cap.RealizedBy[0].Description)

	// Service
	require.Len(t, model.Services, 1)
	svc, ok := model.Services["core-service"]
	require.True(t, ok, "service 'core-service' should be present")
	assert.Equal(t, "Core Team", svc.OwnerTeamName)

	// Team
	require.Len(t, model.Teams, 1)
	team, ok := model.Teams["Core Team"]
	require.True(t, ok, "team 'Core Team' should be present")
	assert.Equal(t, "stream-aligned", team.TeamType.String())
	require.Len(t, team.Owns, 1)
	assert.Equal(t, "Core Capability", team.Owns[0].TargetID.String())
}

func TestYAMLParser_RelationshipForms(t *testing.T) {
	model, err := parser.ParseFile("../../../testdata/relationships.unm.yaml")
	require.NoError(t, err)
	require.NotNil(t, model)

	// Short-form supportedBy
	shortNeed, ok := model.Needs["Short form need"]
	require.True(t, ok)
	require.Len(t, shortNeed.SupportedBy, 2)
	assert.Equal(t, "Search Capability", shortNeed.SupportedBy[0].TargetID.String())
	assert.Equal(t, "", shortNeed.SupportedBy[0].Description)
	assert.Equal(t, "Catalog Capability", shortNeed.SupportedBy[1].TargetID.String())

	// Long-form supportedBy
	longNeed, ok := model.Needs["Long form need"]
	require.True(t, ok)
	require.Len(t, longNeed.SupportedBy, 2)
	assert.Equal(t, "Search Capability", longNeed.SupportedBy[0].TargetID.String())
	assert.Equal(t, "Admin uses search to find records", longNeed.SupportedBy[0].Description)
	assert.Equal(t, "Catalog Capability", longNeed.SupportedBy[1].TargetID.String())
	assert.Equal(t, "Admin manages catalog entries", longNeed.SupportedBy[1].Description)

	// realizedBy is now wired via service.realizes
	searchCap, ok := model.Capabilities["Search Capability"]
	require.True(t, ok)
	require.Len(t, searchCap.RealizedBy, 2)
	// The order depends on service processing order
	svcNames := []string{searchCap.RealizedBy[0].TargetID.String(), searchCap.RealizedBy[1].TargetID.String()}
	assert.Contains(t, svcNames, "search-service")
	assert.Contains(t, svcNames, "search-index-service")

	// DependsOn (short form)
	require.Len(t, searchCap.DependsOn, 1)
	assert.Equal(t, "Catalog Capability", searchCap.DependsOn[0].TargetID.String())

	// Service with mixed dependsOn
	searchSvc, ok := model.Services["search-service"]
	require.True(t, ok)
	require.Len(t, searchSvc.DependsOn, 2)
	assert.Equal(t, "catalog-service", searchSvc.DependsOn[0].TargetID.String())
	assert.Equal(t, "", searchSvc.DependsOn[0].Description)
	assert.Equal(t, "search-index-service", searchSvc.DependsOn[1].TargetID.String())
	assert.Equal(t, "Reads from index", searchSvc.DependsOn[1].Description)

	// Verify GetCapabilitiesForService works (top-down) instead of checking service.Supports
	caps := model.GetCapabilitiesForService("search-service")
	capNames := make([]string, len(caps))
	for i, c := range caps {
		capNames[i] = c.Name
	}
	assert.Contains(t, capNames, "Search Capability")
}

func TestYAMLParser_ExampleModel(t *testing.T) {
	model, err := parser.ParseFile("../../../../examples/nexus.unm.yaml")
	require.NoError(t, err)
	require.NotNil(t, model)

	// System name
	assert.Equal(t, "Nexus", model.System.Name)
	assert.NotEmpty(t, model.System.Description)

	// 4 actors
	assert.Len(t, model.Actors, 4)
	assert.Contains(t, model.Actors, "Seller / Vendor Partner")
	assert.Contains(t, model.Actors, "Shopper / Buyer")
	assert.Contains(t, model.Actors, "Downstream Platform Team")
	assert.Contains(t, model.Actors, "Nexus Platform Engineer")

	// 11 needs
	assert.Len(t, model.Needs, 11)

	// 33 capabilities (flat map)
	assert.Len(t, model.Capabilities, 33)
	assert.Contains(t, model.Capabilities, "Seller Feed Management")
	assert.Contains(t, model.Capabilities, "Product Entity Management")
	assert.Contains(t, model.Capabilities, "Product Serving & Access")

	// 33 services
	assert.Len(t, model.Services, 33)

	// 9 teams
	assert.Len(t, model.Teams, 9)
	assert.Contains(t, model.Teams, "nexus-core-dev")
	assert.Contains(t, model.Teams, "nexus-ingestion-dev")
	assert.Contains(t, model.Teams, "nexus-serving-dev")
	assert.Contains(t, model.Teams, "partner-eng")

	// Interactions present (13 in YAML)
	assert.Len(t, model.Interactions, 13)
}

func TestYAMLParser_EmptyYAML(t *testing.T) {
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(""))
	assert.Error(t, err)
}

func TestYAMLParser_MissingSystemName(t *testing.T) {
	yaml := `
system:
  description: "missing name"
actors:
  - name: "User"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "system.name")
}

func TestYAMLParser_InvalidTeamType(t *testing.T) {
	yaml := `
system:
  name: "Test"
teams:
  - name: "Bad Team"
    type: "not-a-valid-type"
    description: "bad"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team")
}

func TestYAMLParser_ScenariosSectionIgnored(t *testing.T) {
	// Scenarios section was removed in model freeze (phase 10).
	// The parser should reject unknown YAML fields silently (YAML defaults).
	yaml := `
system:
  name: "Test"
actors:
  - name: "User"
needs:
  - name: "Do something"
    actor: "User"
    outcome: "Done"
    supportedBy:
      - "Core Cap"
capabilities:
  - name: "Core Cap"
    description: "A capability"
services:
  - name: "core-svc"
    ownedBy: "Core Team"
    realizes:
      - "Core Cap"
teams:
  - name: "Core Team"
    type: "stream-aligned"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.NotNil(t, model)

	// The Need should be parsed correctly.
	require.Len(t, model.Needs, 1)
	need := model.Needs["Do something"]
	require.NotNil(t, need)
	assert.Equal(t, "User", need.ActorNames[0])
	assert.Equal(t, "Done", need.Outcome)
}

func TestYAMLParser_MultiActorNeed(t *testing.T) {
	src := `
system:
  name: "Multi Actor Test"
actors:
  - name: "Actor A"
  - name: "Actor B"
needs:
  - name: "Shared Need"
    actor:
      - "Actor A"
      - "Actor B"
    outcome: "Something shared"
    supportedBy: []
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(src))
	require.NoError(t, err)
	require.NotNil(t, model)

	need, ok := model.Needs["Shared Need"]
	require.True(t, ok, "need 'Shared Need' should be present")
	assert.Len(t, need.ActorNames, 2)
	assert.Equal(t, "Actor A", need.ActorNames[0])
	assert.Equal(t, "Actor B", need.ActorNames[1])
	assert.Equal(t, "Something shared", need.Outcome)
}

func TestYAMLParser_SingleActorNeedStringForm(t *testing.T) {
	src := `
system:
  name: "Single Actor Test"
actors:
  - name: "Merchant"
needs:
  - name: "Accept Payment"
    actor: "Merchant"
    outcome: "Payment done"
    supportedBy: []
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(src))
	require.NoError(t, err)
	require.NotNil(t, model)

	need, ok := model.Needs["Accept Payment"]
	require.True(t, ok, "need 'Accept Payment' should be present")
	assert.Len(t, need.ActorNames, 1)
	assert.Equal(t, "Merchant", need.ActorNames[0])
}
