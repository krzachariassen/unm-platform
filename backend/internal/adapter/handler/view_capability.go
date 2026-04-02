package handler

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ── 2. Enriched Capability View ───────────────────────────────────────────────

type enrichedCapabilityResponse struct {
	ViewType               string               `json:"view_type"`
	LeafCapabilityCount    int                  `json:"leaf_capability_count"`
	HighSpanServices       []highSpanService    `json:"high_span_services"`
	FragmentedCapabilities []fragmentedCap      `json:"fragmented_capabilities"`
	Capabilities           []enrichedCapability `json:"capabilities"`
	ParentGroups           []capParentGroup     `json:"parent_groups"`
}

type capParentGroup struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Children []string `json:"children"`
}

type fragmentedCap struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	TeamCount int    `json:"team_count"`
}

type enrichedCapability struct {
	ID                string         `json:"id"`
	Label             string         `json:"label"`
	Description       string         `json:"description"`
	Visibility        string         `json:"visibility"`
	IsLeaf            bool           `json:"is_leaf"`
	IsFragmented      bool           `json:"is_fragmented"`
	DependedOnByCount int            `json:"depended_on_by_count"`
	Services          []capServiceRef `json:"services"`
	Teams             []capTeamRef   `json:"teams"`
	DependsOn         []capRef       `json:"depends_on"`
	Children          []capRef       `json:"children"`
	AntiPatterns      []AntiPattern  `json:"anti_patterns,omitempty"`
	ExternalDeps      []CapExtDepRef `json:"external_deps,omitempty"`
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
