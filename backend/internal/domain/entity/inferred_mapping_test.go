package entity

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestNewInferredMapping(t *testing.T) {
	highConf, _ := valueobject.NewConfidence(0.9, "Direct code reference found")
	lowConf, _ := valueobject.NewConfidence(0.3, "Weak naming similarity")

	t.Run("valid construction", func(t *testing.T) {
		m, err := NewInferredMapping("im-1", "payment-service", "payment-processing", highConf, valueobject.Inferred)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.ID.String() != "im-1" {
			t.Errorf("expected ID %q, got %q", "im-1", m.ID.String())
		}
		if m.ServiceName != "payment-service" {
			t.Errorf("expected ServiceName %q, got %q", "payment-service", m.ServiceName)
		}
		if m.CapabilityName != "payment-processing" {
			t.Errorf("expected CapabilityName %q, got %q", "payment-processing", m.CapabilityName)
		}
		if m.Confidence != highConf {
			t.Errorf("expected Confidence %v, got %v", highConf, m.Confidence)
		}
		if m.Status != valueobject.Inferred {
			t.Errorf("expected Status %v, got %v", valueobject.Inferred, m.Status)
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewInferredMapping("", "payment-service", "payment-processing", highConf, valueobject.Inferred)
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("empty serviceName returns error", func(t *testing.T) {
		_, err := NewInferredMapping("im-1", "", "payment-processing", highConf, valueobject.Candidate)
		if err == nil {
			t.Error("expected error for empty serviceName, got nil")
		}
	})

	t.Run("empty capabilityName returns error", func(t *testing.T) {
		_, err := NewInferredMapping("im-1", "payment-service", "", highConf, valueobject.Candidate)
		if err == nil {
			t.Error("expected error for empty capabilityName, got nil")
		}
	})

	t.Run("IsLowConfidence false for high confidence", func(t *testing.T) {
		m, _ := NewInferredMapping("im-1", "svc", "cap", highConf, valueobject.Inferred)
		if m.IsLowConfidence() {
			t.Error("expected IsLowConfidence false for confidence 0.9")
		}
	})

	t.Run("IsLowConfidence true for low confidence", func(t *testing.T) {
		m, _ := NewInferredMapping("im-1", "svc", "cap", lowConf, valueobject.Candidate)
		if !m.IsLowConfidence() {
			t.Error("expected IsLowConfidence true for confidence 0.3")
		}
	})

	t.Run("IsLowConfidence false for exactly 0.5", func(t *testing.T) {
		conf, _ := valueobject.NewConfidence(0.5, "borderline")
		m, _ := NewInferredMapping("im-1", "svc", "cap", conf, valueobject.Inferred)
		if m.IsLowConfidence() {
			t.Error("expected IsLowConfidence false for confidence 0.5 (not strictly less than)")
		}
	})
}
