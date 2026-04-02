package handler

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ── 3. Enriched Ownership View ────────────────────────────────────────────────

// extDepRef is a summary of an external dependency referenced in a view.
type extDepRef struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	ServiceCount int    `json:"service_count"`
}

type enrichedOwnershipResponse struct {
	ViewType                string             `json:"view_type"`
	Lanes                   []ownershipLane    `json:"lanes"`
	UnownedCapabilities     []capRef           `json:"unowned_capabilities"`
	ServiceRows             []serviceRow       `json:"service_rows"`
	CrossTeamCapabilities   []crossTeamCap     `json:"cross_team_capabilities"`
	HighSpanServices        []highSpanService  `json:"high_span_services"`
	OverloadedTeams         []teamRef          `json:"overloaded_teams"`
	NoCapCount              int                `json:"no_cap_count"`
	MultiCapCount           int                `json:"multi_cap_count"`
	ExternalDependencyCount int                `json:"external_dependency_count"`
}

type ownershipLane struct {
	Team         teamNodeRef `json:"team"`
	Caps         []capGroup  `json:"caps"`
	ExternalDeps []extDepRef `json:"external_deps"`
}

type capGroup struct {
	Cap       capRef      `json:"cap"`
	Services  []svcInLane `json:"services"`
	CrossTeam bool        `json:"cross_team"`
}

type svcInLane struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	TeamID    string `json:"team_id"`
	TeamLabel string `json:"team_label"`
	CapCount  int    `json:"cap_count"`
}

type crossTeamCap struct {
	CapID      string   `json:"cap_id"`
	CapLabel   string   `json:"cap_label"`
	TeamLabels []string `json:"team_labels"`
}

func buildEnrichedOwnershipView(m *entity.UNMModel, cfg ...entity.AnalysisConfig) enrichedOwnershipResponse {
	acfg := defaultAnalysisCfg(cfg)
	capSvcMap := buildCapServiceMap(m)
	svcCapCount := buildSvcCapCount(m)

	// Build lanes: team → capabilities it owns → services per cap
	teamNames := make([]string, 0, len(m.Teams))
	for name := range m.Teams {
		teamNames = append(teamNames, name)
	}
	sort.Strings(teamNames)

	lanes := make([]ownershipLane, 0, len(m.Teams))
	overloaded := make([]teamRef, 0)
	ownedCaps := make(map[string]bool)

	for _, teamName := range teamNames {
		t := m.Teams[teamName]
		aps := detectTeamAntiPatterns(t, acfg.OverloadedCapabilityThreshold)
		teamData := map[string]any{
			"type":          t.TeamType.String(),
			"is_overloaded": t.IsOverloaded(acfg.OverloadedCapabilityThreshold),
			"description":   t.Description,
		}
		if len(aps) > 0 {
			teamData["anti_patterns"] = aps
		}

		if t.IsOverloaded(acfg.OverloadedCapabilityThreshold) {
			overloaded = append(overloaded, teamRef{ID: "team-" + t.Name, Label: t.Name})
		}

		// Sort owned capabilities
		ownedCapNames := make([]string, 0, len(t.Owns))
		for _, rel := range t.Owns {
			ownedCapNames = append(ownedCapNames, rel.TargetID.String())
		}
		sort.Strings(ownedCapNames)

		capGroups := make([]capGroup, 0, len(ownedCapNames))
		for _, capName := range ownedCapNames {
			ownedCaps[capName] = true
			c, ok := m.Capabilities[capName]
			if !ok {
				continue
			}
			info := capSvcMap[capName]

			svcsInLane := make([]svcInLane, 0, len(info.services))
			crossTeam := false
			for _, svc := range info.services {
				svcTeamName := svc.OwnerTeamName
				if svcTeamName != teamName && svcTeamName != "" {
					crossTeam = true
				}
				svcTeamLabel := svcTeamName
				svcsInLane = append(svcsInLane, svcInLane{
					ID:        "svc-" + svc.Name,
					Label:     svc.Name,
					TeamID:    "team-" + svcTeamName,
					TeamLabel: svcTeamLabel,
					CapCount:  svcCapCount[svc.Name],
				})
			}

			capData := map[string]any{
				"visibility":  c.Visibility,
				"is_leaf":     c.IsLeaf(),
				"description": c.Description,
			}
			capAps := detectCapabilityAntiPatterns(c, info.isFragmented, len(info.services) > 0)
			if len(capAps) > 0 {
				capData["anti_patterns"] = capAps
			}

			capGroups = append(capGroups, capGroup{
				Cap: capRef{
					ID:    "cap-" + c.Name,
					Label: c.Name,
					Data:  capData,
				},
				Services:  svcsInLane,
				CrossTeam: crossTeam,
			})
		}

		// Build set of service names owned by this team
		teamSvcNames := make(map[string]bool)
		for _, svc := range m.Services {
			if svc.OwnerTeamName == teamName {
				teamSvcNames[svc.Name] = true
			}
		}

		// Find external deps used by any of this team's services
		extDepNames := make([]string, 0)
		extDepNameSet := make(map[string]bool)
		for depName := range m.ExternalDependencies {
			extDepNameSet[depName] = false // placeholder, set later
		}
		type extDepAccum struct {
			dep   *entity.ExternalDependency
			count int
		}
		extDepAccumMap := make(map[string]*extDepAccum)
		for depName, dep := range m.ExternalDependencies {
			count := 0
			for _, usage := range dep.UsedBy {
				if teamSvcNames[usage.ServiceName] {
					count++
				}
			}
			if count > 0 {
				extDepAccumMap[depName] = &extDepAccum{dep: dep, count: count}
				extDepNames = append(extDepNames, depName)
			}
		}
		sort.Strings(extDepNames)
		laneExtDeps := make([]extDepRef, 0, len(extDepNames))
		for _, depName := range extDepNames {
			acc := extDepAccumMap[depName]
			laneExtDeps = append(laneExtDeps, extDepRef{
				ID:           "extdep-" + acc.dep.Name,
				Label:        acc.dep.Name,
				Description:  acc.dep.Description,
				ServiceCount: acc.count,
			})
		}

		lanes = append(lanes, ownershipLane{
			Team: teamNodeRef{
				ID:    "team-" + t.Name,
				Label: t.Name,
				Data:  teamData,
			},
			Caps:         capGroups,
			ExternalDeps: laneExtDeps,
		})
	}

	// Mark parent capabilities as owned if any of their children are owned (transitive)
	changed := true
	for changed {
		changed = false
		for capName, c := range m.Capabilities {
			if ownedCaps[capName] {
				continue
			}
			for _, child := range c.Children {
				if ownedCaps[child.Name] {
					ownedCaps[capName] = true
					changed = true
					break
				}
			}
		}
	}

	// Unowned capabilities
	unowned := make([]capRef, 0)
	for capName, c := range m.Capabilities {
		if !ownedCaps[capName] {
			unowned = append(unowned, capRef{
				ID:    "cap-" + c.Name,
				Label: c.Name,
				Data:  map[string]any{"visibility": c.Visibility},
			})
		}
	}
	sort.Slice(unowned, func(i, j int) bool { return unowned[i].Label < unowned[j].Label })

	// Cross-team capabilities
	crossTeamCaps := make([]crossTeamCap, 0)
	for capName, info := range capSvcMap {
		if len(info.teams) > 1 {
			var labels []string
			for _, t := range info.teams {
				labels = append(labels, t.Name)
			}
			sort.Strings(labels)
			crossTeamCaps = append(crossTeamCaps, crossTeamCap{
				CapID:      "cap-" + capName,
				CapLabel:   capName,
				TeamLabels: labels,
			})
		}
	}
	sort.Slice(crossTeamCaps, func(i, j int) bool { return crossTeamCaps[i].CapLabel < crossTeamCaps[j].CapLabel })

	// High-span services
	highSpan := make([]highSpanService, 0)
	for svcName, count := range svcCapCount {
		if count >= acfg.Signals.HighSpanServiceThreshold {
			highSpan = append(highSpan, highSpanService{Name: svcName, CapabilityCount: count})
		}
	}
	sort.Slice(highSpan, func(i, j int) bool { return highSpan[i].Name < highSpan[j].Name })

	// Service rows + no_cap_count/multi_cap_count
	rows := buildServiceRows(m, svcCapCount)
	noCapCount := 0
	multiCapCount := 0
	for _, svcName := range sortedServiceNames(m) {
		cc := svcCapCount[svcName]
		if cc == 0 {
			noCapCount++
		}
		if cc > 1 {
			multiCapCount++
		}
	}

	return enrichedOwnershipResponse{
		ViewType:                "ownership",
		Lanes:                   lanes,
		UnownedCapabilities:     unowned,
		ServiceRows:             rows,
		CrossTeamCapabilities:   crossTeamCaps,
		HighSpanServices:        highSpan,
		OverloadedTeams:         overloaded,
		NoCapCount:              noCapCount,
		MultiCapCount:           multiCapCount,
		ExternalDependencyCount: len(m.ExternalDependencies),
	}
}
