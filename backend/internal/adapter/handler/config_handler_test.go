package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetConfig(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)

	// Verify analysis section exists with expected defaults
	analysis, ok := body["analysis"].(map[string]interface{})
	require.True(t, ok, "response must contain 'analysis' object")

	assert.EqualValues(t, 5, analysis["default_team_size"])
	assert.EqualValues(t, 6, analysis["overloaded_capability_threshold"])

	// Verify bottleneck sub-section
	bn, ok := analysis["bottleneck"].(map[string]interface{})
	require.True(t, ok)
	assert.EqualValues(t, 5, bn["fan_in_warning"])
	assert.EqualValues(t, 10, bn["fan_in_critical"])

	// Verify signals sub-section
	sig, ok := analysis["signals"].(map[string]interface{})
	require.True(t, ok)
	assert.EqualValues(t, 2, sig["need_team_span_warning"])
	assert.EqualValues(t, 3, sig["need_team_span_critical"])
	assert.EqualValues(t, 3, sig["high_span_service_threshold"])
	assert.EqualValues(t, 4, sig["interaction_over_reliance"])
	assert.EqualValues(t, 4, sig["depth_chain_threshold"])

	// Verify value_chain sub-section
	vc, ok := analysis["value_chain"].(map[string]interface{})
	require.True(t, ok)
	assert.EqualValues(t, 3, vc["at_risk_team_span"])

	// Verify AI section only exposes safe fields (enabled flag, no secrets)
	aiSection, hasAI := body["ai"].(map[string]interface{})
	require.True(t, hasAI, "response must contain 'ai' object")
	_, hasEnabled := aiSection["enabled"]
	assert.True(t, hasEnabled, "ai section must contain 'enabled' flag")
	assert.Len(t, aiSection, 1, "ai section must only contain 'enabled' — no secrets")

	_, hasServer := body["server"]
	assert.False(t, hasServer, "server config must not be in response")
}
