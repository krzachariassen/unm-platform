package usecase

import (
	"errors"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	domainservice "github.com/krzachariassen/unm-platform/internal/domain/service"
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
	VersionCount   int // number of persisted versions (always 1 for memory store)
}

// StoredChangeset holds a changeset with its assigned ID, model association, and timestamps.
type StoredChangeset struct {
	ID        string
	ModelID   string
	Changeset *entity.Changeset
	CreatedAt time.Time
}

// ModelVersionMeta holds metadata for a single model version (no raw content).
type ModelVersionMeta struct {
	ID            string
	ModelID       string
	Version       int
	CommitMessage string
	CommittedAt   time.Time
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

	// ReplaceWithMessage is like Replace but records a commit message in the version row.
	// Memory store ignores the message; PG store persists it.
	ReplaceWithMessage(id string, newModel *entity.UNMModel, message string) error

	// List returns all stored models with version counts populated.
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

	// ListVersions returns metadata for all versions of a model, oldest first.
	// Memory store returns a single entry (version 1).
	ListVersions(modelID string) ([]ModelVersionMeta, error)

	// GetVersion retrieves the model as it was at a specific version number.
	// Returns ErrNotFound if the model or version does not exist.
	GetVersion(modelID string, version int) (*entity.UNMModel, error)

	// DiffVersions computes a structured diff between two version numbers of a model.
	DiffVersions(modelID string, fromV, toV int) (*domainservice.ModelDiff, error)
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
