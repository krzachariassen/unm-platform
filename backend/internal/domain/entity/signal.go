package entity

import (
	"errors"
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// Signal category constants.
const (
	CategoryBottleneck    = "bottleneck"
	CategoryFragmentation = "fragmentation"
	CategoryCognitiveLoad = "cognitive-load"
	CategoryCoupling      = "coupling"
	CategoryGap           = "gap"
)

var validSignalCategories = map[string]bool{
	CategoryBottleneck:    true,
	CategoryFragmentation: true,
	CategoryCognitiveLoad: true,
	CategoryCoupling:      true,
	CategoryGap:           true,
}

// Signal represents an architectural finding such as a bottleneck, fragmentation, or gap.
type Signal struct {
	ID               valueobject.EntityID
	Category         string
	OnEntityName     string
	Description      string
	Evidence         string
	Severity         valueobject.Severity
	AffectedEntities []string
}

// NewSignal constructs a Signal. Returns an error if category is invalid or onEntityName is empty.
func NewSignal(id, category, onEntityName, description, evidence string, severity valueobject.Severity) (*Signal, error) {
	if !validSignalCategories[category] {
		return nil, fmt.Errorf("signal: invalid category %q", category)
	}
	if onEntityName == "" {
		return nil, errors.New("signal: onEntityName must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &Signal{
		ID:               entityID,
		Category:         category,
		OnEntityName:     onEntityName,
		Description:      description,
		Evidence:         evidence,
		Severity:         severity,
		AffectedEntities: []string{},
	}, nil
}

// AddAffected appends an entity name to AffectedEntities.
func (s *Signal) AddAffected(entityName string) {
	s.AffectedEntities = append(s.AffectedEntities, entityName)
}
