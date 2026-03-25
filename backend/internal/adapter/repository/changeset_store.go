package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/uber/unm-platform/internal/domain/entity"
)

// StoredChangeset holds a changeset with its assigned ID, model association, and creation time.
type StoredChangeset struct {
	ID        string
	ModelID   string
	Changeset *entity.Changeset
	CreatedAt time.Time
}

// ChangesetStore is a concurrency-safe in-memory store for changesets.
type ChangesetStore struct {
	mu         sync.RWMutex
	changesets map[string]*StoredChangeset
}

// NewChangesetStore constructs an empty ChangesetStore.
func NewChangesetStore() *ChangesetStore {
	return &ChangesetStore{changesets: make(map[string]*StoredChangeset)}
}

// Store saves a changeset associated with a model and returns its generated ID.
func (s *ChangesetStore) Store(modelID string, cs *entity.Changeset) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", fmt.Errorf("changeset store: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.changesets[id] = &StoredChangeset{
		ID:        id,
		ModelID:   modelID,
		Changeset: cs,
		CreatedAt: time.Now(),
	}
	return id, nil
}

// Get retrieves a stored changeset by ID. Returns nil if not found.
func (s *ChangesetStore) Get(id string) *StoredChangeset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.changesets[id]
}

// ListForModel returns all changesets associated with a given model ID.
func (s *ChangesetStore) ListForModel(modelID string) []*StoredChangeset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*StoredChangeset
	for _, sc := range s.changesets {
		if sc.ModelID == modelID {
			result = append(result, sc)
		}
	}
	return result
}

// DeleteForModel removes all changesets associated with a given model ID.
// Returns the number of changesets deleted.
func (s *ChangesetStore) DeleteForModel(modelID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for id, sc := range s.changesets {
		if sc.ModelID == modelID {
			delete(s.changesets, id)
			count++
		}
	}
	return count
}
