package entity

import (
	"errors"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// Actor represents a person or system that has needs (e.g., Merchant, Eater, Operator).
type Actor struct {
	ID          valueobject.EntityID
	Name        string
	Description string
}

// NewActor constructs an Actor. Returns an error if name is empty.
func NewActor(id, name, description string) (Actor, error) {
	if name == "" {
		return Actor{}, errors.New("actor: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return Actor{}, err
	}
	return Actor{
		ID:          entityID,
		Name:        name,
		Description: description,
	}, nil
}
