package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// setupAuthzTestOrg creates a MemOrgStore with dev user and "local"/"default" workspace.
func setupAuthzTestOrg(t *testing.T) *repository.MemOrgStore {
	t.Helper()
	s := repository.NewMemOrgStore()
	_, _, _, err := s.EnsureDevUser(context.Background())
	require.NoError(t, err)
	return s
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		userRole string
		required string
		want     bool
	}{
		{"admin", "viewer", true},
		{"admin", "editor", true},
		{"admin", "admin", true},
		{"editor", "viewer", true},
		{"editor", "editor", true},
		{"editor", "admin", false},
		{"viewer", "viewer", true},
		{"viewer", "editor", false},
		{"viewer", "admin", false},
		{"", "viewer", false},
	}
	for _, tc := range tests {
		t.Run(tc.userRole+"/"+tc.required, func(t *testing.T) {
			assert.Equal(t, tc.want, hasRole(tc.userRole, tc.required))
		})
	}
}

func TestWorkspaceFromContext_NilWhenAbsent(t *testing.T) {
	ctx := context.Background()
	ws := WorkspaceFromContext(ctx)
	assert.Nil(t, ws)
}

func TestWorkspaceFromContext_ReturnsInjectedWorkspace(t *testing.T) {
	ws := &usecase.WorkspaceInfo{ID: "ws-1", Slug: "alpha", Role: "editor"}
	ctx := context.WithValue(context.Background(), workspaceKey, ws)
	got := WorkspaceFromContext(ctx)
	require.NotNil(t, got)
	assert.Equal(t, "alpha", got.Slug)
}

func TestMakeWorkspaceMiddleware_PassThrough_WhenNoPathValues(t *testing.T) {
	s := setupAuthzTestOrg(t)

	mw := makeWorkspaceMiddleware(s)
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(inner)

	// Request with no orgSlug/wsSlug path values — middleware should pass through.
	req := httptest.NewRequest("GET", "/api/orgs", nil)
	devUser := &usecase.AuthUser{ID: "00000000-0000-0000-0000-000000000001"}
	req = req.WithContext(context.WithValue(req.Context(), authUserKey, devUser))

	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	assert.True(t, called)
}

func TestMakeWorkspaceMiddleware_NotFound_UnknownWorkspace(t *testing.T) {
	s := setupAuthzTestOrg(t)

	mw := makeWorkspaceMiddleware(s)
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(inner)

	req := httptest.NewRequest("GET", "/api/orgs/local/ws/nope/models", nil)
	req.SetPathValue("orgSlug", "local")
	req.SetPathValue("wsSlug", "nope")
	devUser := &usecase.AuthUser{ID: "00000000-0000-0000-0000-000000000001"}
	req = req.WithContext(context.WithValue(req.Context(), authUserKey, devUser))

	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, called)
}

func TestMakeWorkspaceMiddleware_Injects_KnownWorkspace(t *testing.T) {
	s := setupAuthzTestOrg(t)

	mw := makeWorkspaceMiddleware(s)
	var injected *usecase.WorkspaceInfo
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		injected = WorkspaceFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(inner)

	req := httptest.NewRequest("GET", "/api/orgs/local/ws/default/models", nil)
	req.SetPathValue("orgSlug", "local")
	req.SetPathValue("wsSlug", "default")
	devUser := &usecase.AuthUser{ID: "00000000-0000-0000-0000-000000000001"}
	req = req.WithContext(context.WithValue(req.Context(), authUserKey, devUser))

	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, injected)
	assert.Equal(t, "default", injected.Slug)
}

func TestMakeAuthzMiddleware_Forbidden_InsufficientRole(t *testing.T) {
	s := setupAuthzTestOrg(t)
	authzMw := makeAuthzMiddleware(s, "editor")
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Inject workspace with viewer role into context.
	ws := &usecase.WorkspaceInfo{ID: "ws-1", Role: "viewer"}
	devUser := &usecase.AuthUser{ID: "00000000-0000-0000-0000-000000000001"}
	ctx := context.WithValue(context.Background(), workspaceKey, ws)
	ctx = context.WithValue(ctx, authUserKey, devUser)

	req := httptest.NewRequest("POST", "/api/orgs/local/ws/default/models/parse", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	authzMw(inner).ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, called)
}

func TestMakeAuthzMiddleware_Allowed_SufficientRole(t *testing.T) {
	s := setupAuthzTestOrg(t)
	authzMw := makeAuthzMiddleware(s, "editor")
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Inject workspace with admin role.
	ws := &usecase.WorkspaceInfo{ID: "ws-1", Role: "admin"}
	devUser := &usecase.AuthUser{ID: "00000000-0000-0000-0000-000000000001"}
	ctx := context.WithValue(context.Background(), workspaceKey, ws)
	ctx = context.WithValue(ctx, authUserKey, devUser)

	req := httptest.NewRequest("POST", "/api/orgs/local/ws/default/models/parse", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	authzMw(inner).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}

func TestOrgHandler_ListOrgs_DevUser(t *testing.T) {
	s := setupAuthzTestOrg(t)
	cfg := entity.DefaultConfig()

	h := New(HandlerDeps{
		Config:   cfg,
		OrgStore: s,
		Store:    repository.NewModelStore(),
	})

	srv := httptest.NewServer(NewRouter(h, cfg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/orgs")
	require.NoError(t, err)
	defer resp.Body.Close()
	// Dev user exists with "local" org — should return 200.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOrgHandler_GetOrg_NotFound(t *testing.T) {
	s := setupAuthzTestOrg(t)
	cfg := entity.DefaultConfig()

	h := New(HandlerDeps{
		Config:   cfg,
		OrgStore: s,
		Store:    repository.NewModelStore(),
	})

	srv := httptest.NewServer(NewRouter(h, cfg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/orgs/no-such-org")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOrgHandler_GetOrg_Found(t *testing.T) {
	s := setupAuthzTestOrg(t)
	cfg := entity.DefaultConfig()

	h := New(HandlerDeps{
		Config:   cfg,
		OrgStore: s,
		Store:    repository.NewModelStore(),
	})

	srv := httptest.NewServer(NewRouter(h, cfg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/orgs/local")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
