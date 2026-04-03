package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

const pgDevUserID = "00000000-0000-0000-0000-000000000001"

// PGOrgStore is the PostgreSQL-backed OrgRepository implementation.
type PGOrgStore struct {
	db *pgxpool.Pool
}

// NewPGOrgStore returns a PGOrgStore backed by the given pool.
func NewPGOrgStore(db *pgxpool.Pool) *PGOrgStore {
	return &PGOrgStore{db: db}
}

// EnsureUser upserts a user by email and returns the user UUID.
func (s *PGOrgStore) EnsureUser(ctx context.Context, email, name, avatarURL string) (string, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO users (email, name, avatar_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
		  SET name = EXCLUDED.name,
		      avatar_url = EXCLUDED.avatar_url,
		      last_login_at = NOW()
		RETURNING id
	`, email, name, avatarURL).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("pg_org_store: EnsureUser: %w", err)
	}
	return id, nil
}

// EnsureDevUser upserts the hardcoded dev user + "local" org + "default" workspace.
func (s *PGOrgStore) EnsureDevUser(ctx context.Context) (string, string, string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Upsert user.
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, name, avatar_url)
		VALUES ($1, 'local@dev', 'Local Dev User', '')
		ON CONFLICT (id) DO UPDATE SET last_login_at = NOW()
	`, pgDevUserID)
	if err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser upsert user: %w", err)
	}

	orgSlug := "local"
	wsSlug := "default"

	// Upsert org.
	var orgID string
	err = tx.QueryRow(ctx, `
		INSERT INTO organizations (name, slug)
		VALUES ('Local Dev Org', $1)
		ON CONFLICT (slug) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, orgSlug).Scan(&orgID)
	if err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser upsert org: %w", err)
	}

	// Upsert org membership.
	_, err = tx.Exec(ctx, `
		INSERT INTO org_memberships (org_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT (org_id, user_id) DO NOTHING
	`, orgID, pgDevUserID)
	if err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser org membership: %w", err)
	}

	// Upsert workspace.
	var wsID string
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (org_id, name, slug, visibility, created_by)
		VALUES ($1, 'Default', $2, 'org-visible', $3)
		ON CONFLICT (org_id, slug) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, orgID, wsSlug, pgDevUserID).Scan(&wsID)
	if err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser upsert workspace: %w", err)
	}

	// Upsert workspace membership.
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_memberships (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`, wsID, pgDevUserID)
	if err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser ws membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", "", fmt.Errorf("pg_org_store: EnsureDevUser commit: %w", err)
	}
	return pgDevUserID, orgSlug, wsSlug, nil
}

// OnboardNewUser creates a personal org + General workspace for a first-time user.
func (s *PGOrgStore) OnboardNewUser(ctx context.Context, userID, displayName string) (*usecase.OrgInfo, *usecase.WorkspaceInfo, error) {
	orgName := displayName + "'s Org"
	orgSlug := pgSlugify(displayName) + "-org"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("pg_org_store: OnboardNewUser begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert org (retry slug on conflict).
	var orgID string
	base := orgSlug
	for i := 2; ; i++ {
		err = tx.QueryRow(ctx, `
			INSERT INTO organizations (name, slug)
			VALUES ($1, $2)
			ON CONFLICT (slug) DO NOTHING
			RETURNING id
		`, orgName, orgSlug).Scan(&orgID)
		if err == nil && orgID != "" {
			break
		}
		if i > 20 {
			return nil, nil, fmt.Errorf("pg_org_store: OnboardNewUser: could not find unique org slug")
		}
		orgSlug = fmt.Sprintf("%s-%d", base, i)
	}

	// Org membership.
	_, err = tx.Exec(ctx, `
		INSERT INTO org_memberships (org_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, orgID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("pg_org_store: OnboardNewUser org membership: %w", err)
	}

	// General workspace.
	var wsID string
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (org_id, name, slug, visibility, created_by)
		VALUES ($1, 'General', 'general', 'org-visible', $2)
		RETURNING id
	`, orgID, userID).Scan(&wsID)
	if err != nil {
		return nil, nil, fmt.Errorf("pg_org_store: OnboardNewUser workspace: %w", err)
	}

	// Workspace membership.
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_memberships (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, wsID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("pg_org_store: OnboardNewUser ws membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("pg_org_store: OnboardNewUser commit: %w", err)
	}

	orgInfo := &usecase.OrgInfo{ID: orgID, Name: orgName, Slug: orgSlug, Role: "owner"}
	wsInfo := &usecase.WorkspaceInfo{ID: wsID, OrgID: orgID, OrgSlug: orgSlug, Name: "General", Slug: "general", Visibility: "org-visible", Role: "admin"}
	return orgInfo, wsInfo, nil
}

// CreateOrg creates an org owned by ownerID plus a General workspace.
func (s *PGOrgStore) CreateOrg(ctx context.Context, ownerID, name, slug string) (*usecase.OrgInfo, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateOrg begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var orgID string
	err = tx.QueryRow(ctx, `
		INSERT INTO organizations (name, slug)
		VALUES ($1, $2)
		RETURNING id
	`, name, slug).Scan(&orgID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateOrg insert org: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO org_memberships (org_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, orgID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateOrg org membership: %w", err)
	}

	var wsID string
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (org_id, name, slug, visibility, created_by)
		VALUES ($1, 'General', 'general', 'org-visible', $2)
		RETURNING id
	`, orgID, ownerID).Scan(&wsID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateOrg workspace: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_memberships (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, wsID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateOrg ws membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateOrg commit: %w", err)
	}
	return &usecase.OrgInfo{ID: orgID, Name: name, Slug: slug, Role: "owner"}, nil
}

// GetOrg returns org info plus the user's role.
func (s *PGOrgStore) GetOrg(ctx context.Context, userID, orgSlug string) (*usecase.OrgInfo, error) {
	var info usecase.OrgInfo
	err := s.db.QueryRow(ctx, `
		SELECT o.id, o.name, o.slug, om.role
		FROM organizations o
		JOIN org_memberships om ON om.org_id = o.id
		WHERE o.slug = $1 AND om.user_id = $2 AND o.deleted_at IS NULL
	`, orgSlug, userID).Scan(&info.ID, &info.Name, &info.Slug, &info.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, usecase.ErrNotFound
		}
		return nil, fmt.Errorf("pg_org_store: GetOrg: %w", err)
	}
	return &info, nil
}

// ListOrgs returns all orgs the user belongs to.
func (s *PGOrgStore) ListOrgs(ctx context.Context, userID string) ([]*usecase.OrgInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT o.id, o.name, o.slug, om.role
		FROM organizations o
		JOIN org_memberships om ON om.org_id = o.id
		WHERE om.user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: ListOrgs: %w", err)
	}
	defer rows.Close()

	var result []*usecase.OrgInfo
	for rows.Next() {
		var info usecase.OrgInfo
		if err := rows.Scan(&info.ID, &info.Name, &info.Slug, &info.Role); err != nil {
			return nil, fmt.Errorf("pg_org_store: ListOrgs scan: %w", err)
		}
		result = append(result, &info)
	}
	return result, rows.Err()
}

// CreateWorkspace creates a workspace in the org identified by orgSlug.
func (s *PGOrgStore) CreateWorkspace(ctx context.Context, createdByUserID, orgSlug, name, slug, visibility string) (*usecase.WorkspaceInfo, error) {
	var orgID string
	err := s.db.QueryRow(ctx, `SELECT id FROM organizations WHERE slug = $1 AND deleted_at IS NULL`, orgSlug).Scan(&orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, usecase.ErrNotFound
		}
		return nil, fmt.Errorf("pg_org_store: CreateWorkspace lookup org: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateWorkspace begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var wsID string
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (org_id, name, slug, visibility, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, orgID, name, slug, visibility, createdByUserID).Scan(&wsID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateWorkspace insert: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_memberships (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, wsID, createdByUserID)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateWorkspace ws membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("pg_org_store: CreateWorkspace commit: %w", err)
	}
	return &usecase.WorkspaceInfo{ID: wsID, OrgID: orgID, OrgSlug: orgSlug, Name: name, Slug: slug, Visibility: visibility, Role: "admin"}, nil
}

// GetWorkspace returns workspace info plus the user's effective role.
// Org-visible workspaces are returned even when the user is not an explicit member.
func (s *PGOrgStore) GetWorkspace(ctx context.Context, userID, orgSlug, wsSlug string) (*usecase.WorkspaceInfo, error) {
	var info usecase.WorkspaceInfo
	var role *string // nullable — may be NULL for org-visible non-members

	err := s.db.QueryRow(ctx, `
		SELECT w.id, o.id, o.slug, w.name, w.slug, w.visibility,
		       wm.role
		FROM workspaces w
		JOIN organizations o ON o.id = w.org_id
		LEFT JOIN workspace_memberships wm ON wm.workspace_id = w.id AND wm.user_id = $1
		WHERE o.slug = $2 AND w.slug = $3
		  AND w.deleted_at IS NULL AND o.deleted_at IS NULL
		  AND (
		    wm.user_id IS NOT NULL
		    OR (
		      w.visibility = 'org-visible'
		      AND EXISTS (
		        SELECT 1 FROM org_memberships om
		        WHERE om.org_id = o.id AND om.user_id = $1
		      )
		    )
		  )
	`, userID, orgSlug, wsSlug).Scan(&info.ID, &info.OrgID, &info.OrgSlug, &info.Name, &info.Slug, &info.Visibility, &role)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, usecase.ErrNotFound
		}
		return nil, fmt.Errorf("pg_org_store: GetWorkspace: %w", err)
	}
	if role != nil {
		info.Role = *role
	}
	return &info, nil
}

// ListWorkspaces returns workspaces in an org accessible to the user.
func (s *PGOrgStore) ListWorkspaces(ctx context.Context, userID, orgSlug string) ([]*usecase.WorkspaceInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT w.id, o.id, o.slug, w.name, w.slug, w.visibility,
		       wm.role
		FROM workspaces w
		JOIN organizations o ON o.id = w.org_id
		LEFT JOIN workspace_memberships wm ON wm.workspace_id = w.id AND wm.user_id = $1
		WHERE o.slug = $2
		  AND w.deleted_at IS NULL AND o.deleted_at IS NULL
		  AND (
		    wm.user_id IS NOT NULL
		    OR (
		      w.visibility = 'org-visible'
		      AND EXISTS (
		        SELECT 1 FROM org_memberships om
		        WHERE om.org_id = o.id AND om.user_id = $1
		      )
		    )
		  )
		ORDER BY w.name
	`, userID, orgSlug)
	if err != nil {
		return nil, fmt.Errorf("pg_org_store: ListWorkspaces: %w", err)
	}
	defer rows.Close()

	var result []*usecase.WorkspaceInfo
	for rows.Next() {
		var info usecase.WorkspaceInfo
		var role *string
		if err := rows.Scan(&info.ID, &info.OrgID, &info.OrgSlug, &info.Name, &info.Slug, &info.Visibility, &role); err != nil {
			return nil, fmt.Errorf("pg_org_store: ListWorkspaces scan: %w", err)
		}
		if role != nil {
			info.Role = *role
		}
		result = append(result, &info)
	}
	return result, rows.Err()
}

// AddOrgMember grants or updates a user's role in an org.
func (s *PGOrgStore) AddOrgMember(ctx context.Context, orgSlug, targetUserID, role string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO org_memberships (org_id, user_id, role)
		SELECT id, $2, $3 FROM organizations WHERE slug = $1 AND deleted_at IS NULL
		ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, orgSlug, targetUserID, role)
	if err != nil {
		return fmt.Errorf("pg_org_store: AddOrgMember: %w", err)
	}
	return nil
}

// RemoveOrgMember removes a user from an org.
func (s *PGOrgStore) RemoveOrgMember(ctx context.Context, orgSlug, targetUserID string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM org_memberships
		WHERE org_id = (SELECT id FROM organizations WHERE slug = $1 AND deleted_at IS NULL)
		  AND user_id = $2
	`, orgSlug, targetUserID)
	if err != nil {
		return fmt.Errorf("pg_org_store: RemoveOrgMember: %w", err)
	}
	return nil
}

// AddWorkspaceMember grants or updates a user's role in a workspace.
func (s *PGOrgStore) AddWorkspaceMember(ctx context.Context, orgSlug, wsSlug, targetUserID, role string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO workspace_memberships (workspace_id, user_id, role)
		SELECT w.id, $3, $4
		FROM workspaces w
		JOIN organizations o ON o.id = w.org_id
		WHERE o.slug = $1 AND w.slug = $2 AND w.deleted_at IS NULL
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, orgSlug, wsSlug, targetUserID, role)
	if err != nil {
		return fmt.Errorf("pg_org_store: AddWorkspaceMember: %w", err)
	}
	return nil
}

// RemoveWorkspaceMember removes a user from a workspace.
func (s *PGOrgStore) RemoveWorkspaceMember(ctx context.Context, orgSlug, wsSlug, targetUserID string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM workspace_memberships wm
		USING workspaces w, organizations o
		WHERE wm.workspace_id = w.id
		  AND w.org_id = o.id
		  AND o.slug = $1 AND w.slug = $2 AND wm.user_id = $3
		  AND w.deleted_at IS NULL
	`, orgSlug, wsSlug, targetUserID)
	if err != nil {
		return fmt.Errorf("pg_org_store: RemoveWorkspaceMember: %w", err)
	}
	return nil
}

// pgSlugify converts a display name to a URL-safe slug.
func pgSlugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		result = "org"
	}
	return result
}
