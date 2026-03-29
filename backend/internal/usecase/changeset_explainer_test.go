package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func newChangesetExplainer() *usecase.ChangesetExplainer {
	cfg := entity.DefaultConfig().Analysis
	return usecase.NewChangesetExplainer(
		analyzer.NewImpactAnalyzer(cfg),
	)
}

func buildMinimalChangeset() *entity.Changeset {
	cs := &entity.Changeset{}
	return cs
}

func buildMinimalModelForChangeset() *entity.UNMModel {
	m := &entity.UNMModel{}
	m.System.Name = "ChangesetSystem"
	m.Teams = map[string]*entity.Team{}
	m.Services = map[string]*entity.Service{}
	m.Capabilities = map[string]*entity.Capability{}
	m.Actors = map[string]*entity.Actor{}
	m.Needs = map[string]*entity.Need{}
	return m
}

func TestChangesetExplainer_PrepareExplainData_ReturnsData(t *testing.T) {
	explainer := newChangesetExplainer()
	m := buildMinimalModelForChangeset()
	cs := buildMinimalChangeset()

	data, err := explainer.PrepareExplainData(m, cs, "cs-001")
	require.NoError(t, err)

	assert.Equal(t, "ChangesetSystem", data["SystemName"])
	assert.Equal(t, "cs-001", data["ChangesetID"])
	assert.Contains(t, data, "ChangesetActions")
	assert.Contains(t, data, "ImpactDeltas")
}

func TestChangesetExplainer_PrepareExplainData_ChangesetActionsText(t *testing.T) {
	explainer := newChangesetExplainer()
	// Build a model with the service and teams required for a move_service action.
	m := buildMinimalModelForChangeset()
	m.Teams = map[string]*entity.Team{
		"TeamA": {Name: "TeamA"},
		"TeamB": {Name: "TeamB"},
	}
	m.Services = map[string]*entity.Service{
		"svc-a": {Name: "svc-a", OwnerTeamName: "TeamA"},
	}
	cs := &entity.Changeset{
		Actions: []entity.ChangeAction{
			{Type: entity.ActionMoveService, ServiceName: "svc-a", FromTeamName: "TeamA", ToTeamName: "TeamB"},
		},
	}

	data, err := explainer.PrepareExplainData(m, cs, "cs-002")
	require.NoError(t, err)

	actionsText, ok := data["ChangesetActions"].(string)
	require.True(t, ok)
	assert.Contains(t, actionsText, "move_service")
}

func TestChangesetExplainer_PrepareExplainData_ImpactDeltasText(t *testing.T) {
	explainer := newChangesetExplainer()
	m := buildMinimalModelForChangeset()
	cs := buildMinimalChangeset()

	data, err := explainer.PrepareExplainData(m, cs, "cs-003")
	require.NoError(t, err)

	// ImpactDeltas should be a string (may be empty for empty model)
	_, ok := data["ImpactDeltas"].(string)
	assert.True(t, ok, "ImpactDeltas should be a string")
}
