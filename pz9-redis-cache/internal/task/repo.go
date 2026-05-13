package task

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Repo struct {
	mu     sync.RWMutex
	data   map[int64]Task
	nextID int64
}

func NewRepo() *Repo {
	return &Repo{
		data: map[int64]Task{
			1: {
				ID:          1,
				Title:       "Изучить Redis",
				Description: "Разобрать стратегию cache-aside",
				DueDate:     time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			},
			2: {
				ID:          2,
				Title:       "Сделать ПЗ №9",
				Description: "Реализовать кэширование задачи по id",
				DueDate:     time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		nextID: 3,
	}
}

func (r *Repo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Task, 0, len(r.data))
	for _, item := range r.data {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items
}

func (r *Repo) GetByID(id int64) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.data[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}

	return t, nil
}

func (r *Repo) Create(t Task) Task {
	r.mu.Lock()
	defer r.mu.Unlock()

	t.ID = r.nextID
	r.nextID++
	r.data[t.ID] = t

	return t
}

func (r *Repo) Update(t Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[t.ID]; !ok {
		return ErrTaskNotFound
	}

	r.data[t.ID] = t
	return nil
}

func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[id]; !ok {
		return ErrTaskNotFound
	}

	delete(r.data, id)
	return nil
}
