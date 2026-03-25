package entity

import (
	"errors"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// Service represents a concrete implementation that realises one or more capabilities.
// A service is identified by its name, owned by a team, and may depend on other services.
type Service struct {
	ID            valueobject.EntityID
	Name          string
	Description   string
	OwnerTeamName string
	DependsOn     []Relationship
}

// NewService constructs a Service. Returns an error if name is empty.
func NewService(id, name, description, ownerTeamName string) (*Service, error) {
	if name == "" {
		return nil, errors.New("service: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &Service{
		ID:            entityID,
		Name:          name,
		Description:   description,
		OwnerTeamName: ownerTeamName,
		DependsOn:     []Relationship{},
	}, nil
}

// AddDependsOn appends a Relationship to DependsOn.
func (s *Service) AddDependsOn(r Relationship) {
	s.DependsOn = append(s.DependsOn, r)
}
