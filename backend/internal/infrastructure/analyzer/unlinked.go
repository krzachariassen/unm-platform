package analyzer

import (
	"sort"

	"github.com/uber/unm-platform/internal/domain/entity"
)

// UnlinkedCapability holds a leaf capability that is not referenced by any need's SupportedBy.
type UnlinkedCapability struct {
	Capability *entity.Capability
	Visibility string // resolved visibility level
	IsExpected bool   // true when visibility == "infrastructure" (expected to be unlinked)
}

// UnlinkedCapabilityReport holds the result of an unlinked capability analysis pass.
type UnlinkedCapabilityReport struct {
	// UnlinkedLeafCapabilities are leaf capabilities with no need coverage,
	// sorted by visibility order then name.
	UnlinkedLeafCapabilities []UnlinkedCapability
	// TotalLeafCapabilityCount is the total number of leaf capabilities in the model.
	TotalLeafCapabilityCount int
	// LinkedCount is the number of leaf capabilities directly referenced by at least one need.
	LinkedCount int
	// LinkedPercentage is 100 * LinkedCount / TotalLeafCapabilityCount (0 when no leaves exist).
	LinkedPercentage float64
	// ByVisibility maps visibility level to the count of unlinked capabilities at that level.
	ByVisibility map[string]int
}

// UnlinkedCapabilityAnalyzer detects leaf capabilities not referenced by any need's SupportedBy.
type UnlinkedCapabilityAnalyzer struct{}

// NewUnlinkedCapabilityAnalyzer constructs an UnlinkedCapabilityAnalyzer.
func NewUnlinkedCapabilityAnalyzer() *UnlinkedCapabilityAnalyzer {
	return &UnlinkedCapabilityAnalyzer{}
}

// visOrder gives a stable sort position per visibility level.
var visOrder = map[string]int{
	entity.CapVisibilityUserFacing:     0,
	entity.CapVisibilityDomain:         1,
	entity.CapVisibilityFoundational:   2,
	entity.CapVisibilityInfrastructure: 3,
}

// Analyze returns an UnlinkedCapabilityReport.
// A leaf capability is "unlinked" when no need's SupportedBy directly references it.
// Infrastructure-visibility capabilities are expected to be unlinked (IsExpected = true);
// user-facing, domain, and foundational unlinked capabilities indicate a modeling gap.
func (a *UnlinkedCapabilityAnalyzer) Analyze(m *entity.UNMModel) UnlinkedCapabilityReport {
	// Build set of capability names directly referenced by any need.
	referenced := make(map[string]bool, len(m.Needs)*2)
	for _, need := range m.Needs {
		for _, rel := range need.SupportedBy {
			referenced[rel.TargetID.String()] = true
		}
	}

	byVis := make(map[string]int)
	var unlinked []UnlinkedCapability
	totalLeaf := 0

	for _, cap := range m.Capabilities {
		if !cap.IsLeaf() {
			continue
		}
		totalLeaf++
		if referenced[cap.Name] {
			continue
		}
		vis := cap.Visibility
		if vis == "" {
			vis = entity.CapVisibilityFoundational
		}
		byVis[vis]++
		unlinked = append(unlinked, UnlinkedCapability{
			Capability: cap,
			Visibility: vis,
			IsExpected: vis == entity.CapVisibilityInfrastructure,
		})
	}

	// Sort by visibility order then name for determinism.
	sort.Slice(unlinked, func(i, j int) bool {
		oi := visOrder[unlinked[i].Visibility]
		oj := visOrder[unlinked[j].Visibility]
		if oi != oj {
			return oi < oj
		}
		return unlinked[i].Capability.Name < unlinked[j].Capability.Name
	})

	linkedCount := totalLeaf - len(unlinked)
	linkedPct := 0.0
	if totalLeaf > 0 {
		linkedPct = 100.0 * float64(linkedCount) / float64(totalLeaf)
	}

	return UnlinkedCapabilityReport{
		UnlinkedLeafCapabilities: unlinked,
		TotalLeafCapabilityCount: totalLeaf,
		LinkedCount:              linkedCount,
		LinkedPercentage:         linkedPct,
		ByVisibility:             byVis,
	}
}
