package analyzer

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// coupling_test.go helpers use the "coupling" prefix to avoid name collisions
// with helpers declared in other test files in the same package.

func couplingMustDataAsset(t *testing.T, id, name, assetType string) *entity.DataAsset {
	t.Helper()
	da, err := entity.NewDataAsset(id, name, assetType, "")
	if err != nil {
		t.Fatalf("NewDataAsset %q: %v", name, err)
	}
	return da
}

func couplingMustService(t *testing.T, name, ownerTeam string) *entity.Service {
	t.Helper()
	svc, err := entity.NewService(name, name, "", ownerTeam)
	if err != nil {
		t.Fatalf("NewService %q: %v", name, err)
	}
	return svc
}

func couplingMustTeam(t *testing.T, name string) *entity.Team {
	t.Helper()
	team, err := entity.NewTeam(name, name, "", valueobject.StreamAligned)
	if err != nil {
		t.Fatalf("NewTeam %q: %v", name, err)
	}
	return team
}

func couplingAddService(t *testing.T, m *entity.UNMModel, svc *entity.Service) {
	t.Helper()
	if err := m.AddService(svc); err != nil {
		t.Fatalf("AddService %q: %v", svc.Name, err)
	}
}

func couplingAddTeam(t *testing.T, m *entity.UNMModel, team *entity.Team) {
	t.Helper()
	if err := m.AddTeam(team); err != nil {
		t.Fatalf("AddTeam %q: %v", team.Name, err)
	}
}

func couplingAddDataAsset(t *testing.T, m *entity.UNMModel, da *entity.DataAsset) {
	t.Helper()
	if err := m.AddDataAsset(da); err != nil {
		t.Fatalf("AddDataAsset %q: %v", da.Name, err)
	}
}

// Test 1: Empty model → empty CouplingReport
func TestCouplingAnalyzer_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("empty", "")
	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 0 {
		t.Errorf("expected no couplings for empty model, got %d", len(report.DataAssetCouplings))
	}
}

// Test 2: Data asset used by only 1 service → not in report
func TestCouplingAnalyzer_SingleServiceNotReported(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	svcA := couplingMustService(t, "svc-a", "team-alpha")
	couplingAddService(t, m, svcA)

	da := couplingMustDataAsset(t, "db-1", "db-1", entity.TypeDatabase)
	da.AddUsedBy("svc-a", "read-write")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 0 {
		t.Errorf("expected no couplings when only 1 service uses the asset, got %d", len(report.DataAssetCouplings))
	}
}

// Test 3: Data asset used by 2 services from same team → IsCrossteam=false
func TestCouplingAnalyzer_SameTeamNotCrossteam(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	team := couplingMustTeam(t, "team-alpha")
	couplingAddTeam(t, m, team)

	svcA := couplingMustService(t, "svc-a", "team-alpha")
	svcB := couplingMustService(t, "svc-b", "team-alpha")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)

	da := couplingMustDataAsset(t, "db-1", "db-1", entity.TypeDatabase)
	da.AddUsedBy("svc-a", "read-write")
	da.AddUsedBy("svc-b", "read")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling, got %d", len(report.DataAssetCouplings))
	}
	c := report.DataAssetCouplings[0]
	if c.IsCrossteam {
		t.Errorf("expected IsCrossteam=false for same-team services, got true")
	}
	if c.DataAsset.Name != "db-1" {
		t.Errorf("expected asset name db-1, got %q", c.DataAsset.Name)
	}
	if len(c.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(c.Services))
	}
}

// Test 4: Data asset used by 2 services from different teams → IsCrossteam=true
func TestCouplingAnalyzer_DifferentTeamIsCrossteam(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	teamAlpha := couplingMustTeam(t, "team-alpha")
	teamBeta := couplingMustTeam(t, "team-beta")
	couplingAddTeam(t, m, teamAlpha)
	couplingAddTeam(t, m, teamBeta)

	svcA := couplingMustService(t, "svc-a", "team-alpha")
	svcB := couplingMustService(t, "svc-b", "team-beta")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)

	da := couplingMustDataAsset(t, "cache-1", "cache-1", entity.TypeCache)
	da.AddUsedBy("svc-a", "read")
	da.AddUsedBy("svc-b", "write")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling, got %d", len(report.DataAssetCouplings))
	}
	c := report.DataAssetCouplings[0]
	if !c.IsCrossteam {
		t.Errorf("expected IsCrossteam=true for different-team services, got false")
	}
}

// Test 5: Mix of UsedBy + ProducedBy + ConsumedBy → all deduped into Services list
func TestCouplingAnalyzer_DeduplicatesAllSources(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	svcA := couplingMustService(t, "svc-a", "team-alpha")
	svcB := couplingMustService(t, "svc-b", "team-alpha")
	svcC := couplingMustService(t, "svc-c", "team-alpha")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)
	couplingAddService(t, m, svcC)

	da := couplingMustDataAsset(t, "stream-1", "stream-1", entity.TypeEventStream)
	// svc-a appears in UsedBy
	da.AddUsedBy("svc-a", "read-write")
	// svc-b appears in ProducedBy
	da.ProducedBy = "svc-b"
	// svc-c appears in ConsumedBy
	da.ConsumedBy = []string{"svc-c"}
	// svc-a also appears in ConsumedBy (duplicate — must be deduped)
	da.ConsumedBy = append(da.ConsumedBy, "svc-a")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling, got %d", len(report.DataAssetCouplings))
	}
	c := report.DataAssetCouplings[0]
	// Should have svc-a, svc-b, svc-c — deduplicated (svc-a was mentioned twice)
	if len(c.Services) != 3 {
		t.Errorf("expected 3 unique services after deduplication, got %d: %v", len(c.Services), c.Services)
	}
}

// Test 6: Data asset used by 3 services from 2 different teams → IsCrossteam=true
func TestCouplingAnalyzer_ThreeServicesTwoTeams(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	teamAlpha := couplingMustTeam(t, "team-alpha")
	teamBeta := couplingMustTeam(t, "team-beta")
	couplingAddTeam(t, m, teamAlpha)
	couplingAddTeam(t, m, teamBeta)

	svcA := couplingMustService(t, "svc-a", "team-alpha")
	svcB := couplingMustService(t, "svc-b", "team-alpha")
	svcC := couplingMustService(t, "svc-c", "team-beta")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)
	couplingAddService(t, m, svcC)

	da := couplingMustDataAsset(t, "db-shared", "db-shared", entity.TypeDatabase)
	da.AddUsedBy("svc-a", "read")
	da.AddUsedBy("svc-b", "read")
	da.AddUsedBy("svc-c", "write")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling, got %d", len(report.DataAssetCouplings))
	}
	c := report.DataAssetCouplings[0]
	if !c.IsCrossteam {
		t.Errorf("expected IsCrossteam=true when services span 2 teams, got false")
	}
	if len(c.Services) != 3 {
		t.Errorf("expected 3 services, got %d", len(c.Services))
	}
}

// Test 7: Multiple data assets → report has all coupling data assets, sorted by name
func TestCouplingAnalyzer_MultipleAssetsSortedByName(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	svcA := couplingMustService(t, "svc-a", "team-alpha")
	svcB := couplingMustService(t, "svc-b", "team-alpha")
	svcC := couplingMustService(t, "svc-c", "team-alpha")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)
	couplingAddService(t, m, svcC)

	// "zebra-db" should sort after "alpha-cache"
	daZebra := couplingMustDataAsset(t, "zebra-db", "zebra-db", entity.TypeDatabase)
	daZebra.AddUsedBy("svc-a", "read")
	daZebra.AddUsedBy("svc-b", "write")
	couplingAddDataAsset(t, m, daZebra)

	daAlpha := couplingMustDataAsset(t, "alpha-cache", "alpha-cache", entity.TypeCache)
	daAlpha.AddUsedBy("svc-b", "read")
	daAlpha.AddUsedBy("svc-c", "read")
	couplingAddDataAsset(t, m, daAlpha)

	// single-user asset — should NOT appear in report
	daSingle := couplingMustDataAsset(t, "solo-index", "solo-index", entity.TypeSearchIndex)
	daSingle.AddUsedBy("svc-a", "read")
	couplingAddDataAsset(t, m, daSingle)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 2 {
		t.Fatalf("expected 2 couplings (solo-index excluded), got %d", len(report.DataAssetCouplings))
	}
	if report.DataAssetCouplings[0].DataAsset.Name != "alpha-cache" {
		t.Errorf("expected first coupling to be alpha-cache (sorted), got %q", report.DataAssetCouplings[0].DataAsset.Name)
	}
	if report.DataAssetCouplings[1].DataAsset.Name != "zebra-db" {
		t.Errorf("expected second coupling to be zebra-db (sorted), got %q", report.DataAssetCouplings[1].DataAsset.Name)
	}
}

// Test 8: Services with empty OwnerTeamName → IsCrossteam only true when 2 non-empty different names exist
func TestCouplingAnalyzer_EmptyOwnerNotCrossteam(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// svc-a has no owner (empty OwnerTeamName), svc-b has an owner
	svcA := couplingMustService(t, "svc-a", "")
	svcB := couplingMustService(t, "svc-b", "team-beta")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)

	da := couplingMustDataAsset(t, "db-1", "db-1", entity.TypeDatabase)
	da.AddUsedBy("svc-a", "read")
	da.AddUsedBy("svc-b", "write")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling, got %d", len(report.DataAssetCouplings))
	}
	c := report.DataAssetCouplings[0]
	// Only svc-b has a non-empty owner — need at least 2 different non-empty owners for cross-team.
	if c.IsCrossteam {
		t.Errorf("expected IsCrossteam=false when only one service has a non-empty owner team, got true")
	}
}

// Test 8b: Both services have empty OwnerTeamName → IsCrossteam=false
func TestCouplingAnalyzer_BothEmptyOwnerNotCrossteam(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	svcA := couplingMustService(t, "svc-a", "")
	svcB := couplingMustService(t, "svc-b", "")
	couplingAddService(t, m, svcA)
	couplingAddService(t, m, svcB)

	da := couplingMustDataAsset(t, "db-1", "db-1", entity.TypeDatabase)
	da.AddUsedBy("svc-a", "read")
	da.AddUsedBy("svc-b", "write")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling, got %d", len(report.DataAssetCouplings))
	}
	if report.DataAssetCouplings[0].IsCrossteam {
		t.Errorf("expected IsCrossteam=false when both services have empty owner team, got true")
	}
}

// Test: service name referenced in data asset but not present in model → included as-is
func TestCouplingAnalyzer_UnknownServiceNamesKeptAsIs(t *testing.T) {
	m := entity.NewUNMModel("sys", "")

	// No services added to model, but data asset references two service names
	da := couplingMustDataAsset(t, "db-1", "db-1", entity.TypeDatabase)
	da.AddUsedBy("ghost-svc-a", "read")
	da.AddUsedBy("ghost-svc-b", "write")
	couplingAddDataAsset(t, m, da)

	a := NewCouplingAnalyzer()
	report := a.Analyze(m)

	if len(report.DataAssetCouplings) != 1 {
		t.Fatalf("expected 1 coupling even for unknown service names, got %d", len(report.DataAssetCouplings))
	}
	c := report.DataAssetCouplings[0]
	if len(c.Services) != 2 {
		t.Errorf("expected 2 service names kept as-is, got %d", len(c.Services))
	}
	// Neither ghost service is in the model → IsCrossteam=false (no non-empty different owners found)
	if c.IsCrossteam {
		t.Errorf("expected IsCrossteam=false for services not in model, got true")
	}
}
