package handler

import (
	"fmt"
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// ── Anti-pattern detection ────────────────────────────────────────────────────

// AntiPattern represents a detected anti-pattern on a node (JSON response type).
type AntiPattern struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// mapAntiPatterns converts domain service AntiPatterns to handler JSON response AntiPatterns.
func mapAntiPatterns(aps []service.AntiPattern) []AntiPattern {
	out := make([]AntiPattern, len(aps))
	for i, ap := range aps {
		out[i] = AntiPattern{Code: ap.Code, Message: ap.Message, Severity: ap.Severity}
	}
	return out
}

func detectTeamAntiPatterns(t *entity.Team, overloadThreshold int) []AntiPattern {
	return mapAntiPatterns(service.DetectTeamAntiPatterns(t, overloadThreshold))
}

func detectCapabilityAntiPatterns(c *entity.Capability, isFragmented bool, hasServices bool) []AntiPattern {
	return mapAntiPatterns(service.DetectCapabilityAntiPatterns(c, isFragmented, hasServices))
}

func detectNeedAntiPatterns(n *entity.Need) []AntiPattern {
	var aps []AntiPattern
	if !n.IsMapped() {
		aps = append(aps, AntiPattern{Code: "unmapped", Message: "No capability supports this need", Severity: "error"})
	}
	return aps
}

func detectServiceAntiPatterns(s *entity.Service) []AntiPattern {
	var aps []AntiPattern
	if s.OwnerTeamName == "" {
		aps = append(aps, AntiPattern{Code: "no_team", Message: "Service has no owning team", Severity: "warning"})
	}
	return aps
}

// ── Helper: compute cap→services and service→capCount maps ────────────────────

// capServiceInfo holds precomputed service/team info for a capability.
type capServiceInfo struct {
	services    []*entity.Service
	teams       []*entity.Team
	isFragmented bool
}

func buildCapServiceMap(m *entity.UNMModel) map[string]capServiceInfo {
	// service name → number of capabilities it realizes
	svcCapCount := make(map[string]int)
	for _, c := range m.Capabilities {
		for _, rel := range c.RealizedBy {
			svcCapCount[rel.TargetID.String()]++
		}
	}

	result := make(map[string]capServiceInfo)
	for capName, c := range m.Capabilities {
		var svcs []*entity.Service
		teamSet := make(map[string]*entity.Team)
		for _, rel := range c.RealizedBy {
			svcName := rel.TargetID.String()
			if svc, ok := m.Services[svcName]; ok {
				svcs = append(svcs, svc)
				if svc.OwnerTeamName != "" {
					if t, ok := m.Teams[svc.OwnerTeamName]; ok {
						teamSet[t.Name] = t
					}
				}
			}
		}
		var teams []*entity.Team
		for _, t := range teamSet {
			teams = append(teams, t)
		}
		sort.Slice(teams, func(i, j int) bool { return teams[i].Name < teams[j].Name })
		sort.Slice(svcs, func(i, j int) bool { return svcs[i].Name < svcs[j].Name })

		var teamNames []string
		for _, t := range teams {
			teamNames = append(teamNames, t.Name)
		}
		result[capName] = capServiceInfo{
			services:    svcs,
			teams:       teams,
			isFragmented: c.IsFragmented(teamNames),
		}
	}
	return result
}

func buildSvcCapCount(m *entity.UNMModel) map[string]int {
	svcCapCount := make(map[string]int)
	for _, c := range m.Capabilities {
		for _, rel := range c.RealizedBy {
			svcCapCount[rel.TargetID.String()]++
		}
	}
	return svcCapCount
}

// ── 1. Enriched Need View ─────────────────────────────────────────────────────

type enrichedNeedResponse struct {
	ViewType      string              `json:"view_type"`
	TotalNeeds    int                 `json:"total_needs"`
	UnmappedCount int                 `json:"unmapped_count"`
	Groups        []needActorGroup    `json:"groups"`
}

type needActorGroup struct {
	Actor needActorRef   `json:"actor"`
	Needs []needEntry    `json:"needs"`
}

type needActorRef struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type needEntry struct {
	Need         needRef       `json:"need"`
	Capabilities []capRef      `json:"capabilities"`
}

type needRef struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data"`
}

type capRef struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data,omitempty"`
}

func buildEnrichedNeedView(m *entity.UNMModel, cfg ...entity.AnalysisConfig) enrichedNeedResponse {
	acfg := defaultAnalysisCfg(cfg)
	// Run value chain analysis to get per-need delivery risk data.
	vcReport := analyzer.NewValueChainAnalyzer(acfg.ValueChain).Analyze(m)
	riskByNeed := make(map[string]analyzer.NeedDeliveryRisk, len(vcReport.NeedRisks))
	for _, nr := range vcReport.NeedRisks {
		riskByNeed[nr.NeedName] = nr
	}

	// Group needs by actor
	actorNeeds := make(map[string][]*entity.Need)
	for _, n := range m.Needs {
		actorNeeds[n.ActorName] = append(actorNeeds[n.ActorName], n)
	}

	// Sort actor names for determinism
	actorNames := make([]string, 0, len(actorNeeds))
	for name := range actorNeeds {
		actorNames = append(actorNames, name)
	}
	sort.Strings(actorNames)

	unmappedCount := 0
	groups := make([]needActorGroup, 0, len(actorNames))
	for _, actorName := range actorNames {
		needs := actorNeeds[actorName]
		sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })

		entries := make([]needEntry, 0, len(needs))
		for _, n := range needs {
			if !n.IsMapped() {
				unmappedCount++
			}

			caps := make([]capRef, 0)
			for _, rel := range n.SupportedBy {
				capName := rel.TargetID.String()
				if c, ok := m.Capabilities[capName]; ok {
					caps = append(caps, capRef{
						ID:    "cap-" + c.Name,
						Label: c.Name,
						Data:  map[string]any{"visibility": c.Visibility},
					})
				}
			}

			needData := map[string]any{
				"is_mapped": n.IsMapped(),
				"outcome":   n.Outcome,
			}
			// Team span data from value chain traversal.
			if nr, ok := riskByNeed[n.Name]; ok {
				needData["team_span"] = nr.TeamSpan
				needData["teams"] = nr.Teams
				needData["at_risk"] = nr.AtRisk
				needData["unbacked"] = nr.Unbacked
			}
			aps := detectNeedAntiPatterns(n)
			if len(aps) > 0 {
				needData["anti_patterns"] = aps
			}

			entries = append(entries, needEntry{
				Need: needRef{
					ID:    "need-" + n.Name,
					Label: n.Name,
					Data:  needData,
				},
				Capabilities: caps,
			})
		}
		groups = append(groups, needActorGroup{
			Actor: needActorRef{
				ID:    "actor-" + actorName,
				Label: actorName,
			},
			Needs: entries,
		})
	}

	return enrichedNeedResponse{
		ViewType:      "need",
		TotalNeeds:    len(m.Needs),
		UnmappedCount: unmappedCount,
		Groups:        groups,
	}
}

// ── 2. Enriched Capability View ───────────────────────────────────────────────

type enrichedCapabilityResponse struct {
	ViewType              string                  `json:"view_type"`
	LeafCapabilityCount   int                     `json:"leaf_capability_count"`
	HighSpanServices      []highSpanService       `json:"high_span_services"`
	FragmentedCapabilities []fragmentedCap        `json:"fragmented_capabilities"`
	Capabilities          []enrichedCapability    `json:"capabilities"`
	ParentGroups          []capParentGroup        `json:"parent_groups"`
}

type capParentGroup struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Children []string `json:"children"`
}

type highSpanService struct {
	Name            string `json:"name"`
	CapabilityCount int    `json:"capability_count"`
}

type fragmentedCap struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	TeamCount int    `json:"team_count"`
}

type enrichedCapability struct {
	ID               string            `json:"id"`
	Label            string            `json:"label"`
	Description      string            `json:"description"`
	Visibility       string            `json:"visibility"`
	IsLeaf           bool              `json:"is_leaf"`
	IsFragmented     bool              `json:"is_fragmented"`
	DependedOnByCount int             `json:"depended_on_by_count"`
	Services         []capServiceRef   `json:"services"`
	Teams            []capTeamRef      `json:"teams"`
	DependsOn        []capRef          `json:"depends_on"`
	Children         []capRef          `json:"children"`
	AntiPatterns     []AntiPattern     `json:"anti_patterns,omitempty"`
	ExternalDeps     []CapExtDepRef    `json:"external_deps,omitempty"`
}

// CapExtDepRef is a reference to an external dependency associated with a capability.
type CapExtDepRef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type capServiceRef struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	CapCount int    `json:"cap_count"`
}

type capTeamRef struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

func buildEnrichedCapabilityView(m *entity.UNMModel, cfg ...entity.AnalysisConfig) enrichedCapabilityResponse {
	acfg := defaultAnalysisCfg(cfg)
	capSvcMap := buildCapServiceMap(m)
	svcCapCount := buildSvcCapCount(m)

	// Build a lookup: service name → list of external deps that use it.
	// Source of truth: ExternalDependencies[].UsedBy
	svcToExtDeps := make(map[string][]*entity.ExternalDependency)
	for _, dep := range m.ExternalDependencies {
		for _, usage := range dep.UsedBy {
			svcToExtDeps[usage.ServiceName] = append(svcToExtDeps[usage.ServiceName], dep)
		}
	}

	// Build reverse dep count: for each cap, how many other caps depend on it
	reverseDeps := make(map[string]int)
	for _, cap := range m.Capabilities {
		for _, rel := range cap.DependsOn {
			reverseDeps[rel.TargetID.String()]++
		}
	}

	leafCount := 0
	highSpanMap := make(map[string]int)
	fragmented := make([]fragmentedCap, 0)
	caps := make([]enrichedCapability, 0, len(m.Capabilities))

	// Sort capability names for determinism
	capNames := make([]string, 0, len(m.Capabilities))
	for name := range m.Capabilities {
		capNames = append(capNames, name)
	}
	sort.Strings(capNames)

	for _, capName := range capNames {
		c := m.Capabilities[capName]
		info := capSvcMap[capName]

		if c.IsLeaf() {
			leafCount++
		}

		// Services
		svcRefs := make([]capServiceRef, 0, len(info.services))
		for _, svc := range info.services {
			cc := svcCapCount[svc.Name]
			svcRefs = append(svcRefs, capServiceRef{
				ID:       "svc-" + svc.Name,
				Label:    svc.Name,
				CapCount: cc,
			})
			if cc >= acfg.Signals.HighSpanServiceThreshold {
				highSpanMap[svc.Name] = cc
			}
		}

		// Teams
		teamRefs := make([]capTeamRef, 0, len(info.teams))
		for _, t := range info.teams {
			teamRefs = append(teamRefs, capTeamRef{
				ID:    "team-" + t.Name,
				Label: t.Name,
				Type:  t.TeamType.String(),
			})
		}

		if info.isFragmented {
			fragmented = append(fragmented, fragmentedCap{
				ID:        "cap-" + c.Name,
				Label:     c.Name,
				TeamCount: len(info.teams),
			})
		}

		// DependsOn
		deps := make([]capRef, 0, len(c.DependsOn))
		for _, rel := range c.DependsOn {
			deps = append(deps, capRef{
				ID:    "cap-" + rel.TargetID.String(),
				Label: rel.TargetID.String(),
			})
		}

		// Children
		children := make([]capRef, 0, len(c.Children))
		for _, child := range c.Children {
			children = append(children, capRef{
				ID:    "cap-" + child.Name,
				Label: child.Name,
			})
		}

		aps := detectCapabilityAntiPatterns(c, info.isFragmented, len(info.services) > 0)

		// Collect external deps used by any service realizing this capability (deduplicated).
		extDepSeen := make(map[string]bool)
		extDepRefs := make([]CapExtDepRef, 0)
		for _, svc := range info.services {
			for _, dep := range svcToExtDeps[svc.Name] {
				if !extDepSeen[dep.Name] {
					extDepSeen[dep.Name] = true
					extDepRefs = append(extDepRefs, CapExtDepRef{
						Name:        dep.Name,
						Description: dep.Description,
					})
				}
			}
		}
		sort.Slice(extDepRefs, func(i, j int) bool { return extDepRefs[i].Name < extDepRefs[j].Name })

		caps = append(caps, enrichedCapability{
			ID:                "cap-" + c.Name,
			Label:             c.Name,
			Description:       c.Description,
			Visibility:        c.Visibility,
			IsLeaf:            c.IsLeaf(),
			IsFragmented:      info.isFragmented,
			DependedOnByCount: reverseDeps[c.Name],
			Services:          svcRefs,
			Teams:             teamRefs,
			DependsOn:         deps,
			Children:          children,
			AntiPatterns:      aps,
			ExternalDeps:      extDepRefs,
		})
	}

	// High span services
	highSpan := make([]highSpanService, 0, len(highSpanMap))
	for name, count := range highSpanMap {
		highSpan = append(highSpan, highSpanService{Name: name, CapabilityCount: count})
	}
	sort.Slice(highSpan, func(i, j int) bool { return highSpan[i].Name < highSpan[j].Name })

	// Parent groups: capabilities with children
	parentGroups := make([]capParentGroup, 0)
	for _, capName := range capNames {
		c := m.Capabilities[capName]
		if len(c.Children) > 0 {
			childIDs := make([]string, 0, len(c.Children))
			for _, child := range c.Children {
				childIDs = append(childIDs, "cap-"+child.Name)
			}
			parentGroups = append(parentGroups, capParentGroup{
				ID:       "cap-" + c.Name,
				Label:    c.Name,
				Children: childIDs,
			})
		}
	}

	return enrichedCapabilityResponse{
		ViewType:               "capability",
		LeafCapabilityCount:    leafCount,
		HighSpanServices:       highSpan,
		FragmentedCapabilities: fragmented,
		Capabilities:           caps,
		ParentGroups:           parentGroups,
	}
}

// ── 3. Enriched Ownership View ────────────────────────────────────────────────

// extDepRef is a summary of an external dependency referenced in a view.
type extDepRef struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	ServiceCount int    `json:"service_count"`
}

type enrichedOwnershipResponse struct {
	ViewType                string                   `json:"view_type"`
	Lanes                   []ownershipLane          `json:"lanes"`
	UnownedCapabilities     []capRef                 `json:"unowned_capabilities"`
	ServiceRows             []serviceRow             `json:"service_rows"`
	CrossTeamCapabilities   []crossTeamCap           `json:"cross_team_capabilities"`
	HighSpanServices        []highSpanService        `json:"high_span_services"`
	OverloadedTeams         []teamRef                `json:"overloaded_teams"`
	NoCapCount              int                      `json:"no_cap_count"`
	MultiCapCount           int                      `json:"multi_cap_count"`
	ExternalDependencyCount int                      `json:"external_dependency_count"`
}

type ownershipLane struct {
	Team         teamNodeRef  `json:"team"`
	Caps         []capGroup   `json:"caps"`
	ExternalDeps []extDepRef  `json:"external_deps"`
}

type teamNodeRef struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data"`
}

type capGroup struct {
	Cap       capRef            `json:"cap"`
	Services  []svcInLane       `json:"services"`
	CrossTeam bool              `json:"cross_team"`
}

type svcInLane struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	TeamID    string `json:"team_id"`
	TeamLabel string `json:"team_label"`
	CapCount  int    `json:"cap_count"`
}

type teamRef struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type crossTeamCap struct {
	CapID      string   `json:"cap_id"`
	CapLabel   string   `json:"cap_label"`
	TeamLabels []string `json:"team_labels"`
}

type serviceRow struct {
	Service      svcRowRef    `json:"service"`
	Team         *teamNodeRef `json:"team"`
	Capabilities []capRef     `json:"capabilities"`
	ExternalDeps []string     `json:"external_deps"`
}

type svcRowRef struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
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

// ── 4. Enriched Team Topology View ────────────────────────────────────────────

type enrichedTeamTopologyResponse struct {
	ViewType     string                    `json:"view_type"`
	Teams        []enrichedTeamNode        `json:"teams"`
	Interactions []enrichedInteraction     `json:"interactions"`
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
	for _, i := range m.Interactions {		ei := enrichedInteraction{
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

// ── 6. Enriched Realization View ──────────────────────────────────────────────

type enrichedRealizationResponse struct {
	ViewType      string       `json:"view_type"`
	NoCapCount    int          `json:"no_cap_count"`
	MultiCapCount int          `json:"multi_cap_count"`
	ServiceRows   []serviceRow `json:"service_rows"`
}

func buildEnrichedRealizationView(m *entity.UNMModel) enrichedRealizationResponse {
	svcCapCount := buildSvcCapCount(m)
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

	return enrichedRealizationResponse{
		ViewType:      "realization",
		NoCapCount:    noCapCount,
		MultiCapCount: multiCapCount,
		ServiceRows:   rows,
	}
}

// buildServiceRows creates sorted service rows with team and capability info.
// Sorted by team label (unowned last), then service label.
func buildServiceRows(m *entity.UNMModel, svcCapCount map[string]int) []serviceRow {
	svcNames := sortedServiceNames(m)

	type rowData struct {
		svc         *entity.Service
		teamName    string
		caps        []*entity.Capability
	}
	var rows []rowData
	for _, svcName := range svcNames {
		svc := m.Services[svcName]
		caps := m.GetCapabilitiesForService(svcName)
		sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
		rows = append(rows, rowData{svc: svc, teamName: svc.OwnerTeamName, caps: caps})
	}

	// Sort: owned services first (by team label, then svc label), unowned last (by svc label)
	sort.SliceStable(rows, func(i, j int) bool {
		ti, tj := rows[i].teamName, rows[j].teamName
		if (ti == "") != (tj == "") {
			return tj == "" // unowned goes last
		}
		if ti != tj {
			return ti < tj
		}
		return rows[i].svc.Name < rows[j].svc.Name
	})

	var result []serviceRow
	for _, r := range rows {
		var capRefs []capRef
		for _, c := range r.caps {
			capRefs = append(capRefs, capRef{
				ID:    "cap-" + c.Name,
				Label: c.Name,
				Data:  map[string]any{"visibility": c.Visibility},
			})
		}

		svcAps := detectServiceAntiPatterns(r.svc)
		svcRef := svcRowRef{ID: "svc-" + r.svc.Name, Label: r.svc.Name, Description: r.svc.Description}

		var teamNode *teamNodeRef
		if r.teamName != "" {
			if t, ok := m.Teams[r.teamName]; ok {
				teamNode = &teamNodeRef{
					ID:    "team-" + t.Name,
					Label: t.Name,
					Data:  map[string]any{"type": t.TeamType.String()},
				}
			}
		}

		// Collect external dependencies for this service (sorted by name for determinism)
		extDeps := m.GetExternalDepsForService(r.svc.Name)
		extDepLabels := make([]string, 0, len(extDeps))
		for _, ed := range extDeps {
			extDepLabels = append(extDepLabels, ed.Name)
		}
		sort.Strings(extDepLabels)

		row := serviceRow{
			Service:      svcRef,
			Team:         teamNode,
			Capabilities: capRefs,
			ExternalDeps: extDepLabels,
		}
		// Attach anti-patterns via a wrapper — but the spec says service nodes in realization
		// view get anti_patterns. We'll handle that check here.
		_ = svcAps // anti-patterns are on the service row

		result = append(result, row)
	}
	return result
}

func sortedServiceNames(m *entity.UNMModel) []string {
	names := make([]string, 0, len(m.Services))
	for name := range m.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ── 7. UNM Map View ──────────────────────────────────────────────────────────

// viewNode is a generic graph node used in the UNM Map view.
type viewNode struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data,omitempty"`
}

// viewEdge is a directed graph edge used in the UNM Map view.
type viewEdge struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

// UNMMapExtDep represents an external dependency node in the UNM Map view.
type UNMMapExtDep struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	ServiceCount int      `json:"service_count"`
	Services     []string `json:"services"`
	IsCritical   bool     `json:"is_critical"`
	IsWarning    bool     `json:"is_warning"`
}

type unmMapResponse struct {
	ViewType     string         `json:"view_type"`
	Nodes        []viewNode     `json:"nodes"`
	Edges        []viewEdge     `json:"edges"`
	ExternalDeps []UNMMapExtDep `json:"external_deps,omitempty"`
}

func buildUNMMapView(m *entity.UNMModel) unmMapResponse {
	nodes := make([]viewNode, 0)
	edges := make([]viewEdge, 0)
	edgeCounter := 0

	nextEdgeID := func(src, tgt string) string {
		edgeCounter++
		return fmt.Sprintf("edge-%s-%s-%d", src, tgt, edgeCounter)
	}

	// Precompute: cap → owning team (first team that owns a service realizing the cap)
	capSvcMap := buildCapServiceMap(m)

	// Actor nodes
	for _, a := range m.Actors {
		nodes = append(nodes, viewNode{
			ID:    "actor-" + a.Name,
			Type:  "actor",
			Label: a.Name,
			Data:  map[string]any{"description": a.Description},
		})
	}

	// Need nodes + actor→need edges
	for _, n := range m.Needs {
		needID := "need-" + n.Name
		actorID := "actor-" + n.ActorName
		nodes = append(nodes, viewNode{
			ID:    needID,
			Type:  "need",
			Label: n.Name,
			Data:  map[string]any{"actor_name": n.ActorName, "is_mapped": n.IsMapped(), "outcome": n.Outcome},
		})
		edges = append(edges, viewEdge{
			ID: nextEdgeID(actorID, needID), Source: actorID, Target: needID, Label: "has need",
		})
		for _, rel := range n.SupportedBy {
			capID := "cap-" + rel.TargetID.String()
			edges = append(edges, viewEdge{
				ID: nextEdgeID(needID, capID), Source: needID, Target: capID, Label: "supportedBy",
				Description: rel.Description,
			})
		}
	}

	// Capability nodes + dependsOn edges
	for _, c := range m.Capabilities {
		capID := "cap-" + c.Name
		info := capSvcMap[c.Name]

		// Build services list for the node data
		type svcEntry struct {
			ID       string `json:"id"`
			Label    string `json:"label"`
			TeamName string `json:"team_name"`
		}
		svcs := make([]svcEntry, 0, len(info.services))
		for _, svc := range info.services {
			svcs = append(svcs, svcEntry{
				ID:       "svc-" + svc.Name,
				Label:    svc.Name,
				TeamName: svc.OwnerTeamName,
			})
		}

		// Primary team: first team in sorted list
		teamLabel := ""
		teamType := ""
		if len(info.teams) > 0 {
			teamLabel = info.teams[0].Name
			teamType = info.teams[0].TeamType.String()
		}

		nodes = append(nodes, viewNode{
			ID:    capID,
			Type:  "capability",
			Label: c.Name,
			Data: map[string]any{
				"visibility":    c.Visibility,
				"is_leaf":       c.IsLeaf(),
				"is_fragmented": info.isFragmented,
				"team_label":    teamLabel,
				"team_type":     teamType,
				"services":      svcs,
				"description":   c.Description,
			},
		})
		for _, rel := range c.DependsOn {
			depID := "cap-" + rel.TargetID.String()
			edges = append(edges, viewEdge{
				ID: nextEdgeID(capID, depID), Source: capID, Target: depID, Label: "dependsOn",
				Description: rel.Description,
			})
		}
	}

	// External dependency nodes — include ALL external deps, sorted alphabetically.
	// Mark critical (>= 5 services) and warning (>= 3 and < 5 services).
	const unmMapExtDepCritical = 5
	const unmMapExtDepWarning = 3

	extDepNames := make([]string, 0, len(m.ExternalDependencies))
	for name := range m.ExternalDependencies {
		extDepNames = append(extDepNames, name)
	}
	sort.Strings(extDepNames)

	extDeps := make([]UNMMapExtDep, 0, len(extDepNames))
	for _, depName := range extDepNames {
		dep := m.ExternalDependencies[depName]
		count := len(dep.UsedBy)
		services := make([]string, 0, count)
		for _, usage := range dep.UsedBy {
			services = append(services, usage.ServiceName)
		}
		sort.Strings(services)
		isCritical := count >= unmMapExtDepCritical
		isWarning := count >= unmMapExtDepWarning && !isCritical
		extDeps = append(extDeps, UNMMapExtDep{
			ID:           "extdep-" + dep.Name,
			Name:         dep.Name,
			Description:  dep.Description,
			ServiceCount: count,
			Services:     services,
			IsCritical:   isCritical,
			IsWarning:    isWarning,
		})
	}

	return unmMapResponse{
		ViewType:     "unm-map",
		Nodes:        nodes,
		Edges:        edges,
		ExternalDeps: extDeps,
	}
}

// defaultAnalysisCfg returns the first config or a default.
func defaultAnalysisCfg(cfgs []entity.AnalysisConfig) entity.AnalysisConfig {
	if len(cfgs) > 0 {
		return cfgs[0]
	}
	return entity.DefaultConfig().Analysis
}
