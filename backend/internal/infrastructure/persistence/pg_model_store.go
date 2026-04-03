package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/serializer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// PGModelStore implements usecase.ModelRepository backed by PostgreSQL.
// For Phase 14A, the parsed model is also cached in memory after load.
// A "system" user + workspace are bootstrapped on startup until auth (Phase 15).
type PGModelStore struct {
	db                *pgxpool.Pool
	mu                sync.RWMutex
	cache             map[string]*entity.UNMModel
	onDelete          func(modelID string)
	systemUserID      uuid.UUID
	systemWorkspaceID uuid.UUID
}

// NewPGModelStore creates a PGModelStore, bootstrapping the system user and workspace if needed.
func NewPGModelStore(db *pgxpool.Pool) (*PGModelStore, error) {
	s := &PGModelStore{
		db:    db,
		cache: make(map[string]*entity.UNMModel),
	}
	if err := s.bootstrap(context.Background()); err != nil {
		return nil, fmt.Errorf("PGModelStore bootstrap: %w", err)
	}
	return s, nil
}

// bootstrap ensures the system user and default workspace exist.
func (s *PGModelStore) bootstrap(ctx context.Context) error {
	var userID uuid.UUID
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (email, name) VALUES ('system@unm-platform.local', 'System')
		 ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
	).Scan(&userID)
	if err != nil {
		return fmt.Errorf("upsert system user: %w", err)
	}
	s.systemUserID = userID

	// We need an org before a workspace.
	var orgID uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO organizations (name, slug) VALUES ('Default', 'default')
		 ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
	).Scan(&orgID)
	if err != nil {
		return fmt.Errorf("upsert default org: %w", err)
	}

	var wsID uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO workspaces (org_id, name, slug, created_by) VALUES ($1, 'Default', 'default', $2)
		 ON CONFLICT (org_id, slug) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
		orgID, userID,
	).Scan(&wsID)
	if err != nil {
		return fmt.Errorf("upsert default workspace: %w", err)
	}
	s.systemWorkspaceID = wsID
	return nil
}

// Store serializes the model to YAML, inserts it into the models table, and caches it.
func (s *PGModelStore) Store(m *entity.UNMModel) (string, error) {
	raw, err := serializer.MarshalYAML(m)
	if err != nil {
		return "", fmt.Errorf("serialize model: %w", err)
	}

	name := "untitled"
	if m != nil && m.System.Name != "" {
		name = m.System.Name
	}

	var id uuid.UUID
	err = s.db.QueryRow(context.Background(),
		`INSERT INTO models (workspace_id, name, created_by) VALUES ($1, $2, $3) RETURNING id`,
		s.systemWorkspaceID, name, s.systemUserID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert model: %w", err)
	}

	// Store raw content as version 1
	_, err = s.db.Exec(context.Background(),
		`INSERT INTO model_versions (model_id, version, raw_content, committed_by) VALUES ($1, 1, $2, $3)`,
		id, string(raw), s.systemUserID,
	)
	if err != nil {
		return "", fmt.Errorf("insert model version: %w", err)
	}

	idStr := id.String()
	s.mu.Lock()
	s.cache[idStr] = m
	s.mu.Unlock()

	return idStr, nil
}

// Get returns a stored model by ID, loading from DB if not in cache.
func (s *PGModelStore) Get(id string) (*usecase.StoredModel, error) {
	s.mu.RLock()
	m, ok := s.cache[id]
	s.mu.RUnlock()

	if ok {
		return &usecase.StoredModel{
			ID:             id,
			Model:          m,
			CreatedAt:      time.Now(), // approximation for cached items
			LastAccessedAt: time.Now(),
		}, nil
	}

	// Load from DB
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, usecase.ErrNotFound
	}

	var rawContent string
	var createdAt time.Time
	err = s.db.QueryRow(context.Background(),
		`SELECT mv.raw_content, mo.created_at
		 FROM models mo
		 JOIN model_versions mv ON mv.model_id = mo.id AND mv.version = (
		     SELECT MAX(version) FROM model_versions WHERE model_id = mo.id AND deleted_at IS NULL
		 )
		 WHERE mo.id = $1 AND mo.deleted_at IS NULL`,
		uid,
	).Scan(&rawContent, &createdAt)
	if err != nil {
		return nil, usecase.ErrNotFound
	}

	p := parser.NewYAMLParser()
	model, err := p.Parse(bytes.NewBufferString(rawContent))
	if err != nil {
		return nil, fmt.Errorf("parse stored model: %w", err)
	}

	s.mu.Lock()
	s.cache[id] = model
	s.mu.Unlock()

	return &usecase.StoredModel{
		ID:             id,
		Model:          model,
		CreatedAt:      createdAt,
		LastAccessedAt: time.Now(),
	}, nil
}

// Replace updates the in-memory cache and persists a new model version.
func (s *PGModelStore) Replace(id string, newModel *entity.UNMModel) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return usecase.ErrNotFound
	}

	// Verify it exists.
	var exists bool
	err = s.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM models WHERE id = $1 AND deleted_at IS NULL)`, uid,
	).Scan(&exists)
	if err != nil || !exists {
		return usecase.ErrNotFound
	}

	raw, err := serializer.MarshalYAML(newModel)
	if err != nil {
		return fmt.Errorf("serialize model: %w", err)
	}

	_, err = s.db.Exec(context.Background(),
		`INSERT INTO model_versions (model_id, version, raw_content, committed_by)
		 VALUES ($1,
		     (SELECT COALESCE(MAX(version), 0) + 1 FROM model_versions WHERE model_id = $1),
		     $2, $3)`,
		uid, string(raw), s.systemUserID,
	)
	if err != nil {
		return fmt.Errorf("insert replacement version: %w", err)
	}

	_, err = s.db.Exec(context.Background(),
		`UPDATE models SET updated_at = NOW() WHERE id = $1`, uid,
	)
	if err != nil {
		return fmt.Errorf("update model timestamp: %w", err)
	}

	s.mu.Lock()
	s.cache[id] = newModel
	s.mu.Unlock()

	return nil
}

// List returns metadata for all non-deleted models.
func (s *PGModelStore) List() ([]*usecase.StoredModel, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, created_at FROM models WHERE deleted_at IS NULL ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	var result []*usecase.StoredModel
	for rows.Next() {
		var id uuid.UUID
		var createdAt time.Time
		if err := rows.Scan(&id, &createdAt); err != nil {
			return nil, fmt.Errorf("scan model row: %w", err)
		}
		idStr := id.String()
		s.mu.RLock()
		m := s.cache[idStr]
		s.mu.RUnlock()
		result = append(result, &usecase.StoredModel{
			ID:             idStr,
			Model:          m, // may be nil if not cached; callers should use Get for full model
			CreatedAt:      createdAt,
			LastAccessedAt: createdAt,
		})
	}
	return result, rows.Err()
}

// Delete soft-deletes a model. Idempotent — no error if already deleted.
func (s *PGModelStore) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil // not found → treat as already gone
	}

	_, err = s.db.Exec(context.Background(),
		`UPDATE models SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, uid)
	if err != nil {
		return fmt.Errorf("soft-delete model: %w", err)
	}

	s.mu.Lock()
	delete(s.cache, id)
	s.mu.Unlock()

	if s.onDelete != nil {
		s.onDelete(id)
	}
	return nil
}

// SystemUserID returns the bootstrapped system user UUID, needed for PGChangesetStore.
func (s *PGModelStore) SystemUserID() uuid.UUID {
	return s.systemUserID
}

// SetOnDelete stores a callback invoked after a model is deleted.
func (s *PGModelStore) SetOnDelete(fn func(modelID string)) {
	s.mu.Lock()
	s.onDelete = fn
	s.mu.Unlock()
}

// StartEviction is a no-op for PG store (models are persisted, not evicted).
func (s *PGModelStore) StartEviction(ttl, interval time.Duration) {}

// StopEviction is a no-op for PG store.
func (s *PGModelStore) StopEviction() {}

// marshalActions converts a Changeset's actions to JSON (used by PGChangesetStore).
func marshalActions(cs *entity.Changeset) ([]byte, error) {
	if cs == nil || len(cs.Actions) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(cs.Actions)
}
