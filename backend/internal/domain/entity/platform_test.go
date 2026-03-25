package entity

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

func TestNewPlatform(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		p, err := NewPlatform("plat-1", "INCA Platform", "Internal platform for teams")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID.String() != "plat-1" {
			t.Errorf("expected ID %q, got %q", "plat-1", p.ID.String())
		}
		if p.Name != "INCA Platform" {
			t.Errorf("expected Name %q, got %q", "INCA Platform", p.Name)
		}
		if p.Description != "Internal platform for teams" {
			t.Errorf("expected Description %q, got %q", "Internal platform for teams", p.Description)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewPlatform("plat-1", "", "desc")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewPlatform("", "INCA Platform", "desc")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("AddTeam appends team name", func(t *testing.T) {
		p, _ := NewPlatform("plat-1", "INCA Platform", "")
		p.AddTeam("infra-team")
		p.AddTeam("data-team")
		if len(p.TeamNames) != 2 {
			t.Errorf("expected 2 TeamNames, got %d", len(p.TeamNames))
		}
		if p.TeamNames[0] != "infra-team" {
			t.Errorf("expected TeamNames[0] %q, got %q", "infra-team", p.TeamNames[0])
		}
	})

	t.Run("AddProvides appends relationship", func(t *testing.T) {
		p, _ := NewPlatform("plat-1", "INCA Platform", "")
		id, _ := valueobject.NewEntityID("cap-1")
		p.AddProvides(NewRelationship(id, "provides observability", valueobject.Primary))
		if len(p.Provides) != 1 {
			t.Errorf("expected 1 Provides, got %d", len(p.Provides))
		}
	})
}
