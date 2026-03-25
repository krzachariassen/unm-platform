package parser_test

import (
	"strings"
	"testing"

	"github.com/uber/unm-platform/internal/infrastructure/parser"
)

func TestNewParserForPath_YAML(t *testing.T) {
	p := parser.NewParserForPath("model.unm.yaml")
	if _, ok := p.(*parser.YAMLParser); !ok {
		t.Errorf("expected *parser.YAMLParser for .unm.yaml extension, got %T", p)
	}
}

func TestNewParserForPath_YAMLWithDir(t *testing.T) {
	p := parser.NewParserForPath("/some/dir/model.unm.yaml")
	if _, ok := p.(*parser.YAMLParser); !ok {
		t.Errorf("expected *parser.YAMLParser for .unm.yaml extension with directory, got %T", p)
	}
}

func TestNewParserForPath_DSL(t *testing.T) {
	p := parser.NewParserForPath("model.unm")
	if _, ok := p.(*parser.DSLParser); !ok {
		t.Errorf("expected *parser.DSLParser for .unm extension, got %T", p)
	}
}

func TestNewParserForPath_DSLWithDir(t *testing.T) {
	p := parser.NewParserForPath("/some/dir/model.unm")
	if _, ok := p.(*parser.DSLParser); !ok {
		t.Errorf("expected *parser.DSLParser for .unm extension with directory, got %T", p)
	}
}

func TestNewParserForPath_ArbitraryExtension_UsesYAML(t *testing.T) {
	// Anything that is not .unm should use YAML parser
	p := parser.NewParserForPath("model.yaml")
	if _, ok := p.(*parser.YAMLParser); !ok {
		t.Errorf("expected *parser.YAMLParser for .yaml extension, got %T", p)
	}
}

func TestDSLParser_Parse_SimpleModel(t *testing.T) {
	src := `system "Test System" {
  description "A test system"
}
actor "User" {
  description "A basic user"
}
`
	p := parser.NewDSLParser()
	model, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("expected no error parsing simple DSL, got: %v", err)
	}
	if model == nil {
		t.Fatal("expected non-nil model")
	}
	if model.System.Name != "Test System" {
		t.Errorf("expected system name 'Test System', got %q", model.System.Name)
	}
}

func TestDSLParser_Parse_InvalidDSL(t *testing.T) {
	src := `unknown_keyword "something" { }`
	p := parser.NewDSLParser()
	_, err := p.Parse(strings.NewReader(src))
	if err == nil {
		t.Error("expected error parsing invalid DSL, got nil")
	}
}

func TestDSLParser_Parse_EmptyInput(t *testing.T) {
	src := ``
	p := parser.NewDSLParser()
	_, err := p.Parse(strings.NewReader(src))
	// Empty input = no system name — should error during transform
	if err == nil {
		t.Error("expected error parsing empty DSL (no system), got nil")
	}
}
