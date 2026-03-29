package handler

import (
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func (h *Handler) registerSignalsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models/{id}/views/signals", h.handleSignals)
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

// buildSignalsData delegates to the SignalsService use case and returns the signals view data.
// Extracted so it can be reused by the insights endpoint for AI context.
func (h *Handler) buildSignalsData(m *entity.UNMModel) signalsResponse {
	svc := usecase.NewSignalsService(
		h.valueChain,
		h.valueStream,
		h.cognitiveLoad,
		h.bottleneck,
		h.fragmentation,
		h.unlinked,
		h.cfg.Analysis.Signals,
	)
	// On error (shouldn't happen with valid model), return empty response.
	result, err := svc.BuildSignalsData(m)
	if err != nil {
		return signalsResponse{ViewType: "signals"}
	}
	return mapToSignalsResponse(result)
}

// mapToSignalsResponse converts the usecase.SignalsResponse to the handler's signalsResponse type.
func mapToSignalsResponse(r usecase.SignalsResponse) signalsResponse {
	return signalsResponse{
		ViewType: r.ViewType,
		Health: signalHealth{
			UXRisk:           r.Health.UXRisk,
			ArchitectureRisk: r.Health.ArchitectureRisk,
			OrgRisk:          r.Health.OrgRisk,
		},
		UserExperienceLayer: signalUXLayer{
			NeedsRequiring3PlusTeams: mapNeedRisks(r.UserExperienceLayer.NeedsRequiring3PlusTeams),
			NeedsWithNoCapBacking:    mapNeedRisks(r.UserExperienceLayer.NeedsWithNoCapBacking),
			NeedsAtRisk:              mapNeedRisks(r.UserExperienceLayer.NeedsAtRisk),
		},
		ArchitectureLayer: signalArchLayer{
			UserFacingCapsWithCrossTeamServices: mapCapItems(r.ArchitectureLayer.UserFacingCapsWithCrossTeamServices),
			CapabilitiesNotConnectedToAnyNeed:   mapCapItems(r.ArchitectureLayer.CapabilitiesNotConnectedToAnyNeed),
			CapabilitiesFragmentedAcrossTeams:   mapCapItems(r.ArchitectureLayer.CapabilitiesFragmentedAcrossTeams),
		},
		OrganizationLayer: signalOrgLayer{
			TopTeamsByStructuralLoad:   mapTeamItems(r.OrganizationLayer.TopTeamsByStructuralLoad),
			CriticalBottleneckServices: mapSvcItems(r.OrganizationLayer.CriticalBottleneckServices),
			LowCoherenceTeams:          mapTeamItems(r.OrganizationLayer.LowCoherenceTeams),
			CriticalExternalDeps:       mapExtDepItems(r.OrganizationLayer.CriticalExternalDeps),
		},
	}
}

func mapNeedRisks(items []usecase.SignalNeedRisk) []signalNeedRisk {
	out := make([]signalNeedRisk, len(items))
	for i, x := range items {
		out[i] = signalNeedRisk{NeedName: x.NeedName, ActorName: x.ActorName, TeamSpan: x.TeamSpan, Teams: x.Teams}
	}
	return out
}

func mapCapItems(items []usecase.SignalCapItem) []signalCapItem {
	out := make([]signalCapItem, len(items))
	for i, x := range items {
		out[i] = signalCapItem{CapabilityName: x.CapabilityName, Visibility: x.Visibility, TeamCount: x.TeamCount, Teams: x.Teams}
	}
	return out
}

func mapTeamItems(items []usecase.SignalTeamItem) []signalTeamItem {
	out := make([]signalTeamItem, len(items))
	for i, x := range items {
		out[i] = signalTeamItem{
			TeamName:        x.TeamName,
			TeamType:        x.TeamType,
			OverallLevel:    x.OverallLevel,
			CapabilityCount: x.CapabilityCount,
			ServiceCount:    x.ServiceCount,
			CoherenceScore:  x.CoherenceScore,
		}
	}
	return out
}

func mapSvcItems(items []usecase.SignalServiceItem) []signalServiceItem {
	out := make([]signalServiceItem, len(items))
	for i, x := range items {
		out[i] = signalServiceItem{ServiceName: x.ServiceName, FanIn: x.FanIn}
	}
	return out
}

func mapExtDepItems(items []usecase.SignalExtDepItem) []SignalsExtDepItem {
	out := make([]SignalsExtDepItem, len(items))
	for i, x := range items {
		out[i] = SignalsExtDepItem{
			DepName:      x.DepName,
			ServiceCount: x.ServiceCount,
			Services:     x.Services,
			IsCritical:   x.IsCritical,
			IsWarning:    x.IsWarning,
		}
	}
	return out
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
