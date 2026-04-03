package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// buildTestModel creates a minimal UNMModel for testing changeset application.
// It has:
//   - 2 teams: "team-alpha" (stream-aligned), "team-beta" (platform)
//   - 3 services: "svc-a" (owned by team-alpha), "svc-b" (owned by team-alpha), "svc-c" (owned by team-beta)
//   - 2 capabilities: "cap-one", "cap-two"
//   - team-alpha owns cap-one, team-beta owns cap-two
//   - 1 need: "need-one" supported by cap-one
//   - 1 interaction: team-alpha → team-beta (collaboration)
func buildTestModel(t *testing.T) *entity.UNMModel {
	t.Helper()
	m := entity.NewUNMModel("test-system", "test description")

	// Actors
	actor, err := entity.NewActor("actor-1", "merchant", "a merchant")
	require.NoError(t, err)
	require.NoError(t, m.AddActor(&actor))

	// Teams
	teamAlpha, err := entity.NewTeam("team-alpha", "team-alpha", "Alpha team", valueobject.StreamAligned)
	require.NoError(t, err)
	require.NoError(t, m.AddTeam(teamAlpha))

	teamBeta, err := entity.NewTeam("team-beta", "team-beta", "Beta team", valueobject.Platform)
	require.NoError(t, err)
	require.NoError(t, m.AddTeam(teamBeta))

	// Services
	svcA, err := entity.NewService("svc-a", "svc-a", "Service A", "team-alpha")
	require.NoError(t, err)
	require.NoError(t, m.AddService(svcA))

	svcB, err := entity.NewService("svc-b", "svc-b", "Service B", "team-alpha")
	require.NoError(t, err)
	require.NoError(t, m.AddService(svcB))

	svcC, err := entity.NewService("svc-c", "svc-c", "Service C", "team-beta")
	require.NoError(t, err)
	require.NoError(t, m.AddService(svcC))

	// Capabilities
	capOne, err := entity.NewCapability("cap-one", "cap-one", "Capability One")
	require.NoError(t, err)
	require.NoError(t, m.AddCapability(capOne))

	capTwo, err := entity.NewCapability("cap-two", "cap-two", "Capability Two")
	require.NoError(t, err)
	require.NoError(t, m.AddCapability(capTwo))

	// Team owns capabilities
	capOneID, _ := valueobject.NewEntityID("cap-one")
	teamAlpha.AddOwns(entity.NewRelationship(capOneID, "", ""))

	capTwoID, _ := valueobject.NewEntityID("cap-two")
	teamBeta.AddOwns(entity.NewRelationship(capTwoID, "", ""))

	// Need
	need, err := entity.NewNeed("need-one", "need-one", "merchant", "outcome")
	require.NoError(t, err)
	need.AddSupportedBy(entity.NewRelationship(capOneID, "", ""))
	require.NoError(t, m.AddNeed(need))

	// Interaction
	interaction, err := entity.NewInteraction("ix-1", "team-alpha", "team-beta", valueobject.Collaboration, "", "collab")
	require.NoError(t, err)
	m.AddInteraction(interaction)

	return m
}

func TestMoveService_HappyPath(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMoveService,
		ServiceName: "svc-a",
		FromTeamName: "team-alpha",
		ToTeamName:   "team-beta",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	assert.Equal(t, "team-beta", result.Services["svc-a"].OwnerTeamName)
}

func TestMoveService_ServiceNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMoveService,
		ServiceName: "nonexistent",
		FromTeamName: "team-alpha",
		ToTeamName:   "team-beta",
	}))

	_, err = applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestMoveService_TargetTeamNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMoveService,
		ServiceName: "svc-a",
		FromTeamName: "team-alpha",
		ToTeamName:   "no-such-team",
	}))

	_, err = applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no-such-team")
}

func TestSplitTeam(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:             entity.ActionSplitTeam,
		OriginalTeamName: "team-alpha",
		NewTeamAName:     "team-alpha-1",
		NewTeamBName:     "team-alpha-2",
		ServiceAssignment: map[string]string{
			"svc-a": "a",
			"svc-b": "b",
		},
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Original team removed
	_, exists := result.Teams["team-alpha"]
	assert.False(t, exists, "original team should be removed")

	// New teams created
	_, existsA := result.Teams["team-alpha-1"]
	assert.True(t, existsA, "team-alpha-1 should exist")
	_, existsB := result.Teams["team-alpha-2"]
	assert.True(t, existsB, "team-alpha-2 should exist")

	// Services reassigned
	assert.Equal(t, "team-alpha-1", result.Services["svc-a"].OwnerTeamName)
	assert.Equal(t, "team-alpha-2", result.Services["svc-b"].OwnerTeamName)

	// TeamType inherited
	assert.Equal(t, valueobject.StreamAligned, result.Teams["team-alpha-1"].TeamType)
	assert.Equal(t, valueobject.StreamAligned, result.Teams["team-alpha-2"].TeamType)
}

func TestSplitTeam_UnassignedServicesGoToTeamA(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:             entity.ActionSplitTeam,
		OriginalTeamName: "team-alpha",
		NewTeamAName:     "team-alpha-1",
		NewTeamBName:     "team-alpha-2",
		ServiceAssignment: map[string]string{
			"svc-b": "b",
		},
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// svc-a was not explicitly assigned, so goes to TeamA
	assert.Equal(t, "team-alpha-1", result.Services["svc-a"].OwnerTeamName)
	assert.Equal(t, "team-alpha-2", result.Services["svc-b"].OwnerTeamName)
}

func TestSplitTeam_OriginalNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:             entity.ActionSplitTeam,
		OriginalTeamName: "nonexistent",
		NewTeamAName:     "a",
		NewTeamBName:     "b",
	}))

	_, err = applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestMergeTeams(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMergeTeams,
		TeamAName:   "team-alpha",
		TeamBName:   "team-beta",
		NewTeamName: "team-merged",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Both originals removed
	_, existsA := result.Teams["team-alpha"]
	assert.False(t, existsA)
	_, existsB := result.Teams["team-beta"]
	assert.False(t, existsB)

	// New team created
	merged, exists := result.Teams["team-merged"]
	require.True(t, exists)
	assert.Equal(t, valueobject.StreamAligned, merged.TeamType)

	// All services moved to new team
	assert.Equal(t, "team-merged", result.Services["svc-a"].OwnerTeamName)
	assert.Equal(t, "team-merged", result.Services["svc-b"].OwnerTeamName)
	assert.Equal(t, "team-merged", result.Services["svc-c"].OwnerTeamName)

	// Owns combined
	assert.Len(t, merged.Owns, 2) // cap-one from alpha + cap-two from beta
}

func TestMergeTeams_TeamNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMergeTeams,
		TeamAName:   "team-alpha",
		TeamBName:   "nonexistent",
		NewTeamName: "team-merged",
	}))

	_, err = applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestAddCapability(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionAddCapability,
		CapabilityName: "cap-new",
		OwnerTeamName:  "team-alpha",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, exists := result.Capabilities["cap-new"]
	assert.True(t, exists)

	// Check team-alpha owns it
	found := false
	for _, rel := range result.Teams["team-alpha"].Owns {
		if rel.TargetID.String() == "cap-new" {
			found = true
			break
		}
	}
	assert.True(t, found, "team-alpha should own cap-new")
}

func TestAddCapability_NoOwner(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionAddCapability,
		CapabilityName: "cap-orphan",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, exists := result.Capabilities["cap-orphan"]
	assert.True(t, exists)
}

func TestRemoveCapability(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionRemoveCapability,
		CapabilityName: "cap-one",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Capability removed from model
	_, exists := result.Capabilities["cap-one"]
	assert.False(t, exists)

	// Removed from team-alpha's Owns
	for _, rel := range result.Teams["team-alpha"].Owns {
		assert.NotEqual(t, "cap-one", rel.TargetID.String(), "cap-one should be removed from team-alpha Owns")
	}

	// Removed from need's SupportedBy
	for _, rel := range result.Needs["need-one"].SupportedBy {
		assert.NotEqual(t, "cap-one", rel.TargetID.String(), "cap-one should be removed from need-one SupportedBy")
	}
}

func TestReassignCapability(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionReassignCapability,
		CapabilityName: "cap-one",
		FromTeamName:   "team-alpha",
		ToTeamName:     "team-beta",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// cap-one removed from team-alpha Owns
	for _, rel := range result.Teams["team-alpha"].Owns {
		assert.NotEqual(t, "cap-one", rel.TargetID.String())
	}

	// cap-one added to team-beta Owns
	found := false
	for _, rel := range result.Teams["team-beta"].Owns {
		if rel.TargetID.String() == "cap-one" {
			found = true
			break
		}
	}
	assert.True(t, found, "team-beta should own cap-one after reassignment")
}

func TestAddInteraction(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:            entity.ActionAddInteraction,
		SourceTeamName:  "team-beta",
		TargetTeamName:  "team-alpha",
		InteractionMode: "x-as-a-service",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Should have 2 interactions now (original + new)
	assert.Len(t, result.Interactions, 2)

	newIx := result.Interactions[1]
	assert.Equal(t, "team-beta", newIx.FromTeamName)
	assert.Equal(t, "team-alpha", newIx.ToTeamName)
	assert.Equal(t, valueobject.XAsAService, newIx.Mode)
}

func TestAddInteraction_TeamNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:            entity.ActionAddInteraction,
		SourceTeamName:  "nonexistent",
		TargetTeamName:  "team-alpha",
		InteractionMode: "collaboration",
	}))

	_, err = applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestRemoveInteraction(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionRemoveInteraction,
		SourceTeamName: "team-alpha",
		TargetTeamName: "team-beta",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	assert.Len(t, result.Interactions, 0)
}

func TestRemoveInteraction_NotFound_NoError(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionRemoveInteraction,
		SourceTeamName: "team-alpha",
		TargetTeamName: "team-alpha", // no such interaction
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)
	assert.Len(t, result.Interactions, 1) // original still there
}

func TestUpdateTeamSize(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionUpdateTeamSize,
		TeamName: "team-alpha",
		NewSize:  12,
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	assert.Equal(t, 12, result.Teams["team-alpha"].Size)
	assert.True(t, result.Teams["team-alpha"].SizeExplicit)
}

func TestUpdateTeamSize_TeamNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionUpdateTeamSize,
		TeamName: "nonexistent",
		NewSize:  10,
	}))

	_, err = applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestOriginalModelUntouched(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	originalOwner := m.Services["svc-a"].OwnerTeamName

	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMoveService,
		ServiceName: "svc-a",
		FromTeamName: "team-alpha",
		ToTeamName:   "team-beta",
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Result should be changed
	assert.Equal(t, "team-beta", result.Services["svc-a"].OwnerTeamName)

	// Original should be untouched
	assert.Equal(t, originalOwner, m.Services["svc-a"].OwnerTeamName)
}

func TestCompositeChangeset(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "composite test")
	require.NoError(t, err)

	// Action 1: Move svc-a from team-alpha to team-beta
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionMoveService,
		ServiceName: "svc-a",
		FromTeamName: "team-alpha",
		ToTeamName:   "team-beta",
	}))

	// Action 2: Add a new capability
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionAddCapability,
		CapabilityName: "cap-three",
		OwnerTeamName:  "team-beta",
	}))

	// Action 3: Update team-beta size
	require.NoError(t, cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionUpdateTeamSize,
		TeamName: "team-beta",
		NewSize:  20,
	}))

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Verify all 3 actions applied cumulatively
	assert.Equal(t, "team-beta", result.Services["svc-a"].OwnerTeamName)
	_, capExists := result.Capabilities["cap-three"]
	assert.True(t, capExists)
	assert.Equal(t, 20, result.Teams["team-beta"].Size)
}

// ---- New action types tests ----

func TestAddService(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:          entity.ActionAddService,
		ServiceName:   "svc-new",
		OwnerTeamName: "team-alpha",
		Description:   "A new service",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	svc, exists := result.Services["svc-new"]
	require.True(t, exists)
	assert.Equal(t, "svc-new", svc.Name)
	assert.Equal(t, "team-alpha", svc.OwnerTeamName)
	assert.Equal(t, "A new service", svc.Description)
}

func TestAddService_AlreadyExists(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:          entity.ActionAddService,
		ServiceName:   "svc-a",
		OwnerTeamName: "team-alpha",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAddService_TeamNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:          entity.ActionAddService,
		ServiceName:   "svc-new",
		OwnerTeamName: "nonexistent",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRemoveService(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	// Link svc-a to cap-one via service.Realizes so we can verify cleanup
	capOneID, _ := valueobject.NewEntityID("cap-one")
	m.Services["svc-a"].AddRealizes(entity.NewRelationship(capOneID, "", ""))

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionRemoveService,
		ServiceName: "svc-a",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, exists := result.Services["svc-a"]
	assert.False(t, exists)

	// The service was deleted — so no service realizes cap-one with name svc-a
	svcs := result.GetServicesForCapability("cap-one")
	for _, svc := range svcs {
		assert.NotEqual(t, "svc-a", svc.Name)
	}
}

func TestRemoveService_NotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionRemoveService,
		ServiceName: "nonexistent",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestRenameService(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	capOneID, _ := valueobject.NewEntityID("cap-one")
	m.Services["svc-a"].AddRealizes(entity.NewRelationship(capOneID, "", ""))

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionRenameService,
		ServiceName:    "svc-a",
		NewServiceName: "svc-alpha",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, oldExists := result.Services["svc-a"]
	assert.False(t, oldExists)

	svc, newExists := result.Services["svc-alpha"]
	require.True(t, newExists)
	assert.Equal(t, "svc-alpha", svc.Name)
	assert.Equal(t, "team-alpha", svc.OwnerTeamName)

	// The renamed service should still realize cap-one
	svcs := result.GetServicesForCapability("cap-one")
	found := false
	for _, s := range svcs {
		if s.Name == "svc-alpha" {
			found = true
		}
		assert.NotEqual(t, "svc-a", s.Name)
	}
	assert.True(t, found, "svc-alpha should still realize cap-one")
}

func TestAddTeam(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionAddTeam,
		TeamName:    "team-gamma",
		TeamType:    "enabling",
		Description: "Gamma team",
		NewSize:     8,
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	team, exists := result.Teams["team-gamma"]
	require.True(t, exists)
	assert.Equal(t, "team-gamma", team.Name)
	assert.Equal(t, valueobject.Enabling, team.TeamType)
	assert.Equal(t, 8, team.Size)
	assert.Equal(t, "Gamma team", team.Description)
}

func TestAddTeam_AlreadyExists(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionAddTeam,
		TeamName: "team-alpha",
		TeamType: "stream-aligned",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAddTeam_InvalidType(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionAddTeam,
		TeamName: "team-gamma",
		TeamType: "invalid-type",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestRemoveTeam(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	// First move all services away from team-beta
	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:         entity.ActionMoveService,
		ServiceName:  "svc-c",
		FromTeamName: "team-beta",
		ToTeamName:   "team-alpha",
	})
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionRemoveTeam,
		TeamName: "team-beta",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, exists := result.Teams["team-beta"]
	assert.False(t, exists)
	// Interactions involving team-beta removed
	for _, ix := range result.Interactions {
		assert.NotEqual(t, "team-beta", ix.FromTeamName)
		assert.NotEqual(t, "team-beta", ix.ToTeamName)
	}
}

func TestRemoveTeam_StillOwnsServices(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionRemoveTeam,
		TeamName: "team-alpha",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "still owns service")
}

func TestUpdateTeamType(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionUpdateTeamType,
		TeamName: "team-alpha",
		TeamType: "enabling",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	assert.Equal(t, valueobject.Enabling, result.Teams["team-alpha"].TeamType)
}

func TestUpdateTeamType_InvalidType(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionUpdateTeamType,
		TeamName: "team-alpha",
		TeamType: "bogus",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestAddNeed(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionAddNeed,
		NeedName:    "need-two",
		ActorName:   "merchant",
		Outcome:     "better outcome",
		SupportedBy: []string{"cap-one", "cap-two"},
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	need, exists := result.Needs["need-two"]
	require.True(t, exists)
	assert.Equal(t, "need-two", need.Name)
	assert.Equal(t, []string{"merchant"}, need.ActorNames)
	assert.Equal(t, "better outcome", need.Outcome)
	assert.Len(t, need.SupportedBy, 2)
}

func TestAddNeed_ActorNotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:      entity.ActionAddNeed,
		NeedName:  "need-new",
		ActorName: "nonexistent",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRemoveNeed(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionRemoveNeed,
		NeedName: "need-one",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, exists := result.Needs["need-one"]
	assert.False(t, exists)
}

func TestAddActor(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionAddActor,
		ActorName:   "driver",
		Description: "A driver actor",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	actor, exists := result.Actors["driver"]
	require.True(t, exists)
	assert.Equal(t, "driver", actor.Name)
	assert.Equal(t, "A driver actor", actor.Description)
}

func TestAddActor_AlreadyExists(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:      entity.ActionAddActor,
		ActorName: "merchant",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRemoveActor(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	// First remove the need that references the actor
	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:     entity.ActionRemoveNeed,
		NeedName: "need-one",
	})
	_ = cs.AddAction(entity.ChangeAction{
		Type:      entity.ActionRemoveActor,
		ActorName: "merchant",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	_, exists := result.Actors["merchant"]
	assert.False(t, exists)
}

func TestRemoveActor_StillReferenced(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:      entity.ActionRemoveActor,
		ActorName: "merchant",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "still referenced")
}

func TestAddServiceDependency(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:             entity.ActionAddServiceDependency,
		ServiceName:      "svc-a",
		DependsOnService: "svc-c",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	found := false
	for _, rel := range result.Services["svc-a"].DependsOn {
		if rel.TargetID.String() == "svc-c" {
			found = true
		}
	}
	assert.True(t, found, "svc-a should depend on svc-c")
}

func TestRemoveServiceDependency(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	// First add a dependency, then remove it
	depID, _ := valueobject.NewEntityID("svc-c")
	m.Services["svc-a"].DependsOn = append(m.Services["svc-a"].DependsOn,
		entity.NewRelationship(depID, "", ""))

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:             entity.ActionRemoveServiceDependency,
		ServiceName:      "svc-a",
		DependsOnService: "svc-c",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	for _, rel := range result.Services["svc-a"].DependsOn {
		assert.NotEqual(t, "svc-c", rel.TargetID.String())
	}
}

func TestLinkNeedCapability(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionLinkNeedCapability,
		NeedName:       "need-one",
		CapabilityName: "cap-two",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	found := false
	for _, rel := range result.Needs["need-one"].SupportedBy {
		if rel.TargetID.String() == "cap-two" {
			found = true
		}
	}
	assert.True(t, found, "need-one should be supported by cap-two")
}

func TestUnlinkNeedCapability(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionUnlinkNeedCapability,
		NeedName:       "need-one",
		CapabilityName: "cap-one",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	for _, rel := range result.Needs["need-one"].SupportedBy {
		assert.NotEqual(t, "cap-one", rel.TargetID.String())
	}
}

func TestLinkCapabilityService(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionLinkCapabilityService,
		CapabilityName: "cap-one",
		ServiceName:    "svc-a",
		Role:           "primary",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// After LinkCapabilityService, svc-a should realize cap-one
	svcs := result.GetServicesForCapability("cap-one")
	found := false
	for _, svc := range svcs {
		if svc.Name == "svc-a" {
			found = true
			// Check that the role is set on the service's Realizes relationship
			for _, rel := range result.Services["svc-a"].Realizes {
				if rel.TargetID.String() == "cap-one" {
					assert.Equal(t, valueobject.RelationshipRole("primary"), rel.Role)
				}
			}
		}
	}
	assert.True(t, found, "cap-one should be realized by svc-a")
}

func TestUnlinkCapabilityService(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	// First link it via service.Realizes
	capOneID, _ := valueobject.NewEntityID("cap-one")
	m.Services["svc-a"].AddRealizes(entity.NewRelationship(capOneID, "", ""))

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionUnlinkCapabilityService,
		CapabilityName: "cap-one",
		ServiceName:    "svc-a",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// After unlink, svc-a should not realize cap-one
	svcs := result.GetServicesForCapability("cap-one")
	for _, svc := range svcs {
		assert.NotEqual(t, "svc-a", svc.Name)
	}
}

func TestUpdateCapabilityVisibility(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionUpdateCapabilityVisibility,
		CapabilityName: "cap-one",
		Visibility:     "user-facing",
	})

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	assert.Equal(t, "user-facing", result.Capabilities["cap-one"].Visibility)
}

func TestUpdateCapabilityVisibility_Invalid(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:           entity.ActionUpdateCapabilityVisibility,
		CapabilityName: "cap-one",
		Visibility:     "bogus",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestUpdateDescription(t *testing.T) {
	applier := NewChangesetApplier()

	tests := []struct {
		entityType string
		entityName string
	}{
		{"actor", "merchant"},
		{"service", "svc-a"},
		{"team", "team-alpha"},
		{"capability", "cap-one"},
		{"need", "need-one"},
	}

	for _, tc := range tests {
		t.Run(tc.entityType, func(t *testing.T) {
			model := buildTestModel(t)
			cs, _ := entity.NewChangeset("cs-1", "test")
			_ = cs.AddAction(entity.ChangeAction{
				Type:        entity.ActionUpdateDescription,
				EntityType:  tc.entityType,
				EntityName:  tc.entityName,
				Description: "updated description",
			})

			result, err := applier.Apply(model, cs)
			require.NoError(t, err)

			switch tc.entityType {
			case "actor":
				assert.Equal(t, "updated description", result.Actors[tc.entityName].Description)
			case "service":
				assert.Equal(t, "updated description", result.Services[tc.entityName].Description)
			case "team":
				assert.Equal(t, "updated description", result.Teams[tc.entityName].Description)
			case "capability":
				assert.Equal(t, "updated description", result.Capabilities[tc.entityName].Description)
			case "need":
				assert.Equal(t, "updated description", result.Needs[tc.entityName].Outcome)
			}
		})
	}
}

func TestUpdateDescription_NotFound(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionUpdateDescription,
		EntityType:  "service",
		EntityName:  "nonexistent",
		Description: "x",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
}

func TestUpdateDescription_InvalidEntityType(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, _ := entity.NewChangeset("cs-1", "test")
	_ = cs.AddAction(entity.ChangeAction{
		Type:        entity.ActionUpdateDescription,
		EntityType:  "bogus",
		EntityName:  "x",
		Description: "x",
	})

	_, err := applier.Apply(m, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported entity_type")
}

func TestEmptyChangeset(t *testing.T) {
	m := buildTestModel(t)
	applier := NewChangesetApplier()

	cs, err := entity.NewChangeset("cs-1", "empty")
	require.NoError(t, err)

	result, err := applier.Apply(m, cs)
	require.NoError(t, err)

	// Should be a valid deep copy with same data
	assert.Equal(t, len(m.Teams), len(result.Teams))
	assert.Equal(t, len(m.Services), len(result.Services))
}
