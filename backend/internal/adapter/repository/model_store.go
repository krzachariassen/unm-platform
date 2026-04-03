package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
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
// If the model has no version, it is initialized to 1.
// LastModified is stamped to the current time.
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
	s.models[id] = &usecase.StoredModel{ID: id, Model: m, CreatedAt: now, LastAccessedAt: now}
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

// Replace swaps the model stored under the given ID with a new model,
// preserving the ID and CreatedAt timestamp.
// Increments the model version and stamps LastModified.
// Returns ErrNotFound if the ID does not exist.
func (s *ModelStore) Replace(id string, newModel *entity.UNMModel) error {
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
	return nil
}

// List returns all stored models.
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
// Invokes the onDelete cascade callback if one is registered and the model existed.
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

// Len returns the number of stored models.
func (s *ModelStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.models)
}

// SetOnDelete registers a callback invoked after a model is deleted.
// Used to cascade-delete associated changesets.
func (s *ModelStore) SetOnDelete(fn func(modelID string)) {
	s.onDelete = fn
}

// StartEviction launches a background goroutine that periodically removes
// models that have not been accessed within the given TTL.
// Call StopEviction to shut it down cleanly.
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

// stampMeta increments the version (parsing the existing value as an integer,
// defaulting to 1 if absent or non-numeric) and sets LastModified to now.
func stampMeta(meta *entity.ModelMeta, now time.Time) {
	var v int
	if _, err := fmt.Sscanf(meta.Version, "%d", &v); err != nil || v < 1 {
		v = 0
	}
	meta.Version = fmt.Sprintf("%d", v+1)
	meta.LastModified = now.UTC().Format(time.RFC3339)
}
