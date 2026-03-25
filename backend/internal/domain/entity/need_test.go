package entity

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/valueobject"
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
		if n.ActorName != "Merchant" {
			t.Errorf("expected ActorName %q, got %q", "Merchant", n.ActorName)
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
