package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// mustServiceID is a helper to create a valueobject.EntityID from a string, fataling on error.
func mustServiceID(t *testing.T, id string) valueobject.EntityID {
	t.Helper()
	eid, err := valueobject.NewEntityID(id)
	if err != nil {
		t.Fatalf("NewEntityID %q: %v", id, err)
	}
	return eid
}

// mustCapabilityID is a helper to create a valueobject.EntityID for a capability, fataling on error.
func mustCapabilityID(t *testing.T, id string) valueobject.EntityID {
	t.Helper()
	eid, err := valueobject.NewEntityID(id)
	if err != nil {
		t.Fatalf("NewEntityID %q: %v", id, err)
	}
	return eid
}

// mustNewService creates a Service, fataling on error.
func mustNewService(t *testing.T, id, name, owner string) *entity.Service {
	t.Helper()
	svc, err := entity.NewService(id, name, "", owner)
	if err != nil {
		t.Fatalf("NewService %q: %v", name, err)
	}
	return svc
}

// mustNewCapability creates a Capability, fataling on error.
func mustNewCapability(t *testing.T, id, name string) *entity.Capability {
	t.Helper()
	cap, err := entity.NewCapability(id, name, "")
	if err != nil {
		t.Fatalf("NewCapability %q: %v", name, err)
	}
	return cap
}

// TestDependencyAnalyzer_ServiceCycleDetected tests that a 3-node cycle is detected.
// svc-a → svc-b → svc-c → svc-a
func TestDependencyAnalyzer_ServiceCycleDetected(t *testing.T) {
	m := entity.NewUNMModel("cycle-test", "")

	svcA := mustNewService(t, "svc-a", "svc-a", "team-x")
	svcB := mustNewService(t, "svc-b", "svc-b", "team-x")
	svcC := mustNewService(t, "svc-c", "svc-c", "team-x")

	svcA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-b"), "", ""))
	svcB.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-c"), "", ""))
	svcC.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-a"), "", ""))

	if err := m.AddService(svcA); err != nil {
		t.Fatalf("AddService svc-a: %v", err)
	}
	if err := m.AddService(svcB); err != nil {
		t.Fatalf("AddService svc-b: %v", err)
	}
	if err := m.AddService(svcC); err != nil {
		t.Fatalf("AddService svc-c: %v", err)
	}

	a := NewDependencyAnalyzer()
	report := a.Analyze(m)

	if len(report.ServiceCycles) != 1 {
		t.Errorf("expected 1 service cycle, got %d", len(report.ServiceCycles))
	}
}

// TestDependencyAnalyzer_MaxServiceDepth tests that max depth is computed correctly.
// svc-x → svc-y → svc-z (depth 3)
func TestDependencyAnalyzer_MaxServiceDepth(t *testing.T) {
	m := entity.NewUNMModel("depth-test", "")

	svcX := mustNewService(t, "svc-x", "svc-x", "team-y")
	svcY := mustNewService(t, "svc-y", "svc-y", "team-y")
	svcZ := mustNewService(t, "svc-z", "svc-z", "team-y")

	svcX.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-y"), "", ""))
	svcY.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-z"), "", ""))

	if err := m.AddService(svcX); err != nil {
		t.Fatalf("AddService svc-x: %v", err)
	}
	if err := m.AddService(svcY); err != nil {
		t.Fatalf("AddService svc-y: %v", err)
	}
	if err := m.AddService(svcZ); err != nil {
		t.Fatalf("AddService svc-z: %v", err)
	}

	a := NewDependencyAnalyzer()
	report := a.Analyze(m)

	if report.MaxServiceDepth != 3 {
		t.Errorf("expected MaxServiceDepth=3, got %d", report.MaxServiceDepth)
	}
}

// TestDependencyAnalyzer_CriticalServicePath tests that the longest path is returned.
// svc-x → svc-y → svc-z
func TestDependencyAnalyzer_CriticalServicePath(t *testing.T) {
	m := entity.NewUNMModel("path-test", "")

	svcX := mustNewService(t, "svc-x", "svc-x", "team-y")
	svcY := mustNewService(t, "svc-y", "svc-y", "team-y")
	svcZ := mustNewService(t, "svc-z", "svc-z", "team-y")

	svcX.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-y"), "", ""))
	svcY.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-z"), "", ""))

	if err := m.AddService(svcX); err != nil {
		t.Fatalf("AddService svc-x: %v", err)
	}
	if err := m.AddService(svcY); err != nil {
		t.Fatalf("AddService svc-y: %v", err)
	}
	if err := m.AddService(svcZ); err != nil {
		t.Fatalf("AddService svc-z: %v", err)
	}

	a := NewDependencyAnalyzer()
	report := a.Analyze(m)

	want := []string{"svc-x", "svc-y", "svc-z"}
	if len(report.CriticalServicePath) != len(want) {
		t.Fatalf("CriticalServicePath: want %v, got %v", want, report.CriticalServicePath)
	}
	for i, name := range want {
		if report.CriticalServicePath[i] != name {
			t.Errorf("CriticalServicePath[%d]: want %q, got %q", i, name, report.CriticalServicePath[i])
		}
	}
}

// TestDependencyAnalyzer_NoCycles tests that a clean linear chain has no cycles.
func TestDependencyAnalyzer_NoCycles(t *testing.T) {
	m := entity.NewUNMModel("no-cycle-test", "")

	svcA := mustNewService(t, "svc-a", "svc-a", "team-x")
	svcB := mustNewService(t, "svc-b", "svc-b", "team-x")

	svcA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-b"), "", ""))

	if err := m.AddService(svcA); err != nil {
		t.Fatalf("AddService svc-a: %v", err)
	}
	if err := m.AddService(svcB); err != nil {
		t.Fatalf("AddService svc-b: %v", err)
	}

	a := NewDependencyAnalyzer()
	report := a.Analyze(m)

	if len(report.ServiceCycles) != 0 {
		t.Errorf("expected no service cycles, got %d", len(report.ServiceCycles))
	}
}

// TestDependencyAnalyzer_CapabilityCycleDetected tests that a 2-cap cycle is detected.
// cap-a → cap-b → cap-a
func TestDependencyAnalyzer_CapabilityCycleDetected(t *testing.T) {
	m := entity.NewUNMModel("cap-cycle-test", "")

	capA := mustNewCapability(t, "cap-a", "cap-a")
	capB := mustNewCapability(t, "cap-b", "cap-b")

	capA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "cap-b"), "", ""))
	capB.AddDependsOn(entity.NewRelationship(mustServiceID(t, "cap-a"), "", ""))

	if err := m.AddCapability(capA); err != nil {
		t.Fatalf("AddCapability cap-a: %v", err)
	}
	if err := m.AddCapability(capB); err != nil {
		t.Fatalf("AddCapability cap-b: %v", err)
	}

	a := NewDependencyAnalyzer()
	report := a.Analyze(m)

	if len(report.CapabilityCycles) != 1 {
		t.Errorf("expected 1 capability cycle, got %d", len(report.CapabilityCycles))
	}
}

func TestDependencyAnalyzer_SelfLoopNotReportedAsCycle(t *testing.T) {
	// A service that lists itself in DependsOn is a data artifact — not a real cycle.
	m := entity.NewUNMModel("self-loop", "")

	svc, err := entity.NewService("svc-self", "svc-self", "", "team-a")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	selfID := mustServiceID(t, "svc-self")
	svc.AddDependsOn(entity.NewRelationship(selfID, "", ""))
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	a := NewDependencyAnalyzer()
	report := a.Analyze(m)

	if len(report.ServiceCycles) != 0 {
		t.Errorf("self-loop should not be reported as a cycle, got %d cycles: %v",
			len(report.ServiceCycles), report.ServiceCycles)
	}
}
