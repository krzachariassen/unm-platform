package entity

import "testing"

func TestNewDataAsset(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		d, err := NewDataAsset("da-1", "orders-db", TypeDatabase, "Orders database")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.ID.String() != "da-1" {
			t.Errorf("expected ID %q, got %q", "da-1", d.ID.String())
		}
		if d.Name != "orders-db" {
			t.Errorf("expected Name %q, got %q", "orders-db", d.Name)
		}
		if d.Type != TypeDatabase {
			t.Errorf("expected Type %q, got %q", TypeDatabase, d.Type)
		}
		if d.Description != "Orders database" {
			t.Errorf("expected Description %q, got %q", "Orders database", d.Description)
		}
	})

	t.Run("all valid data asset types accepted", func(t *testing.T) {
		types := []string{TypeDatabase, TypeCache, TypeEventStream, TypeBlobStorage, TypeSearchIndex}
		for _, tp := range types {
			_, err := NewDataAsset("da-1", "asset", tp, "desc")
			if err != nil {
				t.Errorf("expected no error for type %q, got %v", tp, err)
			}
		}
	})

	t.Run("free-form type accepted", func(t *testing.T) {
		_, err := NewDataAsset("da-1", "asset", "custom-type", "desc")
		if err != nil {
			t.Errorf("expected no error for custom type, got %v", err)
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewDataAsset("", "orders-db", TypeDatabase, "desc")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewDataAsset("da-1", "", TypeDatabase, "desc")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("AddUsedBy appends service names", func(t *testing.T) {
		d, _ := NewDataAsset("da-1", "orders-db", TypeDatabase, "")
		d.AddUsedBy("order-service")
		d.AddUsedBy("reporting-service")
		if len(d.UsedBy) != 2 {
			t.Errorf("expected 2 UsedBy, got %d", len(d.UsedBy))
		}
		if d.UsedBy[0] != "order-service" {
			t.Errorf("expected UsedBy[0] %q, got %q", "order-service", d.UsedBy[0])
		}
		if d.UsedBy[1] != "reporting-service" {
			t.Errorf("expected UsedBy[1] %q, got %q", "reporting-service", d.UsedBy[1])
		}
	})
}
