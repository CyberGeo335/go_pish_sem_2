package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/CyberGeo335/pz3-logging/internal/student"
	"go.uber.org/zap"
)

type Handler struct {
	repo *student.Repo
	log  *zap.Logger
}

func NewHandler(repo *student.Repo, log *zap.Logger) *Handler {
	return &Handler{
		repo: repo,
		log:  log,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Warn("method not allowed for health endpoint",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h.writePlainError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.log.Debug("health endpoint called")

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Warn("method not allowed for student endpoint",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h.writePlainError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/students/")
	if path == "" || path == r.URL.Path {
		h.log.Warn("student id is missing",
			zap.String("path", r.URL.Path),
		)
		h.writePlainError(w, "student id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		h.log.Warn("invalid student id",
			zap.String("raw_id", path),
			zap.Error(err),
		)
		h.writePlainError(w, "invalid student id", http.StatusBadRequest)
		return
	}

	h.log.Debug("started student lookup",
		zap.Int64("student_id", id),
	)

	st, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, student.ErrStudentNotFound) {
			h.log.Error("student not found",
				zap.Int64("student_id", id),
				zap.Error(err),
			)
			h.writePlainError(w, "student not found", http.StatusNotFound)
			return
		}

		h.log.Error("failed to get student",
			zap.Int64("student_id", id),
			zap.Error(err),
		)
		h.writePlainError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.log.Info("student returned successfully",
		zap.Int64("student_id", st.ID),
		zap.String("group", st.Group),
	)

	h.writeJSON(w, http.StatusOK, st)
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.log.Error("failed to write json response", zap.Error(err))
	}
}

func (h *Handler) writePlainError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
