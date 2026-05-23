package task

import (
	"errors"
	"strings"
	"time"
)

var ErrEmptyTitle = errors.New("task title must not be empty")

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
}

func NewTask(id int64, title string, now time.Time) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	return Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}, nil
}
