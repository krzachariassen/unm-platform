package analyzer

import "github.com/uber/unm-platform/internal/domain/entity"

// GapReport holds the result of a gap analysis pass over a UNMModel.
type GapReport struct {
	// UnmappedNeeds are needs with no SupportedBy relationships.
	UnmappedNeeds []*entity.Need
	// UnrealizedCapabilities are leaf capabilities with no RealizedBy services.
	UnrealizedCapabilities []*entity.Capability
	// UnownedServices are services with an empty OwnerTeamName.
	UnownedServices []*entity.Service
	// UnneededCapabilities are capabilities not referenced by any need's SupportedBy.
	UnneededCapabilities []*entity.Capability
	// OrphanServices are services not referenced by any capability's RealizedBy.
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
	for _, cap := range m.Capabilities {
		// UnrealizedCapabilities: leaf caps with no RealizedBy.
		if cap.IsLeaf() && len(cap.RealizedBy) == 0 {
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

	// OrphanServices: services not referenced by any capability's RealizedBy.
	report.OrphanServices = m.GetOrphanServices()

	return report
}
