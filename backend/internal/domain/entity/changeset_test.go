package entity

import (
	"encoding/json"
	"testing"
)

func TestChangeAction_Validate_MoveService(t *testing.T) {
	valid := ChangeAction{
		Type:         ActionMoveService,
		ServiceName:  "my-svc",
		FromTeamName: "team-a",
		ToTeamName:   "team-b",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	tests := []struct {
		name   string
		action ChangeAction
	}{
		{"missing service", ChangeAction{Type: ActionMoveService, FromTeamName: "a", ToTeamName: "b"}},
		{"missing from", ChangeAction{Type: ActionMoveService, ServiceName: "s", ToTeamName: "b"}},
		{"missing to", ChangeAction{Type: ActionMoveService, ServiceName: "s", FromTeamName: "a"}},
		{"same team", ChangeAction{Type: ActionMoveService, ServiceName: "s", FromTeamName: "x", ToTeamName: "x"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.action.Validate(); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestChangeAction_Validate_SplitTeam(t *testing.T) {
	valid := ChangeAction{
		Type:             ActionSplitTeam,
		OriginalTeamName: "big-team",
		NewTeamAName:     "team-a",
		NewTeamBName:     "team-b",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := (ChangeAction{Type: ActionSplitTeam, OriginalTeamName: "t", NewTeamAName: "x", NewTeamBName: "x"}).Validate(); err == nil {
		t.Error("same team names should be invalid")
	}
}

func TestChangeAction_Validate_MergeTeams(t *testing.T) {
	valid := ChangeAction{
		Type:        ActionMergeTeams,
		TeamAName:   "a",
		TeamBName:   "b",
		NewTeamName: "merged",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := (ChangeAction{Type: ActionMergeTeams, TeamAName: "x", TeamBName: "x", NewTeamName: "m"}).Validate(); err == nil {
		t.Error("same team names should be invalid")
	}
}

func TestChangeAction_Validate_AddRemoveCapability(t *testing.T) {
	if err := (ChangeAction{Type: ActionAddCapability, CapabilityName: "cap"}).Validate(); err != nil {
		t.Errorf("add_capability: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionAddCapability}).Validate(); err == nil {
		t.Error("add_capability: missing name should error")
	}
	if err := (ChangeAction{Type: ActionRemoveCapability, CapabilityName: "cap"}).Validate(); err != nil {
		t.Errorf("remove_capability: expected no error, got %v", err)
	}
}

func TestChangeAction_Validate_ReassignCapability(t *testing.T) {
	valid := ChangeAction{
		Type:           ActionReassignCapability,
		CapabilityName: "cap",
		FromTeamName:   "a",
		ToTeamName:     "b",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionReassignCapability, CapabilityName: "c", FromTeamName: "x", ToTeamName: "x"}).Validate(); err == nil {
		t.Error("same team should be invalid")
	}
}

func TestChangeAction_Validate_Interactions(t *testing.T) {
	addOk := ChangeAction{Type: ActionAddInteraction, SourceTeamName: "a", TargetTeamName: "b", InteractionMode: "collaboration"}
	if err := addOk.Validate(); err != nil {
		t.Errorf("add_interaction: %v", err)
	}
	if err := (ChangeAction{Type: ActionAddInteraction, SourceTeamName: "a", TargetTeamName: "b"}).Validate(); err == nil {
		t.Error("add_interaction: missing mode should error")
	}

	removeOk := ChangeAction{Type: ActionRemoveInteraction, SourceTeamName: "a", TargetTeamName: "b"}
	if err := removeOk.Validate(); err != nil {
		t.Errorf("remove_interaction: %v", err)
	}
}

func TestChangeAction_Validate_UpdateTeamSize(t *testing.T) {
	if err := (ChangeAction{Type: ActionUpdateTeamSize, TeamName: "t", NewSize: 5}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionUpdateTeamSize, TeamName: "t", NewSize: 0}).Validate(); err == nil {
		t.Error("size 0 should be invalid")
	}
	if err := (ChangeAction{Type: ActionUpdateTeamSize, NewSize: 5}).Validate(); err == nil {
		t.Error("missing team name should be invalid")
	}
}

func TestChangeAction_Validate_AddService(t *testing.T) {
	if err := (ChangeAction{Type: ActionAddService, ServiceName: "s", OwnerTeamName: "t"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionAddService, OwnerTeamName: "t"}).Validate(); err == nil {
		t.Error("missing service_name should error")
	}
	if err := (ChangeAction{Type: ActionAddService, ServiceName: "s"}).Validate(); err == nil {
		t.Error("missing owner_team_name should error")
	}
}

func TestChangeAction_Validate_RemoveService(t *testing.T) {
	if err := (ChangeAction{Type: ActionRemoveService, ServiceName: "s"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionRemoveService}).Validate(); err == nil {
		t.Error("missing service_name should error")
	}
}

func TestChangeAction_Validate_RenameService(t *testing.T) {
	if err := (ChangeAction{Type: ActionRenameService, ServiceName: "a", NewServiceName: "b"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionRenameService, ServiceName: "a", NewServiceName: "a"}).Validate(); err == nil {
		t.Error("same names should error")
	}
	if err := (ChangeAction{Type: ActionRenameService, NewServiceName: "b"}).Validate(); err == nil {
		t.Error("missing service_name should error")
	}
}

func TestChangeAction_Validate_AddTeam(t *testing.T) {
	if err := (ChangeAction{Type: ActionAddTeam, TeamName: "t", TeamType: "stream-aligned"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionAddTeam, TeamType: "stream-aligned"}).Validate(); err == nil {
		t.Error("missing team_name should error")
	}
	if err := (ChangeAction{Type: ActionAddTeam, TeamName: "t"}).Validate(); err == nil {
		t.Error("missing team_type should error")
	}
}

func TestChangeAction_Validate_RemoveTeam(t *testing.T) {
	if err := (ChangeAction{Type: ActionRemoveTeam, TeamName: "t"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionRemoveTeam}).Validate(); err == nil {
		t.Error("missing team_name should error")
	}
}

func TestChangeAction_Validate_UpdateTeamType(t *testing.T) {
	if err := (ChangeAction{Type: ActionUpdateTeamType, TeamName: "t", TeamType: "platform"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionUpdateTeamType, TeamType: "platform"}).Validate(); err == nil {
		t.Error("missing team_name should error")
	}
	if err := (ChangeAction{Type: ActionUpdateTeamType, TeamName: "t"}).Validate(); err == nil {
		t.Error("missing team_type should error")
	}
}

func TestChangeAction_Validate_AddNeed(t *testing.T) {
	if err := (ChangeAction{Type: ActionAddNeed, NeedName: "n", ActorName: "a"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionAddNeed, ActorName: "a"}).Validate(); err == nil {
		t.Error("missing need_name should error")
	}
	if err := (ChangeAction{Type: ActionAddNeed, NeedName: "n"}).Validate(); err == nil {
		t.Error("missing actor_name should error")
	}
}

func TestChangeAction_Validate_RemoveNeed(t *testing.T) {
	if err := (ChangeAction{Type: ActionRemoveNeed, NeedName: "n"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionRemoveNeed}).Validate(); err == nil {
		t.Error("missing need_name should error")
	}
}

func TestChangeAction_Validate_AddRemoveActor(t *testing.T) {
	if err := (ChangeAction{Type: ActionAddActor, ActorName: "a"}).Validate(); err != nil {
		t.Errorf("add_actor: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionAddActor}).Validate(); err == nil {
		t.Error("add_actor: missing actor_name should error")
	}
	if err := (ChangeAction{Type: ActionRemoveActor, ActorName: "a"}).Validate(); err != nil {
		t.Errorf("remove_actor: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionRemoveActor}).Validate(); err == nil {
		t.Error("remove_actor: missing actor_name should error")
	}
}

func TestChangeAction_Validate_ServiceDependency(t *testing.T) {
	if err := (ChangeAction{Type: ActionAddServiceDependency, ServiceName: "a", DependsOnService: "b"}).Validate(); err != nil {
		t.Errorf("add: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionAddServiceDependency, ServiceName: "a", DependsOnService: "a"}).Validate(); err == nil {
		t.Error("add: same service should error")
	}
	if err := (ChangeAction{Type: ActionRemoveServiceDependency, ServiceName: "a", DependsOnService: "b"}).Validate(); err != nil {
		t.Errorf("remove: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionRemoveServiceDependency, ServiceName: "a"}).Validate(); err == nil {
		t.Error("remove: missing depends_on_service should error")
	}
}

func TestChangeAction_Validate_NeedCapabilityLink(t *testing.T) {
	if err := (ChangeAction{Type: ActionLinkNeedCapability, NeedName: "n", CapabilityName: "c"}).Validate(); err != nil {
		t.Errorf("link: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionLinkNeedCapability, CapabilityName: "c"}).Validate(); err == nil {
		t.Error("link: missing need_name should error")
	}
	if err := (ChangeAction{Type: ActionUnlinkNeedCapability, NeedName: "n", CapabilityName: "c"}).Validate(); err != nil {
		t.Errorf("unlink: expected no error, got %v", err)
	}
}

func TestChangeAction_Validate_CapabilityServiceLink(t *testing.T) {
	if err := (ChangeAction{Type: ActionLinkCapabilityService, CapabilityName: "c", ServiceName: "s"}).Validate(); err != nil {
		t.Errorf("link: expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionLinkCapabilityService, ServiceName: "s"}).Validate(); err == nil {
		t.Error("link: missing capability_name should error")
	}
	if err := (ChangeAction{Type: ActionUnlinkCapabilityService, CapabilityName: "c", ServiceName: "s"}).Validate(); err != nil {
		t.Errorf("unlink: expected no error, got %v", err)
	}
}

func TestChangeAction_Validate_UpdateCapabilityVisibility(t *testing.T) {
	if err := (ChangeAction{Type: ActionUpdateCapabilityVisibility, CapabilityName: "c", Visibility: "user-facing"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionUpdateCapabilityVisibility, Visibility: "user-facing"}).Validate(); err == nil {
		t.Error("missing capability_name should error")
	}
	if err := (ChangeAction{Type: ActionUpdateCapabilityVisibility, CapabilityName: "c"}).Validate(); err == nil {
		t.Error("missing visibility should error")
	}
}

func TestChangeAction_Validate_UpdateDescription(t *testing.T) {
	if err := (ChangeAction{Type: ActionUpdateDescription, EntityType: "service", EntityName: "s"}).Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := (ChangeAction{Type: ActionUpdateDescription, EntityName: "s"}).Validate(); err == nil {
		t.Error("missing entity_type should error")
	}
	if err := (ChangeAction{Type: ActionUpdateDescription, EntityType: "service"}).Validate(); err == nil {
		t.Error("missing entity_name should error")
	}
}

func TestChangeAction_Validate_UnknownType(t *testing.T) {
	if err := (ChangeAction{Type: "bogus"}).Validate(); err == nil {
		t.Error("unknown action type should error")
	}
}

func TestChangeset_AddAction(t *testing.T) {
	cs, err := NewChangeset("cs-1", "test changeset")
	if err != nil {
		t.Fatalf("NewChangeset: %v", err)
	}
	if !cs.IsEmpty() {
		t.Error("new changeset should be empty")
	}

	action := ChangeAction{
		Type:         ActionMoveService,
		ServiceName:  "svc",
		FromTeamName: "a",
		ToTeamName:   "b",
	}
	if err := cs.AddAction(action); err != nil {
		t.Fatalf("AddAction: %v", err)
	}
	if cs.IsEmpty() {
		t.Error("changeset should not be empty after adding action")
	}
	if len(cs.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(cs.Actions))
	}
}

func TestChangeset_AddAction_InvalidRejected(t *testing.T) {
	cs, _ := NewChangeset("cs-1", "")
	bad := ChangeAction{Type: ActionMoveService} // missing required fields
	if err := cs.AddAction(bad); err == nil {
		t.Error("expected error for invalid action")
	}
	if !cs.IsEmpty() {
		t.Error("invalid action should not be added")
	}
}

func TestChangeset_JSON(t *testing.T) {
	cs, _ := NewChangeset("cs-42", "move ingester")
	_ = cs.AddAction(ChangeAction{
		Type:         ActionMoveService,
		ServiceName:  "ingester",
		FromTeamName: "team-a",
		ToTeamName:   "team-b",
	})

	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out Changeset
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.ID != "cs-42" {
		t.Errorf("ID: want cs-42, got %s", out.ID)
	}
	if len(out.Actions) != 1 {
		t.Errorf("actions: want 1, got %d", len(out.Actions))
	}
	if out.Actions[0].ServiceName != "ingester" {
		t.Errorf("service name: want ingester, got %s", out.Actions[0].ServiceName)
	}
}

func TestNewChangeset_EmptyID(t *testing.T) {
	if _, err := NewChangeset("", "desc"); err == nil {
		t.Error("empty id should error")
	}
}
