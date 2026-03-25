package usecase

import (
	"github.com/uber/unm-platform/internal/domain/entity"
)

// QueryEngine provides structured queries against a UNMModel.
type QueryEngine struct{}

// NewQueryEngine constructs a QueryEngine.
func NewQueryEngine() *QueryEngine {
	return &QueryEngine{}
}

// CapabilitiesForActor returns all Capabilities reachable from the actor's needs.
// Actor → Needs → SupportedBy → Capabilities (deduplicated).
func (q *QueryEngine) CapabilitiesForActor(m *entity.UNMModel, actorName string) []*entity.Capability {
	seen := make(map[string]bool)
	var result []*entity.Capability
	for _, need := range m.Needs {
		if need.ActorName != actorName {
			continue
		}
		for _, rel := range need.SupportedBy {
			capName := rel.TargetID.String()
			if seen[capName] {
				continue
			}
			if cap, ok := m.Capabilities[capName]; ok {
				seen[capName] = true
				result = append(result, cap)
			}
		}
	}
	return result
}

// ServicesForCapability returns all Services that realize a capability — directly
// (via cap.RealizedBy) and transitively (via cap.DependsOn → other caps' RealizedBy).
// Cycle-safe via visited set.
func (q *QueryEngine) ServicesForCapability(m *entity.UNMModel, capName string) []*entity.Service {
	visitedCaps := make(map[string]bool)
	seenSvcs := make(map[string]bool)
	var result []*entity.Service
	q.collectServices(m, capName, visitedCaps, seenSvcs, &result)
	return result
}

func (q *QueryEngine) collectServices(
	m *entity.UNMModel,
	capName string,
	visitedCaps map[string]bool,
	seenSvcs map[string]bool,
	result *[]*entity.Service,
) {
	if visitedCaps[capName] {
		return
	}
	visitedCaps[capName] = true

	cap, ok := m.Capabilities[capName]
	if !ok {
		return
	}

	// Direct services
	for _, rel := range cap.RealizedBy {
		svcName := rel.TargetID.String()
		if seenSvcs[svcName] {
			continue
		}
		if svc, found := m.Services[svcName]; found {
			seenSvcs[svcName] = true
			*result = append(*result, svc)
		}
	}

	// Transitive via dependsOn
	for _, rel := range cap.DependsOn {
		q.collectServices(m, rel.TargetID.String(), visitedCaps, seenSvcs, result)
	}
}

// TeamsForCapability returns all Teams that own the named capability.
func (q *QueryEngine) TeamsForCapability(m *entity.UNMModel, capName string) []*entity.Team {
	return m.GetTeamsForCapability(capName)
}

// CapabilityDependencyClosure returns all Capabilities reachable via DependsOn from
// the named capability (transitive, excluding the start capability itself). Cycle-safe.
func (q *QueryEngine) CapabilityDependencyClosure(m *entity.UNMModel, capName string) []*entity.Capability {
	visited := make(map[string]bool)
	visited[capName] = true // exclude start
	var result []*entity.Capability
	q.collectDeps(m, capName, visited, &result)
	return result
}

func (q *QueryEngine) collectDeps(
	m *entity.UNMModel,
	capName string,
	visited map[string]bool,
	result *[]*entity.Capability,
) {
	cap, ok := m.Capabilities[capName]
	if !ok {
		return
	}
	for _, rel := range cap.DependsOn {
		dep := rel.TargetID.String()
		if visited[dep] {
			continue
		}
		visited[dep] = true
		if depCap, found := m.Capabilities[dep]; found {
			*result = append(*result, depCap)
			q.collectDeps(m, dep, visited, result)
		}
	}
}

// CapabilitiesForTeam returns all Capabilities owned by the named team.
func (q *QueryEngine) CapabilitiesForTeam(m *entity.UNMModel, teamName string) []*entity.Capability {
	return m.GetCapabilitiesForTeam(teamName)
}

// OrphanServices returns services not referenced in any capability's RealizedBy.
func (q *QueryEngine) OrphanServices(m *entity.UNMModel) []*entity.Service {
	return m.GetOrphanServices()
}

// UnmappedNeeds returns needs that have no SupportedBy relationships.
func (q *QueryEngine) UnmappedNeeds(m *entity.UNMModel) []*entity.Need {
	var result []*entity.Need
	for _, need := range m.Needs {
		if !need.IsMapped() {
			result = append(result, need)
		}
	}
	return result
}

// InteractionsForTeam returns all Interactions where the named team is either
// the FromTeam or the ToTeam.
func (q *QueryEngine) InteractionsForTeam(m *entity.UNMModel, teamName string) []*entity.Interaction {
	var result []*entity.Interaction
	for _, i := range m.Interactions {
		if i.FromTeamName == teamName || i.ToTeamName == teamName {
			result = append(result, i)
		}
	}
	return result
}
