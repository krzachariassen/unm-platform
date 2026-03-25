package analyzer

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// NeedDeliveryRisk holds the value-chain traversal result for a single need.
type NeedDeliveryRisk struct {
	NeedName        string   `json:"need_name"`
	ActorName       string   `json:"actor_name"`
	TeamSpan        int      `json:"team_span"`
	Teams           []string `json:"teams"`
	CapabilityCount int      `json:"capability_count"`
	ServiceCount    int      `json:"service_count"`
	CrossTeam       bool     `json:"cross_team"`
	AtRisk          bool     `json:"at_risk"`
	Unbacked        bool     `json:"unbacked"`
}

// ValueChainReport holds the aggregate result of the value chain analysis.
type ValueChainReport struct {
	NeedRisks          []NeedDeliveryRisk `json:"need_risks"`
	CrossTeamNeedCount int                `json:"cross_team_need_count"`
	AtRiskNeedCount    int                `json:"at_risk_need_count"`
	UnbackedNeedCount  int                `json:"unbacked_need_count"`
}

// ValueChainAnalyzer traverses Need → Capability → Service → Team
// to compute delivery risk for each need.
type ValueChainAnalyzer struct {
	cfg entity.ValueChainConfig
}

// NewValueChainAnalyzer constructs a ValueChainAnalyzer.
func NewValueChainAnalyzer(cfg entity.ValueChainConfig) *ValueChainAnalyzer {
	return &ValueChainAnalyzer{cfg: cfg}
}

// Analyze traverses the full UNM value chain for each Need and computes
// delivery risk metrics.
func (a *ValueChainAnalyzer) Analyze(m *entity.UNMModel) ValueChainReport {
	if len(m.Needs) == 0 {
		return ValueChainReport{}
	}

	// Run cognitive load analysis to detect overloaded teams.
	defaultCfg := entity.DefaultConfig().Analysis
	cogReport := NewCognitiveLoadAnalyzer(defaultCfg.CognitiveLoad, defaultCfg.InteractionWeights).Analyze(m)
	highLoadTeams := make(map[string]bool)
	for _, tl := range cogReport.TeamLoads {
		if tl.OverallLevel == LoadHigh {
			highLoadTeams[tl.Team.Name] = true
		}
	}

	// Build service lookup by name for quick access.
	// (m.Services is already a map, but we reference it directly.)

	var risks []NeedDeliveryRisk
	crossCount, atRiskCount, unbackedCount := 0, 0, 0

	for _, need := range m.Needs {
		teamSet := make(map[string]bool)
		serviceSet := make(map[string]bool)
		capCount := 0

		for _, capRel := range need.SupportedBy {
			capName := capRel.TargetID.String()
			cap, ok := m.Capabilities[capName]
			if !ok {
				continue
			}
			capCount++

			for _, svcRel := range cap.RealizedBy {
				svcName := svcRel.TargetID.String()
				svc, ok := m.Services[svcName]
				if !ok {
					continue
				}
				serviceSet[svcName] = true
				if svc.OwnerTeamName != "" {
					teamSet[svc.OwnerTeamName] = true
				}
			}
		}

		teams := make([]string, 0, len(teamSet))
		for t := range teamSet {
			teams = append(teams, t)
		}
		sort.Strings(teams)

		teamSpan := len(teams)
		crossTeam := teamSpan > 1
		unbacked := capCount == 0

		// AtRisk: team span >= configured threshold OR any team in chain has high cognitive load.
		atRisk := teamSpan >= a.cfg.AtRiskTeamSpan
		if !atRisk {
			for _, t := range teams {
				if highLoadTeams[t] {
					atRisk = true
					break
				}
			}
		}

		risks = append(risks, NeedDeliveryRisk{
			NeedName:        need.Name,
			ActorName:       need.ActorName,
			TeamSpan:        teamSpan,
			Teams:           teams,
			CapabilityCount: capCount,
			ServiceCount:    len(serviceSet),
			CrossTeam:       crossTeam,
			AtRisk:          atRisk,
			Unbacked:        unbacked,
		})

		if crossTeam {
			crossCount++
		}
		if atRisk {
			atRiskCount++
		}
		if unbacked {
			unbackedCount++
		}
	}

	// Sort by need name for deterministic output.
	sort.Slice(risks, func(i, j int) bool {
		return risks[i].NeedName < risks[j].NeedName
	})

	return ValueChainReport{
		NeedRisks:          risks,
		CrossTeamNeedCount: crossCount,
		AtRiskNeedCount:    atRiskCount,
		UnbackedNeedCount:  unbackedCount,
	}
}
