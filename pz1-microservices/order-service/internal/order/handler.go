package order

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler contains HTTP handlers for order-service.
type Handler struct {
	repo   *Repo
	client *UserServiceClient
}

func NewHandler(repo *Repo, client *UserServiceClient) *Handler {
	return &Handler{
		repo:   repo,
		client: client,
	}
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Expected path: /orders/{id}
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "order id is required", http.StatusBadRequest)
		return
	}

	if strings.Contains(path, "/") {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) GetOrderWithUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Expected path: /orders/{id}/full
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	if !strings.HasSuffix(path, "/full") {
		http.Error(w, "bad route", http.StatusBadRequest)
		return
	}

	idPart := strings.TrimSuffix(path, "/full")
	idPart = strings.TrimSuffix(idPart, "/")
	if idPart == "" || strings.Contains(idPart, "/") {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.client.GetUserByID(r.Context(), order.UserID)
	if err != nil {
		http.Error(w, "failed to get user data: "+err.Error(), http.StatusBadGateway)
		return
	}

	result := OrderWithUser{
		Order: order,
		User:  user,
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Expected path: /orders/by-user/{userID}
	path := strings.TrimPrefix(r.URL.Path, "/orders/by-user/")
	if path == "" || path == r.URL.Path || strings.Contains(path, "/") {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(path, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	orders := h.repo.GetByUserID(userID)
	writeJSON(w, http.StatusOK, orders)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
