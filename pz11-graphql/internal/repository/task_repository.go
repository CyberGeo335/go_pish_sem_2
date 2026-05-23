package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/CyberGeo335/pz11-graphql/internal/domain"
)

// TaskRepository describes the storage contract used by the service layer.
type TaskRepository interface {
	List(ctx context.Context) ([]*domain.Task, error)
	Get(ctx context.Context, id string) (*domain.Task, error)
	Create(ctx context.Context, input domain.CreateTaskInput) (*domain.Task, error)
	Update(ctx context.Context, id string, input domain.UpdateTaskInput) (*domain.Task, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// MemoryTaskRepository is a thread-safe in-memory storage for the practical work.
type MemoryTaskRepository struct {
	mu     sync.RWMutex
	tasks  map[string]*domain.Task
	nextID int
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	repo := &MemoryTaskRepository{
		tasks:  make(map[string]*domain.Task),
		nextID: 3,
	}

	desc1 := "Учебный пример"
	desc2 := "GraphQL API"
	repo.tasks["t_001"] = &domain.Task{ID: "t_001", Title: "Первая задача", Description: &desc1, Done: false}
	repo.tasks["t_002"] = &domain.Task{ID: "t_002", Title: "Вторая задача", Description: &desc2, Done: true}

	return repo
}

func (r *MemoryTaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.tasks))
	for id := range r.tasks {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]*domain.Task, 0, len(ids))
	for _, id := range ids {
		result = append(result, cloneTask(r.tasks[id]))
	}

	return result, nil
}

func (r *MemoryTaskRepository) Get(ctx context.Context, id string) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, nil
	}

	return cloneTask(task), nil
}

func (r *MemoryTaskRepository) Create(ctx context.Context, input domain.CreateTaskInput) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("t_%03d", r.nextID)
	r.nextID++

	task := &domain.Task{
		ID:          id,
		Title:       input.Title,
		Description: cloneStringPtr(input.Description),
		Done:        false,
	}
	r.tasks[id] = task

	return cloneTask(task), nil
}

func (r *MemoryTaskRepository) Update(ctx context.Context, id string, input domain.UpdateTaskInput) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task with id %q not found", id)
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = cloneStringPtr(input.Description)
	}
	if input.Done != nil {
		task.Done = *input.Done
	}

	return cloneTask(task), nil
}

func (r *MemoryTaskRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return false, nil
	}

	delete(r.tasks, id)
	return true, nil
}

func cloneTask(task *domain.Task) *domain.Task {
	if task == nil {
		return nil
	}
	return &domain.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: cloneStringPtr(task.Description),
		Done:        task.Done,
	}
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
