package analyzer

import "github.com/krzachariassen/unm-platform/internal/domain/entity"

// FragmentedCapability pairs a Capability with all Teams that own it.
type FragmentedCapability struct {
	Capability *entity.Capability
	Teams      []*entity.Team
}

// FragmentationReport holds the result of a fragmentation analysis pass.
type FragmentationReport struct {
	FragmentedCapabilities []FragmentedCapability
}

// FragmentationAnalyzer finds capabilities whose implementation is fragmented across teams.
// Two fragmentation patterns are detected:
//  1. Explicit ownership: capability is listed in team.Owns for more than 2 teams.
//  2. Realization fragmentation: services that realize a capability (via service.Realizes) are owned by more than 1 team.
type FragmentationAnalyzer struct{}

// NewFragmentationAnalyzer constructs a FragmentationAnalyzer.
func NewFragmentationAnalyzer() *FragmentationAnalyzer {
	return &FragmentationAnalyzer{}
}

// Analyze finds all capabilities in m that are fragmented across teams.
func (a *FragmentationAnalyzer) Analyze(m *entity.UNMModel) FragmentationReport {
	var report FragmentationReport

	// Build a lookup: service name → owning team.
	svcTeam := make(map[string]*entity.Team)
	for _, svc := range m.Services {
		if svc.OwnerTeamName != "" {
			if team, ok := m.Teams[svc.OwnerTeamName]; ok {
				svcTeam[svc.Name] = team
			}
		}
	}

	seen := make(map[string]bool) // deduplicate by capability name

	for name, cap := range m.Capabilities {
		// Pattern 1: explicit team.Owns ownership fragmentation (> 2 teams).
		ownerTeams := m.GetTeamsForCapability(name)
		if len(ownerTeams) > 2 && !seen[name] {
			seen[name] = true
			report.FragmentedCapabilities = append(report.FragmentedCapabilities, FragmentedCapability{
				Capability: cap,
				Teams:      ownerTeams,
			})
			continue
		}

		// Pattern 2: realization fragmentation — services realizing this cap span > 1 team.
		teamSet := make(map[string]*entity.Team)
		for _, svc := range m.Services {
			for _, rel := range svc.Realizes {
				if rel.TargetID.String() == name {
					if t, ok := svcTeam[svc.Name]; ok {
						teamSet[t.Name] = t
					}
				}
			}
		}
		if len(teamSet) > 1 && !seen[name] {
			seen[name] = true
			teams := make([]*entity.Team, 0, len(teamSet))
			for _, t := range teamSet {
				teams = append(teams, t)
			}
			report.FragmentedCapabilities = append(report.FragmentedCapabilities, FragmentedCapability{
				Capability: cap,
				Teams:      teams,
			})
		}
	}

	return report
}
