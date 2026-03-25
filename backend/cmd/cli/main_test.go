package main

import (
	"strings"
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
)

func TestRunParseCommand_ValidFile(t *testing.T) {
	var buf strings.Builder
	code := runParseCommand([]string{"../../testdata/simple.unm.yaml"}, &buf)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Simple System") {
		t.Errorf("expected output to contain system name 'Simple System', got:\n%s", out)
	}
	if !strings.Contains(out, "Actors:") {
		t.Errorf("expected output to contain 'Actors:', got:\n%s", out)
	}
	if !strings.Contains(out, "Capabilities:") {
		t.Errorf("expected output to contain 'Capabilities:', got:\n%s", out)
	}
	if !strings.Contains(out, "Validation: PASSED") {
		t.Errorf("expected output to contain 'Validation: PASSED', got:\n%s", out)
	}
}

func TestRunParseCommand_NonExistentFile(t *testing.T) {
	var buf strings.Builder
	code := runParseCommand([]string{"/nonexistent/path/model.unm.yaml"}, &buf)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing file, got %d\noutput:\n%s", code, buf.String())
	}
}

func TestRunParseCommand_InvalidModel(t *testing.T) {
	var buf strings.Builder
	code := runParseCommand([]string{"../../testdata/invalid.unm.yaml"}, &buf)
	if code != 2 {
		t.Fatalf("expected exit code 2 for invalid model, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Validation: FAILED") {
		t.Errorf("expected output to contain 'Validation: FAILED', got:\n%s", out)
	}
}

func TestRunValidateCommand_ValidFile(t *testing.T) {
	var buf strings.Builder
	code := runValidateCommand([]string{"../../testdata/simple.unm.yaml"}, &buf)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Validation: PASSED") {
		t.Errorf("expected output to contain 'Validation: PASSED', got:\n%s", out)
	}
}

func TestRunValidateCommand_NonExistentFile(t *testing.T) {
	var buf strings.Builder
	code := runValidateCommand([]string{"/nonexistent/path/model.unm.yaml"}, &buf)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing file, got %d\noutput:\n%s", code, buf.String())
	}
}

func TestRunAnalyzeCommand_FragmentationOnValidFile(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"fragmentation", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Fragmentation") {
		t.Errorf("expected output to contain 'Fragmentation', got:\n%s", out)
	}
}

func TestRunAnalyzeCommand_CognitiveLoad(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"cognitive-load", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Structural Load Assessment") {
		t.Errorf("expected output to contain 'Structural Load Assessment', got:\n%s", out)
	}
}

func TestRunAnalyzeCommand_Dependencies(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"dependencies", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Dependencies") {
		t.Errorf("expected output to contain 'Dependencies', got:\n%s", out)
	}
}

func TestRunAnalyzeCommand_Gaps(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"gaps", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Gaps") {
		t.Errorf("expected output to contain 'Gaps', got:\n%s", out)
	}
}

func TestRunAnalyzeCommand_All(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"all", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	for _, section := range []string{"Fragmentation", "Structural Load Assessment", "Dependencies", "Gaps", "Bottleneck", "Coupling", "Complexity"} {
		if !strings.Contains(out, section) {
			t.Errorf("expected 'analyze all' output to contain %q, got:\n%s", section, out)
		}
	}
}

func TestRunAnalyzeCommand_Bottleneck(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"bottleneck", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Bottleneck") {
		t.Errorf("expected output to contain 'Bottleneck', got:\n%s", buf.String())
	}
}

func TestRunAnalyzeCommand_Coupling(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"coupling", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Coupling") {
		t.Errorf("expected output to contain 'Coupling', got:\n%s", buf.String())
	}
}

func TestRunAnalyzeCommand_Complexity(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"complexity", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Complexity") {
		t.Errorf("expected output to contain 'Complexity', got:\n%s", buf.String())
	}
}

func TestRunAnalyzeCommand_BottleneckOnINCA(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"bottleneck", "../../../examples/inca.unm.extended.yaml"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Bottleneck") {
		t.Errorf("expected output to contain 'Bottleneck', got:\n%s", out)
	}
	// core and registry should appear as high fan-in services
	if !strings.Contains(out, "core") {
		t.Errorf("expected 'core' service to appear in bottleneck output, got:\n%s", out)
	}
}

// ── DSL format tests ──────────────────────────────────────────────────────────

func TestRunParseCommand_DSLFile(t *testing.T) {
	var buf strings.Builder
	code := runParseCommand([]string{"../../testdata/simple.unm"}, &buf)
	if code != 0 {
		t.Fatalf("expected exit code 0 for DSL file, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Simple DSL System") {
		t.Errorf("expected output to contain system name 'Simple DSL System', got:\n%s", out)
	}
	if !strings.Contains(out, "Actors:") {
		t.Errorf("expected output to contain 'Actors:', got:\n%s", out)
	}
}

func TestRunParseCommand_UnknownExtension_UsesYAML(t *testing.T) {
	// A .yaml file (not .unm.yaml) is routed to the YAML parser.
	// Use a non-existent path so we just test that it tries YAML and fails at file open.
	var buf strings.Builder
	code := runParseCommand([]string{"/nonexistent/model.yaml"}, &buf)
	// Should fail with exit code 1 (file not found), not panic
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing .yaml file, got %d\noutput:\n%s", code, buf.String())
	}
}

func TestRunValidateCommand_DSLFile(t *testing.T) {
	var buf strings.Builder
	code := runValidateCommand([]string{"../../testdata/simple.unm"}, &buf)
	if code != 0 {
		t.Fatalf("expected exit code 0 for DSL validate, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Validation:") {
		t.Errorf("expected output to contain 'Validation:', got:\n%s", out)
	}
}

func TestRunAnalyzeCommand_DSLFile(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"fragmentation", "../../testdata/simple.unm"}, &buf, entity.DefaultConfig())
	if code != 0 {
		t.Fatalf("expected exit code 0 for DSL analyze, got %d\noutput:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Fragmentation") {
		t.Errorf("expected output to contain 'Fragmentation', got:\n%s", out)
	}
}

func TestRunAnalyzeCommand_UnknownSubcommand(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"bogus", "../../testdata/simple.unm.yaml"}, &buf, entity.DefaultConfig())
	if code != 1 {
		t.Fatalf("expected exit code 1 for unknown subcommand, got %d", code)
	}
}

func TestRunAnalyzeCommand_MissingArgs(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{}, &buf, entity.DefaultConfig())
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing args, got %d", code)
	}
}

func TestRunAnalyzeCommand_NonExistentFile(t *testing.T) {
	var buf strings.Builder
	code := runAnalyzeCommand([]string{"fragmentation", "/nonexistent.yaml"}, &buf, entity.DefaultConfig())
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing file, got %d", code)
	}
}
