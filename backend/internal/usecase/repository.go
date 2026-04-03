package usecase

import (
	"errors"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ErrNotFound is returned by repository Get methods when the requested resource does not exist.
var ErrNotFound = errors.New("not found")

// StoredModel holds a parsed UNMModel with its assigned ID and timestamps.
// This type lives in the usecase package so both adapter/repository and
// infrastructure/persistence can implement the same interface without import cycles.
type StoredModel struct {
	ID             string
	Model          *entity.UNMModel
	CreatedAt      time.Time
	LastAccessedAt time.Time
}

// StoredChangeset holds a changeset with its assigned ID, model association, and timestamps.
type StoredChangeset struct {
	ID        string
	ModelID   string
	Changeset *entity.Changeset
	CreatedAt time.Time
}

// ModelRepository is the persistence contract for UNM models.
// Both the in-memory store and the PostgreSQL store implement this interface.
type ModelRepository interface {
	// Store saves a model and returns its generated ID.
	Store(m *entity.UNMModel) (string, error)

	// Get retrieves a stored model by ID.
	// Returns ErrNotFound if the model does not exist.
	Get(id string) (*StoredModel, error)

	// Replace swaps the model stored under the given ID with a new model.
	// Returns ErrNotFound if the ID does not exist.
	Replace(id string, newModel *entity.UNMModel) error

	// List returns all stored models.
	List() ([]*StoredModel, error)

	// Delete removes a model by ID. Idempotent — does not error if the model does not exist.
	// Invokes the onDelete cascade callback if one is registered.
	Delete(id string) error

	// SetOnDelete registers a callback invoked after a model is deleted.
	SetOnDelete(fn func(modelID string))

	// StartEviction launches background TTL eviction (no-op for persistent stores).
	StartEviction(ttl, interval time.Duration)

	// StopEviction stops background TTL eviction (no-op for persistent stores).
	StopEviction()
}

// ChangesetRepository is the persistence contract for UNM changesets.
type ChangesetRepository interface {
	// Store saves a changeset associated with a model and returns its generated ID.
	Store(modelID string, cs *entity.Changeset) (string, error)

	// Get retrieves a stored changeset by ID.
	// Returns ErrNotFound if the changeset does not exist.
	Get(id string) (*StoredChangeset, error)

	// ListForModel returns all changesets associated with a given model ID.
	ListForModel(modelID string) ([]*StoredChangeset, error)

	// Update replaces the changeset actions for an existing changeset.
	// Returns ErrNotFound if the changeset does not exist.
	Update(id string, cs *entity.Changeset) error

	// Delete soft-deletes a single changeset. Idempotent.
	Delete(id string) error

	// DeleteForModel removes all changesets for a model. Returns the count deleted.
	DeleteForModel(modelID string) int
}
