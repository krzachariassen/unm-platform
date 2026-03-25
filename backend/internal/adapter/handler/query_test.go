package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/unm-platform/internal/adapter/repository"
	"github.com/uber/unm-platform/internal/domain/entity"
	domainservice "github.com/uber/unm-platform/internal/domain/service"
	"github.com/uber/unm-platform/internal/infrastructure/analyzer"
	"github.com/uber/unm-platform/internal/infrastructure/parser"
	"github.com/uber/unm-platform/internal/usecase"
)

// setupTestModel parses a minimal model, stores it, and returns the router + model ID.
func setupTestModel(t *testing.T) (http.Handler, string) {
	t.Helper()
	store := repository.NewModelStore()
	cfg := entity.DefaultConfig()
	h := New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), domainservice.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		repository.NewChangesetStore(),
		analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		nil, // aiClient
		store,
	)
	router := NewRouter(h)

	p := parser.NewYAMLParser()
	v := domainservice.NewValidationEngine()
	uc := usecase.NewParseAndValidate(p, v)
	model, _, err := uc.Execute(strings.NewReader(validYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	id, err := store.Store(model)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	return router, id
}

// ── GET /api/models/{id}/capabilities ─────────────────────────────────────────

func TestHandleQueryCapabilities_KnownID_Returns200WithCapabilities(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/capabilities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	caps, ok := resp["capabilities"].([]any)
	require.True(t, ok, "response must contain 'capabilities' array")
	assert.Len(t, caps, 1)

	cap0 := caps[0].(map[string]any)
	assert.Equal(t, "Test Cap", cap0["name"])
	_, hasIsLeaf := cap0["is_leaf"]
	assert.True(t, hasIsLeaf, "capability must include 'is_leaf'")
}

func TestHandleQueryCapabilities_UnknownID_Returns404(t *testing.T) {
	router, _ := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/capabilities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

// ── GET /api/models/{id}/teams ────────────────────────────────────────────────

func TestHandleQueryTeams_KnownID_Returns200WithTeams(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/teams", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	teams, ok := resp["teams"].([]any)
	require.True(t, ok, "response must contain 'teams' array")
	assert.Len(t, teams, 1)

	team0 := teams[0].(map[string]any)
	assert.Equal(t, "Team One", team0["name"])
	assert.Equal(t, "stream-aligned", team0["type"])
	_, hasCapCount := team0["capability_count"]
	assert.True(t, hasCapCount, "team must include 'capability_count'")
	_, hasOverloaded := team0["is_overloaded"]
	assert.True(t, hasOverloaded, "team must include 'is_overloaded'")
}

func TestHandleQueryTeams_UnknownID_Returns404(t *testing.T) {
	router, _ := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/teams", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GET /api/models/{id}/needs ────────────────────────────────────────────────

func TestHandleQueryNeeds_KnownID_Returns200WithNeeds(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/needs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	needs, ok := resp["needs"].([]any)
	require.True(t, ok, "response must contain 'needs' array")
	assert.Len(t, needs, 1)

	need0 := needs[0].(map[string]any)
	assert.Equal(t, "Test Need", need0["name"])
	assert.Equal(t, "User", need0["actor_name"])
	_, hasMapped := need0["is_mapped"]
	assert.True(t, hasMapped, "need must include 'is_mapped'")
}

func TestHandleQueryNeeds_UnknownID_Returns404(t *testing.T) {
	router, _ := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/needs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GET /api/models/{id}/services ─────────────────────────────────────────────

func TestHandleQueryServices_KnownID_Returns200WithServices(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	services, ok := resp["services"].([]any)
	require.True(t, ok, "response must contain 'services' array")
	assert.Len(t, services, 1)

	svc0 := services[0].(map[string]any)
	assert.Equal(t, "Test Svc", svc0["name"])
	assert.Equal(t, "Team One", svc0["owner_team_name"])
}

func TestHandleQueryServices_UnknownID_Returns404(t *testing.T) {
	router, _ := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GET /api/models/{id}/actors ───────────────────────────────────────────────

func TestHandleQueryActors_KnownID_Returns200WithActors(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/actors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	actors, ok := resp["actors"].([]any)
	require.True(t, ok, "response must contain 'actors' array")
	assert.Len(t, actors, 1)

	actor0 := actors[0].(map[string]any)
	assert.Equal(t, "User", actor0["name"])
}

func TestHandleQueryActors_UnknownID_Returns404(t *testing.T) {
	router, _ := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/actors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
