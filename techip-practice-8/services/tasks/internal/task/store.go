package task

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Store struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]Task
}

func NewStore() *Store {
	return &Store{
		nextID: 1,
		items:  make(map[int64]Task),
	}
}

func (s *Store) Create(title string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	created, err := NewTask(s.nextID, title, now)
	if err != nil {
		return Task{}, err
	}

	s.items[created.ID] = created
	s.nextID++
	return created, nil
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Task, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func (s *Store) Get(id int64) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}
	return item, nil
}

func (s *Store) MarkDone(id int64) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}

	item.Done = true
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, nil
}

func (s *Store) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return ErrTaskNotFound
	}

	delete(s.items, id)
	return nil
}
