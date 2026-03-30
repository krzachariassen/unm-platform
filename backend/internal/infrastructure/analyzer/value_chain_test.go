package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// helper to create a minimal model with a need → capability → service → team chain.
func buildValueChainModel(t *testing.T, cfg valueChainTestCfg) *entity.UNMModel {
	t.Helper()
	m := entity.NewUNMModel("test-system", "")

	for _, tm := range cfg.teams {
		team, err := entity.NewTeam(tm.name, tm.name, "", valueobject.StreamAligned)
		if err != nil {
			t.Fatalf("NewTeam %s: %v", tm.name, err)
		}
		team.Size = tm.size
		team.SizeExplicit = true
		if err := m.AddTeam(team); err != nil {
			t.Fatalf("AddTeam %s: %v", tm.name, err)
		}
	}

	for _, svc := range cfg.services {
		s, err := entity.NewService(svc.name, svc.name, "", svc.owner)
		if err != nil {
			t.Fatalf("NewService %s: %v", svc.name, err)
		}
		if err := m.AddService(s); err != nil {
			t.Fatalf("AddService %s: %v", svc.name, err)
		}
	}

	for _, cap := range cfg.capabilities {
		c, err := entity.NewCapability(cap.name, cap.name, "")
		if err != nil {
			t.Fatalf("NewCapability %s: %v", cap.name, err)
		}
		for _, svcName := range cap.realizedBy {
			svcID, err := valueobject.NewEntityID(svcName)
			if err != nil {
				t.Fatalf("NewEntityID %s: %v", svcName, err)
			}
			c.AddRealizedBy(entity.NewRelationship(svcID, "", ""))
		}
		if err := m.AddCapability(c); err != nil {
			t.Fatalf("AddCapability %s: %v", cap.name, err)
		}
	}

	for _, n := range cfg.needs {
		need, err := entity.NewNeed(n.name, n.name, n.actor, "")
		if err != nil {
			t.Fatalf("NewNeed %s: %v", n.name, err)
		}
		for _, capName := range n.supportedBy {
			capID, err := valueobject.NewEntityID(capName)
			if err != nil {
				t.Fatalf("NewEntityID %s: %v", capName, err)
			}
			need.AddSupportedBy(entity.NewRelationship(capID, "", ""))
		}
		if err := m.AddNeed(need); err != nil {
			t.Fatalf("AddNeed %s: %v", n.name, err)
		}
	}

	// Add actor entries so the model is valid (not strictly required for analyzer, but good form)
	for _, a := range cfg.actors {
		actor, err := entity.NewActor(a, a, "")
		if err != nil {
			t.Fatalf("NewActor %s: %v", a, err)
		}
		if err := m.AddActor(&actor); err != nil {
			t.Fatalf("AddActor %s: %v", a, err)
		}
	}

	// If we need overloaded teams for cognitive load, add extra capabilities to team.Owns
	for _, tm := range cfg.teams {
		if tm.ownsCaps > 0 {
			team := m.Teams[tm.name]
			for i := 0; i < tm.ownsCaps; i++ {
				capName := tm.name + "-owned-cap-" + string(rune('a'+i))
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
				team.AddOwns(entity.NewRelationship(capID, "", ""))
			}
		}
	}

	// Add interactions if needed for cognitive load
	for _, ix := range cfg.interactions {
		interaction, err := entity.NewInteraction(ix.from+"-"+ix.to, ix.from, ix.to, ix.mode, "", "")
		if err != nil {
			t.Fatalf("NewInteraction: %v", err)
		}
		m.AddInteraction(interaction)
	}

	return m
}

type teamCfg struct {
	name     string
	size     int
	ownsCaps int // number of extra caps to add to team.Owns (for cognitive load)
}

type serviceCfg struct {
	name  string
	owner string
}

type capCfg struct {
	name       string
	realizedBy []string
}

type needCfg struct {
	name        string
	actor       string
	supportedBy []string
}

type interactionCfg struct {
	from string
	to   string
	mode valueobject.InteractionMode
}

type valueChainTestCfg struct {
	actors       []string
	teams        []teamCfg
	services     []serviceCfg
	capabilities []capCfg
	needs        []needCfg
	interactions []interactionCfg
}

func TestValueChainAnalyzer_SingleTeam(t *testing.T) {
	// Need backed by 1 team → span=1, not cross-team, not at-risk
	cfg := valueChainTestCfg{
		actors: []string{"merchant"},
		teams:  []teamCfg{{name: "team-a", size: 5}},
		services: []serviceCfg{
			{name: "svc-1", owner: "team-a"},
		},
		capabilities: []capCfg{
			{name: "cap-1", realizedBy: []string{"svc-1"}},
		},
		needs: []needCfg{
			{name: "upload-menu", actor: "merchant", supportedBy: []string{"cap-1"}},
		},
	}

	m := buildValueChainModel(t, cfg)
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	if len(report.NeedRisks) != 1 {
		t.Fatalf("expected 1 need risk, got %d", len(report.NeedRisks))
	}

	nr := report.NeedRisks[0]
	if nr.NeedName != "upload-menu" {
		t.Errorf("NeedName: want upload-menu, got %s", nr.NeedName)
	}
	if nr.ActorNames[0] != "merchant" {
		t.Errorf("ActorNames[0]: want merchant, got %s", nr.ActorNames[0])
	}
	if nr.TeamSpan != 1 {
		t.Errorf("TeamSpan: want 1, got %d", nr.TeamSpan)
	}
	if nr.CapabilityCount != 1 {
		t.Errorf("CapabilityCount: want 1, got %d", nr.CapabilityCount)
	}
	if nr.ServiceCount != 1 {
		t.Errorf("ServiceCount: want 1, got %d", nr.ServiceCount)
	}
	if nr.CrossTeam {
		t.Error("CrossTeam: want false, got true")
	}
	if nr.AtRisk {
		t.Error("AtRisk: want false, got true")
	}
	if nr.Unbacked {
		t.Error("Unbacked: want false, got true")
	}
	if report.CrossTeamNeedCount != 0 {
		t.Errorf("CrossTeamNeedCount: want 0, got %d", report.CrossTeamNeedCount)
	}
}

func TestValueChainAnalyzer_CrossTeam(t *testing.T) {
	// Need backed by 2 teams → cross_team=true, but NOT at-risk (only 2 teams)
	cfg := valueChainTestCfg{
		actors: []string{"eater"},
		teams: []teamCfg{
			{name: "team-a", size: 5},
			{name: "team-b", size: 5},
		},
		services: []serviceCfg{
			{name: "svc-1", owner: "team-a"},
			{name: "svc-2", owner: "team-b"},
		},
		capabilities: []capCfg{
			{name: "cap-1", realizedBy: []string{"svc-1"}},
			{name: "cap-2", realizedBy: []string{"svc-2"}},
		},
		needs: []needCfg{
			{name: "order-food", actor: "eater", supportedBy: []string{"cap-1", "cap-2"}},
		},
	}

	m := buildValueChainModel(t, cfg)
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	if len(report.NeedRisks) != 1 {
		t.Fatalf("expected 1 need risk, got %d", len(report.NeedRisks))
	}

	nr := report.NeedRisks[0]
	if nr.TeamSpan != 2 {
		t.Errorf("TeamSpan: want 2, got %d", nr.TeamSpan)
	}
	if !nr.CrossTeam {
		t.Error("CrossTeam: want true, got false")
	}
	if nr.AtRisk {
		t.Error("AtRisk: want false for 2 teams (threshold is >2)")
	}
	if nr.CapabilityCount != 2 {
		t.Errorf("CapabilityCount: want 2, got %d", nr.CapabilityCount)
	}
	if nr.ServiceCount != 2 {
		t.Errorf("ServiceCount: want 2, got %d", nr.ServiceCount)
	}
	if report.CrossTeamNeedCount != 1 {
		t.Errorf("CrossTeamNeedCount: want 1, got %d", report.CrossTeamNeedCount)
	}
}

func TestValueChainAnalyzer_AtRisk(t *testing.T) {
	// Need backed by 3 teams → at_risk=true
	cfg := valueChainTestCfg{
		actors: []string{"operator"},
		teams: []teamCfg{
			{name: "team-a", size: 5},
			{name: "team-b", size: 5},
			{name: "team-c", size: 5},
		},
		services: []serviceCfg{
			{name: "svc-1", owner: "team-a"},
			{name: "svc-2", owner: "team-b"},
			{name: "svc-3", owner: "team-c"},
		},
		capabilities: []capCfg{
			{name: "cap-1", realizedBy: []string{"svc-1"}},
			{name: "cap-2", realizedBy: []string{"svc-2"}},
			{name: "cap-3", realizedBy: []string{"svc-3"}},
		},
		needs: []needCfg{
			{name: "monitor-fleet", actor: "operator", supportedBy: []string{"cap-1", "cap-2", "cap-3"}},
		},
	}

	m := buildValueChainModel(t, cfg)
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	nr := report.NeedRisks[0]
	if nr.TeamSpan != 3 {
		t.Errorf("TeamSpan: want 3, got %d", nr.TeamSpan)
	}
	if !nr.CrossTeam {
		t.Error("CrossTeam: want true")
	}
	if !nr.AtRisk {
		t.Error("AtRisk: want true for 3 teams (threshold is >2)")
	}
	if report.AtRiskNeedCount != 1 {
		t.Errorf("AtRiskNeedCount: want 1, got %d", report.AtRiskNeedCount)
	}
}

func TestValueChainAnalyzer_Unbacked(t *testing.T) {
	// Need with no capabilities → unbacked=true, team_span=0
	cfg := valueChainTestCfg{
		actors: []string{"merchant"},
		teams:  []teamCfg{{name: "team-a", size: 5}},
		needs: []needCfg{
			{name: "orphan-need", actor: "merchant", supportedBy: []string{}},
		},
	}

	m := buildValueChainModel(t, cfg)
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	nr := report.NeedRisks[0]
	if !nr.Unbacked {
		t.Error("Unbacked: want true, got false")
	}
	if nr.TeamSpan != 0 {
		t.Errorf("TeamSpan: want 0, got %d", nr.TeamSpan)
	}
	if nr.CapabilityCount != 0 {
		t.Errorf("CapabilityCount: want 0, got %d", nr.CapabilityCount)
	}
	if nr.ServiceCount != 0 {
		t.Errorf("ServiceCount: want 0, got %d", nr.ServiceCount)
	}
	if report.UnbackedNeedCount != 1 {
		t.Errorf("UnbackedNeedCount: want 1, got %d", report.UnbackedNeedCount)
	}
}

func TestValueChainAnalyzer_OverloadedTeam(t *testing.T) {
	// Need with only 1 team but that team has high cognitive load → at_risk=true
	// We give the team 7 owned caps (domain spread ≥6 → high) to trigger high overall level
	cfg := valueChainTestCfg{
		actors: []string{"merchant"},
		teams:  []teamCfg{{name: "team-overloaded", size: 5, ownsCaps: 7}},
		services: []serviceCfg{
			{name: "svc-1", owner: "team-overloaded"},
		},
		capabilities: []capCfg{
			{name: "cap-1", realizedBy: []string{"svc-1"}},
		},
		needs: []needCfg{
			{name: "upload-menu", actor: "merchant", supportedBy: []string{"cap-1"}},
		},
	}

	m := buildValueChainModel(t, cfg)
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	nr := report.NeedRisks[0]
	if nr.TeamSpan != 1 {
		t.Errorf("TeamSpan: want 1, got %d", nr.TeamSpan)
	}
	if nr.CrossTeam {
		t.Error("CrossTeam: want false (only 1 team)")
	}
	if !nr.AtRisk {
		t.Error("AtRisk: want true (team has high cognitive load)")
	}
	if report.AtRiskNeedCount != 1 {
		t.Errorf("AtRiskNeedCount: want 1, got %d", report.AtRiskNeedCount)
	}
}

func TestValueChainAnalyzer_Counts(t *testing.T) {
	// Multiple needs: 1 single-team, 1 cross-team (3 teams), 1 unbacked
	cfg := valueChainTestCfg{
		actors: []string{"merchant", "eater"},
		teams: []teamCfg{
			{name: "team-a", size: 5},
			{name: "team-b", size: 5},
			{name: "team-c", size: 5},
		},
		services: []serviceCfg{
			{name: "svc-1", owner: "team-a"},
			{name: "svc-2", owner: "team-b"},
			{name: "svc-3", owner: "team-c"},
		},
		capabilities: []capCfg{
			{name: "cap-1", realizedBy: []string{"svc-1"}},
			{name: "cap-2", realizedBy: []string{"svc-2"}},
			{name: "cap-3", realizedBy: []string{"svc-3"}},
		},
		needs: []needCfg{
			{name: "simple-need", actor: "merchant", supportedBy: []string{"cap-1"}},
			{name: "complex-need", actor: "eater", supportedBy: []string{"cap-1", "cap-2", "cap-3"}},
			{name: "orphan-need", actor: "merchant", supportedBy: []string{}},
		},
	}

	m := buildValueChainModel(t, cfg)
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	if len(report.NeedRisks) != 3 {
		t.Fatalf("expected 3 need risks, got %d", len(report.NeedRisks))
	}
	if report.CrossTeamNeedCount != 1 {
		t.Errorf("CrossTeamNeedCount: want 1, got %d", report.CrossTeamNeedCount)
	}
	if report.AtRiskNeedCount != 1 {
		t.Errorf("AtRiskNeedCount: want 1, got %d", report.AtRiskNeedCount)
	}
	if report.UnbackedNeedCount != 1 {
		t.Errorf("UnbackedNeedCount: want 1, got %d", report.UnbackedNeedCount)
	}
}

func TestValueChainAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)

	if len(report.NeedRisks) != 0 {
		t.Errorf("expected empty NeedRisks, got %d", len(report.NeedRisks))
	}
	if report.CrossTeamNeedCount != 0 {
		t.Errorf("CrossTeamNeedCount: want 0, got %d", report.CrossTeamNeedCount)
	}
	if report.AtRiskNeedCount != 0 {
		t.Errorf("AtRiskNeedCount: want 0, got %d", report.AtRiskNeedCount)
	}
	if report.UnbackedNeedCount != 0 {
		t.Errorf("UnbackedNeedCount: want 0, got %d", report.UnbackedNeedCount)
	}
}

// TestValueChainAnalyzer_CustomThresholds verifies that a custom AtRiskTeamSpan changes behavior.
func TestValueChainAnalyzer_CustomThresholds(t *testing.T) {
	m := entity.NewUNMModel("custom", "")

	// Create a need backed by a capability realized by 2 services owned by 2 different teams.
	// With default AtRiskTeamSpan=3, teamSpan=2 is NOT at-risk.
	// With AtRiskTeamSpan=1, teamSpan=2 IS at-risk.
	actor, _ := entity.NewActor("actor", "actor", "")
	_ = m.AddActor(&actor)

	need, _ := entity.NewNeed("n1", "n1", "actor", "")
	capID, _ := valueobject.NewEntityID("cap1")
	need.AddSupportedBy(entity.NewRelationship(capID, "", ""))
	_ = m.AddNeed(need)

	cap1, _ := entity.NewCapability("cap1", "cap1", "")
	svcID1, _ := valueobject.NewEntityID("svc1")
	svcID2, _ := valueobject.NewEntityID("svc2")
	cap1.AddRealizedBy(entity.NewRelationship(svcID1, "", ""))
	cap1.AddRealizedBy(entity.NewRelationship(svcID2, "", ""))
	_ = m.AddCapability(cap1)

	svc1, _ := entity.NewService("svc1", "svc1", "", "team-a")
	svc2, _ := entity.NewService("svc2", "svc2", "", "team-b")
	_ = m.AddService(svc1)
	_ = m.AddService(svc2)

	teamA, _ := entity.NewTeam("team-a", "team-a", "", valueobject.StreamAligned)
	teamB, _ := entity.NewTeam("team-b", "team-b", "", valueobject.StreamAligned)
	_ = m.AddTeam(teamA)
	_ = m.AddTeam(teamB)

	// Default threshold: AtRiskTeamSpan=3, team span=2 → not at-risk
	a := NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	report := a.Analyze(m)
	if len(report.NeedRisks) != 1 {
		t.Fatalf("expected 1 need risk, got %d", len(report.NeedRisks))
	}
	if report.NeedRisks[0].AtRisk {
		t.Error("with default AtRiskTeamSpan=3, team span=2 should NOT be at-risk")
	}

	// Custom threshold: AtRiskTeamSpan=1, team span=2 → at-risk
	a2 := NewValueChainAnalyzer(entity.ValueChainConfig{AtRiskTeamSpan: 1})
	report2 := a2.Analyze(m)
	if !report2.NeedRisks[0].AtRisk {
		t.Error("with AtRiskTeamSpan=1, team span=2 should be at-risk")
	}
}
