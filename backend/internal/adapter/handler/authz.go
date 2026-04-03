package handler

import (
	"context"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// workspaceKey is the context key for the resolved WorkspaceInfo.
const workspaceKey contextKey = "workspace"

// WorkspaceFromContext extracts the workspace injected by workspaceMiddleware.
// Returns nil if no workspace is present.
func WorkspaceFromContext(ctx context.Context) *usecase.WorkspaceInfo {
	ws, _ := ctx.Value(workspaceKey).(*usecase.WorkspaceInfo)
	return ws
}

// roleRank maps workspace roles to numeric ranks for comparison.
var roleRank = map[string]int{
	"viewer": 1,
	"editor": 2,
	"admin":  3,
}

// hasRole returns true if the user's role meets or exceeds the required role.
func hasRole(userRole, required string) bool {
	return roleRank[userRole] >= roleRank[required]
}

// makeWorkspaceMiddleware returns middleware that:
//  1. Extracts orgSlug and wsSlug from the path values.
//  2. Looks up the workspace (including org-visible access).
//  3. Injects the resolved WorkspaceInfo into the request context.
//  4. Returns 404 if the workspace is not found/not accessible.
//
// This middleware is applied to workspace-scoped routes only.
func makeWorkspaceMiddleware(orgStore usecase.OrgRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgSlug := r.PathValue("orgSlug")
			wsSlug := r.PathValue("wsSlug")
			if orgSlug == "" || wsSlug == "" {
				// Not a workspace-scoped route — pass through.
				next.ServeHTTP(w, r)
				return
			}

			user := AuthUserFromContext(r.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}

			ws, err := orgStore.GetWorkspace(r.Context(), user.ID, orgSlug, wsSlug)
			if err != nil {
				writeError(w, http.StatusNotFound, "workspace not found or not accessible")
				return
			}

			ctx := context.WithValue(r.Context(), workspaceKey, ws)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// makeAuthzMiddleware returns middleware that enforces a minimum workspace role.
// It reads the workspace from context (injected by makeWorkspaceMiddleware) and
// the user from context (injected by auth middleware).
//
// Permission matrix:
//   - viewer: read models, views, analysis, history
//   - editor: viewer + create/edit models, create/commit changesets
//   - admin:  editor + delete models, manage workspace members
//
// requiredRole must be one of: "viewer", "editor", "admin".
func makeAuthzMiddleware(orgStore usecase.OrgRepository, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := AuthUserFromContext(r.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}

			ws := WorkspaceFromContext(r.Context())
			if ws == nil {
				// No workspace in context — this middleware is only meaningful for
				// workspace-scoped routes. Fail closed.
				writeError(w, http.StatusForbidden, "workspace context required")
				return
			}

			if !hasRole(ws.Role, requiredRole) {
				writeError(w, http.StatusForbidden, "insufficient workspace role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
