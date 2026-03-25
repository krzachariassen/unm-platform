package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// doAnalyzeRequest fires a POST /api/models/analyze/{analyzeType} with body and returns the recorder.
func doAnalyzeRequest(t *testing.T, h *Handler, analyzeType, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/models/analyze/"+analyzeType, strings.NewReader(body))
	req.SetPathValue("type", analyzeType)
	w := httptest.NewRecorder()
	h.handleAnalyze(w, req)
	return w
}

func TestHandleAnalyze_Fragmentation(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "fragmentation", validYAML)

	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "fragmentation", result["type"])
	assert.Contains(t, result, "fragmented_capabilities")
}

func TestHandleAnalyze_CognitiveLoad(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "cognitive-load", validYAML)

	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "cognitive-load", result["type"])
	assert.Contains(t, result, "team_loads")
}

func TestHandleAnalyze_Dependencies(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "dependencies", validYAML)

	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "dependencies", result["type"])
	assert.Contains(t, result, "max_service_depth")
	assert.Contains(t, result, "max_capability_depth")
	assert.Contains(t, result, "critical_service_path")
	assert.Contains(t, result, "service_cycles")
	assert.Contains(t, result, "capability_cycles")
}

func TestHandleAnalyze_Gaps(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "gaps", validYAML)

	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "gaps", result["type"])
	assert.Contains(t, result, "unmapped_needs")
	assert.Contains(t, result, "unrealized_capabilities")
	assert.Contains(t, result, "unowned_services")
	assert.Contains(t, result, "unneeded_capabilities")
}

func TestHandleAnalyze_All(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "all", validYAML)

	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Equal(t, "all", result["type"])
	assert.Contains(t, result, "fragmentation")
	assert.Contains(t, result, "cognitive_load")
	assert.Contains(t, result, "dependencies")
	assert.Contains(t, result, "gaps")
}

func TestHandleAnalyze_UnknownType(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "bogus", validYAML)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	errMsg, ok := result["error"].(string)
	require.True(t, ok, "error field must be a string")
	assert.Contains(t, errMsg, "bogus")
}

func TestHandleAnalyze_InvalidYAML(t *testing.T) {
	h := newTestHandler(t)
	w := doAnalyzeRequest(t, h, "fragmentation", malformedYAML)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Contains(t, result, "error")
	errMsg, ok := result["error"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, errMsg)
}
