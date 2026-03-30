package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// TestComplexityAnalyzer_EmptyModel verifies an empty model produces an empty report.
func TestComplexityAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	if len(report.Services) != 0 {
		t.Errorf("expected 0 services in report, got %d", len(report.Services))
	}
}

// TestComplexityAnalyzer_SingleServiceAllZero verifies a lone service with no relationships scores 0.
func TestComplexityAnalyzer_SingleServiceAllZero(t *testing.T) {
	m := entity.NewUNMModel("solo", "")
	svc := mustNewService(t, "svc-a", "svc-a", "team-x")
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	if len(report.Services) != 1 {
		t.Fatalf("expected 1 entry in report, got %d", len(report.Services))
	}
	sc := report.Services[0]
	if sc.DependencyScore != 0 {
		t.Errorf("DependencyScore: want 0, got %d", sc.DependencyScore)
	}
	if sc.CapabilityScore != 0 {
		t.Errorf("CapabilityScore: want 0, got %d", sc.CapabilityScore)
	}
	if sc.DataAssetScore != 0 {
		t.Errorf("DataAssetScore: want 0, got %d", sc.DataAssetScore)
	}
	if sc.TotalComplexity != 0 {
		t.Errorf("TotalComplexity: want 0, got %d", sc.TotalComplexity)
	}
}

// TestComplexityAnalyzer_CapabilityScore verifies that a service realized by 2 capabilities
// gets CapabilityScore=2.
func TestComplexityAnalyzer_CapabilityScore(t *testing.T) {
	m := entity.NewUNMModel("cap-score", "")

	svc := mustNewService(t, "svc-a", "svc-a", "team-x")
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	cap1 := mustNewCapability(t, "cap-1", "cap-1")
	cap1.AddRealizedBy(entity.NewRelationship(mustServiceID(t, "svc-a"), "", ""))
	if err := m.AddCapability(cap1); err != nil {
		t.Fatalf("AddCapability cap-1: %v", err)
	}

	cap2 := mustNewCapability(t, "cap-2", "cap-2")
	cap2.AddRealizedBy(entity.NewRelationship(mustServiceID(t, "svc-a"), "", ""))
	if err := m.AddCapability(cap2); err != nil {
		t.Fatalf("AddCapability cap-2: %v", err)
	}

	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	if len(report.Services) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Services))
	}
	sc := report.Services[0]
	if sc.CapabilityScore != 2 {
		t.Errorf("CapabilityScore: want 2, got %d", sc.CapabilityScore)
	}
}

// TestComplexityAnalyzer_DependencyScoreFanOut verifies fan-out (3 outbound deps) yields DependencyScore=3.
func TestComplexityAnalyzer_DependencyScoreFanOut(t *testing.T) {
	m := entity.NewUNMModel("fan-out", "")

	svcA := mustNewService(t, "svc-a", "svc-a", "team-x")
	svcB := mustNewService(t, "svc-b", "svc-b", "team-x")
	svcC := mustNewService(t, "svc-c", "svc-c", "team-x")
	svcD := mustNewService(t, "svc-d", "svc-d", "team-x")

	svcA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-b"), "", ""))
	svcA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-c"), "", ""))
	svcA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-d"), "", ""))

	for _, s := range []*entity.Service{svcA, svcB, svcC, svcD} {
		if err := m.AddService(s); err != nil {
			t.Fatalf("AddService %q: %v", s.Name, err)
		}
	}

	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	var scA ServiceComplexity
	for _, sc := range report.Services {
		if sc.Service.Name == "svc-a" {
			scA = sc
			break
		}
	}
	// fan-out=3, fan-in=0 → DependencyScore=3
	if scA.DependencyScore != 3 {
		t.Errorf("DependencyScore for svc-a: want 3, got %d", scA.DependencyScore)
	}
}

// TestComplexityAnalyzer_DependencyScoreFanIn verifies fan-in (2 other services depend on it) yields DependencyScore=2.
func TestComplexityAnalyzer_DependencyScoreFanIn(t *testing.T) {
	m := entity.NewUNMModel("fan-in", "")

	svcA := mustNewService(t, "svc-a", "svc-a", "team-x") // the hub
	svcB := mustNewService(t, "svc-b", "svc-b", "team-x")
	svcC := mustNewService(t, "svc-c", "svc-c", "team-x")

	// B and C depend on A (fan-in for A = 2)
	svcB.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-a"), "", ""))
	svcC.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-a"), "", ""))

	for _, s := range []*entity.Service{svcA, svcB, svcC} {
		if err := m.AddService(s); err != nil {
			t.Fatalf("AddService %q: %v", s.Name, err)
		}
	}

	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	var scA ServiceComplexity
	for _, sc := range report.Services {
		if sc.Service.Name == "svc-a" {
			scA = sc
			break
		}
	}
	// fan-out=0, fan-in=2 → DependencyScore=2
	if scA.DependencyScore != 2 {
		t.Errorf("DependencyScore for svc-a: want 2, got %d", scA.DependencyScore)
	}
}

// TestComplexityAnalyzer_TotalComplexityWeightedSum verifies the weighted sum formula.
func TestComplexityAnalyzer_TotalComplexityWeightedSum(t *testing.T) {
	m := entity.NewUNMModel("weighted", "")

	// svc-a: fan-out=1 dep, realized by 1 cap, 1 data asset
	svcA := mustNewService(t, "svc-a", "svc-a", "team-x")
	svcB := mustNewService(t, "svc-b", "svc-b", "team-x")
	svcA.AddDependsOn(entity.NewRelationship(mustServiceID(t, "svc-b"), "", ""))

	for _, s := range []*entity.Service{svcA, svcB} {
		if err := m.AddService(s); err != nil {
			t.Fatalf("AddService %q: %v", s.Name, err)
		}
	}

	cap1 := mustNewCapability(t, "cap-1", "cap-1")
	cap1.AddRealizedBy(entity.NewRelationship(mustServiceID(t, "svc-a"), "", ""))
	if err := m.AddCapability(cap1); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	da := mustDataAsset(t, "da-1", "da-1", entity.TypeDatabase)
	da.AddUsedBy("svc-a")
	if err := m.AddDataAsset(da); err != nil {
		t.Fatalf("AddDataAsset: %v", err)
	}

	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	var scA ServiceComplexity
	for _, sc := range report.Services {
		if sc.Service.Name == "svc-a" {
			scA = sc
			break
		}
	}

	// DependencyScore = fan-out(1) + fan-in(0) = 1
	// CapabilityScore = 1
	// DataAssetScore  = 1
	// TotalComplexity = 1*2 + 1*3 + 1*2 = 7
	if scA.DependencyScore != 1 {
		t.Errorf("DependencyScore: want 1, got %d", scA.DependencyScore)
	}
	if scA.CapabilityScore != 1 {
		t.Errorf("CapabilityScore: want 1, got %d", scA.CapabilityScore)
	}
	if scA.DataAssetScore != 1 {
		t.Errorf("DataAssetScore: want 1, got %d", scA.DataAssetScore)
	}
	if scA.TotalComplexity != 7 {
		t.Errorf("TotalComplexity: want 7, got %d", scA.TotalComplexity)
	}
}

// TestComplexityAnalyzer_Ranking verifies services are ranked by TotalComplexity desc, then name asc.
func TestComplexityAnalyzer_Ranking(t *testing.T) {
	m := entity.NewUNMModel("ranking", "")

	// svc-alpha: no deps, no caps → TotalComplexity=0
	svcAlpha := mustNewService(t, "svc-alpha", "svc-alpha", "team-x")
	// svc-beta: 1 cap → TotalComplexity=3
	svcBeta := mustNewService(t, "svc-beta", "svc-beta", "team-x")
	// svc-gamma: 1 cap → TotalComplexity=3 (tie with beta, gamma > beta alphabetically)
	svcGamma := mustNewService(t, "svc-gamma", "svc-gamma", "team-x")

	for _, s := range []*entity.Service{svcAlpha, svcBeta, svcGamma} {
		if err := m.AddService(s); err != nil {
			t.Fatalf("AddService %q: %v", s.Name, err)
		}
	}

	capB := mustNewCapability(t, "cap-b", "cap-b")
	capB.AddRealizedBy(entity.NewRelationship(mustServiceID(t, "svc-beta"), "", ""))
	if err := m.AddCapability(capB); err != nil {
		t.Fatalf("AddCapability cap-b: %v", err)
	}

	capG := mustNewCapability(t, "cap-g", "cap-g")
	capG.AddRealizedBy(entity.NewRelationship(mustServiceID(t, "svc-gamma"), "", ""))
	if err := m.AddCapability(capG); err != nil {
		t.Fatalf("AddCapability cap-g: %v", err)
	}

	a := NewComplexityAnalyzer()
	report := a.Analyze(m)

	if len(report.Services) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(report.Services))
	}

	// Rank 0 and 1: svc-beta and svc-gamma (tied at TotalComplexity=3), alpha < beta alphabetically
	// so beta comes before gamma — both rank before alpha (0).
	if report.Services[0].Service.Name != "svc-beta" {
		t.Errorf("rank 0: want svc-beta, got %s", report.Services[0].Service.Name)
	}
	if report.Services[1].Service.Name != "svc-gamma" {
		t.Errorf("rank 1: want svc-gamma, got %s", report.Services[1].Service.Name)
	}
	if report.Services[2].Service.Name != "svc-alpha" {
		t.Errorf("rank 2: want svc-alpha, got %s", report.Services[2].Service.Name)
	}
}
