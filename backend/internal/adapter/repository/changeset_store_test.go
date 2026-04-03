package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func TestChangesetStore_Store_ReturnsID(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-1", "test changeset")
	require.NoError(t, err)

	id, err := s.Store("model-1", cs)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestChangesetStore_Get_ReturnsStoredChangeset(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-get", "get test")
	require.NoError(t, err)

	id, err := s.Store("model-1", cs)
	require.NoError(t, err)

	got, err := s.Get(id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "model-1", got.ModelID)
	assert.Equal(t, "get test", got.Changeset.Description)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestChangesetStore_Get_NonExistent_ReturnsErrNotFound(t *testing.T) {
	s := NewChangesetStore()

	_, err := s.Get("nonexistent-id")
	assert.True(t, errors.Is(err, usecase.ErrNotFound))
}

func TestChangesetStore_ListForModel_ReturnsMatchingChangesets(t *testing.T) {
	s := NewChangesetStore()

	cs1, err := entity.NewChangeset("cs-a", "first")
	require.NoError(t, err)
	cs2, err := entity.NewChangeset("cs-b", "second")
	require.NoError(t, err)
	cs3, err := entity.NewChangeset("cs-c", "other model")
	require.NoError(t, err)

	_, err = s.Store("model-1", cs1)
	require.NoError(t, err)
	_, err = s.Store("model-1", cs2)
	require.NoError(t, err)
	_, err = s.Store("model-2", cs3)
	require.NoError(t, err)

	results, err := s.ListForModel("model-1")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	for _, sc := range results {
		assert.Equal(t, "model-1", sc.ModelID)
	}
}

func TestChangesetStore_ListForModel_EmptyForUnknownModel(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)

	_, err = s.Store("model-1", cs)
	require.NoError(t, err)

	results, err := s.ListForModel("nonexistent-model")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestChangesetStore_DeleteForModel_RemovesMatchingChangesets(t *testing.T) {
	s := NewChangesetStore()

	cs1, err := entity.NewChangeset("cs-a", "first")
	require.NoError(t, err)
	cs2, err := entity.NewChangeset("cs-b", "second")
	require.NoError(t, err)
	cs3, err := entity.NewChangeset("cs-c", "other")
	require.NoError(t, err)

	_, err = s.Store("model-1", cs1)
	require.NoError(t, err)
	_, err = s.Store("model-1", cs2)
	require.NoError(t, err)
	_, err = s.Store("model-2", cs3)
	require.NoError(t, err)

	count := s.DeleteForModel("model-1")
	assert.Equal(t, 2, count)

	// model-1 changesets should be gone
	results, _ := s.ListForModel("model-1")
	assert.Empty(t, results)

	// model-2 changeset should still exist
	results2, _ := s.ListForModel("model-2")
	assert.Len(t, results2, 1)
}

func TestChangesetStore_DeleteForModel_NoMatchReturnsZero(t *testing.T) {
	s := NewChangesetStore()

	count := s.DeleteForModel("nonexistent-model")
	assert.Equal(t, 0, count)
}

func TestChangesetStore_Store_MultipleChangesets_UniqueIDs(t *testing.T) {
	s := NewChangesetStore()

	ids := make(map[string]bool)
	for i := 0; i < 5; i++ {
		cs, err := entity.NewChangeset("cs", "test")
		require.NoError(t, err)
		id, err := s.Store("model-1", cs)
		require.NoError(t, err)
		assert.False(t, ids[id], "generated ID %q should be unique", id)
		ids[id] = true
	}
}

func TestChangesetStore_Get_AfterDelete_ReturnsErrNotFound(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-1", "to be deleted")
	require.NoError(t, err)

	id, err := s.Store("model-del", cs)
	require.NoError(t, err)

	// Confirm it exists first
	got, err := s.Get(id)
	require.NoError(t, err)
	require.NotNil(t, got)

	// Delete the model's changesets
	s.DeleteForModel("model-del")

	// Should return ErrNotFound now
	_, err = s.Get(id)
	assert.True(t, errors.Is(err, usecase.ErrNotFound))
}

func TestChangesetStore_Update(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-1", "original")
	require.NoError(t, err)

	id, err := s.Store("model-1", cs)
	require.NoError(t, err)

	updated, err := entity.NewChangeset("cs-1", "updated description")
	require.NoError(t, err)

	err = s.Update(id, updated)
	require.NoError(t, err)

	got, err := s.Get(id)
	require.NoError(t, err)
	assert.Equal(t, "updated description", got.Changeset.Description)
}

func TestChangesetStore_Update_NonExistent_ReturnsErrNotFound(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-1", "test")
	require.NoError(t, err)

	err = s.Update("nonexistent-id", cs)
	assert.True(t, errors.Is(err, usecase.ErrNotFound))
}

func TestChangesetStore_Delete_SingleID(t *testing.T) {
	s := NewChangesetStore()
	cs, err := entity.NewChangeset("cs-1", "to delete")
	require.NoError(t, err)

	id, err := s.Store("model-1", cs)
	require.NoError(t, err)

	err = s.Delete(id)
	require.NoError(t, err)

	_, err = s.Get(id)
	assert.True(t, errors.Is(err, usecase.ErrNotFound))
}
