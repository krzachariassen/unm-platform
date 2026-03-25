package entity

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestNewTeam(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		team, err := NewTeam("team-1", "Payments Team", "Owns payment capabilities", valueobject.StreamAligned)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if team.ID.String() != "team-1" {
			t.Errorf("expected ID %q, got %q", "team-1", team.ID.String())
		}
		if team.Name != "Payments Team" {
			t.Errorf("expected Name %q, got %q", "Payments Team", team.Name)
		}
		if team.TeamType != valueobject.StreamAligned {
			t.Errorf("expected TeamType %v, got %v", valueobject.StreamAligned, team.TeamType)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewTeam("team-1", "", "desc", valueobject.StreamAligned)
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewTeam("", "Payments Team", "desc", valueobject.StreamAligned)
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("IsOverloaded false for 6 or fewer capabilities", func(t *testing.T) {
		team, _ := NewTeam("team-1", "Payments Team", "desc", valueobject.StreamAligned)
		for i := 0; i < 6; i++ {
			id, _ := valueobject.NewEntityID("cap-" + string(rune('0'+i+1)))
			team.AddOwns(NewRelationship(id, "", valueobject.RelationshipRole("")))
		}
		if team.IsOverloaded(6) {
			t.Errorf("expected IsOverloaded false for %d capabilities", team.CapabilityCount())
		}
	})

	t.Run("IsOverloaded true for more than 6 capabilities", func(t *testing.T) {
		team, _ := NewTeam("team-1", "Payments Team", "desc", valueobject.StreamAligned)
		for i := 0; i < 7; i++ {
			id, _ := valueobject.NewEntityID("cap-" + string(rune('0'+i+1)))
			team.AddOwns(NewRelationship(id, "", valueobject.RelationshipRole("")))
		}
		if !team.IsOverloaded(6) {
			t.Errorf("expected IsOverloaded true for %d capabilities", team.CapabilityCount())
		}
	})

	t.Run("CapabilityCount matches Owns length", func(t *testing.T) {
		team, _ := NewTeam("team-1", "Payments Team", "desc", valueobject.Platform)
		if team.CapabilityCount() != 0 {
			t.Errorf("expected CapabilityCount 0, got %d", team.CapabilityCount())
		}
		id, _ := valueobject.NewEntityID("cap-1")
		team.AddOwns(NewRelationship(id, "", valueobject.RelationshipRole("")))
		if team.CapabilityCount() != 1 {
			t.Errorf("expected CapabilityCount 1, got %d", team.CapabilityCount())
		}
	})
}
