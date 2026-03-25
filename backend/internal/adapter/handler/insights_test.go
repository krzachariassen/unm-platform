package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
