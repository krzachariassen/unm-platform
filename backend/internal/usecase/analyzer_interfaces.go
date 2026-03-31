package usecase

// This file defines the analyzer interfaces depended on by use cases.
// Concrete implementations live in internal/infrastructure/analyzer and satisfy
// these interfaces implicitly — no changes to the infrastructure package are needed.
//
// Having interfaces defined here (where they are used) follows the Clean Architecture
// dependency inversion rule: the use case layer defines what it needs; the
// infrastructure layer satisfies it.

import (
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// Fragmenter computes capability fragmentation across teams.
type Fragmenter interface {
	Analyze(m *entity.UNMModel) analyzer.FragmentationReport
}

// CognitiveLoader computes cognitive load per team.
type CognitiveLoader interface {
	Analyze(m *entity.UNMModel) analyzer.CognitiveLoadReport
}

// DependencyAnalyzer computes service and capability dependency depth and cycles.
type DependencyAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.DependencyReport
}

// GapAnalyzer detects unmapped needs, unrealized capabilities, and unowned services.
type GapAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.GapReport
}

// BottleneckAnalyzer identifies high fan-in services and external dependencies.
type BottleneckAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.BottleneckReport
}

// CouplingAnalyzer detects data asset coupling across services.
type CouplingAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.CouplingReport
}

// ComplexityAnalyzer scores service complexity by dependency and capability load.
type ComplexityAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.ComplexityReport
}

// InteractionDiversityAnalyzer checks distribution of team interaction modes.
type InteractionDiversityAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.InteractionDiversityReport
}

// UnlinkedCapabilityAnalyzer finds leaf capabilities with no need-to-service path.
type UnlinkedCapabilityAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.UnlinkedCapabilityReport
}

// ValueChainAnalyzer computes need delivery risk through the capability/service chain.
type ValueChainAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.ValueChainReport
}

// ValueStreamAnalyzer computes team coherence via value stream analysis.
type ValueStreamAnalyzer interface {
	Analyze(m *entity.UNMModel) analyzer.ValueStreamReport
}

// SignalGenerator synthesizes findings from multiple analysis reports into suggestions.
type SignalGenerator interface {
	Generate(
		bottleneckReport analyzer.BottleneckReport,
		cognitiveLoadReport analyzer.CognitiveLoadReport,
		fragmentationReport analyzer.FragmentationReport,
		depReport analyzer.DependencyReport,
		unlinkedReport analyzer.UnlinkedCapabilityReport,
		m *entity.UNMModel,
	) analyzer.SignalSuggestionsReport
}

// ImpactRunner runs changeset impact analysis.
type ImpactRunner interface {
	Analyze(m *entity.UNMModel, cs *entity.Changeset) (analyzer.ImpactReport, error)
}
