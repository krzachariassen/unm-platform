package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
)

// newTestHandlerWithFakeAI returns a Handler that has a configured (non-nil) AI client
// using a fake API key. IsConfigured() returns true so the handler does not short-circuit
// before checking the insight cache. The client will fail on any real API call.
func newTestHandlerWithFakeAI(t *testing.T) *Handler {
	t.Helper()
	h := newTestHandler(t)
	h.aiClient = ai.NewOpenAIClientWithKey("fake-key-for-tests", "gpt-4o-mini")
	return h
}

// TestHandleInsights_CachedFailedEntry_Returns200WithStatusFailed proves that when
// the background AI computation fails and the result is cached, the handler returns
// HTTP 200 with status "failed" and an empty (non-null) insights object.
// This is a regression test for 6.10.9: AI failures must surface as status:"failed",
// not as HTTP error codes, so the frontend can distinguish "no findings" from "AI failed".
func TestHandleInsights_CachedFailedEntry_Returns200WithStatusFailed(t *testing.T) {
	h := newTestHandlerWithFakeAI(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	// Pre-seed the cache with a failed entry, simulating what computeAndCacheInsight
	// stores after a background AI call fails (e.g. OpenAI timeout or bad key).
	cacheKey := modelID + ":signals"
	failedResponse := InsightsResponse{
		Domain:       "signals",
		Status:       "failed",
		Insights:     map[string]InsightItem{},
		AIConfigured: true,
		Error:        "ai_unavailable",
	}
	h.insightCache.Store(cacheKey, insightEntry{status: "failed", response: failedResponse})

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/insights/signals", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "AI failures must return HTTP 200, not an error code")

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "failed", resp["status"], "status must be 'failed' when AI computation failed")

	insights, ok := resp["insights"].(map[string]any)
	require.True(t, ok, "insights must be a JSON object (not null) even on failure")
	assert.Empty(t, insights, "insights must be empty when AI failed")

	assert.True(t, resp["ai_configured"].(bool), "ai_configured must be true (AI was configured but failed)")
}

func TestHandleInsights_NoAIClient_Returns200WithAIConfiguredFalse(t *testing.T) {
	h := newTestHandler(t) // aiClient is nil
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/insights/signals", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["ai_configured"].(bool), "ai_configured must be false when AI client is nil")

	insights, ok := resp["insights"].(map[string]any)
	require.True(t, ok, "insights must be an object")
	assert.Empty(t, insights, "insights must be empty when AI not configured")
}

func TestHandleInsights_InvalidDomain_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/insights/not-a-domain", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError, "response must contain 'error' field")
}

func TestHandleInsights_UnknownModel_Returns404(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/does-not-exist/insights/signals", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
