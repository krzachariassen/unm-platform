package repository

import (
	"testing"
	"time"
)

func TestModelStore_StoreAndGet(t *testing.T) {
	s := NewModelStore()

	id, err := s.Store(nil)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	entry := s.Get(id)
	if entry == nil {
		t.Fatal("expected stored entry, got nil")
	}
	if entry.ID != id {
		t.Errorf("entry ID: want %q, got %q", id, entry.ID)
	}
}

func TestModelStore_GetMissing(t *testing.T) {
	s := NewModelStore()
	if entry := s.Get("nonexistent"); entry != nil {
		t.Errorf("expected nil for missing ID, got %+v", entry)
	}
}

func TestModelStore_UniqueIDs(t *testing.T) {
	s := NewModelStore()
	id1, _ := s.Store(nil)
	id2, _ := s.Store(nil)
	if id1 == id2 {
		t.Errorf("expected unique IDs, got duplicate %q", id1)
	}
}

func TestModelStore_Delete(t *testing.T) {
	s := NewModelStore()
	id, _ := s.Store(nil)

	if !s.Delete(id) {
		t.Error("Delete returned false for existing model")
	}
	if s.Get(id) != nil {
		t.Error("model still retrievable after Delete")
	}
	if s.Delete(id) {
		t.Error("Delete returned true for already-deleted model")
	}
}

func TestModelStore_DeleteCascade(t *testing.T) {
	s := NewModelStore()
	var cascadedID string
	s.SetOnDelete(func(modelID string) {
		cascadedID = modelID
	})

	id, _ := s.Store(nil)
	s.Delete(id)

	if cascadedID != id {
		t.Errorf("cascade callback: want %q, got %q", id, cascadedID)
	}
}

func TestModelStore_Len(t *testing.T) {
	s := NewModelStore()
	if s.Len() != 0 {
		t.Errorf("want 0, got %d", s.Len())
	}
	id1, _ := s.Store(nil)
	s.Store(nil)
	if s.Len() != 2 {
		t.Errorf("want 2, got %d", s.Len())
	}
	s.Delete(id1)
	if s.Len() != 1 {
		t.Errorf("want 1, got %d", s.Len())
	}
}

func TestModelStore_GetUpdatesLastAccessed(t *testing.T) {
	s := NewModelStore()
	id, _ := s.Store(nil)

	before := s.Get(id).LastAccessedAt
	time.Sleep(5 * time.Millisecond)
	after := s.Get(id).LastAccessedAt

	if !after.After(before) {
		t.Error("LastAccessedAt not updated on Get")
	}
}

func TestModelStore_EvictExpired(t *testing.T) {
	s := NewModelStore()
	var cascaded []string
	s.SetOnDelete(func(id string) { cascaded = append(cascaded, id) })

	old, _ := s.Store(nil)
	s.mu.Lock()
	s.models[old].LastAccessedAt = time.Now().Add(-3 * time.Hour)
	s.mu.Unlock()

	fresh, _ := s.Store(nil)

	s.evictExpired(2 * time.Hour)

	if s.Get(old) != nil {
		t.Error("expired model should be evicted")
	}
	if s.Get(fresh) == nil {
		t.Error("fresh model should NOT be evicted")
	}
	if len(cascaded) != 1 || cascaded[0] != old {
		t.Errorf("cascade: want [%s], got %v", old, cascaded)
	}
}

func TestChangesetStore_DeleteForModel(t *testing.T) {
	cs := NewChangesetStore()
	cs.Store("model-a", nil)
	cs.Store("model-a", nil)
	cs.Store("model-b", nil)

	n := cs.DeleteForModel("model-a")
	if n != 2 {
		t.Errorf("want 2 deleted, got %d", n)
	}
	if len(cs.ListForModel("model-a")) != 0 {
		t.Error("model-a changesets still present")
	}
	if len(cs.ListForModel("model-b")) != 1 {
		t.Error("model-b changeset should be unaffected")
	}
}
