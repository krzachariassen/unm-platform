package entity_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// buildTestModel creates a rich UNMModel for use across all query method tests.
//
// Actors:     Merchant, Eater
// Needs:      PayOrder (Merchant), TrackOrder (Merchant), FindFood (Eater)
// Capabilities:
//   - Payments (parent)
//   - PaymentProcessing (child) — realizedBy PaymentService, MultiCapService
//   - FraudDetection (child)    — realizedBy FraudService
//   - OrderManagement             — realizedBy OrderService, MultiCapService
//   - Discovery                   — realizedBy SearchService
//   - OrphanCap (no team, no services in RealizedBy)
//   - BigCap0…BigCap6 (7 caps owned by BigTeam, each realizedBy payment-service placeholder)
//
// Services:
//   - PaymentService  (cap.PaymentProcessing.RealizedBy)
//   - FraudService    (cap.FraudDetection.RealizedBy)
//   - OrderService    (cap.OrderManagement.RealizedBy)
//   - SearchService   (cap.Discovery.RealizedBy)
//   - OrphanService   (not in any cap's RealizedBy)
//   - MultiCapService (cap.PaymentProcessing.RealizedBy + cap.OrderManagement.RealizedBy)
//
// Teams:
//   - PaymentsTeam  owns: PaymentProcessing, FraudDetection (2 caps, normal)
//   - OrderTeam     owns: OrderManagement, PaymentProcessing (2 caps, normal)
//   - BigTeam       owns: BigCap0…BigCap6 + PaymentProcessing (8 caps, overloaded)
//
// Fragmented: PaymentProcessing owned by all 3 teams (>2 = fragmented)
// Overloaded: BigTeam (8 > 6)
//
// Platform:    CorePlatform (contains PaymentsTeam)
// Interactions: 2
// Signal:      1 bottleneck
// InferredMapping: 1
func buildTestModel(t *testing.T) *entity.UNMModel {
	t.Helper()

	m := entity.NewUNMModel("INCA", "INCA Platform")

	// --- Actors ---
	merchant, err := entity.NewActor("actor-1", "Merchant", "A merchant on the platform")
	require.NoError(t, err)
	eater, err := entity.NewActor("actor-2", "Eater", "A customer ordering food")
	require.NoError(t, err)
	require.NoError(t, m.AddActor(&merchant))
	require.NoError(t, m.AddActor(&eater))

	// --- Needs ---
	payOrder, err := entity.NewNeed("need-1", "PayOrder", "Merchant", "Payment completed")
	require.NoError(t, err)
	trackOrder, err := entity.NewNeed("need-2", "TrackOrder", "Merchant", "Order tracked")
	require.NoError(t, err)
	findFood, err := entity.NewNeed("need-3", "FindFood", "Eater", "Restaurant found")
	require.NoError(t, err)
	require.NoError(t, m.AddNeed(payOrder))
	require.NoError(t, m.AddNeed(trackOrder))
	require.NoError(t, m.AddNeed(findFood))

	// --- Shared EntityIDs for relationship building ---
	payProcID, err := valueobject.NewEntityID("PaymentProcessing")
	require.NoError(t, err)
	fraudID, err := valueobject.NewEntityID("FraudDetection")
	require.NoError(t, err)
	orderMgmtID, err := valueobject.NewEntityID("OrderManagement")
	require.NoError(t, err)
	paySvcID, err := valueobject.NewEntityID("PaymentService")
	require.NoError(t, err)
	fraudSvcID, err := valueobject.NewEntityID("FraudService")
	require.NoError(t, err)
	orderSvcID, err := valueobject.NewEntityID("OrderService")
	require.NoError(t, err)
	searchSvcID, err := valueobject.NewEntityID("SearchService")
	require.NoError(t, err)
	multiSvcID, err := valueobject.NewEntityID("MultiCapService")
	require.NoError(t, err)

	// --- Capabilities ---
	payments, err := entity.NewCapability("cap-1", "Payments", "Top-level payments capability")
	require.NoError(t, err)

	payProc, err := entity.NewCapability("cap-2", "PaymentProcessing", "Processes payment transactions")
	require.NoError(t, err)
	payProc.AddRealizedBy(entity.NewRelationship(paySvcID, "", valueobject.Primary))
	payProc.AddRealizedBy(entity.NewRelationship(multiSvcID, "", valueobject.Supporting))

	fraud, err := entity.NewCapability("cap-3", "FraudDetection", "Detects fraudulent transactions")
	require.NoError(t, err)
	fraud.AddRealizedBy(entity.NewRelationship(fraudSvcID, "", valueobject.Primary))

	orderMgmt, err := entity.NewCapability("cap-4", "OrderManagement", "Manages orders end-to-end")
	require.NoError(t, err)
	orderMgmt.AddRealizedBy(entity.NewRelationship(orderSvcID, "", valueobject.Primary))
	orderMgmt.AddRealizedBy(entity.NewRelationship(multiSvcID, "", valueobject.Supporting))

	discovery, err := entity.NewCapability("cap-5", "Discovery", "Restaurant discovery and search")
	require.NoError(t, err)
	discovery.AddRealizedBy(entity.NewRelationship(searchSvcID, "", valueobject.Primary))

	orphanCap, err := entity.NewCapability("cap-6", "OrphanCap", "Capability with no team or service")
	require.NoError(t, err)
	// OrphanCap has no RealizedBy — it triggers ErrLeafCapNoService in validator

	// 7 capabilities for BigTeam to trigger overload (> 6)
	bigCaps := make([]*entity.Capability, 7)
	for i := range 7 {
		bigSvcID, err2 := valueobject.NewEntityID("PaymentService") // reuse PaymentService for simplicity
		require.NoError(t, err2)
		c, err2 := entity.NewCapability(
			fmt.Sprintf("big-cap-%d", i),
			fmt.Sprintf("BigCap%d", i),
			"Big team capability",
		)
		require.NoError(t, err2)
		c.AddRealizedBy(entity.NewRelationship(bigSvcID, "", valueobject.Primary))
		bigCaps[i] = c
	}

	// Nest PaymentProcessing and FraudDetection under Payments
	payments.AddChild(payProc)
	payments.AddChild(fraud)

	// AddCapability registers the parent and recursively all children
	require.NoError(t, m.AddCapability(payments))
	require.NoError(t, m.AddCapability(orderMgmt))
	require.NoError(t, m.AddCapability(discovery))
	require.NoError(t, m.AddCapability(orphanCap))
	for _, bc := range bigCaps {
		require.NoError(t, m.AddCapability(bc))
	}

	// --- Services ---
	paySvc, err := entity.NewService("svc-1", "PaymentService", "Handles payments", "PaymentsTeam")
	require.NoError(t, err)

	fraudSvc, err := entity.NewService("svc-2", "FraudService", "Detects fraud", "PaymentsTeam")
	require.NoError(t, err)

	orderSvc, err := entity.NewService("svc-3", "OrderService", "Manages orders", "OrderTeam")
	require.NoError(t, err)

	searchSvc, err := entity.NewService("svc-4", "SearchService", "Powers search", "OrderTeam")
	require.NoError(t, err)

	// OrphanService: not referenced in any cap's RealizedBy → IsOrphan via model query
	orphanSvc, err := entity.NewService("svc-5", "OrphanService", "No capability support", "PaymentsTeam")
	require.NoError(t, err)

	// MultiCapService referenced in PaymentProcessing.RealizedBy + OrderManagement.RealizedBy
	multiSvc, err := entity.NewService("svc-6", "MultiCapService", "Supports multiple caps", "OrderTeam")
	require.NoError(t, err)

	require.NoError(t, m.AddService(paySvc))
	require.NoError(t, m.AddService(fraudSvc))
	require.NoError(t, m.AddService(orderSvc))
	require.NoError(t, m.AddService(searchSvc))
	require.NoError(t, m.AddService(orphanSvc))
	require.NoError(t, m.AddService(multiSvc))

	// --- Teams ---
	paymentsTeam, err := entity.NewTeam("team-1", "PaymentsTeam", "Handles payment domain", valueobject.StreamAligned)
	require.NoError(t, err)
	paymentsTeam.AddOwns(entity.NewRelationship(payProcID, "", valueobject.Primary))
	paymentsTeam.AddOwns(entity.NewRelationship(fraudID, "", valueobject.Primary))

	orderTeam, err := entity.NewTeam("team-2", "OrderTeam", "Handles order domain", valueobject.StreamAligned)
	require.NoError(t, err)
	orderTeam.AddOwns(entity.NewRelationship(orderMgmtID, "", valueobject.Primary))
	// OrderTeam also owns PaymentProcessing → 2nd team owning it
	orderTeam.AddOwns(entity.NewRelationship(payProcID, "", valueobject.Supporting))

	// BigTeam: 7 bigCaps + PaymentProcessing = 8 total (> 6 → overloaded)
	// PaymentProcessing becomes owned by 3 teams → fragmented
	bigTeam, err := entity.NewTeam("team-3", "BigTeam", "Overloaded team", valueobject.Platform)
	require.NoError(t, err)
	for _, bc := range bigCaps {
		bcID, err2 := valueobject.NewEntityID(bc.Name)
		require.NoError(t, err2)
		bigTeam.AddOwns(entity.NewRelationship(bcID, "", valueobject.Primary))
	}
	bigTeam.AddOwns(entity.NewRelationship(payProcID, "", valueobject.Supporting))

	require.NoError(t, m.AddTeam(paymentsTeam))
	require.NoError(t, m.AddTeam(orderTeam))
	require.NoError(t, m.AddTeam(bigTeam))

	// --- Platform ---
	platform, err := entity.NewPlatform("plat-1", "CorePlatform", "Core shared platform")
	require.NoError(t, err)
	platform.AddTeam("PaymentsTeam")
	require.NoError(t, m.AddPlatform(platform))

	// --- Interactions ---
	i1, err := entity.NewInteraction("int-1", "PaymentsTeam", "OrderTeam", valueobject.XAsAService, "Payments API", "Payments exposes API to Order")
	require.NoError(t, err)
	i2, err := entity.NewInteraction("int-2", "OrderTeam", "PaymentsTeam", valueobject.Collaboration, "", "Joint incident response")
	require.NoError(t, err)
	m.AddInteraction(i1)
	m.AddInteraction(i2)

	// --- Signal ---
	sig, err := entity.NewSignal("sig-1", entity.CategoryBottleneck, "PaymentProcessing", "Payment processing is a bottleneck", "High p99 latency", valueobject.SeverityHigh)
	require.NoError(t, err)
	m.AddSignal(sig)

	// --- InferredMapping ---
	conf, err := valueobject.NewConfidence(0.85, "code scan evidence")
	require.NoError(t, err)
	mapping, err := entity.NewInferredMapping("im-1", "OrphanService", "Discovery", conf, valueobject.Candidate)
	require.NoError(t, err)
	m.AddInferredMapping(mapping)

	return m
}

func TestNewUNMModel(t *testing.T) {
	m := entity.NewUNMModel("TestSystem", "A test system")
	assert.Equal(t, "TestSystem", m.System.Name)
	assert.Equal(t, "A test system", m.System.Description)
	assert.NotNil(t, m.Actors)
	assert.NotNil(t, m.Needs)
	assert.NotNil(t, m.Capabilities)
	assert.NotNil(t, m.CapabilityParents)
	assert.NotNil(t, m.Services)
	assert.NotNil(t, m.Teams)
	assert.NotNil(t, m.Platforms)
	assert.NotNil(t, m.Interactions)
	assert.NotNil(t, m.Signals)
	assert.NotNil(t, m.DataAssets)
	assert.NotNil(t, m.ExternalDependencies)
	assert.NotNil(t, m.InferredMappings)
}

func TestDuplicateDetection(t *testing.T) {
	tests := []struct {
		name string
		fn   func(m *entity.UNMModel) error
	}{
		{
			name: "duplicate actor",
			fn: func(m *entity.UNMModel) error {
				a, _ := entity.NewActor("a1", "Alice", "")
				_ = m.AddActor(&a)
				a2, _ := entity.NewActor("a2", "Alice", "duplicate")
				return m.AddActor(&a2)
			},
		},
		{
			name: "duplicate need",
			fn: func(m *entity.UNMModel) error {
				n, _ := entity.NewNeed("n1", "PayOrder", "Alice", "")
				_ = m.AddNeed(n)
				n2, _ := entity.NewNeed("n2", "PayOrder", "Alice", "dup")
				return m.AddNeed(n2)
			},
		},
		{
			name: "duplicate capability",
			fn: func(m *entity.UNMModel) error {
				c, _ := entity.NewCapability("c1", "Payments", "")
				_ = m.AddCapability(c)
				c2, _ := entity.NewCapability("c2", "Payments", "dup")
				return m.AddCapability(c2)
			},
		},
		{
			name: "duplicate service",
			fn: func(m *entity.UNMModel) error {
				s, _ := entity.NewService("s1", "PaySvc", "", "")
				_ = m.AddService(s)
				s2, _ := entity.NewService("s2", "PaySvc", "dup", "")
				return m.AddService(s2)
			},
		},
		{
			name: "duplicate team",
			fn: func(m *entity.UNMModel) error {
				tt, _ := entity.NewTeam("t1", "TeamA", "", valueobject.StreamAligned)
				_ = m.AddTeam(tt)
				tt2, _ := entity.NewTeam("t2", "TeamA", "dup", valueobject.StreamAligned)
				return m.AddTeam(tt2)
			},
		},
		{
			name: "duplicate platform",
			fn: func(m *entity.UNMModel) error {
				p, _ := entity.NewPlatform("p1", "PlatA", "")
				_ = m.AddPlatform(p)
				p2, _ := entity.NewPlatform("p2", "PlatA", "dup")
				return m.AddPlatform(p2)
			},
		},
		{
			name: "duplicate data asset",
			fn: func(m *entity.UNMModel) error {
				d, _ := entity.NewDataAsset("d1", "DB1", entity.TypeDatabase, "")
				_ = m.AddDataAsset(d)
				d2, _ := entity.NewDataAsset("d2", "DB1", entity.TypeDatabase, "dup")
				return m.AddDataAsset(d2)
			},
		},
		{
			name: "duplicate external dependency",
			fn: func(m *entity.UNMModel) error {
				e, _ := entity.NewExternalDependency("e1", "Stripe", "")
				_ = m.AddExternalDependency(e)
				e2, _ := entity.NewExternalDependency("e2", "Stripe", "dup")
				return m.AddExternalDependency(e2)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := entity.NewUNMModel("Sys", "")
			err := tc.fn(m)
			assert.Error(t, err, "expected duplicate detection error")
		})
	}
}

func TestGetNeedsForActor(t *testing.T) {
	m := buildTestModel(t)

	merchantNeeds := m.GetNeedsForActor("Merchant")
	assert.Len(t, merchantNeeds, 2)

	eaterNeeds := m.GetNeedsForActor("Eater")
	assert.Len(t, eaterNeeds, 1)
	assert.Equal(t, "FindFood", eaterNeeds[0].Name)

	unknown := m.GetNeedsForActor("Unknown")
	assert.Empty(t, unknown)
}

func TestGetCapabilitiesForTeam(t *testing.T) {
	m := buildTestModel(t)

	// PaymentsTeam owns PaymentProcessing + FraudDetection = 2
	payCaps := m.GetCapabilitiesForTeam("PaymentsTeam")
	assert.Len(t, payCaps, 2)
	names := make([]string, len(payCaps))
	for i, c := range payCaps {
		names[i] = c.Name
	}
	assert.ElementsMatch(t, []string{"PaymentProcessing", "FraudDetection"}, names)

	// OrderTeam owns OrderManagement + PaymentProcessing = 2
	orderCaps := m.GetCapabilitiesForTeam("OrderTeam")
	assert.Len(t, orderCaps, 2)

	// BigTeam owns 7 bigCaps + PaymentProcessing = 8
	bigCaps := m.GetCapabilitiesForTeam("BigTeam")
	assert.Len(t, bigCaps, 8)

	// Non-existent team returns nil
	none := m.GetCapabilitiesForTeam("Ghost")
	assert.Nil(t, none)
}

func TestGetServicesForCapability(t *testing.T) {
	m := buildTestModel(t)

	// PaymentProcessing.RealizedBy = PaymentService + MultiCapService
	svcs := m.GetServicesForCapability("PaymentProcessing")
	assert.Len(t, svcs, 2)
	svcNames := make([]string, len(svcs))
	for i, s := range svcs {
		svcNames[i] = s.Name
	}
	assert.ElementsMatch(t, []string{"PaymentService", "MultiCapService"}, svcNames)

	// OrderManagement.RealizedBy = OrderService + MultiCapService
	orderSvcs := m.GetServicesForCapability("OrderManagement")
	assert.Len(t, orderSvcs, 2)

	// Discovery.RealizedBy = SearchService
	discSvcs := m.GetServicesForCapability("Discovery")
	assert.Len(t, discSvcs, 1)
	assert.Equal(t, "SearchService", discSvcs[0].Name)

	// OrphanCap has no RealizedBy entries
	orphanSvcs := m.GetServicesForCapability("OrphanCap")
	assert.Empty(t, orphanSvcs)
}

func TestGetCapabilitiesForService(t *testing.T) {
	m := buildTestModel(t)

	// PaymentService referenced in PaymentProcessing.RealizedBy
	caps := m.GetCapabilitiesForService("PaymentService")
	capNames := make([]string, len(caps))
	for i, c := range caps {
		capNames[i] = c.Name
	}
	// PaymentService is also referenced by BigCap0..6 RealizedBy (all reuse PaymentService ID for simplicity)
	assert.Contains(t, capNames, "PaymentProcessing")

	// MultiCapService referenced in PaymentProcessing + OrderManagement
	multiCaps := m.GetCapabilitiesForService("MultiCapService")
	multiCapNames := make([]string, len(multiCaps))
	for i, c := range multiCaps {
		multiCapNames[i] = c.Name
	}
	assert.Contains(t, multiCapNames, "PaymentProcessing")
	assert.Contains(t, multiCapNames, "OrderManagement")

	// OrphanService not referenced anywhere
	orphanCaps := m.GetCapabilitiesForService("OrphanService")
	assert.Empty(t, orphanCaps)
}

func TestGetTeamsForCapability(t *testing.T) {
	m := buildTestModel(t)

	// PaymentProcessing owned by PaymentsTeam + OrderTeam + BigTeam = 3
	teams := m.GetTeamsForCapability("PaymentProcessing")
	assert.Len(t, teams, 3)
	teamNames := make([]string, len(teams))
	for i, t2 := range teams {
		teamNames[i] = t2.Name
	}
	assert.ElementsMatch(t, []string{"PaymentsTeam", "OrderTeam", "BigTeam"}, teamNames)

	// FraudDetection owned by PaymentsTeam only
	fraudTeams := m.GetTeamsForCapability("FraudDetection")
	assert.Len(t, fraudTeams, 1)
	assert.Equal(t, "PaymentsTeam", fraudTeams[0].Name)

	// OrphanCap owned by nobody
	orphanTeams := m.GetTeamsForCapability("OrphanCap")
	assert.Empty(t, orphanTeams)
}

func TestGetOrphanServices(t *testing.T) {
	m := buildTestModel(t)

	orphans := m.GetOrphanServices()
	assert.Len(t, orphans, 1)
	assert.Equal(t, "OrphanService", orphans[0].Name)
}

func TestGetFragmentedCapabilities(t *testing.T) {
	m := buildTestModel(t)

	fragmented := m.GetFragmentedCapabilities()
	assert.Len(t, fragmented, 1)
	assert.Equal(t, "PaymentProcessing", fragmented[0].Name)
}

func TestGetOverloadedTeams(t *testing.T) {
	m := buildTestModel(t)

	overloaded := m.GetOverloadedTeams(6)
	assert.Len(t, overloaded, 1)
	assert.Equal(t, "BigTeam", overloaded[0].Name)
}

func TestSummary(t *testing.T) {
	m := buildTestModel(t)
	s := m.Summary()

	assert.Equal(t, "INCA", s.SystemName)
	assert.Equal(t, 2, s.ActorCount)
	assert.Equal(t, 3, s.NeedCount)
	// Capabilities: Payments, PaymentProcessing, FraudDetection, OrderManagement,
	//               Discovery, OrphanCap + 7 BigCaps = 13
	assert.Equal(t, 13, s.CapabilityCount)
	assert.Equal(t, 6, s.ServiceCount)
	assert.Equal(t, 3, s.TeamCount)
	assert.Equal(t, 1, s.OrphanServiceCount)
	assert.Equal(t, 1, s.FragmentedCapCount)
	assert.Equal(t, 1, s.OverloadedTeamCount)
}

func TestAddInteractionAndSignalNoDedup(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	i1, err := entity.NewInteraction("i1", "TeamA", "TeamB", valueobject.Collaboration, "", "first")
	require.NoError(t, err)
	i2, err := entity.NewInteraction("i2", "TeamA", "TeamB", valueobject.Collaboration, "", "second")
	require.NoError(t, err)
	m.AddInteraction(i1)
	m.AddInteraction(i2)
	assert.Len(t, m.Interactions, 2, "interactions allow duplicates between same teams")

	sig, err := entity.NewSignal("s1", entity.CategoryBottleneck, "SomeEntity", "desc", "evidence", valueobject.SeverityHigh)
	require.NoError(t, err)
	m.AddSignal(sig)
	m.AddSignal(sig)
	assert.Len(t, m.Signals, 2, "signals allow duplicates")
}

func TestAddCapabilityRegistersChildren(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	parent, err := entity.NewCapability("p1", "Parent", "")
	require.NoError(t, err)
	child, err := entity.NewCapability("c1", "Child", "")
	require.NoError(t, err)
	grandchild, err := entity.NewCapability("gc1", "Grandchild", "")
	require.NoError(t, err)

	child.AddChild(grandchild)
	parent.AddChild(child)

	require.NoError(t, m.AddCapability(parent))

	assert.Contains(t, m.Capabilities, "Parent")
	assert.Contains(t, m.Capabilities, "Child")
	assert.Contains(t, m.Capabilities, "Grandchild")

	// CapabilityParents should record the parent-child relationships
	assert.Equal(t, "Parent", m.CapabilityParents["Child"])
	assert.Equal(t, "Child", m.CapabilityParents["Grandchild"])
	_, hasParent := m.CapabilityParents["Parent"]
	assert.False(t, hasParent, "root capability should not have a parent entry")
}

func TestGetRootCapabilities(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	parent, _ := entity.NewCapability("p1", "Parent", "")
	child, _ := entity.NewCapability("c1", "Child", "")
	standalone, _ := entity.NewCapability("s1", "Standalone", "")

	parent.AddChild(child)
	require.NoError(t, m.AddCapability(parent))
	require.NoError(t, m.AddCapability(standalone))

	roots := m.GetRootCapabilities()
	rootNames := make([]string, len(roots))
	for i, r := range roots {
		rootNames[i] = r.Name
	}
	assert.ElementsMatch(t, []string{"Parent", "Standalone"}, rootNames)
	assert.NotContains(t, rootNames, "Child")
}

func TestGetCapabilityPath(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	root, _ := entity.NewCapability("r1", "Root", "")
	mid, _ := entity.NewCapability("m1", "Mid", "")
	leaf, _ := entity.NewCapability("l1", "Leaf", "")

	mid.AddChild(leaf)
	root.AddChild(mid)
	require.NoError(t, m.AddCapability(root))

	t.Run("path from root", func(t *testing.T) {
		path := m.GetCapabilityPath("Root")
		assert.Equal(t, []string{"Root"}, path)
	})

	t.Run("path from mid", func(t *testing.T) {
		path := m.GetCapabilityPath("Mid")
		assert.Equal(t, []string{"Root", "Mid"}, path)
	})

	t.Run("path from leaf", func(t *testing.T) {
		path := m.GetCapabilityPath("Leaf")
		assert.Equal(t, []string{"Root", "Mid", "Leaf"}, path)
	})

	t.Run("not found returns nil", func(t *testing.T) {
		path := m.GetCapabilityPath("DoesNotExist")
		assert.Nil(t, path)
	})
}

func TestGetCapabilitiesByLayer(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	uf, _ := entity.NewCapability("c1", "Cap1", "")
	_ = uf.SetVisibility(entity.CapVisibilityUserFacing)

	dom, _ := entity.NewCapability("c2", "Cap2", "")
	_ = dom.SetVisibility(entity.CapVisibilityDomain)

	noLayer, _ := entity.NewCapability("c3", "Cap3", "")

	require.NoError(t, m.AddCapability(uf))
	require.NoError(t, m.AddCapability(dom))
	require.NoError(t, m.AddCapability(noLayer))

	userFacing := m.GetCapabilitiesByLayer(entity.CapVisibilityUserFacing)
	assert.Len(t, userFacing, 1)
	assert.Equal(t, "Cap1", userFacing[0].Name)

	domain := m.GetCapabilitiesByLayer(entity.CapVisibilityDomain)
	assert.Len(t, domain, 1)
	assert.Equal(t, "Cap2", domain[0].Name)

	empty := m.GetCapabilitiesByLayer("")
	assert.Len(t, empty, 1)
	assert.Equal(t, "Cap3", empty[0].Name)

	foundational := m.GetCapabilitiesByLayer(entity.CapVisibilityFoundational)
	assert.Empty(t, foundational)
}

func TestGetPlatformForTeam(t *testing.T) {
	m := buildTestModel(t)

	// PaymentsTeam is in CorePlatform
	plat := m.GetPlatformForTeam("PaymentsTeam")
	require.NotNil(t, plat)
	assert.Equal(t, "CorePlatform", plat.Name)

	// OrderTeam is not in any platform
	none := m.GetPlatformForTeam("OrderTeam")
	assert.Nil(t, none)

	// Non-existent team
	ghost := m.GetPlatformForTeam("Ghost")
	assert.Nil(t, ghost)
}

func TestBuildValueChain(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	// Create capabilities at different visibility layers.
	capUF1, err := entity.NewCapability("c1", "Zebra", "user-facing Z")
	require.NoError(t, err)
	require.NoError(t, capUF1.SetVisibility(entity.CapVisibilityUserFacing))

	capUF2, err := entity.NewCapability("c2", "Apple", "user-facing A")
	require.NoError(t, err)
	require.NoError(t, capUF2.SetVisibility(entity.CapVisibilityUserFacing))

	capDomain, err := entity.NewCapability("c3", "DomainCap", "domain cap")
	require.NoError(t, err)
	require.NoError(t, capDomain.SetVisibility(entity.CapVisibilityDomain))

	capInfra, err := entity.NewCapability("c4", "InfraCap", "infra cap")
	require.NoError(t, err)
	require.NoError(t, capInfra.SetVisibility(entity.CapVisibilityInfrastructure))

	capUnset, err := entity.NewCapability("c5", "UnsetCap", "no visibility")
	require.NoError(t, err)
	// leave Visibility as ""

	require.NoError(t, m.AddCapability(capUF1))
	require.NoError(t, m.AddCapability(capUF2))
	require.NoError(t, m.AddCapability(capDomain))
	require.NoError(t, m.AddCapability(capInfra))
	require.NoError(t, m.AddCapability(capUnset))

	layers := m.BuildValueChain()

	// foundational is absent — no capabilities at that layer.
	// Expected order: user-facing, domain, infrastructure, ""
	require.Len(t, layers, 4)

	assert.Equal(t, "user-facing", layers[0].Layer)
	assert.Len(t, layers[0].Capabilities, 2)
	// Within user-facing, sorted by name: Apple before Zebra
	assert.Equal(t, "Apple", layers[0].Capabilities[0].Name)
	assert.Equal(t, "Zebra", layers[0].Capabilities[1].Name)

	assert.Equal(t, "domain", layers[1].Layer)
	assert.Len(t, layers[1].Capabilities, 1)
	assert.Equal(t, "DomainCap", layers[1].Capabilities[0].Name)

	assert.Equal(t, "infrastructure", layers[2].Layer)
	assert.Len(t, layers[2].Capabilities, 1)
	assert.Equal(t, "InfraCap", layers[2].Capabilities[0].Name)

	assert.Equal(t, "", layers[3].Layer)
	assert.Len(t, layers[3].Capabilities, 1)
	assert.Equal(t, "UnsetCap", layers[3].Capabilities[0].Name)
}

func TestBuildValueChain_EmptyModel(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")
	layers := m.BuildValueChain()
	assert.Empty(t, layers)
}

func TestBuildUNMMap(t *testing.T) {
	m := entity.NewUNMModel("Sys", "")

	// Actors
	actorA, err := entity.NewActor("a1", "ActorA", "")
	require.NoError(t, err)
	actorB, err := entity.NewActor("a2", "ActorB", "")
	require.NoError(t, err)
	require.NoError(t, m.AddActor(&actorA))
	require.NoError(t, m.AddActor(&actorB))

	// Capabilities
	capX, err := entity.NewCapability("cap-x", "CapX", "capability X")
	require.NoError(t, err)
	capY, err := entity.NewCapability("cap-y", "CapY", "capability Y")
	require.NoError(t, err)
	capZ, err := entity.NewCapability("cap-z", "CapZ", "capability Z")
	require.NoError(t, err)
	require.NoError(t, m.AddCapability(capX))
	require.NoError(t, m.AddCapability(capY))
	require.NoError(t, m.AddCapability(capZ))

	// EntityIDs for relationships
	capXID, err := valueobject.NewEntityID("CapX")
	require.NoError(t, err)
	capYID, err := valueobject.NewEntityID("CapY")
	require.NoError(t, err)
	capZID, err := valueobject.NewEntityID("CapZ")
	require.NoError(t, err)

	// Needs: 2 for ActorA, 1 for ActorB
	needA1, err := entity.NewNeed("n1", "NeedA1", "ActorA", "outcome a1")
	require.NoError(t, err)
	needA1.AddSupportedBy(entity.NewRelationship(capXID, "", valueobject.Primary))
	needA1.AddSupportedBy(entity.NewRelationship(capYID, "", valueobject.Supporting))

	needA2, err := entity.NewNeed("n2", "NeedA2", "ActorA", "outcome a2")
	require.NoError(t, err)
	needA2.AddSupportedBy(entity.NewRelationship(capZID, "", valueobject.Primary))

	needB1, err := entity.NewNeed("n3", "NeedB1", "ActorB", "outcome b1")
	require.NoError(t, err)
	needB1.AddSupportedBy(entity.NewRelationship(capXID, "", valueobject.Primary))

	require.NoError(t, m.AddNeed(needA1))
	require.NoError(t, m.AddNeed(needA2))
	require.NoError(t, m.AddNeed(needB1))

	t.Run("no filter returns all entries", func(t *testing.T) {
		entries := m.BuildUNMMap("")
		require.Len(t, entries, 3)

		// Sorted by ActorName then NeedName: ActorA/NeedA1, ActorA/NeedA2, ActorB/NeedB1
		assert.Equal(t, "ActorA", entries[0].ActorName)
		assert.Equal(t, "NeedA1", entries[0].NeedName)
		capNames0 := capabilityNames(entries[0].Capabilities)
		assert.ElementsMatch(t, []string{"CapX", "CapY"}, capNames0)

		assert.Equal(t, "ActorA", entries[1].ActorName)
		assert.Equal(t, "NeedA2", entries[1].NeedName)
		capNames1 := capabilityNames(entries[1].Capabilities)
		assert.ElementsMatch(t, []string{"CapZ"}, capNames1)

		assert.Equal(t, "ActorB", entries[2].ActorName)
		assert.Equal(t, "NeedB1", entries[2].NeedName)
		capNames2 := capabilityNames(entries[2].Capabilities)
		assert.ElementsMatch(t, []string{"CapX"}, capNames2)
	})

	t.Run("filter by ActorA returns only 2 entries", func(t *testing.T) {
		entries := m.BuildUNMMap("ActorA")
		require.Len(t, entries, 2)
		for _, e := range entries {
			assert.Equal(t, "ActorA", e.ActorName)
		}
		assert.Equal(t, "NeedA1", entries[0].NeedName)
		assert.Equal(t, "NeedA2", entries[1].NeedName)
	})

	t.Run("filter by ActorB returns 1 entry", func(t *testing.T) {
		entries := m.BuildUNMMap("ActorB")
		require.Len(t, entries, 1)
		assert.Equal(t, "ActorB", entries[0].ActorName)
		assert.Equal(t, "NeedB1", entries[0].NeedName)
	})

	t.Run("filter by unknown actor returns empty", func(t *testing.T) {
		entries := m.BuildUNMMap("Ghost")
		assert.Empty(t, entries)
	})
}

// capabilityNames is a helper to extract names from a capability slice.
func capabilityNames(caps []*entity.Capability) []string {
	names := make([]string, len(caps))
	for i, c := range caps {
		names[i] = c.Name
	}
	return names
}

// ── GetServicesForCapability: additional branch coverage ─────────────────────

func TestGetServicesForCapability_NonExistentCap(t *testing.T) {
	m := buildTestModel(t)
	// Capability that doesn't exist in the map → hits the !ok branch → nil
	svcs := m.GetServicesForCapability("DoesNotExist")
	assert.Nil(t, svcs)
}

func TestGetServicesForCapability_DanglingReference(t *testing.T) {
	// Capability whose RealizedBy references a service not in the Services map
	m := entity.NewUNMModel("Test", "")
	cap, err := entity.NewCapability("cap-1", "UnboundCap", "")
	require.NoError(t, err)
	danglingID, err := valueobject.NewEntityID("ghost-service")
	require.NoError(t, err)
	cap.AddRealizedBy(entity.NewRelationship(danglingID, "", valueobject.RelationshipRole("")))
	require.NoError(t, m.AddCapability(cap))

	// ghost-service is NOT in m.Services — the inner `if found` branch is false
	svcs := m.GetServicesForCapability("UnboundCap")
	assert.Empty(t, svcs)
}

// ── GetDataAssetsForService ───────────────────────────────────────────────────

func TestGetDataAssetsForService(t *testing.T) {
	m := entity.NewUNMModel("Test", "")

	// DataAsset found via ProducedBy
	da1, err := entity.NewDataAsset("da-1", "CatalogDB", "database", "Core catalog store")
	require.NoError(t, err)
	da1.ProducedBy = "feed-svc"
	require.NoError(t, m.AddDataAsset(da1))

	// DataAsset found via UsedBy
	da2, err := entity.NewDataAsset("da-2", "CacheLayer", "cache", "Redis cache")
	require.NoError(t, err)
	da2.AddUsedBy("serving-svc", "read")
	require.NoError(t, m.AddDataAsset(da2))

	// DataAsset found via ConsumedBy
	da3, err := entity.NewDataAsset("da-3", "EventStream", "event-stream", "Kafka stream")
	require.NoError(t, err)
	da3.ConsumedBy = append(da3.ConsumedBy, "analytics-svc")
	require.NoError(t, m.AddDataAsset(da3))

	// ProducedBy match
	result := m.GetDataAssetsForService("feed-svc")
	assert.Len(t, result, 1)
	assert.Equal(t, "CatalogDB", result[0].Name)

	// UsedBy match
	result = m.GetDataAssetsForService("serving-svc")
	assert.Len(t, result, 1)
	assert.Equal(t, "CacheLayer", result[0].Name)

	// ConsumedBy match
	result = m.GetDataAssetsForService("analytics-svc")
	assert.Len(t, result, 1)
	assert.Equal(t, "EventStream", result[0].Name)

	// No match
	result = m.GetDataAssetsForService("unknown-svc")
	assert.Empty(t, result)
}

// ── GetExternalDepsForService ─────────────────────────────────────────────────

func TestGetExternalDepsForService(t *testing.T) {
	m := entity.NewUNMModel("Test", "")

	stripe, err := entity.NewExternalDependency("ext-1", "Stripe", "Payment processor")
	require.NoError(t, err)
	stripe.AddUsedBy("payment-svc", "Charges customers")
	require.NoError(t, m.AddExternalDependency(stripe))

	// Found via UsedBy
	result := m.GetExternalDepsForService("payment-svc")
	assert.Len(t, result, 1)
	assert.Equal(t, "Stripe", result[0].Name)

	// Not found
	result = m.GetExternalDepsForService("unknown-svc")
	assert.Empty(t, result)
}

// ── addCapabilityWithParent: duplicate child name error ───────────────────────

func TestAddCapability_DuplicateChildName_Error(t *testing.T) {
	m := entity.NewUNMModel("Test", "")

	// First add a capability named "Child" directly
	child1, err := entity.NewCapability("c-1", "Child", "")
	require.NoError(t, err)
	require.NoError(t, m.AddCapability(child1))

	// Now try to add a parent whose child has the same name → duplicate in recursive call
	parent, err := entity.NewCapability("p-1", "Parent", "")
	require.NoError(t, err)
	child2, err := entity.NewCapability("c-2", "Child", "") // same name as child1
	require.NoError(t, err)
	parent.AddChild(child2)

	err = m.AddCapability(parent)
	assert.Error(t, err, "expected error for duplicate child capability name")
}
