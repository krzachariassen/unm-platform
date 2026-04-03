package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func mustNeed(t *testing.T, name, actor, outcome string) *entity.Need {
	t.Helper()
	n, err := entity.NewNeed(name, name, actor, outcome)
	if err != nil {
		t.Fatalf("NewNeed %s: %v", name, err)
	}
	return n
}

func mustCap(t *testing.T, name string) *entity.Capability {
	t.Helper()
	c, err := entity.NewCapability(name, name, "")
	if err != nil {
		t.Fatalf("NewCapability %s: %v", name, err)
	}
	return c
}

func mustTeamWithType(t *testing.T, name string, tt valueobject.TeamType) *entity.Team {
	t.Helper()
	team, err := entity.NewTeam(name, name, "", tt)
	if err != nil {
		t.Fatalf("NewTeam %s: %v", name, err)
	}
	return team
}

func mustID(t *testing.T, id string) valueobject.EntityID {
	t.Helper()
	eid, err := valueobject.NewEntityID(id)
	if err != nil {
		t.Fatalf("NewEntityID %s: %v", id, err)
	}
	return eid
}

func vsRel(t *testing.T, target string) entity.Relationship {
	t.Helper()
	return entity.NewRelationship(mustID(t, target), "", "")
}

// TestValueStreamAnalyzer_HighCoherence: team serves 3 needs sharing caps → high coherence
func TestValueStreamAnalyzer_HighCoherence(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	capA := mustCap(t, "cap-a")
	capB := mustCap(t, "cap-b")
	capC := mustCap(t, "cap-c")

	// All caps realized by same service owned by team-alpha
	svcA := mustService(t, "svc-a", "team-alpha")
	svcA.AddRealizes(vsRel(t, "cap-a"))
	svcA.AddRealizes(vsRel(t, "cap-b"))
	svcA.AddRealizes(vsRel(t, "cap-c"))

	// need-1 → cap-a, cap-b; need-2 → cap-a, cap-c; need-3 → cap-b, cap-c
	// All pairs share at least one cap → SharedCapEdges = 3, MaxPossible = 3, score = 1.0
	need1 := mustNeed(t, "need-1", "actor-1", "")
	need1.AddSupportedBy(vsRel(t, "cap-a"))
	need1.AddSupportedBy(vsRel(t, "cap-b"))

	need2 := mustNeed(t, "need-2", "actor-1", "")
	need2.AddSupportedBy(vsRel(t, "cap-a"))
	need2.AddSupportedBy(vsRel(t, "cap-c"))

	need3 := mustNeed(t, "need-3", "actor-1", "")
	need3.AddSupportedBy(vsRel(t, "cap-b"))
	need3.AddSupportedBy(vsRel(t, "cap-c"))

	team := mustTeam(t, "team-alpha")

	for _, x := range []*entity.Capability{capA, capB, capC} {
		if err := m.AddCapability(x); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.AddService(svcA); err != nil {
		t.Fatal(err)
	}
	for _, n := range []*entity.Need{need1, need2, need3} {
		if err := m.AddNeed(n); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.AddTeam(team); err != nil {
		t.Fatal(err)
	}

	a := NewValueStreamAnalyzer()
	report := a.Analyze(m)

	if len(report.TeamCoherences) != 1 {
		t.Fatalf("expected 1 team coherence, got %d", len(report.TeamCoherences))
	}
	tc := report.TeamCoherences[0]
	if tc.TeamName != "team-alpha" {
		t.Errorf("expected team-alpha, got %s", tc.TeamName)
	}
	if tc.NeedCount != 3 {
		t.Errorf("expected 3 needs, got %d", tc.NeedCount)
	}
	if tc.SharedCapEdges != 3 {
		t.Errorf("expected 3 shared cap edges, got %d", tc.SharedCapEdges)
	}
	if tc.CoherenceScore != 1.0 {
		t.Errorf("expected coherence 1.0, got %f", tc.CoherenceScore)
	}
	if tc.LowCoherence {
		t.Error("expected LowCoherence=false")
	}
	if report.LowCoherenceCount != 0 {
		t.Errorf("expected 0 low coherence teams, got %d", report.LowCoherenceCount)
	}
}

// TestValueStreamAnalyzer_LowCoherence: team serves 3 unrelated needs (no shared caps) → low coherence, flagged
func TestValueStreamAnalyzer_LowCoherence(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	capA := mustCap(t, "cap-a")
	capB := mustCap(t, "cap-b")
	capC := mustCap(t, "cap-c")

	svc := mustService(t, "svc-x", "team-beta")
	svc.AddRealizes(vsRel(t, "cap-a"))
	svc.AddRealizes(vsRel(t, "cap-b"))
	svc.AddRealizes(vsRel(t, "cap-c"))

	need1 := mustNeed(t, "need-1", "actor-1", "")
	need1.AddSupportedBy(vsRel(t, "cap-a"))

	need2 := mustNeed(t, "need-2", "actor-1", "")
	need2.AddSupportedBy(vsRel(t, "cap-b"))

	need3 := mustNeed(t, "need-3", "actor-1", "")
	need3.AddSupportedBy(vsRel(t, "cap-c"))

	team := mustTeam(t, "team-beta")

	for _, x := range []*entity.Capability{capA, capB, capC} {
		if err := m.AddCapability(x); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.AddService(svc); err != nil {
		t.Fatal(err)
	}
	for _, n := range []*entity.Need{need1, need2, need3} {
		if err := m.AddNeed(n); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.AddTeam(team); err != nil {
		t.Fatal(err)
	}

	a := NewValueStreamAnalyzer()
	report := a.Analyze(m)

	if len(report.TeamCoherences) != 1 {
		t.Fatalf("expected 1 team coherence, got %d", len(report.TeamCoherences))
	}
	tc := report.TeamCoherences[0]
	if tc.NeedCount != 3 {
		t.Errorf("expected 3 needs, got %d", tc.NeedCount)
	}
	if tc.SharedCapEdges != 0 {
		t.Errorf("expected 0 shared cap edges, got %d", tc.SharedCapEdges)
	}
	if tc.CoherenceScore != 0.0 {
		t.Errorf("expected coherence 0.0, got %f", tc.CoherenceScore)
	}
	if !tc.LowCoherence {
		t.Error("expected LowCoherence=true")
	}
	if report.LowCoherenceCount != 1 {
		t.Errorf("expected 1 low coherence team, got %d", report.LowCoherenceCount)
	}
}

// TestValueStreamAnalyzer_SingleNeed: team serves 1 need → coherence=1.0, not flagged
func TestValueStreamAnalyzer_SingleNeed(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	cap := mustCap(t, "cap-a")
	svc := mustService(t, "svc-1", "team-one")
	svc.AddRealizes(vsRel(t, "cap-a"))

	need := mustNeed(t, "need-1", "actor-1", "")
	need.AddSupportedBy(vsRel(t, "cap-a"))

	team := mustTeam(t, "team-one")

	if err := m.AddCapability(cap); err != nil {
		t.Fatal(err)
	}
	if err := m.AddService(svc); err != nil {
		t.Fatal(err)
	}
	if err := m.AddNeed(need); err != nil {
		t.Fatal(err)
	}
	if err := m.AddTeam(team); err != nil {
		t.Fatal(err)
	}

	a := NewValueStreamAnalyzer()
	report := a.Analyze(m)

	if len(report.TeamCoherences) != 1 {
		t.Fatalf("expected 1 team coherence, got %d", len(report.TeamCoherences))
	}
	tc := report.TeamCoherences[0]
	if tc.CoherenceScore != 1.0 {
		t.Errorf("expected coherence 1.0, got %f", tc.CoherenceScore)
	}
	if tc.LowCoherence {
		t.Error("expected LowCoherence=false")
	}
}

// TestValueStreamAnalyzer_SkipsNonStreamTeams: platform team not in report
func TestValueStreamAnalyzer_SkipsNonStreamTeams(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	platformTeam := mustTeamWithType(t, "platform-team", valueobject.Platform)
	enablingTeam := mustTeamWithType(t, "enabling-team", valueobject.Enabling)

	if err := m.AddTeam(platformTeam); err != nil {
		t.Fatal(err)
	}
	if err := m.AddTeam(enablingTeam); err != nil {
		t.Fatal(err)
	}

	a := NewValueStreamAnalyzer()
	report := a.Analyze(m)

	if len(report.TeamCoherences) != 0 {
		t.Errorf("expected 0 team coherences (non-stream teams skipped), got %d", len(report.TeamCoherences))
	}
}

// TestValueStreamAnalyzer_EmptyModel: empty → empty report
func TestValueStreamAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	a := NewValueStreamAnalyzer()
	report := a.Analyze(m)

	if len(report.TeamCoherences) != 0 {
		t.Errorf("expected 0 team coherences, got %d", len(report.TeamCoherences))
	}
	if report.LowCoherenceCount != 0 {
		t.Errorf("expected 0 low coherence count, got %d", report.LowCoherenceCount)
	}
}

// TestValueStreamAnalyzer_TeamServingNoNeeds: stream-aligned team with no need connections → coherence=1.0, not flagged
func TestValueStreamAnalyzer_TeamServingNoNeeds(t *testing.T) {
	m := entity.NewUNMModel("test", "")

	team := mustTeam(t, "lonely-team")
	svc := mustService(t, "lonely-svc", "lonely-team")

	if err := m.AddTeam(team); err != nil {
		t.Fatal(err)
	}
	if err := m.AddService(svc); err != nil {
		t.Fatal(err)
	}

	a := NewValueStreamAnalyzer()
	report := a.Analyze(m)

	if len(report.TeamCoherences) != 1 {
		t.Fatalf("expected 1 team coherence, got %d", len(report.TeamCoherences))
	}
	tc := report.TeamCoherences[0]
	if tc.NeedCount != 0 {
		t.Errorf("expected 0 needs, got %d", tc.NeedCount)
	}
	if tc.CoherenceScore != 1.0 {
		t.Errorf("expected coherence 1.0, got %f", tc.CoherenceScore)
	}
	if tc.LowCoherence {
		t.Error("expected LowCoherence=false")
	}
}
