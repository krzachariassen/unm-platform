package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// ModelStore is a concurrency-safe in-memory store for parsed UNMModels.
// It implements usecase.ModelRepository.
type ModelStore struct {
	mu       sync.RWMutex
	models   map[string]*usecase.StoredModel
	onDelete func(modelID string) // cascade callback
	stopCh   chan struct{}        // signals the eviction goroutine to stop
}

// NewModelStore constructs an empty ModelStore.
func NewModelStore() *ModelStore {
	return &ModelStore{models: make(map[string]*usecase.StoredModel)}
}

// Store saves a model and returns its generated ID.
func (s *ModelStore) Store(m *entity.UNMModel) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", fmt.Errorf("model store: %w", err)
	}
	now := time.Now()
	if m != nil {
		stampMeta(&m.Meta, now)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.models[id] = &usecase.StoredModel{ID: id, Model: m, CreatedAt: now, LastAccessedAt: now, VersionCount: 1}
	return id, nil
}

// Get retrieves a stored model by ID and updates its LastAccessedAt.
// Returns ErrNotFound if the model does not exist.
func (s *ModelStore) Get(id string) (*usecase.StoredModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.models[id]
	if m == nil {
		return nil, usecase.ErrNotFound
	}
	m.LastAccessedAt = time.Now()
	return m, nil
}

// Replace swaps the model stored under the given ID with a new model.
// Returns ErrNotFound if the ID does not exist.
func (s *ModelStore) Replace(id string, newModel *entity.UNMModel) error {
	return s.replaceWithMessage(id, newModel, "")
}

// ReplaceWithMessage is like Replace but accepts a commit message (ignored by the memory store).
func (s *ModelStore) ReplaceWithMessage(id string, newModel *entity.UNMModel, message string) error {
	return s.replaceWithMessage(id, newModel, message)
}

func (s *ModelStore) replaceWithMessage(id string, newModel *entity.UNMModel, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.models[id]
	if !ok {
		return usecase.ErrNotFound
	}
	if newModel != nil && existing.Model != nil {
		newModel.Meta.Version = existing.Model.Meta.Version
		newModel.Meta.Author = existing.Model.Meta.Author
		stampMeta(&newModel.Meta, time.Now())
	}
	existing.Model = newModel
	existing.LastAccessedAt = time.Now()
	existing.VersionCount++
	return nil
}

// List returns all stored models with VersionCount = 1 (memory store has no history).
func (s *ModelStore) List() ([]*usecase.StoredModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*usecase.StoredModel, 0, len(s.models))
	for _, m := range s.models {
		result = append(result, m)
	}
	return result, nil
}

// Delete removes a model by ID. Idempotent — does not error if not found.
func (s *ModelStore) Delete(id string) error {
	s.mu.Lock()
	_, existed := s.models[id]
	delete(s.models, id)
	cb := s.onDelete
	s.mu.Unlock()
	if existed && cb != nil {
		cb(id)
	}
	return nil
}

// ListVersions returns a single version entry for the memory store (no history).
func (s *ModelStore) ListVersions(modelID string) ([]usecase.ModelVersionMeta, error) {
	s.mu.RLock()
	stored, ok := s.models[modelID]
	s.mu.RUnlock()
	if !ok {
		return nil, usecase.ErrNotFound
	}
	return []usecase.ModelVersionMeta{
		{
			ID:          modelID + "-v1",
			ModelID:     modelID,
			Version:     1,
			CommittedAt: stored.CreatedAt,
		},
	}, nil
}

// GetVersion returns the current model for version 1; returns ErrNotFound for any other version.
func (s *ModelStore) GetVersion(modelID string, version int) (*entity.UNMModel, error) {
	if version != 1 {
		return nil, usecase.ErrNotFound
	}
	stored, err := s.Get(modelID)
	if err != nil {
		return nil, err
	}
	return stored.Model, nil
}

// DiffVersions computes a diff. Memory store only has version 1, so fromV and toV must both be 1.
func (s *ModelStore) DiffVersions(modelID string, fromV, toV int) (*service.ModelDiff, error) {
	fromModel, err := s.GetVersion(modelID, fromV)
	if err != nil {
		return nil, fmt.Errorf("from version %d: %w", fromV, err)
	}
	toModel, err := s.GetVersion(modelID, toV)
	if err != nil {
		return nil, fmt.Errorf("to version %d: %w", toV, err)
	}
	diff := service.Diff(fromModel, toModel)
	diff.FromVersion = fromV
	diff.ToVersion = toV
	return diff, nil
}

// Len returns the number of stored models.
func (s *ModelStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.models)
}

// SetOnDelete registers a callback invoked after a model is deleted.
func (s *ModelStore) SetOnDelete(fn func(modelID string)) {
	s.onDelete = fn
}

// StartEviction launches a background goroutine that periodically removes
// models that have not been accessed within the given TTL.
func (s *ModelStore) StartEviction(ttl, interval time.Duration) {
	s.stopCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.evictExpired(ttl)
			case <-s.stopCh:
				return
			}
		}
	}()
	log.Printf("model eviction started: ttl=%v, interval=%v", ttl, interval)
}

// StopEviction stops the background eviction goroutine.
func (s *ModelStore) StopEviction() {
	if s.stopCh != nil {
		close(s.stopCh)
	}
}

func (s *ModelStore) evictExpired(ttl time.Duration) {
	s.mu.RLock()
	var expired []string
	now := time.Now()
	for id, m := range s.models {
		if now.Sub(m.LastAccessedAt) > ttl {
			expired = append(expired, id)
		}
	}
	s.mu.RUnlock()

	for _, id := range expired {
		_ = s.Delete(id)
		log.Printf("evicted model %s (inactive > %v)", id, ttl)
	}
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// stampMeta increments the version and sets LastModified.
func stampMeta(meta *entity.ModelMeta, now time.Time) {
	var v int
	if _, err := fmt.Sscanf(meta.Version, "%d", &v); err != nil || v < 1 {
		v = 0
	}
	meta.Version = fmt.Sprintf("%d", v+1)
	meta.LastModified = now.UTC().Format(time.RFC3339)
}
