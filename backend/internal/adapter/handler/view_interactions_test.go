package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleInteractions_MissingModelID_Returns400(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/views/interactions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleInteractions_UnknownModelID_Returns404(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/views/interactions?model_id=nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleInteractions_ValidModel_Returns200WithShape(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/views/interactions?model_id="+id, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp interactionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, id, resp.ModelID)
	// enrichedTestYAML has 1 interaction: Catalog Team → Search Team via x-as-a-service
	assert.NotNil(t, resp.ModeDistribution)
	assert.NotNil(t, resp.IsolatedTeams)
	assert.NotNil(t, resp.OverReliantTeams)
}

func TestHandleInteractions_ValidModel_IsolatedTeams(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/views/interactions?model_id="+id, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp interactionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// enrichedTestYAML: Menu Team has no interactions → isolated
	assert.Contains(t, resp.IsolatedTeams, "Menu Team")
}

func TestHandleInteractions_ValidModel_ModeDistribution(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/views/interactions?model_id="+id, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp interactionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// enrichedTestYAML has 1 x-as-a-service interaction
	xas, ok := resp.ModeDistribution["x-as-a-service"]
	assert.True(t, ok, "x-as-a-service must be in mode_distribution")
	assert.Equal(t, 1, xas)
}
