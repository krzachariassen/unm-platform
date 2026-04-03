package persistence

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// PGSessionStore is the PostgreSQL-backed SessionRepository implementation.
type PGSessionStore struct {
	db *pgxpool.Pool
}

// NewPGSessionStore returns a PGSessionStore backed by the given pool.
func NewPGSessionStore(db *pgxpool.Pool) *PGSessionStore {
	return &PGSessionStore{db: db}
}

func (s *PGSessionStore) Create(userID, email, name, avatarURL string, ttl time.Duration) (*usecase.UserSession, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(ttl)
	_, err = s.db.Exec(context.Background(),
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		id, userID, expiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &usecase.UserSession{
		ID:        id,
		UserID:    userID,
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *PGSessionStore) Get(id string) (*usecase.UserSession, error) {
	var userID string
	var expiresAt time.Time
	err := s.db.QueryRow(context.Background(),
		`SELECT user_id, expires_at FROM sessions WHERE id = $1 AND expires_at > NOW()`,
		id,
	).Scan(&userID, &expiresAt)
	if err != nil {
		return nil, usecase.ErrNotFound
	}

	// Fetch user info.
	var email, name, avatarURL string
	err = s.db.QueryRow(context.Background(),
		`SELECT email, name, avatar_url FROM users WHERE id = $1`,
		userID,
	).Scan(&email, &name, &avatarURL)
	if err != nil {
		return nil, usecase.ErrNotFound
	}

	return &usecase.UserSession{
		ID:        id,
		UserID:    userID,
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *PGSessionStore) Delete(id string) error {
	_, err := s.db.Exec(context.Background(), `DELETE FROM sessions WHERE id = $1`, id)
	return err
}
