package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// PGChangesetStore implements usecase.ChangesetRepository backed by PostgreSQL.
type PGChangesetStore struct {
	db           *pgxpool.Pool
	systemUserID uuid.UUID
}

// NewPGChangesetStore creates a PGChangesetStore.
// systemUserID must match the UUID bootstrapped by PGModelStore.
func NewPGChangesetStore(db *pgxpool.Pool, systemUserID uuid.UUID) *PGChangesetStore {
	return &PGChangesetStore{db: db, systemUserID: systemUserID}
}

// Store inserts a new changeset row and returns the generated UUID.
func (s *PGChangesetStore) Store(modelID string, cs *entity.Changeset) (string, error) {
	modelUID, err := uuid.Parse(modelID)
	if err != nil {
		return "", fmt.Errorf("invalid model ID: %w", err)
	}

	actionsJSON, err := json.Marshal(cs.Actions)
	if err != nil {
		return "", fmt.Errorf("marshal changeset actions: %w", err)
	}

	var id uuid.UUID
	err = s.db.QueryRow(context.Background(),
		`INSERT INTO changesets (model_id, title, actions_json, created_by)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		modelUID, cs.Description, actionsJSON, s.systemUserID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert changeset: %w", err)
	}

	return id.String(), nil
}

// Get returns a stored changeset by ID.
func (s *PGChangesetStore) Get(id string) (*usecase.StoredChangeset, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, usecase.ErrNotFound
	}

	var modelID uuid.UUID
	var title string
	var actionsJSON []byte
	var createdAt time.Time

	err = s.db.QueryRow(context.Background(),
		`SELECT model_id, title, actions_json, created_at
		 FROM changesets WHERE id = $1 AND deleted_at IS NULL`,
		uid,
	).Scan(&modelID, &title, &actionsJSON, &createdAt)
	if err != nil {
		return nil, usecase.ErrNotFound
	}

	cs, err := entity.NewChangeset(id, title)
	if err != nil {
		return nil, fmt.Errorf("reconstruct changeset: %w", err)
	}

	if err := json.Unmarshal(actionsJSON, &cs.Actions); err != nil {
		return nil, fmt.Errorf("unmarshal changeset actions: %w", err)
	}

	return &usecase.StoredChangeset{
		ID:        id,
		ModelID:   modelID.String(),
		Changeset: cs,
		CreatedAt: createdAt,
	}, nil
}

// ListForModel returns all non-deleted changesets for a given model.
func (s *PGChangesetStore) ListForModel(modelID string) ([]*usecase.StoredChangeset, error) {
	modelUID, err := uuid.Parse(modelID)
	if err != nil {
		return nil, nil // unknown model → empty
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT id, model_id, title, actions_json, created_at
		 FROM changesets WHERE model_id = $1 AND deleted_at IS NULL ORDER BY created_at`,
		modelUID,
	)
	if err != nil {
		return nil, fmt.Errorf("list changesets: %w", err)
	}
	defer rows.Close()

	var result []*usecase.StoredChangeset
	for rows.Next() {
		var id, mid uuid.UUID
		var title string
		var actionsJSON []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &mid, &title, &actionsJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("scan changeset row: %w", err)
		}

		cs, err := entity.NewChangeset(id.String(), title)
		if err != nil {
			return nil, fmt.Errorf("reconstruct changeset: %w", err)
		}
		if err := json.Unmarshal(actionsJSON, &cs.Actions); err != nil {
			return nil, fmt.Errorf("unmarshal actions: %w", err)
		}

		result = append(result, &usecase.StoredChangeset{
			ID:        id.String(),
			ModelID:   mid.String(),
			Changeset: cs,
			CreatedAt: createdAt,
		})
	}
	return result, rows.Err()
}

// Update replaces the actions and title of an existing changeset.
func (s *PGChangesetStore) Update(id string, cs *entity.Changeset) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return usecase.ErrNotFound
	}

	actionsJSON, err := json.Marshal(cs.Actions)
	if err != nil {
		return fmt.Errorf("marshal actions: %w", err)
	}

	tag, err := s.db.Exec(context.Background(),
		`UPDATE changesets SET actions_json = $1, title = $2, updated_at = NOW()
		 WHERE id = $3 AND deleted_at IS NULL`,
		actionsJSON, cs.Description, uid,
	)
	if err != nil {
		return fmt.Errorf("update changeset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return usecase.ErrNotFound
	}
	return nil
}

// Delete soft-deletes a single changeset. Idempotent.
func (s *PGChangesetStore) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	_, err = s.db.Exec(context.Background(),
		`UPDATE changesets SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, uid)
	return err
}

// DeleteForModel soft-deletes all changesets for a model and returns the count deleted.
func (s *PGChangesetStore) DeleteForModel(modelID string) int {
	modelUID, err := uuid.Parse(modelID)
	if err != nil {
		return 0
	}

	tag, err := s.db.Exec(context.Background(),
		`UPDATE changesets SET deleted_at = NOW()
		 WHERE model_id = $1 AND deleted_at IS NULL`,
		modelUID,
	)
	if err != nil {
		return 0
	}
	return int(tag.RowsAffected())
}
