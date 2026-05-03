package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CyberGeo335/pz4-monitoring/internal/metrics"
	"github.com/CyberGeo335/pz4-monitoring/internal/student"
)

type Handler struct {
	repo *student.Repo
}

func NewHandler(repo *student.Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/students/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "student id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}

	studentID := strconv.FormatInt(id, 10)
	metrics.StudentRequestsTotal.WithLabelValues(studentID).Inc()

	start := time.Now()
	defer func() {
		metrics.StudentRequestDuration.WithLabelValues(studentID).Observe(time.Since(start).Seconds())
	}()

	st, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, student.ErrStudentNotFound) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, st)
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}
