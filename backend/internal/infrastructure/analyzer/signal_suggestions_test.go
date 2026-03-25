package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// buildMinimalBottleneckReport creates a BottleneckReport with a critical bottleneck.
func buildCriticalBottleneckReport(svcName string) BottleneckReport {
	svc, _ := entity.NewService(svcName, svcName, "", "team-a")
	return BottleneckReport{
		ServiceBottlenecks: []ServiceBottleneck{
			{Service: svc, FanIn: 12, FanOut: 2, IsCritical: true},
		},
	}
}

func buildWarningBottleneckReport(svcName string) BottleneckReport {
	svc, _ := entity.NewService(svcName, svcName, "", "team-a")
	return BottleneckReport{
		ServiceBottlenecks: []ServiceBottleneck{
			{Service: svc, FanIn: 7, FanOut: 1, IsWarning: true},
		},
	}
}

func buildHighLoadReport(teamName string) CognitiveLoadReport {
	team, _ := entity.NewTeam(teamName, teamName, "", valueobject.StreamAligned)
	return CognitiveLoadReport{
		TeamLoads: []TeamLoad{
			{
				Team:            team,
				CapabilityCount: 8,
				DomainSpread:    LoadDimension{Value: 8, Level: LoadHigh},
				ServiceLoad:     LoadDimension{Value: 2, Level: LoadLow},
				InteractionLoad: LoadDimension{Value: 3, Level: LoadLow},
				DependencyLoad:  LoadDimension{Value: 3, Level: LoadLow},
				OverallLevel:    LoadHigh,
			},
		},
	}
}

func buildLowLoadReport(teamName string) CognitiveLoadReport {
	team, _ := entity.NewTeam(teamName, teamName, "", valueobject.StreamAligned)
	return CognitiveLoadReport{
		TeamLoads: []TeamLoad{
			{
				Team:            team,
				CapabilityCount: 2,
				DomainSpread:    LoadDimension{Value: 2, Level: LoadLow},
				ServiceLoad:     LoadDimension{Value: 1, Level: LoadLow},
				InteractionLoad: LoadDimension{Value: 1, Level: LoadLow},
				DependencyLoad:  LoadDimension{Value: 1, Level: LoadLow},
				OverallLevel:    LoadLow,
			},
		},
	}
}

func emptyBottleneck() BottleneckReport        { return BottleneckReport{} }
func emptyCogLoad() CognitiveLoadReport        { return CognitiveLoadReport{} }
func emptyFrag() FragmentationReport           { return FragmentationReport{} }
func emptyDep() DependencyReport               { return DependencyReport{} }
func emptyUnlinked() UnlinkedCapabilityReport  { return UnlinkedCapabilityReport{} }

// ── tests ─────────────────────────────────────────────────────────────────────

func TestSignalSuggestionGenerator_EmptyReports(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), emptyCogLoad(), emptyFrag(), emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) != 0 {
		t.Errorf("want 0 suggestions from empty reports, got %d", len(r.Suggestions))
	}
}

func TestSignalSuggestionGenerator_CriticalBottleneck(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(buildCriticalBottleneckReport("inca-core"), emptyCogLoad(), emptyFrag(), emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(r.Suggestions))
	}
	s := r.Suggestions[0]
	if s.Category != entity.CategoryBottleneck {
		t.Errorf("want category=bottleneck, got %q", s.Category)
	}
	if s.OnEntityName != "inca-core" {
		t.Errorf("want entity=inca-core, got %q", s.OnEntityName)
	}
	if s.Severity != valueobject.SeverityCritical {
		t.Errorf("want severity=critical, got %q", s.Severity)
	}
	if s.Source != "bottleneck-analyzer" {
		t.Errorf("want source=bottleneck-analyzer, got %q", s.Source)
	}
}

func TestSignalSuggestionGenerator_WarningBottleneckIsHigh(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(buildWarningBottleneckReport("svc-b"), emptyCogLoad(), emptyFrag(), emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(r.Suggestions))
	}
	if r.Suggestions[0].Severity != valueobject.SeverityHigh {
		t.Errorf("want severity=high for warning bottleneck, got %q", r.Suggestions[0].Severity)
	}
}

func TestSignalSuggestionGenerator_CognitiveLoadHighLevel(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), buildHighLoadReport("inca-publisher-dev"), emptyFrag(), emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(r.Suggestions))
	}
	s := r.Suggestions[0]
	if s.Category != entity.CategoryCognitiveLoad {
		t.Errorf("want category=cognitive-load, got %q", s.Category)
	}
	if s.OnEntityName != "inca-publisher-dev" {
		t.Errorf("want entity=inca-publisher-dev, got %q", s.OnEntityName)
	}
	if s.Severity != valueobject.SeverityHigh {
		t.Errorf("want severity=high, got %q", s.Severity)
	}
}

func TestSignalSuggestionGenerator_CognitiveLoadLowLevelNotSuggested(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), buildLowLoadReport("ok-team"), emptyFrag(), emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) != 0 {
		t.Errorf("low-level load should not generate suggestion, got %d suggestions", len(r.Suggestions))
	}
}

func TestSignalSuggestionGenerator_FragmentedCapability(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	cap, _ := entity.NewCapability("async-write", "Async Entity Writing", "")
	team1, _ := entity.NewTeam("t1", "inca-dev", "", valueobject.StreamAligned)
	team2, _ := entity.NewTeam("t2", "inca-publisher-dev", "", valueobject.StreamAligned)
	team3, _ := entity.NewTeam("t3", "inca-serving", "", valueobject.StreamAligned)

	fragReport := FragmentationReport{
		FragmentedCapabilities: []FragmentedCapability{
			{Capability: cap, Teams: []*entity.Team{team1, team2, team3}},
		},
	}

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), emptyCogLoad(), fragReport, emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(r.Suggestions))
	}
	s := r.Suggestions[0]
	if s.Category != entity.CategoryFragmentation {
		t.Errorf("want category=fragmentation, got %q", s.Category)
	}
	if s.Severity != valueobject.SeverityHigh {
		t.Errorf("want severity=high, got %q", s.Severity)
	}
}

func TestSignalSuggestionGenerator_DeepDependencyChain(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	depReport := DependencyReport{
		MaxServiceDepth:     6,
		CriticalServicePath: []string{"svc-a", "svc-b", "svc-c", "svc-d", "svc-e", "svc-f"},
	}

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), emptyCogLoad(), emptyFrag(), depReport, emptyUnlinked(), m)

	if len(r.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(r.Suggestions))
	}
	s := r.Suggestions[0]
	if s.Category != entity.CategoryCoupling {
		t.Errorf("want category=coupling, got %q", s.Category)
	}
	if s.OnEntityName != "svc-f" {
		t.Errorf("want deepest service svc-f, got %q", s.OnEntityName)
	}
	if s.Severity != valueobject.SeverityMedium {
		t.Errorf("want severity=medium for coupling, got %q", s.Severity)
	}
}

func TestSignalSuggestionGenerator_DepthAtThresholdNotSuggested(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	depReport := DependencyReport{
		MaxServiceDepth:     4,
		CriticalServicePath: []string{"svc-a", "svc-b", "svc-c", "svc-d"},
	}

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), emptyCogLoad(), emptyFrag(), depReport, emptyUnlinked(), m)

	// Depth == threshold → no suggestion
	if len(r.Suggestions) != 0 {
		t.Errorf("depth at threshold should not generate suggestion, got %d", len(r.Suggestions))
	}
}

func TestSignalSuggestionGenerator_UnlinkedDomainCapSuggested(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	cap, _ := entity.NewCapability("catalog-idx", "Catalog Indexing", "")
	_ = cap.SetVisibility("domain")

	unlinked := UnlinkedCapabilityReport{
		UnlinkedLeafCapabilities: []UnlinkedCapability{
			{Capability: cap, Visibility: "domain", IsExpected: false},
		},
	}

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), emptyCogLoad(), emptyFrag(), emptyDep(), unlinked, m)

	if len(r.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(r.Suggestions))
	}
	s := r.Suggestions[0]
	if s.Category != entity.CategoryGap {
		t.Errorf("want category=gap, got %q", s.Category)
	}
	if s.Severity != valueobject.SeverityHigh {
		t.Errorf("want severity=high, got %q", s.Severity)
	}
}

func TestSignalSuggestionGenerator_InfrastructureUnlinkedNotSuggested(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	cap, _ := entity.NewCapability("blob", "Blob Storage", "")
	_ = cap.SetVisibility("infrastructure")

	unlinked := UnlinkedCapabilityReport{
		UnlinkedLeafCapabilities: []UnlinkedCapability{
			{Capability: cap, Visibility: "infrastructure", IsExpected: true},
		},
	}

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(emptyBottleneck(), emptyCogLoad(), emptyFrag(), emptyDep(), unlinked, m)

	if len(r.Suggestions) != 0 {
		t.Errorf("infrastructure unlinked cap should not generate suggestion, got %d", len(r.Suggestions))
	}
}

func TestSignalSuggestionGenerator_NoDuplicatesWhenSignalExists(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	// Pre-existing signal for inca-core bottleneck
	existing, err := entity.NewSignal("sig1", entity.CategoryBottleneck, "inca-core", "existing bottleneck", "known", valueobject.SeverityCritical)
	if err != nil {
		t.Fatalf("NewSignal: %v", err)
	}
	m.Signals = append(m.Signals, existing)

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(buildCriticalBottleneckReport("inca-core"), emptyCogLoad(), emptyFrag(), emptyDep(), emptyUnlinked(), m)

	// Should produce 0 new suggestions since inca-core bottleneck already exists
	if len(r.Suggestions) != 0 {
		t.Errorf("want 0 suggestions (already exists), got %d", len(r.Suggestions))
	}
}

func TestSignalSuggestionGenerator_SortOrder(t *testing.T) {
	// Critical bottleneck + high cognitive-load → critical first
	m := entity.NewUNMModel("sys", "")

	bottleneck := buildCriticalBottleneckReport("inca-core")
	cogLoad := buildHighLoadReport("heavy-team")

	g := NewSignalSuggestionGenerator(entity.DefaultConfig().Analysis.Signals)
	r := g.Generate(bottleneck, cogLoad, emptyFrag(), emptyDep(), emptyUnlinked(), m)

	if len(r.Suggestions) < 2 {
		t.Fatalf("want at least 2 suggestions, got %d", len(r.Suggestions))
	}
	if r.Suggestions[0].Severity != valueobject.SeverityCritical {
		t.Errorf("first suggestion should be critical, got %q", r.Suggestions[0].Severity)
	}
}
