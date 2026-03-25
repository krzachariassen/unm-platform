package analyzer

import (
	"sort"

	"github.com/uber/unm-platform/internal/domain/entity"
)

// ServiceBottleneck holds fan-in / fan-out metrics for a single service.
type ServiceBottleneck struct {
	Service    *entity.Service
	FanIn      int  // count of other services that list this service in their DependsOn
	FanOut     int  // count of services this service depends on (excluding self-loops)
	IsCritical bool // fan-in > 10
	IsWarning  bool // fan-in > 5 and not critical (i.e. 6–10)
}

// ExternalDepBottleneck holds fan-in metrics for a single external dependency.
// ServiceCount is the number of internal services that use this external dependency.
// IsCritical: ServiceCount >= 5. IsWarning: ServiceCount >= 3 and < 5.
// Only deps with ServiceCount >= 3 are included in the report.
type ExternalDepBottleneck struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	ServiceCount int      `json:"service_count"`
	Services     []string `json:"services"`
	IsCritical   bool     `json:"is_critical"`
	IsWarning    bool     `json:"is_warning"`
}

// BottleneckReport holds the result of a bottleneck analysis pass over a UNMModel.
type BottleneckReport struct {
	// ServiceBottlenecks contains all services ranked by FanIn descending, then name ascending.
	ServiceBottlenecks []ServiceBottleneck
	// ExternalDependencyBottlenecks contains external deps with ServiceCount >= 3,
	// ranked by ServiceCount descending, then name ascending.
	ExternalDependencyBottlenecks []ExternalDepBottleneck
}

// BottleneckAnalyzer detects high fan-in and fan-out services in the service dependency graph.
type BottleneckAnalyzer struct {
	cfg entity.BottleneckConfig
}

// NewBottleneckAnalyzer constructs a BottleneckAnalyzer.
func NewBottleneckAnalyzer(cfg entity.BottleneckConfig) *BottleneckAnalyzer {
	return &BottleneckAnalyzer{cfg: cfg}
}

// Analyze inspects m and returns a BottleneckReport.
// Fan-in: count of other services that list this service in their DependsOn (self-loops excluded).
// Fan-out: count of services this service depends on (self-loops excluded).
// IsCritical: fan-in > 10.
// IsWarning: fan-in > 5 and not critical (6–10).
// All services are included, ranked by fan-in descending, then name ascending.
func (a *BottleneckAnalyzer) Analyze(m *entity.UNMModel) BottleneckReport {
	var report BottleneckReport

	if len(m.Services) > 0 {
		// Count fan-in for each service (how many other services depend on it).
		fanIn := make(map[string]int, len(m.Services))
		// Initialise every service to 0 so all services appear in the report.
		for name := range m.Services {
			fanIn[name] = 0
		}

		// Count fan-out for each service.
		fanOut := make(map[string]int, len(m.Services))

		for name, svc := range m.Services {
			for _, rel := range svc.DependsOn {
				target := rel.TargetID.String()
				if target == name {
					// Self-loop: skip for both fan-in and fan-out.
					continue
				}
				fanOut[name]++
				fanIn[target]++
			}
		}

		bottlenecks := make([]ServiceBottleneck, 0, len(m.Services))
		for name, svc := range m.Services {
			fi := fanIn[name]
			fo := fanOut[name]
			bottlenecks = append(bottlenecks, ServiceBottleneck{
				Service:    svc,
				FanIn:      fi,
				FanOut:     fo,
				IsCritical: fi > a.cfg.FanInCritical,
				IsWarning:  fi > a.cfg.FanInWarning && fi <= a.cfg.FanInCritical,
			})
		}

		// Sort: fan-in descending, then name ascending for ties.
		sort.Slice(bottlenecks, func(i, j int) bool {
			if bottlenecks[i].FanIn != bottlenecks[j].FanIn {
				return bottlenecks[i].FanIn > bottlenecks[j].FanIn
			}
			return bottlenecks[i].Service.Name < bottlenecks[j].Service.Name
		})

		report.ServiceBottlenecks = bottlenecks
	}

	// External dependency fan-in: count how many services use each external dep.
	// Thresholds: IsCritical if count >= 5, IsWarning if count >= 3 and < 5.
	// Only include deps with count >= 3.
	const extDepCritical = 5
	const extDepWarning = 3

	extBottlenecks := make([]ExternalDepBottleneck, 0)
	for _, dep := range m.ExternalDependencies {
		count := len(dep.UsedBy)
		if count < extDepWarning {
			continue
		}
		services := make([]string, 0, count)
		for _, usage := range dep.UsedBy {
			services = append(services, usage.ServiceName)
		}
		sort.Strings(services)
		isCritical := count >= extDepCritical
		extBottlenecks = append(extBottlenecks, ExternalDepBottleneck{
			Name:         dep.Name,
			Description:  dep.Description,
			ServiceCount: count,
			Services:     services,
			IsCritical:   isCritical,
			IsWarning:    !isCritical,
		})
	}

	// Sort: service count descending, then name ascending for ties.
	sort.Slice(extBottlenecks, func(i, j int) bool {
		if extBottlenecks[i].ServiceCount != extBottlenecks[j].ServiceCount {
			return extBottlenecks[i].ServiceCount > extBottlenecks[j].ServiceCount
		}
		return extBottlenecks[i].Name < extBottlenecks[j].Name
	})

	report.ExternalDependencyBottlenecks = extBottlenecks

	return report
}
