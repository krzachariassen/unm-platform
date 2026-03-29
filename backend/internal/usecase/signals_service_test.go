package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func buildMinimalModelForSignals() *entity.UNMModel {
	m := &entity.UNMModel{}
	m.System.Name = "TestSystem"
	m.Teams = map[string]*entity.Team{
		"TeamA": {Name: "TeamA", TeamType: valueobject.StreamAligned},
		"TeamB": {Name: "TeamB", TeamType: valueobject.StreamAligned},
	}
	m.Services = map[string]*entity.Service{
		"svc-a": {Name: "svc-a", OwnerTeamName: "TeamA"},
		"svc-b": {Name: "svc-b", OwnerTeamName: "TeamB"},
	}
	m.Capabilities = map[string]*entity.Capability{
		"CapX": {Name: "CapX", Visibility: entity.CapVisibilityUserFacing},
	}
	m.Actors = map[string]*entity.Actor{
		"User": {Name: "User"},
	}
	m.Needs = map[string]*entity.Need{
		"NeedA": {Name: "NeedA", ActorName: "User"},
	}
	return m
}

func TestSignalsService_BuildSignalsData_ReturnsSignalsResponse(t *testing.T) {
	cfg := entity.DefaultConfig().Analysis
	cl := analyzer.NewCognitiveLoadAnalyzer(cfg.CognitiveLoad, cfg.InteractionWeights)
	svc := usecase.NewSignalsService(
		analyzer.NewValueChainAnalyzer(cfg.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		cl,
		analyzer.NewBottleneckAnalyzer(cfg.Bottleneck),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		cfg.Signals,
	)

	m := buildMinimalModelForSignals()
	resp, err := svc.BuildSignalsData(m)
	require.NoError(t, err)
	assert.Equal(t, "signals", resp.ViewType)
	assert.NotEmpty(t, resp.Health.UXRisk)
	assert.NotEmpty(t, resp.Health.ArchitectureRisk)
	assert.NotEmpty(t, resp.Health.OrgRisk)
}

func TestSignalsService_HealthLevels_AreLegalValues(t *testing.T) {
	cfg := entity.DefaultConfig().Analysis
	cl := analyzer.NewCognitiveLoadAnalyzer(cfg.CognitiveLoad, cfg.InteractionWeights)
	svc := usecase.NewSignalsService(
		analyzer.NewValueChainAnalyzer(cfg.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		cl,
		analyzer.NewBottleneckAnalyzer(cfg.Bottleneck),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		cfg.Signals,
	)

	m := buildMinimalModelForSignals()
	resp, err := svc.BuildSignalsData(m)
	require.NoError(t, err)

	validLevels := map[string]bool{"red": true, "amber": true, "green": true}
	assert.True(t, validLevels[resp.Health.UXRisk], "UXRisk must be red/amber/green, got: %s", resp.Health.UXRisk)
	assert.True(t, validLevels[resp.Health.ArchitectureRisk], "ArchRisk must be red/amber/green")
	assert.True(t, validLevels[resp.Health.OrgRisk], "OrgRisk must be red/amber/green")
}

func TestSignalsService_NilSlicesAreCoalescedToEmpty(t *testing.T) {
	cfg := entity.DefaultConfig().Analysis
	cl := analyzer.NewCognitiveLoadAnalyzer(cfg.CognitiveLoad, cfg.InteractionWeights)
	svc := usecase.NewSignalsService(
		analyzer.NewValueChainAnalyzer(cfg.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		cl,
		analyzer.NewBottleneckAnalyzer(cfg.Bottleneck),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		cfg.Signals,
	)

	// Empty model — all slices should be non-nil (for clean JSON)
	m := &entity.UNMModel{}
	m.System.Name = "Empty"
	m.Teams = map[string]*entity.Team{}
	m.Services = map[string]*entity.Service{}
	m.Capabilities = map[string]*entity.Capability{}
	m.Actors = map[string]*entity.Actor{}
	m.Needs = map[string]*entity.Need{}

	resp, err := svc.BuildSignalsData(m)
	require.NoError(t, err)
	assert.NotNil(t, resp.UserExperienceLayer.NeedsRequiring3PlusTeams)
	assert.NotNil(t, resp.UserExperienceLayer.NeedsWithNoCapBacking)
	assert.NotNil(t, resp.UserExperienceLayer.NeedsAtRisk)
	assert.NotNil(t, resp.ArchitectureLayer.UserFacingCapsWithCrossTeamServices)
	assert.NotNil(t, resp.ArchitectureLayer.CapabilitiesNotConnectedToAnyNeed)
	assert.NotNil(t, resp.ArchitectureLayer.CapabilitiesFragmentedAcrossTeams)
	assert.NotNil(t, resp.OrganizationLayer.TopTeamsByStructuralLoad)
	assert.NotNil(t, resp.OrganizationLayer.CriticalBottleneckServices)
	assert.NotNil(t, resp.OrganizationLayer.LowCoherenceTeams)
	assert.NotNil(t, resp.OrganizationLayer.CriticalExternalDeps)
}
