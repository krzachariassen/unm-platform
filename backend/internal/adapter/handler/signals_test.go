package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSignals_StoredModel_Returns200WithLayers(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, validYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/views/signals", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "user_experience_layer")
	assert.Contains(t, resp, "architecture_layer")
	assert.Contains(t, resp, "organization_layer")
}

func TestHandleSignals_UnknownModel_Returns404(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/does-not-exist/views/signals", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
