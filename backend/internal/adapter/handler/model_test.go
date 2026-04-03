package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// newTestHandler constructs a Handler wired with real implementations for testing.
func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	cfg := entity.DefaultConfig()
	return New(HandlerDeps{
		Config:            cfg,
		ParseAndValidate:  usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
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
		AIClient:          nil, // aiClient
		Store:             repository.NewModelStore(),
	})
}

// validYAML is a minimal complete UNM model that passes validation.
const validYAML = `system:
  name: "Test System"
actors:
  - name: "User"
needs:
  - name: "Test Need"
    actor: "User"
    supportedBy:
      - "Test Cap"
capabilities:
  - name: "Test Cap"
services:
  - name: "Test Svc"
    ownedBy: "Team One"
    realizes:
      - "Test Cap"
teams:
  - name: "Team One"
    type: "stream-aligned"
`

// invalidModelYAML is a model that parses OK but fails validation (service with no owner).
const invalidModelYAML = `system:
  name: "Invalid System"
services:
  - name: "Orphan Svc"
`

// malformedYAML is text that cannot be parsed as a UNM model.
const malformedYAML = `{[not valid yaml at all`

// validDSL is a minimal complete UNM model in DSL format that passes validation.
const validDSL = `system "DSL Test System" {
  description "A test system in DSL format"
}

actor "User" {
  description "A basic user"
}

need "Test Need" {
  actor "User"
  supportedBy "Test Cap"
}

capability "Test Cap" {
  description "Core capability"
}

service "Test Svc" {
  description "Core service"
  ownedBy "Team One"
  realizes "Test Cap"
}

team "Team One" {
  type "stream-aligned"
  description "The main team"
}
`

func TestHandleParse_DSL_Returns200(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/parse?format=dsl", strings.NewReader(validDSL))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "DSL Test System", resp["system_name"])

	summary, ok := resp["summary"].(map[string]any)
	require.True(t, ok, "response must contain 'summary' object")
	assert.EqualValues(t, 1, summary["actors"])
	assert.EqualValues(t, 1, summary["capabilities"])
	assert.EqualValues(t, 1, summary["services"])
	assert.EqualValues(t, 1, summary["teams"])

	validation, ok := resp["validation"].(map[string]any)
	require.True(t, ok, "response must contain 'validation' object")
	assert.True(t, validation["is_valid"].(bool))
}

func TestHandleValidate_DSL_Returns200(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate?format=dsl", strings.NewReader(validDSL))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	isValid, ok := resp["is_valid"].(bool)
	require.True(t, ok, "response must contain bool 'is_valid'")
	assert.True(t, isValid)
}

func TestHandleParse_InvalidDSL_Returns400(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/parse?format=dsl", strings.NewReader(`unknown_kw "bad" {}`))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError, "response must contain 'error' field for invalid DSL")
}

func TestHandleParse_NoFormatParam_UsesYAML(t *testing.T) {
	h := newTestHandler(t)

	// Valid YAML sent without ?format=dsl should still work
	req := httptest.NewRequest(http.MethodPost, "/api/models/parse", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}


func TestHandleParse_ValidYAML_Returns200WithSummary(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/parse", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Must include an ID
	id, ok := resp["id"].(string)
	require.True(t, ok, "response must contain string 'id'")
	assert.NotEmpty(t, id)

	// Must include system_name
	assert.Equal(t, "Test System", resp["system_name"])

	// Must include summary with entity counts
	summary, ok := resp["summary"].(map[string]any)
	require.True(t, ok, "response must contain 'summary' object")
	assert.EqualValues(t, 1, summary["actors"])
	assert.EqualValues(t, 1, summary["needs"])
	assert.EqualValues(t, 1, summary["capabilities"])
	assert.EqualValues(t, 1, summary["services"])
	assert.EqualValues(t, 1, summary["teams"])

	// Must include validation object
	validation, ok := resp["validation"].(map[string]any)
	require.True(t, ok, "response must contain 'validation' object")
	assert.True(t, validation["is_valid"].(bool))
}

func TestHandleParse_ValidYAML_StoredModelIsRetrievable(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/parse", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	id := resp["id"].(string)
	stored, err := h.store.Get(id)
	require.NoError(t, err)
	require.NotNil(t, stored, "model must be retrievable from the store using the returned ID")
	assert.Equal(t, "Test System", stored.Model.System.Name)
}

func TestHandleParse_InvalidYAML_Returns400(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/parse", strings.NewReader(malformedYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError, "response must contain 'error' field")
}

// ── POST /api/models/validate ──────────────────────────────────────────────────

func TestHandleValidate_ValidYAML_Returns200IsValidTrue(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	isValid, ok := resp["is_valid"].(bool)
	require.True(t, ok, "response must contain bool 'is_valid'")
	assert.True(t, isValid)

	errors, ok := resp["errors"].([]any)
	require.True(t, ok, "response must contain 'errors' array")
	assert.Empty(t, errors)
}

func TestHandleValidate_InvalidModel_Returns200IsValidFalseWithErrors(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate", strings.NewReader(invalidModelYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	// Validation failure is domain data, not an HTTP error — must be 200
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	isValid, ok := resp["is_valid"].(bool)
	require.True(t, ok, "response must contain bool 'is_valid'")
	assert.False(t, isValid)

	errors, ok := resp["errors"].([]any)
	require.True(t, ok, "response must contain 'errors' array")
	assert.NotEmpty(t, errors, "errors array must be non-empty for invalid model")

	// Each error must have code, message, entity fields
	firstErr, ok := errors[0].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, firstErr["code"])
	assert.NotEmpty(t, firstErr["message"])
}

func TestHandleValidate_InvalidYAML_Returns400(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate", strings.NewReader(malformedYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError, "response must contain 'error' field")
}

func TestHandleValidate_DoesNotStoreModel(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Validate should not return an id field
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasID := resp["id"]
	assert.False(t, hasID, "validate endpoint must not return an 'id' (model should not be stored)")
}

func TestHandleValidate_ResponseIncludesWarningsArray(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate", strings.NewReader(validYAML))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	_, hasWarnings := resp["warnings"]
	assert.True(t, hasWarnings, "response must contain 'warnings' array")
}

// ── Format auto-detection ──────────────────────────────────────────────────────

func TestHandleParse_DSLContentAutoDetected(t *testing.T) {
	h := newTestHandler(t)

	// Send DSL content WITHOUT ?format=dsl — handler must auto-detect.
	req := httptest.NewRequest(http.MethodPost, "/api/models/parse", strings.NewReader(validDSL))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.handleParse(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "DSL auto-detect must succeed: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "DSL Test System", resp["system_name"])
}

func TestHandleValidate_DSLContentAutoDetected(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/validate", strings.NewReader(validDSL))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.handleValidate(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "DSL auto-detect validate must succeed: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	isValid, ok := resp["is_valid"].(bool)
	require.True(t, ok)
	assert.True(t, isValid)
}

// ── GET /api/models ────────────────────────────────────────────────────────────

func TestHandleListModels_Returns200(t *testing.T) {
	h := newTestHandler(t)
	// Store a model first.
	parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	models, ok := resp["models"].([]any)
	require.True(t, ok, "response must contain 'models' array")
	assert.GreaterOrEqual(t, len(models), 1)

	total, ok := resp["total"].(float64)
	require.True(t, ok, "response must contain numeric 'total'")
	assert.GreaterOrEqual(t, int(total), 1)

	// Each model entry must have id, name, version_count.
	first := models[0].(map[string]any)
	assert.NotEmpty(t, first["id"])
	assert.NotEmpty(t, first["name"])
	_, hasVC := first["version_count"]
	assert.True(t, hasVC, "model entry must have version_count")
}

func TestHandleListModels_Empty(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	models, ok := resp["models"].([]any)
	require.True(t, ok)
	assert.Empty(t, models)
	assert.EqualValues(t, 0, resp["total"])
}

// ── GET /api/models/{id}/history ──────────────────────────────────────────────

func TestHandleListVersions_Returns200(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/history", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, modelID, resp["model_id"])
	versions, ok := resp["versions"].([]any)
	require.True(t, ok, "response must contain 'versions' array")
	assert.GreaterOrEqual(t, len(versions), 1)

	v := versions[0].(map[string]any)
	assert.NotEmpty(t, v["id"])
	assert.EqualValues(t, 1, v["version"])
}

func TestHandleListVersions_UnknownModel_Returns404(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent-id/history", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GET /api/models/{id}/versions/{v} ─────────────────────────────────────────

func TestHandleGetVersion_ValidVersion_Returns200(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/versions/1", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "Test System", resp["system_name"])
	_, hasValidation := resp["validation"]
	assert.True(t, hasValidation, "response must contain 'validation'")
}

func TestHandleGetVersion_InvalidVersion_Returns404(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/versions/99", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetVersion_BadVersionParam_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/versions/bad", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── GET /api/models/{id}/diff ──────────────────────────────────────────────────

func TestHandleDiffVersions_Returns200(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	// Memory store only has version 1; diff v1→v1 is valid and returns empty change sets.
	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/diff?from=1&to=1", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, modelID, resp["model_id"])
	assert.EqualValues(t, 1, resp["from_version"])
	assert.EqualValues(t, 1, resp["to_version"])
	_, hasAdded := resp["added"]
	assert.True(t, hasAdded, "response must contain 'added'")
	_, hasRemoved := resp["removed"]
	assert.True(t, hasRemoved, "response must contain 'removed'")
}

func TestHandleDiffVersions_MissingParams_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/diff", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
