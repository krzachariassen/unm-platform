package entity

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestNewCapability(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		c, err := NewCapability("cap-1", "Payment Processing", "Handles payment transactions")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID.String() != "cap-1" {
			t.Errorf("expected ID %q, got %q", "cap-1", c.ID.String())
		}
		if c.Name != "Payment Processing" {
			t.Errorf("expected Name %q, got %q", "Payment Processing", c.Name)
		}
		if c.Description != "Handles payment transactions" {
			t.Errorf("expected Description %q, got %q", "Handles payment transactions", c.Description)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewCapability("cap-1", "", "desc")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewCapability("", "Payment Processing", "desc")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("IsLeaf true for new capability", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		if !c.IsLeaf() {
			t.Error("expected IsLeaf true for new capability with no children")
		}
	})

	t.Run("IsLeaf false after AddChild", func(t *testing.T) {
		parent, _ := NewCapability("cap-1", "Payment Processing", "")
		child, _ := NewCapability("cap-2", "Card Processing", "")
		parent.AddChild(child)
		if parent.IsLeaf() {
			t.Error("expected IsLeaf false after adding a child")
		}
	})

	t.Run("Depth is 1 for leaf", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		if c.Depth() != 1 {
			t.Errorf("expected Depth 1 for leaf, got %d", c.Depth())
		}
	})

	t.Run("Depth reflects nested hierarchy", func(t *testing.T) {
		root, _ := NewCapability("cap-1", "Root", "")
		mid, _ := NewCapability("cap-2", "Mid", "")
		leaf, _ := NewCapability("cap-3", "Leaf", "")

		mid.AddChild(leaf)
		root.AddChild(mid)

		if root.Depth() != 3 {
			t.Errorf("expected Depth 3 for 3-level hierarchy, got %d", root.Depth())
		}
		if mid.Depth() != 2 {
			t.Errorf("expected Depth 2 for mid node, got %d", mid.Depth())
		}
		if leaf.Depth() != 1 {
			t.Errorf("expected Depth 1 for leaf, got %d", leaf.Depth())
		}
	})

	t.Run("DecomposesTo is alias for Children", func(t *testing.T) {
		parent, _ := NewCapability("cap-1", "Parent", "")
		child, _ := NewCapability("cap-2", "Child", "")
		parent.AddChild(child)
		if len(parent.DecomposesTo) != len(parent.Children) {
			t.Errorf("expected DecomposesTo len %d == Children len %d", len(parent.DecomposesTo), len(parent.Children))
		}
	})

	t.Run("IsFragmented false for 2 teams", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		if c.IsFragmented([]string{"team-a", "team-b"}) {
			t.Error("expected IsFragmented false for 2 teams")
		}
	})

	t.Run("IsFragmented true for more than 2 teams", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		if !c.IsFragmented([]string{"team-a", "team-b", "team-c"}) {
			t.Error("expected IsFragmented true for 3 teams")
		}
	})

	t.Run("AddRealizes on service links capability", func(t *testing.T) {
		s, _ := NewService("svc-1", "svc-1", "payment service", "team-a")
		id, _ := valueobject.NewEntityID("cap-1")
		s.AddRealizes(NewRelationship(id, "payment service", valueobject.Primary))
		if len(s.Realizes) != 1 {
			t.Errorf("expected 1 Realizes, got %d", len(s.Realizes))
		}
	})

	t.Run("AddDependsOn appends relationship", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		id, _ := valueobject.NewEntityID("cap-2")
		c.AddDependsOn(NewRelationship(id, "depends on auth", valueobject.Supporting))
		if len(c.DependsOn) != 1 {
			t.Errorf("expected 1 DependsOn, got %d", len(c.DependsOn))
		}
	})

	t.Run("Visibility defaults to empty", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		if c.Visibility != "" {
			t.Errorf("expected empty Visibility, got %q", c.Visibility)
		}
	})

	t.Run("SetVisibility accepts valid values", func(t *testing.T) {
		validValues := []string{
			CapVisibilityUserFacing,
			CapVisibilityDomain,
			CapVisibilityFoundational,
			CapVisibilityInfrastructure,
			"", // empty is allowed
		}
		for _, v := range validValues {
			c, _ := NewCapability("cap-1", "Payment Processing", "")
			if err := c.SetVisibility(v); err != nil {
				t.Errorf("expected no error for visibility %q, got %v", v, err)
			}
			if c.Visibility != v {
				t.Errorf("expected Visibility %q, got %q", v, c.Visibility)
			}
		}
	})

	t.Run("SetVisibility rejects invalid value", func(t *testing.T) {
		c, _ := NewCapability("cap-1", "Payment Processing", "")
		if err := c.SetVisibility("bogus-layer"); err == nil {
			t.Error("expected error for invalid visibility, got nil")
		}
	})
}
