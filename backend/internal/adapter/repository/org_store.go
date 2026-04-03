package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/krzachariassen/unm-platform/internal/usecase"
)

const devUserID = "00000000-0000-0000-0000-000000000001"

// MemOrgStore is an in-memory stub implementing usecase.OrgRepository.
// It is used in tests and when storage.driver=memory.
// NOT safe for concurrent writes in production — use PGOrgStore instead.
type MemOrgStore struct {
	mu           sync.RWMutex
	users        map[string]*memUser        // id → user
	emailIndex   map[string]string          // email → id
	orgs         map[string]*memOrg         // slug → org
	workspaces   map[string]*memWorkspace   // "orgSlug/wsSlug" → ws
	orgMembers   map[string]map[string]string // orgSlug → {userID → role}
	wsMembers    map[string]map[string]string // "orgSlug/wsSlug" → {userID → role}
	counter      int
}

type memUser struct {
	id        string
	email     string
	name      string
	avatarURL string
}

type memOrg struct {
	id   string
	name string
	slug string
}

type memWorkspace struct {
	id         string
	orgSlug    string
	name       string
	slug       string
	visibility string
}

// NewMemOrgStore returns an initialised in-memory org store.
func NewMemOrgStore() *MemOrgStore {
	return &MemOrgStore{
		users:      make(map[string]*memUser),
		emailIndex: make(map[string]string),
		orgs:       make(map[string]*memOrg),
		workspaces: make(map[string]*memWorkspace),
		orgMembers: make(map[string]map[string]string),
		wsMembers:  make(map[string]map[string]string),
	}
}

func (s *MemOrgStore) nextID() string {
	s.counter++
	return fmt.Sprintf("mem-%d", s.counter)
}

// EnsureUser upserts a user by email.
func (s *MemOrgStore) EnsureUser(_ context.Context, email, name, avatarURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.emailIndex[email]; ok {
		u := s.users[id]
		u.name = name
		u.avatarURL = avatarURL
		return id, nil
	}
	id := s.nextID()
	s.users[id] = &memUser{id: id, email: email, name: name, avatarURL: avatarURL}
	s.emailIndex[email] = id
	return id, nil
}

// EnsureDevUser creates the hardcoded dev user + "local" org + "default" workspace.
func (s *MemOrgStore) EnsureDevUser(ctx context.Context) (string, string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := devUserID
	if _, ok := s.users[id]; !ok {
		s.users[id] = &memUser{id: id, email: "local@dev", name: "Local Dev User"}
		s.emailIndex["local@dev"] = id
	}

	const orgSlug = "local"
	const wsSlug = "default"

	if _, ok := s.orgs[orgSlug]; !ok {
		s.orgs[orgSlug] = &memOrg{id: s.nextID(), name: "Local Dev Org", slug: orgSlug}
		s.orgMembers[orgSlug] = map[string]string{id: "owner"}
	}

	wsKey := orgSlug + "/" + wsSlug
	if _, ok := s.workspaces[wsKey]; !ok {
		s.workspaces[wsKey] = &memWorkspace{
			id:         s.nextID(),
			orgSlug:    orgSlug,
			name:       "Default",
			slug:       wsSlug,
			visibility: "org-visible",
		}
		s.wsMembers[wsKey] = map[string]string{id: "admin"}
	}

	return id, orgSlug, wsSlug, nil
}

// OnboardNewUser creates a personal org + General workspace for a new user.
func (s *MemOrgStore) OnboardNewUser(ctx context.Context, userID, displayName string) (*usecase.OrgInfo, *usecase.WorkspaceInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slug := slugify(displayName) + "-org"
	// Ensure uniqueness.
	base := slug
	for i := 2; ; i++ {
		if _, taken := s.orgs[slug]; !taken {
			break
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}

	org := &memOrg{id: s.nextID(), name: displayName + "'s Org", slug: slug}
	s.orgs[slug] = org
	s.orgMembers[slug] = map[string]string{userID: "owner"}

	wsKey := slug + "/general"
	ws := &memWorkspace{id: s.nextID(), orgSlug: slug, name: "General", slug: "general", visibility: "org-visible"}
	s.workspaces[wsKey] = ws
	s.wsMembers[wsKey] = map[string]string{userID: "admin"}

	orgInfo := &usecase.OrgInfo{ID: org.id, Name: org.name, Slug: org.slug, Role: "owner"}
	wsInfo := &usecase.WorkspaceInfo{ID: ws.id, OrgID: org.id, OrgSlug: slug, Name: ws.name, Slug: ws.slug, Visibility: ws.visibility, Role: "admin"}
	return orgInfo, wsInfo, nil
}

// CreateOrg creates an org with owner and a General workspace.
func (s *MemOrgStore) CreateOrg(ctx context.Context, ownerID, name, slug string) (*usecase.OrgInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orgs[slug]; ok {
		return nil, fmt.Errorf("org slug %q already taken", slug)
	}
	org := &memOrg{id: s.nextID(), name: name, slug: slug}
	s.orgs[slug] = org
	s.orgMembers[slug] = map[string]string{ownerID: "owner"}

	wsKey := slug + "/general"
	s.workspaces[wsKey] = &memWorkspace{id: s.nextID(), orgSlug: slug, name: "General", slug: "general", visibility: "org-visible"}
	s.wsMembers[wsKey] = map[string]string{ownerID: "admin"}

	return &usecase.OrgInfo{ID: org.id, Name: name, Slug: slug, Role: "owner"}, nil
}

// GetOrg returns org info for a user.
func (s *MemOrgStore) GetOrg(_ context.Context, userID, orgSlug string) (*usecase.OrgInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgSlug]
	if !ok {
		return nil, usecase.ErrNotFound
	}
	members := s.orgMembers[orgSlug]
	role, isMember := members[userID]
	if !isMember {
		return nil, usecase.ErrNotFound
	}
	return &usecase.OrgInfo{ID: org.id, Name: org.name, Slug: org.slug, Role: role}, nil
}

// ListOrgs returns all orgs the user belongs to.
func (s *MemOrgStore) ListOrgs(_ context.Context, userID string) ([]*usecase.OrgInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*usecase.OrgInfo
	for slug, members := range s.orgMembers {
		if role, ok := members[userID]; ok {
			org := s.orgs[slug]
			result = append(result, &usecase.OrgInfo{ID: org.id, Name: org.name, Slug: org.slug, Role: role})
		}
	}
	return result, nil
}

// CreateWorkspace creates a workspace in an org.
func (s *MemOrgStore) CreateWorkspace(_ context.Context, createdByUserID, orgSlug, name, slug, visibility string) (*usecase.WorkspaceInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	org, ok := s.orgs[orgSlug]
	if !ok {
		return nil, usecase.ErrNotFound
	}
	wsKey := orgSlug + "/" + slug
	if _, taken := s.workspaces[wsKey]; taken {
		return nil, fmt.Errorf("workspace slug %q already exists in org", slug)
	}
	ws := &memWorkspace{id: s.nextID(), orgSlug: orgSlug, name: name, slug: slug, visibility: visibility}
	s.workspaces[wsKey] = ws
	s.wsMembers[wsKey] = map[string]string{createdByUserID: "admin"}

	return &usecase.WorkspaceInfo{ID: ws.id, OrgID: org.id, OrgSlug: orgSlug, Name: name, Slug: slug, Visibility: visibility, Role: "admin"}, nil
}

// GetWorkspace returns workspace info for a user. Includes org-visible workspaces.
func (s *MemOrgStore) GetWorkspace(_ context.Context, userID, orgSlug, wsSlug string) (*usecase.WorkspaceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, orgOk := s.orgs[orgSlug]
	if !orgOk {
		return nil, usecase.ErrNotFound
	}

	wsKey := orgSlug + "/" + wsSlug
	ws, wsOk := s.workspaces[wsKey]
	if !wsOk {
		return nil, usecase.ErrNotFound
	}

	members := s.wsMembers[wsKey]
	role := members[userID]

	// Org-visible: accessible if user is an org member.
	if role == "" && ws.visibility == "org-visible" {
		if _, isOrgMember := s.orgMembers[orgSlug][userID]; !isOrgMember {
			return nil, usecase.ErrNotFound
		}
	} else if role == "" {
		return nil, usecase.ErrNotFound
	}

	return &usecase.WorkspaceInfo{ID: ws.id, OrgID: org.id, OrgSlug: orgSlug, Name: ws.name, Slug: ws.slug, Visibility: ws.visibility, Role: role}, nil
}

// ListWorkspaces returns all workspaces accessible to a user in an org.
func (s *MemOrgStore) ListWorkspaces(_ context.Context, userID, orgSlug string) ([]*usecase.WorkspaceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, ok := s.orgs[orgSlug]
	if !ok {
		return nil, usecase.ErrNotFound
	}

	_, isOrgMember := s.orgMembers[orgSlug][userID]

	prefix := orgSlug + "/"
	var result []*usecase.WorkspaceInfo
	for key, ws := range s.workspaces {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		members := s.wsMembers[key]
		role := members[userID]
		if role == "" && ws.visibility == "org-visible" && isOrgMember {
			role = ""
		} else if role == "" {
			continue
		}
		result = append(result, &usecase.WorkspaceInfo{
			ID: ws.id, OrgID: org.id, OrgSlug: orgSlug,
			Name: ws.name, Slug: ws.slug, Visibility: ws.visibility, Role: role,
		})
	}
	return result, nil
}

// AddOrgMember grants or updates a user's role in an org.
func (s *MemOrgStore) AddOrgMember(_ context.Context, orgSlug, targetUserID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[orgSlug]; !ok {
		return usecase.ErrNotFound
	}
	if s.orgMembers[orgSlug] == nil {
		s.orgMembers[orgSlug] = make(map[string]string)
	}
	s.orgMembers[orgSlug][targetUserID] = role
	return nil
}

// RemoveOrgMember removes a user from an org.
func (s *MemOrgStore) RemoveOrgMember(_ context.Context, orgSlug, targetUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[orgSlug]; !ok {
		return usecase.ErrNotFound
	}
	delete(s.orgMembers[orgSlug], targetUserID)
	return nil
}

// AddWorkspaceMember grants or updates a user's role in a workspace.
func (s *MemOrgStore) AddWorkspaceMember(_ context.Context, orgSlug, wsSlug, targetUserID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	wsKey := orgSlug + "/" + wsSlug
	if _, ok := s.workspaces[wsKey]; !ok {
		return usecase.ErrNotFound
	}
	if s.wsMembers[wsKey] == nil {
		s.wsMembers[wsKey] = make(map[string]string)
	}
	s.wsMembers[wsKey][targetUserID] = role
	return nil
}

// RemoveWorkspaceMember removes a user from a workspace.
func (s *MemOrgStore) RemoveWorkspaceMember(_ context.Context, orgSlug, wsSlug, targetUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	wsKey := orgSlug + "/" + wsSlug
	if _, ok := s.workspaces[wsKey]; !ok {
		return usecase.ErrNotFound
	}
	delete(s.wsMembers[wsKey], targetUserID)
	return nil
}

// slugify converts a display name into a URL-safe slug (lowercase, hyphens).
func slugify(s string) string {
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
