package service

import "github.com/krzachariassen/unm-platform/internal/domain/entity"

// deepCopyModel creates a full structural deep copy of a UNMModel.
func deepCopyModel(m *entity.UNMModel) *entity.UNMModel {
	result := &entity.UNMModel{
		System:           m.System,
		Actors:           make(map[string]*entity.Actor, len(m.Actors)),
		Needs:            make(map[string]*entity.Need, len(m.Needs)),
		Capabilities:     make(map[string]*entity.Capability, len(m.Capabilities)),
		CapabilityParents: make(map[string]string, len(m.CapabilityParents)),
		Services:         make(map[string]*entity.Service, len(m.Services)),
		Teams:            make(map[string]*entity.Team, len(m.Teams)),
		Platforms:        make(map[string]*entity.Platform, len(m.Platforms)),
		Interactions:     make([]*entity.Interaction, 0, len(m.Interactions)),
		Signals:          make([]*entity.Signal, 0, len(m.Signals)),
		DataAssets:       make(map[string]*entity.DataAsset, len(m.DataAssets)),
		ExternalDependencies: make(map[string]*entity.ExternalDependency, len(m.ExternalDependencies)),
		InferredMappings: make([]*entity.InferredMapping, 0, len(m.InferredMappings)),
		Transitions:      make([]*entity.Transition, 0, len(m.Transitions)),
	}

	// Actors (value-like, no internal slices to worry about beyond the struct)
	for k, v := range m.Actors {
		copied := *v
		result.Actors[k] = &copied
	}

	// Needs
	for k, v := range m.Needs {
		copied := *v
		copied.SupportedBy = copyRelationships(v.SupportedBy)
		result.Needs[k] = &copied
	}

	// Capabilities — deep copy including Children tree
	// First pass: copy all capabilities without children references
	capCopies := make(map[string]*entity.Capability, len(m.Capabilities))
	for k, v := range m.Capabilities {
		copied := *v
		copied.RealizedBy = copyRelationships(v.RealizedBy)
		copied.DependsOn = copyRelationships(v.DependsOn)
		copied.Children = nil
		copied.DecomposesTo = nil
		capCopies[k] = &copied
	}
	// Second pass: wire up children using the copies
	for k, v := range m.Capabilities {
		if len(v.Children) > 0 {
			children := make([]*entity.Capability, len(v.Children))
			for i, child := range v.Children {
				children[i] = capCopies[child.Name]
			}
			capCopies[k].Children = children
			capCopies[k].DecomposesTo = children
		} else {
			capCopies[k].Children = []*entity.Capability{}
			capCopies[k].DecomposesTo = capCopies[k].Children
		}
	}
	result.Capabilities = capCopies

	// CapabilityParents
	for k, v := range m.CapabilityParents {
		result.CapabilityParents[k] = v
	}

	// Services
	for k, v := range m.Services {
		copied := *v
		copied.DependsOn = copyRelationships(v.DependsOn)
		result.Services[k] = &copied
	}

	// Teams
	for k, v := range m.Teams {
		copied := *v
		copied.Owns = copyRelationships(v.Owns)
		copied.InteractsWith = make([]entity.TeamInteraction, len(v.InteractsWith))
		copy(copied.InteractsWith, v.InteractsWith)
		result.Teams[k] = &copied
	}

	// Platforms
	for k, v := range m.Platforms {
		copied := *v
		copied.TeamNames = make([]string, len(v.TeamNames))
		copy(copied.TeamNames, v.TeamNames)
		copied.Provides = copyRelationships(v.Provides)
		result.Platforms[k] = &copied
	}

	// Interactions
	for _, v := range m.Interactions {
		copied := *v
		result.Interactions = append(result.Interactions, &copied)
	}

	// Signals
	for _, v := range m.Signals {
		copied := *v
		copied.AffectedEntities = make([]string, len(v.AffectedEntities))
		copy(copied.AffectedEntities, v.AffectedEntities)
		result.Signals = append(result.Signals, &copied)
	}

	// DataAssets
	for k, v := range m.DataAssets {
		copied := *v
		copied.UsedBy = make([]entity.DataAssetServiceUsage, len(v.UsedBy))
		copy(copied.UsedBy, v.UsedBy)
		copied.ConsumedBy = make([]string, len(v.ConsumedBy))
		copy(copied.ConsumedBy, v.ConsumedBy)
		result.DataAssets[k] = &copied
	}

	// ExternalDependencies
	for k, v := range m.ExternalDependencies {
		copied := *v
		copied.UsedBy = make([]entity.ExternalUsage, len(v.UsedBy))
		copy(copied.UsedBy, v.UsedBy)
		result.ExternalDependencies[k] = &copied
	}

	// InferredMappings
	for _, v := range m.InferredMappings {
		copied := *v
		result.InferredMappings = append(result.InferredMappings, &copied)
	}

	// Transitions
	for _, v := range m.Transitions {
		copied := *v
		copied.Current = make([]entity.TransitionBinding, len(v.Current))
		copy(copied.Current, v.Current)
		copied.Target = make([]entity.TransitionBinding, len(v.Target))
		copy(copied.Target, v.Target)
		copied.Steps = make([]entity.TransitionStep, len(v.Steps))
		copy(copied.Steps, v.Steps)
		result.Transitions = append(result.Transitions, &copied)
	}

	return result
}

// copyRelationships creates a new slice with copied Relationship values.
func copyRelationships(rels []entity.Relationship) []entity.Relationship {
	if rels == nil {
		return []entity.Relationship{}
	}
	copied := make([]entity.Relationship, len(rels))
	copy(copied, rels)
	return copied
}
