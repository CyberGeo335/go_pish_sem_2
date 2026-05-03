package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/CyberGeo335/pz5-security/internal/student"
)

var emailAllowList = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Handler struct {
	getByIDStmt    *sql.Stmt
	getByEmailStmt *sql.Stmt
}

func NewHandler(getByIDStmt, getByEmailStmt *sql.Stmt) *Handler {
	return &Handler{
		getByIDStmt:    getByIDStmt,
		getByEmailStmt: getByEmailStmt,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"scheme": "https",
	})
}

func (h *Handler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	rawID := r.URL.Query().Get("id")
	if rawID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var st student.Student
	err = h.getByIDStmt.QueryRow(id).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		handleStudentError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, st)
}

func (h *Handler) GetStudentByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	if !emailAllowList.MatchString(email) {
		writeError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	var st student.Student
	err := h.getByEmailStmt.QueryRow(email).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		handleStudentError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, st)
}

func handleStudentError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, student.ErrStudentNotFound) {
		writeError(w, http.StatusNotFound, "student not found")
		return
	}

	writeError(w, http.StatusInternalServerError, "internal server error")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
