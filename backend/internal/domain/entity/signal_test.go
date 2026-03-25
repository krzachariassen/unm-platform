package entity

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestNewSignal(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		s, err := NewSignal("sig-1", CategoryBottleneck, "payment-service", "Payment service is a bottleneck", "High latency observed", valueobject.SeverityHigh)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.ID.String() != "sig-1" {
			t.Errorf("expected ID %q, got %q", "sig-1", s.ID.String())
		}
		if s.Category != CategoryBottleneck {
			t.Errorf("expected Category %q, got %q", CategoryBottleneck, s.Category)
		}
		if s.OnEntityName != "payment-service" {
			t.Errorf("expected OnEntityName %q, got %q", "payment-service", s.OnEntityName)
		}
		if s.Severity != valueobject.SeverityHigh {
			t.Errorf("expected Severity %v, got %v", valueobject.SeverityHigh, s.Severity)
		}
	})

	t.Run("all valid categories accepted", func(t *testing.T) {
		categories := []string{
			CategoryBottleneck, CategoryFragmentation, CategoryCognitiveLoad,
			CategoryCoupling, CategoryGap,
		}
		for _, cat := range categories {
			_, err := NewSignal("sig-1", cat, "some-entity", "desc", "evidence", valueobject.SeverityLow)
			if err != nil {
				t.Errorf("expected no error for category %q, got %v", cat, err)
			}
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewSignal("", CategoryGap, "entity", "desc", "evidence", valueobject.SeverityLow)
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("invalid category returns error", func(t *testing.T) {
		_, err := NewSignal("sig-1", "invalid-category", "entity", "desc", "evidence", valueobject.SeverityLow)
		if err == nil {
			t.Error("expected error for invalid category, got nil")
		}
	})

	t.Run("empty onEntityName returns error", func(t *testing.T) {
		_, err := NewSignal("sig-1", CategoryGap, "", "desc", "evidence", valueobject.SeverityLow)
		if err == nil {
			t.Error("expected error for empty onEntityName, got nil")
		}
	})

	t.Run("AddAffected appends entity name", func(t *testing.T) {
		s, _ := NewSignal("sig-1", CategoryCoupling, "svc-a", "Tight coupling", "shared DB", valueobject.SeverityMedium)
		s.AddAffected("svc-b")
		s.AddAffected("svc-c")
		if len(s.AffectedEntities) != 2 {
			t.Errorf("expected 2 AffectedEntities, got %d", len(s.AffectedEntities))
		}
	})
}
