package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// ChangesetStore is a concurrency-safe in-memory store for changesets.
// It implements usecase.ChangesetRepository.
type ChangesetStore struct {
	mu         sync.RWMutex
	changesets map[string]*usecase.StoredChangeset
}

// NewChangesetStore constructs an empty ChangesetStore.
func NewChangesetStore() *ChangesetStore {
	return &ChangesetStore{changesets: make(map[string]*usecase.StoredChangeset)}
}

// Store saves a changeset associated with a model and returns its generated ID.
func (s *ChangesetStore) Store(modelID string, cs *entity.Changeset) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", fmt.Errorf("changeset store: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.changesets[id] = &usecase.StoredChangeset{
		ID:        id,
		ModelID:   modelID,
		Changeset: cs,
		CreatedAt: time.Now(),
	}
	return id, nil
}

// Get retrieves a stored changeset by ID.
// Returns ErrNotFound if the changeset does not exist.
func (s *ChangesetStore) Get(id string) (*usecase.StoredChangeset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc := s.changesets[id]
	if sc == nil {
		return nil, usecase.ErrNotFound
	}
	return sc, nil
}

// ListForModel returns all changesets associated with a given model ID.
func (s *ChangesetStore) ListForModel(modelID string) ([]*usecase.StoredChangeset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*usecase.StoredChangeset
	for _, sc := range s.changesets {
		if sc.ModelID == modelID {
			result = append(result, sc)
		}
	}
	return result, nil
}

// Update replaces the changeset actions for an existing changeset.
// Returns ErrNotFound if the changeset does not exist.
func (s *ChangesetStore) Update(id string, cs *entity.Changeset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc, ok := s.changesets[id]
	if !ok {
		return usecase.ErrNotFound
	}
	sc.Changeset = cs
	return nil
}

// Delete removes a single changeset by ID. Idempotent.
func (s *ChangesetStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.changesets, id)
	return nil
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
