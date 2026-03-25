package entity

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestNewService(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		s, err := NewService("svc-1", "Payment Service", "Handles payments", "payments-team")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.ID.String() != "svc-1" {
			t.Errorf("expected ID %q, got %q", "svc-1", s.ID.String())
		}
		if s.Name != "Payment Service" {
			t.Errorf("expected Name %q, got %q", "Payment Service", s.Name)
		}
		if s.OwnerTeamName != "payments-team" {
			t.Errorf("expected OwnerTeamName %q, got %q", "payments-team", s.OwnerTeamName)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewService("svc-1", "", "desc", "team")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewService("", "Payment Service", "desc", "team")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("AddDependsOn appends relationship", func(t *testing.T) {
		s, _ := NewService("svc-1", "Service", "desc", "team")
		id, _ := valueobject.NewEntityID("svc-2")
		s.AddDependsOn(NewRelationship(id, "depends on auth", valueobject.Supporting))
		if len(s.DependsOn) != 1 {
			t.Errorf("expected 1 DependsOn, got %d", len(s.DependsOn))
		}
	})
}
