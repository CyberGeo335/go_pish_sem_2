package graph

import (
	"context"

	"github.com/CyberGeo335/pz11-graphql/graph/generated"
	"github.com/CyberGeo335/pz11-graphql/graph/model"
	"github.com/CyberGeo335/pz11-graphql/internal/domain"
)

// CreateTask is the resolver for the createTask field.
func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	task, err := r.TaskService.CreateTask(ctx, domain.CreateTaskInput{
		Title:       input.Title,
		Description: input.Description,
	})
	if err != nil {
		return nil, err
	}
	return toModelTask(task), nil
}

// UpdateTask is the resolver for the updateTask field.
func (r *mutationResolver) UpdateTask(ctx context.Context, id string, input model.UpdateTaskInput) (*model.Task, error) {
	task, err := r.TaskService.UpdateTask(ctx, id, domain.UpdateTaskInput{
		Title:       input.Title,
		Description: input.Description,
		Done:        input.Done,
	})
	if err != nil {
		return nil, err
	}
	return toModelTask(task), nil
}

// DeleteTask is the resolver for the deleteTask field.
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (bool, error) {
	return r.TaskService.DeleteTask(ctx, id)
}

// Tasks is the resolver for the tasks field.
func (r *queryResolver) Tasks(ctx context.Context) ([]*model.Task, error) {
	tasks, err := r.TaskService.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Task, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, toModelTask(task))
	}

	return result, nil
}

// Task is the resolver for the task field.
func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	task, err := r.TaskService.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return toModelTask(task), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

func toModelTask(task *domain.Task) *model.Task {
	if task == nil {
		return nil
	}
	return &model.Task{
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
