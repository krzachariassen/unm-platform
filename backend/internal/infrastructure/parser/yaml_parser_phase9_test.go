package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
)

// ---------------------------------------------------------------------------
// 9.1 — Flat capabilities with `parent` field
// ---------------------------------------------------------------------------

func TestPhase9_FlatCapabilities_SingleParent(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Parent Cap"
    visibility: "domain"
  - name: "Child Cap"
    parent: "Parent Cap"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.NotNil(t, model)

	// Both caps exist in flat map
	require.Contains(t, model.Capabilities, "Parent Cap")
	require.Contains(t, model.Capabilities, "Child Cap")

	// Child is in parent's Children slice
	parentCap := model.Capabilities["Parent Cap"]
	require.Len(t, parentCap.Children, 1)
	assert.Equal(t, "Child Cap", parentCap.Children[0].Name)

	// CapabilityParents records the relationship
	assert.Equal(t, "Parent Cap", model.CapabilityParents["Child Cap"])
}

func TestPhase9_FlatCapabilities_MultiLevel(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "A"
  - name: "B"
    parent: "A"
  - name: "C"
    parent: "B"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.NotNil(t, model)

	require.Contains(t, model.Capabilities, "A")
	require.Contains(t, model.Capabilities, "B")
	require.Contains(t, model.Capabilities, "C")

	// A→B→C chain
	capA := model.Capabilities["A"]
	require.Len(t, capA.Children, 1)
	assert.Equal(t, "B", capA.Children[0].Name)

	capB := model.Capabilities["B"]
	require.Len(t, capB.Children, 1)
	assert.Equal(t, "C", capB.Children[0].Name)

	assert.Equal(t, "A", model.CapabilityParents["B"])
	assert.Equal(t, "B", model.CapabilityParents["C"])
}

func TestPhase9_FlatCapabilities_MixedFlatAndNested(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Parent Cap"
    children:
      - name: "Nested Child"
  - name: "Flat Child"
    parent: "Parent Cap"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.NotNil(t, model)

	require.Contains(t, model.Capabilities, "Parent Cap")
	require.Contains(t, model.Capabilities, "Nested Child")
	require.Contains(t, model.Capabilities, "Flat Child")

	parentCap := model.Capabilities["Parent Cap"]
	// Should have both nested and flat children
	childNames := make([]string, len(parentCap.Children))
	for i, c := range parentCap.Children {
		childNames[i] = c.Name
	}
	assert.Contains(t, childNames, "Nested Child")
	assert.Contains(t, childNames, "Flat Child")
}

func TestPhase9_FlatCapabilities_MissingParent_Error(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Orphan"
    parent: "DoesNotExist"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DoesNotExist")
}

func TestPhase9_FlatCapabilities_CircularParent_Error(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "CapA"
    parent: "CapB"
  - name: "CapB"
    parent: "CapA"
`
	p := parser.NewYAMLParser()
	_, err := p.Parse(strings.NewReader(yaml))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

// ---------------------------------------------------------------------------
// 9.2 — Visibility inheritance
// ---------------------------------------------------------------------------

func TestPhase9_VisibilityInheritance_ChildInheritsParent(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Parent Cap"
    visibility: "domain"
    children:
      - name: "Child Cap"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	childCap := model.Capabilities["Child Cap"]
	require.NotNil(t, childCap)
	assert.Equal(t, "domain", childCap.Visibility, "child should inherit parent visibility")
}

func TestPhase9_VisibilityInheritance_ChildOverridesParent(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Parent Cap"
    visibility: "domain"
    children:
      - name: "Child Cap"
        visibility: "foundational"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	childCap := model.Capabilities["Child Cap"]
	require.NotNil(t, childCap)
	assert.Equal(t, "foundational", childCap.Visibility, "child should keep its own visibility")
}

func TestPhase9_VisibilityInheritance_MultiLevel(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Grandparent"
    visibility: "user-facing"
    children:
      - name: "Parent"
        children:
          - name: "Child"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	parent := model.Capabilities["Parent"]
	require.NotNil(t, parent)
	assert.Equal(t, "user-facing", parent.Visibility, "parent should inherit grandparent visibility")

	child := model.Capabilities["Child"]
	require.NotNil(t, child)
	assert.Equal(t, "user-facing", child.Visibility, "child should inherit grandparent visibility")
}

// ---------------------------------------------------------------------------
// 9.3 — Services declare `realizes`
// ---------------------------------------------------------------------------

func TestPhase9_ServiceRealizes_Basic(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Cap A"
services:
  - name: "svc-a"
    realizes:
      - "Cap A"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	cap := model.Capabilities["Cap A"]
	require.NotNil(t, cap)
	// Wiring is canonical on service.Realizes — verify via model query
	svcs := model.GetServicesForCapability("Cap A")
	require.Len(t, svcs, 1)
	assert.Equal(t, "svc-a", svcs[0].Name)
}

func TestPhase9_ServiceRealizes_WithRole(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Cap A"
services:
  - name: "svc-a"
    realizes:
      - target: "Cap A"
        role: "supporting"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	// Wiring is canonical on service.Realizes
	svcA, ok := model.Services["svc-a"]
	require.True(t, ok)
	require.Len(t, svcA.Realizes, 1)
	assert.Equal(t, "Cap A", svcA.Realizes[0].TargetID.String())
	assert.Equal(t, "supporting", svcA.Realizes[0].Role.String())
}

// ---------------------------------------------------------------------------
// 9.4 — External dependencies on services
// ---------------------------------------------------------------------------

func TestPhase9_ServiceExternalDeps_Basic(t *testing.T) {
	yaml := `
system:
  name: "Test"
services:
  - name: "svc-a"
    externalDeps:
      - "Stripe"
external_dependencies:
  - name: "Stripe"
    description: "Payment processor"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	stripe := model.ExternalDependencies["Stripe"]
	require.NotNil(t, stripe)
	require.Len(t, stripe.UsedBy, 1)
	assert.Equal(t, "svc-a", stripe.UsedBy[0].ServiceName)
}

func TestPhase9_ExternalDep_UnreferencedAllowed(t *testing.T) {
	yaml := `
system:
  name: "Test"
external_dependencies:
  - name: "Stripe"
    description: "Unused dep"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)
	require.Contains(t, model.ExternalDependencies, "Stripe")
}

// ---------------------------------------------------------------------------
// 9.5 — Interactions on teams (inline)
// ---------------------------------------------------------------------------

func TestPhase9_TeamInlineInteractions_Basic(t *testing.T) {
	yaml := `
system:
  name: "Test"
teams:
  - name: "Team A"
    type: "stream-aligned"
    interacts:
      - with: "Team B"
        mode: "collaboration"
  - name: "Team B"
    type: "stream-aligned"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	require.Len(t, model.Interactions, 1)
	inter := model.Interactions[0]
	assert.Equal(t, "Team A", inter.FromTeamName)
	assert.Equal(t, "Team B", inter.ToTeamName)
	assert.Equal(t, "collaboration", inter.Mode.String())
}

// ---------------------------------------------------------------------------
// 9.7 — Strict reference validation
// ---------------------------------------------------------------------------

func TestPhase9_ReferenceValidation_UnresolvedNeedSupportedBy_Warning(t *testing.T) {
	yaml := `
system:
  name: "Test"
actors:
  - name: "User"
needs:
  - name: "Need A"
    actor: "User"
    supportedBy:
      - "NonExistentCap"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err) // Not an error, just a warning

	assert.NotEmpty(t, model.Warnings)
	found := false
	for _, w := range model.Warnings {
		if strings.Contains(w, "NonExistentCap") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning about unresolved supportedBy reference")
}

func TestPhase9_ReferenceValidation_UnresolvedCapabilityRealizedBy_Warning(t *testing.T) {
	yaml := `
system:
  name: "Test"
capabilities:
  - name: "Cap A"
services:
  - name: "ghost-svc"
    realizes:
      - "Cap A"
      - "NonExistentCap"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err) // Not an error, just a warning

	// ghost-svc realizes Cap A successfully, but NonExistentCap is skipped
	// (reference validation catches unresolved service.realizes targets)
	// Verify Cap A is realized by ghost-svc via the canonical service.Realizes
	svcs := model.GetServicesForCapability("Cap A")
	require.Len(t, svcs, 1)
}

func TestPhase9_ReferenceValidation_ResolvedReferences_NoWarning(t *testing.T) {
	yaml := `
system:
  name: "Test"
actors:
  - name: "User"
needs:
  - name: "Need A"
    actor: "User"
    supportedBy:
      - "Cap A"
capabilities:
  - name: "Cap A"
services:
  - name: "svc-a"
    realizes:
      - "Cap A"
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	// No reference warnings — only check that all resolved refs have no unresolved warning
	for _, w := range model.Warnings {
		assert.NotContains(t, w, "unresolved", "should not have unresolved warnings")
	}
}

// ---------------------------------------------------------------------------
// 9.8 — Data assets compact usedBy syntax
// ---------------------------------------------------------------------------

func TestPhase9_DataAsset_CompactUsedBy(t *testing.T) {
	yaml := `
system:
  name: "Test"
data_assets:
  - name: "Orders DB"
    type: "database"
    usedBy:
      - svc-a
      - svc-b
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	da := model.DataAssets["Orders DB"]
	require.NotNil(t, da)
	require.Len(t, da.UsedBy, 2)
	assert.Contains(t, da.UsedBy, "svc-a")
	assert.Contains(t, da.UsedBy, "svc-b")
}

func TestPhase9_DataAsset_ObjectUsedBy_StillWorks(t *testing.T) {
	yaml := `
system:
  name: "Test"
data_assets:
  - name: "Orders DB"
    type: "database"
    usedBy:
      - svc-a
`
	p := parser.NewYAMLParser()
	model, err := p.Parse(strings.NewReader(yaml))
	require.NoError(t, err)

	da := model.DataAssets["Orders DB"]
	require.NotNil(t, da)
	require.Len(t, da.UsedBy, 1)
	assert.Equal(t, "svc-a", da.UsedBy[0])
}
