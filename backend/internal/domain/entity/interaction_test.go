package entity

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/valueobject"
)

func TestNewInteraction(t *testing.T) {
	t.Run("valid construction", func(t *testing.T) {
		i, err := NewInteraction("int-1", "payments-team", "platform-team", valueobject.XAsAService, "API", "Payments uses platform API")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if i.ID.String() != "int-1" {
			t.Errorf("expected ID %q, got %q", "int-1", i.ID.String())
		}
		if i.FromTeamName != "payments-team" {
			t.Errorf("expected FromTeamName %q, got %q", "payments-team", i.FromTeamName)
		}
		if i.ToTeamName != "platform-team" {
			t.Errorf("expected ToTeamName %q, got %q", "platform-team", i.ToTeamName)
		}
		if i.Mode != valueobject.XAsAService {
			t.Errorf("expected Mode %v, got %v", valueobject.XAsAService, i.Mode)
		}
		if i.Via != "API" {
			t.Errorf("expected Via %q, got %q", "API", i.Via)
		}
		if i.Description != "Payments uses platform API" {
			t.Errorf("expected Description %q, got %q", "Payments uses platform API", i.Description)
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		_, err := NewInteraction("", "from-team", "to-team", valueobject.Collaboration, "", "")
		if err == nil {
			t.Error("expected error for empty id, got nil")
		}
	})

	t.Run("empty fromTeam returns error", func(t *testing.T) {
		_, err := NewInteraction("int-1", "", "platform-team", valueobject.Collaboration, "", "")
		if err == nil {
			t.Error("expected error for empty fromTeam, got nil")
		}
	})

	t.Run("empty toTeam returns error", func(t *testing.T) {
		_, err := NewInteraction("int-1", "payments-team", "", valueobject.Collaboration, "", "")
		if err == nil {
			t.Error("expected error for empty toTeam, got nil")
		}
	})
}
