package entity

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestNewNeed(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		n, err := NewNeed("need-1", "Accept Payment", "Merchant", "Payment processed successfully")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n.ID.String() != "need-1" {
			t.Errorf("expected ID %q, got %q", "need-1", n.ID.String())
		}
		if n.Name != "Accept Payment" {
			t.Errorf("expected Name %q, got %q", "Accept Payment", n.Name)
		}
		// ActorNames stores single actor as slice
		if len(n.ActorNames) != 1 {
			t.Fatalf("expected 1 ActorName, got %d", len(n.ActorNames))
		}
		if n.ActorNames[0] != "Merchant" {
			t.Errorf("expected ActorNames[0] %q, got %q", "Merchant", n.ActorNames[0])
		}
		if n.Outcome != "Payment processed successfully" {
			t.Errorf("expected Outcome %q, got %q", "Payment processed successfully", n.Outcome)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewNeed("need-1", "", "Merchant", "outcome")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewNeed("", "Accept Payment", "Merchant", "outcome")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("empty actorName returns error", func(t *testing.T) {
		_, err := NewNeed("need-1", "Accept Payment", "", "outcome")
		if err == nil {
			t.Error("expected error for empty actorName, got nil")
		}
	})

	t.Run("IsMapped false when no relationships", func(t *testing.T) {
		n, _ := NewNeed("need-1", "Accept Payment", "Merchant", "outcome")
		if n.IsMapped() {
			t.Error("expected IsMapped false for new need with no SupportedBy")
		}
	})

	t.Run("IsMapped true after AddSupportedBy", func(t *testing.T) {
		n, _ := NewNeed("need-1", "Accept Payment", "Merchant", "outcome")
		id, _ := valueobject.NewEntityID("cap-1")
		n.AddSupportedBy(NewRelationship(id, "payment capability", valueobject.Primary))
		if !n.IsMapped() {
			t.Error("expected IsMapped true after AddSupportedBy")
		}
		if len(n.SupportedBy) != 1 {
			t.Errorf("expected 1 SupportedBy, got %d", len(n.SupportedBy))
		}
	})
}

func TestNewNeedMultiActor(t *testing.T) {
	t.Run("stores multiple actors", func(t *testing.T) {
		n, err := NewNeedMultiActor("need-1", "Shared Need", []string{"Actor A", "Actor B"}, "some outcome")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(n.ActorNames) != 2 {
			t.Fatalf("expected 2 ActorNames, got %d", len(n.ActorNames))
		}
		if n.ActorNames[0] != "Actor A" {
			t.Errorf("expected ActorNames[0] %q, got %q", "Actor A", n.ActorNames[0])
		}
		if n.ActorNames[1] != "Actor B" {
			t.Errorf("expected ActorNames[1] %q, got %q", "Actor B", n.ActorNames[1])
		}
	})

	t.Run("empty actorNames returns error", func(t *testing.T) {
		_, err := NewNeedMultiActor("need-1", "Need", []string{}, "outcome")
		if err == nil {
			t.Error("expected error for empty actorNames slice")
		}
	})

	t.Run("actorNames with empty string returns error", func(t *testing.T) {
		_, err := NewNeedMultiActor("need-1", "Need", []string{""}, "outcome")
		if err == nil {
			t.Error("expected error for empty actor name in slice")
		}
	})
}

func TestNeed_HasActor(t *testing.T) {
	t.Run("single actor", func(t *testing.T) {
		n, _ := NewNeed("need-1", "Need", "Merchant", "outcome")
		if !n.HasActor("Merchant") {
			t.Error("expected HasActor(Merchant) = true")
		}
		if n.HasActor("Driver") {
			t.Error("expected HasActor(Driver) = false")
		}
	})

	t.Run("multi actor", func(t *testing.T) {
		n, _ := NewNeedMultiActor("need-1", "Need", []string{"Actor A", "Actor B"}, "outcome")
		if !n.HasActor("Actor A") {
			t.Error("expected HasActor(Actor A) = true")
		}
		if !n.HasActor("Actor B") {
			t.Error("expected HasActor(Actor B) = true")
		}
		if n.HasActor("Actor C") {
			t.Error("expected HasActor(Actor C) = false")
		}
	})
}
