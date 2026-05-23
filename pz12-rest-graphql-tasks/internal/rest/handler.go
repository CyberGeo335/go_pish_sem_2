package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/task"
)

// Handler exposes the Task domain through REST endpoints.
type Handler struct {
	repo *task.Repository
}

// NewHandler creates a REST handler.
func NewHandler(repo *task.Repository) *Handler {
	return &Handler{repo: repo}
}

// Register attaches REST routes to a mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/tasks", h.tasks)
	mux.HandleFunc("/v1/tasks/", h.taskByID)
}

func (h *Handler) tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.repo.List())
	case http.MethodPost:
		var input task.CreateInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		item, err := h.repo.Create(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) taskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := h.repo.Get(id)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, task.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var input task.UpdateInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		item, err := h.repo.Update(id, input)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, task.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
