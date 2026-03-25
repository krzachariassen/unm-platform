package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// mustChangeset creates a Changeset, fataling on error.
func mustChangeset(t *testing.T, id, description string) *entity.Changeset {
	t.Helper()
	cs, err := entity.NewChangeset(id, description)
	if err != nil {
		t.Fatalf("NewChangeset %q: %v", id, err)
	}
	return cs
}

func mustEntityID(t *testing.T, id string) valueobject.EntityID {
	t.Helper()
	eid, err := valueobject.NewEntityID(id)
	if err != nil {
		t.Fatalf("NewEntityID %q: %v", id, err)
	}
	return eid
}

func TestImpactAnalyzer_EmptyChangeset(t *testing.T) {
	m := entity.NewUNMModel("test", "test system")
	cs := mustChangeset(t, "cs-1", "empty changeset")

	analyzer := NewImpactAnalyzer(entity.DefaultConfig().Analysis)
	report, err := analyzer.Analyze(m, cs)
	require.NoError(t, err)
	assert.Equal(t, "cs-1", report.ChangesetID)
	// All dimensions should be unchanged
	for _, d := range report.Deltas {
		assert.Equal(t, Unchanged, d.Change, "dimension %s should be unchanged", d.Dimension)
	}
}

func TestImpactAnalyzer_FragmentationImproved(t *testing.T) {
	// Setup: a capability realized by services in 2 different teams → fragmented.
	// Changeset: move one service to the other team → no longer fragmented.
	m := entity.NewUNMModel("test", "test system")
	teamA := mustTeam(t, "team-a")
	teamB := mustTeam(t, "team-b")
	addTeamToModel(t, m, teamA)
	addTeamToModel(t, m, teamB)

	svcA := mustService(t, "svc-a", "team-a")
	svcB := mustService(t, "svc-b", "team-b")
	addServiceToModel(t, m, svcA)
	addServiceToModel(t, m, svcB)

	cap := mustCap(t, "cap-1")
	cap.RealizedBy = []entity.Relationship{
		entity.NewRelationship(mustEntityID(t, "svc-a"), "", ""),
		entity.NewRelationship(mustEntityID(t, "svc-b"), "", ""),
	}
	require.NoError(t, m.AddCapability(cap))

	// Changeset: move svc-b to team-a → both services now in same team
	cs := mustChangeset(t, "cs-frag", "consolidate services")
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:         entity.ActionMoveService,
		ServiceName:  "svc-b",
		FromTeamName: "team-b",
		ToTeamName:   "team-a",
	}))

	analyzer := NewImpactAnalyzer(entity.DefaultConfig().Analysis)
	report, err := analyzer.Analyze(m, cs)
	require.NoError(t, err)

	fragDelta := findDelta(report, "fragmentation")
	require.NotNil(t, fragDelta, "should have fragmentation delta")
	assert.Equal(t, Improved, fragDelta.Change)
}

func TestImpactAnalyzer_CognitiveLoadRegressed(t *testing.T) {
	// Setup: team-a owns 3 capabilities (low domain spread).
	// Changeset: add 4 more capabilities to team-a → 7 total (high domain spread).
	m := entity.NewUNMModel("test", "test system")
	teamA := mustTeam(t, "team-a")
	addTeamToModel(t, m, teamA)

	svc := mustService(t, "svc-a", "team-a")
	addServiceToModel(t, m, svc)

	// Give team-a 3 capabilities
	for i := 1; i <= 3; i++ {
		capName := "cap-" + string(rune('a'+i-1))
		cap := mustCap(t, capName)
		require.NoError(t, m.AddCapability(cap))
		teamA.Owns = append(teamA.Owns, entity.NewRelationship(mustEntityID(t, capName), "", ""))
	}

	// Changeset: add 4 more capabilities owned by team-a
	cs := mustChangeset(t, "cs-cog", "add capabilities")
	for i := 4; i <= 7; i++ {
		capName := "cap-new-" + string(rune('a'+i-4))
		require.NoError(t, cs.AddAction(entity.ChangeAction{
			Type:           entity.ActionAddCapability,
			CapabilityName: capName,
			OwnerTeamName:  "team-a",
		}))
	}

	analyzer := NewImpactAnalyzer(entity.DefaultConfig().Analysis)
	report, err := analyzer.Analyze(m, cs)
	require.NoError(t, err)

	cogDelta := findDelta(report, "cognitive_load")
	require.NotNil(t, cogDelta, "should have cognitive_load delta")
	assert.Equal(t, Regressed, cogDelta.Change)
}

func TestImpactAnalyzer_UnchangedDimensions(t *testing.T) {
	// Setup: minimal model. Changeset that only affects fragmentation.
	// Other dimensions should remain unchanged.
	m := entity.NewUNMModel("test", "test system")
	teamA := mustTeam(t, "team-a")
	teamB := mustTeam(t, "team-b")
	addTeamToModel(t, m, teamA)
	addTeamToModel(t, m, teamB)

	svcA := mustService(t, "svc-a", "team-a")
	svcB := mustService(t, "svc-b", "team-b")
	addServiceToModel(t, m, svcA)
	addServiceToModel(t, m, svcB)

	cap := mustCap(t, "cap-1")
	cap.RealizedBy = []entity.Relationship{
		entity.NewRelationship(mustEntityID(t, "svc-a"), "", ""),
		entity.NewRelationship(mustEntityID(t, "svc-b"), "", ""),
	}
	require.NoError(t, m.AddCapability(cap))

	// Move service to consolidate → affects fragmentation only
	cs := mustChangeset(t, "cs-unch", "consolidate")
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:         entity.ActionMoveService,
		ServiceName:  "svc-b",
		FromTeamName: "team-b",
		ToTeamName:   "team-a",
	}))

	analyzer := NewImpactAnalyzer(entity.DefaultConfig().Analysis)
	report, err := analyzer.Analyze(m, cs)
	require.NoError(t, err)

	// Bottleneck, coupling, value_chain, value_stream should be unchanged
	for _, dim := range []string{"bottleneck", "coupling", "value_chain", "value_stream"} {
		d := findDelta(report, dim)
		require.NotNil(t, d, "should have %s delta", dim)
		assert.Equal(t, Unchanged, d.Change, "dimension %s should be unchanged", dim)
	}
}

func TestImpactAnalyzer_AllDimensionsPresent(t *testing.T) {
	m := entity.NewUNMModel("test", "test system")
	cs := mustChangeset(t, "cs-all", "empty")

	analyzer := NewImpactAnalyzer(entity.DefaultConfig().Analysis)
	report, err := analyzer.Analyze(m, cs)
	require.NoError(t, err)

	expected := []string{"fragmentation", "cognitive_load", "bottleneck", "coupling", "value_chain", "value_stream"}
	var actual []string
	for _, d := range report.Deltas {
		actual = append(actual, d.Dimension)
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestImpactAnalyzer_ApplierError(t *testing.T) {
	m := entity.NewUNMModel("test", "test system")
	cs := mustChangeset(t, "cs-err", "bad changeset")
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:         entity.ActionMoveService,
		ServiceName:  "nonexistent",
		FromTeamName: "no-team",
		ToTeamName:   "no-team2",
	}))

	analyzer := NewImpactAnalyzer(entity.DefaultConfig().Analysis)
	_, err := analyzer.Analyze(m, cs)
	assert.Error(t, err)
}

// findDelta returns the DimensionDelta with the given dimension name, or nil if not found.
func findDelta(report ImpactReport, dimension string) *DimensionDelta {
	for i := range report.Deltas {
		if report.Deltas[i].Dimension == dimension {
			return &report.Deltas[i]
		}
	}
	return nil
}

// TestImpactAnalyzer_UsesProvidedConfig verifies that NewImpactAnalyzer accepts a custom
// AnalysisConfig and uses it for cognitive load thresholds instead of hardcoded defaults.
// With a very high DomainSpreadThreshold, a team that would be "high load" under the
// default config should appear as "unchanged" (i.e. not high load).
func TestImpactAnalyzer_UsesProvidedConfig(t *testing.T) {
	// Build a model where team-a owns many capabilities (would be high load under defaults).
	m := entity.NewUNMModel("test", "custom-cfg")
	teamA := mustTeam(t, "team-a")
	addTeamToModel(t, m, teamA)
	svc := mustService(t, "svc-a", "team-a")
	addServiceToModel(t, m, svc)
	for i := 1; i <= 8; i++ {
		capName := "cap-" + string(rune('a'+i-1))
		cap := mustCap(t, capName)
		require.NoError(t, m.AddCapability(cap))
		teamA.Owns = append(teamA.Owns, entity.NewRelationship(mustEntityID(t, capName), "", ""))
	}

	// Use a custom config with very permissive thresholds so team-a is NOT high-load.
	customCfg := entity.DefaultConfig().Analysis
	customCfg.CognitiveLoad.DomainSpreadThresholds = [2]int{100, 200} // extremely high

	// Also apply a minimal changeset so there is a before/after to diff.
	cs := mustChangeset(t, "cs-cfg", "empty")

	ia := NewImpactAnalyzer(customCfg)
	report, err := ia.Analyze(m, cs)
	require.NoError(t, err)

	cogDelta := findDelta(report, "cognitive_load")
	require.NotNil(t, cogDelta, "should have cognitive_load delta")
	// With permissive thresholds, team-a is NOT high load → counts are 0 before and after → unchanged.
	assert.Equal(t, Unchanged, cogDelta.Change,
		"cognitive_load should be unchanged when custom thresholds are permissive; got %s (%s → %s)",
		cogDelta.Change, cogDelta.Before, cogDelta.After)
}
