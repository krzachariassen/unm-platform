package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func TestMemOrgStore_EnsureDevUser(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	userID, orgSlug, wsSlug, err := s.EnsureDevUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", userID)
	assert.Equal(t, "local", orgSlug)
	assert.Equal(t, "default", wsSlug)

	// Idempotent — calling again should not error.
	userID2, orgSlug2, wsSlug2, err := s.EnsureDevUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, userID, userID2)
	assert.Equal(t, orgSlug, orgSlug2)
	assert.Equal(t, wsSlug, wsSlug2)
}

func TestMemOrgStore_EnsureUser(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	id, err := s.EnsureUser(ctx, "test@example.com", "Test User", "https://avatar.example.com/u")
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// Upsert — same email returns same ID.
	id2, err := s.EnsureUser(ctx, "test@example.com", "Test User Updated", "")
	require.NoError(t, err)
	assert.Equal(t, id, id2)
}

func TestMemOrgStore_CreateOrg(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, err := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	require.NoError(t, err)

	org, err := s.CreateOrg(ctx, ownerID, "ACME Corp", "acme")
	require.NoError(t, err)
	assert.Equal(t, "ACME Corp", org.Name)
	assert.Equal(t, "acme", org.Slug)
	assert.Equal(t, "owner", org.Role)
}

func TestMemOrgStore_GetOrg(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, _ := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	_, _ = s.CreateOrg(ctx, ownerID, "ACME Corp", "acme")

	// Owner can get org.
	org, err := s.GetOrg(ctx, ownerID, "acme")
	require.NoError(t, err)
	assert.Equal(t, "acme", org.Slug)
	assert.Equal(t, "owner", org.Role)

	// Unknown user gets ErrNotFound.
	_, err = s.GetOrg(ctx, "unknown-user", "acme")
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestMemOrgStore_ListOrgs(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	userID, _ := s.EnsureUser(ctx, "u@example.com", "User", "")
	_, _ = s.CreateOrg(ctx, userID, "Org A", "org-a")
	_, _ = s.CreateOrg(ctx, userID, "Org B", "org-b")

	orgs, err := s.ListOrgs(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestMemOrgStore_CreateWorkspace(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, _ := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	_, _ = s.CreateOrg(ctx, ownerID, "ACME", "acme")

	ws, err := s.CreateWorkspace(ctx, ownerID, "acme", "Team Alpha", "alpha", "private")
	require.NoError(t, err)
	assert.Equal(t, "alpha", ws.Slug)
	assert.Equal(t, "acme", ws.OrgSlug)
	assert.Equal(t, "admin", ws.Role)
	assert.Equal(t, "private", ws.Visibility)
}

func TestMemOrgStore_GetWorkspace(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, _ := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	_, _ = s.CreateOrg(ctx, ownerID, "ACME", "acme")
	_, _ = s.CreateWorkspace(ctx, ownerID, "acme", "Alpha", "alpha", "private")

	// Owner gets workspace.
	ws, err := s.GetWorkspace(ctx, ownerID, "acme", "alpha")
	require.NoError(t, err)
	assert.Equal(t, "alpha", ws.Slug)
	assert.Equal(t, "admin", ws.Role)

	// Unknown user cannot access private workspace.
	otherID, _ := s.EnsureUser(ctx, "other@example.com", "Other", "")
	_, err = s.GetWorkspace(ctx, otherID, "acme", "alpha")
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestMemOrgStore_OrgVisibleWorkspace(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, _ := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	_, _ = s.CreateOrg(ctx, ownerID, "ACME", "acme")
	_, _ = s.CreateWorkspace(ctx, ownerID, "acme", "Public WS", "pub", "org-visible")

	// Org member (not explicit ws member) can access org-visible workspace.
	memberID, _ := s.EnsureUser(ctx, "member@example.com", "Member", "")
	_ = s.AddOrgMember(ctx, "acme", memberID, "member")

	ws, err := s.GetWorkspace(ctx, memberID, "acme", "pub")
	require.NoError(t, err)
	assert.Equal(t, "pub", ws.Slug)
	// Role is empty for non-members (read-only access).
	assert.Equal(t, "", ws.Role)
}

func TestMemOrgStore_AddRemoveMember(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, _ := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	memberID, _ := s.EnsureUser(ctx, "member@example.com", "Member", "")
	_, _ = s.CreateOrg(ctx, ownerID, "ACME", "acme")

	err := s.AddOrgMember(ctx, "acme", memberID, "member")
	require.NoError(t, err)

	org, err := s.GetOrg(ctx, memberID, "acme")
	require.NoError(t, err)
	assert.Equal(t, "member", org.Role)

	err = s.RemoveOrgMember(ctx, "acme", memberID)
	require.NoError(t, err)

	_, err = s.GetOrg(ctx, memberID, "acme")
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestMemOrgStore_OnboardNewUser(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	userID, _ := s.EnsureUser(ctx, "newuser@example.com", "New User", "")
	orgInfo, wsInfo, err := s.OnboardNewUser(ctx, userID, "New User")
	require.NoError(t, err)
	assert.NotEmpty(t, orgInfo.Slug)
	assert.Equal(t, "owner", orgInfo.Role)
	assert.Equal(t, "general", wsInfo.Slug)
	assert.Equal(t, "admin", wsInfo.Role)
	assert.Equal(t, "org-visible", wsInfo.Visibility)
}

func TestMemOrgStore_WorkspaceMembership(t *testing.T) {
	s := repository.NewMemOrgStore()
	ctx := context.Background()

	ownerID, _ := s.EnsureUser(ctx, "owner@example.com", "Owner", "")
	editorID, _ := s.EnsureUser(ctx, "editor@example.com", "Editor", "")
	_, _ = s.CreateOrg(ctx, ownerID, "ACME", "acme")
	_, _ = s.CreateWorkspace(ctx, ownerID, "acme", "Alpha", "alpha", "private")

	err := s.AddWorkspaceMember(ctx, "acme", "alpha", editorID, "editor")
	require.NoError(t, err)

	ws, err := s.GetWorkspace(ctx, editorID, "acme", "alpha")
	require.NoError(t, err)
	assert.Equal(t, "editor", ws.Role)

	err = s.RemoveWorkspaceMember(ctx, "acme", "alpha", editorID)
	require.NoError(t, err)

	_, err = s.GetWorkspace(ctx, editorID, "acme", "alpha")
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}
