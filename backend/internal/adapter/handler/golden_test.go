package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	domainservice "github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// goldenDir returns the absolute path to the testdata/golden directory.
// Go tests run with cwd = the package directory, so we walk up from there.
func goldenDir(t *testing.T) string {
	t.Helper()
	// Package dir: backend/internal/adapter/handler/
	// testdata/golden: backend/testdata/golden
	cwd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(cwd, "..", "..", "..", "testdata", "golden")
}

// nexusYAMLPath returns the path to examples/nexus.unm.yaml.
// Test cwd is backend/internal/adapter/handler/, examples/ is at repo root.
func nexusYAMLPath(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// handler → adapter → internal → backend → unm-platform (repo root) → examples
	return filepath.Join(cwd, "..", "..", "..", "..", "examples", "nexus.unm.yaml")
}

// setupNexusTestModel parses examples/nexus.unm.yaml and returns a router + model ID.
func setupNexusTestModel(t *testing.T) (http.Handler, string) {
	t.Helper()

	nexusPath := nexusYAMLPath(t)
	f, err := os.Open(nexusPath)
	if err != nil {
		t.Skipf("nexus.unm.yaml not found at %s — skipping golden tests", nexusPath)
	}
	defer f.Close()

	store := repository.NewModelStore()
	cfg := entity.DefaultConfig()
	h := New(HandlerDeps{
		Config:            cfg,
		ParseAndValidate:  usecase.NewParseAndValidate(parser.NewYAMLParser(), domainservice.NewValidationEngine()),
		Fragmentation:     analyzer.NewFragmentationAnalyzer(),
		CognitiveLoad:     analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		Dependency:        analyzer.NewDependencyAnalyzer(),
		Gap:               analyzer.NewGapAnalyzer(),
		Bottleneck:        analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		Coupling:          analyzer.NewCouplingAnalyzer(),
		Complexity:        analyzer.NewComplexityAnalyzer(),
		Interactions:      analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		Unlinked:          analyzer.NewUnlinkedCapabilityAnalyzer(),
		SignalSuggestions: analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		ValueChain:        analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		ValueStream:       analyzer.NewValueStreamAnalyzer(),
		ChangesetStore:    repository.NewChangesetStore(),
		ImpactAnalyzer:    analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		AIClient:          nil,
		Store:             store,
	})
	router := NewRouter(h)

	p := parser.NewYAMLParser()
	v := domainservice.NewValidationEngine()
	uc := usecase.NewParseAndValidate(p, v)
	model, _, err := uc.Execute(f)
	require.NoError(t, err, "parse nexus.unm.yaml")

	id, err := store.Store(model)
	require.NoError(t, err)

	return router, id
}

// checkGolden compares responseBody against a golden file.
// If UPDATE_GOLDEN=1 is set, it writes the golden file instead.
func checkGolden(t *testing.T, goldenFile string, responseBody []byte) {
	t.Helper()

	// Normalise: pretty-print so diffs are readable.
	var pretty any
	require.NoError(t, json.Unmarshal(responseBody, &pretty))
	normalized, err := json.MarshalIndent(pretty, "", "  ")
	require.NoError(t, err)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenFile), 0o755))
		require.NoError(t, os.WriteFile(goldenFile, append(normalized, '\n'), 0o644))
		t.Logf("Updated golden file: %s", goldenFile)
		return
	}

	goldenData, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 to generate: %s", goldenFile)
	require.JSONEq(t, strings.TrimSpace(string(goldenData)), string(normalized))
}

// runGoldenViewTest hits a view endpoint with the nexus model and calls checkGolden.
func runGoldenViewTest(t *testing.T, viewType string) {
	t.Helper()
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/"+viewType, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "view %s returned non-200: %s", viewType, w.Body.String())

	goldenFile := filepath.Join(goldenDir(t), "nexus-"+viewType+"-view.json")
	checkGolden(t, goldenFile, w.Body.Bytes())
}

// ── Golden tests ──────────────────────────────────────────────────────────────

func TestGolden_NeedView(t *testing.T) {
	runGoldenViewTest(t, "need")
}

func TestGolden_CapabilityView(t *testing.T) {
	runGoldenViewTest(t, "capability")
}

func TestGolden_OwnershipView(t *testing.T) {
	runGoldenViewTest(t, "ownership")
}

func TestGolden_TeamTopologyView(t *testing.T) {
	runGoldenViewTest(t, "team-topology")
}

func TestGolden_CognitiveLoadView(t *testing.T) {
	// NOTE: The cognitive-load view's team_loads array is built from map iteration,
	// making its order non-deterministic. We verify structure but not exact ordering.
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/cognitive-load", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "cognitive-load view returned non-200: %s", w.Body.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, "cognitive-load", result["view_type"])
	teamLoads, ok := result["team_loads"].([]any)
	require.True(t, ok, "team_loads must be an array")
	require.NotEmpty(t, teamLoads, "team_loads must not be empty")
}

func TestGolden_SignalsView(t *testing.T) {
	// NOTE: The signals view contains arrays built from map iteration (fragmented
	// capabilities, low-coherence teams, etc.) whose order is non-deterministic.
	// We verify the view returns 200 and is valid JSON, but do not compare against
	// a golden snapshot until the signals service sorts its output deterministically.
	router, id := setupNexusTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/signals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "signals view returned non-200: %s", w.Body.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, "signals", result["view_type"])
}
