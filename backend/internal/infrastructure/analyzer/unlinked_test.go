package analyzer

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// ── unlinked-specific helpers ─────────────────────────────────────────────────

func mustCapVis(t *testing.T, name, vis string) *entity.Capability {
	t.Helper()
	cap, err := entity.NewCapability(name, name, name)
	if err != nil {
		t.Fatalf("NewCapability %q: %v", name, err)
	}
	if vis != "" {
		if err := cap.SetVisibility(vis); err != nil {
			t.Fatalf("SetVisibility %q on %q: %v", vis, name, err)
		}
	}
	return cap
}

func addCapVis(t *testing.T, m *entity.UNMModel, cap *entity.Capability) {
	t.Helper()
	if err := m.AddCapability(cap); err != nil {
		t.Fatalf("AddCapability %q: %v", cap.Name, err)
	}
}

func mustNeedRef(t *testing.T, name, actor, capTarget string) *entity.Need {
	t.Helper()
	need, err := entity.NewNeed(name, name, actor, "outcome")
	if err != nil {
		t.Fatalf("NewNeed %q: %v", name, err)
	}
	tid, err := valueobject.NewEntityID(capTarget)
	if err != nil {
		t.Fatalf("NewEntityID %q: %v", capTarget, err)
	}
	need.AddSupportedBy(entity.NewRelationship(tid, "", valueobject.RelationshipRole("")))
	return need
}

func addNeedRef(t *testing.T, m *entity.UNMModel, need *entity.Need) {
	t.Helper()
	if err := m.AddNeed(need); err != nil {
		t.Fatalf("AddNeed %q: %v", need.Name, err)
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestUnlinkedCapabilityAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	a := NewUnlinkedCapabilityAnalyzer()
	r := a.Analyze(m)

	if r.TotalLeafCapabilityCount != 0 {
		t.Errorf("want 0 leaves, got %d", r.TotalLeafCapabilityCount)
	}
	if len(r.UnlinkedLeafCapabilities) != 0 {
		t.Errorf("want no unlinked, got %d", len(r.UnlinkedLeafCapabilities))
	}
	if r.LinkedPercentage != 0.0 {
		t.Errorf("want 0%% linked, got %v", r.LinkedPercentage)
	}
}

func TestUnlinkedCapabilityAnalyzer_AllLinked(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	addCapVis(t, m, mustCapVis(t, "cap-a", "domain"))

	need := mustNeedRef(t, "need-a", "actor-a", "cap-a")
	addNeedRef(t, m, need)

	a := NewUnlinkedCapabilityAnalyzer()
	r := a.Analyze(m)

	if r.TotalLeafCapabilityCount != 1 {
		t.Errorf("want 1 leaf, got %d", r.TotalLeafCapabilityCount)
	}
	if r.LinkedCount != 1 {
		t.Errorf("want 1 linked, got %d", r.LinkedCount)
	}
	if len(r.UnlinkedLeafCapabilities) != 0 {
		t.Errorf("want 0 unlinked, got %d", len(r.UnlinkedLeafCapabilities))
	}
	if r.LinkedPercentage != 100.0 {
		t.Errorf("want 100%% linked, got %v", r.LinkedPercentage)
	}
}

func TestUnlinkedCapabilityAnalyzer_InfrastructureExpected(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	addCapVis(t, m, mustCapVis(t, "blob-storage", "infrastructure"))

	a := NewUnlinkedCapabilityAnalyzer()
	r := a.Analyze(m)

	if r.TotalLeafCapabilityCount != 1 {
		t.Errorf("want 1 leaf, got %d", r.TotalLeafCapabilityCount)
	}
	if len(r.UnlinkedLeafCapabilities) != 1 {
		t.Fatalf("want 1 unlinked, got %d", len(r.UnlinkedLeafCapabilities))
	}
	uc := r.UnlinkedLeafCapabilities[0]
	if !uc.IsExpected {
		t.Errorf("infrastructure unlinked cap should be expected")
	}
	if uc.Visibility != "infrastructure" {
		t.Errorf("want visibility=infrastructure, got %q", uc.Visibility)
	}
	if r.ByVisibility["infrastructure"] != 1 {
		t.Errorf("want ByVisibility[infrastructure]=1, got %d", r.ByVisibility["infrastructure"])
	}
}

func TestUnlinkedCapabilityAnalyzer_DomainUnlinkedIsSuspicious(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	addCapVis(t, m, mustCapVis(t, "catalog-indexing", "domain"))

	a := NewUnlinkedCapabilityAnalyzer()
	r := a.Analyze(m)

	if len(r.UnlinkedLeafCapabilities) != 1 {
		t.Fatalf("want 1 unlinked, got %d", len(r.UnlinkedLeafCapabilities))
	}
	uc := r.UnlinkedLeafCapabilities[0]
	if uc.IsExpected {
		t.Errorf("domain unlinked cap should NOT be expected")
	}
	if uc.Visibility != "domain" {
		t.Errorf("want visibility=domain, got %q", uc.Visibility)
	}
}

func TestUnlinkedCapabilityAnalyzer_MixedVisibility(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// One linked domain cap
	addCapVis(t, m, mustCapVis(t, "dom-linked", "domain"))
	addNeedRef(t, m, mustNeedRef(t, "need-1", "actor", "dom-linked"))

	// One unlinked domain cap (suspicious)
	addCapVis(t, m, mustCapVis(t, "dom-unlinked", "domain"))

	// One unlinked foundational cap (suspicious)
	addCapVis(t, m, mustCapVis(t, "found-unlinked", "foundational"))

	// One unlinked infrastructure cap (expected)
	addCapVis(t, m, mustCapVis(t, "infra-unlinked", "infrastructure"))

	a := NewUnlinkedCapabilityAnalyzer()
	r := a.Analyze(m)

	if r.TotalLeafCapabilityCount != 4 {
		t.Errorf("want 4 total leaves, got %d", r.TotalLeafCapabilityCount)
	}
	if r.LinkedCount != 1 {
		t.Errorf("want 1 linked, got %d", r.LinkedCount)
	}
	if len(r.UnlinkedLeafCapabilities) != 3 {
		t.Errorf("want 3 unlinked, got %d", len(r.UnlinkedLeafCapabilities))
	}

	// Ordering: domain < foundational < infrastructure
	if r.UnlinkedLeafCapabilities[0].Visibility != "domain" {
		t.Errorf("first unlinked should be domain, got %q", r.UnlinkedLeafCapabilities[0].Visibility)
	}
	if r.UnlinkedLeafCapabilities[1].Visibility != "foundational" {
		t.Errorf("second unlinked should be foundational, got %q", r.UnlinkedLeafCapabilities[1].Visibility)
	}
	if r.UnlinkedLeafCapabilities[2].Visibility != "infrastructure" {
		t.Errorf("third unlinked should be infrastructure, got %q", r.UnlinkedLeafCapabilities[2].Visibility)
	}

	// IsExpected flags
	for _, uc := range r.UnlinkedLeafCapabilities {
		if uc.Visibility == "infrastructure" && !uc.IsExpected {
			t.Errorf("infrastructure should be IsExpected=true")
		}
		if uc.Visibility != "infrastructure" && uc.IsExpected {
			t.Errorf("%q should not be IsExpected", uc.Visibility)
		}
	}

	// LinkedPercentage: 1/4 = 25
	if r.LinkedPercentage != 25.0 {
		t.Errorf("want 25%% linked, got %v", r.LinkedPercentage)
	}

	if r.ByVisibility["domain"] != 1 {
		t.Errorf("want domain=1, got %d", r.ByVisibility["domain"])
	}
	if r.ByVisibility["foundational"] != 1 {
		t.Errorf("want foundational=1, got %d", r.ByVisibility["foundational"])
	}
	if r.ByVisibility["infrastructure"] != 1 {
		t.Errorf("want infrastructure=1, got %d", r.ByVisibility["infrastructure"])
	}
}

func TestUnlinkedCapabilityAnalyzer_ParentCapsNotCounted(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// Parent cap with a child: AddCapability registers both via recursive walk
	parent := mustCapVis(t, "parent", "domain")
	child := mustCapVis(t, "child", "domain")
	parent.AddChild(child)
	addCapVis(t, m, parent) // also registers child transitively

	// child is a leaf and unlinked → should appear; parent is not a leaf
	a := NewUnlinkedCapabilityAnalyzer()
	r := a.Analyze(m)

	if r.TotalLeafCapabilityCount != 1 {
		t.Errorf("want 1 leaf (child only), got %d", r.TotalLeafCapabilityCount)
	}
	if len(r.UnlinkedLeafCapabilities) != 1 {
		t.Fatalf("want 1 unlinked, got %d", len(r.UnlinkedLeafCapabilities))
	}
	if r.UnlinkedLeafCapabilities[0].Capability.Name != "child" {
		t.Errorf("expected child to be unlinked, got %q", r.UnlinkedLeafCapabilities[0].Capability.Name)
	}
}
