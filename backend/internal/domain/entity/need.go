package entity

import (
	"errors"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// Need represents what an actor is trying to achieve.
type Need struct {
	ID          valueobject.EntityID
	Name        string
	ActorName   string
	Outcome     string
	SupportedBy []Relationship
}

// NewNeed constructs a Need. Returns an error if name or actorName is empty.
func NewNeed(id, name, actorName, outcome string) (*Need, error) {
	if name == "" {
		return nil, errors.New("need: name must not be empty")
	}
	if actorName == "" {
		return nil, errors.New("need: actorName must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &Need{
		ID:          entityID,
		Name:        name,
		ActorName:   actorName,
		Outcome:     outcome,
		SupportedBy: []Relationship{},
	}, nil
}

// AddSupportedBy appends a Relationship to the SupportedBy slice.
func (n *Need) AddSupportedBy(r Relationship) {
	n.SupportedBy = append(n.SupportedBy, r)
}

// IsMapped returns true if the Need has at least one supporting relationship.
func (n *Need) IsMapped() bool {
	return len(n.SupportedBy) > 0
}
