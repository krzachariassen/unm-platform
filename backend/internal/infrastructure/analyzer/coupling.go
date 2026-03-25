package analyzer

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// DataAssetCoupling describes coupling between services mediated by a shared data asset.
type DataAssetCoupling struct {
	DataAsset   *entity.DataAsset
	Services    []string // all service names that touch this asset (deduplicated, sorted)
	IsCrossteam bool     // true if at least two services have different non-empty owner team names
}

// CouplingReport holds the result of a coupling analysis pass.
type CouplingReport struct {
	DataAssetCouplings []DataAssetCoupling // only data assets used by >= 2 services
}

// CouplingAnalyzer detects implicit coupling between services that share data assets.
type CouplingAnalyzer struct{}

// NewCouplingAnalyzer constructs a CouplingAnalyzer.
func NewCouplingAnalyzer() *CouplingAnalyzer {
	return &CouplingAnalyzer{}
}

// Analyze finds all data assets in m that are shared by 2 or more services and
// determines whether the coupling crosses team boundaries.
func (a *CouplingAnalyzer) Analyze(m *entity.UNMModel) CouplingReport {
	var couplings []DataAssetCoupling

	for _, da := range m.DataAssets {
		services := collectServices(da)
		if len(services) < 2 {
			continue
		}
		crossTeam := isCrossteam(services, m)
		couplings = append(couplings, DataAssetCoupling{
			DataAsset:   da,
			Services:    services,
			IsCrossteam: crossTeam,
		})
	}

	// Sort by data asset name for determinism.
	sort.Slice(couplings, func(i, j int) bool {
		return couplings[i].DataAsset.Name < couplings[j].DataAsset.Name
	})

	return CouplingReport{DataAssetCouplings: couplings}
}

// collectServices returns a deduplicated, sorted list of all service names
// that touch the given data asset (from UsedBy, ProducedBy, and ConsumedBy).
func collectServices(da *entity.DataAsset) []string {
	seen := make(map[string]struct{})

	for _, u := range da.UsedBy {
		if u.ServiceName != "" {
			seen[u.ServiceName] = struct{}{}
		}
	}
	if da.ProducedBy != "" {
		seen[da.ProducedBy] = struct{}{}
	}
	for _, name := range da.ConsumedBy {
		if name != "" {
			seen[name] = struct{}{}
		}
	}

	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// isCrossteam returns true if at least two services in the list have different,
// non-empty OwnerTeamName values as found in the model.
func isCrossteam(serviceNames []string, m *entity.UNMModel) bool {
	var firstOwner string
	for _, name := range serviceNames {
		svc, ok := m.Services[name]
		if !ok {
			continue
		}
		owner := svc.OwnerTeamName
		if owner == "" {
			continue
		}
		if firstOwner == "" {
			firstOwner = owner
			continue
		}
		if owner != firstOwner {
			return true
		}
	}
	return false
}
