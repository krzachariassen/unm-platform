package analyzer

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// TeamModeOverload holds information about a team that over-relies on a single interaction mode.
type TeamModeOverload struct {
	TeamName string
	Mode     valueobject.InteractionMode
	Count    int
}

// InteractionDiversityReport holds the result of an interaction diversity analysis.
type InteractionDiversityReport struct {
	// ModeDistribution is the count of each interaction mode across all model interactions.
	ModeDistribution map[valueobject.InteractionMode]int
	// IsolatedTeams are team names with zero declared interactions (neither from nor to).
	IsolatedTeams []string
	// OverReliantTeams are teams that appear in 4 or more interactions of the same mode.
	OverReliantTeams []TeamModeOverload
	// AllModesSame is true when all interactions use exactly one mode (and there are interactions).
	AllModesSame bool
}

// InteractionDiversityAnalyzer detects uniform or low-diversity team interaction patterns.
type InteractionDiversityAnalyzer struct {
	cfg entity.SignalsConfig
}

// NewInteractionDiversityAnalyzer constructs an InteractionDiversityAnalyzer.
func NewInteractionDiversityAnalyzer(cfg entity.SignalsConfig) *InteractionDiversityAnalyzer {
	return &InteractionDiversityAnalyzer{cfg: cfg}
}

// Analyze returns an InteractionDiversityReport for the model.
// A team is isolated when it appears in m.Teams but in no m.Interactions entry (as from or to).
// A team is over-reliant when it participates in 4+ interactions of the same mode.
// AllModesSame is set when the entire model uses exactly one interaction mode.
func (a *InteractionDiversityAnalyzer) Analyze(m *entity.UNMModel) InteractionDiversityReport {
	modeDist := make(map[valueobject.InteractionMode]int)

	// teamParticipates tracks which team names appear in at least one interaction.
	teamParticipates := make(map[string]bool)

	// teamModeCounts[teamName][mode] = count of interactions of that mode for that team
	// (a team increments for both its from-side and to-side).
	teamModeCounts := make(map[string]map[valueobject.InteractionMode]int)

	for _, intr := range m.Interactions {
		modeDist[intr.Mode]++
		from := intr.FromTeamName
		to := intr.ToTeamName
		teamParticipates[from] = true
		teamParticipates[to] = true
		if teamModeCounts[from] == nil {
			teamModeCounts[from] = make(map[valueobject.InteractionMode]int)
		}
		if teamModeCounts[to] == nil {
			teamModeCounts[to] = make(map[valueobject.InteractionMode]int)
		}
		teamModeCounts[from][intr.Mode]++
		teamModeCounts[to][intr.Mode]++
	}

	// Isolated teams: registered in the model but absent from all interactions.
	var isolated []string
	for name := range m.Teams {
		if !teamParticipates[name] {
			isolated = append(isolated, name)
		}
	}
	sort.Strings(isolated)

	// Over-reliant teams: threshold+ interactions of a single mode.
	var overReliant []TeamModeOverload
	for teamName, modes := range teamModeCounts {
		for mode, count := range modes {
			if count >= a.cfg.InteractionOverReliance {
				overReliant = append(overReliant, TeamModeOverload{
					TeamName: teamName,
					Mode:     mode,
					Count:    count,
				})
			}
		}
	}
	sort.Slice(overReliant, func(i, j int) bool {
		if overReliant[i].TeamName != overReliant[j].TeamName {
			return overReliant[i].TeamName < overReliant[j].TeamName
		}
		return string(overReliant[i].Mode) < string(overReliant[j].Mode)
	})

	allSame := len(modeDist) == 1 && len(m.Interactions) > 0

	return InteractionDiversityReport{
		ModeDistribution: modeDist,
		IsolatedTeams:    isolated,
		OverReliantTeams: overReliant,
		AllModesSame:     allSame,
	}
}
