package task

// Task is the single domain model used by both REST and GraphQL handlers.
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// CreateInput describes fields required to create a task.
type CreateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateInput describes optional fields for partial task updates.
type UpdateInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}
