package handler

import "github.com/krzachariassen/unm-platform/internal/domain/entity"

// changeActionDTO is the adapter-layer representation of a ChangeAction.
// JSON tags belong here, not on the domain entity.
type changeActionDTO struct {
	Type ChangeActionTypeDTO `json:"type"`

	// MoveService
	ServiceName  string `json:"service_name,omitempty"`
	FromTeamName string `json:"from_team_name,omitempty"`
	ToTeamName   string `json:"to_team_name,omitempty"`

	// SplitTeam
	OriginalTeamName  string            `json:"original_team_name,omitempty"`
	NewTeamAName      string            `json:"new_team_a_name,omitempty"`
	NewTeamBName      string            `json:"new_team_b_name,omitempty"`
	ServiceAssignment map[string]string `json:"service_assignment,omitempty"`

	// MergeTeams
	TeamAName   string `json:"team_a_name,omitempty"`
	TeamBName   string `json:"team_b_name,omitempty"`
	NewTeamName string `json:"new_team_name,omitempty"`

	// Capability operations
	CapabilityName string `json:"capability_name,omitempty"`
	OwnerTeamName  string `json:"owner_team_name,omitempty"`
	Visibility     string `json:"visibility,omitempty"`
	Role           string `json:"role,omitempty"`

	// Interaction operations
	SourceTeamName  string `json:"source_team_name,omitempty"`
	TargetTeamName  string `json:"target_team_name,omitempty"`
	InteractionMode string `json:"interaction_mode,omitempty"`

	// Team operations
	TeamName    string `json:"team_name,omitempty"`
	NewSize     int    `json:"new_size,omitempty"`
	TeamType    string `json:"team_type,omitempty"`
	Description string `json:"description,omitempty"`

	// RenameService
	NewServiceName string `json:"new_service_name,omitempty"`

	// Need / Actor operations
	NeedName    string   `json:"need_name,omitempty"`
	Outcome     string   `json:"outcome,omitempty"`
	SupportedBy []string `json:"supported_by,omitempty"`
	ActorName   string   `json:"actor_name,omitempty"`

	// Service dependency operations
	DependsOnService string `json:"depends_on_service,omitempty"`

	// UpdateDescription
	EntityType string `json:"entity_type,omitempty"`
	EntityName string `json:"entity_name,omitempty"`
}

// ChangeActionTypeDTO is the string representation of a ChangeActionType in the adapter layer.
type ChangeActionTypeDTO = string

// changesetDTO is the adapter-layer representation of a Changeset.
type changesetDTO struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Actions     []changeActionDTO `json:"actions"`
}

// toChangeActionDTO maps a domain ChangeAction to its DTO representation.
func toChangeActionDTO(a entity.ChangeAction) changeActionDTO {
	return changeActionDTO{
		Type:              string(a.Type),
		ServiceName:       a.ServiceName,
		FromTeamName:      a.FromTeamName,
		ToTeamName:        a.ToTeamName,
		OriginalTeamName:  a.OriginalTeamName,
		NewTeamAName:      a.NewTeamAName,
		NewTeamBName:      a.NewTeamBName,
		ServiceAssignment: a.ServiceAssignment,
		TeamAName:         a.TeamAName,
		TeamBName:         a.TeamBName,
		NewTeamName:       a.NewTeamName,
		CapabilityName:    a.CapabilityName,
		OwnerTeamName:     a.OwnerTeamName,
		Visibility:        a.Visibility,
		Role:              a.Role,
		SourceTeamName:    a.SourceTeamName,
		TargetTeamName:    a.TargetTeamName,
		InteractionMode:   a.InteractionMode,
		TeamName:          a.TeamName,
		NewSize:           a.NewSize,
		TeamType:          a.TeamType,
		Description:       a.Description,
		NewServiceName:    a.NewServiceName,
		NeedName:          a.NeedName,
		Outcome:           a.Outcome,
		SupportedBy:       a.SupportedBy,
		ActorName:         a.ActorName,
		DependsOnService:  a.DependsOnService,
		EntityType:        a.EntityType,
		EntityName:        a.EntityName,
	}
}

// fromChangeActionDTO maps a DTO representation to the domain ChangeAction.
func fromChangeActionDTO(d changeActionDTO) entity.ChangeAction {
	return entity.ChangeAction{
		Type:              entity.ChangeActionType(d.Type),
		ServiceName:       d.ServiceName,
		FromTeamName:      d.FromTeamName,
		ToTeamName:        d.ToTeamName,
		OriginalTeamName:  d.OriginalTeamName,
		NewTeamAName:      d.NewTeamAName,
		NewTeamBName:      d.NewTeamBName,
		ServiceAssignment: d.ServiceAssignment,
		TeamAName:         d.TeamAName,
		TeamBName:         d.TeamBName,
		NewTeamName:       d.NewTeamName,
		CapabilityName:    d.CapabilityName,
		OwnerTeamName:     d.OwnerTeamName,
		Visibility:        d.Visibility,
		Role:              d.Role,
		SourceTeamName:    d.SourceTeamName,
		TargetTeamName:    d.TargetTeamName,
		InteractionMode:   d.InteractionMode,
		TeamName:          d.TeamName,
		NewSize:           d.NewSize,
		TeamType:          d.TeamType,
		Description:       d.Description,
		NewServiceName:    d.NewServiceName,
		NeedName:          d.NeedName,
		Outcome:           d.Outcome,
		SupportedBy:       d.SupportedBy,
		ActorName:         d.ActorName,
		DependsOnService:  d.DependsOnService,
		EntityType:        d.EntityType,
		EntityName:        d.EntityName,
	}
}

// toChangesetDTO maps a domain Changeset to its DTO representation.
func toChangesetDTO(cs entity.Changeset) changesetDTO {
	actions := make([]changeActionDTO, len(cs.Actions))
	for i, a := range cs.Actions {
		actions[i] = toChangeActionDTO(a)
	}
	return changesetDTO{
		ID:          cs.ID,
		Description: cs.Description,
		Actions:     actions,
	}
}
