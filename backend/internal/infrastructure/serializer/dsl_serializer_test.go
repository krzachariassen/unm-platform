package serializer

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser/dsl"
)

func TestMarshalDSL_ProducesValidOutput(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)

	assert.Contains(t, out, `system "Test System"`)
	assert.Contains(t, out, `actor "Merchant"`)
	assert.Contains(t, out, `actor "Eater"`)
	assert.Contains(t, out, `capability "Cap One"`)
	assert.Contains(t, out, `visibility user-facing`)
	assert.Contains(t, out, `service "svc-a"`)
	assert.Contains(t, out, `ownedBy "Team Alpha"`)
	assert.Contains(t, out, `team "Team Alpha"`)
	assert.Contains(t, out, `type stream-aligned`)
	assert.Contains(t, out, `size 8`)
	assert.Contains(t, out, `need "Browse menu"`)
	assert.Contains(t, out, `interacts "Team Beta" mode collaboration`)
}

func TestMarshalDSL_RealizesOnServices(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)

	assert.Contains(t, out, `realizes "Cap One"`)
	assert.Contains(t, out, `realizes "Cap Two"`)
}

func TestMarshalDSL_ExternalDepsOnServices(t *testing.T) {
	m := entity.NewUNMModel("ExtDeps Test", "")

	svc, _ := entity.NewService("svc-x", "svc-x", "Service X", "Team X")
	_ = m.AddService(svc)

	extDep, _ := entity.NewExternalDependency("stripe", "Stripe", "Payment gateway")
	extDep.AddUsedBy("svc-x", "")
	_ = m.AddExternalDependency(extDep)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `externalDeps "Stripe"`)
}

func TestMarshalDSL_DeterministicOrder(t *testing.T) {
	m := buildExportTestModel(t)

	data1, _ := MarshalDSL(m)
	data2, _ := MarshalDSL(m)

	assert.Equal(t, string(data1), string(data2), "DSL output should be deterministic")
}

func TestMarshalDSL_RoundTrip(t *testing.T) {
	m := buildExportTestModel(t)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	f, err := dsl.Parse(string(data))
	require.NoError(t, err)

	m2, err := dsl.Transform(f)
	require.NoError(t, err)

	assert.Equal(t, m.System.Name, m2.System.Name)
	assert.Len(t, m2.Actors, len(m.Actors))
	assert.Len(t, m2.Needs, len(m.Needs))
	assert.Len(t, m2.Capabilities, len(m.Capabilities))
	assert.Len(t, m2.Services, len(m.Services))
	assert.Len(t, m2.Teams, len(m.Teams))
	assert.Len(t, m2.Interactions, len(m.Interactions))
}

func TestMarshalDSL_RoundTrip_Nexus(t *testing.T) {
	m, err := parser.ParseFile("../../../../examples/nexus.unm.yaml")
	require.NoError(t, err)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	f, err := dsl.Parse(string(data))
	require.NoError(t, err, "Generated DSL should be parseable:\n%s", string(data))

	m2, err := dsl.Transform(f)
	require.NoError(t, err)

	assert.Len(t, m2.Services, len(m.Services))
	assert.Len(t, m2.Capabilities, len(m.Capabilities))
	assert.Len(t, m2.Teams, len(m.Teams))
	assert.Len(t, m2.Interactions, len(m.Interactions))
}

func TestMarshalDSL_RoundTrip_DSLFile(t *testing.T) {
	p := parser.NewParserForPath("nexus.unm.yaml")
	src, err := os.Open("../../../../examples/nexus.unm.yaml")
	require.NoError(t, err)
	defer src.Close()
	m, err := p.Parse(src)
	require.NoError(t, err)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	parsed, err := dsl.Parse(string(data))
	require.NoError(t, err, "Generated DSL should be parseable")

	m2, err := dsl.Transform(parsed)
	require.NoError(t, err)

	assert.Equal(t, m.System.Name, m2.System.Name)
	assert.Len(t, m2.Services, len(m.Services))
	assert.Len(t, m2.Capabilities, len(m.Capabilities))
	assert.Len(t, m2.Teams, len(m.Teams))
	assert.Len(t, m2.Needs, len(m.Needs))
	assert.Len(t, m2.DataAssets, len(m.DataAssets))
	assert.Len(t, m2.ExternalDependencies, len(m.ExternalDependencies))
}

func TestMarshalDSL_MultiActorNeed(t *testing.T) {
	m := entity.NewUNMModel("Multi-Actor", "")

	a1, _ := entity.NewActor("dev", "Dev", "")
	a2, _ := entity.NewActor("ops", "Ops", "")
	_ = m.AddActor(&a1)
	_ = m.AddActor(&a2)

	need, _ := entity.NewNeedMultiActor("shared-need", "Shared Need", []string{"Dev", "Ops"}, "Both actors need this")
	_ = m.AddNeed(need)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `actor "Dev", "Ops"`)
}

func TestMarshalDSL_ExternalDepUsedByDescription(t *testing.T) {
	m := entity.NewUNMModel("ExtDesc Test", "")

	extDep, _ := entity.NewExternalDependency("db", "PostgreSQL", "Main database")
	extDep.AddUsedBy("svc-a", "Stores user data")
	_ = m.AddExternalDependency(extDep)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `usedBy "svc-a" : "Stores user data"`)
}

func TestMarshalDSL_DataAssets(t *testing.T) {
	m := entity.NewUNMModel("DA Test", "")

	da, _ := entity.NewDataAsset("events", "events", "event-stream", "Change events")
	da.AddUsedBy("svc-a")
	_ = m.AddDataAsset(da)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `data_asset "events"`)
	assert.Contains(t, out, `type event-stream`)
	assert.Contains(t, out, `usedBy "svc-a"`)
}

func TestMarshalDSL_CapabilityParent(t *testing.T) {
	m := entity.NewUNMModel("Parent Test", "")

	parent, _ := entity.NewCapability("cap-parent", "Parent Cap", "")
	parent.Visibility = "user-facing"
	child, _ := entity.NewCapability("cap-child", "Child Cap", "")
	parent.AddChild(child)
	_ = m.AddCapability(parent)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `capability "Parent Cap"`)
	assert.Contains(t, out, `capability "Child Cap"`)
}

func TestMarshalDSL_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("Empty", "")

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `system "Empty"`)
	assert.NotContains(t, out, "actor")
	assert.NotContains(t, out, "need")
}

func TestMarshalDSL_PlatformBlock(t *testing.T) {
	m := entity.NewUNMModel("Plat Test", "")

	team, _ := entity.NewTeam("team-a", "Team A", "", valueobject.Platform)
	_ = m.AddTeam(team)

	plat, _ := entity.NewPlatform("plat-1", "My Platform", "Platform desc")
	plat.AddTeam("Team A")
	_ = m.AddPlatform(plat)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `platform "My Platform"`)
	assert.Contains(t, out, `teams ["Team A"]`)
}

// DSL grammar does not support escape sequences in quoted strings.
// Embedded double quotes are replaced with single quotes to avoid breaking the grammar.
func TestMarshalDSL_QuotingPreservesContent(t *testing.T) {
	m := entity.NewUNMModel("System with 'quotes' & specials", "Description with special chars: <>&")

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	f, err := dsl.Parse(string(data))
	require.NoError(t, err)

	m2, err := dsl.Transform(f)
	require.NoError(t, err)

	assert.Equal(t, m.System.Name, m2.System.Name)
	assert.Equal(t, m.System.Description, m2.System.Description)
}

func TestMarshalDSL_EmbeddedDoubleQuotesReplacedWithSingleQuotes(t *testing.T) {
	m := entity.NewUNMModel(`Has "quotes" inside`, "")

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `"Has 'quotes' inside"`)

	f, err := dsl.Parse(out)
	require.NoError(t, err)

	m2, err := dsl.Transform(f)
	require.NoError(t, err)

	assert.Equal(t, "Has 'quotes' inside", m2.System.Name)
}

// TestMarshalDSL_RoundTrip_ICA tests round-trip through the inca model if it exists.
func TestMarshalDSL_RoundTrip_Inca(t *testing.T) {
	m, err := parser.ParseFile("../../../../examples/inca.unm.yaml")
	if err != nil {
		t.Skip("inca.unm.yaml not found — skipping")
	}

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	f, err := dsl.Parse(string(data))
	require.NoError(t, err, "Generated DSL should be parseable")

	m2, err := dsl.Transform(f)
	require.NoError(t, err)

	assert.Len(t, m2.Services, len(m.Services))
	assert.Len(t, m2.Capabilities, len(m.Capabilities))
	assert.Len(t, m2.Teams, len(m.Teams))
}

func TestMarshalDSL_VersionMetaRoundTrip(t *testing.T) {
	m := entity.NewUNMModel("Versioned", "A versioned model")
	m.Meta = entity.ModelMeta{
		Version:      "5",
		LastModified: "2026-03-17T10:00:00Z",
		Author:       "kristianz",
	}

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.Contains(t, out, `version "5"`)
	assert.Contains(t, out, `author "kristianz"`)
	assert.Contains(t, out, `lastModified "2026-03-17T10:00:00Z"`)

	// Round-trip: parse back and check all meta preserved
	f, err := dsl.Parse(out)
	require.NoError(t, err)

	m2, err := dsl.Transform(f)
	require.NoError(t, err)

	assert.Equal(t, "5", m2.Meta.Version)
	assert.Equal(t, "kristianz", m2.Meta.Author)
	assert.Equal(t, "2026-03-17T10:00:00Z", m2.Meta.LastModified)
}

func TestMarshalDSL_NoMetaWhenEmpty(t *testing.T) {
	m := entity.NewUNMModel("NoMeta", "")

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	out := string(data)
	assert.NotContains(t, out, "lastModified")
	assert.NotContains(t, out, "version")
	assert.NotContains(t, out, "author")
}

func TestMarshalYAML_VersionMetaRoundTrip(t *testing.T) {
	m := entity.NewUNMModel("Versioned", "")
	m.Meta = entity.ModelMeta{
		Version:      "3",
		LastModified: "2026-03-17T12:00:00Z",
		Author:       "test-user",
	}

	data, err := MarshalYAML(m)
	require.NoError(t, err)

	yamlStr := string(data)
	assert.Contains(t, yamlStr, "version: \"3\"")
	assert.Contains(t, yamlStr, "author: test-user")
	assert.Contains(t, yamlStr, "lastModified:")
}

func TestMarshalDSL_NoRelationshipDescDuplicate(t *testing.T) {
	m := entity.NewUNMModel("Desc Test", "")

	need, _ := entity.NewNeed("n1", "Need One", "Actor A", "outcome")
	capID, _ := valueobject.NewEntityID("cap-1")
	need.AddSupportedBy(entity.NewRelationship(capID, "through API", ""))
	_ = m.AddNeed(need)

	a, _ := entity.NewActor("a", "Actor A", "")
	_ = m.AddActor(&a)

	data, err := MarshalDSL(m)
	require.NoError(t, err)

	// Colon form for relationship with description but no role
	assert.Contains(t, string(data), `supportedBy "cap-1" : "through API"`)
	assert.Equal(t, 1, strings.Count(string(data), "through API"))
}
