package task

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	// ErrNotFound is returned when a task ID does not exist.
	ErrNotFound = errors.New("task not found")

	// ErrInvalidInput is returned when input data cannot be accepted.
	ErrInvalidInput = errors.New("invalid task input")
)

// Repository stores tasks in memory. It is shared by REST and GraphQL handlers.
type Repository struct {
	mu     sync.RWMutex
	nextID int
	items  map[string]Task
}

// NewRepository creates a repository with identical seed data for both APIs.
func NewRepository() *Repository {
	r := &Repository{
		nextID: 3,
		items: map[string]Task{
			"t_001": {
				ID:          "t_001",
				Title:       "Первая задача",
				Description: "Учебный пример",
				Done:        false,
			},
			"t_002": {
				ID:          "t_002",
				Title:       "Вторая задача",
				Description: "Проверка API",
				Done:        true,
			},
		},
	}
	return r
}

// List returns all tasks sorted by ID for deterministic output.
func (r *Repository) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.items))
	for id := range r.items {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]Task, 0, len(ids))
	for _, id := range ids {
		result = append(result, r.items[id])
	}
	return result
}

// Get returns one task by ID.
func (r *Repository) Get(id string) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return item, nil
}

// Create stores a new task. New tasks are created with Done=false.
func (r *Repository) Create(input CreateInput) (Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Task{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("t_%03d", r.nextID)
	r.nextID++

	item := Task{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Done:        false,
	}
	r.items[id] = item
	return item, nil
}

// Update partially updates an existing task.
func (r *Repository) Update(id string, input UpdateInput) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return Task{}, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
		}
		item.Title = title
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.Done != nil {
		item.Done = *input.Done
	}

	r.items[id] = item
	return item, nil
}
