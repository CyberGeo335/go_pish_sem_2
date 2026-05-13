package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/CyberGeo335/pz9-redis-cache/internal/service"
	"github.com/CyberGeo335/pz9-redis-cache/internal/task"
)

type Handler struct {
	service *service.TaskService
}

func NewHandler(service *service.TaskService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListTasks(w, r)
	case http.MethodPost:
		h.CreateTask(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) TaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetTaskByID(w, r, id)
	case http.MethodPatch:
		h.PatchTask(w, r, id)
	case http.MethodDelete:
		h.DeleteTask(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) ListTasks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": h.service.ListTasks()})
}

func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request, id int64) {
	t, err := h.service.GetTaskByID(r.Context(), id)
	if errors.Is(err, task.ErrTaskNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t task.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	created := h.service.CreateTask(r.Context(), t)
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request, id int64) {
	var t task.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	t.ID = id
	if err := h.service.UpdateTask(r.Context(), t); errors.Is(err, task.ErrTaskNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.DeleteTask(r.Context(), id); errors.Is(err, task.ErrTaskNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func idFromPath(path string) (int64, error) {
	rawID := strings.TrimPrefix(path, "/v1/tasks/")
	if rawID == "" || strings.Contains(rawID, "/") {
		return 0, strconv.ErrSyntax
	}

	return strconv.ParseInt(rawID, 10, 64)
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}
