package entity

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

func TestNewRelationship(t *testing.T) {
	id, err := valueobject.NewEntityID("cap-1")
	if err != nil {
		t.Fatalf("unexpected error creating EntityID: %v", err)
	}

	t.Run("constructs with all fields", func(t *testing.T) {
		r := NewRelationship(id, "supports checkout", valueobject.Primary)
		if r.TargetID != id {
			t.Errorf("expected TargetID %v, got %v", id, r.TargetID)
		}
		if r.Description != "supports checkout" {
			t.Errorf("expected Description %q, got %q", "supports checkout", r.Description)
		}
		if r.Role != valueobject.Primary {
			t.Errorf("expected Role %v, got %v", valueobject.Primary, r.Role)
		}
	})

	t.Run("constructs with empty description and no role", func(t *testing.T) {
		r := NewRelationship(id, "", valueobject.RelationshipRole(""))
		if r.TargetID != id {
			t.Errorf("expected TargetID %v, got %v", id, r.TargetID)
		}
		if r.Description != "" {
			t.Errorf("expected empty description, got %q", r.Description)
		}
	})
}
