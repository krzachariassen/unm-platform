package analyzer

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ServiceComplexity holds the complexity breakdown for a single service.
type ServiceComplexity struct {
	Service         *entity.Service
	DependencyScore int // fan-out + fan-in (excluding self-loops)
	CapabilityScore int // count of capabilities this service realizes (via service.Realizes)
	DataAssetScore  int // count of data assets this service is involved in
	TotalComplexity int // DependencyScore*2 + CapabilityScore*3 + DataAssetScore*2
}

// ComplexityReport holds per-service complexity scores ranked by TotalComplexity desc, then name asc.
type ComplexityReport struct {
	Services []ServiceComplexity
}

// ComplexityAnalyzer computes per-service complexity scores.
type ComplexityAnalyzer struct{}

// NewComplexityAnalyzer constructs a ComplexityAnalyzer.
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{}
}

// Analyze computes complexity scores for all services in m and returns a ComplexityReport.
func (a *ComplexityAnalyzer) Analyze(m *entity.UNMModel) ComplexityReport {
	if len(m.Services) == 0 {
		return ComplexityReport{}
	}

	// Compute fan-in counts: how many other services declare a dependency on each service name.
	fanIn := make(map[string]int, len(m.Services))
	for name, svc := range m.Services {
		for _, rel := range svc.DependsOn {
			target := rel.TargetID.String()
			if target != name { // exclude self-loops
				fanIn[target]++
			}
		}
	}

	results := make([]ServiceComplexity, 0, len(m.Services))

	for _, svc := range m.Services {
		// DependencyScore: fan-out (outbound, no self-loops) + fan-in.
		fanOut := 0
		for _, rel := range svc.DependsOn {
			if rel.TargetID.String() != svc.Name {
				fanOut++
			}
		}
		depScore := fanOut + fanIn[svc.Name]

		// CapabilityScore: capabilities realized by this service (via GetCapabilitiesForService / service.Realizes).
		capScore := len(m.GetCapabilitiesForService(svc.Name))

		// DataAssetScore: data assets involving this service.
		daScore := len(m.GetDataAssetsForService(svc.Name))

		total := depScore*2 + capScore*3 + daScore*2

		results = append(results, ServiceComplexity{
			Service:         svc,
			DependencyScore: depScore,
			CapabilityScore: capScore,
			DataAssetScore:  daScore,
			TotalComplexity: total,
		})
	}

	// Rank: highest TotalComplexity first; break ties by name ascending.
	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalComplexity != results[j].TotalComplexity {
			return results[i].TotalComplexity > results[j].TotalComplexity
		}
		return results[i].Service.Name < results[j].Service.Name
	})

	return ComplexityReport{Services: results}
}
