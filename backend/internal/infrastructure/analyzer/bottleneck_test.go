package analyzer

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
	"github.com/uber/unm-platform/internal/infrastructure/parser"
)

// mustAddService is a helper that creates and adds a service to the model, fataling on error.
func mustAddService(t *testing.T, m *entity.UNMModel, id, name, owner string) *entity.Service {
	t.Helper()
	svc, err := entity.NewService(id, name, "", owner)
	if err != nil {
		t.Fatalf("NewService %q: %v", name, err)
	}
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService %q: %v", name, err)
	}
	return svc
}

// mustDepID is a helper that creates a valueobject.EntityID, fataling on error.
func mustDepID(t *testing.T, id string) valueobject.EntityID {
	t.Helper()
	eid, err := valueobject.NewEntityID(id)
	if err != nil {
		t.Fatalf("NewEntityID %q: %v", id, err)
	}
	return eid
}

// findBottleneck returns the ServiceBottleneck for the named service, or nil if not found.
func findBottleneck(report BottleneckReport, name string) *ServiceBottleneck {
	for i := range report.ServiceBottlenecks {
		if report.ServiceBottlenecks[i].Service.Name == name {
			return &report.ServiceBottlenecks[i]
		}
	}
	return nil
}

// TestBottleneckAnalyzer_EmptyModel verifies an empty model returns an empty report.
func TestBottleneckAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	if len(report.ServiceBottlenecks) != 0 {
		t.Errorf("expected 0 bottlenecks for empty model, got %d", len(report.ServiceBottlenecks))
	}
}

// TestBottleneckAnalyzer_SingleServiceNoDeps verifies a lone service has fan-in=0 and fan-out=0.
func TestBottleneckAnalyzer_SingleServiceNoDeps(t *testing.T) {
	m := entity.NewUNMModel("single", "")
	mustAddService(t, m, "svc-alone", "svc-alone", "team-a")

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	if len(report.ServiceBottlenecks) != 1 {
		t.Fatalf("expected 1 bottleneck entry, got %d", len(report.ServiceBottlenecks))
	}
	b := report.ServiceBottlenecks[0]
	if b.FanIn != 0 {
		t.Errorf("expected FanIn=0, got %d", b.FanIn)
	}
	if b.FanOut != 0 {
		t.Errorf("expected FanOut=0, got %d", b.FanOut)
	}
	if b.IsWarning {
		t.Errorf("expected IsWarning=false, got true")
	}
	if b.IsCritical {
		t.Errorf("expected IsCritical=false, got true")
	}
}

// TestBottleneckAnalyzer_FanInThree verifies a service depended on by 3 others has fan-in=3,
// no warning, no critical.
func TestBottleneckAnalyzer_FanInThree(t *testing.T) {
	m := entity.NewUNMModel("fanin3", "")

	hub := mustAddService(t, m, "hub", "hub", "team-a")
	_ = hub

	for i, name := range []string{"caller-1", "caller-2", "caller-3"} {
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub"), "", ""))
		_ = i
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	b := findBottleneck(report, "hub")
	if b == nil {
		t.Fatal("hub not found in report")
	}
	if b.FanIn != 3 {
		t.Errorf("expected FanIn=3, got %d", b.FanIn)
	}
	if b.IsWarning {
		t.Errorf("expected IsWarning=false for fan-in=3, got true")
	}
	if b.IsCritical {
		t.Errorf("expected IsCritical=false for fan-in=3, got true")
	}
}

// TestBottleneckAnalyzer_FanInSix verifies fan-in=6 triggers IsWarning=true, IsCritical=false.
func TestBottleneckAnalyzer_FanInSix(t *testing.T) {
	m := entity.NewUNMModel("fanin6", "")
	mustAddService(t, m, "hub", "hub", "team-a")

	for i := 0; i < 6; i++ {
		name := "caller"
		switch i {
		case 0:
			name = "caller-a"
		case 1:
			name = "caller-b"
		case 2:
			name = "caller-c"
		case 3:
			name = "caller-d"
		case 4:
			name = "caller-e"
		case 5:
			name = "caller-f"
		}
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub"), "", ""))
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	b := findBottleneck(report, "hub")
	if b == nil {
		t.Fatal("hub not found in report")
	}
	if b.FanIn != 6 {
		t.Errorf("expected FanIn=6, got %d", b.FanIn)
	}
	if !b.IsWarning {
		t.Errorf("expected IsWarning=true for fan-in=6, got false")
	}
	if b.IsCritical {
		t.Errorf("expected IsCritical=false for fan-in=6, got true")
	}
}

// TestBottleneckAnalyzer_FanInEleven verifies fan-in=11 triggers IsCritical=true, IsWarning=false.
func TestBottleneckAnalyzer_FanInEleven(t *testing.T) {
	m := entity.NewUNMModel("fanin11", "")
	mustAddService(t, m, "hub", "hub", "team-a")

	callerNames := []string{
		"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10", "c11",
	}
	for _, name := range callerNames {
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub"), "", ""))
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	b := findBottleneck(report, "hub")
	if b == nil {
		t.Fatal("hub not found in report")
	}
	if b.FanIn != 11 {
		t.Errorf("expected FanIn=11, got %d", b.FanIn)
	}
	if b.IsWarning {
		t.Errorf("expected IsWarning=false for fan-in=11 (critical), got true")
	}
	if !b.IsCritical {
		t.Errorf("expected IsCritical=true for fan-in=11, got false")
	}
}

// TestBottleneckAnalyzer_SelfLoopNotCounted verifies that a service depending on itself
// does not count toward its own fan-in or fan-out.
func TestBottleneckAnalyzer_SelfLoopNotCounted(t *testing.T) {
	m := entity.NewUNMModel("self-loop", "")
	svc := mustAddService(t, m, "svc-self", "svc-self", "team-a")
	svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "svc-self"), "", ""))

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	b := findBottleneck(report, "svc-self")
	if b == nil {
		t.Fatal("svc-self not found in report")
	}
	if b.FanIn != 0 {
		t.Errorf("self-loop: expected FanIn=0, got %d", b.FanIn)
	}
	if b.FanOut != 0 {
		t.Errorf("self-loop: expected FanOut=0, got %d", b.FanOut)
	}
}

// TestBottleneckAnalyzer_Ranking verifies the report is sorted by fan-in descending,
// then by name ascending for ties.
func TestBottleneckAnalyzer_Ranking(t *testing.T) {
	m := entity.NewUNMModel("ranking", "")

	// hub-b: fan-in = 2
	mustAddService(t, m, "hub-b", "hub-b", "team-a")
	// hub-a: fan-in = 2 (tie with hub-b — should come first alphabetically)
	mustAddService(t, m, "hub-a", "hub-a", "team-a")
	// hub-c: fan-in = 3 (should rank first)
	mustAddService(t, m, "hub-c", "hub-c", "team-a")
	// hub-d: fan-in = 0
	mustAddService(t, m, "hub-d", "hub-d", "team-a")

	// 2 callers depend on hub-a
	for _, name := range []string{"caller-x", "caller-y"} {
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub-a"), "", ""))
	}

	// 2 callers depend on hub-b
	for _, name := range []string{"caller-p", "caller-q"} {
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub-b"), "", ""))
	}

	// 3 callers depend on hub-c
	for _, name := range []string{"caller-1", "caller-2", "caller-3"} {
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub-c"), "", ""))
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	// Find hub entries (there will be caller entries too — we check only the hubs' relative order).
	bottlenecks := report.ServiceBottlenecks

	// hub-c (fan-in=3) should appear before hub-a and hub-b (fan-in=2).
	// hub-a should appear before hub-b (alphabetical tie-break).
	// hub-d (fan-in=0) should appear after hub-a, hub-b.

	indexOf := func(name string) int {
		for i, b := range bottlenecks {
			if b.Service.Name == name {
				return i
			}
		}
		return -1
	}

	idxC := indexOf("hub-c")
	idxA := indexOf("hub-a")
	idxB := indexOf("hub-b")
	idxD := indexOf("hub-d")

	if idxC < 0 || idxA < 0 || idxB < 0 || idxD < 0 {
		t.Fatalf("missing hub entries: c=%d a=%d b=%d d=%d", idxC, idxA, idxB, idxD)
	}

	if idxC >= idxA {
		t.Errorf("hub-c (fan-in=3) should rank before hub-a (fan-in=2): positions c=%d a=%d", idxC, idxA)
	}
	if idxA >= idxB {
		t.Errorf("hub-a should rank before hub-b (alphabetical tie-break): positions a=%d b=%d", idxA, idxB)
	}
	if idxB >= idxD {
		t.Errorf("hub-b (fan-in=2) should rank before hub-d (fan-in=0): positions b=%d d=%d", idxB, idxD)
	}
}

// TestBottleneckAnalyzer_FanOut verifies fan-out equals the number of non-self dependencies.
func TestBottleneckAnalyzer_FanOut(t *testing.T) {
	m := entity.NewUNMModel("fanout", "")

	caller := mustAddService(t, m, "caller", "caller", "team-a")
	for _, name := range []string{"dep-1", "dep-2", "dep-3", "dep-4"} {
		mustAddService(t, m, name, name, "team-b")
		caller.AddDependsOn(entity.NewRelationship(mustDepID(t, name), "", ""))
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	b := findBottleneck(report, "caller")
	if b == nil {
		t.Fatal("caller not found in report")
	}
	if b.FanOut != 4 {
		t.Errorf("expected FanOut=4, got %d", b.FanOut)
	}
}

// testDir returns the directory of the current test file.
func testDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

// TestBottleneckAnalyzer_ExternalDepFanIn verifies external dependency bottleneck detection
// using the INCA model. Cadence is used by 12 services and must appear as IsCritical=true.
func TestBottleneckAnalyzer_ExternalDepFanIn(t *testing.T) {
	incaPath := filepath.Join(testDir(), "..", "..", "..", "..", "examples", "inca.unm.yaml")
	m, err := parser.ParseFile(incaPath)
	if err != nil {
		t.Fatalf("failed to parse INCA model: %v", err)
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	// Find cadence in ExternalDependencyBottlenecks
	var found *ExternalDepBottleneck
	for i := range report.ExternalDependencyBottlenecks {
		if report.ExternalDependencyBottlenecks[i].Name == "cadence" {
			found = &report.ExternalDependencyBottlenecks[i]
			break
		}
	}
	if found == nil {
		t.Fatal("cadence not found in ExternalDependencyBottlenecks")
	}
	if found.ServiceCount != 12 {
		t.Errorf("expected cadence ServiceCount=12, got %d", found.ServiceCount)
	}
	if !found.IsCritical {
		t.Errorf("expected cadence IsCritical=true (12 services >= threshold 5), got false")
	}
	if found.IsWarning {
		t.Errorf("expected cadence IsWarning=false (critical takes priority), got true")
	}
	if len(found.Services) != 12 {
		t.Errorf("expected 12 service names, got %d", len(found.Services))
	}
}

// TestBottleneckAnalyzer_ExternalDepWarning verifies that a dep used by 3-4 services is IsWarning.
func TestBottleneckAnalyzer_ExternalDepWarning(t *testing.T) {
	m := entity.NewUNMModel("test-ext-dep-warning", "")

	dep, err := entity.NewExternalDependency("dep-1", "my-dep", "A shared external dep")
	if err != nil {
		t.Fatalf("NewExternalDependency: %v", err)
	}
	for _, svcName := range []string{"svc-a", "svc-b", "svc-c"} {
		dep.AddUsedBy(svcName, "")
		svc, err := entity.NewService(svcName, svcName, "", "team-a")
		if err != nil {
			t.Fatalf("NewService: %v", err)
		}
		if err := m.AddService(svc); err != nil {
			t.Fatalf("AddService: %v", err)
		}
	}
	if err := m.AddExternalDependency(dep); err != nil {
		t.Fatalf("AddExternalDependency: %v", err)
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	var found *ExternalDepBottleneck
	for i := range report.ExternalDependencyBottlenecks {
		if report.ExternalDependencyBottlenecks[i].Name == "my-dep" {
			found = &report.ExternalDependencyBottlenecks[i]
			break
		}
	}
	if found == nil {
		t.Fatal("my-dep not found in ExternalDependencyBottlenecks")
	}
	if found.ServiceCount != 3 {
		t.Errorf("expected ServiceCount=3, got %d", found.ServiceCount)
	}
	if found.IsCritical {
		t.Errorf("expected IsCritical=false for 3 services, got true")
	}
	if !found.IsWarning {
		t.Errorf("expected IsWarning=true for 3 services, got false")
	}
}

// TestBottleneckAnalyzer_ExternalDepBelowThreshold verifies deps with fewer than 3 services are excluded.
func TestBottleneckAnalyzer_ExternalDepBelowThreshold(t *testing.T) {
	m := entity.NewUNMModel("test-ext-dep-below", "")

	dep, err := entity.NewExternalDependency("dep-1", "rare-dep", "Rarely used dep")
	if err != nil {
		t.Fatalf("NewExternalDependency: %v", err)
	}
	for _, svcName := range []string{"svc-a", "svc-b"} {
		dep.AddUsedBy(svcName, "")
		svc, err := entity.NewService(svcName, svcName, "", "team-a")
		if err != nil {
			t.Fatalf("NewService: %v", err)
		}
		if err := m.AddService(svc); err != nil {
			t.Fatalf("AddService: %v", err)
		}
	}
	if err := m.AddExternalDependency(dep); err != nil {
		t.Fatalf("AddExternalDependency: %v", err)
	}

	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)

	for _, b := range report.ExternalDependencyBottlenecks {
		if b.Name == "rare-dep" {
			t.Errorf("rare-dep (2 services) should not appear in ExternalDependencyBottlenecks")
		}
	}
}

// TestBottleneckAnalyzer_CustomThresholds verifies that non-default thresholds change behavior.
func TestBottleneckAnalyzer_CustomThresholds(t *testing.T) {
	m := entity.NewUNMModel("custom", "")
	mustAddService(t, m, "hub", "hub", "team-a")

	// 2 callers → fan-in = 2
	for _, name := range []string{"caller-1", "caller-2"} {
		svc := mustAddService(t, m, name, name, "team-b")
		svc.AddDependsOn(entity.NewRelationship(mustDepID(t, "hub"), "", ""))
	}

	// With default thresholds (warning=5, critical=10), fan-in=2 is not a warning.
	a := NewBottleneckAnalyzer(entity.DefaultConfig().Analysis.Bottleneck)
	report := a.Analyze(m)
	b := findBottleneck(report, "hub")
	if b.IsWarning {
		t.Error("with defaults, fan-in=2 should not be a warning")
	}

	// With custom thresholds (warning=1, critical=5), fan-in=2 IS a warning.
	customCfg := entity.BottleneckConfig{FanInWarning: 1, FanInCritical: 5}
	a2 := NewBottleneckAnalyzer(customCfg)
	report2 := a2.Analyze(m)
	b2 := findBottleneck(report2, "hub")
	if !b2.IsWarning {
		t.Error("with FanInWarning=1, fan-in=2 should be a warning")
	}
}
