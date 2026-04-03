package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// orgHandler holds dependencies for org and workspace HTTP handlers.
type orgHandler struct {
	store usecase.OrgRepository
}

// ---- Request/Response types ------------------------------------------------

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type orgResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role,omitempty"`
}

type workspaceResponse struct {
	ID         string `json:"id"`
	OrgID      string `json:"org_id"`
	OrgSlug    string `json:"org_slug"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Visibility string `json:"visibility"`
	Role       string `json:"role,omitempty"`
}

type memberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Remove bool   `json:"remove,omitempty"`
}

type createWorkspaceRequest struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Visibility string `json:"visibility"` // "private" | "org-visible"
}

// ---- Helpers ---------------------------------------------------------------

func orgInfoToResponse(o *usecase.OrgInfo) orgResponse {
	return orgResponse{ID: o.ID, Name: o.Name, Slug: o.Slug, Role: o.Role}
}

func wsInfoToResponse(w *usecase.WorkspaceInfo) workspaceResponse {
	return workspaceResponse{
		ID:         w.ID,
		OrgID:      w.OrgID,
		OrgSlug:    w.OrgSlug,
		Name:       w.Name,
		Slug:       w.Slug,
		Visibility: w.Visibility,
		Role:       w.Role,
	}
}

// requireAuthUser extracts the authenticated user from the request context.
// If no user is present it writes a 401 and returns nil.
func requireAuthUser(w http.ResponseWriter, r *http.Request) *usecase.AuthUser {
	user := AuthUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return nil
	}
	return user
}

// ---- Org handlers ----------------------------------------------------------

// handleListOrgs returns all orgs the authenticated user belongs to.
// GET /api/orgs
func (h *Handler) handleListOrgs(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}

	orgs, err := h.orgH.store.ListOrgs(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list orgs")
		return
	}

	resp := make([]orgResponse, 0, len(orgs))
	for _, o := range orgs {
		resp = append(resp, orgInfoToResponse(o))
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleCreateOrg creates a new org owned by the authenticated user.
// POST /api/orgs
func (h *Handler) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}

	var req createOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	org, err := h.orgH.store.CreateOrg(r.Context(), user.ID, req.Name, req.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create org")
		return
	}
	writeJSON(w, http.StatusCreated, orgInfoToResponse(org))
}

// handleGetOrg returns an org's details.
// GET /api/orgs/{orgSlug}
func (h *Handler) handleGetOrg(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}
	orgSlug := r.PathValue("orgSlug")

	org, err := h.orgH.store.GetOrg(r.Context(), user.ID, orgSlug)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			writeError(w, http.StatusNotFound, "org not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get org")
		return
	}

	// Also include workspaces.
	wss, _ := h.orgH.store.ListWorkspaces(r.Context(), user.ID, orgSlug)
	wsResp := make([]workspaceResponse, 0, len(wss))
	for _, ws := range wss {
		wsResp = append(wsResp, wsInfoToResponse(ws))
	}

	type orgDetailResponse struct {
		orgResponse
		Workspaces []workspaceResponse `json:"workspaces"`
	}
	writeJSON(w, http.StatusOK, orgDetailResponse{orgResponse: orgInfoToResponse(org), Workspaces: wsResp})
}

// handleUpdateOrgMembers adds or removes an org member (admin only).
// PUT /api/orgs/{orgSlug}/members
func (h *Handler) handleUpdateOrgMembers(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}
	orgSlug := r.PathValue("orgSlug")

	// Check requesting user is owner/admin.
	org, err := h.orgH.store.GetOrg(r.Context(), user.ID, orgSlug)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			writeError(w, http.StatusNotFound, "org not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to check org membership")
		return
	}
	if org.Role != "owner" && org.Role != "admin" {
		writeError(w, http.StatusForbidden, "only org admins can manage members")
		return
	}

	var req memberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if req.Remove {
		if err := h.orgH.store.RemoveOrgMember(r.Context(), orgSlug, req.UserID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to remove member")
			return
		}
	} else {
		if req.Role == "" {
			req.Role = "member"
		}
		if err := h.orgH.store.AddOrgMember(r.Context(), orgSlug, req.UserID, req.Role); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to add member")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Workspace handlers ----------------------------------------------------

// handleListWorkspaces lists workspaces the user can access within an org.
// GET /api/orgs/{orgSlug}/workspaces
func (h *Handler) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}
	orgSlug := r.PathValue("orgSlug")

	wss, err := h.orgH.store.ListWorkspaces(r.Context(), user.ID, orgSlug)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			writeError(w, http.StatusNotFound, "org not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to list workspaces")
		return
	}

	resp := make([]workspaceResponse, 0, len(wss))
	for _, ws := range wss {
		resp = append(resp, wsInfoToResponse(ws))
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleCreateWorkspace creates a workspace inside an org.
// POST /api/orgs/{orgSlug}/workspaces
func (h *Handler) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}
	orgSlug := r.PathValue("orgSlug")

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if req.Visibility == "" {
		req.Visibility = "private"
	}

	ws, err := h.orgH.store.CreateWorkspace(r.Context(), user.ID, orgSlug, req.Name, req.Slug, req.Visibility)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			writeError(w, http.StatusNotFound, "org not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create workspace")
		return
	}
	writeJSON(w, http.StatusCreated, wsInfoToResponse(ws))
}

// handleGetWorkspace returns a workspace's details.
// GET /api/orgs/{orgSlug}/ws/{wsSlug}
func (h *Handler) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}
	orgSlug := r.PathValue("orgSlug")
	wsSlug := r.PathValue("wsSlug")

	ws, err := h.orgH.store.GetWorkspace(r.Context(), user.ID, orgSlug, wsSlug)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get workspace")
		return
	}
	writeJSON(w, http.StatusOK, wsInfoToResponse(ws))
}

// handleUpdateWorkspaceMembers adds or removes a workspace member (ws admin only).
// PUT /api/orgs/{orgSlug}/ws/{wsSlug}/members
func (h *Handler) handleUpdateWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	user := requireAuthUser(w, r)
	if user == nil {
		return
	}
	orgSlug := r.PathValue("orgSlug")
	wsSlug := r.PathValue("wsSlug")

	// Check requesting user is workspace admin.
	ws, err := h.orgH.store.GetWorkspace(r.Context(), user.ID, orgSlug, wsSlug)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to check workspace membership")
		return
	}
	if ws.Role != "admin" {
		writeError(w, http.StatusForbidden, "only workspace admins can manage members")
		return
	}

	var req memberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if req.Remove {
		if err := h.orgH.store.RemoveWorkspaceMember(r.Context(), orgSlug, wsSlug, req.UserID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to remove workspace member")
			return
		}
	} else {
		if req.Role == "" {
			req.Role = "viewer"
		}
		if err := h.orgH.store.AddWorkspaceMember(r.Context(), orgSlug, wsSlug, req.UserID, req.Role); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to add workspace member")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// registerOrgRoutes registers org and workspace management routes on the mux.
func (h *Handler) registerOrgRoutes(mux *http.ServeMux) {
	// Org routes.
	mux.HandleFunc("GET /api/orgs", h.handleListOrgs)
	mux.HandleFunc("POST /api/orgs", h.handleCreateOrg)
	mux.HandleFunc("GET /api/orgs/{orgSlug}", h.handleGetOrg)
	mux.HandleFunc("PUT /api/orgs/{orgSlug}/members", h.handleUpdateOrgMembers)

	// Workspace routes.
	mux.HandleFunc("GET /api/orgs/{orgSlug}/workspaces", h.handleListWorkspaces)
	mux.HandleFunc("POST /api/orgs/{orgSlug}/workspaces", h.handleCreateWorkspace)
	mux.HandleFunc("GET /api/orgs/{orgSlug}/ws/{wsSlug}", h.handleGetWorkspace)
	mux.HandleFunc("PUT /api/orgs/{orgSlug}/ws/{wsSlug}/members", h.handleUpdateWorkspaceMembers)

	// Workspace-scoped model routes (delegate to existing model handlers).
	mux.HandleFunc("GET /api/orgs/{orgSlug}/ws/{wsSlug}/models", h.handleListModels)
	mux.HandleFunc("POST /api/orgs/{orgSlug}/ws/{wsSlug}/models/parse", h.handleParse)
	mux.HandleFunc("POST /api/orgs/{orgSlug}/ws/{wsSlug}/models/validate", h.handleValidate)
	mux.HandleFunc("GET /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/export", h.handleExport)
	mux.HandleFunc("GET /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/history", h.handleListVersions)
	mux.HandleFunc("GET /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/versions/{v}", h.handleGetVersion)
	mux.HandleFunc("GET /api/orgs/{orgSlug}/ws/{wsSlug}/models/{id}/diff", h.handleDiffVersions)
}
