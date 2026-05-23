package task

import "testing"

func TestRepositoryCreateGetUpdate(t *testing.T) {
	repo := NewRepository()

	created, err := repo.Create(CreateInput{Title: "Новая задача", Description: "Описание"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if created.ID != "t_003" {
		t.Fatalf("expected id t_003, got %s", created.ID)
	}
	if created.Done {
		t.Fatal("new task must be created with done=false")
	}

	done := true
	updated, err := repo.Update(created.ID, UpdateInput{Done: &done})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if !updated.Done {
		t.Fatal("expected updated task done=true")
	}

	got, err := repo.Get(created.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.ID != created.ID || !got.Done {
		t.Fatalf("unexpected task after update: %#v", got)
	}
}

func TestRepositoryRejectsEmptyTitle(t *testing.T) {
	repo := NewRepository()
	if _, err := repo.Create(CreateInput{Title: "   "}); err == nil {
		t.Fatal("expected error for empty title")
	}
}
