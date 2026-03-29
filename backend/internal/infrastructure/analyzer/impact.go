package analyzer

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
)

// ChangeKind describes whether a metric improved, regressed, or stayed the same.
type ChangeKind string

const (
	Improved  ChangeKind = "improved"
	Regressed ChangeKind = "regressed"
	Unchanged ChangeKind = "unchanged"
)

// DimensionDelta captures a before/after snapshot of one analysis dimension.
type DimensionDelta struct {
	Dimension string     // human-readable name: "fragmentation", "cognitive_load", etc.
	Before    string     // human-readable before value
	After     string     // human-readable after value
	Change    ChangeKind // improved / regressed / unchanged
	Detail    string     // optional: which specific entities changed
}

// ImpactReport summarizes the effect of applying a Changeset to a UNMModel.
type ImpactReport struct {
	ChangesetID string
	Deltas      []DimensionDelta
}

// ImpactAnalyzer runs all supported analyzers on both the original and projected
// models and diffs the results to produce an ImpactReport.
type ImpactAnalyzer struct {
	cfg entity.AnalysisConfig
}

// NewImpactAnalyzer constructs an ImpactAnalyzer using the provided AnalysisConfig
// for threshold-sensitive analyzers (e.g. cognitive load).
func NewImpactAnalyzer(cfg entity.AnalysisConfig) *ImpactAnalyzer {
	return &ImpactAnalyzer{cfg: cfg}
}

// Analyze applies the changeset to produce a projected model, runs analyzers on
// both, and returns a diff report.
func (a *ImpactAnalyzer) Analyze(m *entity.UNMModel, cs *entity.Changeset) (ImpactReport, error) {
	applier := service.NewChangesetApplier()
	projected, err := applier.Apply(m, cs)
	if err != nil {
		return ImpactReport{}, fmt.Errorf("impact analyzer: %w", err)
	}

	report := ImpactReport{
		ChangesetID: cs.ID,
	}

	report.Deltas = append(report.Deltas, diffFragmentation(m, projected))
	report.Deltas = append(report.Deltas, diffCognitiveLoad(m, projected, a.cfg))
	report.Deltas = append(report.Deltas, diffBottleneck(m, projected, a.cfg))
	report.Deltas = append(report.Deltas, diffCoupling(m, projected))
	report.Deltas = append(report.Deltas, diffValueChain(m, projected, a.cfg))
	report.Deltas = append(report.Deltas, diffValueStream(m, projected))

	return report, nil
}

func diffFragmentation(before, after *entity.UNMModel) DimensionDelta {
	fa := NewFragmentationAnalyzer()
	bReport := fa.Analyze(before)
	aReport := fa.Analyze(after)

	bCount := len(bReport.FragmentedCapabilities)
	aCount := len(aReport.FragmentedCapabilities)

	change := compareCountsLowerIsBetter(bCount, aCount)

	detail := ""
	if change != Unchanged {
		removed, added := diffNameSets(
			fragCapNames(bReport.FragmentedCapabilities),
			fragCapNames(aReport.FragmentedCapabilities),
		)
		detail = formatDiffDetail(removed, added)
	}

	return DimensionDelta{
		Dimension: "fragmentation",
		Before:    fmt.Sprintf("%d fragmented capabilities", bCount),
		After:     fmt.Sprintf("%d fragmented capabilities", aCount),
		Change:    change,
		Detail:    detail,
	}
}

func diffCognitiveLoad(before, after *entity.UNMModel, cfg entity.AnalysisConfig) DimensionDelta {
	ca := NewCognitiveLoadAnalyzer(cfg.CognitiveLoad, cfg.InteractionWeights)
	bReport := ca.Analyze(before)
	aReport := ca.Analyze(after)

	bCount := countHighLoadTeams(bReport.TeamLoads)
	aCount := countHighLoadTeams(aReport.TeamLoads)

	change := compareCountsLowerIsBetter(bCount, aCount)

	detail := ""
	if change != Unchanged {
		bNames := highLoadTeamNames(bReport.TeamLoads)
		aNames := highLoadTeamNames(aReport.TeamLoads)
		removed, added := diffNameSets(bNames, aNames)
		detail = formatDiffDetail(removed, added)
	}

	return DimensionDelta{
		Dimension: "cognitive_load",
		Before:    fmt.Sprintf("%d teams at high load", bCount),
		After:     fmt.Sprintf("%d teams at high load", aCount),
		Change:    change,
		Detail:    detail,
	}
}

func diffBottleneck(before, after *entity.UNMModel, cfg entity.AnalysisConfig) DimensionDelta {
	ba := NewBottleneckAnalyzer(cfg.Bottleneck)
	bReport := ba.Analyze(before)
	aReport := ba.Analyze(after)

	bCount := countCriticalBottlenecks(bReport.ServiceBottlenecks)
	aCount := countCriticalBottlenecks(aReport.ServiceBottlenecks)

	change := compareCountsLowerIsBetter(bCount, aCount)

	return DimensionDelta{
		Dimension: "bottleneck",
		Before:    fmt.Sprintf("%d critical bottlenecks", bCount),
		After:     fmt.Sprintf("%d critical bottlenecks", aCount),
		Change:    change,
	}
}

func diffCoupling(before, after *entity.UNMModel) DimensionDelta {
	ca := NewCouplingAnalyzer()
	bReport := ca.Analyze(before)
	aReport := ca.Analyze(after)

	bCount := countHighCoupling(bReport.DataAssetCouplings)
	aCount := countHighCoupling(aReport.DataAssetCouplings)

	change := compareCountsLowerIsBetter(bCount, aCount)

	return DimensionDelta{
		Dimension: "coupling",
		Before:    fmt.Sprintf("%d highly coupled assets", bCount),
		After:     fmt.Sprintf("%d highly coupled assets", aCount),
		Change:    change,
	}
}

func diffValueChain(before, after *entity.UNMModel, cfg entity.AnalysisConfig) DimensionDelta {
	va := NewValueChainAnalyzer(cfg.ValueChain)
	bReport := va.Analyze(before)
	aReport := va.Analyze(after)

	bCount := countAtRiskNeeds(bReport.NeedRisks)
	aCount := countAtRiskNeeds(aReport.NeedRisks)

	change := compareCountsLowerIsBetter(bCount, aCount)

	return DimensionDelta{
		Dimension: "value_chain",
		Before:    fmt.Sprintf("%d at-risk needs", bCount),
		After:     fmt.Sprintf("%d at-risk needs", aCount),
		Change:    change,
	}
}

func diffValueStream(before, after *entity.UNMModel) DimensionDelta {
	va := NewValueStreamAnalyzer()
	bReport := va.Analyze(before)
	aReport := va.Analyze(after)

	bCount := countLowCoherence(bReport.TeamCoherences)
	aCount := countLowCoherence(aReport.TeamCoherences)

	change := compareCountsLowerIsBetter(bCount, aCount)

	return DimensionDelta{
		Dimension: "value_stream",
		Before:    fmt.Sprintf("%d low-coherence teams", bCount),
		After:     fmt.Sprintf("%d low-coherence teams", aCount),
		Change:    change,
	}
}

// compareCountsLowerIsBetter returns the ChangeKind when lower values are better.
func compareCountsLowerIsBetter(before, after int) ChangeKind {
	if after < before {
		return Improved
	}
	if after > before {
		return Regressed
	}
	return Unchanged
}

// --- counting helpers ---

func countHighLoadTeams(loads []TeamLoad) int {
	count := 0
	for _, tl := range loads {
		if tl.OverallLevel == LoadHigh {
			count++
		}
	}
	return count
}

func highLoadTeamNames(loads []TeamLoad) []string {
	var names []string
	for _, tl := range loads {
		if tl.OverallLevel == LoadHigh {
			names = append(names, tl.Team.Name)
		}
	}
	return names
}

func countCriticalBottlenecks(bns []ServiceBottleneck) int {
	count := 0
	for _, b := range bns {
		if b.IsCritical {
			count++
		}
	}
	return count
}

func countHighCoupling(couplings []DataAssetCoupling) int {
	count := 0
	for _, c := range couplings {
		if len(c.Services) > 2 {
			count++
		}
	}
	return count
}

func countAtRiskNeeds(risks []NeedDeliveryRisk) int {
	count := 0
	for _, r := range risks {
		if r.AtRisk {
			count++
		}
	}
	return count
}

func countLowCoherence(coherences []TeamStreamCoherence) int {
	count := 0
	for _, c := range coherences {
		if c.LowCoherence {
			count++
		}
	}
	return count
}

func fragCapNames(caps []FragmentedCapability) []string {
	names := make([]string, len(caps))
	for i, fc := range caps {
		names[i] = fc.Capability.Name
	}
	return names
}

// diffNameSets returns names removed and added between two sets.
func diffNameSets(before, after []string) (removed, added []string) {
	bSet := make(map[string]bool, len(before))
	for _, n := range before {
		bSet[n] = true
	}
	aSet := make(map[string]bool, len(after))
	for _, n := range after {
		aSet[n] = true
	}
	for _, n := range before {
		if !aSet[n] {
			removed = append(removed, n)
		}
	}
	for _, n := range after {
		if !bSet[n] {
			added = append(added, n)
		}
	}
	return
}

func formatDiffDetail(removed, added []string) string {
	detail := ""
	if len(removed) > 0 {
		detail += fmt.Sprintf("removed: %v", removed)
	}
	if len(added) > 0 {
		if detail != "" {
			detail += "; "
		}
		detail += fmt.Sprintf("added: %v", added)
	}
	return detail
}
