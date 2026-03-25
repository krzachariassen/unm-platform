package entity

import "testing"

func TestNewActor(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		a, err := NewActor("actor-1", "Merchant", "A merchant using the platform")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.ID.String() != "actor-1" {
			t.Errorf("expected ID %q, got %q", "actor-1", a.ID.String())
		}
		if a.Name != "Merchant" {
			t.Errorf("expected Name %q, got %q", "Merchant", a.Name)
		}
		if a.Description != "A merchant using the platform" {
			t.Errorf("expected Description %q, got %q", "A merchant using the platform", a.Description)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewActor("actor-1", "", "some description")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewActor("", "Merchant", "some description")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})
}
