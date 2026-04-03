package usecase

import "time"

// UserSession holds the data stored per authenticated session.
type UserSession struct {
	ID        string
	UserID    string
	Email     string
	Name      string
	AvatarURL string
	ExpiresAt time.Time
}

// OrgMembership holds an org the user belongs to and their role within it.
type OrgMembership struct {
	OrgID   string
	OrgName string
	OrgSlug string
	Role    string // "owner" | "admin" | "member"
}

// AuthUser is injected into request context by the auth or dev-mode middleware.
type AuthUser struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
	Orgs      []OrgMembership
}

// SessionRepository is the persistence contract for authentication sessions.
// Implementations: MemorySessionStore (dev) and PGSessionStore (postgres).
type SessionRepository interface {
	// Create stores a new session and returns its ID.
	Create(userID, email, name, avatarURL string, ttl time.Duration) (*UserSession, error)

	// Get retrieves a session by ID. Returns ErrNotFound if absent or expired.
	Get(id string) (*UserSession, error)

	// Delete removes a session by ID. Idempotent.
	Delete(id string) error
}
