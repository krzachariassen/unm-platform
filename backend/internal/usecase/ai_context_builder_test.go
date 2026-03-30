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

func buildModelForAIContext() *entity.UNMModel {
	m := &entity.UNMModel{}
	m.System.Name = "AIContextSystem"
	m.System.Description = "Test system"
	m.Teams = map[string]*entity.Team{
		"TeamA": {Name: "TeamA", TeamType: valueobject.StreamAligned},
	}
	m.Services = map[string]*entity.Service{
		"svc-a": {Name: "svc-a", OwnerTeamName: "TeamA"},
	}
	m.Capabilities = map[string]*entity.Capability{
		"CapX": {Name: "CapX", Visibility: entity.CapVisibilityUserFacing},
	}
	m.Actors = map[string]*entity.Actor{
		"User": {Name: "User"},
	}
	m.Needs = map[string]*entity.Need{
		"NeedA": {Name: "NeedA", ActorNames: []string{"User"}},
	}
	m.Interactions = []*entity.Interaction{}
	m.Signals = []*entity.Signal{}
	m.ExternalDependencies = map[string]*entity.ExternalDependency{}
	return m
}

func newAIContextBuilder() *usecase.AIContextBuilder {
	cfg := entity.DefaultConfig().Analysis
	return usecase.NewAIContextBuilder(
		analyzer.NewCognitiveLoadAnalyzer(cfg.CognitiveLoad, cfg.InteractionWeights),
		analyzer.NewValueChainAnalyzer(cfg.ValueChain),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewValueStreamAnalyzer(),
	)
}

func TestAIContextBuilder_BuildPromptData_ContainsRequiredKeys(t *testing.T) {
	builder := newAIContextBuilder()
	m := buildModelForAIContext()

	data, err := builder.BuildPromptData(m, "What is the system health?")
	require.NoError(t, err)

	// Required keys for all prompt templates
	assert.Contains(t, data, "SystemName")
	assert.Contains(t, data, "Teams")
	assert.Contains(t, data, "Services")
	assert.Contains(t, data, "Capabilities")
	assert.Contains(t, data, "Needs")
	assert.Contains(t, data, "Interactions")
	assert.Contains(t, data, "UserQuestion")
	assert.Contains(t, data, "Question")
}

func TestAIContextBuilder_BuildPromptData_SystemName(t *testing.T) {
	builder := newAIContextBuilder()
	m := buildModelForAIContext()

	data, err := builder.BuildPromptData(m, "test question")
	require.NoError(t, err)

	assert.Equal(t, "AIContextSystem", data["SystemName"])
}

func TestAIContextBuilder_BuildPromptData_UserQuestion(t *testing.T) {
	builder := newAIContextBuilder()
	m := buildModelForAIContext()
	question := "How is the cognitive load distributed?"

	data, err := builder.BuildPromptData(m, question)
	require.NoError(t, err)

	assert.Equal(t, question, data["UserQuestion"])
	assert.Equal(t, question, data["Question"])
}

func TestAIContextBuilder_BuildPromptData_ContainsAnalysisData(t *testing.T) {
	builder := newAIContextBuilder()
	m := buildModelForAIContext()

	data, err := builder.BuildPromptData(m, "check bottlenecks")
	require.NoError(t, err)

	// All analysis keys should be present
	assert.Contains(t, data, "Bottlenecks")
	assert.Contains(t, data, "Couplings")
	assert.Contains(t, data, "Gaps")
	assert.Contains(t, data, "Complexities")
	assert.Contains(t, data, "ServiceCycles")
	assert.Contains(t, data, "CapabilityCycles")
	assert.Contains(t, data, "ValueStreamCoherence")
	assert.Contains(t, data, "CognitiveLoadDetails")
	assert.Contains(t, data, "FragmentedCapabilities")
}
