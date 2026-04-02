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

// modelWithTwoTeamsYAML is a model with two teams and a service, suitable for changeset testing.
const modelWithTwoTeamsYAML = `system:
  name: "Changeset Test System"
actors:
  - name: "User"
needs:
  - name: "Browse"
    actor: "User"
    supportedBy:
      - "Browsing"
capabilities:
  - name: "Browsing"
services:
  - name: "browse-svc"
    ownedBy: "Team Alpha"
    realizes:
      - "Browsing"
teams:
  - name: "Team Alpha"
    type: "stream-aligned"
  - name: "Team Beta"
    type: "stream-aligned"
`

// parseAndStoreModel is a test helper that parses a YAML model via the handler
// and returns the model ID.
func parseAndStoreModel(t *testing.T, h *Handler, yaml string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/models/parse", strings.NewReader(yaml))
	req.Header.Set("Content-Type", "application/yaml")
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "parse must succeed: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	id, ok := resp["id"].(string)
	require.True(t, ok)
	return id
}

// ── POST /api/models/{id}/changesets ────────────────────────────────────────────

func TestCreateChangeset_Returns201(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	body := `{
		"id": "cs-1",
		"description": "Move browse-svc to Team Beta",
		"actions": [
			{
				"type": "move_service",
				"service_name": "browse-svc",
				"from_team_name": "Team Alpha",
				"to_team_name": "Team Beta"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, modelID, resp["model_id"])
	assert.Equal(t, "Move browse-svc to Team Beta", resp["description"])
	assert.EqualValues(t, 1, resp["action_count"])
}

func TestCreateChangeset_ModelNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)

	body := `{"id": "cs-1", "description": "test", "actions": []}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/nonexistent/changesets", strings.NewReader(body))
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateChangeset_InvalidAction_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	// move_service without required fields
	body := `{
		"id": "cs-bad",
		"description": "bad action",
		"actions": [{"type": "move_service"}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateChangeset_EmptyID_Returns400(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	body := `{"id": "", "description": "no id", "actions": []}`
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── GET /api/models/{id}/changesets/{csId} ──────────────────────────────────────

func TestGetChangeset_Returns200(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	// Create a changeset first
	body := `{
		"id": "cs-get",
		"description": "test get",
		"actions": [
			{
				"type": "move_service",
				"service_name": "browse-svc",
				"from_team_name": "Team Alpha",
				"to_team_name": "Team Beta"
			}
		]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	createW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp map[string]any
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	csID := createResp["id"].(string)

	// Get the changeset
	getReq := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/changesets/"+csID, nil)
	getW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var getResp map[string]any
	require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &getResp))
	assert.Equal(t, csID, getResp["id"])
	assert.Equal(t, modelID, getResp["model_id"])
	assert.Equal(t, "test get", getResp["description"])
	assert.NotEmpty(t, getResp["created_at"])

	actions, ok := getResp["actions"].([]any)
	require.True(t, ok)
	assert.Len(t, actions, 1)
}

func TestGetChangeset_NotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/changesets/nonexistent", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetChangeset_ModelNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/changesets/any", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GET /api/models/{id}/changesets/{csId}/projected ────────────────────────────

func TestProjectedModel_Returns200(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	body := `{
		"id": "cs-proj",
		"description": "project test",
		"actions": [
			{
				"type": "move_service",
				"service_name": "browse-svc",
				"from_team_name": "Team Alpha",
				"to_team_name": "Team Beta"
			}
		]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	createW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp map[string]any
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	csID := createResp["id"].(string)

	// Get projected model
	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/changesets/"+csID+"/projected", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	// Should return a view with view_type
	assert.Contains(t, resp, "view_type")
	assert.Equal(t, "need", resp["view_type"])
}

func TestProjectedModel_ChangesetNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/changesets/nonexistent/projected", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GET /api/models/{id}/changesets/{csId}/impact ───────────────────────────────

func TestImpact_Returns200WithDeltas(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	body := `{
		"id": "cs-impact",
		"description": "impact test",
		"actions": [
			{
				"type": "move_service",
				"service_name": "browse-svc",
				"from_team_name": "Team Alpha",
				"to_team_name": "Team Beta"
			}
		]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	createW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp map[string]any
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	csID := createResp["id"].(string)

	// Get impact
	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/changesets/"+csID+"/impact", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "cs-impact", resp["changeset_id"])
	deltas, ok := resp["deltas"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, deltas)

	// Each delta should have the expected fields
	firstDelta, ok := deltas[0].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, firstDelta, "dimension")
	assert.Contains(t, firstDelta, "before")
	assert.Contains(t, firstDelta, "after")
	assert.Contains(t, firstDelta, "change")
}

func TestImpact_ChangesetNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+modelID+"/changesets/nonexistent/impact", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── POST /api/models/{id}/changesets/{csId}/apply ───────────────────────────────

func TestApply_Returns200(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	body := `{
		"id": "cs-apply",
		"description": "apply test",
		"actions": [
			{
				"type": "move_service",
				"service_name": "browse-svc",
				"from_team_name": "Team Alpha",
				"to_team_name": "Team Beta"
			}
		]
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets", strings.NewReader(body))
	createW := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResp map[string]any
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))
	csID := createResp["id"].(string)

	// Apply
	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets/"+csID+"/apply", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Changeset Test System", resp["system_name"])
	assert.Contains(t, resp, "summary")
}

func TestApply_ChangesetNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)
	modelID := parseAndStoreModel(t, h, modelWithTwoTeamsYAML)

	req := httptest.NewRequest(http.MethodPost, "/api/models/"+modelID+"/changesets/nonexistent/apply", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestApply_ModelNotFound_Returns404(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/models/nonexistent/changesets/any/apply", nil)
	w := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
