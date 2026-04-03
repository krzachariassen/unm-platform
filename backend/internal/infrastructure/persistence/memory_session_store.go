package persistence

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// MemorySessionStore is the in-memory SessionRepository implementation (dev mode / tests).
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*usecase.UserSession
}

// NewMemorySessionStore returns a ready-to-use MemorySessionStore.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]*usecase.UserSession)}
}

func (s *MemorySessionStore) Create(userID, email, name, avatarURL string, ttl time.Duration) (*usecase.UserSession, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}
	sess := &usecase.UserSession{
		ID:        id,
		UserID:    userID,
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
		ExpiresAt: time.Now().Add(ttl),
	}
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
	return sess, nil
}

func (s *MemorySessionStore) Get(id string) (*usecase.UserSession, error) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return nil, usecase.ErrNotFound
	}
	if time.Now().After(sess.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, id)
		s.mu.Unlock()
		return nil, usecase.ErrNotFound
	}
	return sess, nil
}

func (s *MemorySessionStore) Delete(id string) error {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
	return nil
}

// generateSessionID returns a 32-byte cryptographically random hex string.
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
