package handler

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// ── 4. Enriched Team Topology View ────────────────────────────────────────────

type enrichedTeamTopologyResponse struct {
	ViewType     string                `json:"view_type"`
	Teams        []enrichedTeamNode    `json:"teams"`
	Interactions []enrichedInteraction `json:"interactions"`
}

type enrichedTeamNode struct {
	ID              string                `json:"id"`
	Label           string                `json:"label"`
	Description     string                `json:"description"`
	Type            string                `json:"type"`
	IsOverloaded    bool                  `json:"is_overloaded"`
	CapabilityCount int                   `json:"capability_count"`
	ServiceCount    int                   `json:"service_count"`
	Services        []string              `json:"services"`
	Capabilities    []string              `json:"capabilities"`
	Interactions    []enrichedInteraction `json:"interactions"`
	AntiPatterns    []AntiPattern         `json:"anti_patterns,omitempty"`
}

type enrichedInteraction struct {
	SourceID    string `json:"source_id"`
	TargetID    string `json:"target_id"`
	Mode        string `json:"mode"`
	Via         string `json:"via"`
	Description string `json:"description"`
}

func buildEnrichedTeamTopologyView(m *entity.UNMModel, cfg ...entity.AnalysisConfig) enrichedTeamTopologyResponse {
	acfg := defaultAnalysisCfg(cfg)
	// Count services per team and collect service names
	svcCountByTeam := make(map[string]int)
	svcNamesByTeam := make(map[string][]string)
	for _, svc := range m.Services {
		if svc.OwnerTeamName != "" {
			svcCountByTeam[svc.OwnerTeamName]++
			svcNamesByTeam[svc.OwnerTeamName] = append(svcNamesByTeam[svc.OwnerTeamName], svc.Name)
		}
	}
	for _, names := range svcNamesByTeam {
		sort.Strings(names)
	}

	// Build interaction lookup: team → interactions involving it
	teamInteractions := make(map[string][]enrichedInteraction)
	allInteractions := make([]enrichedInteraction, 0, len(m.Interactions))
	for _, i := range m.Interactions {
		ei := enrichedInteraction{
			SourceID:    "team-" + i.FromTeamName,
			TargetID:    "team-" + i.ToTeamName,
			Mode:        i.Mode.String(),
			Via:         i.Via,
			Description: i.Description,
		}
		allInteractions = append(allInteractions, ei)
		teamInteractions[i.FromTeamName] = append(teamInteractions[i.FromTeamName], ei)
		teamInteractions[i.ToTeamName] = append(teamInteractions[i.ToTeamName], ei)
	}

	// Sort team names for determinism
	teamNames := make([]string, 0, len(m.Teams))
	for name := range m.Teams {
		teamNames = append(teamNames, name)
	}
	sort.Strings(teamNames)

	teams := make([]enrichedTeamNode, 0, len(m.Teams))
	for _, name := range teamNames {
		t := m.Teams[name]
		aps := detectTeamAntiPatterns(t, acfg.OverloadedCapabilityThreshold)
		ints := teamInteractions[t.Name]
		if ints == nil {
			ints = make([]enrichedInteraction, 0)
		}

		// Collect capability names from team's Owns relationships
		capNames := make([]string, 0, len(t.Owns))
		for _, rel := range t.Owns {
			capNames = append(capNames, rel.TargetID.String())
		}
		sort.Strings(capNames)

		// Collect service names for this team
		svcNames := svcNamesByTeam[t.Name]
		if svcNames == nil {
			svcNames = make([]string, 0)
		}

		teams = append(teams, enrichedTeamNode{
			ID:              "team-" + t.Name,
			Label:           t.Name,
			Description:     t.Description,
			Type:            t.TeamType.String(),
			IsOverloaded:    t.IsOverloaded(acfg.OverloadedCapabilityThreshold),
			CapabilityCount: len(t.Owns),
			ServiceCount:    svcCountByTeam[t.Name],
			Services:        svcNames,
			Capabilities:    capNames,
			Interactions:    ints,
			AntiPatterns:    aps,
		})
	}

	return enrichedTeamTopologyResponse{
		ViewType:     "team-topology",
		Teams:        teams,
		Interactions: allInteractions,
	}
}

// ── 5. Enriched Cognitive Load View ───────────────────────────────────────────

type enrichedCognitiveLoadResponse struct {
	ViewType  string             `json:"view_type"`
	TeamLoads []enrichedTeamLoad `json:"team_loads"`
}

type loadDimensionView struct {
	Value float64 `json:"value"`
	Level string  `json:"level"`
}

type enrichedTeamLoad struct {
	Team             teamLoadRef       `json:"team"`
	CapabilityCount  int               `json:"capability_count"`
	ServiceCount     int               `json:"service_count"`
	DependencyCount  int               `json:"dependency_count"`
	InteractionCount int               `json:"interaction_count"`
	InteractionScore int               `json:"interaction_score"`
	TeamSize         int               `json:"team_size"`
	SizeIsExplicit   bool              `json:"size_is_explicit"`
	DomainSpread     loadDimensionView `json:"domain_spread"`
	ServiceLoad      loadDimensionView `json:"service_load"`
	InteractionLoad  loadDimensionView `json:"interaction_load"`
	DependencyLoad   loadDimensionView `json:"dependency_load"`
	OverallLevel     string            `json:"overall_level"`
	Services         []string          `json:"services"`
	Capabilities     []string          `json:"capabilities"`
}

type teamLoadRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func dimView(d analyzer.LoadDimension) loadDimensionView {
	return loadDimensionView{Value: d.Value, Level: string(d.Level)}
}

func buildEnrichedCognitiveLoadView(m *entity.UNMModel, cfg ...entity.AnalysisConfig) enrichedCognitiveLoadResponse {
	acfg := defaultAnalysisCfg(cfg)
	cl := analyzer.NewCognitiveLoadAnalyzer(acfg.CognitiveLoad, acfg.InteractionWeights)
	report := cl.Analyze(m)

	svcNamesByTeam := make(map[string][]string)
	for _, svc := range m.Services {
		if svc.OwnerTeamName != "" {
			svcNamesByTeam[svc.OwnerTeamName] = append(svcNamesByTeam[svc.OwnerTeamName], svc.Name)
		}
	}
	for _, names := range svcNamesByTeam {
		sort.Strings(names)
	}

	loads := make([]enrichedTeamLoad, 0, len(report.TeamLoads))
	for _, tl := range report.TeamLoads {
		capNames := make([]string, 0, len(tl.Team.Owns))
		for _, rel := range tl.Team.Owns {
			capNames = append(capNames, rel.TargetID.String())
		}
		sort.Strings(capNames)

		svcNames := svcNamesByTeam[tl.Team.Name]
		if svcNames == nil {
			svcNames = make([]string, 0)
		}

		loads = append(loads, enrichedTeamLoad{
			Team: teamLoadRef{
				Name: tl.Team.Name,
				Type: tl.Team.TeamType.String(),
			},
			CapabilityCount:  tl.CapabilityCount,
			ServiceCount:     tl.ServiceCount,
			DependencyCount:  tl.DependencyCount,
			InteractionCount: tl.InteractionCount,
			InteractionScore: tl.InteractionScore,
			TeamSize:         tl.TeamSize,
			SizeIsExplicit:   tl.SizeIsExplicit,
			DomainSpread:     dimView(tl.DomainSpread),
			ServiceLoad:      dimView(tl.ServiceLoad),
			InteractionLoad:  dimView(tl.InteractionLoad),
			DependencyLoad:   dimView(tl.DependencyLoad),
			OverallLevel:     string(tl.OverallLevel),
			Services:         svcNames,
			Capabilities:     capNames,
		})
	}

	return enrichedCognitiveLoadResponse{
		ViewType:  "cognitive-load",
		TeamLoads: loads,
	}
}
