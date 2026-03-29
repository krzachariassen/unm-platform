package service

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// ChangesetApplier is a domain service that applies a Changeset to a UNMModel,
// producing a new model with changes applied. The original model is never mutated.
type ChangesetApplier struct{}

// NewChangesetApplier constructs a ChangesetApplier.
func NewChangesetApplier() *ChangesetApplier {
	return &ChangesetApplier{}
}

// Apply returns a new model with all changeset actions applied sequentially.
// Returns an error if any action references a non-existent entity.
// The original model m is never mutated.
func (a *ChangesetApplier) Apply(m *entity.UNMModel, cs *entity.Changeset) (*entity.UNMModel, error) {
	result := deepCopyModel(m)

	for i, action := range cs.Actions {
		var err error
		switch action.Type {
		case entity.ActionMoveService:
			err = applyMoveService(result, action)
		case entity.ActionSplitTeam:
			err = applySplitTeam(result, action)
		case entity.ActionMergeTeams:
			err = applyMergeTeams(result, action)
		case entity.ActionAddCapability:
			err = applyAddCapability(result, action)
		case entity.ActionRemoveCapability:
			err = applyRemoveCapability(result, action)
		case entity.ActionReassignCapability:
			err = applyReassignCapability(result, action)
		case entity.ActionAddInteraction:
			err = applyAddInteraction(result, action)
		case entity.ActionRemoveInteraction:
			applyRemoveInteraction(result, action)
		case entity.ActionUpdateTeamSize:
			err = applyUpdateTeamSize(result, action)
		case entity.ActionAddService:
			err = applyAddService(result, action)
		case entity.ActionRemoveService:
			err = applyRemoveService(result, action)
		case entity.ActionRenameService:
			err = applyRenameService(result, action)
		case entity.ActionAddTeam:
			err = applyAddTeam(result, action)
		case entity.ActionRemoveTeam:
			err = applyRemoveTeam(result, action)
		case entity.ActionUpdateTeamType:
			err = applyUpdateTeamType(result, action)
		case entity.ActionAddNeed:
			err = applyAddNeed(result, action)
		case entity.ActionRemoveNeed:
			err = applyRemoveNeed(result, action)
		case entity.ActionAddActor:
			err = applyAddActor(result, action)
		case entity.ActionRemoveActor:
			err = applyRemoveActor(result, action)
		case entity.ActionAddServiceDependency:
			err = applyAddServiceDependency(result, action)
		case entity.ActionRemoveServiceDependency:
			err = applyRemoveServiceDependency(result, action)
		case entity.ActionLinkNeedCapability:
			err = applyLinkNeedCapability(result, action)
		case entity.ActionUnlinkNeedCapability:
			err = applyUnlinkNeedCapability(result, action)
		case entity.ActionLinkCapabilityService:
			err = applyLinkCapabilityService(result, action)
		case entity.ActionUnlinkCapabilityService:
			err = applyUnlinkCapabilityService(result, action)
		case entity.ActionUpdateCapabilityVisibility:
			err = applyUpdateCapabilityVisibility(result, action)
		case entity.ActionUpdateDescription:
			err = applyUpdateDescription(result, action)
		default:
			err = fmt.Errorf("unknown action type: %q", action.Type)
		}
		if err != nil {
			return nil, fmt.Errorf("changeset %s action %d (%s): %w", cs.ID, i, action.Type, err)
		}
	}

	return result, nil
}

func applyMoveService(m *entity.UNMModel, action entity.ChangeAction) error {
	svc, ok := m.Services[action.ServiceName]
	if !ok {
		return fmt.Errorf("service %q not found", action.ServiceName)
	}
	if _, ok := m.Teams[action.ToTeamName]; !ok {
		return fmt.Errorf("target team %q not found", action.ToTeamName)
	}
	svc.OwnerTeamName = action.ToTeamName
	return nil
}

func applySplitTeam(m *entity.UNMModel, action entity.ChangeAction) error {
	original, ok := m.Teams[action.OriginalTeamName]
	if !ok {
		return fmt.Errorf("team %q not found", action.OriginalTeamName)
	}

	teamAID, _ := valueobject.NewEntityID(action.NewTeamAName)
	teamA := &entity.Team{
		ID:            teamAID,
		Name:          action.NewTeamAName,
		TeamType:      original.TeamType,
		Size:          original.Size,
		Owns:          []entity.Relationship{},
		InteractsWith: []entity.TeamInteraction{},
	}

	teamBID, _ := valueobject.NewEntityID(action.NewTeamBName)
	teamB := &entity.Team{
		ID:            teamBID,
		Name:          action.NewTeamBName,
		TeamType:      original.TeamType,
		Size:          original.Size,
		Owns:          []entity.Relationship{},
		InteractsWith: []entity.TeamInteraction{},
	}

	// Build set of services assigned to team B
	teamBServices := make(map[string]bool)
	for svcName, assignment := range action.ServiceAssignment {
		if assignment == "b" {
			teamBServices[svcName] = true
		}
	}

	// Move services
	for _, svc := range m.Services {
		if svc.OwnerTeamName == action.OriginalTeamName {
			if teamBServices[svc.Name] {
				svc.OwnerTeamName = action.NewTeamBName
			} else {
				svc.OwnerTeamName = action.NewTeamAName
			}
		}
	}

	// Transfer capability ownership based on service assignments.
	// For each cap in original team's Owns, check which services realize it.
	// If all realizing services go to B, cap goes to B. Otherwise cap goes to A (default).
	for _, rel := range original.Owns {
		capName := rel.TargetID.String()
		cap, capExists := m.Capabilities[capName]

		goesToB := false
		if capExists && len(cap.RealizedBy) > 0 {
			allB := true
			for _, rb := range cap.RealizedBy {
				svcName := rb.TargetID.String()
				svc, svcExists := m.Services[svcName]
				if !svcExists || svc.OwnerTeamName != action.NewTeamBName {
					allB = false
					break
				}
			}
			goesToB = allB
		}

		if goesToB {
			teamB.Owns = append(teamB.Owns, rel)
		} else {
			teamA.Owns = append(teamA.Owns, rel)
		}
	}

	delete(m.Teams, action.OriginalTeamName)
	m.Teams[action.NewTeamAName] = teamA
	m.Teams[action.NewTeamBName] = teamB

	return nil
}

func applyMergeTeams(m *entity.UNMModel, action entity.ChangeAction) error {
	teamA, okA := m.Teams[action.TeamAName]
	if !okA {
		return fmt.Errorf("team %q not found", action.TeamAName)
	}
	teamB, okB := m.Teams[action.TeamBName]
	if !okB {
		return fmt.Errorf("team %q not found", action.TeamBName)
	}

	newTeamID, _ := valueobject.NewEntityID(action.NewTeamName)
	newTeam := &entity.Team{
		ID:            newTeamID,
		Name:          action.NewTeamName,
		TeamType:      valueobject.StreamAligned,
		Size:          5,
		Owns:          []entity.Relationship{},
		InteractsWith: []entity.TeamInteraction{},
	}

	// Combine Owns from both teams
	newTeam.Owns = append(newTeam.Owns, teamA.Owns...)
	newTeam.Owns = append(newTeam.Owns, teamB.Owns...)

	// Move all services from both teams to new team
	for _, svc := range m.Services {
		if svc.OwnerTeamName == action.TeamAName || svc.OwnerTeamName == action.TeamBName {
			svc.OwnerTeamName = action.NewTeamName
		}
	}

	delete(m.Teams, action.TeamAName)
	delete(m.Teams, action.TeamBName)
	m.Teams[action.NewTeamName] = newTeam

	return nil
}

func applyAddCapability(m *entity.UNMModel, action entity.ChangeAction) error {
	capID, _ := valueobject.NewEntityID(action.CapabilityName)
	cap := &entity.Capability{
		ID:         capID,
		Name:       action.CapabilityName,
		Children:   []*entity.Capability{},
		RealizedBy: []entity.Relationship{},
		DependsOn:  []entity.Relationship{},
	}
	cap.DecomposesTo = cap.Children
	m.Capabilities[action.CapabilityName] = cap

	if action.OwnerTeamName != "" {
		team, ok := m.Teams[action.OwnerTeamName]
		if ok {
			team.Owns = append(team.Owns, entity.NewRelationship(capID, "", ""))
		}
	}

	return nil
}

func applyRemoveCapability(m *entity.UNMModel, action entity.ChangeAction) error {
	delete(m.Capabilities, action.CapabilityName)
	delete(m.CapabilityParents, action.CapabilityName)

	// Remove from all teams' Owns
	for _, team := range m.Teams {
		filtered := make([]entity.Relationship, 0, len(team.Owns))
		for _, rel := range team.Owns {
			if rel.TargetID.String() != action.CapabilityName {
				filtered = append(filtered, rel)
			}
		}
		team.Owns = filtered
	}

	// Remove from all needs' SupportedBy
	for _, need := range m.Needs {
		filtered := make([]entity.Relationship, 0, len(need.SupportedBy))
		for _, rel := range need.SupportedBy {
			if rel.TargetID.String() != action.CapabilityName {
				filtered = append(filtered, rel)
			}
		}
		need.SupportedBy = filtered
	}

	return nil
}

func applyReassignCapability(m *entity.UNMModel, action entity.ChangeAction) error {
	fromTeam, ok := m.Teams[action.FromTeamName]
	if !ok {
		return fmt.Errorf("from team %q not found", action.FromTeamName)
	}
	toTeam, ok := m.Teams[action.ToTeamName]
	if !ok {
		return fmt.Errorf("to team %q not found", action.ToTeamName)
	}

	// Remove from fromTeam
	filtered := make([]entity.Relationship, 0, len(fromTeam.Owns))
	for _, rel := range fromTeam.Owns {
		if rel.TargetID.String() != action.CapabilityName {
			filtered = append(filtered, rel)
		}
	}
	fromTeam.Owns = filtered

	// Add to toTeam
	capID, _ := valueobject.NewEntityID(action.CapabilityName)
	toTeam.Owns = append(toTeam.Owns, entity.NewRelationship(capID, "", ""))

	return nil
}

func applyAddInteraction(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, ok := m.Teams[action.SourceTeamName]; !ok {
		return fmt.Errorf("source team %q not found", action.SourceTeamName)
	}
	if _, ok := m.Teams[action.TargetTeamName]; !ok {
		return fmt.Errorf("target team %q not found", action.TargetTeamName)
	}

	mode, err := valueobject.NewInteractionMode(action.InteractionMode)
	if err != nil {
		return err
	}

	ixID, _ := valueobject.NewEntityID(fmt.Sprintf("ix-%s-%s", action.SourceTeamName, action.TargetTeamName))
	ix := &entity.Interaction{
		ID:           ixID,
		FromTeamName: action.SourceTeamName,
		ToTeamName:   action.TargetTeamName,
		Mode:         mode,
	}
	m.Interactions = append(m.Interactions, ix)

	return nil
}

func applyRemoveInteraction(m *entity.UNMModel, action entity.ChangeAction) {
	filtered := make([]*entity.Interaction, 0, len(m.Interactions))
	for _, ix := range m.Interactions {
		if (ix.FromTeamName == action.SourceTeamName && ix.ToTeamName == action.TargetTeamName) ||
			(ix.FromTeamName == action.TargetTeamName && ix.ToTeamName == action.SourceTeamName) {
			continue
		}
		filtered = append(filtered, ix)
	}
	m.Interactions = filtered
}

func applyUpdateTeamSize(m *entity.UNMModel, action entity.ChangeAction) error {
	team, ok := m.Teams[action.TeamName]
	if !ok {
		return fmt.Errorf("team %q not found", action.TeamName)
	}
	team.Size = action.NewSize
	team.SizeExplicit = true
	return nil
}

func applyAddService(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, exists := m.Services[action.ServiceName]; exists {
		return fmt.Errorf("service %q already exists", action.ServiceName)
	}
	if _, ok := m.Teams[action.OwnerTeamName]; !ok {
		return fmt.Errorf("owner team %q not found", action.OwnerTeamName)
	}
	svcID, _ := valueobject.NewEntityID(action.ServiceName)
	m.Services[action.ServiceName] = &entity.Service{
		ID:            svcID,
		Name:          action.ServiceName,
		Description:   action.Description,
		OwnerTeamName: action.OwnerTeamName,
		DependsOn:     []entity.Relationship{},
	}
	return nil
}

func applyRemoveService(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, ok := m.Services[action.ServiceName]; !ok {
		return fmt.Errorf("service %q not found", action.ServiceName)
	}
	delete(m.Services, action.ServiceName)
	for _, cap := range m.Capabilities {
		filtered := make([]entity.Relationship, 0, len(cap.RealizedBy))
		for _, rel := range cap.RealizedBy {
			if rel.TargetID.String() != action.ServiceName {
				filtered = append(filtered, rel)
			}
		}
		cap.RealizedBy = filtered
	}
	for _, svc := range m.Services {
		filtered := make([]entity.Relationship, 0, len(svc.DependsOn))
		for _, rel := range svc.DependsOn {
			if rel.TargetID.String() != action.ServiceName {
				filtered = append(filtered, rel)
			}
		}
		svc.DependsOn = filtered
	}
	return nil
}

func applyRenameService(m *entity.UNMModel, action entity.ChangeAction) error {
	svc, ok := m.Services[action.ServiceName]
	if !ok {
		return fmt.Errorf("service %q not found", action.ServiceName)
	}
	if _, exists := m.Services[action.NewServiceName]; exists {
		return fmt.Errorf("service %q already exists", action.NewServiceName)
	}
	newID, _ := valueobject.NewEntityID(action.NewServiceName)
	svc.ID = newID
	svc.Name = action.NewServiceName
	delete(m.Services, action.ServiceName)
	m.Services[action.NewServiceName] = svc
	for _, cap := range m.Capabilities {
		for i, rel := range cap.RealizedBy {
			if rel.TargetID.String() == action.ServiceName {
				cap.RealizedBy[i].TargetID = newID
			}
		}
	}
	for _, other := range m.Services {
		for i, rel := range other.DependsOn {
			if rel.TargetID.String() == action.ServiceName {
				other.DependsOn[i].TargetID = newID
			}
		}
	}
	return nil
}

func applyAddTeam(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, exists := m.Teams[action.TeamName]; exists {
		return fmt.Errorf("team %q already exists", action.TeamName)
	}
	tt, err := valueobject.NewTeamType(action.TeamType)
	if err != nil {
		return err
	}
	teamID, _ := valueobject.NewEntityID(action.TeamName)
	size := action.NewSize
	if size <= 0 {
		size = 5
	}
	m.Teams[action.TeamName] = &entity.Team{
		ID:            teamID,
		Name:          action.TeamName,
		Description:   action.Description,
		TeamType:      tt,
		Size:          size,
		Owns:          []entity.Relationship{},
		InteractsWith: []entity.TeamInteraction{},
	}
	return nil
}

func applyRemoveTeam(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, ok := m.Teams[action.TeamName]; !ok {
		return fmt.Errorf("team %q not found", action.TeamName)
	}
	for _, svc := range m.Services {
		if svc.OwnerTeamName == action.TeamName {
			return fmt.Errorf("team %q still owns service %q", action.TeamName, svc.Name)
		}
	}
	delete(m.Teams, action.TeamName)
	filtered := make([]*entity.Interaction, 0, len(m.Interactions))
	for _, ix := range m.Interactions {
		if ix.FromTeamName != action.TeamName && ix.ToTeamName != action.TeamName {
			filtered = append(filtered, ix)
		}
	}
	m.Interactions = filtered
	return nil
}

func applyUpdateTeamType(m *entity.UNMModel, action entity.ChangeAction) error {
	team, ok := m.Teams[action.TeamName]
	if !ok {
		return fmt.Errorf("team %q not found", action.TeamName)
	}
	tt, err := valueobject.NewTeamType(action.TeamType)
	if err != nil {
		return err
	}
	team.TeamType = tt
	return nil
}

func applyAddNeed(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, exists := m.Needs[action.NeedName]; exists {
		return fmt.Errorf("need %q already exists", action.NeedName)
	}
	if _, ok := m.Actors[action.ActorName]; !ok {
		return fmt.Errorf("actor %q not found", action.ActorName)
	}
	needID, _ := valueobject.NewEntityID(action.NeedName)
	need := &entity.Need{
		ID:          needID,
		Name:        action.NeedName,
		ActorName:   action.ActorName,
		Outcome:     action.Outcome,
		SupportedBy: []entity.Relationship{},
	}
	for _, capName := range action.SupportedBy {
		if _, ok := m.Capabilities[capName]; !ok {
			return fmt.Errorf("capability %q not found (referenced in supported_by)", capName)
		}
		capID, _ := valueobject.NewEntityID(capName)
		need.SupportedBy = append(need.SupportedBy, entity.NewRelationship(capID, "", ""))
	}
	m.Needs[action.NeedName] = need
	return nil
}

func applyRemoveNeed(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, ok := m.Needs[action.NeedName]; !ok {
		return fmt.Errorf("need %q not found", action.NeedName)
	}
	delete(m.Needs, action.NeedName)
	return nil
}

func applyAddActor(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, exists := m.Actors[action.ActorName]; exists {
		return fmt.Errorf("actor %q already exists", action.ActorName)
	}
	actorID, _ := valueobject.NewEntityID(action.ActorName)
	m.Actors[action.ActorName] = &entity.Actor{
		ID:          actorID,
		Name:        action.ActorName,
		Description: action.Description,
	}
	return nil
}

func applyRemoveActor(m *entity.UNMModel, action entity.ChangeAction) error {
	if _, ok := m.Actors[action.ActorName]; !ok {
		return fmt.Errorf("actor %q not found", action.ActorName)
	}
	for _, need := range m.Needs {
		if need.ActorName == action.ActorName {
			return fmt.Errorf("actor %q still referenced by need %q", action.ActorName, need.Name)
		}
	}
	delete(m.Actors, action.ActorName)
	return nil
}

func applyAddServiceDependency(m *entity.UNMModel, action entity.ChangeAction) error {
	svc, ok := m.Services[action.ServiceName]
	if !ok {
		return fmt.Errorf("service %q not found", action.ServiceName)
	}
	if _, ok := m.Services[action.DependsOnService]; !ok {
		return fmt.Errorf("dependency target service %q not found", action.DependsOnService)
	}
	depID, _ := valueobject.NewEntityID(action.DependsOnService)
	svc.DependsOn = append(svc.DependsOn, entity.NewRelationship(depID, "", ""))
	return nil
}

func applyRemoveServiceDependency(m *entity.UNMModel, action entity.ChangeAction) error {
	svc, ok := m.Services[action.ServiceName]
	if !ok {
		return fmt.Errorf("service %q not found", action.ServiceName)
	}
	filtered := make([]entity.Relationship, 0, len(svc.DependsOn))
	for _, rel := range svc.DependsOn {
		if rel.TargetID.String() != action.DependsOnService {
			filtered = append(filtered, rel)
		}
	}
	svc.DependsOn = filtered
	return nil
}

func applyLinkNeedCapability(m *entity.UNMModel, action entity.ChangeAction) error {
	need, ok := m.Needs[action.NeedName]
	if !ok {
		return fmt.Errorf("need %q not found", action.NeedName)
	}
	if _, ok := m.Capabilities[action.CapabilityName]; !ok {
		return fmt.Errorf("capability %q not found", action.CapabilityName)
	}
	capID, _ := valueobject.NewEntityID(action.CapabilityName)
	need.SupportedBy = append(need.SupportedBy, entity.NewRelationship(capID, "", ""))
	return nil
}

func applyUnlinkNeedCapability(m *entity.UNMModel, action entity.ChangeAction) error {
	need, ok := m.Needs[action.NeedName]
	if !ok {
		return fmt.Errorf("need %q not found", action.NeedName)
	}
	filtered := make([]entity.Relationship, 0, len(need.SupportedBy))
	for _, rel := range need.SupportedBy {
		if rel.TargetID.String() != action.CapabilityName {
			filtered = append(filtered, rel)
		}
	}
	need.SupportedBy = filtered
	return nil
}

func applyLinkCapabilityService(m *entity.UNMModel, action entity.ChangeAction) error {
	cap, ok := m.Capabilities[action.CapabilityName]
	if !ok {
		return fmt.Errorf("capability %q not found", action.CapabilityName)
	}
	if _, ok := m.Services[action.ServiceName]; !ok {
		return fmt.Errorf("service %q not found", action.ServiceName)
	}
	svcID, _ := valueobject.NewEntityID(action.ServiceName)
	role, err := valueobject.NewRelationshipRole(action.Role)
	if err != nil {
		return err
	}
	cap.RealizedBy = append(cap.RealizedBy, entity.NewRelationship(svcID, "", role))
	return nil
}

func applyUnlinkCapabilityService(m *entity.UNMModel, action entity.ChangeAction) error {
	cap, ok := m.Capabilities[action.CapabilityName]
	if !ok {
		return fmt.Errorf("capability %q not found", action.CapabilityName)
	}
	filtered := make([]entity.Relationship, 0, len(cap.RealizedBy))
	for _, rel := range cap.RealizedBy {
		if rel.TargetID.String() != action.ServiceName {
			filtered = append(filtered, rel)
		}
	}
	cap.RealizedBy = filtered
	return nil
}

func applyUpdateCapabilityVisibility(m *entity.UNMModel, action entity.ChangeAction) error {
	cap, ok := m.Capabilities[action.CapabilityName]
	if !ok {
		return fmt.Errorf("capability %q not found", action.CapabilityName)
	}
	return cap.SetVisibility(action.Visibility)
}

func applyUpdateDescription(m *entity.UNMModel, action entity.ChangeAction) error {
	switch action.EntityType {
	case "actor":
		a, ok := m.Actors[action.EntityName]
		if !ok {
			return fmt.Errorf("actor %q not found", action.EntityName)
		}
		a.Description = action.Description
	case "need":
		n, ok := m.Needs[action.EntityName]
		if !ok {
			return fmt.Errorf("need %q not found", action.EntityName)
		}
		n.Outcome = action.Description
	case "capability":
		c, ok := m.Capabilities[action.EntityName]
		if !ok {
			return fmt.Errorf("capability %q not found", action.EntityName)
		}
		c.Description = action.Description
	case "service":
		s, ok := m.Services[action.EntityName]
		if !ok {
			return fmt.Errorf("service %q not found", action.EntityName)
		}
		s.Description = action.Description
	case "team":
		t, ok := m.Teams[action.EntityName]
		if !ok {
			return fmt.Errorf("team %q not found", action.EntityName)
		}
		t.Description = action.Description
	default:
		return fmt.Errorf("unsupported entity_type %q", action.EntityType)
	}
	return nil
}
