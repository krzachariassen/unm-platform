package dsl

import (
	"strings"
	"testing"
)

func TestParse_EmptyInput(t *testing.T) {
	f, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.System != nil {
		t.Error("expected nil system for empty input")
	}
	if len(f.Actors) != 0 {
		t.Errorf("expected no actors, got %d", len(f.Actors))
	}
}

func TestParse_SystemBlock(t *testing.T) {
	src := `system "My System" { description "A test system" }`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.System == nil {
		t.Fatal("expected system node")
	}
	if f.System.Name != "My System" {
		t.Errorf("expected name %q, got %q", "My System", f.System.Name)
	}
	if f.System.Description != "A test system" {
		t.Errorf("expected description %q, got %q", "A test system", f.System.Description)
	}
}

func TestParse_SystemBlock_Empty(t *testing.T) {
	src := `system "Empty" {}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.System == nil {
		t.Fatal("expected system node")
	}
	if f.System.Name != "Empty" {
		t.Errorf("expected name %q, got %q", "Empty", f.System.Name)
	}
	if f.System.Description != "" {
		t.Errorf("expected empty description, got %q", f.System.Description)
	}
}

func TestParse_ActorBlock(t *testing.T) {
	src := `actor "Merchant" { description "A merchant actor" }`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Actors) != 1 {
		t.Fatalf("expected 1 actor, got %d", len(f.Actors))
	}
	a := f.Actors[0]
	if a.Name != "Merchant" {
		t.Errorf("expected name %q, got %q", "Merchant", a.Name)
	}
	if a.Description != "A merchant actor" {
		t.Errorf("expected description %q, got %q", "A merchant actor", a.Description)
	}
}

func TestParse_MultipleActors(t *testing.T) {
	src := `
actor "Merchant" {}
actor "Eater" {}
actor "Operator" {}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Actors) != 3 {
		t.Fatalf("expected 3 actors, got %d", len(f.Actors))
	}
}

func TestParse_NeedBlock(t *testing.T) {
	src := `
need "Accept Payments" {
  actor "Merchant"
  description "Merchants need to accept card payments"
  supportedBy "Payment Processing"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Needs) != 1 {
		t.Fatalf("expected 1 need, got %d", len(f.Needs))
	}
	n := f.Needs[0]
	if n.Name != "Accept Payments" {
		t.Errorf("expected name %q, got %q", "Accept Payments", n.Name)
	}
	if n.Actors[0] != "Merchant" {
		t.Errorf("expected actor %q, got %q", "Merchant", n.Actors[0])
	}
	if n.Description != "Merchants need to accept card payments" {
		t.Errorf("unexpected description %q", n.Description)
	}
	if len(n.SupportedBy) != 1 {
		t.Fatalf("expected 1 supportedBy, got %d", len(n.SupportedBy))
	}
	if n.SupportedBy[0].Target != "Payment Processing" {
		t.Errorf("expected supportedBy target %q, got %q", "Payment Processing", n.SupportedBy[0].Target)
	}
}

func TestParse_NeedBlock_MultipleSupportedBy(t *testing.T) {
	src := `
need "Buy Items" {
  actor "Eater"
  supportedBy "Catalog"
  supportedBy "Checkout"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := f.Needs[0]
	if len(n.SupportedBy) != 2 {
		t.Fatalf("expected 2 supportedBy, got %d", len(n.SupportedBy))
	}
}

func TestParse_CapabilityBlock(t *testing.T) {
	src := `
capability "Payment Processing" {
  visibility user-facing
  description "Handles payments"
  realizedBy "payment-service"
  dependsOn "Fraud Detection"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Capabilities) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(f.Capabilities))
	}
	c := f.Capabilities[0]
	if c.Name != "Payment Processing" {
		t.Errorf("expected name %q, got %q", "Payment Processing", c.Name)
	}
	if c.Visibility != "user-facing" {
		t.Errorf("expected visibility %q, got %q", "user-facing", c.Visibility)
	}
	if len(c.RealizedBy) != 1 {
		t.Fatalf("expected 1 realizedBy, got %d", len(c.RealizedBy))
	}
	if c.RealizedBy[0].Target != "payment-service" {
		t.Errorf("expected realizedBy target %q, got %q", "payment-service", c.RealizedBy[0].Target)
	}
	if len(c.DependsOn) != 1 {
		t.Fatalf("expected 1 dependsOn, got %d", len(c.DependsOn))
	}
}

func TestParse_CapabilityBlock_WithChildren(t *testing.T) {
	src := `
capability "Payment Processing" {
  visibility domain
  capability "Payment Capture" {
    realizedBy "capture-service"
  }
  capability "Payment Auth" {
    realizedBy "auth-service"
  }
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Capabilities) != 1 {
		t.Fatalf("expected 1 top-level capability, got %d", len(f.Capabilities))
	}
	c := f.Capabilities[0]
	if len(c.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(c.Children))
	}
	if c.Children[0].Name != "Payment Capture" {
		t.Errorf("expected first child %q, got %q", "Payment Capture", c.Children[0].Name)
	}
	if c.Children[1].Name != "Payment Auth" {
		t.Errorf("expected second child %q, got %q", "Payment Auth", c.Children[1].Name)
	}
}

func TestParse_ServiceBlock(t *testing.T) {
	src := `
service "payment-service" {
  description "Core payment processing"
  ownedBy "payments-team"
  dependsOn "fraud-service"
  dependsOn "capture-service" { description "For settlement" }
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(f.Services))
	}
	s := f.Services[0]
	if s.Name != "payment-service" {
		t.Errorf("expected name %q, got %q", "payment-service", s.Name)
	}
	if s.OwnedBy != "payments-team" {
		t.Errorf("expected ownedBy %q, got %q", "payments-team", s.OwnedBy)
	}
	if len(s.DependsOn) != 2 {
		t.Fatalf("expected 2 dependsOn, got %d", len(s.DependsOn))
	}
	if s.DependsOn[1].Description != "For settlement" {
		t.Errorf("expected dependency description %q, got %q", "For settlement", s.DependsOn[1].Description)
	}
}

func TestParse_TeamBlock(t *testing.T) {
	src := `
team "payments-team" {
  type stream-aligned
  description "Owns payment flows"
  owns "payment-service"
  owns "capture-service"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(f.Teams))
	}
	team := f.Teams[0]
	if team.Name != "payments-team" {
		t.Errorf("expected name %q, got %q", "payments-team", team.Name)
	}
	if team.Type != "stream-aligned" {
		t.Errorf("expected type %q, got %q", "stream-aligned", team.Type)
	}
	if len(team.Owns) != 2 {
		t.Fatalf("expected 2 owns, got %d", len(team.Owns))
	}
}

func TestParse_PlatformBlock(t *testing.T) {
	src := `
platform "Payments Platform" {
  description "All payment teams"
  teams ["payments-team", "fraud-team"]
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Platforms) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(f.Platforms))
	}
	pl := f.Platforms[0]
	if pl.Name != "Payments Platform" {
		t.Errorf("expected name %q, got %q", "Payments Platform", pl.Name)
	}
	if len(pl.Teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(pl.Teams))
	}
	if pl.Teams[0] != "payments-team" {
		t.Errorf("expected first team %q, got %q", "payments-team", pl.Teams[0])
	}
	if pl.Teams[1] != "fraud-team" {
		t.Errorf("expected second team %q, got %q", "fraud-team", pl.Teams[1])
	}
}

func TestParse_InteractionBlock(t *testing.T) {
	src := `
interaction "payments-team" -> "fraud-team" {
  mode x-as-a-service
  description "Fraud screening API"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(f.Interactions))
	}
	inter := f.Interactions[0]
	if inter.From != "payments-team" {
		t.Errorf("expected from %q, got %q", "payments-team", inter.From)
	}
	if inter.To != "fraud-team" {
		t.Errorf("expected to %q, got %q", "fraud-team", inter.To)
	}
	if inter.Mode != "x-as-a-service" {
		t.Errorf("expected mode %q, got %q", "x-as-a-service", inter.Mode)
	}
	if inter.Description != "Fraud screening API" {
		t.Errorf("expected description %q, got %q", "Fraud screening API", inter.Description)
	}
}

func TestParse_DataAssetBlock(t *testing.T) {
	src := `
data_asset "payments-db" {
  type database
  description "Payment records"
  usedBy "payment-service"
  usedBy "capture-service"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.DataAssets) != 1 {
		t.Fatalf("expected 1 data_asset, got %d", len(f.DataAssets))
	}
	da := f.DataAssets[0]
	if da.Name != "payments-db" {
		t.Errorf("expected name %q, got %q", "payments-db", da.Name)
	}
	if da.Type != "database" {
		t.Errorf("expected type %q, got %q", "database", da.Type)
	}
	if len(da.UsedBy) != 2 {
		t.Fatalf("expected 2 usedBy, got %d", len(da.UsedBy))
	}
}

func TestParse_ExternalDependencyBlock(t *testing.T) {
	src := `
external_dependency "stripe" {
  description "Payment gateway"
  usedBy "gateway-service"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.ExternalDependencies) != 1 {
		t.Fatalf("expected 1 external_dependency, got %d", len(f.ExternalDependencies))
	}
	ext := f.ExternalDependencies[0]
	if ext.Name != "stripe" {
		t.Errorf("expected name %q, got %q", "stripe", ext.Name)
	}
	if ext.Description != "Payment gateway" {
		t.Errorf("expected description %q, got %q", "Payment gateway", ext.Description)
	}
	if len(ext.UsedBy) != 1 || ext.UsedBy[0].Target != "gateway-service" {
		t.Errorf("expected usedBy [gateway-service], got %v", ext.UsedBy)
	}
}

func TestParse_SignalBlock(t *testing.T) {
	src := `
signal "Fragmentation risk" {
  category fragmentation
  severity high
  description "Too many owners"
  onEntity "Payment Processing"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(f.Signals))
	}
	s := f.Signals[0]
	if s.Name != "Fragmentation risk" {
		t.Errorf("expected name %q, got %q", "Fragmentation risk", s.Name)
	}
	if s.Category != "fragmentation" {
		t.Errorf("expected category %q, got %q", "fragmentation", s.Category)
	}
	if s.Severity != "high" {
		t.Errorf("expected severity %q, got %q", "high", s.Severity)
	}
	if s.OnEntity != "Payment Processing" {
		t.Errorf("expected onEntity %q, got %q", "Payment Processing", s.OnEntity)
	}
}

func TestParse_RelationshipWithModifiers(t *testing.T) {
	src := `
capability "Payment Processing" {
  realizedBy "gateway-service" { description "Handles routing" role primary }
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Capabilities) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(f.Capabilities))
	}
	c := f.Capabilities[0]
	if len(c.RealizedBy) != 1 {
		t.Fatalf("expected 1 realizedBy, got %d", len(c.RealizedBy))
	}
	rel := c.RealizedBy[0]
	if rel.Target != "gateway-service" {
		t.Errorf("expected target %q, got %q", "gateway-service", rel.Target)
	}
	if rel.Description != "Handles routing" {
		t.Errorf("expected description %q, got %q", "Handles routing", rel.Description)
	}
	if rel.Role != "primary" {
		t.Errorf("expected role %q, got %q", "primary", rel.Role)
	}
}

func TestParse_RelationshipWithDescriptionOnly(t *testing.T) {
	src := `
need "Accept Payments" {
  actor "Merchant"
  supportedBy "Fraud Detection" { description "Validates transaction" }
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := f.Needs[0]
	if n.SupportedBy[0].Description != "Validates transaction" {
		t.Errorf("expected description %q, got %q", "Validates transaction", n.SupportedBy[0].Description)
	}
	if n.SupportedBy[0].Role != "" {
		t.Errorf("expected empty role, got %q", n.SupportedBy[0].Role)
	}
}

func TestParse_Comments(t *testing.T) {
	src := `
// This is a comment
system "My System" {
  // another comment
  description "desc"
}
// end comment
actor "Merchant" {}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.System == nil {
		t.Fatal("expected system node")
	}
	if f.System.Name != "My System" {
		t.Errorf("expected name %q, got %q", "My System", f.System.Name)
	}
	if len(f.Actors) != 1 {
		t.Errorf("expected 1 actor, got %d", len(f.Actors))
	}
}

func TestParse_MissingOpenBrace(t *testing.T) {
	src := `system "My System" description "desc" }`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for missing open brace")
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected error to contain line number, got: %v", err)
	}
}

func TestParse_UnknownTopLevelBlock(t *testing.T) {
	src := `unknown "foo" {}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for unknown keyword")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected error to mention 'unknown', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 5.4 — Error Reporting Quality
// ---------------------------------------------------------------------------

func TestParse_ParseError_Type(t *testing.T) {
	// Parse() should return a *ParseError for parse failures
	src := `unknown_kw "foo" {}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line <= 0 {
		t.Errorf("expected positive line number, got %d", pe.Line)
	}
	if pe.Message == "" {
		t.Error("expected non-empty ParseError.Message")
	}
}

func TestParse_ParseError_ErrorString(t *testing.T) {
	// ParseError.Error() should format as "line N: message"
	pe := &ParseError{Line: 5, Message: "unexpected token"}
	got := pe.Error()
	if !strings.HasPrefix(got, "line 5:") {
		t.Errorf("expected error to start with 'line 5:', got %q", got)
	}
	if !strings.Contains(got, "unexpected token") {
		t.Errorf("expected error to contain 'unexpected token', got %q", got)
	}
}

func TestParse_ErrorContainsLine_MissingBrace(t *testing.T) {
	src := "system \"My System\" description \"desc\" }"
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for missing open brace")
	}
	if !strings.Contains(err.Error(), "line 1:") {
		t.Errorf("expected error to contain 'line 1:', got: %v", err)
	}
}

func TestParse_ErrorContainsLine_MultiLine(t *testing.T) {
	// Error on line 4
	src := "system \"Sys\" {\n  description \"desc\"\n}\nunknown_kw \"x\" {}"
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for unknown keyword on line 4")
	}
	if !strings.Contains(err.Error(), "line 4:") {
		t.Errorf("expected error to contain 'line 4:', got: %v", err)
	}
}

func TestParse_ErrorContainsLine_UnexpectedEndOfInput(t *testing.T) {
	// system keyword with no name (EOF immediately)
	src := `system`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for unexpected end of input")
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected error to contain 'line', got: %v", err)
	}
}

func TestParse_ErrorContainsLine_UnterminatedBlock(t *testing.T) {
	// actor block missing closing brace
	src := "system \"S\" {}\nactor \"Merchant\" {\n  description \"x\"\n"
	_, err := Parse(src)
	// This may or may not error (parser may tolerate EOF in block), but if it does error, check for line
	if err != nil {
		if !strings.Contains(err.Error(), "line") {
			t.Errorf("expected error to contain 'line', got: %v", err)
		}
	}
}

func TestParse_ErrorContainsLine_UnexpectedFieldInActor(t *testing.T) {
	src := "actor \"Merchant\" {\n  badfield \"x\"\n}"
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for bad field")
	}
	if !strings.Contains(err.Error(), "line 2:") {
		t.Errorf("expected error to contain 'line 2:', got: %v", err)
	}
}

func TestParse_FullModel(t *testing.T) {
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
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.System == nil || f.System.Name != "INCA" {
		t.Error("expected system INCA")
	}
	if len(f.Actors) != 1 {
		t.Errorf("expected 1 actor, got %d", len(f.Actors))
	}
	if len(f.Needs) != 1 {
		t.Errorf("expected 1 need, got %d", len(f.Needs))
	}
	if len(f.Capabilities) != 1 {
		t.Errorf("expected 1 capability, got %d", len(f.Capabilities))
	}
	if len(f.Capabilities[0].Children) != 1 {
		t.Errorf("expected 1 child capability, got %d", len(f.Capabilities[0].Children))
	}
	if len(f.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(f.Services))
	}
	if len(f.Teams) != 1 {
		t.Errorf("expected 1 team, got %d", len(f.Teams))
	}
	if len(f.Platforms) != 1 {
		t.Errorf("expected 1 platform, got %d", len(f.Platforms))
	}
	if len(f.Interactions) != 1 {
		t.Errorf("expected 1 interaction, got %d", len(f.Interactions))
	}
	if len(f.DataAssets) != 1 {
		t.Errorf("expected 1 data_asset, got %d", len(f.DataAssets))
	}
	if len(f.ExternalDependencies) != 1 {
		t.Errorf("expected 1 external_dependency, got %d", len(f.ExternalDependencies))
	}
	if len(f.Signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(f.Signals))
	}
}

// ---------------------------------------------------------------------------
// 5.5 — Import Syntax
// ---------------------------------------------------------------------------

func TestParse_Import_Simple(t *testing.T) {
	src := `import "other-file.unm"`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(f.Imports))
	}
	imp := f.Imports[0]
	if imp.Path != "other-file.unm" {
		t.Errorf("expected path %q, got %q", "other-file.unm", imp.Path)
	}
	if imp.Alias != "" {
		t.Errorf("expected empty alias, got %q", imp.Alias)
	}
}

func TestParse_Import_Named(t *testing.T) {
	src := `import actors from "shared/actors.unm"`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(f.Imports))
	}
	imp := f.Imports[0]
	if imp.Path != "shared/actors.unm" {
		t.Errorf("expected path %q, got %q", "shared/actors.unm", imp.Path)
	}
	if imp.Alias != "actors" {
		t.Errorf("expected alias %q, got %q", "actors", imp.Alias)
	}
}

func TestParse_Import_MultipleImports(t *testing.T) {
	src := `
import "common.unm"
import actors from "shared/actors.unm"
import teams from "shared/teams.unm"
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(f.Imports))
	}
	if f.Imports[0].Path != "common.unm" || f.Imports[0].Alias != "" {
		t.Errorf("unexpected first import: %+v", f.Imports[0])
	}
	if f.Imports[1].Path != "shared/actors.unm" || f.Imports[1].Alias != "actors" {
		t.Errorf("unexpected second import: %+v", f.Imports[1])
	}
	if f.Imports[2].Path != "shared/teams.unm" || f.Imports[2].Alias != "teams" {
		t.Errorf("unexpected third import: %+v", f.Imports[2])
	}
}

func TestParse_Import_MissingPath_Error(t *testing.T) {
	// "import" followed by EOF — should error
	src := `import`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for missing import path")
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected error to contain 'line', got: %v", err)
	}
}

func TestParse_Import_MissingFromKeyword_Error(t *testing.T) {
	// "import actors path.unm" missing "from"
	src := `import actors "missing-from.unm"`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for missing 'from' keyword")
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected error to contain 'line', got: %v", err)
	}
}

func TestParse_Import_WithOtherEntities(t *testing.T) {
	src := `
import "common.unm"
system "MySystem" {
  description "Test"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Imports) != 1 {
		t.Errorf("expected 1 import, got %d", len(f.Imports))
	}
	if f.System == nil || f.System.Name != "MySystem" {
		t.Error("expected system MySystem")
	}
}

// ---------------------------------------------------------------------------
// 5.6 — Inferred Mapping Syntax
// ---------------------------------------------------------------------------

func TestParse_InferredMapping_Complete(t *testing.T) {
	src := `
inferred {
    from "need-name"
    to "capability-name"
    confidence 0.85
    evidence "Detected via API scan of feed-api service"
    status suggested
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.InferredMappings) != 1 {
		t.Fatalf("expected 1 inferred mapping, got %d", len(f.InferredMappings))
	}
	im := f.InferredMappings[0]
	if im.From != "need-name" {
		t.Errorf("expected from %q, got %q", "need-name", im.From)
	}
	if im.To != "capability-name" {
		t.Errorf("expected to %q, got %q", "capability-name", im.To)
	}
	if im.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", im.Confidence)
	}
	if im.Evidence != "Detected via API scan of feed-api service" {
		t.Errorf("expected evidence %q, got %q", "Detected via API scan of feed-api service", im.Evidence)
	}
	if im.Status != "suggested" {
		t.Errorf("expected status %q, got %q", "suggested", im.Status)
	}
}

func TestParse_InferredMapping_Partial(t *testing.T) {
	// Only from and to, no confidence/evidence/status
	src := `
inferred {
    from "my-need"
    to "my-cap"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.InferredMappings) != 1 {
		t.Fatalf("expected 1 inferred mapping, got %d", len(f.InferredMappings))
	}
	im := f.InferredMappings[0]
	if im.From != "my-need" {
		t.Errorf("expected from %q, got %q", "my-need", im.From)
	}
	if im.To != "my-cap" {
		t.Errorf("expected to %q, got %q", "my-cap", im.To)
	}
	if im.Confidence != 0.0 {
		t.Errorf("expected default confidence 0.0, got %f", im.Confidence)
	}
}

func TestParse_InferredMapping_LowConfidence(t *testing.T) {
	src := `inferred { from "n" to "c" confidence 0.3 status confirmed }`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.InferredMappings) != 1 {
		t.Fatalf("expected 1 inferred mapping, got %d", len(f.InferredMappings))
	}
	if f.InferredMappings[0].Confidence != 0.3 {
		t.Errorf("expected confidence 0.3, got %f", f.InferredMappings[0].Confidence)
	}
}

func TestParse_InferredMapping_InvalidConfidence(t *testing.T) {
	src := `inferred { from "n" to "c" confidence not-a-number }`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for invalid confidence value")
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected error to contain 'line', got: %v", err)
	}
}

func TestParse_InferredMapping_Multiple(t *testing.T) {
	src := `
inferred {
    from "n1"
    to "c1"
    confidence 0.9
    status suggested
}
inferred {
    from "n2"
    to "c2"
    confidence 0.6
    status confirmed
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.InferredMappings) != 2 {
		t.Fatalf("expected 2 inferred mappings, got %d", len(f.InferredMappings))
	}
}

// ---------------------------------------------------------------------------
// 5.7 — Transition Syntax
// ---------------------------------------------------------------------------

func TestParse_TransitionBlock(t *testing.T) {
	src := `
transition "Consolidate catalog ownership" {
  description "Move from fragmented to stream-aligned catalog ownership"

  current {
    capability "Catalog publication" ownedBy team "Team A"
    capability "Catalog publication" ownedBy team "Team B"
  }

  target {
    capability "Catalog publication" ownedBy team "Catalog Stream"
  }

  step 1 "Align Team A and Team B" {
    action merge team "Team A" team "Team B" into team "Catalog Stream"
    expected_outcome "Single team owns ingestion and validation"
  }

  step 2 "Extract platform capabilities" {
    action extract capability "Catalog storage" to team "Catalog Platform"
    expected_outcome "Storage becomes x-as-a-service"
  }
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(f.Transitions))
	}
	tr := f.Transitions[0]
	if tr.Name != "Consolidate catalog ownership" {
		t.Errorf("expected name %q, got %q", "Consolidate catalog ownership", tr.Name)
	}
	if tr.Description != "Move from fragmented to stream-aligned catalog ownership" {
		t.Errorf("unexpected description: %q", tr.Description)
	}
	if len(tr.Current) != 2 {
		t.Fatalf("expected 2 current bindings, got %d", len(tr.Current))
	}
	if tr.Current[0].CapabilityName != "Catalog publication" || tr.Current[0].TeamName != "Team A" {
		t.Errorf("unexpected current[0]: %+v", tr.Current[0])
	}
	if tr.Current[1].TeamName != "Team B" {
		t.Errorf("unexpected current[1]: %+v", tr.Current[1])
	}
	if len(tr.Target) != 1 {
		t.Fatalf("expected 1 target binding, got %d", len(tr.Target))
	}
	if tr.Target[0].TeamName != "Catalog Stream" {
		t.Errorf("unexpected target[0]: %+v", tr.Target[0])
	}
	if len(tr.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(tr.Steps))
	}
	if tr.Steps[0].Number != 1 {
		t.Errorf("expected step 1 number=1, got %d", tr.Steps[0].Number)
	}
	if tr.Steps[0].Label != "Align Team A and Team B" {
		t.Errorf("unexpected step 1 label: %q", tr.Steps[0].Label)
	}
	if tr.Steps[0].ExpectedOutcome != "Single team owns ingestion and validation" {
		t.Errorf("unexpected step 1 expected_outcome: %q", tr.Steps[0].ExpectedOutcome)
	}
	if tr.Steps[1].Number != 2 {
		t.Errorf("expected step 2 number=2, got %d", tr.Steps[1].Number)
	}
}

func TestParse_TransitionBlock_MinimalEmpty(t *testing.T) {
	src := `transition "Empty plan" {}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(f.Transitions))
	}
	tr := f.Transitions[0]
	if tr.Name != "Empty plan" {
		t.Errorf("expected name %q, got %q", "Empty plan", tr.Name)
	}
	if len(tr.Current) != 0 || len(tr.Target) != 0 || len(tr.Steps) != 0 {
		t.Error("expected empty current/target/steps for minimal transition")
	}
}

func TestParse_TransitionBlock_UnknownField_Error(t *testing.T) {
	src := `transition "Bad" { badfield "value" }`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for unknown field in transition block")
	}
	if !strings.Contains(err.Error(), "unexpected field") {
		t.Errorf("expected 'unexpected field' in error, got: %v", err)
	}
}

// P0: Need outcome field
func TestParse_NeedOutcome(t *testing.T) {
	src := `
need "Fast checkout" {
  description "Checkout is fast"
  outcome "User completes checkout in under 3 clicks"
  actor "Shopper"
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Needs) != 1 {
		t.Fatalf("expected 1 need, got %d", len(f.Needs))
	}
	n := f.Needs[0]
	if n.Outcome != "User completes checkout in under 3 clicks" {
		t.Errorf("expected outcome %q, got %q", "User completes checkout in under 3 clicks", n.Outcome)
	}
	if n.Description != "Checkout is fast" {
		t.Errorf("expected description %q, got %q", "Checkout is fast", n.Description)
	}
}

// P0: Team size field
func TestParse_TeamSize(t *testing.T) {
	src := `
team "Platform" {
  type platform
  size 6
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(f.Teams))
	}
	if f.Teams[0].Size != 6 {
		t.Errorf("expected size 6, got %d", f.Teams[0].Size)
	}
}

// P0: Interaction via field
func TestParse_InteractionVia(t *testing.T) {
	src := `
interaction "service-a" -> "service-b" {
  via "api-gateway"
  mode x-as-a-service
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(f.Interactions))
	}
	if f.Interactions[0].Via != "api-gateway" {
		t.Errorf("expected via %q, got %q", "api-gateway", f.Interactions[0].Via)
	}
}

// P1: Colon shorthand for relationship descriptions
func TestParse_RelationshipColonShorthand(t *testing.T) {
	src := `
capability "Payment" {
  realizedBy "payment-service" : "primary implementation"
  dependsOn "fraud-service" : "risk checks"
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Capabilities) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(f.Capabilities))
	}
	cap := f.Capabilities[0]
	if len(cap.RealizedBy) != 1 {
		t.Fatalf("expected 1 realizedBy, got %d", len(cap.RealizedBy))
	}
	if cap.RealizedBy[0].Description != "primary implementation" {
		t.Errorf("expected realizedBy description %q, got %q", "primary implementation", cap.RealizedBy[0].Description)
	}
	if cap.DependsOn[0].Description != "risk checks" {
		t.Errorf("expected dependsOn description %q, got %q", "risk checks", cap.DependsOn[0].Description)
	}
}

// P1: DataAsset usedBy (plain service name)
func TestParse_DataAssetUsedByWithAccess(t *testing.T) {
	src := `
data_asset "payments-db" {
  type database
  usedBy "payment-service"
  usedBy "reporting-service"
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.DataAssets) != 1 {
		t.Fatalf("expected 1 data_asset, got %d", len(f.DataAssets))
	}
	da := f.DataAssets[0]
	if len(da.UsedBy) != 2 {
		t.Fatalf("expected 2 usedBy, got %d", len(da.UsedBy))
	}
	if da.UsedBy[0] != "payment-service" {
		t.Errorf("expected target %q, got %q", "payment-service", da.UsedBy[0])
	}
	if da.UsedBy[1] != "reporting-service" {
		t.Errorf("expected target %q, got %q", "reporting-service", da.UsedBy[1])
	}
}

// P1: DataAsset producedBy and consumedBy are not DSL fields — expect parse error
func TestParse_DataAssetProducedConsumed(t *testing.T) {
	src := `
data_asset "events" {
  type event-stream
  producedBy "event-service"
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected parse error for unknown field 'producedBy', got nil")
	}
}

// P1: ExternalDependency usedBy with colon description
func TestParse_ExternalDepUsedByWithDescription(t *testing.T) {
	src := `
external_dependency "stripe" {
  description "Payment gateway"
  usedBy "checkout-service" : "processes payments"
  usedBy "billing-service"
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.ExternalDependencies) != 1 {
		t.Fatalf("expected 1 external_dependency, got %d", len(f.ExternalDependencies))
	}
	ext := f.ExternalDependencies[0]
	if len(ext.UsedBy) != 2 {
		t.Fatalf("expected 2 usedBy, got %d", len(ext.UsedBy))
	}
	if ext.UsedBy[0].Target != "checkout-service" {
		t.Errorf("expected target %q, got %q", "checkout-service", ext.UsedBy[0].Target)
	}
	if ext.UsedBy[0].Description != "processes payments" {
		t.Errorf("expected description %q, got %q", "processes payments", ext.UsedBy[0].Description)
	}
	if ext.UsedBy[1].Target != "billing-service" {
		t.Errorf("expected target %q, got %q", "billing-service", ext.UsedBy[1].Target)
	}
	if ext.UsedBy[1].Description != "" {
		t.Errorf("expected empty description, got %q", ext.UsedBy[1].Description)
	}
}

// P2: `external` alias for external_dependency
func TestParse_ExternalAlias(t *testing.T) {
	src := `
external "twilio" {
  description "SMS provider"
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.ExternalDependencies) != 1 {
		t.Fatalf("expected 1 external_dependency via alias, got %d", len(f.ExternalDependencies))
	}
	if f.ExternalDependencies[0].Name != "twilio" {
		t.Errorf("expected name %q, got %q", "twilio", f.ExternalDependencies[0].Name)
	}
}

// P2: `data` alias for data_asset
func TestParse_DataAlias(t *testing.T) {
	src := `
data "user-profiles" {
  type database
  description "User profile store"
}`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.DataAssets) != 1 {
		t.Fatalf("expected 1 data_asset via alias, got %d", len(f.DataAssets))
	}
	if f.DataAssets[0].Name != "user-profiles" {
		t.Errorf("expected name %q, got %q", "user-profiles", f.DataAssets[0].Name)
	}
}

// ---------------------------------------------------------------------------
// 9.1.2 — Flat capabilities with parent in DSL
// ---------------------------------------------------------------------------

func TestParse_CapabilityFlatParent(t *testing.T) {
	src := `
capability "Child Cap" {
  parent "Parent Cap"
  visibility internal
}
capability "Parent Cap" {
  visibility domain
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Capabilities) != 2 {
		t.Fatalf("expected 2 capabilities, got %d", len(f.Capabilities))
	}
	child := f.Capabilities[0]
	if child.Name != "Child Cap" {
		t.Errorf("expected name %q, got %q", "Child Cap", child.Name)
	}
	if child.Parent != "Parent Cap" {
		t.Errorf("expected parent %q, got %q", "Parent Cap", child.Parent)
	}
	if child.Visibility != "internal" {
		t.Errorf("expected visibility %q, got %q", "internal", child.Visibility)
	}
}

func TestParse_CapabilityFlatParent_MultiLevel(t *testing.T) {
	src := `
capability "Grandchild" {
  parent "Child"
}
capability "Child" {
  parent "Root"
}
capability "Root" {
  visibility domain
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Capabilities) != 3 {
		t.Fatalf("expected 3 capabilities, got %d", len(f.Capabilities))
	}
	// verify parent references are stored
	grandchild := f.Capabilities[0]
	if grandchild.Parent != "Child" {
		t.Errorf("expected grandchild parent %q, got %q", "Child", grandchild.Parent)
	}
	child := f.Capabilities[1]
	if child.Parent != "Root" {
		t.Errorf("expected child parent %q, got %q", "Root", child.Parent)
	}
}

func TestParse_CapabilityMixedFlatAndNested(t *testing.T) {
	src := `
capability "Root" {
  visibility domain
  capability "Nested Child" {
    visibility foundational
  }
}
capability "Flat Child" {
  parent "Root"
  visibility user-facing
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 top-level capabilities: Root (with nested child) and Flat Child
	if len(f.Capabilities) != 2 {
		t.Fatalf("expected 2 top-level capabilities, got %d", len(f.Capabilities))
	}
	root := f.Capabilities[0]
	if len(root.Children) != 1 {
		t.Errorf("expected 1 nested child, got %d", len(root.Children))
	}
	flatChild := f.Capabilities[1]
	if flatChild.Parent != "Root" {
		t.Errorf("expected flat child parent %q, got %q", "Root", flatChild.Parent)
	}
}

// ---------------------------------------------------------------------------
// 9.3.5 — realizes on service blocks in DSL
// ---------------------------------------------------------------------------

func TestParse_ServiceRealizes(t *testing.T) {
	src := `
service "my-service" {
  ownedBy "team-a"
  realizes "Cap A"
  realizes "Cap B" role "supporting"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(f.Services))
	}
	s := f.Services[0]
	if len(s.Realizes) != 2 {
		t.Fatalf("expected 2 realizes, got %d", len(s.Realizes))
	}
	if s.Realizes[0].Target != "Cap A" {
		t.Errorf("expected realizes[0] target %q, got %q", "Cap A", s.Realizes[0].Target)
	}
	if s.Realizes[0].Role != "" {
		t.Errorf("expected realizes[0] role to be empty, got %q", s.Realizes[0].Role)
	}
	if s.Realizes[1].Target != "Cap B" {
		t.Errorf("expected realizes[1] target %q, got %q", "Cap B", s.Realizes[1].Target)
	}
	if s.Realizes[1].Role != "supporting" {
		t.Errorf("expected realizes[1] role %q, got %q", "supporting", s.Realizes[1].Role)
	}
}

// ---------------------------------------------------------------------------
// 9.4.3 — externalDeps on service blocks in DSL
// ---------------------------------------------------------------------------

func TestParse_ServiceExternalDeps(t *testing.T) {
	src := `
service "my-service" {
  ownedBy "team-a"
  externalDeps "temporal"
  externalDeps "postgres"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(f.Services))
	}
	s := f.Services[0]
	if len(s.ExternalDeps) != 2 {
		t.Fatalf("expected 2 externalDeps, got %d", len(s.ExternalDeps))
	}
	if s.ExternalDeps[0] != "temporal" {
		t.Errorf("expected externalDeps[0] %q, got %q", "temporal", s.ExternalDeps[0])
	}
	if s.ExternalDeps[1] != "postgres" {
		t.Errorf("expected externalDeps[1] %q, got %q", "postgres", s.ExternalDeps[1])
	}
}

// ---------------------------------------------------------------------------
// 9.5.3 — interacts on team blocks in DSL
// ---------------------------------------------------------------------------

func TestParse_TeamInteracts(t *testing.T) {
	src := `
team "team-a" {
  type stream-aligned
  interacts "team-b" mode x-as-a-service via "API"
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(f.Teams))
	}
	team := f.Teams[0]
	if len(team.Interacts) != 1 {
		t.Fatalf("expected 1 interacts, got %d", len(team.Interacts))
	}
	inter := team.Interacts[0]
	if inter.With != "team-b" {
		t.Errorf("expected with %q, got %q", "team-b", inter.With)
	}
	if inter.Mode != "x-as-a-service" {
		t.Errorf("expected mode %q, got %q", "x-as-a-service", inter.Mode)
	}
	if inter.Via != "API" {
		t.Errorf("expected via %q, got %q", "API", inter.Via)
	}
}

func TestParse_TeamInteracts_MultipleInteractions(t *testing.T) {
	src := `
team "team-a" {
  type stream-aligned
  interacts "team-b" mode x-as-a-service via "API"
  interacts "team-c" mode collaboration
}
`
	f, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	team := f.Teams[0]
	if len(team.Interacts) != 2 {
		t.Fatalf("expected 2 interacts, got %d", len(team.Interacts))
	}
	if team.Interacts[1].With != "team-c" {
		t.Errorf("expected second interaction with %q, got %q", "team-c", team.Interacts[1].With)
	}
	if team.Interacts[1].Via != "" {
		t.Errorf("expected empty via, got %q", team.Interacts[1].Via)
	}
}

