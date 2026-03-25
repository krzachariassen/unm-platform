package handler

import (
	"net/http"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/infrastructure/analyzer"
)

func (h *Handler) registerSignalsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models/{id}/views/signals", h.handleSignals)
}

// signalHealth is the three-layer traffic-light summary.
type signalHealth struct {
	UXRisk           string `json:"ux_risk"`
	ArchitectureRisk string `json:"architecture_risk"`
	OrgRisk          string `json:"org_risk"`
}

type signalNeedRisk struct {
	NeedName  string   `json:"need_name"`
	ActorName string   `json:"actor_name"`
	TeamSpan  int      `json:"team_span"`
	Teams     []string `json:"teams"`
}

type signalCapItem struct {
	CapabilityName string   `json:"capability_name"`
	Visibility     string   `json:"visibility,omitempty"`
	TeamCount      int      `json:"team_count,omitempty"`
	Teams          []string `json:"teams,omitempty"`
}

type signalTeamItem struct {
	TeamName        string  `json:"team_name"`
	TeamType        string  `json:"team_type"`
	OverallLevel    string  `json:"overall_level"`
	CapabilityCount int     `json:"capability_count"`
	ServiceCount    int     `json:"service_count"`
	CoherenceScore  float64 `json:"coherence_score,omitempty"`
}

type signalServiceItem struct {
	ServiceName string `json:"service_name"`
	FanIn       int    `json:"fan_in"`
}

type signalUXLayer struct {
	NeedsRequiring3PlusTeams []signalNeedRisk `json:"needs_requiring_3plus_teams"`
	NeedsWithNoCapBacking    []signalNeedRisk `json:"needs_with_no_capability_backing"`
	NeedsAtRisk              []signalNeedRisk `json:"needs_at_risk"`
}

type signalArchLayer struct {
	UserFacingCapsWithCrossTeamServices []signalCapItem `json:"user_facing_caps_with_cross_team_services"`
	CapabilitiesNotConnectedToAnyNeed   []signalCapItem `json:"capabilities_not_connected_to_any_need"`
	CapabilitiesFragmentedAcrossTeams   []signalCapItem `json:"capabilities_fragmented_across_teams"`
}

type signalOrgLayer struct {
	TopTeamsByStructuralLoad   []signalTeamItem   `json:"top_teams_by_structural_load"`
	CriticalBottleneckServices []signalServiceItem `json:"critical_bottleneck_services"`
	LowCoherenceTeams          []signalTeamItem   `json:"low_coherence_teams"`
	CriticalExternalDeps       []SignalsExtDepItem `json:"critical_external_deps"`
}

// SignalsExtDepItem summarises an external dependency with notable service fan-in.
type SignalsExtDepItem struct {
	DepName      string   `json:"dep_name"`
	ServiceCount int      `json:"service_count"`
	Services     []string `json:"services"`
	IsCritical   bool     `json:"is_critical"`
	IsWarning    bool     `json:"is_warning"`
}

type signalsResponse struct {
	ViewType            string          `json:"view_type"`
	Health              signalHealth    `json:"health"`
	UserExperienceLayer signalUXLayer   `json:"user_experience_layer"`
	ArchitectureLayer   signalArchLayer `json:"architecture_layer"`
	OrganizationLayer   signalOrgLayer  `json:"organization_layer"`
}

func (h *Handler) handleSignals(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	writeJSON(w, http.StatusOK, h.buildSignalsData(stored.Model))
}

// buildSignalsData computes and returns the full signals view data for a model.
// Extracted so it can be reused by the insights endpoint for AI context.
func (h *Handler) buildSignalsData(m *entity.UNMModel) signalsResponse {
	// Run all needed analyzers.
	vcReport := h.valueChain.Analyze(m)
	vsReport := h.valueStream.Analyze(m)
	clReport := h.cognitiveLoad.Analyze(m)
	bnReport := h.bottleneck.Analyze(m)
	fragReport := h.fragmentation.Analyze(m)
	ulReport := h.unlinked.Analyze(m)

	// --- UX Risk layer ---
	var needs3Plus, needsUnbacked, needsAtRisk []signalNeedRisk
	for _, nr := range vcReport.NeedRisks {
		item := signalNeedRisk{
			NeedName:  nr.NeedName,
			ActorName: nr.ActorName,
			TeamSpan:  nr.TeamSpan,
			Teams:     nr.Teams,
		}
		switch {
		case nr.Unbacked:
			needsUnbacked = append(needsUnbacked, item)
		case nr.TeamSpan >= h.cfg.Analysis.Signals.NeedTeamSpanCritical:
			needs3Plus = append(needs3Plus, item)
		case nr.AtRisk:
			needsAtRisk = append(needsAtRisk, item)
		}
	}

	// --- Architecture Risk layer ---
	// User-facing caps with cross-team services (fragmented user-facing caps).
	var crossTeamUserFacing []signalCapItem
	for _, fc := range fragReport.FragmentedCapabilities {
		if fc.Capability.Visibility != entity.CapVisibilityUserFacing {
			continue
		}
		teams := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teams = append(teams, t.Name)
		}
		crossTeamUserFacing = append(crossTeamUserFacing, signalCapItem{
			CapabilityName: fc.Capability.Name,
			Visibility:     fc.Capability.Visibility,
			TeamCount:      len(fc.Teams),
			Teams:          teams,
		})
	}

	// Caps not connected to any need (unexpected unlinked = not infrastructure).
	var unlinkedCaps []signalCapItem
	for _, uc := range ulReport.UnlinkedLeafCapabilities {
		if uc.IsExpected {
			continue
		}
		unlinkedCaps = append(unlinkedCaps, signalCapItem{
			CapabilityName: uc.Capability.Name,
			Visibility:     uc.Visibility,
		})
	}

	// All fragmented caps (cross-team ownership).
	var fragCaps []signalCapItem
	for _, fc := range fragReport.FragmentedCapabilities {
		teams := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teams = append(teams, t.Name)
		}
		fragCaps = append(fragCaps, signalCapItem{
			CapabilityName: fc.Capability.Name,
			Visibility:     fc.Capability.Visibility,
			TeamCount:      len(fc.Teams),
			Teams:          teams,
		})
	}

	// --- Org Risk layer ---
	// Top teams by structural load — include all high teams, then top 5 medium.
	var topTeams []signalTeamItem

	// Build coherence lookup by team name.
	coherenceByTeam := make(map[string]float64)
	for _, tc := range vsReport.TeamCoherences {
		coherenceByTeam[tc.TeamName] = tc.CoherenceScore
	}

	for _, tl := range clReport.TeamLoads {
		if tl.OverallLevel != analyzer.LoadHigh {
			continue
		}
		topTeams = append(topTeams, signalTeamItem{
			TeamName:        tl.Team.Name,
			TeamType:        string(tl.Team.TeamType),
			OverallLevel:    string(tl.OverallLevel),
			CapabilityCount: tl.CapabilityCount,
			ServiceCount:    tl.ServiceCount,
			CoherenceScore:  coherenceByTeam[tl.Team.Name],
		})
	}
	// If fewer than 5, pad with medium teams.
	if len(topTeams) < 5 {
		for _, tl := range clReport.TeamLoads {
			if tl.OverallLevel != analyzer.LoadMedium {
				continue
			}
			topTeams = append(topTeams, signalTeamItem{
				TeamName:        tl.Team.Name,
				TeamType:        string(tl.Team.TeamType),
				OverallLevel:    string(tl.OverallLevel),
				CapabilityCount: tl.CapabilityCount,
				ServiceCount:    tl.ServiceCount,
				CoherenceScore:  coherenceByTeam[tl.Team.Name],
			})
			if len(topTeams) >= 5 {
				break
			}
		}
	}

	// Critical bottleneck services.
	var criticalServices []signalServiceItem
	for _, sbn := range bnReport.ServiceBottlenecks {
		if sbn.IsCritical {
			criticalServices = append(criticalServices, signalServiceItem{
				ServiceName: sbn.Service.Name,
				FanIn:       sbn.FanIn,
			})
		}
	}

	// Low coherence teams.
	var lowCoherenceTeams []signalTeamItem
	for _, tc := range vsReport.TeamCoherences {
		if !tc.LowCoherence {
			continue
		}
		lowCoherenceTeams = append(lowCoherenceTeams, signalTeamItem{
			TeamName:       tc.TeamName,
			CoherenceScore: tc.CoherenceScore,
		})
	}

	// Critical external dependency fan-in (deps used by >= 3 services).
	var criticalExtDeps []SignalsExtDepItem
	for _, extDep := range bnReport.ExternalDependencyBottlenecks {
		criticalExtDeps = append(criticalExtDeps, SignalsExtDepItem{
			DepName:      extDep.Name,
			ServiceCount: extDep.ServiceCount,
			Services:     extDep.Services,
			IsCritical:   extDep.IsCritical,
			IsWarning:    extDep.IsWarning,
		})
	}

	// --- Health assessment ---
	uxRisk := healthLevel(len(needs3Plus) > 0 || len(needsUnbacked) > 0, len(needsAtRisk) > 0)
	archRisk := healthLevel(len(crossTeamUserFacing) > 0 && len(criticalServices) > 0,
		len(crossTeamUserFacing) > 0 || len(fragCaps) > 0 || len(unlinkedCaps) > 0)
	// Org risk: red if any high-load teams, amber if low-coherence teams exist.
	orgRisk := healthLevel(anyHighLoad(clReport.TeamLoads), len(lowCoherenceTeams) > 0)

	return signalsResponse{
		ViewType: "signals",
		Health: signalHealth{
			UXRisk:           uxRisk,
			ArchitectureRisk: archRisk,
			OrgRisk:          orgRisk,
		},
		UserExperienceLayer: signalUXLayer{
			NeedsRequiring3PlusTeams: coalesceNeedRisks(needs3Plus),
			NeedsWithNoCapBacking:    coalesceNeedRisks(needsUnbacked),
			NeedsAtRisk:              coalesceNeedRisks(needsAtRisk),
		},
		ArchitectureLayer: signalArchLayer{
			UserFacingCapsWithCrossTeamServices: coalesceCapItems(crossTeamUserFacing),
			CapabilitiesNotConnectedToAnyNeed:   coalesceCapItems(unlinkedCaps),
			CapabilitiesFragmentedAcrossTeams:   coalesceCapItems(fragCaps),
		},
		OrganizationLayer: signalOrgLayer{
			TopTeamsByStructuralLoad:   coalesceTeamItems(topTeams),
			CriticalBottleneckServices: coalesceSvcItems(criticalServices),
			LowCoherenceTeams:          coalesceTeamItems(lowCoherenceTeams),
			CriticalExternalDeps:       coalesceExtDepItems(criticalExtDeps),
		},
	}
}

// healthLevel returns "red", "amber", or "green" based on two severity flags.
func healthLevel(red, amber bool) string {
	if red {
		return "red"
	}
	if amber {
		return "amber"
	}
	return "green"
}

func anyHighLoad(loads []analyzer.TeamLoad) bool {
	for _, tl := range loads {
		if tl.OverallLevel == analyzer.LoadHigh {
			return true
		}
	}
	return false
}

// coalesce helpers return empty slices (never nil) for clean JSON serialisation.
func coalesceNeedRisks(s []signalNeedRisk) []signalNeedRisk {
	if s == nil {
		return []signalNeedRisk{}
	}
	return s
}

func coalesceCapItems(s []signalCapItem) []signalCapItem {
	if s == nil {
		return []signalCapItem{}
	}
	return s
}

func coalesceTeamItems(s []signalTeamItem) []signalTeamItem {
	if s == nil {
		return []signalTeamItem{}
	}
	return s
}

func coalesceSvcItems(s []signalServiceItem) []signalServiceItem {
	if s == nil {
		return []signalServiceItem{}
	}
	return s
}

func coalesceExtDepItems(s []SignalsExtDepItem) []SignalsExtDepItem {
	if s == nil {
		return []SignalsExtDepItem{}
	}
	return s
}
