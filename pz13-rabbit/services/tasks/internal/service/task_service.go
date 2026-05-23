package service

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TaskService struct {
	mu      sync.RWMutex
	nextID  int
	storage map[string]Task
}

func NewTaskService() *TaskService {
	return &TaskService{
		nextID:  1,
		storage: make(map[string]Task),
	}
}

func (s *TaskService) Create(req CreateTaskRequest) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("t_%03d", s.nextID)
	s.nextID++

	task := Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   time.Now().UTC(),
	}
	s.storage[id] = task

	return task
}

func (s *TaskService) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.storage))
	for _, task := range s.storage {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks
}
