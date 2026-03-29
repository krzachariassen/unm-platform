package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func buildMinimalModelForAnalysis() *entity.UNMModel {
	m := &entity.UNMModel{}
	m.System.Name = "AnalysisTest"
	m.Teams = map[string]*entity.Team{}
	m.Services = map[string]*entity.Service{}
	m.Capabilities = map[string]*entity.Capability{}
	m.Actors = map[string]*entity.Actor{}
	m.Needs = map[string]*entity.Need{}
	return m
}

func newAnalysisRunner() *usecase.AnalysisRunner {
	cfg := entity.DefaultConfig().Analysis
	return usecase.NewAnalysisRunner(
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.CognitiveLoad, cfg.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Signals),
	)
}

func TestAnalysisRunner_RunAnalysis_Fragmentation(t *testing.T) {
	runner := newAnalysisRunner()
	m := buildMinimalModelForAnalysis()
	result, err := runner.RunAnalysis("fragmentation", m)
	require.NoError(t, err)
	assert.Equal(t, "fragmentation", result["type"])
}

func TestAnalysisRunner_RunAnalysis_CognitiveLoad(t *testing.T) {
	runner := newAnalysisRunner()
	m := buildMinimalModelForAnalysis()
	result, err := runner.RunAnalysis("cognitive-load", m)
	require.NoError(t, err)
	assert.Equal(t, "cognitive-load", result["type"])
}

func TestAnalysisRunner_RunAnalysis_All(t *testing.T) {
	runner := newAnalysisRunner()
	m := buildMinimalModelForAnalysis()
	result, err := runner.RunAnalysis("all", m)
	require.NoError(t, err)
	assert.Equal(t, "all", result["type"])
	// Should contain sub-analyses
	assert.Contains(t, result, "fragmentation")
	assert.Contains(t, result, "cognitive_load")
	assert.Contains(t, result, "bottleneck")
}

func TestAnalysisRunner_RunAnalysis_UnknownType_ReturnsError(t *testing.T) {
	runner := newAnalysisRunner()
	m := buildMinimalModelForAnalysis()
	_, err := runner.RunAnalysis("nonexistent-type", m)
	assert.Error(t, err)
}

func TestAnalysisRunner_ValidAnalysisType(t *testing.T) {
	validTypes := []string{
		"fragmentation", "cognitive-load", "dependencies", "gaps",
		"bottleneck", "coupling", "complexity", "interactions", "unlinked", "all",
	}
	for _, at := range validTypes {
		assert.True(t, usecase.ValidAnalysisType(at), "expected %q to be valid", at)
	}
	assert.False(t, usecase.ValidAnalysisType("bad-type"))
}
