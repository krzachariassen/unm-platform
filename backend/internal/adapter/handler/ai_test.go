package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// skipUnlessAITests skips the current test unless UNM_AI_TESTS=true is set.
// This is separate from UNM_OPENAI_API_KEY: the key may be present in CI for
// the running server, but AI tests should only run in a dedicated daily job.
func skipUnlessAITests(t *testing.T) {
	t.Helper()
	if os.Getenv("UNM_AI_TESTS") != "true" {
		t.Skip("UNM_AI_TESTS not enabled — skipping real AI test (set UNM_AI_TESTS=true to run)")
	}
}

// newTestHandlerWithAI constructs a Handler wired with a real OpenAI client.
// Requires UNM_OPENAI_API_KEY to be set (always is when UNM_AI_TESTS=true in CI).
func newTestHandlerWithAI(t *testing.T) *Handler {
	t.Helper()
	skipUnlessAITests(t)
	var aiClient *ai.OpenAIClient
	c, err := ai.NewOpenAIClient()
	require.NoError(t, err)
	aiClient = c
	cfg := entity.DefaultConfig()
	return New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
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
		aiClient,
		repository.NewModelStore(),
	)
}

// ── POST /api/models/{id}/ask — no AI configured ────────────────────────────────

func TestHandleAsk_NotConfigured_Returns200WithMessage(t *testing.T) {
	h := newTestHandler(t) // aiClient is nil
	modelID := parseAndStoreModel(t, h, validYAML)

	body := `{"question": "What are the biggest risks?", "category": "general"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["ai_configured"].(bool))
	assert.NotEmpty(t, resp["answer"])
}

func TestHandleAsk_ModelNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)

	body := `{"question": "What are the biggest risks?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/nonexistent/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleAsk_InvalidCategory_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	body := `{"question": "test?", "category": "not-a-real-category"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleAsk_DefaultsToGeneralCategory(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	// No category field — should default to "general" and not fail
	body := `{"question": "What teams are in this system?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	// With nil aiClient it returns 200 with "not configured" message
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "general", resp["category"])
}

func TestHandleAsk_EmptyQuestion_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	body := `{"question": "   ", "category": "general"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "question")
}

// ── POST /api/models/{id}/ask — real OpenAI ─────────────────────────────────────

func TestHandleAsk_RealAI_General(t *testing.T) {
	skipUnlessAITests(t)

	h := newTestHandlerWithAI(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	body := `{"question": "How many teams are in this system?", "category": "general"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["ai_configured"].(bool), "ai_configured must be true")
	assert.NotEmpty(t, resp["answer"], "answer must be non-empty")
	assert.Equal(t, modelID, resp["model_id"])
	assert.Equal(t, "general", resp["category"])
}

func TestHandleAsk_RealAI_StructuralLoad(t *testing.T) {
	skipUnlessAITests(t)

	h := newTestHandlerWithAI(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	body := `{"question": "Which teams have the highest cognitive load?", "category": "structural-load"}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["ai_configured"].(bool))
	assert.NotEmpty(t, resp["answer"])
}

// ── POST /api/models/{id}/changesets/{csId}/explain — real OpenAI ───────────────

func TestHandleExplainChangeset_NotConfigured_Returns200WithMessage(t *testing.T) {
	h := newTestHandler(t) // aiClient is nil
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	// Create a changeset
	csBody := `{
		"id": "cs-explain",
		"description": "Move browse-svc",
		"actions": [{"type": "move_service", "service_name": "browse-svc", "from_team_name": "Team Alpha", "to_team_name": "Team Beta"}]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(csBody))
	createW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp map[string]any
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	csID := createResp["id"].(string)

	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets/"+csID+"/explain", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["ai_configured"].(bool))
	assert.NotEmpty(t, resp["explanation"])
}

func TestHandleExplainChangeset_RealAI(t *testing.T) {
	skipUnlessAITests(t)

	h := newTestHandlerWithAI(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	// Create a changeset
	csBody := `{
		"id": "cs-explain-ai",
		"description": "Move browse-svc to Team Beta",
		"actions": [{"type": "move_service", "service_name": "browse-svc", "from_team_name": "Team Alpha", "to_team_name": "Team Beta"}]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(csBody))
	createW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp map[string]any
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	csID := createResp["id"].(string)

	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets/"+csID+"/explain", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["ai_configured"].(bool))
	assert.NotEmpty(t, resp["explanation"])
	assert.Equal(t, csID, resp["changeset_id"])
}
