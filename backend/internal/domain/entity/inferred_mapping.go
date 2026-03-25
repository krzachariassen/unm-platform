package entity

import (
	"errors"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// InferredMapping represents an AI-suggested relationship between a service and a capability.
type InferredMapping struct {
	ID             valueobject.EntityID
	ServiceName    string
	CapabilityName string
	Confidence     valueobject.Confidence
	Status         valueobject.MappingStatus
	Evidence       string
}

// NewInferredMapping constructs an InferredMapping.
// Returns an error if serviceName or capabilityName is empty.
func NewInferredMapping(id, serviceName, capabilityName string, conf valueobject.Confidence, status valueobject.MappingStatus) (*InferredMapping, error) {
	if serviceName == "" {
		return nil, errors.New("inferred_mapping: serviceName must not be empty")
	}
	if capabilityName == "" {
		return nil, errors.New("inferred_mapping: capabilityName must not be empty")
	}
	entityID, err := valueobject.NewEntityID(id)
	if err != nil {
		return nil, err
	}
	return &InferredMapping{
		ID:             entityID,
		ServiceName:    serviceName,
		CapabilityName: capabilityName,
		Confidence:     conf,
		Status:         status,
	}, nil
}

// IsLowConfidence returns true if the mapping's confidence score is below 0.5.
func (m *InferredMapping) IsLowConfidence() bool {
	return m.Confidence.IsLow()
}
