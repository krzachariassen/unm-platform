package entity

import (
	"errors"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// Need represents what an actor (or multiple actors) is trying to achieve.
type Need struct {
	ID          valueobject.EntityID
	Name        string
	ActorNames  []string
	Outcome     string
	SupportedBy []Relationship
}

// NewNeed constructs a Need with a single actor. Returns an error if name or actorName is empty.
// The actorName is stored as []string{actorName} so all callers still work unchanged.
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
		ActorNames:  []string{actorName},
		Outcome:     outcome,
		SupportedBy: []Relationship{},
	}, nil
}

// NewNeedMultiActor constructs a Need shared by multiple actors.
// actorNames must be non-empty and contain no empty strings.
func NewNeedMultiActor(id, name string, actorNames []string, outcome string) (*Need, error) {
	if name == "" {
		return nil, errors.New("need: name must not be empty")
	}
	if len(actorNames) == 0 {
		return nil, errors.New("need: actorNames must not be empty")
	}
	for _, a := range actorNames {
		if a == "" {
			return nil, errors.New("need: actorNames must not contain empty strings")
		}
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(actorNames))
	copy(names, actorNames)
	return &Need{
		ID:          entityID,
		Name:        name,
		ActorNames:  names,
		Outcome:     outcome,
		SupportedBy: []Relationship{},
	}, nil
}

// HasActor returns true if any of the need's ActorNames matches the given name.
func (n *Need) HasActor(actorName string) bool {
	for _, a := range n.ActorNames {
		if a == actorName {
			return true
		}
	}
	return false
}

// AddSupportedBy appends a Relationship to the SupportedBy slice.
func (n *Need) AddSupportedBy(r Relationship) {
	n.SupportedBy = append(n.SupportedBy, r)
}

// IsMapped returns true if the Need has at least one supporting relationship.
func (n *Need) IsMapped() bool {
	return len(n.SupportedBy) > 0
}
