package persistence_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/persistence"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// connectTestDB opens a pool to UNM_TEST_DB_URL, runs migrations, and returns the pool.
// The test is skipped when the env var is not set.
func connectTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("UNM_TEST_DB_URL")
	if url == "" {
		t.Skip("UNM_TEST_DB_URL not set — skipping postgres contract tests")
	}

	if err := persistence.RunMigrations(url); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	db, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// setupTestOrgWorkspace inserts a unique org + workspace into the DB and returns
// (orgID, workspaceID, userID) for use in store constructors.
// A t.Cleanup is registered to hard-delete the org (CASCADE removes all children).
func setupTestOrgWorkspace(t *testing.T, db *pgxpool.Pool) (orgID, workspaceID uuid.UUID, userID uuid.UUID) {
	t.Helper()

	// Ensure a system user exists (reuse the bootstrap pattern).
	err := db.QueryRow(context.Background(),
		`INSERT INTO users (email, name) VALUES ('system@unm-platform.local', 'System')
		 ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
	).Scan(&userID)
	require.NoError(t, err)

	// Create a unique org for this test run.
	slug := fmt.Sprintf("test-%d-%d", time.Now().UnixNano(), rand.Intn(100000))
	err = db.QueryRow(context.Background(),
		`INSERT INTO organizations (name, slug) VALUES ($1, $2) RETURNING id`,
		slug, slug,
	).Scan(&orgID)
	require.NoError(t, err)

	// Create a workspace in that org.
	err = db.QueryRow(context.Background(),
		`INSERT INTO workspaces (org_id, name, slug, created_by) VALUES ($1, $2, $3, $4) RETURNING id`,
		orgID, slug, slug, userID,
	).Scan(&workspaceID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(),
			`DELETE FROM organizations WHERE id = $1`, orgID)
	})

	return orgID, workspaceID, userID
}

// minimalModel returns a small but valid UNMModel for testing.
func minimalModel(name string) *entity.UNMModel {
	return &entity.UNMModel{
		System: entity.System{Name: name},
	}
}

// minimalChangeset returns a Changeset for testing.
func minimalChangeset(desc string) *entity.Changeset {
	cs, _ := entity.NewChangeset("test-id", desc)
	return cs
}

// ── Model Repository contract ──────────────────────────────────────────────

func runModelRepositoryTests(t *testing.T, repo usecase.ModelRepository) {
	t.Helper()

	t.Run("StoreAndGet", func(t *testing.T) {
		m := minimalModel("Contract Test")
		id, err := repo.Store(m)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		stored, err := repo.Get(id)
		require.NoError(t, err)
		require.NotNil(t, stored)
		assert.Equal(t, id, stored.ID)
		assert.NotNil(t, stored.Model)
		assert.False(t, stored.CreatedAt.IsZero())
	})

	t.Run("GetMissing_ReturnsErrNotFound", func(t *testing.T) {
		_, err := repo.Get("00000000-0000-0000-0000-000000000000")
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("List", func(t *testing.T) {
		// Store two models
		_, err := repo.Store(minimalModel("List A"))
		require.NoError(t, err)
		_, err = repo.Store(minimalModel("List B"))
		require.NoError(t, err)

		items, err := repo.List()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 2)
	})

	t.Run("Delete_Idempotent", func(t *testing.T) {
		id, err := repo.Store(minimalModel("Delete Me"))
		require.NoError(t, err)

		err = repo.Delete(id)
		require.NoError(t, err)

		_, err = repo.Get(id)
		assert.True(t, errors.Is(err, usecase.ErrNotFound))

		// Second delete must not error.
		err = repo.Delete(id)
		require.NoError(t, err)
	})

	t.Run("Replace", func(t *testing.T) {
		id, err := repo.Store(minimalModel("Replace Original"))
		require.NoError(t, err)

		newModel := minimalModel("Replace Updated")
		err = repo.Replace(id, newModel)
		require.NoError(t, err)

		stored, err := repo.Get(id)
		require.NoError(t, err)
		// System name should reflect the update (PG store re-reads latest version)
		assert.NotNil(t, stored.Model)
	})

	t.Run("Replace_Missing_ReturnsErrNotFound", func(t *testing.T) {
		err := repo.Replace("00000000-0000-0000-0000-000000000000", minimalModel("Ghost"))
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("ReplaceWithMessage_NoError", func(t *testing.T) {
		id, err := repo.Store(minimalModel("RWM Original"))
		require.NoError(t, err)

		err = repo.ReplaceWithMessage(id, minimalModel("RWM Updated"), "test commit message")
		require.NoError(t, err)

		stored, err := repo.Get(id)
		require.NoError(t, err)
		assert.NotNil(t, stored.Model)
	})

	t.Run("ListVersions_ReturnsAtLeastOne", func(t *testing.T) {
		id, err := repo.Store(minimalModel("Versions Model"))
		require.NoError(t, err)

		versions, err := repo.ListVersions(id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(versions), 1)
		assert.Equal(t, id, versions[0].ModelID)
		assert.Equal(t, 1, versions[0].Version)
	})

	t.Run("ListVersions_UnknownModel_ReturnsErrNotFound", func(t *testing.T) {
		_, err := repo.ListVersions("00000000-0000-0000-0000-000000000000")
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("GetVersion_Valid_ReturnsModel", func(t *testing.T) {
		id, err := repo.Store(minimalModel("GetVersion Model"))
		require.NoError(t, err)

		m, err := repo.GetVersion(id, 1)
		require.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("GetVersion_Invalid_ReturnsErrNotFound", func(t *testing.T) {
		id, err := repo.Store(minimalModel("GetVersion Missing"))
		require.NoError(t, err)

		_, err = repo.GetVersion(id, 99)
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("DiffVersions_NoError", func(t *testing.T) {
		id, err := repo.Store(minimalModel("Diff Model"))
		require.NoError(t, err)

		diff, err := repo.DiffVersions(id, 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, diff)
		assert.Equal(t, 1, diff.FromVersion)
		assert.Equal(t, 1, diff.ToVersion)
	})
}

// ── Changeset Repository contract ──────────────────────────────────────────

func runChangesetRepositoryTests(t *testing.T, modelRepo usecase.ModelRepository, csRepo usecase.ChangesetRepository) {
	t.Helper()

	// Create a model to hang changesets from.
	modelID, err := modelRepo.Store(minimalModel("Changeset Owner"))
	require.NoError(t, err)

	t.Run("StoreAndGet", func(t *testing.T) {
		cs := minimalChangeset("first changeset")
		id, err := csRepo.Store(modelID, cs)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		got, err := csRepo.Get(id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, modelID, got.ModelID)
		assert.Equal(t, "first changeset", got.Changeset.Description)
		assert.False(t, got.CreatedAt.IsZero())
	})

	t.Run("GetMissing_ReturnsErrNotFound", func(t *testing.T) {
		_, err := csRepo.Get("00000000-0000-0000-0000-000000000000")
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("ListForModel", func(t *testing.T) {
		cs1 := minimalChangeset("list-a")
		cs2 := minimalChangeset("list-b")
		_, err := csRepo.Store(modelID, cs1)
		require.NoError(t, err)
		_, err = csRepo.Store(modelID, cs2)
		require.NoError(t, err)

		results, err := csRepo.ListForModel(modelID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)
		for _, r := range results {
			assert.Equal(t, modelID, r.ModelID)
		}
	})

	t.Run("ListForModel_UnknownModel_Empty", func(t *testing.T) {
		results, err := csRepo.ListForModel("00000000-0000-0000-0000-000000000000")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Update", func(t *testing.T) {
		id, err := csRepo.Store(modelID, minimalChangeset("original"))
		require.NoError(t, err)

		updated := minimalChangeset("updated description")
		err = csRepo.Update(id, updated)
		require.NoError(t, err)

		got, err := csRepo.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "updated description", got.Changeset.Description)
	})

	t.Run("Update_Missing_ReturnsErrNotFound", func(t *testing.T) {
		err := csRepo.Update("00000000-0000-0000-0000-000000000000", minimalChangeset("ghost"))
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("Delete_SingleID", func(t *testing.T) {
		id, err := csRepo.Store(modelID, minimalChangeset("to delete"))
		require.NoError(t, err)

		err = csRepo.Delete(id)
		require.NoError(t, err)

		_, err = csRepo.Get(id)
		assert.True(t, errors.Is(err, usecase.ErrNotFound))
	})

	t.Run("DeleteForModel", func(t *testing.T) {
		// Use a fresh model to avoid counting changesets from other tests.
		freshModel, err := modelRepo.Store(minimalModel("Cascade Model"))
		require.NoError(t, err)

		_, _ = csRepo.Store(freshModel, minimalChangeset("cs1"))
		_, _ = csRepo.Store(freshModel, minimalChangeset("cs2"))

		n := csRepo.DeleteForModel(freshModel)
		assert.Equal(t, 2, n)

		results, _ := csRepo.ListForModel(freshModel)
		assert.Empty(t, results)
	})
}

// ── Memory implementations ─────────────────────────────────────────────────

func TestMemoryModelRepository(t *testing.T) {
	runModelRepositoryTests(t, repository.NewModelStore())
}

func TestMemoryChangesetRepository(t *testing.T) {
	modelRepo := repository.NewModelStore()
	runChangesetRepositoryTests(t, modelRepo, repository.NewChangesetStore())
}

// ── PostgreSQL implementations ─────────────────────────────────────────────

func TestPostgresModelRepository(t *testing.T) {
	db := connectTestDB(t)
	_, wsID, userID := setupTestOrgWorkspace(t, db)
	store := persistence.NewPGModelStoreWithWorkspace(db, userID, wsID)
	runModelRepositoryTests(t, store)
}

func TestPostgresChangesetRepository(t *testing.T) {
	db := connectTestDB(t)
	_, wsID, userID := setupTestOrgWorkspace(t, db)
	store := persistence.NewPGModelStoreWithWorkspace(db, userID, wsID)
	csStore := persistence.NewPGChangesetStore(db, userID)
	runChangesetRepositoryTests(t, store, csStore)
}

// ── 14D.5: Purge unit test ─────────────────────────────────────────────────

// TestPostgresPurgeExpired verifies that purgeExpired (called via StartEviction)
// hard-deletes rows whose deleted_at is older than the retention window.
func TestPostgresPurgeExpired(t *testing.T) {
	db := connectTestDB(t)
	_, wsID, userID := setupTestOrgWorkspace(t, db)
	store := persistence.NewPGModelStoreWithWorkspace(db, userID, wsID)

	// Store a model and soft-delete it.
	id, err := store.Store(minimalModel("Purge Me"))
	require.NoError(t, err)
	require.NoError(t, store.Delete(id))

	// Back-date deleted_at to 8 days ago so it falls within any reasonable retention window.
	uid, err := uuid.Parse(id)
	require.NoError(t, err)
	_, err = db.Exec(context.Background(),
		`UPDATE models SET deleted_at = NOW() - INTERVAL '8 days' WHERE id = $1`, uid)
	require.NoError(t, err)

	// Trigger purge with zero retention — everything soft-deleted qualifies.
	store.StartEviction(0, 50*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	store.StopEviction()

	// The model row should now be gone.
	var count int
	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM models WHERE id = $1`, uid).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "model row should have been hard-deleted by purge")
}
