package valueobject_test

import (
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

func TestSourceType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		val      valueobject.SourceType
		expected string
	}{
		{"model_fact", valueobject.SourceModelFact, "model_fact"},
		{"analyzer_finding", valueobject.SourceAnalyzerFinding, "analyzer_finding"},
		{"ai_interpretation", valueobject.SourceAIInterpretation, "ai_interpretation"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.val) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.val))
			}
		})
	}
}
