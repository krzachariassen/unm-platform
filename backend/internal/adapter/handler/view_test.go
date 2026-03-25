package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	domainservice "github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// enrichedTestYAML is a richer model for testing enriched views.
const enrichedTestYAML = `system:
  name: "Enriched Test"
actors:
  - name: "Merchant"
  - name: "Eater"
needs:
  - name: "Upload catalog"
    actor: "Merchant"
    outcome: "Catalog is available"
    supportedBy:
      - "Catalog Ingestion"
  - name: "Browse menu"
    actor: "Eater"
    supportedBy:
      - "Menu Display"
  - name: "Track order"
    actor: "Eater"
capabilities:
  - name: "Catalog Ingestion"
    visibility: "domain"
    realizedBy:
      - "feed-api"
      - "catalog-worker"
  - name: "Menu Display"
    visibility: "user-facing"
    realizedBy:
      - "menu-service"
  - name: "Auth"
    visibility: "infrastructure"
  - name: "Search"
    visibility: "domain"
    realizedBy:
      - "search-service"
    dependsOn:
      - "Auth"
services:
  - name: "feed-api"
    ownedBy: "Catalog Team"
  - name: "catalog-worker"
    ownedBy: "Catalog Team"
  - name: "menu-service"
    ownedBy: "Menu Team"
  - name: "search-service"
    ownedBy: "Search Team"
  - name: "orphan-svc"
teams:
  - name: "Catalog Team"
    type: "stream-aligned"
    owns:
      - "Catalog Ingestion"
  - name: "Menu Team"
    type: "stream-aligned"
    owns:
      - "Menu Display"
  - name: "Search Team"
    type: "complicated-subsystem"
    owns:
      - "Search"
interactions:
  - from: "Catalog Team"
    to: "Search Team"
    mode: "x-as-a-service"
`

// setupEnrichedTestModel parses the enriched YAML and stores it.
func setupEnrichedTestModel(t *testing.T) (http.Handler, string) {
	t.Helper()
	store := repository.NewModelStore()
	cfg := entity.DefaultConfig()
	h := New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), domainservice.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		repository.NewChangesetStore(),
		analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		nil, // aiClient
		store,
	)
	router := NewRouter(h)

	p := parser.NewYAMLParser()
	v := domainservice.NewValidationEngine()
	uc := usecase.NewParseAndValidate(p, v)
	model, _, err := uc.Execute(strings.NewReader(enrichedTestYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	id, err := store.Store(model)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	return router, id
}

// ── GET /api/models/{id}/views/{viewType} — basic ─────────────────────────────

func TestHandleView_NeedView_Returns200WithEnrichedStructure(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/need", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "need", resp["view_type"])
	assert.EqualValues(t, 1, resp["total_needs"])
	assert.EqualValues(t, 0, resp["unmapped_count"])
	groups := resp["groups"].([]any)
	require.Len(t, groups, 1)
}

func TestHandleView_CapabilityView_Returns200(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/capability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "capability", resp["view_type"])
	_, hasCaps := resp["capabilities"]
	assert.True(t, hasCaps, "response must contain 'capabilities'")
}

func TestHandleView_RealizationView_Returns200(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/realization", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "realization", resp["view_type"])
	_, hasRows := resp["service_rows"]
	assert.True(t, hasRows, "response must contain 'service_rows'")
}

func TestHandleView_OwnershipView_Returns200(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ownership", resp["view_type"])
	_, hasLanes := resp["lanes"]
	assert.True(t, hasLanes, "response must contain 'lanes'")
}

func TestHandleView_TeamTopologyView_Returns200(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/team-topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "team-topology", resp["view_type"])
	_, hasTeams := resp["teams"]
	assert.True(t, hasTeams, "response must contain 'teams'")
}

func TestHandleView_CognitiveLoadView_Returns200(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/cognitive-load", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "cognitive-load", resp["view_type"])
	_, hasLoads := resp["team_loads"]
	assert.True(t, hasLoads, "response must contain 'team_loads'")
}

func TestHandleView_UNMMapView_Returns200(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/unm-map", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "unm-map", resp["view_type"])
	_, hasNodes := resp["nodes"]
	assert.True(t, hasNodes, "response must contain 'nodes'")
	_, hasEdges := resp["edges"]
	assert.True(t, hasEdges, "response must contain 'edges'")
}

func TestHandleView_UnknownID_Returns404(t *testing.T) {
	router, _ := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/nonexistent/views/need", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasError := resp["error"]
	assert.True(t, hasError)
}

func TestHandleView_UnknownViewType_Returns400(t *testing.T) {
	router, id := setupTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/bogus", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errMsg, ok := resp["error"].(string)
	require.True(t, ok)
	assert.Contains(t, errMsg, "bogus")
}

// ── Enriched Need View content tests ──────────────────────────────────────────

func TestHandleView_EnrichedNeedView_GroupsByActor(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/need", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedNeedResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "need", resp.ViewType)
	assert.Equal(t, 3, resp.TotalNeeds)
	assert.Equal(t, 1, resp.UnmappedCount) // "Track order" is unmapped
	assert.Len(t, resp.Groups, 2)          // Eater, Merchant

	// Groups are sorted by actor name
	assert.Equal(t, "Eater", resp.Groups[0].Actor.Label)
	assert.Equal(t, "Merchant", resp.Groups[1].Actor.Label)

	// Eater has 2 needs
	assert.Len(t, resp.Groups[0].Needs, 2)

	// Check "Track order" is unmapped
	var trackOrder *needEntry
	for i, ne := range resp.Groups[0].Needs {
		if ne.Need.Label == "Track order" {
			trackOrder = &resp.Groups[0].Needs[i]
			break
		}
	}
	require.NotNil(t, trackOrder)
	assert.Equal(t, false, trackOrder.Need.Data["is_mapped"])
	aps, _ := trackOrder.Need.Data["anti_patterns"].([]any)
	require.Len(t, aps, 1)
}

func TestHandleView_EnrichedNeedView_MappedNeedHasCapabilities(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/need", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp enrichedNeedResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Find "Upload catalog" under Merchant
	var upload *needEntry
	for _, g := range resp.Groups {
		if g.Actor.Label == "Merchant" {
			for i, ne := range g.Needs {
				if ne.Need.Label == "Upload catalog" {
					upload = &g.Needs[i]
				}
			}
		}
	}
	require.NotNil(t, upload)
	assert.True(t, upload.Need.Data["is_mapped"].(bool))
	assert.Len(t, upload.Capabilities, 1)
	assert.Equal(t, "Catalog Ingestion", upload.Capabilities[0].Label)
}

// ── Enriched Capability View content tests ────────────────────────────────────

func TestHandleView_EnrichedCapabilityView_Structure(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/capability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedCapabilityResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "capability", resp.ViewType)
	assert.Equal(t, 4, resp.LeafCapabilityCount) // all 4 are leaves
	assert.Len(t, resp.Capabilities, 4)

	// Find Catalog Ingestion
	var catCap *enrichedCapability
	for i, c := range resp.Capabilities {
		if c.Label == "Catalog Ingestion" {
			catCap = &resp.Capabilities[i]
			break
		}
	}
	require.NotNil(t, catCap)
	assert.True(t, catCap.IsLeaf)
	assert.Len(t, catCap.Services, 2) // feed-api, catalog-worker
	assert.Len(t, catCap.Teams, 1)    // Catalog Team
	assert.False(t, catCap.IsFragmented)
}

func TestHandleView_EnrichedCapabilityView_AuthHasNoServices(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/capability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp enrichedCapabilityResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	var authCap *enrichedCapability
	for i, c := range resp.Capabilities {
		if c.Label == "Auth" {
			authCap = &resp.Capabilities[i]
			break
		}
	}
	require.NotNil(t, authCap)
	assert.Empty(t, authCap.Services)
	// Auth is a leaf with no services → anti-pattern
	require.Len(t, authCap.AntiPatterns, 1)
	assert.Equal(t, "no_services", authCap.AntiPatterns[0].Code)
}

func TestHandleView_EnrichedCapabilityView_SearchDependsOnAuth(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/capability", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp enrichedCapabilityResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	var searchCap *enrichedCapability
	for i, c := range resp.Capabilities {
		if c.Label == "Search" {
			searchCap = &resp.Capabilities[i]
			break
		}
	}
	require.NotNil(t, searchCap)
	require.Len(t, searchCap.DependsOn, 1)
	assert.Equal(t, "Auth", searchCap.DependsOn[0].Label)
}

// ── Enriched Ownership View content tests ─────────────────────────────────────

func TestHandleView_EnrichedOwnershipView_Lanes(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedOwnershipResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "ownership", resp.ViewType)
	assert.Len(t, resp.Lanes, 3) // Catalog, Menu, Search teams

	// Unowned caps: Auth has no owning team
	assert.Len(t, resp.UnownedCapabilities, 1)
	assert.Equal(t, "Auth", resp.UnownedCapabilities[0].Label)

	// no_cap_count: orphan-svc has no capabilities
	assert.Equal(t, 1, resp.NoCapCount)
}

func TestHandleView_EnrichedOwnershipView_ServiceRows(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp enrichedOwnershipResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Should have 5 service rows
	assert.Len(t, resp.ServiceRows, 5)

	// orphan-svc should be last (no team)
	lastRow := resp.ServiceRows[len(resp.ServiceRows)-1]
	assert.Equal(t, "orphan-svc", lastRow.Service.Label)
	assert.Nil(t, lastRow.Team)
}

// ── Enriched Team Topology View content tests ─────────────────────────────────

func TestHandleView_EnrichedTeamTopologyView_Structure(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/team-topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedTeamTopologyResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "team-topology", resp.ViewType)
	assert.Len(t, resp.Teams, 3)
	assert.Len(t, resp.Interactions, 1) // Catalog→Search

	// Check Catalog Team
	var catTeam *enrichedTeamNode
	for i, tm := range resp.Teams {
		if tm.Label == "Catalog Team" {
			catTeam = &resp.Teams[i]
			break
		}
	}
	require.NotNil(t, catTeam)
	assert.Equal(t, "stream-aligned", catTeam.Type)
	assert.Equal(t, 1, catTeam.CapabilityCount)
	assert.Equal(t, 2, catTeam.ServiceCount) // feed-api, catalog-worker
	require.Len(t, catTeam.Interactions, 1)
	assert.Equal(t, "x-as-a-service", catTeam.Interactions[0].Mode)
}

// ── Enriched Cognitive Load View content tests ────────────────────────────────

func TestHandleView_EnrichedCognitiveLoadView_Structure(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/cognitive-load", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedCognitiveLoadResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "cognitive-load", resp.ViewType)
	assert.Len(t, resp.TeamLoads, 3)

	// All test fixture teams have small loads → overall level should be "low"
	for _, tl := range resp.TeamLoads {
		assert.Contains(t, []string{"low", "medium", "high"}, tl.OverallLevel)
		assert.NotEmpty(t, tl.DomainSpread.Level)
		assert.NotEmpty(t, tl.ServiceLoad.Level)
		assert.NotEmpty(t, tl.InteractionLoad.Level)
		assert.NotEmpty(t, tl.DependencyLoad.Level)
	}
}

// ── Enriched Realization View content tests ───────────────────────────────────

func TestHandleView_EnrichedRealizationView_Structure(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/realization", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedRealizationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "realization", resp.ViewType)
	assert.Len(t, resp.ServiceRows, 5)
	assert.Equal(t, 1, resp.NoCapCount) // orphan-svc
}

// ── Phase 4.9: External Deps in Ownership and Realization Views ───────────────

// externalDepsTestYAML is a model that includes external_dependencies for testing 4.9.
const externalDepsTestYAML = `system:
  name: "ExtDep Test"
actors:
  - name: "Merchant"
needs:
  - name: "Upload catalog"
    actor: "Merchant"
    supportedBy:
      - "Catalog Ingestion"
capabilities:
  - name: "Catalog Ingestion"
    visibility: "domain"
    realizedBy:
      - "feed-api"
      - "catalog-worker"
services:
  - name: "feed-api"
    ownedBy: "Catalog Team"
  - name: "catalog-worker"
    ownedBy: "Catalog Team"
  - name: "other-svc"
    ownedBy: "Other Team"
teams:
  - name: "Catalog Team"
    type: "stream-aligned"
    owns:
      - "Catalog Ingestion"
  - name: "Other Team"
    type: "platform"
external_dependencies:
  - name: "Stripe"
    description: "Payment gateway"
    usedBy:
      - target: "feed-api"
        description: "Charges merchants"
      - target: "other-svc"
        description: "Also uses Stripe"
  - name: "SendGrid"
    description: "Email provider"
    usedBy:
      - target: "catalog-worker"
        description: "Sends notifications"
`

// setupExtDepTestModel parses the external deps YAML and returns a router + stored model ID.
func setupExtDepTestModel(t *testing.T) (http.Handler, string) {
	t.Helper()
	store := repository.NewModelStore()
	cfg := entity.DefaultConfig()
	h := New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), domainservice.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		repository.NewChangesetStore(),
		analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		nil, // aiClient
		store,
	)
	router := NewRouter(h)

	p := parser.NewYAMLParser()
	v := domainservice.NewValidationEngine()
	uc := usecase.NewParseAndValidate(p, v)
	model, _, err := uc.Execute(strings.NewReader(externalDepsTestYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	id, err := store.Store(model)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	return router, id
}

func TestHandleView_OwnershipView_ExternalDependencyCount(t *testing.T) {
	router, id := setupExtDepTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedOwnershipResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Top-level count: 2 external deps (Stripe, SendGrid)
	assert.Equal(t, 2, resp.ExternalDependencyCount)
}

func TestHandleView_OwnershipView_LaneExternalDeps(t *testing.T) {
	router, id := setupExtDepTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedOwnershipResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Find Catalog Team lane
	var catalogLane *ownershipLane
	for i, lane := range resp.Lanes {
		if lane.Team.Label == "Catalog Team" {
			catalogLane = &resp.Lanes[i]
			break
		}
	}
	require.NotNil(t, catalogLane, "Catalog Team lane must exist")

	// Catalog Team owns feed-api (uses Stripe) and catalog-worker (uses SendGrid)
	// So both Stripe and SendGrid should appear in the lane's external_deps
	require.Len(t, catalogLane.ExternalDeps, 2)

	// Check that service_count is correct
	depsByID := make(map[string]extDepRef)
	for _, d := range catalogLane.ExternalDeps {
		depsByID[d.Label] = d
	}
	stripe, ok := depsByID["Stripe"]
	require.True(t, ok, "Stripe must appear in Catalog Team lane")
	assert.Equal(t, 1, stripe.ServiceCount) // only feed-api in this team uses Stripe

	sendgrid, ok := depsByID["SendGrid"]
	require.True(t, ok, "SendGrid must appear in Catalog Team lane")
	assert.Equal(t, 1, sendgrid.ServiceCount) // only catalog-worker in this team uses SendGrid
}

func TestHandleView_OwnershipView_OtherTeamLane_OnlyStripe(t *testing.T) {
	router, id := setupExtDepTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedOwnershipResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Find Other Team lane
	var otherLane *ownershipLane
	for i, lane := range resp.Lanes {
		if lane.Team.Label == "Other Team" {
			otherLane = &resp.Lanes[i]
			break
		}
	}
	require.NotNil(t, otherLane, "Other Team lane must exist")

	// Other Team owns other-svc which uses Stripe but not SendGrid
	require.Len(t, otherLane.ExternalDeps, 1)
	assert.Equal(t, "Stripe", otherLane.ExternalDeps[0].Label)
	assert.Equal(t, 1, otherLane.ExternalDeps[0].ServiceCount)
}

func TestHandleView_OwnershipView_LaneExternalDeps_EmptyWhenNone(t *testing.T) {
	router, id := setupEnrichedTestModel(t) // this model has no external_dependencies

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/ownership", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedOwnershipResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, 0, resp.ExternalDependencyCount)
	for _, lane := range resp.Lanes {
		assert.Empty(t, lane.ExternalDeps, "lane %s must have empty external_deps", lane.Team.Label)
	}
}

func TestHandleView_RealizationView_ServiceRowExternalDeps(t *testing.T) {
	router, id := setupExtDepTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/realization", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedRealizationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	rowsByLabel := make(map[string]serviceRow)
	for _, r := range resp.ServiceRows {
		rowsByLabel[r.Service.Label] = r
	}

	// feed-api uses Stripe
	feedAPI, ok := rowsByLabel["feed-api"]
	require.True(t, ok)
	assert.Equal(t, []string{"Stripe"}, feedAPI.ExternalDeps)

	// catalog-worker uses SendGrid
	catalogWorker, ok := rowsByLabel["catalog-worker"]
	require.True(t, ok)
	assert.Equal(t, []string{"SendGrid"}, catalogWorker.ExternalDeps)

	// other-svc uses Stripe
	otherSvc, ok := rowsByLabel["other-svc"]
	require.True(t, ok)
	assert.Equal(t, []string{"Stripe"}, otherSvc.ExternalDeps)
}

func TestHandleView_RealizationView_ServiceRowExternalDeps_EmptyWhenNone(t *testing.T) {
	router, id := setupEnrichedTestModel(t) // no external_dependencies

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/realization", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp enrichedRealizationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	for _, row := range resp.ServiceRows {
		assert.Empty(t, row.ExternalDeps, "service %s must have empty external_deps", row.Service.Label)
	}
}

// ── UNM Map View content tests ────────────────────────────────────────────────

func TestHandleView_UNMMapView_ContainsAllNodeTypes(t *testing.T) {
	router, id := setupEnrichedTestModel(t)

	req := httptest.NewRequest(http.MethodGet, "/api/models/"+id+"/views/unm-map", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp unmMapResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "unm-map", resp.ViewType)

	nodeTypes := map[string]int{}
	for _, n := range resp.Nodes {
		nodeTypes[n.Type]++
	}
	assert.Equal(t, 2, nodeTypes["actor"])
	assert.Equal(t, 3, nodeTypes["need"])
	assert.Equal(t, 4, nodeTypes["capability"])
	// Services and teams are embedded in capability node data, not emitted as separate nodes.
	assert.Equal(t, 0, nodeTypes["service"])
	assert.Equal(t, 0, nodeTypes["team"])

	// Should have edges: has_need + supportedBy + capability dependsOn edges
	assert.GreaterOrEqual(t, len(resp.Edges), 5)
}
