package service_test

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// helpers

func mustNewEntityID(t *testing.T, s string) valueobject.EntityID {
	t.Helper()
	id, err := valueobject.NewEntityID(s)
	if err != nil {
		t.Fatalf("NewEntityID(%q): %v", s, err)
	}
	return id
}

func mustNewRelationship(t *testing.T, targetName string) entity.Relationship {
	t.Helper()
	id := mustNewEntityID(t, targetName)
	return entity.NewRelationship(id, "", valueobject.RelationshipRole(""))
}

func mustNewNeed(t *testing.T, id, name, actor, outcome string) *entity.Need {
	t.Helper()
	n, err := entity.NewNeed(id, name, actor, outcome)
	if err != nil {
		t.Fatalf("NewNeed: %v", err)
	}
	return n
}

func mustNewCapability(t *testing.T, id, name, desc string) *entity.Capability {
	t.Helper()
	c, err := entity.NewCapability(id, name, desc)
	if err != nil {
		t.Fatalf("NewCapability: %v", err)
	}
	return c
}

func mustNewService(t *testing.T, id, name, owner string) *entity.Service {
	t.Helper()
	s, err := entity.NewService(id, name, "", owner)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return s
}

func mustNewTeam(t *testing.T, id, name string, tt valueobject.TeamType) *entity.Team {
	t.Helper()
	tm, err := entity.NewTeam(id, name, "", tt)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	return tm
}

func mustNewInteraction(t *testing.T, id, from, to string, mode valueobject.InteractionMode) *entity.Interaction {
	t.Helper()
	i, err := entity.NewInteraction(id, from, to, mode, "", "")
	if err != nil {
		t.Fatalf("NewInteraction: %v", err)
	}
	return i
}

func mustNewInferredMapping(t *testing.T, id, svc, cap string, score float64) *entity.InferredMapping {
	t.Helper()
	conf, err := valueobject.NewConfidence(score, "")
	if err != nil {
		t.Fatalf("NewConfidence: %v", err)
	}
	im, err := entity.NewInferredMapping(id, svc, cap, conf, valueobject.Inferred)
	if err != nil {
		t.Fatalf("NewInferredMapping: %v", err)
	}
	return im
}

func hasError(result service.ValidationResult, code service.ValidationErrorCode) bool {
	for _, e := range result.Errors {
		if e.Code == code {
			return true
		}
	}
	return false
}

func hasWarning(result service.ValidationResult, code service.ValidationWarningCode) bool {
	for _, w := range result.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

// buildValidModel constructs a minimal valid model:
// - one actor, one need (supported by one capability)
// - one leaf capability (realized by one service)
// - one service (owned by one team)
// - one team
// Capability.RealizedBy is the canonical source of truth (top-down).
func buildValidModel(t *testing.T) *entity.UNMModel {
	t.Helper()
	m := entity.NewUNMModel("TestSystem", "")

	actor, err := entity.NewActor("actor-1", "Merchant", "")
	if err != nil {
		t.Fatalf("NewActor: %v", err)
	}
	if err := m.AddActor(&actor); err != nil {
		t.Fatalf("AddActor: %v", err)
	}

	need := mustNewNeed(t, "need-1", "Accept Payments", "Merchant", "")
	need.AddSupportedBy(mustNewRelationship(t, "Payment Processing"))
	if err := m.AddNeed(need); err != nil {
		t.Fatalf("AddNeed: %v", err)
	}

	cap := mustNewCapability(t, "cap-1", "Payment Processing", "")
	// Capability.RealizedBy is canonical; service is referenced here.
	cap.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	if err := m.AddCapability(cap); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	svc := mustNewService(t, "svc-1", "payment-service", "payments-team")
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	team := mustNewTeam(t, "team-1", "payments-team", valueobject.StreamAligned)
	team.SizeExplicit = true
	team.AddOwns(mustNewRelationship(t, "Payment Processing"))
	if err := m.AddTeam(team); err != nil {
		t.Fatalf("AddTeam: %v", err)
	}

	interaction, err := entity.NewInteraction("int-1", "payments-team", "payments-team",
		valueobject.Collaboration, "", "self-ref for coverage")
	if err != nil {
		t.Fatalf("NewInteraction: %v", err)
	}
	m.AddInteraction(interaction)

	return m
}

// ── Test 1: valid minimal model ──────────────────────────────────────────────

func TestValidate_ValidModel_NoErrorsNoWarnings(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if !result.IsValid() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
	if result.HasWarnings() {
		t.Errorf("expected no warnings, got: %v", result.Warnings)
	}
}

// ── Test 2: ErrNeedNoCapability ──────────────────────────────────────────────

func TestValidate_NeedNoCapability_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	unmapped := mustNewNeed(t, "need-2", "Track Order", "Merchant", "")
	// intentionally do NOT add SupportedBy
	if err := m.AddNeed(unmapped); err != nil {
		t.Fatalf("AddNeed: %v", err)
	}

	result := engine.Validate(m)
	if !hasError(result, service.ErrNeedNoCapability) {
		t.Errorf("expected ErrNeedNoCapability, got errors: %v", result.Errors)
	}
}

func TestValidate_NeedWithCapability_NoError(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if hasError(result, service.ErrNeedNoCapability) {
		t.Errorf("did not expect ErrNeedNoCapability for valid model")
	}
}

// ── Test 3: ErrLeafCapNoService ───────────────────────────────────────────────

func TestValidate_LeafCapNoService_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	unrealized := mustNewCapability(t, "cap-2", "Refund Processing", "")
	// no RealizedBy — leaf with no service
	if err := m.AddCapability(unrealized); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	result := engine.Validate(m)
	if !hasError(result, service.ErrLeafCapNoService) {
		t.Errorf("expected ErrLeafCapNoService, got errors: %v", result.Errors)
	}
}

func TestValidate_LeafCapWithService_NoError(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if hasError(result, service.ErrLeafCapNoService) {
		t.Errorf("did not expect ErrLeafCapNoService for valid model")
	}
}

func TestValidate_NonLeafCap_NoLeafServiceError(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// parent cap with no RealizedBy — but not a leaf, so should NOT trigger ErrLeafCapNoService
	parent := mustNewCapability(t, "cap-parent", "Payments Suite", "")
	child := mustNewCapability(t, "cap-child", "Tokenisation", "")
	child.AddRealizedBy(mustNewRelationship(t, "token-svc"))
	parent.AddChild(child)
	if err := m.AddCapability(parent); err != nil {
		t.Fatalf("AddCapability parent: %v", err)
	}

	// add a service so Tokenisation child is realized
	svc := mustNewService(t, "svc-token", "token-svc", "payments-team")
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	result := engine.Validate(m)
	// parent is not a leaf — should not produce the error for "Payments Suite"
	for _, e := range result.Errors {
		if e.Code == service.ErrLeafCapNoService && e.Entity == "Payments Suite" {
			t.Errorf("should not flag non-leaf capability 'Payments Suite' with ErrLeafCapNoService")
		}
	}
}

// ── Test 4: ErrServiceNoOwner ─────────────────────────────────────────────────

func TestValidate_ServiceNoOwner_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	orphanSvc := mustNewService(t, "svc-2", "orphan-service", "some-team")
	orphanSvc.OwnerTeamName = "" // forcibly clear
	// Add to a capability's RealizedBy so we don't get WarnOrphanService too
	cap := m.Capabilities["Payment Processing"]
	cap.AddRealizedBy(mustNewRelationship(t, "orphan-service"))
	if err := m.AddService(orphanSvc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	result := engine.Validate(m)
	if !hasError(result, service.ErrServiceNoOwner) {
		t.Errorf("expected ErrServiceNoOwner, got errors: %v", result.Errors)
	}
}

func TestValidate_ServiceWithOwner_NoError(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if hasError(result, service.ErrServiceNoOwner) {
		t.Errorf("did not expect ErrServiceNoOwner for valid model")
	}
}

// ── Test 5: ErrInvalidInteraction ─────────────────────────────────────────────

func TestValidate_InteractionUnknownTeam_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Interaction referencing a team that doesn't exist in model
	badInteraction := mustNewInteraction(t, "int-2", "payments-team", "ghost-team",
		valueobject.Collaboration)
	m.AddInteraction(badInteraction)

	result := engine.Validate(m)
	if !hasError(result, service.ErrInvalidInteraction) {
		t.Errorf("expected ErrInvalidInteraction, got errors: %v", result.Errors)
	}
}

func TestValidate_InteractionEmptyMode_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	badInteraction := &entity.Interaction{
		FromTeamName: "payments-team",
		ToTeamName:   "payments-team",
		Mode:         valueobject.InteractionMode(""),
	}
	// Give it an ID
	id, _ := valueobject.NewEntityID("int-bad")
	badInteraction.ID = id

	m.AddInteraction(badInteraction)

	result := engine.Validate(m)
	if !hasError(result, service.ErrInvalidInteraction) {
		t.Errorf("expected ErrInvalidInteraction for empty mode, got errors: %v", result.Errors)
	}
}

func TestValidate_ValidInteraction_NoError(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	// buildValidModel already adds a self-referencing interaction with valid mode
	result := engine.Validate(m)
	if hasError(result, service.ErrInvalidInteraction) {
		t.Errorf("did not expect ErrInvalidInteraction for valid model")
	}
}

// ── Test 6: ErrInvalidConfidence ──────────────────────────────────────────────

func TestValidate_InvalidConfidence_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Construct with valid score, then forcibly set to invalid value
	im := mustNewInferredMapping(t, "im-1", "payment-service", "Payment Processing", 0.9)
	im.Confidence = valueobject.Confidence{Score: 1.5} // bypass constructor validation
	m.AddInferredMapping(im)

	result := engine.Validate(m)
	if !hasError(result, service.ErrInvalidConfidence) {
		t.Errorf("expected ErrInvalidConfidence, got errors: %v", result.Errors)
	}
}

func TestValidate_NegativeConfidence_Error(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	im := mustNewInferredMapping(t, "im-2", "payment-service", "Payment Processing", 0.8)
	im.Confidence = valueobject.Confidence{Score: -0.1}
	m.AddInferredMapping(im)

	result := engine.Validate(m)
	if !hasError(result, service.ErrInvalidConfidence) {
		t.Errorf("expected ErrInvalidConfidence, got errors: %v", result.Errors)
	}
}

func TestValidate_ValidConfidence_NoError(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	im := mustNewInferredMapping(t, "im-3", "payment-service", "Payment Processing", 0.9)
	m.AddInferredMapping(im)
	result := engine.Validate(m)
	if hasError(result, service.ErrInvalidConfidence) {
		t.Errorf("did not expect ErrInvalidConfidence for score=0.9")
	}
}

// ── Test 7: WarnFragmentation ─────────────────────────────────────────────────

func TestValidate_FragmentedCapability_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Add two more teams that also own "Payment Processing" → 3 owners total
	for i, name := range []string{"team-alpha", "team-beta"} {
		tm := mustNewTeam(t, name, name, valueobject.StreamAligned)
		tm.AddOwns(mustNewRelationship(t, "Payment Processing"))
		if err := m.AddTeam(tm); err != nil {
			t.Fatalf("AddTeam[%d]: %v", i, err)
		}
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnFragmentation) {
		t.Errorf("expected WarnFragmentation, got warnings: %v", result.Warnings)
	}
}

func TestValidate_NonFragmentedCapability_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if hasWarning(result, service.WarnFragmentation) {
		t.Errorf("did not expect WarnFragmentation for valid model")
	}
}

// ── Test 8: WarnCognitiveLoad ──────────────────────────────────────────────────

func TestValidate_OverloadedTeam_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Add 6 more capabilities owned by the same team → 7 total
	team, ok := m.Teams["payments-team"]
	if !ok {
		t.Fatal("payments-team not found")
	}

	for i := 0; i < 6; i++ {
		capName := "Extra Cap" + string(rune('A'+i))
		capID := "cap-extra-" + string(rune('a'+i))
		svcName := "extra-svc-" + string(rune('a'+i))
		svcID := "svc-extra-" + string(rune('a'+i))
		cap := mustNewCapability(t, capID, capName, "")
		cap.AddRealizedBy(mustNewRelationship(t, svcName))
		if err := m.AddCapability(cap); err != nil {
			t.Fatalf("AddCapability extra: %v", err)
		}
		svc := mustNewService(t, svcID, svcName, "payments-team")
		if err := m.AddService(svc); err != nil {
			t.Fatalf("AddService extra: %v", err)
		}
		team.AddOwns(mustNewRelationship(t, capName))
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnCognitiveLoad) {
		t.Errorf("expected WarnCognitiveLoad, got warnings: %v", result.Warnings)
	}
}

func TestValidate_NonOverloadedTeam_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if hasWarning(result, service.WarnCognitiveLoad) {
		t.Errorf("did not expect WarnCognitiveLoad for valid model")
	}
}

// ── Test: WarnTeamSizeUnset ────────────────────────────────────────────────────

// TestValidate_TeamSizeUnset_Warning checks that a team without SizeExplicit
// produces a WarnTeamSizeUnset warning — structural load assessment uses a
// default of 5 people which may be wrong.
func TestValidate_TeamSizeUnset_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// NewTeam sets Size=5 but SizeExplicit=false by default (parser didn't set it).
	team, ok := m.Teams["payments-team"]
	if !ok {
		t.Fatal("payments-team not found")
	}
	team.SizeExplicit = false

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnTeamSizeUnset) {
		t.Errorf("expected WarnTeamSizeUnset when SizeExplicit is false, got warnings: %v", result.Warnings)
	}
}

// TestValidate_TeamSizeSet_NoWarning verifies that a team with an explicit
// size does NOT produce a WarnTeamSizeUnset warning.
func TestValidate_TeamSizeSet_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Mark all teams as having explicit size.
	for _, team := range m.Teams {
		team.SizeExplicit = true
	}

	result := engine.Validate(m)
	if hasWarning(result, service.WarnTeamSizeUnset) {
		t.Errorf("did not expect WarnTeamSizeUnset when SizeExplicit is true")
	}
}

// ── Test 9: WarnOrphanService ─────────────────────────────────────────────────

func TestValidate_OrphanService_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Service not referenced in any capability's RealizedBy → orphan
	orphan := mustNewService(t, "svc-orphan", "orphan-svc", "payments-team")
	if err := m.AddService(orphan); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnOrphanService) {
		t.Errorf("expected WarnOrphanService, got warnings: %v", result.Warnings)
	}
}

func TestValidate_NonOrphanService_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	result := engine.Validate(m)
	if hasWarning(result, service.WarnOrphanService) {
		t.Errorf("did not expect WarnOrphanService for valid model")
	}
}

// ── Test 10: WarnCircularDep ───────────────────────────────────────────────────

func TestValidate_CircularDependency_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// cap-a → depends on cap-b → depends on cap-a (cycle)
	capA := mustNewCapability(t, "cap-a", "Cap A", "")
	capB := mustNewCapability(t, "cap-b", "Cap B", "")
	capA.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	capB.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	capA.AddDependsOn(mustNewRelationship(t, "Cap B"))
	capB.AddDependsOn(mustNewRelationship(t, "Cap A"))

	if err := m.AddCapability(capA); err != nil {
		t.Fatalf("AddCapability capA: %v", err)
	}
	if err := m.AddCapability(capB); err != nil {
		t.Fatalf("AddCapability capB: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnCircularDep) {
		t.Errorf("expected WarnCircularDep, got warnings: %v", result.Warnings)
	}
}

func TestValidate_NoCircularDependency_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)
	// buildValidModel has one capability with no DependsOn
	result := engine.Validate(m)
	if hasWarning(result, service.WarnCircularDep) {
		t.Errorf("did not expect WarnCircularDep for valid model")
	}
}

func TestValidate_LinearDependency_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// cap-x → depends on cap-y (no cycle)
	capX := mustNewCapability(t, "cap-x", "Cap X", "")
	capY := mustNewCapability(t, "cap-y", "Cap Y", "")
	capX.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	capY.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	capX.AddDependsOn(mustNewRelationship(t, "Cap Y"))

	if err := m.AddCapability(capX); err != nil {
		t.Fatalf("AddCapability capX: %v", err)
	}
	if err := m.AddCapability(capY); err != nil {
		t.Fatalf("AddCapability capY: %v", err)
	}

	result := engine.Validate(m)
	if hasWarning(result, service.WarnCircularDep) {
		t.Errorf("did not expect WarnCircularDep for linear dependency")
	}
}

// ── Test 11: WarnLowConfidence ─────────────────────────────────────────────────

func TestValidate_LowConfidence_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	im := mustNewInferredMapping(t, "im-low", "payment-service", "Payment Processing", 0.3)
	m.AddInferredMapping(im)

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnLowConfidence) {
		t.Errorf("expected WarnLowConfidence, got warnings: %v", result.Warnings)
	}
}

func TestValidate_HighConfidence_NoLowWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	im := mustNewInferredMapping(t, "im-high", "payment-service", "Payment Processing", 0.8)
	m.AddInferredMapping(im)

	result := engine.Validate(m)
	if hasWarning(result, service.WarnLowConfidence) {
		t.Errorf("did not expect WarnLowConfidence for score=0.8")
	}
}

// ── Test 12: WarnParentCapHasServices ──────────────────────────────────────────

func TestValidate_ParentCapWithServices_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Parent capability with RealizedBy entries (should warn)
	parent := mustNewCapability(t, "cap-parent", "Parent Cap", "")
	child := mustNewCapability(t, "cap-child2", "Child Cap", "")
	child.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	parent.AddChild(child)
	// also add RealizedBy on the parent — this should trigger the warning
	parent.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	if err := m.AddCapability(parent); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnParentCapHasServices) {
		t.Errorf("expected WarnParentCapHasServices, got warnings: %v", result.Warnings)
	}
}

func TestValidate_ParentCapNoServices_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Parent with no RealizedBy — only child has services
	parent := mustNewCapability(t, "cap-parent2", "Parent Cap 2", "")
	child := mustNewCapability(t, "cap-child3", "Child Cap 3", "")
	child.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	parent.AddChild(child)
	if err := m.AddCapability(parent); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	result := engine.Validate(m)
	for _, w := range result.Warnings {
		if w.Code == service.WarnParentCapHasServices && w.Entity == "Parent Cap 2" {
			t.Errorf("did not expect WarnParentCapHasServices for parent with no RealizedBy")
		}
	}
}

// ── Test 13: WarnNonPlatformTeamInPlatform ──────────────────────────────────────

func TestValidate_NonPlatformTeamInPlatform_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// payments-team is stream-aligned, add it to a platform
	plat, _ := entity.NewPlatform("plat-1", "MyPlatform", "")
	plat.AddTeam("payments-team") // stream-aligned, not platform type
	if err := m.AddPlatform(plat); err != nil {
		t.Fatalf("AddPlatform: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnNonPlatformTeamInPlatform) {
		t.Errorf("expected WarnNonPlatformTeamInPlatform, got warnings: %v", result.Warnings)
	}
}

func TestValidate_PlatformTeamInPlatform_NoWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Add a platform-type team
	platTeam := mustNewTeam(t, "pt-1", "platform-team", valueobject.Platform)
	platTeam.AddOwns(mustNewRelationship(t, "Payment Processing"))
	if err := m.AddTeam(platTeam); err != nil {
		t.Fatalf("AddTeam: %v", err)
	}

	plat, _ := entity.NewPlatform("plat-2", "ProperPlatform", "")
	plat.AddTeam("platform-team")
	if err := m.AddPlatform(plat); err != nil {
		t.Fatalf("AddPlatform: %v", err)
	}

	result := engine.Validate(m)
	for _, w := range result.Warnings {
		if w.Code == service.WarnNonPlatformTeamInPlatform && w.Entity == "ProperPlatform" {
			t.Errorf("did not expect WarnNonPlatformTeamInPlatform for platform-type team")
		}
	}
}

// ── Test 14: WarnSelfDependency ────────────────────────────────────────────────

func TestValidate_ServiceSelfDependency_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Service that depends on itself
	selfSvc := mustNewService(t, "svc-self", "self-svc", "payments-team")
	selfSvc.AddDependsOn(mustNewRelationship(t, "self-svc"))
	// Wire it up to a capability so we don't get WarnOrphanService
	cap := m.Capabilities["Payment Processing"]
	cap.AddRealizedBy(mustNewRelationship(t, "self-svc"))
	if err := m.AddService(selfSvc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnSelfDependency) {
		t.Errorf("expected WarnSelfDependency for service with self-dep, got warnings: %v", result.Warnings)
	}
	if hasWarning(result, service.WarnCircularDep) {
		t.Errorf("service self-dep should NOT produce WarnCircularDep, got warnings: %v", result.Warnings)
	}
}

func TestValidate_CapabilitySelfDependency_Warning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Capability that depends on itself
	selfCap := mustNewCapability(t, "cap-self", "Self Cap", "")
	selfCap.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	selfCap.AddDependsOn(mustNewRelationship(t, "Self Cap"))
	if err := m.AddCapability(selfCap); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnSelfDependency) {
		t.Errorf("expected WarnSelfDependency for capability with self-dep, got warnings: %v", result.Warnings)
	}
	if hasWarning(result, service.WarnCircularDep) {
		t.Errorf("capability self-dep should NOT produce WarnCircularDep, got warnings: %v", result.Warnings)
	}
}

func TestValidate_ServiceLegitDep_NoSelfDependencyWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Service with a legitimate dep on another service — no self-dependency warning
	otherSvc := mustNewService(t, "svc-other", "other-svc", "payments-team")
	cap := m.Capabilities["Payment Processing"]
	cap.AddRealizedBy(mustNewRelationship(t, "other-svc"))
	if err := m.AddService(otherSvc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	// payment-service depends on other-svc (not itself)
	svc := m.Services["payment-service"]
	svc.AddDependsOn(mustNewRelationship(t, "other-svc"))

	result := engine.Validate(m)
	if hasWarning(result, service.WarnSelfDependency) {
		t.Errorf("did not expect WarnSelfDependency for legitimate cross-service dep, got warnings: %v", result.Warnings)
	}
}

func TestValidate_CapabilityRealCycle_CircularDepNotSelfDep(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	// Two-node cycle: Cap C → Cap D → Cap C (genuine circular dep, not self-dep)
	capC := mustNewCapability(t, "cap-c", "Cap C", "")
	capD := mustNewCapability(t, "cap-d", "Cap D", "")
	capC.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	capD.AddRealizedBy(mustNewRelationship(t, "payment-service"))
	capC.AddDependsOn(mustNewRelationship(t, "Cap D"))
	capD.AddDependsOn(mustNewRelationship(t, "Cap C"))

	if err := m.AddCapability(capC); err != nil {
		t.Fatalf("AddCapability capC: %v", err)
	}
	if err := m.AddCapability(capD); err != nil {
		t.Fatalf("AddCapability capD: %v", err)
	}

	result := engine.Validate(m)
	if !hasWarning(result, service.WarnCircularDep) {
		t.Errorf("expected WarnCircularDep for 2-node cycle, got warnings: %v", result.Warnings)
	}
	if hasWarning(result, service.WarnSelfDependency) {
		t.Errorf("2-node cycle should NOT produce WarnSelfDependency, got warnings: %v", result.Warnings)
	}
}

func TestValidate_NoSelfDeps_NoSelfDependencyWarning(t *testing.T) {
	engine := service.NewValidationEngine()
	m := buildValidModel(t)

	result := engine.Validate(m)
	if hasWarning(result, service.WarnSelfDependency) {
		t.Errorf("did not expect WarnSelfDependency for model with no self-deps, got warnings: %v", result.Warnings)
	}
}

// ── Test 15: ValidationResult methods ─────────────────────────────────────────

func TestValidationResult_IsValid(t *testing.T) {
	r := service.ValidationResult{}
	if !r.IsValid() {
		t.Error("empty result should be valid")
	}
	r.Errors = append(r.Errors, service.ValidationError{Code: service.ErrNeedNoCapability, Entity: "x"})
	if r.IsValid() {
		t.Error("result with errors should not be valid")
	}
}

func TestValidationResult_HasWarnings(t *testing.T) {
	r := service.ValidationResult{}
	if r.HasWarnings() {
		t.Error("empty result should have no warnings")
	}
	r.Warnings = append(r.Warnings, service.ValidationWarning{Code: service.WarnFragmentation, Entity: "x"})
	if !r.HasWarnings() {
		t.Error("result with warnings should report HasWarnings true")
	}
}
