package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CyberGeo335/tasks/internal/config"
)

// Task is a simple in-memory task model used for demonstration and health checks.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type taskStore struct {
	mu     sync.RWMutex
	nextID int
	tasks  map[int]Task
}

func newTaskStore() *taskStore {
	return &taskStore{
		nextID: 1,
		tasks:  make(map[int]Task),
	}
}

func (s *taskStore) create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	task := Task{
		ID:        s.nextID,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.tasks[task.ID] = task
	s.nextID++

	return task
}

func (s *taskStore) list() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

func (s *taskStore) get(id int) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}

func (s *taskStore) update(id int, title *string, done *bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}

	if title != nil {
		task.Title = strings.TrimSpace(*title)
	}

	if done != nil {
		task.Done = *done
	}

	task.UpdatedAt = time.Now().UTC()
	s.tasks[id] = task

	return task, true
}

func (s *taskStore) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}

	delete(s.tasks, id)
	return true
}

// NewRouter returns all HTTP routes for the tasks service.
func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	store := newTaskStore()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "tasks",
		})
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.list())
	})

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}

		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		req.Title = strings.TrimSpace(req.Title)
		if req.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}

		writeJSON(w, http.StatusCreated, store.create(req.Title))
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := taskIDFromRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		task, ok := store.get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}

		writeJSON(w, http.StatusOK, task)
	})

	mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := taskIDFromRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var req struct {
			Title *string `json:"title"`
			Done  *bool   `json:"done"`
		}

		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
			writeError(w, http.StatusBadRequest, "title must not be empty")
			return
		}

		task, ok := store.update(id, req.Title, req.Done)
		if !ok {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}

		writeJSON(w, http.StatusOK, task)
	})

	mux.HandleFunc("DELETE /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := taskIDFromRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		if !store.delete(id) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return loggingMiddleware(cfg, mux)
}

func loggingMiddleware(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)

		if cfg.LogLevel != "silent" {
			log.Printf("method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(started))
		}
	})
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func taskIDFromRequest(r *http.Request) (int, error) {
	rawID := r.PathValue("id")
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}
