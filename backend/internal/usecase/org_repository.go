package usecase

import "context"

// OrgInfo holds an organization's data plus the requesting user's role.
type OrgInfo struct {
	ID   string
	Name string
	Slug string
	Role string // "owner" | "admin" | "member"
}

// WorkspaceInfo holds a workspace's data plus the requesting user's effective role.
type WorkspaceInfo struct {
	ID         string
	OrgID      string
	OrgSlug    string
	Name       string
	Slug       string
	Visibility string // "private" | "org-visible"
	Role       string // "admin" | "editor" | "viewer", or "" for org-visible read-only access
}

// OrgRepository is the persistence contract for organisations, workspaces, and
// user onboarding. Implementations: PGOrgStore (postgres) and MemOrgStore (in-memory stub).
type OrgRepository interface {
	// EnsureUser upserts a user row (called on login/OAuth callback).
	// Returns the user's UUID.
	EnsureUser(ctx context.Context, email, name, avatarURL string) (string, error)

	// EnsureDevUser upserts the hardcoded dev user (id 00000000-0000-0000-0000-000000000001)
	// and ensures a personal org (slug "local") and workspace (slug "default") exist.
	// Returns the user ID, org slug, and workspace slug.
	EnsureDevUser(ctx context.Context) (userID string, orgSlug string, wsSlug string, err error)

	// OnboardNewUser creates a personal org + "General" workspace for a first-time user.
	// The org slug is derived from displayName. Returns the created OrgInfo and WorkspaceInfo.
	OnboardNewUser(ctx context.Context, userID, displayName string) (*OrgInfo, *WorkspaceInfo, error)

	// CreateOrg creates an organisation owned by ownerID.
	// Also creates a "General" workspace and grants the owner admin access.
	CreateOrg(ctx context.Context, ownerID, name, slug string) (*OrgInfo, error)

	// GetOrg returns the org identified by orgSlug plus the requesting user's role.
	// Returns ErrNotFound if the org does not exist or the user is not a member.
	GetOrg(ctx context.Context, userID, orgSlug string) (*OrgInfo, error)

	// ListOrgs returns all orgs the user belongs to (via org_memberships).
	ListOrgs(ctx context.Context, userID string) ([]*OrgInfo, error)

	// CreateWorkspace creates a workspace in the org identified by orgSlug.
	// The workspace is owned by createdByUserID (granted admin role).
	CreateWorkspace(ctx context.Context, createdByUserID, orgSlug, name, slug, visibility string) (*WorkspaceInfo, error)

	// GetWorkspace returns workspace info plus the requesting user's effective role.
	// Org-visible workspaces are returned even when the user is not an explicit member.
	// Returns ErrNotFound if the workspace does not exist or is not accessible.
	GetWorkspace(ctx context.Context, userID, orgSlug, wsSlug string) (*WorkspaceInfo, error)

	// ListWorkspaces returns workspaces in an org accessible to the user.
	// Includes explicitly-joined workspaces and org-visible ones.
	ListWorkspaces(ctx context.Context, userID, orgSlug string) ([]*WorkspaceInfo, error)

	// AddOrgMember grants (or updates) a user's role in an org.
	AddOrgMember(ctx context.Context, orgSlug, targetUserID, role string) error

	// RemoveOrgMember removes a user from an org.
	RemoveOrgMember(ctx context.Context, orgSlug, targetUserID string) error

	// AddWorkspaceMember grants (or updates) a user's role in a workspace.
	AddWorkspaceMember(ctx context.Context, orgSlug, wsSlug, targetUserID, role string) error

	// RemoveWorkspaceMember removes a user from a workspace.
	RemoveWorkspaceMember(ctx context.Context, orgSlug, wsSlug, targetUserID string) error
}
