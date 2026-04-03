package analyzer

import "github.com/krzachariassen/unm-platform/internal/domain/entity"

// GapReport holds the result of a gap analysis pass over a UNMModel.
type GapReport struct {
	// UnmappedNeeds are needs with no SupportedBy relationships.
	UnmappedNeeds []*entity.Need
	// UnrealizedCapabilities are leaf capabilities with no services realizing them.
	UnrealizedCapabilities []*entity.Capability
	// UnownedServices are services with an empty OwnerTeamName.
	UnownedServices []*entity.Service
	// UnneededCapabilities are capabilities not referenced by any need's SupportedBy.
	UnneededCapabilities []*entity.Capability
	// OrphanServices are services not realizing any capability.
	OrphanServices []*entity.Service
}

// GapAnalyzer detects structural gaps in a UNMModel.
type GapAnalyzer struct{}

// NewGapAnalyzer constructs a GapAnalyzer.
func NewGapAnalyzer() *GapAnalyzer {
	return &GapAnalyzer{}
}

// Analyze inspects m and returns a GapReport describing every detected gap.
func (a *GapAnalyzer) Analyze(m *entity.UNMModel) GapReport {
	var report GapReport

	// UnmappedNeeds: needs with no SupportedBy.
	for _, need := range m.Needs {
		if !need.IsMapped() {
			report.UnmappedNeeds = append(report.UnmappedNeeds, need)
		}
	}

	// Build the set of capability names directly referenced by needs.
	directlyNeeded := make(map[string]bool)
	for _, need := range m.Needs {
		for _, rel := range need.SupportedBy {
			directlyNeeded[rel.TargetID.String()] = true
		}
	}

	// Expand: any ancestor of a directly-needed capability is also considered needed.
	// This makes the analysis hierarchy-aware: if a need references a leaf capability,
	// its parent group capabilities are implicitly needed too.
	neededCaps := make(map[string]bool)
	for capName := range directlyNeeded {
		neededCaps[capName] = true
		current := capName
		for {
			parent, hasParent := m.CapabilityParents[current]
			if !hasParent {
				break
			}
			neededCaps[parent] = true
			current = parent
		}
	}

	// UnrealizedCapabilities and UnneededCapabilities.
	// Build a set of capability names that have at least one service realizing them.
	realizedCaps := make(map[string]bool)
	for _, svc := range m.Services {
		for _, rel := range svc.Realizes {
			realizedCaps[rel.TargetID.String()] = true
		}
	}
	for _, cap := range m.Capabilities {
		// UnrealizedCapabilities: leaf caps with no services realizing them.
		if cap.IsLeaf() && !realizedCaps[cap.Name] {
			report.UnrealizedCapabilities = append(report.UnrealizedCapabilities, cap)
		}
		// UnneededCapabilities: caps not in the needed set (including via ancestry).
		if !neededCaps[cap.Name] {
			report.UnneededCapabilities = append(report.UnneededCapabilities, cap)
		}
	}

	// UnownedServices: services with empty OwnerTeamName.
	for _, svc := range m.Services {
		if svc.OwnerTeamName == "" {
			report.UnownedServices = append(report.UnownedServices, svc)
		}
	}

	// OrphanServices: services not realizing any capability.
	report.OrphanServices = m.GetOrphanServices()

	return report
}
