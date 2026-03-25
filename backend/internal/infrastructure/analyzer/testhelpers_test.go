package analyzer

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// mustTeam creates a Team, fataling on error.
func mustTeam(t *testing.T, name string) *entity.Team {
	t.Helper()
	team, err := entity.NewTeam(name, name, "", valueobject.StreamAligned)
	if err != nil {
		t.Fatalf("NewTeam %q: %v", name, err)
	}
	return team
}

// mustService creates a Service with the given name and owner team, fataling on error.
func mustService(t *testing.T, name, ownerTeam string) *entity.Service {
	t.Helper()
	svc, err := entity.NewService(name, name, "", ownerTeam)
	if err != nil {
		t.Fatalf("NewService %q: %v", name, err)
	}
	return svc
}

// mustDataAsset creates a DataAsset with the given id, name, and asset type, fataling on error.
func mustDataAsset(t *testing.T, id, name, assetType string) *entity.DataAsset {
	t.Helper()
	da, err := entity.NewDataAsset(id, name, assetType, "")
	if err != nil {
		t.Fatalf("NewDataAsset %q: %v", name, err)
	}
	return da
}

// addTeamToModel adds a Team to the model, fataling on error.
func addTeamToModel(t *testing.T, m *entity.UNMModel, team *entity.Team) {
	t.Helper()
	if err := m.AddTeam(team); err != nil {
		t.Fatalf("AddTeam %q: %v", team.Name, err)
	}
}

// addServiceToModel adds a Service to the model, fataling on error.
func addServiceToModel(t *testing.T, m *entity.UNMModel, svc *entity.Service) {
	t.Helper()
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService %q: %v", svc.Name, err)
	}
}

// addDataAssetToModel adds a DataAsset to the model, fataling on error.
func addDataAssetToModel(t *testing.T, m *entity.UNMModel, da *entity.DataAsset) {
	t.Helper()
	if err := m.AddDataAsset(da); err != nil {
		t.Fatalf("AddDataAsset %q: %v", da.Name, err)
	}
}
