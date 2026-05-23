package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const serviceName = "tasks"

type config struct {
	Port        string `json:"tasks_port"`
	AuthBaseURL string `json:"auth_base_url"`
	LogLevel    string `json:"log_level"`
}

type task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type taskStore struct {
	mu     sync.RWMutex
	nextID int
	items  map[int]task
}

func newTaskStore() *taskStore {
	now := time.Now().UTC()
	return &taskStore{
		nextID: 3,
		items: map[int]task{
			1: {ID: 1, Title: "Подготовить Docker-образ", Done: true, CreatedAt: now, UpdatedAt: now},
			2: {ID: 2, Title: "Опубликовать сервис в Kubernetes", Done: false, CreatedAt: now, UpdatedAt: now},
		},
	}
}

func (s *taskStore) list() []task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]task, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

func (s *taskStore) get(id int) (task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

func (s *taskStore) create(title string, done bool) task {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	item := task{
		ID:        s.nextID,
		Title:     title,
		Done:      done,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.items[item.ID] = item
	s.nextID++
	return item
}

func (s *taskStore) update(id int, title *string, done *bool) (task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return task{}, false
	}
	if title != nil {
		item.Title = strings.TrimSpace(*title)
	}
	if done != nil {
		item.Done = *done
	}
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, true
}

func (s *taskStore) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

type app struct {
	cfg   config
	store *taskStore
}

func main() {
	cfg := loadConfig()
	application := &app{
		cfg:   cfg,
		store: newTaskStore(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", application.handleRoot)
	mux.HandleFunc("/health", application.handleHealth)
	mux.HandleFunc("/tasks", application.handleTasks)
	mux.HandleFunc("/tasks/", application.handleTaskByID)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("service=%s port=%s auth_base_url=%s log_level=%s", serviceName, cfg.Port, cfg.AuthBaseURL, cfg.LogLevel)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}
}

func loadConfig() config {
	return config{
		Port:        getEnv("TASKS_PORT", "8082"),
		AuthBaseURL: getEnv("AUTH_BASE_URL", "http://auth:8081"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func (a *app) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service": serviceName,
		"message": "Tasks backend is running",
		"endpoints": []string{
			"GET /health",
			"GET /tasks",
			"POST /tasks",
			"GET /tasks/{id}",
			"PATCH /tasks/{id}",
			"DELETE /tasks/{id}",
		},
	})
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": serviceName,
		"time":    time.Now().UTC().Format(time.RFC3339),
		"config":  a.cfg,
	})
}

func (a *app) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": a.store.list()})
	case http.MethodPost:
		var request struct {
			Title string `json:"title"`
			Done  bool   `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		request.Title = strings.TrimSpace(request.Title)
		if request.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		writeJSON(w, http.StatusCreated, a.store.create(request.Title, request.Done))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *app) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := taskIDFromPath(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, ok := a.store.get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var request struct {
			Title *string `json:"title"`
			Done  *bool   `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if request.Title != nil && strings.TrimSpace(*request.Title) == "" {
			writeError(w, http.StatusBadRequest, "title must not be empty")
			return
		}
		item, ok := a.store.update(id, request.Title, request.Done)
		if !ok {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if !a.store.delete(id) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func taskIDFromPath(path string) (int, error) {
	idText := strings.TrimPrefix(path, "/tasks/")
	if idText == "" || strings.Contains(idText, "/") {
		return 0, fmt.Errorf("invalid task id")
	}
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid task id")
	}
	return id, nil
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(started))
	})
}
