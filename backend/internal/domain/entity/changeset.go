package entity

import (
	"errors"
	"fmt"
)

// ChangeActionType identifies the kind of structural change a ChangeAction represents.
type ChangeActionType string

const (
	ActionMoveService        ChangeActionType = "move_service"
	ActionSplitTeam          ChangeActionType = "split_team"
	ActionMergeTeams         ChangeActionType = "merge_teams"
	ActionAddCapability      ChangeActionType = "add_capability"
	ActionRemoveCapability   ChangeActionType = "remove_capability"
	ActionReassignCapability ChangeActionType = "reassign_capability"
	ActionAddInteraction     ChangeActionType = "add_interaction"
	ActionRemoveInteraction  ChangeActionType = "remove_interaction"
	ActionUpdateTeamSize     ChangeActionType = "update_team_size"

	ActionAddService                 ChangeActionType = "add_service"
	ActionRemoveService              ChangeActionType = "remove_service"
	ActionRenameService              ChangeActionType = "rename_service"
	ActionAddTeam                    ChangeActionType = "add_team"
	ActionRemoveTeam                 ChangeActionType = "remove_team"
	ActionUpdateTeamType             ChangeActionType = "update_team_type"
	ActionAddNeed                    ChangeActionType = "add_need"
	ActionRemoveNeed                 ChangeActionType = "remove_need"
	ActionAddActor                   ChangeActionType = "add_actor"
	ActionRemoveActor                ChangeActionType = "remove_actor"
	ActionAddServiceDependency       ChangeActionType = "add_service_dependency"
	ActionRemoveServiceDependency    ChangeActionType = "remove_service_dependency"
	ActionLinkNeedCapability         ChangeActionType = "link_need_capability"
	ActionUnlinkNeedCapability       ChangeActionType = "unlink_need_capability"
	ActionLinkCapabilityService      ChangeActionType = "link_capability_service"
	ActionUnlinkCapabilityService    ChangeActionType = "unlink_capability_service"
	ActionUpdateCapabilityVisibility ChangeActionType = "update_capability_visibility"
	ActionUpdateDescription          ChangeActionType = "update_description"
)

// ChangeAction is a single proposed structural modification to a UNM model.
// Each action is self-contained and carries all information needed to apply it.
type ChangeAction struct {
	Type ChangeActionType `json:"type"`

	// MoveService: move a service from one team to another.
	ServiceName  string `json:"service_name,omitempty"`
	FromTeamName string `json:"from_team_name,omitempty"`
	ToTeamName   string `json:"to_team_name,omitempty"`

	// SplitTeam: split one team into two, with explicit service assignments.
	OriginalTeamName  string            `json:"original_team_name,omitempty"`
	NewTeamAName      string            `json:"new_team_a_name,omitempty"`
	NewTeamBName      string            `json:"new_team_b_name,omitempty"`
	ServiceAssignment map[string]string `json:"service_assignment,omitempty"` // service → "a"|"b"

	// MergeTeams: merge two teams into one.
	TeamAName   string `json:"team_a_name,omitempty"`
	TeamBName   string `json:"team_b_name,omitempty"`
	NewTeamName string `json:"new_team_name,omitempty"`

	// AddCapability / RemoveCapability: manage capabilities.
	CapabilityName string `json:"capability_name,omitempty"`
	OwnerTeamName  string `json:"owner_team_name,omitempty"` // for AddCapability

	// ReassignCapability: move capability ownership between teams.
	// Uses CapabilityName, FromTeamName, ToTeamName.

	// AddInteraction / RemoveInteraction.
	SourceTeamName  string `json:"source_team_name,omitempty"`
	TargetTeamName  string `json:"target_team_name,omitempty"`
	InteractionMode string `json:"interaction_mode,omitempty"`

	// UpdateTeamSize: change the explicit headcount for a team.
	TeamName string `json:"team_name,omitempty"`
	NewSize  int    `json:"new_size,omitempty"`

	// AddService / RenameService: manage services.
	// Uses ServiceName for existing service name.
	NewServiceName string `json:"new_service_name,omitempty"` // for RenameService
	Description    string `json:"description,omitempty"`      // for add/update operations

	// AddTeam / UpdateTeamType.
	// Uses TeamName for existing team name.
	TeamType string `json:"team_type,omitempty"` // for AddTeam / UpdateTeamType

	// AddNeed: create a new need.
	NeedName    string   `json:"need_name,omitempty"`
	ActorName   string   `json:"actor_name,omitempty"`
	Outcome     string   `json:"outcome,omitempty"`
	SupportedBy []string `json:"supported_by,omitempty"` // capability names

	// AddActor.
	// Uses ActorName for the actor name.

	// Service dependency management.
	DependsOnService string `json:"depends_on_service,omitempty"`

	// Capability-service linking.
	// Uses CapabilityName, ServiceName.
	Role string `json:"role,omitempty"` // optional: primary/supporting/consuming

	// UpdateCapabilityVisibility.
	Visibility string `json:"visibility,omitempty"`

	// UpdateDescription.
	EntityType string `json:"entity_type,omitempty"` // "actor", "need", "capability", "service", "team"
	EntityName string `json:"entity_name,omitempty"`
	// Uses Description field.
}

// Validate checks that the action has all required fields for its type.
// It does NOT validate against a model — that is the Applier's responsibility.
func (a ChangeAction) Validate() error {
	switch a.Type {
	case ActionMoveService:
		if a.ServiceName == "" {
			return errors.New("move_service: service_name required")
		}
		if a.FromTeamName == "" {
			return errors.New("move_service: from_team_name required")
		}
		if a.ToTeamName == "" {
			return errors.New("move_service: to_team_name required")
		}
		if a.FromTeamName == a.ToTeamName {
			return errors.New("move_service: from_team_name and to_team_name must differ")
		}

	case ActionSplitTeam:
		if a.OriginalTeamName == "" {
			return errors.New("split_team: original_team_name required")
		}
		if a.NewTeamAName == "" {
			return errors.New("split_team: new_team_a_name required")
		}
		if a.NewTeamBName == "" {
			return errors.New("split_team: new_team_b_name required")
		}
		if a.NewTeamAName == a.NewTeamBName {
			return errors.New("split_team: new_team_a_name and new_team_b_name must differ")
		}

	case ActionMergeTeams:
		if a.TeamAName == "" {
			return errors.New("merge_teams: team_a_name required")
		}
		if a.TeamBName == "" {
			return errors.New("merge_teams: team_b_name required")
		}
		if a.NewTeamName == "" {
			return errors.New("merge_teams: new_team_name required")
		}
		if a.TeamAName == a.TeamBName {
			return errors.New("merge_teams: team_a_name and team_b_name must differ")
		}

	case ActionAddCapability:
		if a.CapabilityName == "" {
			return errors.New("add_capability: capability_name required")
		}

	case ActionRemoveCapability:
		if a.CapabilityName == "" {
			return errors.New("remove_capability: capability_name required")
		}

	case ActionReassignCapability:
		if a.CapabilityName == "" {
			return errors.New("reassign_capability: capability_name required")
		}
		if a.FromTeamName == "" {
			return errors.New("reassign_capability: from_team_name required")
		}
		if a.ToTeamName == "" {
			return errors.New("reassign_capability: to_team_name required")
		}
		if a.FromTeamName == a.ToTeamName {
			return errors.New("reassign_capability: from_team_name and to_team_name must differ")
		}

	case ActionAddInteraction:
		if a.SourceTeamName == "" {
			return errors.New("add_interaction: source_team_name required")
		}
		if a.TargetTeamName == "" {
			return errors.New("add_interaction: target_team_name required")
		}
		if a.InteractionMode == "" {
			return errors.New("add_interaction: interaction_mode required")
		}

	case ActionRemoveInteraction:
		if a.SourceTeamName == "" {
			return errors.New("remove_interaction: source_team_name required")
		}
		if a.TargetTeamName == "" {
			return errors.New("remove_interaction: target_team_name required")
		}

	case ActionUpdateTeamSize:
		if a.TeamName == "" {
			return errors.New("update_team_size: team_name required")
		}
		if a.NewSize <= 0 {
			return errors.New("update_team_size: new_size must be > 0")
		}

	case ActionAddService:
		if a.ServiceName == "" {
			return errors.New("add_service: service_name required")
		}
		if a.OwnerTeamName == "" {
			return errors.New("add_service: owner_team_name required")
		}

	case ActionRemoveService:
		if a.ServiceName == "" {
			return errors.New("remove_service: service_name required")
		}

	case ActionRenameService:
		if a.ServiceName == "" {
			return errors.New("rename_service: service_name required")
		}
		if a.NewServiceName == "" {
			return errors.New("rename_service: new_service_name required")
		}
		if a.ServiceName == a.NewServiceName {
			return errors.New("rename_service: service_name and new_service_name must differ")
		}

	case ActionAddTeam:
		if a.TeamName == "" {
			return errors.New("add_team: team_name required")
		}
		if a.TeamType == "" {
			return errors.New("add_team: team_type required")
		}

	case ActionRemoveTeam:
		if a.TeamName == "" {
			return errors.New("remove_team: team_name required")
		}

	case ActionUpdateTeamType:
		if a.TeamName == "" {
			return errors.New("update_team_type: team_name required")
		}
		if a.TeamType == "" {
			return errors.New("update_team_type: team_type required")
		}

	case ActionAddNeed:
		if a.NeedName == "" {
			return errors.New("add_need: need_name required")
		}
		if a.ActorName == "" {
			return errors.New("add_need: actor_name required")
		}

	case ActionRemoveNeed:
		if a.NeedName == "" {
			return errors.New("remove_need: need_name required")
		}

	case ActionAddActor:
		if a.ActorName == "" {
			return errors.New("add_actor: actor_name required")
		}

	case ActionRemoveActor:
		if a.ActorName == "" {
			return errors.New("remove_actor: actor_name required")
		}

	case ActionAddServiceDependency:
		if a.ServiceName == "" {
			return errors.New("add_service_dependency: service_name required")
		}
		if a.DependsOnService == "" {
			return errors.New("add_service_dependency: depends_on_service required")
		}
		if a.ServiceName == a.DependsOnService {
			return errors.New("add_service_dependency: service_name and depends_on_service must differ")
		}

	case ActionRemoveServiceDependency:
		if a.ServiceName == "" {
			return errors.New("remove_service_dependency: service_name required")
		}
		if a.DependsOnService == "" {
			return errors.New("remove_service_dependency: depends_on_service required")
		}

	case ActionLinkNeedCapability:
		if a.NeedName == "" {
			return errors.New("link_need_capability: need_name required")
		}
		if a.CapabilityName == "" {
			return errors.New("link_need_capability: capability_name required")
		}

	case ActionUnlinkNeedCapability:
		if a.NeedName == "" {
			return errors.New("unlink_need_capability: need_name required")
		}
		if a.CapabilityName == "" {
			return errors.New("unlink_need_capability: capability_name required")
		}

	case ActionLinkCapabilityService:
		if a.CapabilityName == "" {
			return errors.New("link_capability_service: capability_name required")
		}
		if a.ServiceName == "" {
			return errors.New("link_capability_service: service_name required")
		}

	case ActionUnlinkCapabilityService:
		if a.CapabilityName == "" {
			return errors.New("unlink_capability_service: capability_name required")
		}
		if a.ServiceName == "" {
			return errors.New("unlink_capability_service: service_name required")
		}

	case ActionUpdateCapabilityVisibility:
		if a.CapabilityName == "" {
			return errors.New("update_capability_visibility: capability_name required")
		}
		if a.Visibility == "" {
			return errors.New("update_capability_visibility: visibility required")
		}

	case ActionUpdateDescription:
		if a.EntityType == "" {
			return errors.New("update_description: entity_type required")
		}
		if a.EntityName == "" {
			return errors.New("update_description: entity_name required")
		}

	default:
		return fmt.Errorf("unknown action type: %q", a.Type)
	}
	return nil
}

// Changeset is an ordered list of proposed structural changes to a UNM model.
// It is applied as a unit: all actions are applied sequentially to produce
// a projected model. The source model is never mutated.
type Changeset struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Actions     []ChangeAction `json:"actions"`
}

// NewChangeset constructs an empty changeset with the given description.
func NewChangeset(id, description string) (*Changeset, error) {
	if id == "" {
		return nil, errors.New("changeset: id must not be empty")
	}
	return &Changeset{
		ID:          id,
		Description: description,
		Actions:     []ChangeAction{},
	}, nil
}

// AddAction appends a validated action to the changeset.
// Returns an error if the action fails its own structural validation.
func (cs *Changeset) AddAction(a ChangeAction) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("changeset %s: invalid action: %w", cs.ID, err)
	}
	cs.Actions = append(cs.Actions, a)
	return nil
}

// IsEmpty returns true if the changeset contains no actions.
func (cs *Changeset) IsEmpty() bool {
	return len(cs.Actions) == 0
}
