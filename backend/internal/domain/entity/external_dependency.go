package entity

import (
	"errors"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// ExternalUsage records a service's use of an external dependency.
type ExternalUsage struct {
	ServiceName string
	Description string
}

// ExternalDependency represents an external system or service that internal services depend on.
type ExternalDependency struct {
	ID          valueobject.EntityID
	Name        string
	Description string
	UsedBy      []ExternalUsage
}

// NewExternalDependency constructs an ExternalDependency. Returns an error if name is empty.
func NewExternalDependency(id, name, description string) (*ExternalDependency, error) {
	if name == "" {
		return nil, errors.New("external_dependency: name must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &ExternalDependency{
		ID:          entityID,
		Name:        name,
		Description: description,
		UsedBy:      []ExternalUsage{},
	}, nil
}

// AddUsedBy appends an ExternalUsage record to UsedBy.
func (e *ExternalDependency) AddUsedBy(serviceName, description string) {
	e.UsedBy = append(e.UsedBy, ExternalUsage{
		ServiceName: serviceName,
		Description: description,
	})
}
