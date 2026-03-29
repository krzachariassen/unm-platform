package usecase

import (
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// SignalHealth is the three-layer traffic-light summary.
type SignalHealth struct {
	UXRisk           string `json:"ux_risk"`
	ArchitectureRisk string `json:"architecture_risk"`
	OrgRisk          string `json:"org_risk"`
}

// SignalNeedRisk holds per-need risk data.
type SignalNeedRisk struct {
	NeedName  string   `json:"need_name"`
	ActorName string   `json:"actor_name"`
	TeamSpan  int      `json:"team_span"`
	Teams     []string `json:"teams"`
}

// SignalCapItem holds per-capability signal data.
type SignalCapItem struct {
	CapabilityName string   `json:"capability_name"`
	Visibility     string   `json:"visibility,omitempty"`
	TeamCount      int      `json:"team_count,omitempty"`
	Teams          []string `json:"teams,omitempty"`
}

// SignalTeamItem holds per-team signal data.
type SignalTeamItem struct {
	TeamName        string  `json:"team_name"`
	TeamType        string  `json:"team_type"`
	OverallLevel    string  `json:"overall_level"`
	CapabilityCount int     `json:"capability_count"`
	ServiceCount    int     `json:"service_count"`
	CoherenceScore  float64 `json:"coherence_score,omitempty"`
}

// SignalServiceItem holds per-service signal data.
type SignalServiceItem struct {
	ServiceName string `json:"service_name"`
	FanIn       int    `json:"fan_in"`
}

// SignalExtDepItem summarises an external dependency with notable service fan-in.
type SignalExtDepItem struct {
	DepName      string   `json:"dep_name"`
	ServiceCount int      `json:"service_count"`
	Services     []string `json:"services"`
	IsCritical   bool     `json:"is_critical"`
	IsWarning    bool     `json:"is_warning"`
}

// SignalUXLayer groups UX-risk signals.
type SignalUXLayer struct {
	NeedsRequiring3PlusTeams []SignalNeedRisk `json:"needs_requiring_3plus_teams"`
	NeedsWithNoCapBacking    []SignalNeedRisk `json:"needs_with_no_capability_backing"`
	NeedsAtRisk              []SignalNeedRisk `json:"needs_at_risk"`
}

// SignalArchLayer groups architecture-risk signals.
type SignalArchLayer struct {
	UserFacingCapsWithCrossTeamServices []SignalCapItem `json:"user_facing_caps_with_cross_team_services"`
	CapabilitiesNotConnectedToAnyNeed   []SignalCapItem `json:"capabilities_not_connected_to_any_need"`
	CapabilitiesFragmentedAcrossTeams   []SignalCapItem `json:"capabilities_fragmented_across_teams"`
}

// SignalOrgLayer groups org-risk signals.
type SignalOrgLayer struct {
	TopTeamsByStructuralLoad   []SignalTeamItem    `json:"top_teams_by_structural_load"`
	CriticalBottleneckServices []SignalServiceItem  `json:"critical_bottleneck_services"`
	LowCoherenceTeams          []SignalTeamItem    `json:"low_coherence_teams"`
	CriticalExternalDeps       []SignalExtDepItem   `json:"critical_external_deps"`
}

// SignalsResponse is the full signals view payload.
type SignalsResponse struct {
	ViewType            string          `json:"view_type"`
	Health              SignalHealth     `json:"health"`
	UserExperienceLayer SignalUXLayer    `json:"user_experience_layer"`
	ArchitectureLayer   SignalArchLayer  `json:"architecture_layer"`
	OrganizationLayer   SignalOrgLayer   `json:"organization_layer"`
}

// SignalsService computes the signals view data from a UNM model.
type SignalsService struct {
	valueChain    *analyzer.ValueChainAnalyzer
	valueStream   *analyzer.ValueStreamAnalyzer
	cognitiveLoad *analyzer.CognitiveLoadAnalyzer
	bottleneck    *analyzer.BottleneckAnalyzer
	fragmentation *analyzer.FragmentationAnalyzer
	unlinked      *analyzer.UnlinkedCapabilityAnalyzer
	signalsCfg    entity.SignalsConfig
}

// NewSignalsService constructs a SignalsService.
func NewSignalsService(
	vc *analyzer.ValueChainAnalyzer,
	vs *analyzer.ValueStreamAnalyzer,
	cl *analyzer.CognitiveLoadAnalyzer,
	bn *analyzer.BottleneckAnalyzer,
	frag *analyzer.FragmentationAnalyzer,
	unl *analyzer.UnlinkedCapabilityAnalyzer,
	signalsCfg entity.SignalsConfig,
) *SignalsService {
	return &SignalsService{
		valueChain:    vc,
		valueStream:   vs,
		cognitiveLoad: cl,
		bottleneck:    bn,
		fragmentation: frag,
		unlinked:      unl,
		signalsCfg:    signalsCfg,
	}
}

// BuildSignalsData runs all required analyzers and computes the signals view for m.
func (s *SignalsService) BuildSignalsData(m *entity.UNMModel) (SignalsResponse, error) {
	vcReport := s.valueChain.Analyze(m)
	vsReport := s.valueStream.Analyze(m)
	clReport := s.cognitiveLoad.Analyze(m)
	bnReport := s.bottleneck.Analyze(m)
	fragReport := s.fragmentation.Analyze(m)
	ulReport := s.unlinked.Analyze(m)

	// --- UX Risk layer ---
	var needs3Plus, needsUnbacked, needsAtRisk []SignalNeedRisk
	for _, nr := range vcReport.NeedRisks {
		item := SignalNeedRisk{
			NeedName:  nr.NeedName,
			ActorName: nr.ActorName,
			TeamSpan:  nr.TeamSpan,
			Teams:     nr.Teams,
		}
		switch {
		case nr.Unbacked:
			needsUnbacked = append(needsUnbacked, item)
		case nr.TeamSpan >= s.signalsCfg.NeedTeamSpanCritical:
			needs3Plus = append(needs3Plus, item)
		case nr.AtRisk:
			needsAtRisk = append(needsAtRisk, item)
		}
	}

	// --- Architecture Risk layer ---
	var crossTeamUserFacing []SignalCapItem
	for _, fc := range fragReport.FragmentedCapabilities {
		if fc.Capability.Visibility != entity.CapVisibilityUserFacing {
			continue
		}
		teams := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teams = append(teams, t.Name)
		}
		crossTeamUserFacing = append(crossTeamUserFacing, SignalCapItem{
			CapabilityName: fc.Capability.Name,
			Visibility:     fc.Capability.Visibility,
			TeamCount:      len(fc.Teams),
			Teams:          teams,
		})
	}

	var unlinkedCaps []SignalCapItem
	for _, uc := range ulReport.UnlinkedLeafCapabilities {
		if uc.IsExpected {
			continue
		}
		unlinkedCaps = append(unlinkedCaps, SignalCapItem{
			CapabilityName: uc.Capability.Name,
			Visibility:     uc.Visibility,
		})
	}

	var fragCaps []SignalCapItem
	for _, fc := range fragReport.FragmentedCapabilities {
		teams := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teams = append(teams, t.Name)
		}
		fragCaps = append(fragCaps, SignalCapItem{
			CapabilityName: fc.Capability.Name,
			Visibility:     fc.Capability.Visibility,
			TeamCount:      len(fc.Teams),
			Teams:          teams,
		})
	}

	// --- Org Risk layer ---
	coherenceByTeam := make(map[string]float64)
	for _, tc := range vsReport.TeamCoherences {
		coherenceByTeam[tc.TeamName] = tc.CoherenceScore
	}

	var topTeams []SignalTeamItem
	for _, tl := range clReport.TeamLoads {
		if tl.OverallLevel != analyzer.LoadHigh {
			continue
		}
		topTeams = append(topTeams, SignalTeamItem{
			TeamName:        tl.Team.Name,
			TeamType:        string(tl.Team.TeamType),
			OverallLevel:    string(tl.OverallLevel),
			CapabilityCount: tl.CapabilityCount,
			ServiceCount:    tl.ServiceCount,
			CoherenceScore:  coherenceByTeam[tl.Team.Name],
		})
	}
	if len(topTeams) < 5 {
		for _, tl := range clReport.TeamLoads {
			if tl.OverallLevel != analyzer.LoadMedium {
				continue
			}
			topTeams = append(topTeams, SignalTeamItem{
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

	var criticalServices []SignalServiceItem
	for _, sbn := range bnReport.ServiceBottlenecks {
		if sbn.IsCritical {
			criticalServices = append(criticalServices, SignalServiceItem{
				ServiceName: sbn.Service.Name,
				FanIn:       sbn.FanIn,
			})
		}
	}

	var lowCoherenceTeams []SignalTeamItem
	for _, tc := range vsReport.TeamCoherences {
		if !tc.LowCoherence {
			continue
		}
		lowCoherenceTeams = append(lowCoherenceTeams, SignalTeamItem{
			TeamName:       tc.TeamName,
			CoherenceScore: tc.CoherenceScore,
		})
	}

	var criticalExtDeps []SignalExtDepItem
	for _, extDep := range bnReport.ExternalDependencyBottlenecks {
		criticalExtDeps = append(criticalExtDeps, SignalExtDepItem{
			DepName:      extDep.Name,
			ServiceCount: extDep.ServiceCount,
			Services:     extDep.Services,
			IsCritical:   extDep.IsCritical,
			IsWarning:    extDep.IsWarning,
		})
	}

	// --- Health assessment ---
	uxRisk := signalHealthLevel(len(needs3Plus) > 0 || len(needsUnbacked) > 0, len(needsAtRisk) > 0)
	archRisk := signalHealthLevel(
		len(crossTeamUserFacing) > 0 && len(criticalServices) > 0,
		len(crossTeamUserFacing) > 0 || len(fragCaps) > 0 || len(unlinkedCaps) > 0,
	)
	orgRisk := signalHealthLevel(anyHighLoad(clReport.TeamLoads), len(lowCoherenceTeams) > 0)

	return SignalsResponse{
		ViewType: "signals",
		Health: SignalHealth{
			UXRisk:           uxRisk,
			ArchitectureRisk: archRisk,
			OrgRisk:          orgRisk,
		},
		UserExperienceLayer: SignalUXLayer{
			NeedsRequiring3PlusTeams: coalesceNeedRisks(needs3Plus),
			NeedsWithNoCapBacking:    coalesceNeedRisks(needsUnbacked),
			NeedsAtRisk:              coalesceNeedRisks(needsAtRisk),
		},
		ArchitectureLayer: SignalArchLayer{
			UserFacingCapsWithCrossTeamServices: coalesceCapItems(crossTeamUserFacing),
			CapabilitiesNotConnectedToAnyNeed:   coalesceCapItems(unlinkedCaps),
			CapabilitiesFragmentedAcrossTeams:   coalesceCapItems(fragCaps),
		},
		OrganizationLayer: SignalOrgLayer{
			TopTeamsByStructuralLoad:   coalesceTeamItems(topTeams),
			CriticalBottleneckServices: coalesceSvcItems(criticalServices),
			LowCoherenceTeams:          coalesceTeamItems(lowCoherenceTeams),
			CriticalExternalDeps:       coalesceExtDepItems(criticalExtDeps),
		},
	}, nil
}

func signalHealthLevel(red, amber bool) string {
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

func coalesceNeedRisks(s []SignalNeedRisk) []SignalNeedRisk {
	if s == nil {
		return []SignalNeedRisk{}
	}
	return s
}

func coalesceCapItems(s []SignalCapItem) []SignalCapItem {
	if s == nil {
		return []SignalCapItem{}
	}
	return s
}

func coalesceTeamItems(s []SignalTeamItem) []SignalTeamItem {
	if s == nil {
		return []SignalTeamItem{}
	}
	return s
}

func coalesceSvcItems(s []SignalServiceItem) []SignalServiceItem {
	if s == nil {
		return []SignalServiceItem{}
	}
	return s
}

func coalesceExtDepItems(s []SignalExtDepItem) []SignalExtDepItem {
	if s == nil {
		return []SignalExtDepItem{}
	}
	return s
}
