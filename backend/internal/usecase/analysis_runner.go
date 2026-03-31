package usecase

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// AnalysisRunner dispatches analysis requests to the appropriate analyzer
// and returns the result as a JSON-ready map.
type AnalysisRunner struct {
	fragmentation     Fragmenter
	cognitiveLoad     CognitiveLoader
	dependency        DependencyAnalyzer
	gap               GapAnalyzer
	bottleneck        BottleneckAnalyzer
	coupling          CouplingAnalyzer
	complexity        ComplexityAnalyzer
	interactions      InteractionDiversityAnalyzer
	unlinked          UnlinkedCapabilityAnalyzer
	signalSuggestions SignalGenerator
}

// NewAnalysisRunner constructs an AnalysisRunner.
func NewAnalysisRunner(
	frag Fragmenter,
	cl CognitiveLoader,
	dep DependencyAnalyzer,
	g GapAnalyzer,
	bn BottleneckAnalyzer,
	cp CouplingAnalyzer,
	cx ComplexityAnalyzer,
	intr InteractionDiversityAnalyzer,
	unl UnlinkedCapabilityAnalyzer,
	sg SignalGenerator,
) *AnalysisRunner {
	return &AnalysisRunner{
		fragmentation:     frag,
		cognitiveLoad:     cl,
		dependency:        dep,
		gap:               g,
		bottleneck:        bn,
		coupling:          cp,
		complexity:        cx,
		interactions:      intr,
		unlinked:          unl,
		signalSuggestions: sg,
	}
}

// ValidAnalysisType returns true if analyzeType is a supported analysis type key.
func ValidAnalysisType(analyzeType string) bool {
	switch analyzeType {
	case "fragmentation", "cognitive-load", "dependencies", "gaps",
		"bottleneck", "coupling", "complexity", "interactions", "unlinked", "all":
		return true
	}
	return false
}

// RunAnalysis executes the named analysis on m and returns a JSON-ready map.
// Returns an error if the analysis type is not valid.
func (r *AnalysisRunner) RunAnalysis(analyzeType string, m *entity.UNMModel) (map[string]any, error) {
	switch analyzeType {
	case "fragmentation":
		return buildFragmentationResult(r.fragmentation.Analyze(m)), nil
	case "cognitive-load":
		return buildCognitiveLoadResult(r.cognitiveLoad.Analyze(m)), nil
	case "dependencies":
		return buildDependenciesResult(r.dependency.Analyze(m)), nil
	case "gaps":
		return buildGapsResult(r.gap.Analyze(m)), nil
	case "bottleneck":
		return buildBottleneckResult(r.bottleneck.Analyze(m)), nil
	case "coupling":
		return buildCouplingResult(r.coupling.Analyze(m)), nil
	case "complexity":
		return buildComplexityResult(r.complexity.Analyze(m)), nil
	case "interactions":
		return buildInteractionsResult(r.interactions.Analyze(m)), nil
	case "unlinked":
		return buildUnlinkedResult(r.unlinked.Analyze(m)), nil
	case "all":
		fragReport := r.fragmentation.Analyze(m)
		cogReport := r.cognitiveLoad.Analyze(m)
		depReport := r.dependency.Analyze(m)
		gapReport := r.gap.Analyze(m)
		bnReport := r.bottleneck.Analyze(m)
		cpReport := r.coupling.Analyze(m)
		cxReport := r.complexity.Analyze(m)
		intrReport := r.interactions.Analyze(m)
		unlReport := r.unlinked.Analyze(m)
		sgReport := r.signalSuggestions.Generate(bnReport, cogReport, fragReport, depReport, unlReport, m)
		return map[string]any{
			"type":               "all",
			"fragmentation":      buildFragmentationResult(fragReport),
			"cognitive_load":     buildCognitiveLoadResult(cogReport),
			"dependencies":       buildDependenciesResult(depReport),
			"gaps":               buildGapsResult(gapReport),
			"bottleneck":         buildBottleneckResult(bnReport),
			"coupling":           buildCouplingResult(cpReport),
			"complexity":         buildComplexityResult(cxReport),
			"interactions":       buildInteractionsResult(intrReport),
			"unlinked":           buildUnlinkedResult(unlReport),
			"signal_suggestions": buildSignalSuggestionsResult(sgReport),
		}, nil
	}
	return nil, fmt.Errorf("unknown analysis type: %q", analyzeType)
}

// ── Result builders ───────────────────────────────────────────────────────────
// These are pure data transformations from analyzer report types to JSON-ready maps.

type analysisCapabilityView struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type analysisTeamView struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type analysisNeedView struct {
	Name       string   `json:"name"`
	ActorNames []string `json:"actor_names"`
}

func buildFragmentationResult(report analyzer.FragmentationReport) map[string]any {
	type fragmentedCapabilityView struct {
		Capability analysisCapabilityView `json:"capability"`
		Teams      []analysisTeamView     `json:"teams"`
	}

	items := make([]fragmentedCapabilityView, 0, len(report.FragmentedCapabilities))
	for _, fc := range report.FragmentedCapabilities {
		teams := make([]analysisTeamView, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teams = append(teams, analysisTeamView{Name: t.Name, Type: t.TeamType.String()})
		}
		items = append(items, fragmentedCapabilityView{
			Capability: analysisCapabilityView{Name: fc.Capability.Name, Description: fc.Capability.Description},
			Teams:      teams,
		})
	}
	return map[string]any{
		"type":                    "fragmentation",
		"fragmented_capabilities": items,
	}
}

func buildCognitiveLoadResult(report analyzer.CognitiveLoadReport) map[string]any {
	type dimView struct {
		Value float64 `json:"value"`
		Level string  `json:"level"`
	}
	type teamLoadView struct {
		Team             analysisTeamView `json:"team"`
		CapabilityCount  int              `json:"capability_count"`
		ServiceCount     int              `json:"service_count"`
		DependencyCount  int              `json:"dependency_count"`
		InteractionCount int              `json:"interaction_count"`
		InteractionScore int              `json:"interaction_score"`
		TeamSize         int              `json:"team_size"`
		SizeIsExplicit   bool             `json:"size_is_explicit"`
		DomainSpread     dimView          `json:"domain_spread"`
		ServiceLoad      dimView          `json:"service_load"`
		InteractionLoad  dimView          `json:"interaction_load"`
		DependencyLoad   dimView          `json:"dependency_load"`
		OverallLevel     string           `json:"overall_level"`
		IsOverloaded     bool             `json:"is_overloaded"`
	}

	loads := make([]teamLoadView, 0, len(report.TeamLoads))
	for _, tl := range report.TeamLoads {
		loads = append(loads, teamLoadView{
			Team:             analysisTeamView{Name: tl.Team.Name, Type: tl.Team.TeamType.String()},
			CapabilityCount:  tl.CapabilityCount,
			ServiceCount:     tl.ServiceCount,
			DependencyCount:  tl.DependencyCount,
			InteractionCount: tl.InteractionCount,
			InteractionScore: tl.InteractionScore,
			TeamSize:         tl.TeamSize,
			SizeIsExplicit:   tl.SizeIsExplicit,
			DomainSpread:     dimView{Value: tl.DomainSpread.Value, Level: string(tl.DomainSpread.Level)},
			ServiceLoad:      dimView{Value: tl.ServiceLoad.Value, Level: string(tl.ServiceLoad.Level)},
			InteractionLoad:  dimView{Value: tl.InteractionLoad.Value, Level: string(tl.InteractionLoad.Level)},
			DependencyLoad:   dimView{Value: tl.DependencyLoad.Value, Level: string(tl.DependencyLoad.Level)},
			OverallLevel:     string(tl.OverallLevel),
			IsOverloaded:     tl.OverallLevel == analyzer.LoadHigh,
		})
	}
	return map[string]any{
		"type":       "cognitive-load",
		"team_loads": loads,
	}
}

func buildDependenciesResult(report analyzer.DependencyReport) map[string]any {
	critPath := report.CriticalServicePath
	if critPath == nil {
		critPath = []string{}
	}
	return map[string]any{
		"type":                  "dependencies",
		"max_service_depth":     report.MaxServiceDepth,
		"max_capability_depth":  report.MaxCapabilityDepth,
		"critical_service_path": critPath,
		"service_cycles":        analysisDepCyclePaths(report.ServiceCycles),
		"capability_cycles":     analysisDepCyclePaths(report.CapabilityCycles),
	}
}

func analysisDepCyclePaths(cycles []analyzer.DependencyCycle) [][]string {
	if cycles == nil {
		return [][]string{}
	}
	paths := make([][]string, 0, len(cycles))
	for _, c := range cycles {
		paths = append(paths, c.Path)
	}
	return paths
}

func buildGapsResult(report analyzer.GapReport) map[string]any {
	unmappedNeeds := make([]analysisNeedView, 0, len(report.UnmappedNeeds))
	for _, n := range report.UnmappedNeeds {
		unmappedNeeds = append(unmappedNeeds, analysisNeedView{Name: n.Name, ActorNames: n.ActorNames})
	}

	unrealizedCaps := make([]analysisCapabilityView, 0, len(report.UnrealizedCapabilities))
	for _, c := range report.UnrealizedCapabilities {
		unrealizedCaps = append(unrealizedCaps, analysisCapabilityView{Name: c.Name, Description: c.Description})
	}

	unownedSvcs := make([]map[string]string, 0, len(report.UnownedServices))
	for _, s := range report.UnownedServices {
		unownedSvcs = append(unownedSvcs, map[string]string{"name": s.Name})
	}

	unneededCaps := make([]analysisCapabilityView, 0, len(report.UnneededCapabilities))
	for _, c := range report.UnneededCapabilities {
		unneededCaps = append(unneededCaps, analysisCapabilityView{Name: c.Name, Description: c.Description})
	}

	return map[string]any{
		"type":                    "gaps",
		"unmapped_needs":          unmappedNeeds,
		"unrealized_capabilities": unrealizedCaps,
		"unowned_services":        unownedSvcs,
		"unneeded_capabilities":   unneededCaps,
	}
}

func buildBottleneckResult(report analyzer.BottleneckReport) map[string]any {
	type bottleneckView struct {
		Service    string `json:"service"`
		FanIn      int    `json:"fan_in"`
		FanOut     int    `json:"fan_out"`
		IsCritical bool   `json:"is_critical"`
		IsWarning  bool   `json:"is_warning"`
	}

	items := make([]bottleneckView, 0, len(report.ServiceBottlenecks))
	for _, b := range report.ServiceBottlenecks {
		items = append(items, bottleneckView{
			Service:    b.Service.Name,
			FanIn:      b.FanIn,
			FanOut:     b.FanOut,
			IsCritical: b.IsCritical,
			IsWarning:  b.IsWarning,
		})
	}
	return map[string]any{
		"type":        "bottleneck",
		"bottlenecks": items,
	}
}

func buildCouplingResult(report analyzer.CouplingReport) map[string]any {
	type couplingView struct {
		DataAsset   string   `json:"data_asset"`
		Type        string   `json:"type"`
		Services    []string `json:"services"`
		IsCrossteam bool     `json:"is_crossteam"`
	}

	items := make([]couplingView, 0, len(report.DataAssetCouplings))
	for _, c := range report.DataAssetCouplings {
		assetType := ""
		if c.DataAsset != nil {
			assetType = c.DataAsset.Type
		}
		items = append(items, couplingView{
			DataAsset:   c.DataAsset.Name,
			Type:        assetType,
			Services:    c.Services,
			IsCrossteam: c.IsCrossteam,
		})
	}
	return map[string]any{
		"type":      "coupling",
		"couplings": items,
	}
}

func buildComplexityResult(report analyzer.ComplexityReport) map[string]any {
	type complexityView struct {
		Service         string `json:"service"`
		DependencyScore int    `json:"dependency_score"`
		CapabilityScore int    `json:"capability_score"`
		DataAssetScore  int    `json:"data_asset_score"`
		TotalComplexity int    `json:"total_complexity"`
	}

	items := make([]complexityView, 0, len(report.Services))
	for _, s := range report.Services {
		items = append(items, complexityView{
			Service:         s.Service.Name,
			DependencyScore: s.DependencyScore,
			CapabilityScore: s.CapabilityScore,
			DataAssetScore:  s.DataAssetScore,
			TotalComplexity: s.TotalComplexity,
		})
	}
	return map[string]any{
		"type":     "complexity",
		"services": items,
	}
}

func buildInteractionsResult(report analyzer.InteractionDiversityReport) map[string]any {
	type overReliantView struct {
		TeamName string `json:"team_name"`
		Mode     string `json:"mode"`
		Count    int    `json:"count"`
	}

	modeDist := make(map[string]int, len(report.ModeDistribution))
	for mode, count := range report.ModeDistribution {
		modeDist[string(mode)] = count
	}

	overReliant := make([]overReliantView, 0, len(report.OverReliantTeams))
	for _, or_ := range report.OverReliantTeams {
		overReliant = append(overReliant, overReliantView{
			TeamName: or_.TeamName,
			Mode:     string(or_.Mode),
			Count:    or_.Count,
		})
	}

	isolated := report.IsolatedTeams
	if isolated == nil {
		isolated = []string{}
	}

	return map[string]any{
		"type":               "interactions",
		"mode_distribution":  modeDist,
		"all_modes_same":     report.AllModesSame,
		"isolated_teams":     isolated,
		"over_reliant_teams": overReliant,
	}
}

func buildUnlinkedResult(report analyzer.UnlinkedCapabilityReport) map[string]any {
	type unlinkedView struct {
		Capability string `json:"capability"`
		Visibility string `json:"visibility"`
		IsExpected bool   `json:"is_expected"`
	}

	items := make([]unlinkedView, 0, len(report.UnlinkedLeafCapabilities))
	for _, uc := range report.UnlinkedLeafCapabilities {
		items = append(items, unlinkedView{
			Capability: uc.Capability.Name,
			Visibility: uc.Visibility,
			IsExpected: uc.IsExpected,
		})
	}

	byVis := make(map[string]int, len(report.ByVisibility))
	for k, v := range report.ByVisibility {
		byVis[k] = v
	}

	return map[string]any{
		"type":                       "unlinked",
		"total_leaf_capabilities":    report.TotalLeafCapabilityCount,
		"linked_count":               report.LinkedCount,
		"linked_percentage":          report.LinkedPercentage,
		"by_visibility":              byVis,
		"unlinked_leaf_capabilities": items,
	}
}

func buildSignalSuggestionsResult(report analyzer.SignalSuggestionsReport) map[string]any {
	type suggestionView struct {
		Category     string `json:"category"`
		OnEntityName string `json:"on_entity_name"`
		Description  string `json:"description"`
		Evidence     string `json:"evidence"`
		Severity     string `json:"severity"`
		Source       string `json:"source"`
	}

	items := make([]suggestionView, 0, len(report.Suggestions))
	for _, s := range report.Suggestions {
		items = append(items, suggestionView{
			Category:     s.Category,
			OnEntityName: s.OnEntityName,
			Description:  s.Description,
			Evidence:     s.Evidence,
			Severity:     string(s.Severity),
			Source:       s.Source,
		})
	}

	return map[string]any{
		"type":        "signal_suggestions",
		"suggestions": items,
	}
}
