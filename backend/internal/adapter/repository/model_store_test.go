package repository

import (
	"errors"
	"testing"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
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

	entry, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry == nil {
		t.Fatal("expected stored entry, got nil")
	}
	if entry.ID != id {
		t.Errorf("entry ID: want %q, got %q", id, entry.ID)
	}
}

func TestModelStore_GetMissing(t *testing.T) {
	s := NewModelStore()
	_, err := s.Get("nonexistent")
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing ID, got %v", err)
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

	if err := s.Delete(id); err != nil {
		t.Errorf("Delete returned error for existing model: %v", err)
	}
	if _, err := s.Get(id); !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("model still retrievable after Delete: %v", err)
	}
	// Delete is idempotent — second delete should not error
	if err := s.Delete(id); err != nil {
		t.Errorf("Delete returned error for already-deleted model: %v", err)
	}
}

func TestModelStore_Replace(t *testing.T) {
	s := NewModelStore()
	m := &entity.UNMModel{}
	id, err := s.Store(m)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	newModel := &entity.UNMModel{}
	if err := s.Replace(id, newModel); err != nil {
		t.Errorf("Replace returned error for existing model: %v", err)
	}

	if err := s.Replace("nonexistent", newModel); !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("Replace should return ErrNotFound for missing ID, got %v", err)
	}
}

func TestModelStore_List(t *testing.T) {
	s := NewModelStore()
	_, _ = s.Store(nil)
	_, _ = s.Store(nil)

	items, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}
