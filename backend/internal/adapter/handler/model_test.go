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
	stored := h.store.Get(id)
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
