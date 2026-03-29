package service

import (
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// AntiPattern represents a detected anti-pattern on a node.
type AntiPattern struct {
	Code     string
	Message  string
	Severity string
}

// DetectTeamAntiPatterns returns anti-patterns detected for the given team.
// overloadThreshold is the maximum number of owned capabilities before the team
// is considered overloaded.
func DetectTeamAntiPatterns(t *entity.Team, overloadThreshold int) []AntiPattern {
	var aps []AntiPattern
	if t.IsOverloaded(overloadThreshold) {
		aps = append(aps, AntiPattern{
			Code:     "overloaded",
			Message:  "Team cognitive load exceeds threshold",
			Severity: "warning",
		})
	}
	return aps
}

// DetectCapabilityAntiPatterns returns anti-patterns detected for the given capability.
// isFragmented indicates the capability is owned by multiple teams.
// hasServices indicates the capability has at least one realizing service.
func DetectCapabilityAntiPatterns(c *entity.Capability, isFragmented bool, hasServices bool) []AntiPattern {
	var aps []AntiPattern
	if isFragmented {
		aps = append(aps, AntiPattern{
			Code:     "fragmented",
			Message:  "Capability delivered by multiple teams",
			Severity: "warning",
		})
	}
	if c.IsLeaf() && !hasServices {
		aps = append(aps, AntiPattern{
			Code:     "no_services",
			Message:  "Capability has no realizing services",
			Severity: "warning",
		})
	}
	return aps
}
