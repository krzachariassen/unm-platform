package usecase

import (
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/valueobject"
)

// buildQueryTestModel builds a small but representative model for query tests.
//
// Actors:  Merchant, Operator
// Needs:   "Accept Payment" (Merchant→cap-pay), "View Reports" (Operator→cap-rep)
// Capabilities:
//
//	cap-pay    (realized by: svc-pay, svc-gateway; dependsOn: cap-rep)
//	cap-rep    (realized by: svc-rep)
//	cap-orphan (no realizedBy, no services)
//
// Services: svc-pay (team: payments), svc-gateway (team: payments), svc-rep (team: ops), svc-lonely (no team claim)
// Teams:    payments (owns cap-pay), ops (owns cap-rep)
// Interactions: payments→ops (Collaboration)
func buildQueryTestModel(t *testing.T) *entity.UNMModel {
	t.Helper()
	m := entity.NewUNMModel("Test System", "")

	// Actors
	merchant, _ := entity.NewActor("actor-1", "Merchant", "")
	operator, _ := entity.NewActor("actor-2", "Operator", "")
	_ = m.AddActor(&merchant)
	_ = m.AddActor(&operator)

	// Capabilities
	capPay, _ := entity.NewCapability("cap-pay", "cap-pay", "")
	capRep, _ := entity.NewCapability("cap-rep", "cap-rep", "")
	capOrphan, _ := entity.NewCapability("cap-orphan", "cap-orphan", "")

	// cap-pay realized by svc-pay, svc-gateway
	idSvcPay, _ := valueobject.NewEntityID("svc-pay")
	idSvcGateway, _ := valueobject.NewEntityID("svc-gateway")
	capPay.AddRealizedBy(entity.NewRelationship(idSvcPay, "", valueobject.Primary))
	capPay.AddRealizedBy(entity.NewRelationship(idSvcGateway, "", valueobject.Supporting))

	// cap-pay depends on cap-rep
	idCapRep, _ := valueobject.NewEntityID("cap-rep")
	capPay.AddDependsOn(entity.NewRelationship(idCapRep, "", valueobject.RelationshipRole("")))

	// cap-rep realized by svc-rep
	idSvcRep, _ := valueobject.NewEntityID("svc-rep")
	capRep.AddRealizedBy(entity.NewRelationship(idSvcRep, "", valueobject.Primary))

	_ = m.AddCapability(capPay)
	_ = m.AddCapability(capRep)
	_ = m.AddCapability(capOrphan)

	// Needs
	needPay, _ := entity.NewNeed("need-1", "Accept Payment", "Merchant", "Payment done")
	idCapPay, _ := valueobject.NewEntityID("cap-pay")
	needPay.AddSupportedBy(entity.NewRelationship(idCapPay, "", valueobject.Primary))
	_ = m.AddNeed(needPay)

	needRep, _ := entity.NewNeed("need-2", "View Reports", "Operator", "Reports visible")
	idCapRep2, _ := valueobject.NewEntityID("cap-rep")
	needRep.AddSupportedBy(entity.NewRelationship(idCapRep2, "", valueobject.Primary))
	_ = m.AddNeed(needRep)

	// Unmapped need (no SupportedBy)
	needUnmapped, _ := entity.NewNeed("need-3", "Export Data", "Operator", "Export done")
	_ = m.AddNeed(needUnmapped)

	// Services
	svcPay, _ := entity.NewService("svc-pay", "svc-pay", "", "payments")
	svcGateway, _ := entity.NewService("svc-gateway", "svc-gateway", "", "payments")
	svcRep, _ := entity.NewService("svc-rep", "svc-rep", "", "ops")
	svcLonely, _ := entity.NewService("svc-lonely", "svc-lonely", "", "") // no team — but parser won't validate in unit test
	_ = m.AddService(svcPay)
	_ = m.AddService(svcGateway)
	_ = m.AddService(svcRep)
	_ = m.AddService(svcLonely)

	// Teams
	paymentsTeam, _ := entity.NewTeam("team-1", "payments", "", valueobject.StreamAligned)
	idCapPayOwn, _ := valueobject.NewEntityID("cap-pay")
	paymentsTeam.AddOwns(entity.NewRelationship(idCapPayOwn, "", valueobject.RelationshipRole("")))
	_ = m.AddTeam(paymentsTeam)

	opsTeam, _ := entity.NewTeam("team-2", "ops", "", valueobject.StreamAligned)
	idCapRepOwn, _ := valueobject.NewEntityID("cap-rep")
	opsTeam.AddOwns(entity.NewRelationship(idCapRepOwn, "", valueobject.RelationshipRole("")))
	_ = m.AddTeam(opsTeam)

	// Interaction
	interaction, _ := entity.NewInteraction("int-1", "payments", "ops", valueobject.Collaboration, "", "")
	m.AddInteraction(interaction)

	return m
}

func TestQueryEngine_CapabilitiesForActor(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	caps := q.CapabilitiesForActor(m, "Merchant")
	if len(caps) != 1 {
		t.Fatalf("expected 1 capability for Merchant, got %d", len(caps))
	}
	if caps[0].Name != "cap-pay" {
		t.Errorf("expected cap-pay, got %q", caps[0].Name)
	}

	// Operator has 1 mapped need (View Reports → cap-rep) + 1 unmapped need
	caps = q.CapabilitiesForActor(m, "Operator")
	if len(caps) != 1 {
		t.Fatalf("expected 1 capability for Operator, got %d", len(caps))
	}
	if caps[0].Name != "cap-rep" {
		t.Errorf("expected cap-rep, got %q", caps[0].Name)
	}

	// Unknown actor returns empty
	caps = q.CapabilitiesForActor(m, "Ghost")
	if len(caps) != 0 {
		t.Errorf("expected 0 capabilities for unknown actor, got %d", len(caps))
	}
}

func TestQueryEngine_ServicesForCapability_Direct(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	// cap-pay has 2 direct services + 1 transitive (svc-rep via cap-rep dependsOn)
	svcs := q.ServicesForCapability(m, "cap-pay")
	if len(svcs) != 3 {
		t.Fatalf("expected 3 services for cap-pay (2 direct + 1 transitive), got %d", len(svcs))
	}
}

func TestQueryEngine_ServicesForCapability_Transitive(t *testing.T) {
	// cap-a → cap-b (dependsOn) → cap-c
	// cap-c realized by svc-c
	m := entity.NewUNMModel("Transitive", "")

	capA, _ := entity.NewCapability("cap-a", "cap-a", "")
	capB, _ := entity.NewCapability("cap-b", "cap-b", "")
	capC, _ := entity.NewCapability("cap-c", "cap-c", "")

	idCapB, _ := valueobject.NewEntityID("cap-b")
	capA.AddDependsOn(entity.NewRelationship(idCapB, "", valueobject.RelationshipRole("")))
	idCapC, _ := valueobject.NewEntityID("cap-c")
	capB.AddDependsOn(entity.NewRelationship(idCapC, "", valueobject.RelationshipRole("")))
	idSvcC, _ := valueobject.NewEntityID("svc-c")
	capC.AddRealizedBy(entity.NewRelationship(idSvcC, "", valueobject.Primary))

	_ = m.AddCapability(capA)
	_ = m.AddCapability(capB)
	_ = m.AddCapability(capC)

	svcC, _ := entity.NewService("svc-c", "svc-c", "", "team-c")
	_ = m.AddService(svcC)

	q := NewQueryEngine()
	svcs := q.ServicesForCapability(m, "cap-a")
	// cap-a has no direct services, but transitively via cap-b → cap-c → svc-c
	if len(svcs) != 1 {
		t.Fatalf("expected 1 transitive service, got %d", len(svcs))
	}
	if svcs[0].Name != "svc-c" {
		t.Errorf("expected svc-c, got %q", svcs[0].Name)
	}
}

func TestQueryEngine_ServicesForCapability_CycleGuard(t *testing.T) {
	// cap-x depends on cap-y, cap-y depends on cap-x (cycle)
	m := entity.NewUNMModel("Cycle", "")
	capX, _ := entity.NewCapability("cap-x", "cap-x", "")
	capY, _ := entity.NewCapability("cap-y", "cap-y", "")
	idCapY, _ := valueobject.NewEntityID("cap-y")
	idCapX, _ := valueobject.NewEntityID("cap-x")
	capX.AddDependsOn(entity.NewRelationship(idCapY, "", valueobject.RelationshipRole("")))
	capY.AddDependsOn(entity.NewRelationship(idCapX, "", valueobject.RelationshipRole("")))
	_ = m.AddCapability(capX)
	_ = m.AddCapability(capY)

	q := NewQueryEngine()
	// Should not infinite loop — result may be empty but not panic
	svcs := q.ServicesForCapability(m, "cap-x")
	_ = svcs // nil or empty both acceptable — just must not loop
}

func TestQueryEngine_TeamsForCapability(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	teams := q.TeamsForCapability(m, "cap-pay")
	if len(teams) != 1 || teams[0].Name != "payments" {
		t.Errorf("expected [payments], got %v", teams)
	}

	teams = q.TeamsForCapability(m, "cap-orphan")
	if len(teams) != 0 {
		t.Errorf("expected 0 teams for cap-orphan, got %d", len(teams))
	}
}

func TestQueryEngine_CapabilityDependencyClosure(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	// cap-pay depends on cap-rep
	deps := q.CapabilityDependencyClosure(m, "cap-pay")
	if len(deps) != 1 {
		t.Fatalf("expected 1 transitive dep, got %d", len(deps))
	}
	if deps[0].Name != "cap-rep" {
		t.Errorf("expected cap-rep, got %q", deps[0].Name)
	}
}

func TestQueryEngine_CapabilitiesForTeam(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	caps := q.CapabilitiesForTeam(m, "payments")
	if len(caps) != 1 || caps[0].Name != "cap-pay" {
		t.Errorf("expected [cap-pay], got %v", caps)
	}
}

func TestQueryEngine_OrphanServices(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	orphans := q.OrphanServices(m)
	// svc-lonely is not referenced in any cap.RealizedBy
	names := make(map[string]bool)
	for _, s := range orphans {
		names[s.Name] = true
	}
	if !names["svc-lonely"] {
		t.Errorf("expected svc-lonely to be orphan, got %v", orphans)
	}
	if names["svc-pay"] || names["svc-gateway"] || names["svc-rep"] {
		t.Errorf("expected no false-positive orphans, got %v", orphans)
	}
}

func TestQueryEngine_UnmappedNeeds(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	unmapped := q.UnmappedNeeds(m)
	if len(unmapped) != 1 {
		t.Fatalf("expected 1 unmapped need, got %d", len(unmapped))
	}
	if unmapped[0].Name != "Export Data" {
		t.Errorf("expected 'Export Data', got %q", unmapped[0].Name)
	}
}

func TestQueryEngine_InteractionsForTeam(t *testing.T) {
	m := buildQueryTestModel(t)
	q := NewQueryEngine()

	ints := q.InteractionsForTeam(m, "payments")
	if len(ints) != 1 {
		t.Fatalf("expected 1 interaction for payments, got %d", len(ints))
	}

	ints = q.InteractionsForTeam(m, "ops")
	if len(ints) != 1 {
		t.Fatalf("expected 1 interaction for ops, got %d", len(ints))
	}

	ints = q.InteractionsForTeam(m, "ghost-team")
	if len(ints) != 0 {
		t.Errorf("expected 0 for unknown team, got %d", len(ints))
	}
}
