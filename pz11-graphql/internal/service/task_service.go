package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CyberGeo335/pz11-graphql/internal/domain"
	"github.com/CyberGeo335/pz11-graphql/internal/repository"
)

// TaskService contains business rules and hides storage details from GraphQL resolvers.
type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) ListTasks(ctx context.Context) ([]*domain.Task, error) {
	return s.repo.List(ctx)
}

func (s *TaskService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("task id is required")
	}
	return s.repo.Get(ctx, id)
}

func (s *TaskService) CreateTask(ctx context.Context, input domain.CreateTaskInput) (*domain.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("task title is required")
	}

	input.Title = title
	return s.repo.Create(ctx, input)
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, input domain.UpdateTaskInput) (*domain.Task, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("task id is required")
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, fmt.Errorf("task title cannot be empty")
		}
		input.Title = &title
	}

	return s.repo.Update(ctx, id, input)
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("task id is required")
	}
	return s.repo.Delete(ctx, id)
}
