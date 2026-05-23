package task

import (
	"errors"
	"testing"
)

func TestStoreCreateAndList(t *testing.T) {
	store := NewStore()

	first, err := store.Create("  write pipeline  ")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	second, err := store.Create("run tests")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("unexpected IDs: got %d and %d", first.ID, second.ID)
	}
	if first.Title != "write pipeline" {
		t.Fatalf("title was not trimmed: %q", first.Title)
	}

	items := store.List()
	if len(items) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(items))
	}
	if items[0].ID != 1 || items[1].ID != 2 {
		t.Fatalf("tasks are not sorted by ID: %+v", items)
	}
}

func TestStoreRejectsEmptyTitle(t *testing.T) {
	store := NewStore()

	_, err := store.Create("   ")
	if !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestStoreMarkDoneAndDelete(t *testing.T) {
	store := NewStore()
	created, err := store.Create("docker build")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	updated, err := store.MarkDone(created.ID)
	if err != nil {
		t.Fatalf("MarkDone returned error: %v", err)
	}
	if !updated.Done {
		t.Fatal("task was not marked as done")
	}

	if err := store.Delete(created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	_, err = store.Get(created.ID)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound after delete, got %v", err)
	}
}
