package dsl

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// Transform converts a parsed DSL File AST into a UNMModel domain object.
func Transform(f *File) (*entity.UNMModel, error) {
	systemName := ""
	systemDesc := ""
	if f.System != nil {
		systemName = f.System.Name
		systemDesc = f.System.Description
	}
	if systemName == "" {
		return nil, fmt.Errorf("transform: system name is required")
	}

	model := entity.NewUNMModel(systemName, systemDesc)

	if err := transformActors(model, f.Actors); err != nil {
		return nil, err
	}
	if err := transformNeeds(model, f.Needs); err != nil {
		return nil, err
	}
	if err := transformCapabilities(model, f.Capabilities); err != nil {
		return nil, err
	}
	if err := transformServices(model, f.Services); err != nil {
		return nil, err
	}
	if err := transformTeams(model, f.Teams); err != nil {
		return nil, err
	}
	if err := transformPlatforms(model, f.Platforms); err != nil {
		return nil, err
	}
	if err := transformInteractions(model, f.Interactions); err != nil {
		return nil, err
	}
	if err := transformDataAssets(model, f.DataAssets); err != nil {
		return nil, err
	}
	if err := transformExternalDependencies(model, f.ExternalDependencies); err != nil {
		return nil, err
	}
	if err := transformSignals(model, f.Signals); err != nil {
		return nil, err
	}
	if err := transformInferredMappings(model, f.InferredMappings); err != nil {
		return nil, err
	}
	transformTransitions(model, f.Transitions)

	return model, nil
}

func transformActors(model *entity.UNMModel, nodes []*ActorNode) error {
	for _, n := range nodes {
		actor, err := entity.NewActor(n.Name, n.Name, n.Description)
		if err != nil {
			return fmt.Errorf("transform: actor %q: %w", n.Name, err)
		}
		if err := model.AddActor(&actor); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformNeeds(model *entity.UNMModel, nodes []*NeedNode) error {
	for _, n := range nodes {
		outcome := n.Outcome
		if outcome == "" {
			outcome = n.Description
		}
		need, err := entity.NewNeed(n.Name, n.Name, n.Actor, outcome)
		if err != nil {
			return fmt.Errorf("transform: need %q: %w", n.Name, err)
		}
		for _, rel := range n.SupportedBy {
			r, err := buildEntityRelationship(rel)
			if err != nil {
				return fmt.Errorf("transform: need %q supportedBy: %w", n.Name, err)
			}
			need.AddSupportedBy(r)
		}
		if err := model.AddNeed(need); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformCapabilities(model *entity.UNMModel, nodes []*CapabilityNode) error {
	for _, n := range nodes {
		cap, err := buildEntityCapability(n)
		if err != nil {
			return err
		}
		if err := model.AddCapability(cap); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func buildEntityCapability(n *CapabilityNode) (*entity.Capability, error) {
	cap, err := entity.NewCapability(n.Name, n.Name, n.Description)
	if err != nil {
		return nil, fmt.Errorf("transform: capability %q: %w", n.Name, err)
	}
	if n.Visibility != "" {
		if err := cap.SetVisibility(n.Visibility); err != nil {
			return nil, fmt.Errorf("transform: capability %q: %w", n.Name, err)
		}
	}
	for _, child := range n.Children {
		childCap, err := buildEntityCapability(child)
		if err != nil {
			return nil, err
		}
		cap.AddChild(childCap)
	}
	for _, rel := range n.RealizedBy {
		r, err := buildEntityRelationship(rel)
		if err != nil {
			return nil, fmt.Errorf("transform: capability %q realizedBy: %w", n.Name, err)
		}
		cap.AddRealizedBy(r)
	}
	for _, rel := range n.DependsOn {
		r, err := buildEntityRelationship(rel)
		if err != nil {
			return nil, fmt.Errorf("transform: capability %q dependsOn: %w", n.Name, err)
		}
		cap.AddDependsOn(r)
	}
	return cap, nil
}

func transformServices(model *entity.UNMModel, nodes []*ServiceNode) error {
	for _, n := range nodes {
		svc, err := entity.NewService(n.Name, n.Name, n.Description, n.OwnedBy)
		if err != nil {
			return fmt.Errorf("transform: service %q: %w", n.Name, err)
		}
		for _, rel := range n.DependsOn {
			r, err := buildEntityRelationship(rel)
			if err != nil {
				return fmt.Errorf("transform: service %q dependsOn: %w", n.Name, err)
			}
			svc.AddDependsOn(r)
		}
		if err := model.AddService(svc); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformTeams(model *entity.UNMModel, nodes []*TeamNode) error {
	for _, n := range nodes {
		teamType, err := valueobject.NewTeamType(n.Type)
		if err != nil {
			return fmt.Errorf("transform: team %q type: %w", n.Name, err)
		}
		team, err := entity.NewTeam(n.Name, n.Name, n.Description, teamType)
		if err != nil {
			return fmt.Errorf("transform: team %q: %w", n.Name, err)
		}
		if n.Size > 0 {
			team.Size = n.Size
			team.SizeExplicit = true
		}
		for _, owns := range n.Owns {
			targetID, err := valueobject.NewEntityID(owns)
			if err != nil {
				return fmt.Errorf("transform: team %q owns %q: %w", n.Name, owns, err)
			}
			team.AddOwns(entity.NewRelationship(targetID, "", valueobject.RelationshipRole("")))
		}
		if err := model.AddTeam(team); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformPlatforms(model *entity.UNMModel, nodes []*PlatformNode) error {
	for _, n := range nodes {
		platform, err := entity.NewPlatform(n.Name, n.Name, n.Description)
		if err != nil {
			return fmt.Errorf("transform: platform %q: %w", n.Name, err)
		}
		for _, teamName := range n.Teams {
			platform.AddTeam(teamName)
		}
		if err := model.AddPlatform(platform); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformInteractions(model *entity.UNMModel, nodes []*InteractionNode) error {
	for i, n := range nodes {
		mode, err := valueobject.NewInteractionMode(n.Mode)
		if err != nil {
			return fmt.Errorf("transform: interaction[%d] mode: %w", i, err)
		}
		id := fmt.Sprintf("%s->%s", n.From, n.To)
		interaction, err := entity.NewInteraction(id, n.From, n.To, mode, n.Via, n.Description)
		if err != nil {
			return fmt.Errorf("transform: interaction[%d]: %w", i, err)
		}
		model.AddInteraction(interaction)
	}
	return nil
}

func transformDataAssets(model *entity.UNMModel, nodes []*DataAssetNode) error {
	for _, n := range nodes {
		da, err := entity.NewDataAsset(n.Name, n.Name, n.Type, n.Description)
		if err != nil {
			return fmt.Errorf("transform: data_asset %q: %w", n.Name, err)
		}
		for _, usage := range n.UsedBy {
			da.AddUsedBy(usage.Target, usage.Access)
		}
		da.ProducedBy = n.ProducedBy
		for _, consumer := range n.ConsumedBy {
			da.ConsumedBy = append(da.ConsumedBy, consumer)
		}
		if err := model.AddDataAsset(da); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformExternalDependencies(model *entity.UNMModel, nodes []*ExternalDependencyNode) error {
	for _, n := range nodes {
		ext, err := entity.NewExternalDependency(n.Name, n.Name, n.Description)
		if err != nil {
			return fmt.Errorf("transform: external_dependency %q: %w", n.Name, err)
		}
		for _, usage := range n.UsedBy {
			ext.AddUsedBy(usage.Target, usage.Description)
		}
		if err := model.AddExternalDependency(ext); err != nil {
			return fmt.Errorf("transform: %w", err)
		}
	}
	return nil
}

func transformSignals(model *entity.UNMModel, nodes []*SignalNode) error {
	for i, n := range nodes {
		severity, err := valueobject.NewSeverity(n.Severity)
		if err != nil {
			return fmt.Errorf("transform: signal[%d] severity: %w", i, err)
		}
		id := fmt.Sprintf("signal-%s-%d", n.OnEntity, i)
		sig, err := entity.NewSignal(id, n.Category, n.OnEntity, n.Description, "", severity)
		if err != nil {
			return fmt.Errorf("transform: signal[%d]: %w", i, err)
		}
		for _, affected := range n.Affected {
			sig.AddAffected(affected)
		}
		model.AddSignal(sig)
	}
	return nil
}

func transformInferredMappings(model *entity.UNMModel, nodes []*InferredMappingNode) error {
	for i, n := range nodes {
		conf, err := valueobject.NewConfidence(n.Confidence, n.Evidence)
		if err != nil {
			return fmt.Errorf("transform: inferred_mapping[%d] confidence: %w", i, err)
		}
		status, err := valueobject.NewMappingStatus(n.Status)
		if err != nil {
			return fmt.Errorf("transform: inferred_mapping[%d] status: %w", i, err)
		}
		id := fmt.Sprintf("inferred-%d", i)
		im, err := entity.NewInferredMapping(id, n.From, n.To, conf, status)
		if err != nil {
			return fmt.Errorf("transform: inferred_mapping[%d]: %w", i, err)
		}
		model.AddInferredMapping(im)
	}
	return nil
}

func transformTransitions(model *entity.UNMModel, nodes []*TransitionNode) {
	for _, n := range nodes {
		tr := &entity.Transition{
			Name:        n.Name,
			Description: n.Description,
		}
		for _, b := range n.Current {
			tr.Current = append(tr.Current, entity.TransitionBinding{
				CapabilityName: b.CapabilityName,
				TeamName:       b.TeamName,
			})
		}
		for _, b := range n.Target {
			tr.Target = append(tr.Target, entity.TransitionBinding{
				CapabilityName: b.CapabilityName,
				TeamName:       b.TeamName,
			})
		}
		for _, s := range n.Steps {
			tr.Steps = append(tr.Steps, entity.TransitionStep{
				Number:          s.Number,
				Label:           s.Label,
				ActionText:      s.ActionText,
				ExpectedOutcome: s.ExpectedOutcome,
			})
		}
		model.AddTransition(tr)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildEntityRelationship(rel RelationshipNode) (entity.Relationship, error) {
	targetID, err := valueobject.NewEntityID(rel.Target)
	if err != nil {
		return entity.Relationship{}, fmt.Errorf("relationship target: %w", err)
	}
	role, err := valueobject.NewRelationshipRole(rel.Role)
	if err != nil {
		return entity.Relationship{}, fmt.Errorf("relationship role: %w", err)
	}
	return entity.NewRelationship(targetID, rel.Description, role), nil
}
