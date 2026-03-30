package dsl

import (
	"strings"
	"testing"
)

func TestTransform_SystemName(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test System", Description: "A system"},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.System.Name != "Test System" {
		t.Errorf("expected system name %q, got %q", "Test System", model.System.Name)
	}
	if model.System.Description != "A system" {
		t.Errorf("expected description %q, got %q", "A system", model.System.Description)
	}
}

func TestTransform_MissingSystem(t *testing.T) {
	f := &File{}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for missing system")
	}
}

func TestTransform_Actor(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Actors: []*ActorNode{
			{Name: "Merchant", Description: "A merchant"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actor, ok := model.Actors["Merchant"]
	if !ok {
		t.Fatal("expected actor 'Merchant' in model")
	}
	if actor.Name != "Merchant" {
		t.Errorf("expected name %q, got %q", "Merchant", actor.Name)
	}
	if actor.Description != "A merchant" {
		t.Errorf("expected description %q, got %q", "A merchant", actor.Description)
	}
}

func TestTransform_Need(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Needs: []*NeedNode{
			{
				Name:        "Accept Payments",
				Actors:      []string{"Merchant"},
				Description: "Payment acceptance",
				SupportedBy: []RelationshipNode{
					{Target: "Payment Processing", Description: "Core capability"},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	need, ok := model.Needs["Accept Payments"]
	if !ok {
		t.Fatal("expected need 'Accept Payments' in model")
	}
	if need.ActorNames[0] != "Merchant" {
		t.Errorf("expected actor %q, got %q", "Merchant", need.ActorNames[0])
	}
	if len(need.SupportedBy) != 1 {
		t.Fatalf("expected 1 supportedBy, got %d", len(need.SupportedBy))
	}
	if need.SupportedBy[0].TargetID.String() != "Payment Processing" {
		t.Errorf("expected supportedBy target %q, got %q", "Payment Processing", need.SupportedBy[0].TargetID.String())
	}
	if need.SupportedBy[0].Description != "Core capability" {
		t.Errorf("expected description %q, got %q", "Core capability", need.SupportedBy[0].Description)
	}
}

func TestTransform_CapabilityHierarchy(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{
				Name:       "Payment Processing",
				Visibility: "user-facing",
				Children: []*CapabilityNode{
					{
						Name: "Payment Capture",
						RealizedBy: []RelationshipNode{
							{Target: "capture-service"},
						},
					},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both parent and child should be in flat map
	parent, ok := model.Capabilities["Payment Processing"]
	if !ok {
		t.Fatal("expected 'Payment Processing' in capabilities")
	}
	if parent.Visibility != "user-facing" {
		t.Errorf("expected visibility %q, got %q", "user-facing", parent.Visibility)
	}
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}

	child, ok := model.Capabilities["Payment Capture"]
	if !ok {
		t.Fatal("expected 'Payment Capture' in capabilities")
	}
	if child.Name != "Payment Capture" {
		t.Errorf("expected child name %q, got %q", "Payment Capture", child.Name)
	}

	// Verify parent tracking
	parentName, hasParent := model.CapabilityParents["Payment Capture"]
	if !hasParent {
		t.Fatal("expected Payment Capture to have a parent")
	}
	if parentName != "Payment Processing" {
		t.Errorf("expected parent %q, got %q", "Payment Processing", parentName)
	}
}

func TestTransform_Service(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Services: []*ServiceNode{
			{
				Name:        "payment-service",
				Description: "Core payment service",
				OwnedBy:     "payments-team",
				DependsOn: []RelationshipNode{
					{Target: "fraud-service", Description: "For validation"},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc, ok := model.Services["payment-service"]
	if !ok {
		t.Fatal("expected service 'payment-service'")
	}
	if svc.OwnerTeamName != "payments-team" {
		t.Errorf("expected owner %q, got %q", "payments-team", svc.OwnerTeamName)
	}
	if len(svc.DependsOn) != 1 {
		t.Fatalf("expected 1 dependsOn, got %d", len(svc.DependsOn))
	}
	if svc.DependsOn[0].TargetID.String() != "fraud-service" {
		t.Errorf("expected dependsOn target %q, got %q", "fraud-service", svc.DependsOn[0].TargetID.String())
	}
}

func TestTransform_Team(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Teams: []*TeamNode{
			{
				Name:        "payments-team",
				Type:        "stream-aligned",
				Description: "Payments",
				Owns:        []string{"payment-service"},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	team, ok := model.Teams["payments-team"]
	if !ok {
		t.Fatal("expected team 'payments-team'")
	}
	if team.TeamType.String() != "stream-aligned" {
		t.Errorf("expected type %q, got %q", "stream-aligned", team.TeamType.String())
	}
	if len(team.Owns) != 1 {
		t.Fatalf("expected 1 owns, got %d", len(team.Owns))
	}
	if team.Owns[0].TargetID.String() != "payment-service" {
		t.Errorf("expected owns %q, got %q", "payment-service", team.Owns[0].TargetID.String())
	}
}

func TestTransform_Platform(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Platforms: []*PlatformNode{
			{
				Name:        "Payments Platform",
				Description: "All payment teams",
				Teams:       []string{"payments-team", "fraud-team"},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pl, ok := model.Platforms["Payments Platform"]
	if !ok {
		t.Fatal("expected platform 'Payments Platform'")
	}
	if len(pl.TeamNames) != 2 {
		t.Fatalf("expected 2 team names, got %d", len(pl.TeamNames))
	}
}

func TestTransform_Interaction(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Interactions: []*InteractionNode{
			{
				From:        "payments-team",
				To:          "fraud-team",
				Mode:        "x-as-a-service",
				Description: "Fraud screening API",
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(model.Interactions))
	}
	inter := model.Interactions[0]
	if inter.FromTeamName != "payments-team" {
		t.Errorf("expected from %q, got %q", "payments-team", inter.FromTeamName)
	}
	if inter.ToTeamName != "fraud-team" {
		t.Errorf("expected to %q, got %q", "fraud-team", inter.ToTeamName)
	}
	if inter.Mode.String() != "x-as-a-service" {
		t.Errorf("expected mode %q, got %q", "x-as-a-service", inter.Mode.String())
	}
}

func TestTransform_DataAsset(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		DataAssets: []*DataAssetNode{
			{
				Name:        "payments-db",
				Type:        "database",
				Description: "Payment records",
				UsedBy:      []string{"payment-service"},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	da, ok := model.DataAssets["payments-db"]
	if !ok {
		t.Fatal("expected data asset 'payments-db'")
	}
	if da.Type != "database" {
		t.Errorf("expected type %q, got %q", "database", da.Type)
	}
	if len(da.UsedBy) != 1 {
		t.Fatalf("expected 1 usedBy, got %d", len(da.UsedBy))
	}
	if da.UsedBy[0] != "payment-service" {
		t.Errorf("expected usedBy[0] %q, got %q", "payment-service", da.UsedBy[0])
	}
}

func TestTransform_ExternalDependency(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		ExternalDependencies: []*ExternalDependencyNode{
			{
				Name:        "stripe",
				Description: "Payment gateway",
				UsedBy:      []ExternalDepUsageNode{{Target: "gateway-service"}},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ext, ok := model.ExternalDependencies["stripe"]
	if !ok {
		t.Fatal("expected external dependency 'stripe'")
	}
	if len(ext.UsedBy) != 1 {
		t.Fatalf("expected 1 usedBy, got %d", len(ext.UsedBy))
	}
	if ext.UsedBy[0].ServiceName != "gateway-service" {
		t.Errorf("expected usedBy service %q, got %q", "gateway-service", ext.UsedBy[0].ServiceName)
	}
}

func TestTransform_Signal(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Signals: []*SignalNode{
			{
				Name:        "Fragmentation risk",
				Category:    "fragmentation",
				Severity:    "high",
				Description: "Too many owners",
				OnEntity:    "Payment Processing",
				Affected:    []string{"Entity A", "Entity B"},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(model.Signals))
	}
	sig := model.Signals[0]
	if sig.Category != "fragmentation" {
		t.Errorf("expected category %q, got %q", "fragmentation", sig.Category)
	}
	if sig.Severity.String() != "high" {
		t.Errorf("expected severity %q, got %q", "high", sig.Severity.String())
	}
	if sig.OnEntityName != "Payment Processing" {
		t.Errorf("expected onEntity %q, got %q", "Payment Processing", sig.OnEntityName)
	}
	if len(sig.AffectedEntities) != 2 {
		t.Fatalf("expected 2 affected, got %d", len(sig.AffectedEntities))
	}
}

func TestTransform_InvalidTeamType(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Teams: []*TeamNode{
			{Name: "bad-team", Type: "not-a-valid-type"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for invalid team type")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("expected error to mention 'type', got: %v", err)
	}
}

func TestTransform_InvalidVisibility(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Bad Cap", Visibility: "not-valid"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for invalid visibility")
	}
}

func TestTransform_InvalidInteractionMode(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Interactions: []*InteractionNode{
			{From: "a", To: "b", Mode: "invalid-mode"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for invalid interaction mode")
	}
}

func TestTransform_FreeFormDataAssetType(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		DataAssets: []*DataAssetNode{
			{Name: "custom-asset", Type: "custom-type"},
		},
	}
	_, err := Transform(f)
	if err != nil {
		t.Fatalf("expected no error for free-form data asset type, got: %v", err)
	}
}

// P0: Need outcome flows through transformer
func TestTransform_NeedOutcome(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Needs: []*NeedNode{
			{Name: "Fast checkout", Actors: []string{"Shopper"}, Description: "desc", Outcome: "checkout in 3 clicks"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	need, ok := model.Needs["Fast checkout"]
	if !ok {
		t.Fatal("expected need 'Fast checkout'")
	}
	if need.Outcome != "checkout in 3 clicks" {
		t.Errorf("expected outcome %q, got %q", "checkout in 3 clicks", need.Outcome)
	}
}

// P0: Need outcome falls back to description when not set
func TestTransform_NeedOutcomeFallback(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Needs: []*NeedNode{
			{Name: "Fast checkout", Actors: []string{"Shopper"}, Description: "fallback desc"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	need, ok := model.Needs["Fast checkout"]
	if !ok {
		t.Fatal("expected need 'Fast checkout'")
	}
	if need.Outcome != "fallback desc" {
		t.Errorf("expected outcome fallback to description %q, got %q", "fallback desc", need.Outcome)
	}
}

// P0: Team size and SizeExplicit flow through transformer
func TestTransform_TeamSize(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Teams: []*TeamNode{
			{Name: "Platform", Type: "platform", Size: 6},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	team, ok := model.Teams["Platform"]
	if !ok {
		t.Fatal("expected team 'Platform'")
	}
	if team.Size != 6 {
		t.Errorf("expected size 6, got %d", team.Size)
	}
	if !team.SizeExplicit {
		t.Error("expected SizeExplicit = true")
	}
}

// P0: Interaction via flows through transformer
func TestTransform_InteractionVia(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Interactions: []*InteractionNode{
			{From: "service-a", To: "service-b", Via: "api-gateway", Mode: "x-as-a-service"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(model.Interactions))
	}
	if model.Interactions[0].Via != "api-gateway" {
		t.Errorf("expected via %q, got %q", "api-gateway", model.Interactions[0].Via)
	}
}

func TestTransform_InvalidSignalSeverity(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Signals: []*SignalNode{
			{Name: "s", Category: "bottleneck", Severity: "invalid", OnEntity: "X"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for invalid signal severity")
	}
}

func TestTransform_FullRoundTrip(t *testing.T) {
	src := `
system "INCA" {
  description "INCA Platform"
}
actor "Merchant" {
  description "Sells items"
}
need "Accept Payments" {
  actor "Merchant"
  supportedBy "Payment Processing"
}
capability "Payment Processing" {
  visibility user-facing
  realizedBy "payment-service" { description "Core handler" role primary }
  capability "Payment Capture" {
    realizedBy "capture-service"
  }
}
service "payment-service" {
  ownedBy "payments-team"
  dependsOn "fraud-service"
}
team "payments-team" {
  type stream-aligned
  owns "payment-service"
}
platform "Payments Platform" {
  teams ["payments-team"]
}
interaction "payments-team" -> "fraud-team" {
  mode x-as-a-service
}
data_asset "payments-db" {
  type database
  usedBy "payment-service"
}
external_dependency "stripe" {
  description "Gateway"
  usedBy "payment-service"
}
signal "Bottleneck" {
  category bottleneck
  severity medium
  onEntity "Payment Processing"
}
`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	model, err := Transform(ast)
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}

	if model.System.Name != "INCA" {
		t.Errorf("expected system name %q, got %q", "INCA", model.System.Name)
	}
	if _, ok := model.Actors["Merchant"]; !ok {
		t.Error("expected actor 'Merchant'")
	}
	if _, ok := model.Needs["Accept Payments"]; !ok {
		t.Error("expected need 'Accept Payments'")
	}
	if _, ok := model.Capabilities["Payment Processing"]; !ok {
		t.Error("expected capability 'Payment Processing'")
	}
	if _, ok := model.Capabilities["Payment Capture"]; !ok {
		t.Error("expected capability 'Payment Capture'")
	}
	if _, ok := model.Services["payment-service"]; !ok {
		t.Error("expected service 'payment-service'")
	}
	if _, ok := model.Teams["payments-team"]; !ok {
		t.Error("expected team 'payments-team'")
	}
	if _, ok := model.Platforms["Payments Platform"]; !ok {
		t.Error("expected platform 'Payments Platform'")
	}
	if len(model.Interactions) != 1 {
		t.Errorf("expected 1 interaction, got %d", len(model.Interactions))
	}
	if _, ok := model.DataAssets["payments-db"]; !ok {
		t.Error("expected data asset 'payments-db'")
	}
	if _, ok := model.ExternalDependencies["stripe"]; !ok {
		t.Error("expected external dependency 'stripe'")
	}
	if len(model.Signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(model.Signals))
	}

	// Verify relationship role
	cap := model.Capabilities["Payment Processing"]
	if cap.RealizedBy[0].Role.String() != "primary" {
		t.Errorf("expected role %q, got %q", "primary", cap.RealizedBy[0].Role.String())
	}
}

// ---------------------------------------------------------------------------
// 5.6 — Inferred Mapping Transformer
// ---------------------------------------------------------------------------

func TestTransform_InferredMapping(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		InferredMappings: []*InferredMappingNode{
			{
				From:       "my-service",
				To:         "my-capability",
				Confidence: 0.85,
				Evidence:   "Detected via scan",
				Status:     "inferred",
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.InferredMappings) != 1 {
		t.Fatalf("expected 1 inferred mapping, got %d", len(model.InferredMappings))
	}
	im := model.InferredMappings[0]
	if im.ServiceName != "my-service" {
		t.Errorf("expected ServiceName %q, got %q", "my-service", im.ServiceName)
	}
	if im.CapabilityName != "my-capability" {
		t.Errorf("expected CapabilityName %q, got %q", "my-capability", im.CapabilityName)
	}
	if im.Confidence.Score != 0.85 {
		t.Errorf("expected confidence score 0.85, got %f", im.Confidence.Score)
	}
	if im.Confidence.Evidence != "Detected via scan" {
		t.Errorf("expected evidence %q, got %q", "Detected via scan", im.Confidence.Evidence)
	}
	if im.Status.String() != "inferred" {
		t.Errorf("expected status %q, got %q", "inferred", im.Status.String())
	}
}

func TestTransform_InferredMapping_InvalidStatus(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		InferredMappings: []*InferredMappingNode{
			{From: "s", To: "c", Confidence: 0.5, Status: "not-valid"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for invalid mapping status")
	}
}

func TestTransform_InferredMapping_InvalidConfidence(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		InferredMappings: []*InferredMappingNode{
			{From: "s", To: "c", Confidence: 1.5, Status: "inferred"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for confidence out of range")
	}
}

func TestTransform_InferredMapping_FullRoundTrip(t *testing.T) {
	src := `
system "Test" {}
inferred {
    from "feed-service"
    to "Feed Ingestion"
    confidence 0.85
    evidence "Detected via API scan of feed-api service"
    status inferred
}
`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	model, err := Transform(ast)
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}
	if len(model.InferredMappings) != 1 {
		t.Fatalf("expected 1 inferred mapping, got %d", len(model.InferredMappings))
	}
}

// ---------------------------------------------------------------------------
// 5.7 — Transition Transformer
// ---------------------------------------------------------------------------

func TestTransform_Transitions(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Transitions: []*TransitionNode{
			{
				Name:        "Consolidate catalog",
				Description: "Move to stream-aligned",
				Current: []TransitionBindingNode{
					{CapabilityName: "Catalog publication", TeamName: "Team A"},
					{CapabilityName: "Catalog publication", TeamName: "Team B"},
				},
				Target: []TransitionBindingNode{
					{CapabilityName: "Catalog publication", TeamName: "Catalog Stream"},
				},
				Steps: []TransitionStepNode{
					{Number: 1, Label: "Merge teams", ActionText: "merge team A team B", ExpectedOutcome: "Single team"},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(model.Transitions))
	}
	tr := model.Transitions[0]
	if tr.Name != "Consolidate catalog" {
		t.Errorf("expected name %q, got %q", "Consolidate catalog", tr.Name)
	}
	if tr.Description != "Move to stream-aligned" {
		t.Errorf("unexpected description: %q", tr.Description)
	}
	if len(tr.Current) != 2 {
		t.Fatalf("expected 2 current bindings, got %d", len(tr.Current))
	}
	if tr.Current[0].CapabilityName != "Catalog publication" || tr.Current[0].TeamName != "Team A" {
		t.Errorf("unexpected current[0]: %+v", tr.Current[0])
	}
	if len(tr.Target) != 1 {
		t.Fatalf("expected 1 target binding, got %d", len(tr.Target))
	}
	if tr.Target[0].TeamName != "Catalog Stream" {
		t.Errorf("unexpected target[0]: %+v", tr.Target[0])
	}
	if len(tr.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(tr.Steps))
	}
	if tr.Steps[0].Number != 1 || tr.Steps[0].Label != "Merge teams" {
		t.Errorf("unexpected step: %+v", tr.Steps[0])
	}
	if tr.Steps[0].ActionText != "merge team A team B" {
		t.Errorf("unexpected action text: %q", tr.Steps[0].ActionText)
	}
	if tr.Steps[0].ExpectedOutcome != "Single team" {
		t.Errorf("unexpected expected outcome: %q", tr.Steps[0].ExpectedOutcome)
	}
}

// ---------------------------------------------------------------------------
// 9.1.2 — Flat capabilities with parent + two-pass resolution
// ---------------------------------------------------------------------------

func TestTransform_FlatCapabilityParent(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Child Cap", Parent: "Parent Cap", Visibility: "domain"},
			{Name: "Parent Cap", Visibility: "domain"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parent, ok := model.Capabilities["Parent Cap"]
	if !ok {
		t.Fatal("expected 'Parent Cap' in capabilities")
	}
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child on Parent Cap, got %d", len(parent.Children))
	}
	if parent.Children[0].Name != "Child Cap" {
		t.Errorf("expected child name %q, got %q", "Child Cap", parent.Children[0].Name)
	}
	_, ok = model.Capabilities["Child Cap"]
	if !ok {
		t.Fatal("expected 'Child Cap' in capabilities")
	}
	parentName, hasParent := model.CapabilityParents["Child Cap"]
	if !hasParent {
		t.Fatal("expected Child Cap to have a parent")
	}
	if parentName != "Parent Cap" {
		t.Errorf("expected parent %q, got %q", "Parent Cap", parentName)
	}
}

func TestTransform_FlatCapabilityParent_MultiLevel(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Grandchild", Parent: "Child", Visibility: "domain"},
			{Name: "Child", Parent: "Root", Visibility: "domain"},
			{Name: "Root", Visibility: "domain"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Capabilities) != 3 {
		t.Fatalf("expected 3 capabilities, got %d", len(model.Capabilities))
	}
	root := model.Capabilities["Root"]
	if len(root.Children) != 1 {
		t.Errorf("expected 1 child on Root, got %d", len(root.Children))
	}
	child := model.Capabilities["Child"]
	if len(child.Children) != 1 {
		t.Errorf("expected 1 child on Child, got %d", len(child.Children))
	}
}

func TestTransform_FlatCapabilityParent_MissingParent_Error(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Child Cap", Parent: "Nonexistent Parent", Visibility: "domain"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for missing parent capability")
	}
	if !strings.Contains(err.Error(), "parent") {
		t.Errorf("expected error to mention 'parent', got: %v", err)
	}
}

func TestTransform_FlatCapabilityParent_Circular_Error(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Cap A", Parent: "Cap B", Visibility: "domain"},
			{Name: "Cap B", Parent: "Cap A", Visibility: "domain"},
		},
	}
	_, err := Transform(f)
	if err == nil {
		t.Fatal("expected error for circular parent reference")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected error to mention 'circular', got: %v", err)
	}
}

func TestTransform_MixedFlatAndNested(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{
				Name:       "Root",
				Visibility: "domain",
				Children: []*CapabilityNode{
					{Name: "Nested Child", Visibility: "foundational"},
				},
			},
			{Name: "Flat Child", Parent: "Root", Visibility: "user-facing"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Capabilities) != 3 {
		t.Fatalf("expected 3 capabilities, got %d", len(model.Capabilities))
	}
	root := model.Capabilities["Root"]
	// Root should have both nested child and flat child
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children on Root (nested + flat), got %d", len(root.Children))
	}
}

// ---------------------------------------------------------------------------
// 9.2.2 — Visibility inheritance in DSL transformer
// ---------------------------------------------------------------------------

func TestTransform_VisibilityInheritance_ChildInheritsParent(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Child Cap", Parent: "Parent Cap"}, // no explicit visibility
			{Name: "Parent Cap", Visibility: "domain"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	child := model.Capabilities["Child Cap"]
	if child.Visibility != "domain" {
		t.Errorf("expected child to inherit parent visibility 'domain', got %q", child.Visibility)
	}
}

func TestTransform_VisibilityInheritance_ChildOverrides(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Child Cap", Parent: "Parent Cap", Visibility: "foundational"},
			{Name: "Parent Cap", Visibility: "domain"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	child := model.Capabilities["Child Cap"]
	if child.Visibility != "foundational" {
		t.Errorf("expected child to keep explicit visibility 'foundational', got %q", child.Visibility)
	}
}

func TestTransform_VisibilityInheritance_MultiLevel(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Grandchild", Parent: "Child"}, // no visibility
			{Name: "Child", Parent: "Root"},        // no visibility
			{Name: "Root", Visibility: "user-facing"},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grandchild := model.Capabilities["Grandchild"]
	if grandchild.Visibility != "user-facing" {
		t.Errorf("expected grandchild to inherit 'user-facing', got %q", grandchild.Visibility)
	}
	child := model.Capabilities["Child"]
	if child.Visibility != "user-facing" {
		t.Errorf("expected child to inherit 'user-facing', got %q", child.Visibility)
	}
}

func TestTransform_VisibilityInheritance_NestedChildInherits(t *testing.T) {
	// Nested (not flat) children should also inherit
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{
				Name:       "Parent",
				Visibility: "domain",
				Children: []*CapabilityNode{
					{Name: "Nested Child"}, // no visibility
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	child := model.Capabilities["Nested Child"]
	if child.Visibility != "domain" {
		t.Errorf("expected nested child to inherit 'domain', got %q", child.Visibility)
	}
}

// ---------------------------------------------------------------------------
// 9.3.5 — realizes on service blocks
// ---------------------------------------------------------------------------

func TestTransform_ServiceRealizes_WiresCapability(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Cap A", Visibility: "domain"},
		},
		Services: []*ServiceNode{
			{
				Name:    "my-service",
				OwnedBy: "team-a",
				Realizes: []ServiceRealizesNode{
					{Target: "Cap A"},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cap, ok := model.Capabilities["Cap A"]
	if !ok {
		t.Fatal("expected 'Cap A' in capabilities")
	}
	if len(cap.RealizedBy) != 1 {
		t.Fatalf("expected 1 realizedBy on Cap A, got %d", len(cap.RealizedBy))
	}
	if cap.RealizedBy[0].TargetID.String() != "my-service" {
		t.Errorf("expected realizedBy target %q, got %q", "my-service", cap.RealizedBy[0].TargetID.String())
	}
}

func TestTransform_ServiceRealizes_WithRole(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Capabilities: []*CapabilityNode{
			{Name: "Cap B", Visibility: "domain"},
		},
		Services: []*ServiceNode{
			{
				Name:    "svc-b",
				OwnedBy: "team-a",
				Realizes: []ServiceRealizesNode{
					{Target: "Cap B", Role: "supporting"},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cap := model.Capabilities["Cap B"]
	if cap.RealizedBy[0].Role.String() != "supporting" {
		t.Errorf("expected role %q, got %q", "supporting", cap.RealizedBy[0].Role.String())
	}
}

// ---------------------------------------------------------------------------
// 9.4.3 — externalDeps on service blocks
// ---------------------------------------------------------------------------

func TestTransform_ServiceExternalDeps_PopulatesUsedBy(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		ExternalDependencies: []*ExternalDependencyNode{
			{Name: "temporal", Description: "Workflow engine"},
			{Name: "postgres", Description: "Database"},
		},
		Services: []*ServiceNode{
			{
				Name:         "my-service",
				OwnedBy:      "team-a",
				ExternalDeps: []string{"temporal", "postgres"},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	temporal, ok := model.ExternalDependencies["temporal"]
	if !ok {
		t.Fatal("expected 'temporal' in external dependencies")
	}
	if len(temporal.UsedBy) != 1 {
		t.Fatalf("expected 1 usedBy on temporal, got %d", len(temporal.UsedBy))
	}
	if temporal.UsedBy[0].ServiceName != "my-service" {
		t.Errorf("expected usedBy service %q, got %q", "my-service", temporal.UsedBy[0].ServiceName)
	}
	postgres := model.ExternalDependencies["postgres"]
	if len(postgres.UsedBy) != 1 || postgres.UsedBy[0].ServiceName != "my-service" {
		t.Errorf("expected postgres.UsedBy[my-service], got %v", postgres.UsedBy)
	}
}

// ---------------------------------------------------------------------------
// 9.5.3 — interacts on team blocks
// ---------------------------------------------------------------------------

func TestTransform_TeamInteracts_CreatesInteraction(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Teams: []*TeamNode{
			{
				Name: "team-a",
				Type: "stream-aligned",
				Interacts: []TeamInteractionNode{
					{With: "team-b", Mode: "x-as-a-service", Via: "API", Description: "Fraud API"},
				},
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(model.Interactions))
	}
	inter := model.Interactions[0]
	if inter.FromTeamName != "team-a" {
		t.Errorf("expected from %q, got %q", "team-a", inter.FromTeamName)
	}
	if inter.ToTeamName != "team-b" {
		t.Errorf("expected to %q, got %q", "team-b", inter.ToTeamName)
	}
	if inter.Mode.String() != "x-as-a-service" {
		t.Errorf("expected mode %q, got %q", "x-as-a-service", inter.Mode.String())
	}
	if inter.Via != "API" {
		t.Errorf("expected via %q, got %q", "API", inter.Via)
	}
	if inter.Description != "Fraud API" {
		t.Errorf("expected description %q, got %q", "Fraud API", inter.Description)
	}
}

// ---------------------------------------------------------------------------
// Full round-trip: 9.x DSL features parsed and transformed
// ---------------------------------------------------------------------------

func TestTransform_Phase9_FullRoundTrip(t *testing.T) {
	src := `
system "Test" {}

capability "Root Cap" {
  visibility domain
}
capability "Child Cap" {
  parent "Root Cap"
}

external_dependency "temporal" {}

service "my-service" {
  ownedBy "team-a"
  realizes "Root Cap" role supporting
  externalDeps "temporal"
}

team "team-a" {
  type stream-aligned
  interacts "team-b" mode x-as-a-service via "API"
}
`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	model, err := Transform(ast)
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}

	// Check capabilities hierarchy
	root, ok := model.Capabilities["Root Cap"]
	if !ok {
		t.Fatal("expected 'Root Cap' in capabilities")
	}
	if len(root.Children) != 1 || root.Children[0].Name != "Child Cap" {
		t.Errorf("expected Root Cap to have Child Cap as child, got %v", root.Children)
	}

	// Check visibility inheritance
	child := model.Capabilities["Child Cap"]
	if child.Visibility != "domain" {
		t.Errorf("expected Child Cap to inherit 'domain' visibility, got %q", child.Visibility)
	}

	// Check realizes wiring
	if len(root.RealizedBy) != 1 {
		t.Fatalf("expected 1 realizedBy on Root Cap, got %d", len(root.RealizedBy))
	}
	if root.RealizedBy[0].TargetID.String() != "my-service" {
		t.Errorf("expected realizedBy target %q, got %q", "my-service", root.RealizedBy[0].TargetID.String())
	}

	// Check externalDeps wiring
	temporal := model.ExternalDependencies["temporal"]
	if len(temporal.UsedBy) != 1 || temporal.UsedBy[0].ServiceName != "my-service" {
		t.Errorf("expected temporal.UsedBy[my-service], got %v", temporal.UsedBy)
	}

	// Check team interaction
	if len(model.Interactions) != 1 {
		t.Fatalf("expected 1 interaction from team interacts, got %d", len(model.Interactions))
	}
}

func TestTransform_Transitions_FullRoundTrip(t *testing.T) {
	src := `
system "Test" {}
transition "Consolidate" {
  description "Move to stream-aligned"
  current {
    capability "Cap A" ownedBy team "Team A"
  }
  target {
    capability "Cap A" ownedBy team "Merged Team"
  }
  step 1 "Merge" {
    action merge team "Team A" into team "Merged Team"
    expected_outcome "Done"
  }
}
`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	model, err := Transform(ast)
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}
	if len(model.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(model.Transitions))
	}
	tr := model.Transitions[0]
	if tr.Name != "Consolidate" {
		t.Errorf("expected name %q, got %q", "Consolidate", tr.Name)
	}
	if len(tr.Current) != 1 || len(tr.Target) != 1 || len(tr.Steps) != 1 {
		t.Errorf("unexpected counts: current=%d target=%d steps=%d", len(tr.Current), len(tr.Target), len(tr.Steps))
	}
}


func TestTransform_NeedMultiActor(t *testing.T) {
	f := &File{
		System: &SystemNode{Name: "Test"},
		Needs: []*NeedNode{
			{
				Name:   "Shared Need",
				Actors: []string{"Actor A", "Actor B"},
				Outcome: "some shared outcome",
			},
		},
	}
	model, err := Transform(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	need, ok := model.Needs["Shared Need"]
	if !ok {
		t.Fatal("expected need 'Shared Need' in model")
	}
	if len(need.ActorNames) != 2 {
		t.Fatalf("expected 2 ActorNames, got %d", len(need.ActorNames))
	}
	if need.ActorNames[0] != "Actor A" {
		t.Errorf("expected ActorNames[0] %q, got %q", "Actor A", need.ActorNames[0])
	}
	if need.ActorNames[1] != "Actor B" {
		t.Errorf("expected ActorNames[1] %q, got %q", "Actor B", need.ActorNames[1])
	}
}
