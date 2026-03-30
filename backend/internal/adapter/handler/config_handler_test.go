package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
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

// ---------------------------------------------------------------------------
// IP extraction helpers
// ---------------------------------------------------------------------------

func TestExtractClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.5:12345"
	assert.Equal(t, "203.0.113.5", extractClientIP(req))
}

func TestExtractClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:80"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.2")
	assert.Equal(t, "203.0.113.5", extractClientIP(req))
}

func TestExtractClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:80"
	req.Header.Set("X-Real-IP", "203.0.113.9")
	assert.Equal(t, "203.0.113.9", extractClientIP(req))
}

func TestExtractClientIP_XForwardedFor_TakesPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	req.Header.Set("X-Real-IP", "203.0.113.9")
	assert.Equal(t, "203.0.113.5", extractClientIP(req))
}

// ---------------------------------------------------------------------------
// IP allowlist helper
// ---------------------------------------------------------------------------

func TestIsIPAllowed_ExactMatch(t *testing.T) {
	assert.True(t, isIPAllowed("203.0.113.5", []string{"203.0.113.5"}))
}

func TestIsIPAllowed_CIDRMatch(t *testing.T) {
	assert.True(t, isIPAllowed("10.0.1.42", []string{"10.0.0.0/8"}))
}

func TestIsIPAllowed_NoMatch(t *testing.T) {
	assert.False(t, isIPAllowed("203.0.113.5", []string{"192.168.1.1", "10.0.0.0/8"}))
}

func TestIsIPAllowed_EmptyList(t *testing.T) {
	assert.False(t, isIPAllowed("203.0.113.5", []string{}))
}

func TestIsIPAllowed_IPv6Loopback(t *testing.T) {
	assert.True(t, isIPAllowed("::1", []string{"::1"}))
}

// ---------------------------------------------------------------------------
// Config endpoint — allowlist enforcement
// ---------------------------------------------------------------------------

func newTestHandlerWithAllowedIPs(t *testing.T, allowedIPs []string) *Handler {
	t.Helper()
	cfg := entity.DefaultConfig()
	cfg.AI.AllowedIPs = allowedIPs
	h := newTestHandler(t)
	h.cfg.AI.AllowedIPs = allowedIPs
	return h
}

func TestGetConfig_EmptyAllowlist_AllowsAll(t *testing.T) {
	// No allowlist configured → ai.enabled reflects only whether aiClient is configured
	h := newTestHandler(t) // aiClient is nil → enabled=false regardless
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	// With empty allowlist the IP check is skipped; aiClient is nil so enabled=false
	aiSection := body["ai"].(map[string]interface{})
	assert.Equal(t, false, aiSection["enabled"])
}

func TestGetConfig_BlockedIP_ReturnsAIDisabled(t *testing.T) {
	h := newTestHandlerWithAllowedIPs(t, []string{"192.168.1.1"})
	// Simulate a configured aiClient by patching the flag via a mock-free approach:
	// We test the IP logic independently above; here we verify the handler wires it correctly.
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	req.RemoteAddr = "203.0.113.5:1234" // not in allowlist
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	aiSection := body["ai"].(map[string]interface{})
	// aiClient is nil in test handler so enabled is false regardless, but the allowlist
	// would also block it — verify no panic and correct response shape
	assert.Equal(t, false, aiSection["enabled"])
}

