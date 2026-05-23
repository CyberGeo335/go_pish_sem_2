package domain

// Task is the internal business entity used by repository and service layers.
type Task struct {
	ID          string
	Title       string
	Description *string
	Done        bool
}

type CreateTaskInput struct {
	Title       string
	Description *string
}

type UpdateTaskInput struct {
	Title       *string
	Description *string
	Done        *bool
}
