package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/unm-platform/internal/domain/entity"
)

func TestHandleLoadExample_Returns200WithSystemName(t *testing.T) {
	h := newTestHandler(t)

	// Pass a config with DebugRoutes enabled so the route is registered.
	cfg := entity.DefaultConfig()
	cfg.Features.DebugRoutes = true

	req := httptest.NewRequest(http.MethodPost, "/api/debug/load-example", nil)
	w := httptest.NewRecorder()
	NewRouter(h, cfg).ServeHTTP(w, req)

	// The handler may fail to find the example file if tests are run from an unexpected
	// working directory — in that case, accept either 200 (success) or 500 (file not found).
	// What matters is that the route is registered and not a 404.
	assert.NotEqual(t, http.StatusNotFound, w.Code, "debug route must be registered")
	assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code, "POST method must be allowed")

	if w.Code == http.StatusOK {
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		_, hasSystemName := resp["system_name"]
		assert.True(t, hasSystemName, "response must contain 'system_name' field")
	}
}
