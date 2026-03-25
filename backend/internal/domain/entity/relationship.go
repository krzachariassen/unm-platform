package entity

import "github.com/krzachariassen/unm-platform/internal/domain/valueobject"

// Relationship is a directed link to another entity, with optional description and role.
type Relationship struct {
	TargetID    valueobject.EntityID
	Description string                       // optional human-readable label for edge
	Role        valueobject.RelationshipRole // optional: primary/supporting/consuming
}

// NewRelationship constructs a Relationship with the given target ID, description, and role.
func NewRelationship(targetID valueobject.EntityID, description string, role valueobject.RelationshipRole) Relationship {
	return Relationship{
		TargetID:    targetID,
		Description: description,
		Role:        role,
	}
}
