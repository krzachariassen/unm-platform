package entity

import (
	"errors"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// TeamInteraction describes how a team interacts with another team.
type TeamInteraction struct {
	TargetTeamName string
	Mode           valueobject.InteractionMode
	Via            string
	Description    string
}

// Team represents an organizational unit that owns services/capabilities.
type Team struct {
	ID            valueobject.EntityID
	Name          string
	Description   string
	TeamType      valueobject.TeamType
	Size          int  // number of people on the team; defaults to 5 if unset
	SizeExplicit  bool // true when the YAML/DSL explicitly set the team size
	Owns          []Relationship
	InteractsWith []TeamInteraction
}

// NewTeam constructs a Team. Returns an error if name is empty.
func NewTeam(id, name, description string, teamType valueobject.TeamType) (*Team, error) {
	if name == "" {
		return nil, errors.New("team: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &Team{
		ID:            entityID,
		Name:          name,
		Description:   description,
		TeamType:      teamType,
		Size:          5, // Team Topologies default: typical small team
		Owns:          []Relationship{},
		InteractsWith: []TeamInteraction{},
	}, nil
}

// AddOwns appends a Relationship to Owns.
func (t *Team) AddOwns(r Relationship) {
	t.Owns = append(t.Owns, r)
}

// CapabilityCount returns the number of capabilities owned by this team.
func (t *Team) CapabilityCount() int {
	return len(t.Owns)
}

// EffectiveSize returns the team's size, defaulting to 5 if unset.
func (t *Team) EffectiveSize() int {
	if t.Size <= 0 {
		return 5
	}
	return t.Size
}

// IsOverloaded returns true if the team owns more capabilities than the given threshold.
// This is a quick heuristic — the CognitiveLoadAnalyzer provides a more
// accurate assessment based on team size, type, and all responsibility signals.
func (t *Team) IsOverloaded(threshold int) bool {
	return t.CapabilityCount() > threshold
}
