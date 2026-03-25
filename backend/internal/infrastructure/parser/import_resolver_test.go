package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uber/unm-platform/internal/infrastructure/parser"
)

func TestParseDSLFile_SimpleFile(t *testing.T) {
	// Write a temp .unm file
	dir := t.TempDir()
	path := filepath.Join(dir, "simple.unm")
	content := `system "My System" {
  description "A simple system for testing"
}
actor "Operator" {
  description "Runs the system"
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	model, err := parser.ParseDSLFile(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if model == nil {
		t.Fatal("expected non-nil model")
	}
	if model.System.Name != "My System" {
		t.Errorf("expected system name 'My System', got %q", model.System.Name)
	}
}

func TestParseDSLFile_NonExistentFile(t *testing.T) {
	_, err := parser.ParseDSLFile("/tmp/does-not-exist-xyzzy.unm")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestParseDSLFile_CircularImportDetection(t *testing.T) {
	// File A imports file B, and file B imports file A → circular import error.
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.unm")
	pathB := filepath.Join(dir, "b.unm")

	contentA := `system "A" {}
import "b.unm"
`
	contentB := `system "B" {}
import "a.unm"
`
	if err := os.WriteFile(pathA, []byte(contentA), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.WriteFile(pathB, []byte(contentB), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := parser.ParseDSLFile(pathA)
	if err == nil {
		t.Fatal("expected circular import error, got nil")
	}
	if !strings.Contains(err.Error(), "circular import") {
		t.Errorf("expected 'circular import' in error, got: %v", err)
	}
}

func TestParseDSLFile_SelfImportDetection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "self.unm")
	content := `system "Self" {}
import "self.unm"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := parser.ParseDSLFile(path)
	if err == nil {
		t.Fatal("expected circular import error for self-import, got nil")
	}
	if !strings.Contains(err.Error(), "circular import") {
		t.Errorf("expected 'circular import' in error, got: %v", err)
	}
}

func TestParseDSLFile_ImportMergesEntities(t *testing.T) {
	dir := t.TempDir()

	// Imported file defines actors and services
	imported := `system "Shared" {}
actor "Merchant" {
  description "Sells things"
}
service "feed-api" {
  description "Feed API"
  ownedBy "catalog-team"
}
`
	if err := os.WriteFile(filepath.Join(dir, "shared.unm"), []byte(imported), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Main file imports shared.unm and adds its own entities
	main := `system "Main" {
  description "Main system"
}
import "shared.unm"

actor "Eater" {
  description "Eats things"
}
`
	mainPath := filepath.Join(dir, "main.unm")
	if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	model, err := parser.ParseDSLFile(mainPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Main system name should be preserved
	if model.System.Name != "Main" {
		t.Errorf("expected system name 'Main', got %q", model.System.Name)
	}

	// Both actors should be present
	if _, ok := model.Actors["Merchant"]; !ok {
		t.Error("expected Merchant actor from import, not found")
	}
	if _, ok := model.Actors["Eater"]; !ok {
		t.Error("expected Eater actor from main file, not found")
	}

	// Service from import should be present
	if _, ok := model.Services["feed-api"]; !ok {
		t.Error("expected feed-api service from import, not found")
	}
}

func TestParseDSLFile_TransitiveImports(t *testing.T) {
	dir := t.TempDir()

	// C defines an actor
	c := `system "C" {}
actor "Driver" {
  description "Drives"
}
`
	if err := os.WriteFile(filepath.Join(dir, "c.unm"), []byte(c), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// B imports C
	b := `system "B" {}
import "c.unm"
actor "Merchant" {
  description "Sells"
}
`
	if err := os.WriteFile(filepath.Join(dir, "b.unm"), []byte(b), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// A imports B → transitively gets C's entities too
	a := `system "A" {
  description "Top level"
}
import "b.unm"
actor "Eater" {
  description "Eats"
}
`
	aPath := filepath.Join(dir, "a.unm")
	if err := os.WriteFile(aPath, []byte(a), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	model, err := parser.ParseDSLFile(aPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if model.System.Name != "A" {
		t.Errorf("expected system name 'A', got %q", model.System.Name)
	}
	for _, expected := range []string{"Driver", "Merchant", "Eater"} {
		if _, ok := model.Actors[expected]; !ok {
			t.Errorf("expected actor %q (transitive import), not found", expected)
		}
	}
}

func TestParseDSLFile_ImportFileNotFound(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.unm")
	content := `system "Main" {}
import "does-not-exist.unm"
`
	if err := os.WriteFile(mainPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := parser.ParseDSLFile(mainPath)
	if err == nil {
		t.Fatal("expected error for missing import file, got nil")
	}
}

func TestParseDSLFile_ImportDuplicateEntity(t *testing.T) {
	dir := t.TempDir()

	// Both files define the same actor
	imported := `system "Shared" {}
actor "Merchant" {
  description "From import"
}
`
	if err := os.WriteFile(filepath.Join(dir, "shared.unm"), []byte(imported), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	main := `system "Main" {}
import "shared.unm"
actor "Merchant" {
  description "From main"
}
`
	mainPath := filepath.Join(dir, "main.unm")
	if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := parser.ParseDSLFile(mainPath)
	if err == nil {
		t.Fatal("expected duplicate entity error, got nil")
	}
}

func TestParseDSLFile_InvalidDSLSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.unm")
	if err := os.WriteFile(path, []byte(`blarg "nope" { }`), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	_, err := parser.ParseDSLFile(path)
	if err == nil {
		t.Error("expected error for invalid DSL syntax, got nil")
	}
}
