package serializer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
)

func buildExportTestModel(t *testing.T) *entity.UNMModel {
	t.Helper()
	m := entity.NewUNMModel("Test System", "A test system for YAML export")

	actor1, _ := entity.NewActor("merchant", "Merchant", "A merchant actor")
	_ = m.AddActor(&actor1)
	actor2, _ := entity.NewActor("eater", "Eater", "An eater actor")
	_ = m.AddActor(&actor2)

	teamA, _ := entity.NewTeam("team-alpha", "Team Alpha", "Alpha team", valueobject.StreamAligned)
	teamA.Size = 8
	teamA.SizeExplicit = true
	_ = m.AddTeam(teamA)

	teamB, _ := entity.NewTeam("team-beta", "Team Beta", "Beta team", valueobject.Platform)
	_ = m.AddTeam(teamB)

	svcA, _ := entity.NewService("svc-a", "svc-a", "Service A", "Team Alpha")
	_ = m.AddService(svcA)

	svcB, _ := entity.NewService("svc-b", "svc-b", "Service B", "Team Beta")
	_ = m.AddService(svcB)

	svcAID, _ := valueobject.NewEntityID("svc-a")
	svcBID, _ := valueobject.NewEntityID("svc-b")
	svcA.DependsOn = append(svcA.DependsOn, entity.NewRelationship(svcBID, "", ""))

	cap1, _ := entity.NewCapability("cap-one", "Cap One", "First capability")
	cap1.Visibility = "user-facing"
	cap1.RealizedBy = append(cap1.RealizedBy, entity.NewRelationship(svcAID, "Main impl", valueobject.Primary))
	_ = m.AddCapability(cap1)

	cap2, _ := entity.NewCapability("cap-two", "Cap Two", "Second capability")
	cap2.RealizedBy = append(cap2.RealizedBy, entity.NewRelationship(svcBID, "", ""))
	_ = m.AddCapability(cap2)

	capOneID, _ := valueobject.NewEntityID("cap-one")
	capTwoID, _ := valueobject.NewEntityID("cap-two")
	teamA.AddOwns(entity.NewRelationship(capOneID, "", ""))
	teamB.AddOwns(entity.NewRelationship(capTwoID, "", ""))

	need1, _ := entity.NewNeed("need-1", "Browse menu", "Eater", "See current menu")
	need1.SupportedBy = append(need1.SupportedBy, entity.NewRelationship(capOneID, "", ""))
	_ = m.AddNeed(need1)

	ix, _ := entity.NewInteraction("ix-1", "Team Alpha", "Team Beta", valueobject.Collaboration, "", "collab")
	m.AddInteraction(ix)

	return m
}

func TestMarshalYAML_ProducesValidOutput(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)

	assert.Contains(t, yamlStr, "name: Test System")
	assert.Contains(t, yamlStr, "name: Merchant")
	assert.Contains(t, yamlStr, "name: Eater")
	assert.Contains(t, yamlStr, "name: Cap One")
	assert.Contains(t, yamlStr, "visibility: user-facing")
	assert.Contains(t, yamlStr, "name: svc-a")
	assert.Contains(t, yamlStr, "ownedBy: Team Alpha")
	assert.Contains(t, yamlStr, "name: Team Alpha")
	assert.Contains(t, yamlStr, "type: stream-aligned")
	assert.Contains(t, yamlStr, "size: 8")
	assert.Contains(t, yamlStr, "name: Browse menu")
	assert.Contains(t, yamlStr, "with: Team Beta")
	assert.Contains(t, yamlStr, "mode: collaboration")
}

func TestMarshalYAML_ShortFormRelationships(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)
	// svc-a depends on svc-b with no description/role -> should be short form (just "- svc-b")
	assert.Contains(t, yamlStr, "- svc-b")
}

func TestMarshalYAML_LongFormRelationships(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)
	// v2: svc-a realizes cap-one with description and role (moved from capability.realizedBy to service.realizes)
	assert.Contains(t, yamlStr, "target: Cap One")
	assert.Contains(t, yamlStr, "description: Main impl")
	assert.Contains(t, yamlStr, "role: primary")
}

// ── v2 format tests ──────────────────────────────────────────────────────────

func TestMarshalYAML_V2_RealizesOnServices(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)

	// v2: services have realizes field, capabilities do NOT have realizedBy
	assert.Contains(t, yamlStr, "realizes:")
	assert.NotContains(t, yamlStr, "realizedBy:")

	// svc-a should have realizes: Cap One
	assert.Contains(t, yamlStr, "Cap One")
}

func TestMarshalYAML_V2_FlatCapabilities(t *testing.T) {
	m := entity.NewUNMModel("Hierarchy Test", "")

	parent, _ := entity.NewCapability("cap-parent", "Parent Cap", "A parent")
	child, _ := entity.NewCapability("cap-child", "Child Cap", "A child")
	parent.AddChild(child)
	_ = m.AddCapability(parent)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)

	// v2: both caps at top level (flat), child has parent field
	assert.Contains(t, yamlStr, "name: Child Cap")
	assert.Contains(t, yamlStr, "parent: Parent Cap")

	// v2: no nested children block
	assert.NotContains(t, yamlStr, "children:")
}

func TestMarshalYAML_V2_ExternalDepsOnServices(t *testing.T) {
	m := entity.NewUNMModel("ExtDeps Test", "")

	svc, _ := entity.NewService("svc-x", "svc-x", "Service X", "Team X")
	_ = m.AddService(svc)

	extDep, _ := entity.NewExternalDependency("stripe", "Stripe", "Payment gateway")
	extDep.AddUsedBy("svc-x", "")
	_ = m.AddExternalDependency(extDep)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)

	// v2: service has externalDeps list
	assert.Contains(t, yamlStr, "externalDeps:")
	assert.Contains(t, yamlStr, "Stripe")

	// v2: external_dependencies section does NOT have usedBy
	assert.NotContains(t, yamlStr, "usedBy:")
}

func TestMarshalYAML_V2_InteractsOnTeams(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)

	// v2: teams have interacts block
	assert.Contains(t, yamlStr, "interacts:")

	// v2: top-level interactions section is empty/omitted
	// The interaction is stored on the team, not in the top-level section
	// Check that "from: Team Alpha" is NOT in a top-level interactions block
	// (it may appear within team's interacts block as "with:" instead)
	assert.NotContains(t, yamlStr, "from: Team Alpha")
}

func TestMarshalYAML_V2_DeterministicOrder(t *testing.T) {
	m := buildExportTestModel(t)

	data1, _ := MarshalYAML(m)
	data2, _ := MarshalYAML(m)

	assert.Equal(t, string(data1), string(data2), "v2 YAML output should be deterministic")
}

func TestYAMLSerializer_RoundTrip_Nexus(t *testing.T) {
	m, err := parser.ParseFile("../../../../examples/nexus.unm.yaml")
	require.NoError(t, err)

	// Serialize to v2 YAML
	out, err := MarshalYAML(m)
	require.NoError(t, err)

	// Parse the v2 YAML back
	m2, err := parser.NewYAMLParser().Parse(strings.NewReader(string(out)))
	require.NoError(t, err)

	// Counts must match
	assert.Len(t, m2.Services, len(m.Services))
	assert.Len(t, m2.Capabilities, len(m.Capabilities))
	assert.Len(t, m2.Teams, len(m.Teams))
	assert.Len(t, m2.Interactions, len(m.Interactions))
}

func TestMarshalYAML_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("Empty", "")

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)
	assert.Contains(t, yamlStr, "name: Empty")
	assert.NotContains(t, yamlStr, "actors:")
	assert.NotContains(t, yamlStr, "needs:")
}

func TestMarshalYAML_DeterministicOrder(t *testing.T) {
	m := buildExportTestModel(t)

	data1, _ := MarshalYAML(m)
	data2, _ := MarshalYAML(m)

	assert.Equal(t, string(data1), string(data2), "YAML output should be deterministic")
}
