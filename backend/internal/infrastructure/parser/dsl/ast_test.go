package dsl

import (
	"testing"
)

func TestRelationshipNode_DescriptionAndRole(t *testing.T) {
	r := RelationshipNode{
		Target:      "my-service",
		Description: "Handles routing",
		Role:        "primary",
	}
	if r.Target != "my-service" {
		t.Errorf("expected target %q, got %q", "my-service", r.Target)
	}
	if r.Description != "Handles routing" {
		t.Errorf("expected description %q, got %q", "Handles routing", r.Description)
	}
	if r.Role != "primary" {
		t.Errorf("expected role %q, got %q", "primary", r.Role)
	}
}

func TestRelationshipNode_EmptyRole(t *testing.T) {
	r := RelationshipNode{Target: "svc"}
	if r.Role != "" {
		t.Errorf("expected empty role, got %q", r.Role)
	}
}

func TestCapabilityNode_NestedChildren(t *testing.T) {
	child := &CapabilityNode{Name: "Child Cap"}
	parent := &CapabilityNode{
		Name:     "Parent Cap",
		Children: []*CapabilityNode{child},
	}
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0].Name != "Child Cap" {
		t.Errorf("expected child name %q, got %q", "Child Cap", parent.Children[0].Name)
	}
}

func TestCapabilityNode_DeepNesting(t *testing.T) {
	grandchild := &CapabilityNode{Name: "Grandchild"}
	child := &CapabilityNode{
		Name:     "Child",
		Children: []*CapabilityNode{grandchild},
	}
	root := &CapabilityNode{
		Name:     "Root",
		Children: []*CapabilityNode{child},
	}
	if root.Children[0].Children[0].Name != "Grandchild" {
		t.Errorf("expected grandchild name %q, got %q", "Grandchild", root.Children[0].Children[0].Name)
	}
}

func TestFile_AllSlicesPresent(t *testing.T) {
	f := &File{}
	// Verify zero values are nil slices (valid initial state)
	if f.System != nil {
		t.Error("expected nil system")
	}
	if f.Actors != nil {
		t.Error("expected nil actors")
	}
}

func TestImportNode_SimpleForm(t *testing.T) {
	n := &ImportNode{Path: "other.unm"}
	if n.Path != "other.unm" {
		t.Errorf("expected path %q, got %q", "other.unm", n.Path)
	}
	if n.Alias != "" {
		t.Errorf("expected empty alias, got %q", n.Alias)
	}
}

func TestImportNode_NamedForm(t *testing.T) {
	n := &ImportNode{Path: "shared/actors.unm", Alias: "actors"}
	if n.Alias != "actors" {
		t.Errorf("expected alias %q, got %q", "actors", n.Alias)
	}
	if n.Path != "shared/actors.unm" {
		t.Errorf("expected path %q, got %q", "shared/actors.unm", n.Path)
	}
}

func TestInferredMappingNode_Fields(t *testing.T) {
	n := &InferredMappingNode{
		From:       "my-service",
		To:         "my-capability",
		Confidence: 0.75,
		Evidence:   "found via scan",
		Status:     "suggested",
	}
	if n.From != "my-service" {
		t.Errorf("expected From %q, got %q", "my-service", n.From)
	}
	if n.To != "my-capability" {
		t.Errorf("expected To %q, got %q", "my-capability", n.To)
	}
	if n.Confidence != 0.75 {
		t.Errorf("expected Confidence 0.75, got %f", n.Confidence)
	}
	if n.Evidence != "found via scan" {
		t.Errorf("expected Evidence %q, got %q", "found via scan", n.Evidence)
	}
	if n.Status != "suggested" {
		t.Errorf("expected Status %q, got %q", "suggested", n.Status)
	}
}

func TestParseError_Error(t *testing.T) {
	pe := &ParseError{Line: 7, Message: "unexpected token"}
	got := pe.Error()
	expected := "line 7: unexpected token"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestParseError_ImplementsError(t *testing.T) {
	var err error = &ParseError{Line: 1, Message: "test"}
	if err == nil {
		t.Error("*ParseError should implement error interface")
	}
}
