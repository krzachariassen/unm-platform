package entity

import (
	"errors"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// Platform groups platform teams that provide shared capabilities.
type Platform struct {
	ID          valueobject.EntityID
	Name        string
	Description string
	TeamNames   []string
	Provides    []Relationship
}

// NewPlatform constructs a Platform. Returns an error if name is empty.
func NewPlatform(id, name, description string) (*Platform, error) {
	if name == "" {
		return nil, errors.New("platform: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &Platform{
		ID:          entityID,
		Name:        name,
		Description: description,
		TeamNames:   []string{},
		Provides:    []Relationship{},
	}, nil
}

// AddTeam appends a team name to the Platform.
func (p *Platform) AddTeam(teamName string) {
	p.TeamNames = append(p.TeamNames, teamName)
}

// AddProvides appends a Relationship to Provides.
func (p *Platform) AddProvides(r Relationship) {
	p.Provides = append(p.Provides, r)
}
