package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func mustInteraction(t *testing.T, id, from, to string, mode valueobject.InteractionMode) *entity.Interaction {
	t.Helper()
	intr, err := entity.NewInteraction(id, from, to, mode, "", "")
	if err != nil {
		t.Fatalf("NewInteraction %q→%q: %v", from, to, err)
	}
	return intr
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestInteractionDiversityAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	if len(r.ModeDistribution) != 0 {
		t.Errorf("want empty mode dist, got %v", r.ModeDistribution)
	}
	if r.AllModesSame {
		t.Errorf("AllModesSame should be false with no interactions")
	}
	if len(r.IsolatedTeams) != 0 {
		t.Errorf("want no isolated teams (no teams registered), got %v", r.IsolatedTeams)
	}
}

func TestInteractionDiversityAnalyzer_DiverseModes(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	// Three teams, three different modes
	m.AddInteraction(mustInteraction(t, "i1", "team-a", "team-b", valueobject.XAsAService))
	m.AddInteraction(mustInteraction(t, "i2", "team-b", "team-c", valueobject.Collaboration))
	m.AddInteraction(mustInteraction(t, "i3", "team-a", "team-c", valueobject.Facilitating))

	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	if r.AllModesSame {
		t.Errorf("AllModesSame should be false with three distinct modes")
	}
	if r.ModeDistribution[valueobject.XAsAService] != 1 {
		t.Errorf("want x-as-a-service=1, got %d", r.ModeDistribution[valueobject.XAsAService])
	}
	if r.ModeDistribution[valueobject.Collaboration] != 1 {
		t.Errorf("want collaboration=1, got %d", r.ModeDistribution[valueobject.Collaboration])
	}
	if r.ModeDistribution[valueobject.Facilitating] != 1 {
		t.Errorf("want facilitating=1, got %d", r.ModeDistribution[valueobject.Facilitating])
	}
}

func TestInteractionDiversityAnalyzer_AllSameMode(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	m.AddInteraction(mustInteraction(t, "i1", "team-a", "team-b", valueobject.XAsAService))
	m.AddInteraction(mustInteraction(t, "i2", "team-b", "team-c", valueobject.XAsAService))
	m.AddInteraction(mustInteraction(t, "i3", "team-c", "team-d", valueobject.XAsAService))

	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	if !r.AllModesSame {
		t.Errorf("AllModesSame should be true when all interactions use x-as-a-service")
	}
	if r.ModeDistribution[valueobject.XAsAService] != 3 {
		t.Errorf("want x-as-a-service=3, got %d", r.ModeDistribution[valueobject.XAsAService])
	}
	if len(r.ModeDistribution) != 1 {
		t.Errorf("want exactly 1 mode, got %d", len(r.ModeDistribution))
	}
}

func TestInteractionDiversityAnalyzer_IsolatedTeams(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// Register three teams but only two interact
	teamA := mustTeam(t, "team-a")
	teamB := mustTeam(t, "team-b")
	teamC := mustTeam(t, "team-c") // isolated

	addTeamToModel(t, m, teamA)
	addTeamToModel(t, m, teamB)
	addTeamToModel(t, m, teamC)

	m.AddInteraction(mustInteraction(t, "i1", "team-a", "team-b", valueobject.XAsAService))

	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	if len(r.IsolatedTeams) != 1 {
		t.Fatalf("want 1 isolated team, got %v", r.IsolatedTeams)
	}
	if r.IsolatedTeams[0] != "team-c" {
		t.Errorf("want isolated=team-c, got %q", r.IsolatedTeams[0])
	}
}

func TestInteractionDiversityAnalyzer_AllTeamsIsolated(t *testing.T) {
	m := entity.NewUNMModel("sys", "")
	addTeamToModel(t, m, mustTeam(t, "team-x"))
	addTeamToModel(t, m, mustTeam(t, "team-y"))

	// No interactions registered
	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	if len(r.IsolatedTeams) != 2 {
		t.Errorf("want 2 isolated teams, got %v", r.IsolatedTeams)
	}
	// Sorted
	if r.IsolatedTeams[0] != "team-x" || r.IsolatedTeams[1] != "team-y" {
		t.Errorf("want sorted [team-x team-y], got %v", r.IsolatedTeams)
	}
}

func TestInteractionDiversityAnalyzer_OverReliantTeam(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// team-hub is the to-side of 5 x-as-a-service interactions
	for i := 0; i < 5; i++ {
		from := "team-consumer-" + string(rune('a'+i))
		m.AddInteraction(mustInteraction(t, "i"+string(rune('0'+i)), from, "team-hub", valueobject.XAsAService))
	}

	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	if len(r.OverReliantTeams) == 0 {
		t.Fatal("expected at least one over-reliant team")
	}
	// Find team-hub in results
	var found *TeamModeOverload
	for i := range r.OverReliantTeams {
		if r.OverReliantTeams[i].TeamName == "team-hub" {
			found = &r.OverReliantTeams[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("team-hub not in over-reliant list: %v", r.OverReliantTeams)
	}
	if found.Mode != valueobject.XAsAService {
		t.Errorf("want mode=x-as-a-service, got %q", found.Mode)
	}
	if found.Count < 4 {
		t.Errorf("want count>=4, got %d", found.Count)
	}
}

func TestInteractionDiversityAnalyzer_BelowOverloadThreshold(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// 3 interactions — below threshold of 4
	for i := 0; i < 3; i++ {
		from := "team-" + string(rune('a'+i))
		m.AddInteraction(mustInteraction(t, "i"+string(rune('0'+i)), from, "team-hub", valueobject.XAsAService))
	}

	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)

	// team-hub has 3 x-as-a-service interactions — should NOT be over-reliant
	for _, ot := range r.OverReliantTeams {
		if ot.TeamName == "team-hub" && ot.Mode == valueobject.XAsAService {
			t.Errorf("team-hub with 3 interactions should not be over-reliant (threshold is 4)")
		}
	}
}

func TestInteractionDiversityAnalyzer_OverReliantSortedDeterministic(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// Two teams each with 4 x-as-a-service
	for i := 0; i < 4; i++ {
		m.AddInteraction(mustInteraction(t, "ia"+string(rune('0'+i)), "team-alpha-"+string(rune('a'+i)), "team-hub-1", valueobject.XAsAService))
		m.AddInteraction(mustInteraction(t, "ib"+string(rune('0'+i)), "team-beta-"+string(rune('a'+i)), "team-hub-2", valueobject.XAsAService))
	}

	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r1 := a.Analyze(m)
	r2 := a.Analyze(m)

	if len(r1.OverReliantTeams) != len(r2.OverReliantTeams) {
		t.Errorf("non-deterministic over-reliant list lengths")
	}
	for i := range r1.OverReliantTeams {
		if r1.OverReliantTeams[i].TeamName != r2.OverReliantTeams[i].TeamName {
			t.Errorf("non-deterministic sort at index %d: %q vs %q", i, r1.OverReliantTeams[i].TeamName, r2.OverReliantTeams[i].TeamName)
		}
	}
}

// TestInteractionDiversityAnalyzer_CustomThresholds verifies that a non-default
// InteractionOverReliance threshold changes which teams are flagged.
func TestInteractionDiversityAnalyzer_CustomThresholds(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// team-a has 2 x-as-a-service interactions (below default 4, above custom 1)
	m.AddInteraction(mustInteraction(t, "i1", "team-a", "team-b", valueobject.XAsAService))
	m.AddInteraction(mustInteraction(t, "i2", "team-a", "team-c", valueobject.XAsAService))

	// With default threshold=4, 2 interactions should NOT trigger over-reliance.
	a := NewInteractionDiversityAnalyzer(entity.DefaultConfig().Analysis.Signals)
	r := a.Analyze(m)
	if len(r.OverReliantTeams) != 0 {
		t.Errorf("with default threshold=4, expected no over-reliant teams, got %d", len(r.OverReliantTeams))
	}

	// With custom threshold=1, 2 interactions SHOULD trigger over-reliance.
	customCfg := entity.DefaultConfig().Analysis.Signals
	customCfg.InteractionOverReliance = 1
	a2 := NewInteractionDiversityAnalyzer(customCfg)
	r2 := a2.Analyze(m)
	if len(r2.OverReliantTeams) == 0 {
		t.Error("with InteractionOverReliance=1, 2 same-mode interactions should trigger over-reliance")
	}
}
