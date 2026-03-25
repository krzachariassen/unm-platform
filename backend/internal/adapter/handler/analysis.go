package handler

import (
	"fmt"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// registerAnalysisRoutes registers analysis endpoints:
//   - POST /api/analyze/{type}             – submit YAML, get analysis (no stored model)
//   - GET  /api/models/{id}/analyze/{type} – analyze a stored model by ID
func (h *Handler) registerAnalysisRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/analyze/{type}", h.handleAnalyze)
	mux.HandleFunc("GET /api/models/{id}/analyze/{type}", h.handleAnalyzeStored)
}

// handleAnalyze runs the requested analysis type on a submitted YAML model.
func (h *Handler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	analyzeType := r.PathValue("type")
	if !validAnalysisType(analyzeType) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown analysis type: %q", analyzeType))
		return
	}

	model, _, err := h.parseAndValidate.Execute(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.runAnalysis(analyzeType, model))
}

// handleAnalyzeStored runs analysis on a model already stored by ID.
func (h *Handler) handleAnalyzeStored(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	analyzeType := r.PathValue("type")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	if !validAnalysisType(analyzeType) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown analysis type: %q", analyzeType))
		return
	}

	writeJSON(w, http.StatusOK, h.runAnalysis(analyzeType, stored.Model))
}

func validAnalysisType(t string) bool {
	switch t {
	case "fragmentation", "cognitive-load", "dependencies", "gaps",
		"bottleneck", "coupling", "complexity", "interactions", "unlinked", "all":
		return true
	}
	return false
}

// runAnalysis executes the named analysis on m and returns a JSON-ready map.
func (h *Handler) runAnalysis(analyzeType string, m *entity.UNMModel) map[string]any {
	switch analyzeType {
	case "fragmentation":
		return buildFragmentationResponse(h.fragmentation.Analyze(m))
	case "cognitive-load":
		return buildCognitiveLoadResponse(h.cognitiveLoad.Analyze(m))
	case "dependencies":
		return buildDependenciesResponse(h.dependency.Analyze(m))
	case "gaps":
		return buildGapsResponse(h.gap.Analyze(m))
	case "bottleneck":
		return buildBottleneckResponse(h.bottleneck.Analyze(m))
	case "coupling":
		return buildCouplingResponse(h.coupling.Analyze(m))
	case "complexity":
		return buildComplexityResponse(h.complexity.Analyze(m))
	case "interactions":
		return buildInteractionsResponse(h.interactions.Analyze(m))
	case "unlinked":
		return buildUnlinkedResponse(h.unlinked.Analyze(m))
	case "all":
		fragReport   := h.fragmentation.Analyze(m)
		cogReport    := h.cognitiveLoad.Analyze(m)
		depReport    := h.dependency.Analyze(m)
		gapReport    := h.gap.Analyze(m)
		bnReport     := h.bottleneck.Analyze(m)
		cpReport     := h.coupling.Analyze(m)
		cxReport     := h.complexity.Analyze(m)
		intrReport   := h.interactions.Analyze(m)
		unlReport    := h.unlinked.Analyze(m)
		sgReport     := h.signalSuggestions.Generate(bnReport, cogReport, fragReport, depReport, unlReport, m)
		return map[string]any{
			"type":                 "all",
			"fragmentation":        buildFragmentationResponse(fragReport),
			"cognitive_load":       buildCognitiveLoadResponse(cogReport),
			"dependencies":         buildDependenciesResponse(depReport),
			"gaps":                 buildGapsResponse(gapReport),
			"bottleneck":           buildBottleneckResponse(bnReport),
			"coupling":             buildCouplingResponse(cpReport),
			"complexity":           buildComplexityResponse(cxReport),
			"interactions":         buildInteractionsResponse(intrReport),
			"unlinked":             buildUnlinkedResponse(unlReport),
			"signal_suggestions":   buildSignalSuggestionsResponse(sgReport),
		}
	}
	return map[string]any{}
}

// ── Response builders ──────────────────────────────────────────────────────────

// capabilityView is the JSON shape for a Capability.
type capabilityView struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// teamView is the JSON shape for a Team.
type teamView struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// needView is the JSON shape for a Need in gaps analysis.
type needView struct {
	Name      string `json:"name"`
	ActorName string `json:"actor_name"`
}

func buildFragmentationResponse(report analyzer.FragmentationReport) map[string]any {
	type fragmentedCapabilityView struct {
		Capability capabilityView `json:"capability"`
		Teams      []teamView     `json:"teams"`
	}

	items := make([]fragmentedCapabilityView, 0, len(report.FragmentedCapabilities))
	for _, fc := range report.FragmentedCapabilities {
		teams := make([]teamView, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teams = append(teams, teamView{Name: t.Name, Type: t.TeamType.String()})
		}
		items = append(items, fragmentedCapabilityView{
			Capability: capabilityView{Name: fc.Capability.Name, Description: fc.Capability.Description},
			Teams:      teams,
		})
	}

	return map[string]any{
		"type":                    "fragmentation",
		"fragmented_capabilities": items,
	}
}

func buildCognitiveLoadResponse(report analyzer.CognitiveLoadReport) map[string]any {
	type dimView struct {
		Value float64 `json:"value"`
		Level string  `json:"level"`
	}
	type teamLoadView struct {
		Team             teamView `json:"team"`
		CapabilityCount  int      `json:"capability_count"`
		ServiceCount     int      `json:"service_count"`
		DependencyCount  int      `json:"dependency_count"`
		InteractionCount int      `json:"interaction_count"`
		InteractionScore int      `json:"interaction_score"`
		TeamSize         int      `json:"team_size"`
		SizeIsExplicit   bool     `json:"size_is_explicit"`
		DomainSpread     dimView  `json:"domain_spread"`
		ServiceLoad      dimView  `json:"service_load"`
		InteractionLoad  dimView  `json:"interaction_load"`
		DependencyLoad   dimView  `json:"dependency_load"`
		OverallLevel     string   `json:"overall_level"`
		IsOverloaded     bool     `json:"is_overloaded"`
	}

	loads := make([]teamLoadView, 0, len(report.TeamLoads))
	for _, tl := range report.TeamLoads {
		loads = append(loads, teamLoadView{
			Team:             teamView{Name: tl.Team.Name, Type: tl.Team.TeamType.String()},
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

func buildDependenciesResponse(report analyzer.DependencyReport) map[string]any {
	serviceCycles := cyclePathsOrEmpty(report.ServiceCycles)
	capCycles := cyclePathsOrEmpty(report.CapabilityCycles)
	critPath := report.CriticalServicePath
	if critPath == nil {
		critPath = []string{}
	}

	return map[string]any{
		"type":                  "dependencies",
		"max_service_depth":     report.MaxServiceDepth,
		"max_capability_depth":  report.MaxCapabilityDepth,
		"critical_service_path": critPath,
		"service_cycles":        serviceCycles,
		"capability_cycles":     capCycles,
	}
}

func cyclePathsOrEmpty(cycles []analyzer.DependencyCycle) [][]string {
	if cycles == nil {
		return [][]string{}
	}
	paths := make([][]string, 0, len(cycles))
	for _, c := range cycles {
		paths = append(paths, c.Path)
	}
	return paths
}

func buildGapsResponse(report analyzer.GapReport) map[string]any {
	unmappedNeeds := make([]needView, 0, len(report.UnmappedNeeds))
	for _, n := range report.UnmappedNeeds {
		unmappedNeeds = append(unmappedNeeds, needView{Name: n.Name, ActorName: n.ActorName})
	}

	unrealizedCaps := capabilityViewsOrEmpty(report.UnrealizedCapabilities)
	unownedSvcs := serviceViewsOrEmpty(report.UnownedServices)
	unneededCaps := capabilityViewsOrEmpty(report.UnneededCapabilities)

	return map[string]any{
		"type":                    "gaps",
		"unmapped_needs":          unmappedNeeds,
		"unrealized_capabilities": unrealizedCaps,
		"unowned_services":        unownedSvcs,
		"unneeded_capabilities":   unneededCaps,
	}
}

func capabilityViewsOrEmpty(caps []*entity.Capability) []capabilityView {
	views := make([]capabilityView, 0, len(caps))
	for _, c := range caps {
		views = append(views, capabilityView{Name: c.Name, Description: c.Description})
	}
	return views
}

func serviceViewsOrEmpty(svcs []*entity.Service) []map[string]string {
	views := make([]map[string]string, 0, len(svcs))
	for _, s := range svcs {
		views = append(views, map[string]string{"name": s.Name})
	}
	return views
}

func buildBottleneckResponse(report analyzer.BottleneckReport) map[string]any {
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

func buildCouplingResponse(report analyzer.CouplingReport) map[string]any {
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

func buildComplexityResponse(report analyzer.ComplexityReport) map[string]any {
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

func buildInteractionsResponse(report analyzer.InteractionDiversityReport) map[string]any {
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
		"type":              "interactions",
		"mode_distribution": modeDist,
		"all_modes_same":    report.AllModesSame,
		"isolated_teams":    isolated,
		"over_reliant_teams": overReliant,
	}
}

func buildUnlinkedResponse(report analyzer.UnlinkedCapabilityReport) map[string]any {
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
		"type":                      "unlinked",
		"total_leaf_capabilities":   report.TotalLeafCapabilityCount,
		"linked_count":              report.LinkedCount,
		"linked_percentage":         report.LinkedPercentage,
		"by_visibility":             byVis,
		"unlinked_leaf_capabilities": items,
	}
}

func buildSignalSuggestionsResponse(report analyzer.SignalSuggestionsReport) map[string]any {
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
