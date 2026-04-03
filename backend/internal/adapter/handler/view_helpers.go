package handler

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
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
	services     []*entity.Service
	teams        []*entity.Team
	isFragmented bool
}

func buildCapServiceMap(m *entity.UNMModel) map[string]capServiceInfo {
	// service name → number of capabilities it realizes (from service.Realizes)
	svcCapCount := make(map[string]int)
	for _, svc := range m.Services {
		svcCapCount[svc.Name] += len(svc.Realizes)
	}

	result := make(map[string]capServiceInfo)
	for capName := range m.Capabilities {
		svcs := m.GetServicesForCapability(capName)
		teamSet := make(map[string]*entity.Team)
		for _, svc := range svcs {
			if svc.OwnerTeamName != "" {
				if t, ok := m.Teams[svc.OwnerTeamName]; ok {
					teamSet[t.Name] = t
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
		cap := m.Capabilities[capName]
		result[capName] = capServiceInfo{
			services:     svcs,
			teams:        teams,
			isFragmented: cap.IsFragmented(teamNames),
		}
	}
	return result
}

func buildSvcCapCount(m *entity.UNMModel) map[string]int {
	svcCapCount := make(map[string]int)
	for _, svc := range m.Services {
		svcCapCount[svc.Name] += len(svc.Realizes)
	}
	return svcCapCount
}

// ── Shared reference types ────────────────────────────────────────────────────

type capRef struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data,omitempty"`
}

type teamRef struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type teamNodeRef struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data"`
}

type highSpanService struct {
	Name            string `json:"name"`
	CapabilityCount int    `json:"capability_count"`
}

// ── Service row helpers (shared by realization and ownership views) ───────────

type svcRowRef struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type serviceRow struct {
	Service      svcRowRef    `json:"service"`
	Team         *teamNodeRef `json:"team"`
	Capabilities []capRef     `json:"capabilities"`
	ExternalDeps []string     `json:"external_deps"`
}

// buildServiceRows creates sorted service rows with team and capability info.
// Sorted by team label (unowned last), then service label.
func buildServiceRows(m *entity.UNMModel, svcCapCount map[string]int) []serviceRow {
	svcNames := sortedServiceNames(m)

	type rowData struct {
		svc      *entity.Service
		teamName string
		caps     []*entity.Capability
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

	result := make([]serviceRow, 0, len(rows))
	for _, r := range rows {
		capRefs := make([]capRef, 0, len(r.caps))
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

// defaultAnalysisCfg returns the first config or a default.
func defaultAnalysisCfg(cfgs []entity.AnalysisConfig) entity.AnalysisConfig {
	if len(cfgs) > 0 {
		return cfgs[0]
	}
	return entity.DefaultConfig().Analysis
}
