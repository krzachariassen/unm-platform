package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// buildGapModel creates a model with deliberate gaps:
//   - need "Unmapped Need" (Merchant) with NO SupportedBy → unmapped need
//   - cap "Unrealized Cap" (leaf, no service realizing it, no children) → unrealized capability
//   - cap "Parent Cap" with 1 child that has a service realizing it → parent NOT flagged
//   - svc "Unowned Svc" with OwnerTeamName="" → unowned service
//   - cap "Unneeded Cap" not referenced by any need's SupportedBy → unneeded capability
//   - need "Mapped Need" (Merchant) with SupportedBy pointing to "Realized Cap" → mapped need (negative case)
//   - cap "Realized Cap" (leaf) with service realizing it → realized capability (negative case)
func buildGapModel(t *testing.T) *entity.UNMModel {
	t.Helper()

	m := entity.NewUNMModel("gap-system", "")

	// Actor
	actor, err := entity.NewActor("merchant", "Merchant", "")
	if err != nil {
		t.Fatalf("NewActor: %v", err)
	}
	if err := m.AddActor(&actor); err != nil {
		t.Fatalf("AddActor: %v", err)
	}

	// Unmapped need — no SupportedBy.
	unmappedNeed, err := entity.NewNeed("unmapped-need", "Unmapped Need", "Merchant", "")
	if err != nil {
		t.Fatalf("NewNeed unmapped: %v", err)
	}
	if err := m.AddNeed(unmappedNeed); err != nil {
		t.Fatalf("AddNeed unmapped: %v", err)
	}

	// Mapped need — has SupportedBy pointing to "Realized Cap".
	mappedNeed, err := entity.NewNeed("mapped-need", "Mapped Need", "Merchant", "")
	if err != nil {
		t.Fatalf("NewNeed mapped: %v", err)
	}
	realizedCapID, err := valueobject.NewEntityID("Realized Cap")
	if err != nil {
		t.Fatalf("NewEntityID Realized Cap: %v", err)
	}
	mappedNeed.AddSupportedBy(entity.NewRelationship(realizedCapID, "", ""))
	if err := m.AddNeed(mappedNeed); err != nil {
		t.Fatalf("AddNeed mapped: %v", err)
	}

	// Unrealized Cap — leaf, no service realizes it.
	unrealizedCap, err := entity.NewCapability("unrealized-cap", "Unrealized Cap", "")
	if err != nil {
		t.Fatalf("NewCapability unrealized: %v", err)
	}
	if err := m.AddCapability(unrealizedCap); err != nil {
		t.Fatalf("AddCapability unrealized: %v", err)
	}

	// Realized Cap — leaf, with a service realizing it.
	realizedCap, err := entity.NewCapability("realized-cap", "Realized Cap", "")
	if err != nil {
		t.Fatalf("NewCapability realized: %v", err)
	}
	if err := m.AddCapability(realizedCap); err != nil {
		t.Fatalf("AddCapability realized: %v", err)
	}

	// Parent Cap with one child that has a service realizing it.
	childCap, err := entity.NewCapability("child-cap", "Child Cap", "")
	if err != nil {
		t.Fatalf("NewCapability child: %v", err)
	}

	parentCap, err := entity.NewCapability("parent-cap", "Parent Cap", "")
	if err != nil {
		t.Fatalf("NewCapability parent: %v", err)
	}
	parentCap.AddChild(childCap)
	if err := m.AddCapability(parentCap); err != nil {
		t.Fatalf("AddCapability parent: %v", err)
	}

	// Unneeded Cap — not referenced by any need.
	unneededCap, err := entity.NewCapability("unneeded-cap", "Unneeded Cap", "")
	if err != nil {
		t.Fatalf("NewCapability unneeded: %v", err)
	}
	if err := m.AddCapability(unneededCap); err != nil {
		t.Fatalf("AddCapability unneeded: %v", err)
	}

	// Owned service — realizes Realized Cap, Child Cap, and Unneeded Cap.
	ownedSvc, err := entity.NewService("owned-svc", "Owned Svc", "", "Team Alpha")
	if err != nil {
		t.Fatalf("NewService owned: %v", err)
	}
	ownedSvc.AddRealizes(entity.NewRelationship(realizedCapID, "", ""))
	childCapID, _ := valueobject.NewEntityID("Child Cap")
	ownedSvc.AddRealizes(entity.NewRelationship(childCapID, "", ""))
	unneededCapID, _ := valueobject.NewEntityID("Unneeded Cap")
	ownedSvc.AddRealizes(entity.NewRelationship(unneededCapID, "", ""))
	if err := m.AddService(ownedSvc); err != nil {
		t.Fatalf("AddService owned: %v", err)
	}

	// Unowned service — OwnerTeamName is empty.
	unownedSvc, err := entity.NewService("unowned-svc", "Unowned Svc", "", "")
	if err != nil {
		t.Fatalf("NewService unowned: %v", err)
	}
	if err := m.AddService(unownedSvc); err != nil {
		t.Fatalf("AddService unowned: %v", err)
	}

	return m
}

func TestGapAnalyzer_UnmappedNeeds(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	if len(report.UnmappedNeeds) != 1 {
		t.Fatalf("expected 1 unmapped need, got %d", len(report.UnmappedNeeds))
	}
	if report.UnmappedNeeds[0].Name != "Unmapped Need" {
		t.Errorf("expected unmapped need to be 'Unmapped Need', got %q", report.UnmappedNeeds[0].Name)
	}
}

func TestGapAnalyzer_MappedNeedNotReported(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	for _, n := range report.UnmappedNeeds {
		if n.Name == "Mapped Need" {
			t.Errorf("'Mapped Need' should not be in UnmappedNeeds")
		}
	}
}

func TestGapAnalyzer_UnrealizedCapabilities(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	if len(report.UnrealizedCapabilities) != 1 {
		t.Fatalf("expected 1 unrealized capability, got %d", len(report.UnrealizedCapabilities))
	}
	if report.UnrealizedCapabilities[0].Name != "Unrealized Cap" {
		t.Errorf("expected 'Unrealized Cap', got %q", report.UnrealizedCapabilities[0].Name)
	}
}

func TestGapAnalyzer_ParentCapNotFlaggedUnrealized(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	for _, c := range report.UnrealizedCapabilities {
		if c.Name == "Parent Cap" {
			t.Errorf("'Parent Cap' (non-leaf) should not be in UnrealizedCapabilities")
		}
	}
}

func TestGapAnalyzer_RealizedCapNotFlagged(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	for _, c := range report.UnrealizedCapabilities {
		if c.Name == "Realized Cap" {
			t.Errorf("'Realized Cap' should not be in UnrealizedCapabilities")
		}
	}
}

func TestGapAnalyzer_UnownedServices(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	if len(report.UnownedServices) != 1 {
		t.Fatalf("expected 1 unowned service, got %d", len(report.UnownedServices))
	}
	if report.UnownedServices[0].Name != "Unowned Svc" {
		t.Errorf("expected 'Unowned Svc', got %q", report.UnownedServices[0].Name)
	}
}

func TestGapAnalyzer_OwnedServiceNotFlagged(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	for _, s := range report.UnownedServices {
		if s.Name == "Owned Svc" {
			t.Errorf("'Owned Svc' should not be in UnownedServices")
		}
	}
}

func TestGapAnalyzer_UnneededCapabilities(t *testing.T) {
	m := buildGapModel(t)
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	// Unneeded Cap and Unrealized Cap are not referenced by any need.
	// Realized Cap IS referenced by Mapped Need.
	// Parent Cap and Child Cap are not referenced by any need either.
	unneededNames := make(map[string]bool)
	for _, c := range report.UnneededCapabilities {
		unneededNames[c.Name] = true
	}

	if !unneededNames["Unneeded Cap"] {
		t.Errorf("'Unneeded Cap' should be in UnneededCapabilities")
	}
	if unneededNames["Realized Cap"] {
		t.Errorf("'Realized Cap' is referenced by Mapped Need, should NOT be in UnneededCapabilities")
	}
}

func TestGapAnalyzer_ParentOfNeededChildNotFlaggedUnneeded(t *testing.T) {
	// Parent capability whose child leaf is referenced by a need should NOT be flagged as unneeded.
	m := entity.NewUNMModel("hierarchy-test", "")

	childLeaf, _ := entity.NewCapability("child-leaf", "Child Leaf", "")
	parentGroup, _ := entity.NewCapability("parent-group", "Parent Group", "")
	parentGroup.AddChild(childLeaf)
	_ = m.AddCapability(parentGroup)

	childLeafID, _ := valueobject.NewEntityID("Child Leaf")
	need, _ := entity.NewNeed("need-1", "Need 1", "Actor", "")
	need.AddSupportedBy(entity.NewRelationship(childLeafID, "", ""))
	_ = m.AddNeed(need)

	a := NewGapAnalyzer()
	report := a.Analyze(m)

	for _, c := range report.UnneededCapabilities {
		if c.Name == "Parent Group" {
			t.Error("'Parent Group' should not be unneeded — its child 'Child Leaf' is referenced by a need")
		}
	}
}

func TestGapAnalyzer_GrandparentOfNeededLeafNotFlaggedUnneeded(t *testing.T) {
	// Three-level hierarchy: grandparent → parent → leaf (referenced). Grandparent must not be flagged.
	m := entity.NewUNMModel("deep-hierarchy", "")

	leaf, _ := entity.NewCapability("leaf", "Leaf", "")
	parent, _ := entity.NewCapability("parent", "Parent", "")
	grandparent, _ := entity.NewCapability("grandparent", "Grandparent", "")
	parent.AddChild(leaf)
	grandparent.AddChild(parent)
	_ = m.AddCapability(grandparent)

	leafID, _ := valueobject.NewEntityID("Leaf")
	need, _ := entity.NewNeed("need-1", "Need 1", "Actor", "")
	need.AddSupportedBy(entity.NewRelationship(leafID, "", ""))
	_ = m.AddNeed(need)

	a := NewGapAnalyzer()
	report := a.Analyze(m)

	flagged := make(map[string]bool)
	for _, c := range report.UnneededCapabilities {
		flagged[c.Name] = true
	}
	if flagged["Grandparent"] {
		t.Error("'Grandparent' should not be unneeded — its descendant 'Leaf' is referenced by a need")
	}
	if flagged["Parent"] {
		t.Error("'Parent' should not be unneeded — its child 'Leaf' is referenced by a need")
	}
	if flagged["Leaf"] {
		t.Error("'Leaf' should not be unneeded — it is directly referenced by a need")
	}
}

func TestGapAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewGapAnalyzer()
	report := a.Analyze(m)

	if len(report.UnmappedNeeds) != 0 {
		t.Errorf("expected 0 unmapped needs, got %d", len(report.UnmappedNeeds))
	}
	if len(report.UnrealizedCapabilities) != 0 {
		t.Errorf("expected 0 unrealized capabilities, got %d", len(report.UnrealizedCapabilities))
	}
	if len(report.UnownedServices) != 0 {
		t.Errorf("expected 0 unowned services, got %d", len(report.UnownedServices))
	}
	if len(report.UnneededCapabilities) != 0 {
		t.Errorf("expected 0 unneeded capabilities, got %d", len(report.UnneededCapabilities))
	}
	if len(report.OrphanServices) != 0 {
		t.Errorf("expected 0 orphan services, got %d", len(report.OrphanServices))
	}
}

