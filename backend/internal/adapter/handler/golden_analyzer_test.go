package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// runGoldenAnalyzerTest hits GET /api/models/{id}/analyze/{analyzeType} with the nexus model
// and compares the response against a golden file.
func runGoldenAnalyzerTest(t *testing.T, analyzeType string) {
	t.Helper()
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/analyze/"+analyzeType, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "analyze %s returned non-200: %s", analyzeType, w.Body.String())

	goldenFile := filepath.Join(goldenDir(t), "nexus-"+analyzeType+"-signals.json")
	checkGolden(t, goldenFile, w.Body.Bytes())
}

// ── Bottleneck ─────────────────────────────────────────────────────────────────

func TestGolden_BottleneckAnalyzer(t *testing.T) {
	// Bottleneck analyzer sorts by fan-in desc, then fan-out desc, then name asc — deterministic.
	runGoldenAnalyzerTest(t, "bottleneck")
}

// ── Coupling ──────────────────────────────────────────────────────────────────

func TestGolden_CouplingAnalyzer(t *testing.T) {
	// Coupling analyzer sorts couplings alphabetically — deterministic.
	runGoldenAnalyzerTest(t, "coupling")
}

// ── Complexity ────────────────────────────────────────────────────────────────

func TestGolden_ComplexityAnalyzer(t *testing.T) {
	// Complexity analyzer sorts by TotalComplexity desc, then name asc — deterministic.
	runGoldenAnalyzerTest(t, "complexity")
}

// ── Fragmentation ─────────────────────────────────────────────────────────────

func TestGolden_FragmentationAnalyzer(t *testing.T) {
	// NOTE: FragmentedCapabilities is built from map iteration — non-deterministic order.
	// We verify structure and presence of fragmented capabilities, not exact ordering.
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/analyze/fragmentation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "fragmentation analyze returned non-200: %s", w.Body.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, "fragmentation", result["type"])
	_, ok := result["fragmented_capabilities"].([]any)
	require.True(t, ok, "fragmented_capabilities must be an array")
}

// ── Gap ───────────────────────────────────────────────────────────────────────

func TestGolden_GapAnalyzer(t *testing.T) {
	// NOTE: Gap report arrays are built from map iteration — non-deterministic order.
	// We verify structure and required fields, not exact content or ordering.
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/analyze/gaps", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "gaps analyze returned non-200: %s", w.Body.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, "gaps", result["type"])
	require.Contains(t, result, "unmapped_needs")
	require.Contains(t, result, "unrealized_capabilities")
	require.Contains(t, result, "unowned_services")
	require.Contains(t, result, "unneeded_capabilities")
}

// ── Cognitive Load ────────────────────────────────────────────────────────────

func TestGolden_CognitiveLoadAnalyzer(t *testing.T) {
	// NOTE: team_loads is built from map iteration — non-deterministic order.
	// We verify structure and presence of team loads, not exact ordering.
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/analyze/cognitive-load", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "cognitive-load analyze returned non-200: %s", w.Body.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, "cognitive-load", result["type"])
	teamLoads, ok := result["team_loads"].([]any)
	require.True(t, ok, "team_loads must be an array")
	require.NotEmpty(t, teamLoads, "team_loads must not be empty")

	// Verify each team load has required fields.
	for _, tl := range teamLoads {
		tlMap, ok := tl.(map[string]any)
		require.True(t, ok)
		require.Contains(t, tlMap, "team")
		require.Contains(t, tlMap, "capability_count")
		require.Contains(t, tlMap, "service_count")
		require.Contains(t, tlMap, "overall_level")
	}
}
