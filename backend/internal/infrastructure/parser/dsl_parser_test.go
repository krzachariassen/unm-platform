package parser

import (
	"strings"
	"testing"
)

func TestDSLParser_SimpleModel(t *testing.T) {
	src := `
system "Test" {}
actor "User" {}
`
	p := NewDSLParser()
	model, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.System.Name != "Test" {
		t.Errorf("expected system name %q, got %q", "Test", model.System.Name)
	}
	if _, ok := model.Actors["User"]; !ok {
		t.Error("expected actor 'User' in model")
	}
}

func TestDSLParser_FullModel(t *testing.T) {
	src := `
system "Full System" {
  description "Complete DSL model test"
}

actor "Merchant" {
  description "Sells goods"
}

need "Process Orders" {
  actor "Merchant"
  supportedBy "Order Processing"
}

capability "Order Processing" {
  visibility domain
}

service "order-service" {
  ownedBy "orders-team"
  realizes "Order Processing" role "primary"
}

team "orders-team" {
  type stream-aligned
  owns "order-service"
}
`
	p := NewDSLParser()
	model, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.System.Name != "Full System" {
		t.Errorf("expected %q, got %q", "Full System", model.System.Name)
	}
	if len(model.Actors) != 1 {
		t.Errorf("expected 1 actor, got %d", len(model.Actors))
	}
	if len(model.Needs) != 1 {
		t.Errorf("expected 1 need, got %d", len(model.Needs))
	}
	if len(model.Capabilities) != 1 {
		t.Errorf("expected 1 capability, got %d", len(model.Capabilities))
	}
	if len(model.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(model.Services))
	}
	if len(model.Teams) != 1 {
		t.Errorf("expected 1 team, got %d", len(model.Teams))
	}
}

func TestDSLParser_ParseError(t *testing.T) {
	src := `unknown "foo" {}`
	p := NewDSLParser()
	_, err := p.Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for unknown keyword")
	}
}

func TestDSLParser_TransformError(t *testing.T) {
	// Missing system means transform will error
	src := `actor "User" {}`
	p := NewDSLParser()
	_, err := p.Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error when system is missing")
	}
}
