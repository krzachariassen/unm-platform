package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
)

// ---------------------------------------------------------------------------
// Error-path tests for addActors
// ---------------------------------------------------------------------------

func TestYAMLParser_EmptyActorName(t *testing.T) {
	yaml := `
system:
  name: "Test"
actors:
  - name: ""
    description: "missing name"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "actor.name")
}

func TestYAMLParser_DuplicateActor(t *testing.T) {
	yaml := `
system:
  name: "Test"
actors:
  - name: "User"
  - name: "User"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Error-path tests for addNeeds
// ---------------------------------------------------------------------------

func TestYAMLParser_EmptyNeedName(t *testing.T) {
	yaml := `
system:
  name: "Test"
needs:
  - name: ""
    actor: "User"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "need.name")
}

func TestYAMLParser_MissingNeedActor(t *testing.T) {
	yaml := `
system:
  name: "Test"
needs:
  - name: "Pay"
    actor: ""
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "actor")
}

func TestYAMLParser_DuplicateNeed(t *testing.T) {
	yaml := `
system:
  name: "Test"
needs:
  - name: "Pay"
    actor: "User"
  - name: "Pay"
    actor: "Admin"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_NeedWithInvalidRelationshipTarget(t *testing.T) {
	yaml := `
system:
  name: "Test"
needs:
  - name: "Pay"
    actor: "User"
    supportedBy:
      - target: ""
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Error-path tests for addCapabilities
// ---------------------------------------------------------------------------

func TestYAMLParser_EmptyCapabilityName(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: ""
    description: "no name"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "capability.name")
}

func TestYAMLParser_DuplicateCapability(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Search"
  - name: "Search"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_InvalidCapabilityVisibility(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Search"
    visibility: "not-a-real-layer"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_NestedCapabilityEmptyChildName(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Parent"
    children:
      - name: ""
        description: "no child name"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_CapabilityWithInvalidRealizedByTarget(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Search"
    realizedBy:
      - target: ""
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_CapabilityWithInvalidDependsOnTarget(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Search"
    dependsOn:
      - target: ""
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Error-path tests for addServices
// ---------------------------------------------------------------------------

func TestYAMLParser_EmptyServiceName(t *testing.T) {
	yaml := `
system:
  name: "Test"
services:
  - name: ""
    ownedBy: "Team"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service.name")
}

func TestYAMLParser_DuplicateService(t *testing.T) {
	yaml := `
system:
  name: "Test"
services:
  - name: "svc"
    ownedBy: "Team"
  - name: "svc"
    ownedBy: "OtherTeam"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_ServiceWithInvalidDependsOnTarget(t *testing.T) {
	yaml := `
system:
  name: "Test"
services:
  - name: "svc"
    ownedBy: "Team"
    dependsOn:
      - target: ""
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Error-path tests for addTeams
// ---------------------------------------------------------------------------

func TestYAMLParser_EmptyTeamName(t *testing.T) {
	yaml := `
system:
  name: "Test"
teams:
  - name: ""
    type: "stream-aligned"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "team.name")
}

func TestYAMLParser_DuplicateTeam(t *testing.T) {
	yaml := `
system:
  name: "Test"
teams:
  - name: "Alpha"
    type: "stream-aligned"
  - name: "Alpha"
    type: "platform"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_TeamWithInvalidOwnsTarget(t *testing.T) {
	yaml := `
system:
  name: "Test"
teams:
  - name: "Alpha"
    type: "stream-aligned"
    owns:
      - target: ""
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Error-path tests for addPlatforms
// ---------------------------------------------------------------------------

func TestYAMLParser_EmptyPlatformName(t *testing.T) {
	yaml := `
system:
  name: "Test"
platforms:
  - name: ""
    teams:
      - "Alpha"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "platform.name")
}

func TestYAMLParser_DuplicatePlatform(t *testing.T) {
	yaml := `
system:
  name: "Test"
platforms:
  - name: "MyPlatform"
  - name: "MyPlatform"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Error-path tests for addInteractions
// ---------------------------------------------------------------------------

func TestYAMLParser_InvalidInteractionMode(t *testing.T) {
	yaml := `
system:
  name: "Test"
interactions:
  - from: "Team A"
    to: "Team B"
    mode: "not-a-mode"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interaction")
}

func TestYAMLParser_DataAssets(t *testing.T) {
	yaml := `
system:
  name: "Test"
data_assets:
  - name: "Catalog Data"
    type: "database"
    description: "Core catalog records"
    producedBy: "feed-svc"
    usedBy:
      - target: "serving-svc"
        access: "read"
    consumedBy:
      - "analytics-svc"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.Len(t, model.DataAssets, 1)
	da := model.DataAssets["Catalog Data"]
	require.NotNil(t, da)
	assert.Equal(t, "database", da.Type)
	assert.Equal(t, "feed-svc", da.ProducedBy)
	require.Len(t, da.UsedBy, 1)
	assert.Equal(t, "serving-svc", da.UsedBy[0].ServiceName)
	assert.Equal(t, "read", da.UsedBy[0].Access)
	assert.Contains(t, da.ConsumedBy, "analytics-svc")
}

func TestYAMLParser_DataAssetEmptyName(t *testing.T) {
	yaml := `
system:
  name: "Test"
data_assets:
  - name: ""
    type: "dataset"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data_asset.name")
}

func TestYAMLParser_DuplicateDataAsset(t *testing.T) {
	yaml := `
system:
  name: "Test"
data_assets:
  - name: "Catalog Data"
    type: "dataset"
  - name: "Catalog Data"
    type: "stream"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

func TestYAMLParser_ExternalDependencies(t *testing.T) {
	yaml := `
system:
  name: "Test"
external_dependencies:
  - name: "Stripe"
    description: "Payment processing"
    usedBy:
      - target: "payment-svc"
        description: "Charges customers via Stripe API"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.Len(t, model.ExternalDependencies, 1)
	ext := model.ExternalDependencies["Stripe"]
	require.NotNil(t, ext)
	assert.Equal(t, "Payment processing", ext.Description)
	require.Len(t, ext.UsedBy, 1)
	assert.Equal(t, "payment-svc", ext.UsedBy[0].ServiceName)
}

func TestYAMLParser_ExternalDependencyEmptyName(t *testing.T) {
	yaml := `
system:
  name: "Test"
external_dependencies:
  - name: ""
    description: "no name"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "external_dependency.name")
}

func TestYAMLParser_DuplicateExternalDependency(t *testing.T) {
	yaml := `
system:
  name: "Test"
external_dependencies:
  - name: "Stripe"
  - name: "Stripe"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// ParseFile error path
// ---------------------------------------------------------------------------

func TestParseFile_NonExistentFile(t *testing.T) {
	_, err := parser.ParseFile("/does/not/exist.unm.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open file")
}

// ---------------------------------------------------------------------------
// Invalid YAML
// ---------------------------------------------------------------------------

func TestYAMLParser_InvalidYAML(t *testing.T) {
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader("::invalid::yaml::"))
	require.Error(t, err)
}
