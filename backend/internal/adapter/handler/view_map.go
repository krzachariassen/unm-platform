package handler

import (
	"fmt"
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

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
		// Use first actor for primary edge; need data includes all actor names.
		primaryActor := ""
		if len(n.ActorNames) > 0 {
			primaryActor = n.ActorNames[0]
		}
		actorID := "actor-" + primaryActor
		nodes = append(nodes, viewNode{
			ID:    needID,
			Type:  "need",
			Label: n.Name,
			Data:  map[string]any{"actor_names": n.ActorNames, "is_mapped": n.IsMapped(), "outcome": n.Outcome},
		})
		// Create an edge for each actor the need belongs to
		for _, actorName := range n.ActorNames {
			aID := "actor-" + actorName
			edges = append(edges, viewEdge{
				ID: nextEdgeID(aID, needID), Source: aID, Target: needID, Label: "has need",
			})
		}
		_ = actorID // primary actor used for backward compat above; edges now done per actor
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
