package analyzer

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// buildCognitiveLoadModel creates a model where:
//   - team "alpha" (stream-aligned, size 5): 7 caps, 3 services each with 2 deps, 1 collaboration interaction
//   - team "beta" (stream-aligned, size 5): 2 caps, 1 service, 1 x-as-a-service interaction
func buildCognitiveLoadModel(t *testing.T) *entity.UNMModel {
	t.Helper()

	m := entity.NewUNMModel("test-system", "")

	alpha, err := entity.NewTeam("alpha", "alpha", "", valueobject.StreamAligned)
	if err != nil {
		t.Fatalf("NewTeam alpha: %v", err)
	}
	alpha.Size = 5
	alpha.SizeExplicit = true
	beta, err := entity.NewTeam("beta", "beta", "", valueobject.StreamAligned)
	if err != nil {
		t.Fatalf("NewTeam beta: %v", err)
	}
	beta.Size = 5
	beta.SizeExplicit = true

	for i := 1; i <= 7; i++ {
		capName := "alpha-cap-" + string(rune('0'+i))
		cap, err := entity.NewCapability(capName, capName, "")
		if err != nil {
			t.Fatalf("NewCapability %s: %v", capName, err)
		}
		if err := m.AddCapability(cap); err != nil {
			t.Fatalf("AddCapability %s: %v", capName, err)
		}
		capID, err := valueobject.NewEntityID(capName)
		if err != nil {
			t.Fatalf("NewEntityID %s: %v", capName, err)
		}
		alpha.AddOwns(entity.NewRelationship(capID, "", ""))
	}

	for i := 1; i <= 2; i++ {
		capName := "beta-cap-" + string(rune('0'+i))
		cap, err := entity.NewCapability(capName, capName, "")
		if err != nil {
			t.Fatalf("NewCapability %s: %v", capName, err)
		}
		if err := m.AddCapability(cap); err != nil {
			t.Fatalf("AddCapability %s: %v", capName, err)
		}
		capID, err := valueobject.NewEntityID(capName)
		if err != nil {
			t.Fatalf("NewEntityID %s: %v", capName, err)
		}
		beta.AddOwns(entity.NewRelationship(capID, "", ""))
	}

	if err := m.AddTeam(alpha); err != nil {
		t.Fatalf("AddTeam alpha: %v", err)
	}
	if err := m.AddTeam(beta); err != nil {
		t.Fatalf("AddTeam beta: %v", err)
	}

	betaSvc, err := entity.NewService("beta-svc-1", "beta-svc-1", "", "beta")
	if err != nil {
		t.Fatalf("NewService beta-svc-1: %v", err)
	}
	if err := m.AddService(betaSvc); err != nil {
		t.Fatalf("AddService beta-svc-1: %v", err)
	}

	betaSvcID, err := valueobject.NewEntityID("beta-svc-1")
	if err != nil {
		t.Fatalf("NewEntityID beta-svc-1: %v", err)
	}
	for i := 1; i <= 3; i++ {
		svcName := "alpha-svc-" + string(rune('0'+i))
		svc, err := entity.NewService(svcName, svcName, "", "alpha")
		if err != nil {
			t.Fatalf("NewService %s: %v", svcName, err)
		}
		svc.AddDependsOn(entity.NewRelationship(betaSvcID, "dep-1", ""))
		svc.AddDependsOn(entity.NewRelationship(betaSvcID, "dep-2", ""))
		if err := m.AddService(svc); err != nil {
			t.Fatalf("AddService %s: %v", svcName, err)
		}
	}

	// 1 collaboration interaction: alpha → beta (weight 3)
	interaction, err := entity.NewInteraction("i1", "alpha", "beta", valueobject.Collaboration, "", "")
	if err != nil {
		t.Fatalf("NewInteraction: %v", err)
	}
	m.AddInteraction(interaction)

	return m
}

func TestCognitiveLoadAnalyzer_AlphaDimensions(t *testing.T) {
	m := buildCognitiveLoadModel(t)
	a := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights)
	report := a.Analyze(m)

	var alpha *TeamLoad
	for i := range report.TeamLoads {
		if report.TeamLoads[i].Team.Name == "alpha" {
			alpha = &report.TeamLoads[i]
			break
		}
	}
	if alpha == nil {
		t.Fatal("expected team load for alpha, got none")
	}

	if alpha.CapabilityCount != 7 {
		t.Errorf("alpha CapabilityCount: want 7, got %d", alpha.CapabilityCount)
	}
	if alpha.ServiceCount != 3 {
		t.Errorf("alpha ServiceCount: want 3, got %d", alpha.ServiceCount)
	}
	if alpha.DependencyCount != 6 {
		t.Errorf("alpha DependencyCount: want 6 (3 svcs × 2 deps), got %d", alpha.DependencyCount)
	}
	if alpha.InteractionCount != 1 {
		t.Errorf("alpha InteractionCount: want 1, got %d", alpha.InteractionCount)
	}

	// Domain spread: 7 caps → high (threshold 6+)
	if alpha.DomainSpread.Level != LoadHigh {
		t.Errorf("alpha DomainSpread.Level: want high, got %s", alpha.DomainSpread.Level)
	}
	if alpha.DomainSpread.Value != 7 {
		t.Errorf("alpha DomainSpread.Value: want 7, got %d", alpha.DomainSpread.Value)
	}

	// Service load: 3 services / 5 people = 0.6 → low (threshold >2)
	if alpha.ServiceLoad.Level != LoadLow {
		t.Errorf("alpha ServiceLoad.Level: want low, got %s", alpha.ServiceLoad.Level)
	}

	// Interaction load: 1 collaboration interaction = weight 3 → low (threshold ≥4)
	if alpha.InteractionScore != 3 {
		t.Errorf("alpha InteractionScore: want 3, got %d", alpha.InteractionScore)
	}
	if alpha.InteractionLoad.Level != LoadLow {
		t.Errorf("alpha InteractionLoad.Level: want low, got %s", alpha.InteractionLoad.Level)
	}

	// Dependency load: 6 deps → medium (threshold 5-8)
	if alpha.DependencyLoad.Level != LoadMedium {
		t.Errorf("alpha DependencyLoad.Level: want medium, got %d deps", alpha.DependencyCount)
	}

	// Overall: worst is high (from domain spread)
	if alpha.OverallLevel != LoadHigh {
		t.Errorf("alpha OverallLevel: want high, got %s", alpha.OverallLevel)
	}

	if alpha.SizeIsExplicit != true {
		t.Errorf("alpha SizeIsExplicit: want true, got false")
	}
}

func TestCognitiveLoadAnalyzer_BetaDimensions(t *testing.T) {
	m := buildCognitiveLoadModel(t)
	a := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights)
	report := a.Analyze(m)

	var beta *TeamLoad
	for i := range report.TeamLoads {
		if report.TeamLoads[i].Team.Name == "beta" {
			beta = &report.TeamLoads[i]
			break
		}
	}
	if beta == nil {
		t.Fatal("expected team load for beta, got none")
	}

	if beta.CapabilityCount != 2 {
		t.Errorf("beta CapabilityCount: want 2, got %d", beta.CapabilityCount)
	}
	// Domain spread: 2 caps → low
	if beta.DomainSpread.Level != LoadLow {
		t.Errorf("beta DomainSpread.Level: want low, got %s", beta.DomainSpread.Level)
	}
	// Service load: 1 service / 5 people = 0.2 → low
	if beta.ServiceLoad.Level != LoadLow {
		t.Errorf("beta ServiceLoad.Level: want low, got %s", beta.ServiceLoad.Level)
	}
	// Interaction load: 1 xaas = 1 (weight comes from the collab, beta gets weight 3 too as the "to" side)
	// Actually: alpha→beta collaboration, so beta gets weight 3 on the "to" side.
	// That's ≥ 0 and < 4, so low.
	if beta.InteractionLoad.Level != LoadLow {
		t.Errorf("beta InteractionLoad.Level: want low, got %s (score=%d)", beta.InteractionLoad.Level, beta.InteractionScore)
	}
	// Overall: all low → low
	if beta.OverallLevel != LoadLow {
		t.Errorf("beta OverallLevel: want low, got %s", beta.OverallLevel)
	}
}

func TestCognitiveLoadAnalyzer_InteractionModeWeighting(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	team, _ := entity.NewTeam("t1", "t1", "", valueobject.StreamAligned)
	team.Size = 5
	other, _ := entity.NewTeam("t2", "t2", "", valueobject.StreamAligned)
	other.Size = 5
	_ = m.AddTeam(team)
	_ = m.AddTeam(other)

	// 3 collaboration interactions (weight 3 each) = 9 → high (≥7)
	for i := 0; i < 3; i++ {
		ix, _ := entity.NewInteraction("c"+string(rune('0'+i)), "t1", "t2", valueobject.Collaboration, "", "")
		m.AddInteraction(ix)
	}

	report := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights).Analyze(m)
	var t1 *TeamLoad
	for i := range report.TeamLoads {
		if report.TeamLoads[i].Team.Name == "t1" {
			t1 = &report.TeamLoads[i]
		}
	}
	if t1 == nil {
		t.Fatal("t1 not found")
	}
	if t1.InteractionScore != 9 {
		t.Errorf("t1 InteractionScore: want 9 (3 collab × 3), got %d", t1.InteractionScore)
	}
	if t1.InteractionLoad.Level != LoadHigh {
		t.Errorf("t1 InteractionLoad.Level: want high (score 9 ≥ 7), got %s", t1.InteractionLoad.Level)
	}
}

func TestCognitiveLoadAnalyzer_ServiceLoadPerPerson(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	team, _ := entity.NewTeam("small", "small", "", valueobject.StreamAligned)
	team.Size = 2 // small team
	_ = m.AddTeam(team)

	// 7 services / 2 people = 3.5 → high (>3)
	for i := 0; i < 7; i++ {
		svc, _ := entity.NewService("svc-"+string(rune('a'+i)), "svc-"+string(rune('a'+i)), "", "small")
		_ = m.AddService(svc)
	}

	report := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights).Analyze(m)
	tl := report.TeamLoads[0]
	if tl.ServiceLoad.Level != LoadHigh {
		t.Errorf("ServiceLoad.Level: want high (7 svcs / 2 people = 3.5), got %s", tl.ServiceLoad.Level)
	}
}

func TestCognitiveLoadAnalyzer_SortedByOverallThenComposite(t *testing.T) {
	m := buildCognitiveLoadModel(t)
	a := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights)
	report := a.Analyze(m)

	if len(report.TeamLoads) < 2 {
		t.Fatalf("expected at least 2 team loads, got %d", len(report.TeamLoads))
	}
	// alpha has high domain spread → should be first
	if report.TeamLoads[0].Team.Name != "alpha" {
		t.Errorf("expected alpha first (highest load), got %q", report.TeamLoads[0].Team.Name)
	}
	// alpha overall should outrank beta overall
	if levelRank(report.TeamLoads[0].OverallLevel) < levelRank(report.TeamLoads[1].OverallLevel) {
		t.Error("team loads not sorted by overall level descending")
	}
}

func TestCognitiveLoadAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights)
	report := a.Analyze(m)

	if len(report.TeamLoads) != 0 {
		t.Errorf("expected empty report for model with no teams, got %d entries", len(report.TeamLoads))
	}
}

func TestCognitiveLoadAnalyzer_SizeIsExplicit(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	explicit, _ := entity.NewTeam("explicit", "explicit", "", valueobject.StreamAligned)
	explicit.Size = 8
	explicit.SizeExplicit = true
	implicit, _ := entity.NewTeam("implicit", "implicit", "", valueobject.StreamAligned)
	// SizeExplicit stays false (simulates parser NOT setting size)

	_ = m.AddTeam(explicit)
	_ = m.AddTeam(implicit)

	report := NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights).Analyze(m)

	for _, tl := range report.TeamLoads {
		switch tl.Team.Name {
		case "explicit":
			if !tl.SizeIsExplicit {
				t.Error("explicit team should have SizeIsExplicit=true")
			}
			if tl.TeamSize != 8 {
				t.Errorf("explicit team size: want 8, got %d", tl.TeamSize)
			}
		case "implicit":
			if tl.SizeIsExplicit {
				t.Error("implicit team should have SizeIsExplicit=false")
			}
			if tl.TeamSize != 5 {
				t.Errorf("implicit team size: want 5 (default), got %d", tl.TeamSize)
			}
		}
	}
}

func TestInteractionWeight(t *testing.T) {
	tests := []struct {
		mode valueobject.InteractionMode
		want int
	}{
		{valueobject.Collaboration, 3},
		{valueobject.Facilitating, 2},
		{valueobject.XAsAService, 1},
	}
	for _, tt := range tests {
		got := InteractionWeight(tt.mode)
		if got != tt.want {
			t.Errorf("InteractionWeight(%s): want %d, got %d", tt.mode, tt.want, got)
		}
	}
}

func TestLoadDimensionThresholds(t *testing.T) {
	defaultCfg := entity.DefaultConfig().Analysis.CognitiveLoad

	// Domain spread
	for _, tc := range []struct {
		caps int
		want LoadLevel
	}{
		{0, LoadLow}, {3, LoadLow}, {4, LoadMedium}, {5, LoadMedium}, {6, LoadHigh}, {10, LoadHigh},
	} {
		got := assessDomainSpreadCfg(tc.caps, defaultCfg.DomainSpreadThresholds)
		if got.Level != tc.want {
			t.Errorf("assessDomainSpreadCfg(%d): want %s, got %s", tc.caps, tc.want, got.Level)
		}
	}

	// Service load (per person)
	for _, tc := range []struct {
		svcs, size int
		want       LoadLevel
	}{
		{10, 5, LoadLow},  // 2.0 exactly → low
		{11, 5, LoadMedium}, // 2.2 → medium
		{16, 5, LoadHigh},  // 3.2 → high
		{0, 5, LoadLow},
	} {
		got := assessServiceLoadCfg(tc.svcs, tc.size, defaultCfg.ServiceLoadThresholds)
		if got.Level != tc.want {
			t.Errorf("assessServiceLoadCfg(%d, %d): want %s, got %s", tc.svcs, tc.size, tc.want, got.Level)
		}
	}

	// Interaction load (weighted score)
	for _, tc := range []struct {
		score int
		want  LoadLevel
	}{
		{0, LoadLow}, {3, LoadLow}, {4, LoadMedium}, {6, LoadMedium}, {7, LoadHigh}, {15, LoadHigh},
	} {
		got := assessInteractionLoadCfg(tc.score, defaultCfg.InteractionLoadThresholds)
		if got.Level != tc.want {
			t.Errorf("assessInteractionLoadCfg(%d): want %s, got %s", tc.score, tc.want, got.Level)
		}
	}

	// Dependency load
	for _, tc := range []struct {
		deps int
		want LoadLevel
	}{
		{0, LoadLow}, {4, LoadLow}, {5, LoadMedium}, {8, LoadMedium}, {9, LoadHigh}, {20, LoadHigh},
	} {
		got := assessDependencyLoadCfg(tc.deps, defaultCfg.DependencyLoadThresholds)
		if got.Level != tc.want {
			t.Errorf("assessDependencyLoadCfg(%d): want %s, got %s", tc.deps, tc.want, got.Level)
		}
	}
}

// TestCognitiveLoadAnalyzer_CustomThresholds verifies that non-default thresholds change behavior.
func TestCognitiveLoadAnalyzer_CustomThresholds(t *testing.T) {
	m := entity.NewUNMModel("custom", "")
	team, _ := entity.NewTeam("t1", "t1", "", valueobject.StreamAligned)
	team.Size = 5
	// 2 capabilities — would be "low" with defaults [4,6], but "high" with [1,2]
	for i := 0; i < 2; i++ {
		capName := "cap-" + string(rune('a'+i))
		cap, _ := entity.NewCapability(capName, capName, "")
		_ = m.AddCapability(cap)
		capID, _ := valueobject.NewEntityID(capName)
		team.AddOwns(entity.NewRelationship(capID, "", ""))
	}
	_ = m.AddTeam(team)

	cfg := entity.CognitiveLoadConfig{
		DomainSpreadThresholds:    [2]int{1, 2},
		ServiceLoadThresholds:     [2]float64{2.0, 3.0},
		InteractionLoadThresholds: [2]int{4, 7},
		DependencyLoadThresholds:  [2]int{5, 9},
	}
	weights := entity.DefaultConfig().Analysis.InteractionWeights

	a := NewCognitiveLoadAnalyzer(cfg, weights)
	report := a.Analyze(m)

	if len(report.TeamLoads) != 1 {
		t.Fatalf("expected 1 team load, got %d", len(report.TeamLoads))
	}
	tl := report.TeamLoads[0]
	if tl.DomainSpread.Level != LoadHigh {
		t.Errorf("with DomainSpread threshold [1,2], 2 caps should be high, got %s", tl.DomainSpread.Level)
	}
}
