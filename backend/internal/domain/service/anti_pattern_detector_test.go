package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
)

func TestDetectTeamAntiPatterns_Overloaded(t *testing.T) {
	team := &entity.Team{
		Name: "overloaded-team",
	}
	// Simulate 7 owned capabilities (> threshold of 6)
	for i := 0; i < 7; i++ {
		team.Owns = append(team.Owns, entity.Relationship{})
	}

	aps := service.DetectTeamAntiPatterns(team, 6)
	assert.Len(t, aps, 1)
	assert.Equal(t, "overloaded", aps[0].Code)
	assert.Equal(t, "warning", aps[0].Severity)
}

func TestDetectTeamAntiPatterns_NotOverloaded(t *testing.T) {
	team := &entity.Team{
		Name: "healthy-team",
	}
	for i := 0; i < 3; i++ {
		team.Owns = append(team.Owns, entity.Relationship{})
	}

	aps := service.DetectTeamAntiPatterns(team, 6)
	assert.Len(t, aps, 0)
}

func TestDetectCapabilityAntiPatterns_Fragmented(t *testing.T) {
	cap := &entity.Capability{Name: "Auth"}
	aps := service.DetectCapabilityAntiPatterns(cap, true, true)
	assert.Len(t, aps, 1)
	assert.Equal(t, "fragmented", aps[0].Code)
	assert.Equal(t, "warning", aps[0].Severity)
}

func TestDetectCapabilityAntiPatterns_NoServices(t *testing.T) {
	// Leaf capability with no services — should trigger no_services
	cap := &entity.Capability{Name: "LeafCap"}
	// IsLeaf() returns true when there are no children
	aps := service.DetectCapabilityAntiPatterns(cap, false, false)
	assert.Len(t, aps, 1)
	assert.Equal(t, "no_services", aps[0].Code)
}

func TestDetectCapabilityAntiPatterns_NoIssues(t *testing.T) {
	cap := &entity.Capability{Name: "HealthyCap"}
	aps := service.DetectCapabilityAntiPatterns(cap, false, true)
	assert.Len(t, aps, 0)
}

func TestDetectCapabilityAntiPatterns_FragmentedAndNoServices(t *testing.T) {
	cap := &entity.Capability{Name: "BadCap"}
	aps := service.DetectCapabilityAntiPatterns(cap, true, false)
	// Both fragmented and no_services
	assert.Len(t, aps, 2)
}

func TestAntiPatternCode_HasExpectedFields(t *testing.T) {
	team := &entity.Team{Name: "big-team"}
	for i := 0; i < 10; i++ {
		team.Owns = append(team.Owns, entity.Relationship{})
	}
	aps := service.DetectTeamAntiPatterns(team, 6)
	assert.NotEmpty(t, aps[0].Code)
	assert.NotEmpty(t, aps[0].Message)
	assert.NotEmpty(t, aps[0].Severity)
}
