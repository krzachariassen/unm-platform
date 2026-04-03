package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGaps_MissingModelID_Returns400(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/views/gaps", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleGaps_UnknownModelID_Returns404(t *testing.T) {
	h := newTestHandler(t)
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/views/gaps?model_id=nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleGaps_ValidModel_Returns200WithShape(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/views/gaps?model_id="+id, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp gapsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, id, resp.ModelID)
	// enrichedTestYAML has: Track order (unmapped), Auth (unrealized), orphan-svc (orphan)
	assert.Contains(t, resp.UnmappedNeeds, "Track order")
	assert.Contains(t, resp.UnrealizedCapabilities, "Auth")
	assert.Contains(t, resp.OrphanServices, "orphan-svc")
	// Fields must be non-nil slices (not null in JSON)
	assert.NotNil(t, resp.UnownedServices)
	assert.NotNil(t, resp.UnneededCapabilities)
}
