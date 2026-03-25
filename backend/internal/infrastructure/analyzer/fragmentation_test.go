package analyzer

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// buildFragmentationModel creates a model where:
//   - cap-A is owned by 3 teams (alpha, beta, gamma) → fragmented
//   - cap-B is owned by 2 teams (alpha, beta) → NOT fragmented (threshold is > 2)
func buildFragmentationModel(t *testing.T) *entity.UNMModel {
	t.Helper()

	m := entity.NewUNMModel("test-system", "")

	capA, err := entity.NewCapability("cap-a", "cap-A", "")
	if err != nil {
		t.Fatalf("NewCapability cap-A: %v", err)
	}
	capB, err := entity.NewCapability("cap-b", "cap-B", "")
	if err != nil {
		t.Fatalf("NewCapability cap-B: %v", err)
	}
	if err := m.AddCapability(capA); err != nil {
		t.Fatalf("AddCapability cap-A: %v", err)
	}
	if err := m.AddCapability(capB); err != nil {
		t.Fatalf("AddCapability cap-B: %v", err)
	}

	teamNames := []string{"alpha", "beta", "gamma"}
	for _, name := range teamNames {
		team, err := entity.NewTeam(name, name, "", valueobject.StreamAligned)
		if err != nil {
			t.Fatalf("NewTeam %s: %v", name, err)
		}
		// All three teams own cap-A.
		idA, err := valueobject.NewEntityID("cap-A")
		if err != nil {
			t.Fatalf("NewEntityID cap-A: %v", err)
		}
		team.AddOwns(entity.NewRelationship(idA, "", ""))

		// Only alpha and beta own cap-B.
		if name == "alpha" || name == "beta" {
			idB, err := valueobject.NewEntityID("cap-B")
			if err != nil {
				t.Fatalf("NewEntityID cap-B: %v", err)
			}
			team.AddOwns(entity.NewRelationship(idB, "", ""))
		}
		if err := m.AddTeam(team); err != nil {
			t.Fatalf("AddTeam %s: %v", name, err)
		}
	}
	return m
}

func TestFragmentationAnalyzer_OnlyCapAIsFragmented(t *testing.T) {
	m := buildFragmentationModel(t)
	a := NewFragmentationAnalyzer()
	report := a.Analyze(m)

	if len(report.FragmentedCapabilities) != 1 {
		t.Fatalf("expected 1 fragmented capability, got %d", len(report.FragmentedCapabilities))
	}
	if report.FragmentedCapabilities[0].Capability.Name != "cap-A" {
		t.Errorf("expected fragmented capability to be cap-A, got %q", report.FragmentedCapabilities[0].Capability.Name)
	}
}

func TestFragmentationAnalyzer_CapBNotFragmented(t *testing.T) {
	m := buildFragmentationModel(t)
	a := NewFragmentationAnalyzer()
	report := a.Analyze(m)

	for _, fc := range report.FragmentedCapabilities {
		if fc.Capability.Name == "cap-B" {
			t.Errorf("cap-B should not be fragmented (only 2 owners), but it was reported")
		}
	}
}

func TestFragmentationAnalyzer_TeamsCount(t *testing.T) {
	m := buildFragmentationModel(t)
	a := NewFragmentationAnalyzer()
	report := a.Analyze(m)

	if len(report.FragmentedCapabilities) != 1 {
		t.Fatalf("expected 1 fragmented capability, got %d", len(report.FragmentedCapabilities))
	}
	fc := report.FragmentedCapabilities[0]
	if len(fc.Teams) != 3 {
		t.Errorf("expected 3 teams for cap-A, got %d", len(fc.Teams))
	}
}

func TestFragmentationAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewFragmentationAnalyzer()
	report := a.Analyze(m)

	if len(report.FragmentedCapabilities) != 0 {
		t.Errorf("expected empty report for model with no fragmentation, got %d entries", len(report.FragmentedCapabilities))
	}
}

func TestFragmentationAnalyzer_NoFragmentation(t *testing.T) {
	m := entity.NewUNMModel("no-frag", "")

	cap, err := entity.NewCapability("cap-1", "cap-One", "")
	if err != nil {
		t.Fatalf("NewCapability: %v", err)
	}
	if err := m.AddCapability(cap); err != nil {
		t.Fatalf("AddCapability: %v", err)
	}

	// Only 2 teams own cap-One → not fragmented.
	for _, name := range []string{"team-x", "team-y"} {
		team, err := entity.NewTeam(name, name, "", valueobject.StreamAligned)
		if err != nil {
			t.Fatalf("NewTeam: %v", err)
		}
		id, err := valueobject.NewEntityID("cap-One")
		if err != nil {
			t.Fatalf("NewEntityID: %v", err)
		}
		team.AddOwns(entity.NewRelationship(id, "", ""))
		if err := m.AddTeam(team); err != nil {
			t.Fatalf("AddTeam: %v", err)
		}
	}

	a := NewFragmentationAnalyzer()
	report := a.Analyze(m)

	if len(report.FragmentedCapabilities) != 0 {
		t.Errorf("expected no fragmented capabilities, got %d", len(report.FragmentedCapabilities))
	}
}
