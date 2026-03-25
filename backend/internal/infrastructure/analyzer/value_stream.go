package analyzer

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// TeamStreamCoherence holds the value stream coherence assessment for a single stream-aligned team.
type TeamStreamCoherence struct {
	TeamName       string   `json:"teamName"`
	NeedsServed    []string `json:"needsServed"`
	NeedCount      int      `json:"needCount"`
	CoherenceScore float64  `json:"coherenceScore"`
	LowCoherence   bool     `json:"lowCoherence"`
	SharedCapEdges int      `json:"sharedCapEdges"`
}

// ValueStreamReport holds the result of a value stream coherence analysis.
type ValueStreamReport struct {
	TeamCoherences    []TeamStreamCoherence `json:"teamCoherences"`
	LowCoherenceCount int                  `json:"lowCoherenceCount"`
}

// ValueStreamAnalyzer assesses whether each stream-aligned team is a coherent value stream.
type ValueStreamAnalyzer struct{}

// NewValueStreamAnalyzer constructs a ValueStreamAnalyzer.
func NewValueStreamAnalyzer() *ValueStreamAnalyzer {
	return &ValueStreamAnalyzer{}
}

// Analyze computes a ValueStreamReport for all stream-aligned teams in the model.
func (a *ValueStreamAnalyzer) Analyze(m *entity.UNMModel) ValueStreamReport {
	// Build reverse index: service name → team name
	svcToTeam := make(map[string]string)
	for _, svc := range m.Services {
		if svc.OwnerTeamName != "" {
			svcToTeam[svc.Name] = svc.OwnerTeamName
		}
	}

	// Build reverse index: capability name → set of service names that realize it
	capToServices := make(map[string]map[string]bool)
	for _, cap := range m.Capabilities {
		svcSet := make(map[string]bool)
		for _, rel := range cap.RealizedBy {
			svcSet[rel.TargetID.String()] = true
		}
		capToServices[cap.Name] = svcSet
	}

	// For each need, find which teams serve it
	// need name → set of capability names
	needCaps := make(map[string]map[string]bool)
	// need name → set of team names that contribute
	needTeams := make(map[string]map[string]bool)

	for _, need := range m.Needs {
		caps := make(map[string]bool)
		teams := make(map[string]bool)
		for _, rel := range need.SupportedBy {
			capName := rel.TargetID.String()
			caps[capName] = true
			// Walk cap → services → team
			for svcName := range capToServices[capName] {
				if teamName, ok := svcToTeam[svcName]; ok {
					teams[teamName] = true
				}
			}
		}
		needCaps[need.Name] = caps
		needTeams[need.Name] = teams
	}

	var coherences []TeamStreamCoherence
	lowCount := 0

	for _, team := range m.Teams {
		if team.TeamType != valueobject.StreamAligned {
			continue
		}

		// Find all needs this team serves
		var needNames []string
		for needName, teams := range needTeams {
			if teams[team.Name] {
				needNames = append(needNames, needName)
			}
		}
		sort.Strings(needNames)

		n := len(needNames)
		var coherenceScore float64
		var sharedEdges int

		if n <= 1 {
			coherenceScore = 1.0
		} else {
			// Count pairs of needs that share at least one capability
			for i := 0; i < n; i++ {
				for j := i + 1; j < n; j++ {
					capsI := needCaps[needNames[i]]
					capsJ := needCaps[needNames[j]]
					if hasIntersection(capsI, capsJ) {
						sharedEdges++
					}
				}
			}
			maxEdges := n * (n - 1) / 2
			coherenceScore = float64(sharedEdges) / float64(maxEdges)
		}

		lowCoherence := coherenceScore < 0.4 && n >= 2

		if lowCoherence {
			lowCount++
		}

		coherences = append(coherences, TeamStreamCoherence{
			TeamName:       team.Name,
			NeedsServed:    needNames,
			NeedCount:      n,
			CoherenceScore: coherenceScore,
			LowCoherence:   lowCoherence,
			SharedCapEdges: sharedEdges,
		})
	}

	sort.Slice(coherences, func(i, j int) bool {
		return coherences[i].TeamName < coherences[j].TeamName
	})

	return ValueStreamReport{
		TeamCoherences:    coherences,
		LowCoherenceCount: lowCount,
	}
}

// hasIntersection returns true if two sets share at least one element.
func hasIntersection(a, b map[string]bool) bool {
	// Iterate over the smaller set for efficiency
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}
