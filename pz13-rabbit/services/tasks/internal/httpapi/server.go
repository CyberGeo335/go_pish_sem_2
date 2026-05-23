package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CyberGeo335/pz13-rabbit/services/tasks/internal/publisher"
	"github.com/CyberGeo335/pz13-rabbit/services/tasks/internal/service"
)

type Server struct {
	tasks     *service.TaskService
	publisher *publisher.Publisher
	logger    *log.Logger
}

func NewServer(tasks *service.TaskService, publisher *publisher.Publisher, logger *log.Logger) *Server {
	return &Server{tasks: tasks, publisher: publisher, logger: logger}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /v1/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /v1/tasks", s.handleListTasks)
	return logRequests(s.logger, mux)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req service.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	req.Description = strings.TrimSpace(req.Description)
	requestID := requestIDFrom(r)

	task := s.tasks.Create(req)

	publishStatus := "published"
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.publisher.PublishTaskCreated(ctx, task.ID, requestID); err != nil {
		publishStatus = "failed"
		s.logger.Printf("publish task.created failed: task_id=%s request_id=%s error=%v", task.ID, requestID, err)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"task":                 task,
		"event_publish_status": publishStatus,
		"request_id":           requestID,
	})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"tasks": s.tasks.List()})
}

func requestIDFrom(r *http.Request) string {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID != "" {
		return requestID
	}
	return "req_" + time.Now().UTC().Format("20060102150405.000000000")
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func logRequests(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("%s %s completed in %s", r.Method, r.URL.Path, time.Since(startedAt).Round(time.Millisecond))
	})
}
