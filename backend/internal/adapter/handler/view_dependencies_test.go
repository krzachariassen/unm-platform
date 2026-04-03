package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDependencies_MissingModelID_Returns400(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/views/dependencies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleDependencies_UnknownModelID_Returns404(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/views/dependencies?model_id=nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleDependencies_ValidModel_Returns200WithShape(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/views/dependencies?model_id="+id, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp dependenciesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, id, resp.ModelID)
	// Fields must be present (non-nil slices, numeric depths)
	assert.NotNil(t, resp.ServiceCycles)
	assert.NotNil(t, resp.CapabilityCycles)
	assert.GreaterOrEqual(t, resp.MaxServiceDepth, 0)
	assert.GreaterOrEqual(t, resp.MaxCapabilityDepth, 0)
	assert.NotNil(t, resp.CriticalServicePath)
}

func TestHandleDependencies_ValidModel_CapabilityDepth(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/views/dependencies?model_id="+id, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp dependenciesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// enrichedTestYAML has Search → Auth (1 dependency edge), so depth >= 2
	assert.GreaterOrEqual(t, resp.MaxCapabilityDepth, 2)
}
