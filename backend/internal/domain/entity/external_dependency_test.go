package entity

import "testing"

func TestNewExternalDependency(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		e, err := NewExternalDependency("ext-1", "Stripe", "Payment gateway")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.ID.String() != "ext-1" {
			t.Errorf("expected ID %q, got %q", "ext-1", e.ID.String())
		}
		if e.Name != "Stripe" {
			t.Errorf("expected Name %q, got %q", "Stripe", e.Name)
		}
		if e.Description != "Payment gateway" {
			t.Errorf("expected Description %q, got %q", "Payment gateway", e.Description)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewExternalDependency("ext-1", "", "desc")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewExternalDependency("", "Stripe", "desc")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("AddUsedBy appends usage", func(t *testing.T) {
		e, _ := NewExternalDependency("ext-1", "Stripe", "Payment gateway")
		e.AddUsedBy("payment-service", "Processes card payments")
		e.AddUsedBy("refund-service", "Issues refunds")
		if len(e.UsedBy) != 2 {
			t.Errorf("expected 2 UsedBy, got %d", len(e.UsedBy))
		}
		if e.UsedBy[0].ServiceName != "payment-service" {
			t.Errorf("expected ServiceName %q, got %q", "payment-service", e.UsedBy[0].ServiceName)
		}
		if e.UsedBy[0].Description != "Processes card payments" {
			t.Errorf("expected Description %q, got %q", "Processes card payments", e.UsedBy[0].Description)
		}
	})
}
