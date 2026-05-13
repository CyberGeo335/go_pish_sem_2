package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type Server struct {
	instanceID string
	tasks      []Task
}

func main() {
	instanceID := getenv("INSTANCE_ID", "tasks-unknown")
	port := getenv("APP_PORT", "8082")

	srv := &Server{
		instanceID: instanceID,
		tasks: []Task{
			{ID: 1, Title: "Изучить NGINX"},
			{ID: 2, Title: "Понять load balancing"},
			{ID: 3, Title: "Проверить горизонтальное масштабирование"},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", srv.handleHealth)
	mux.HandleFunc("GET /v1/tasks", srv.handleListTasks)
	mux.HandleFunc("GET /whoami", srv.handleWhoami)

	wrappedMux := requestLogger(instanceID, mux)
	addr := ":" + port

	log.Printf("tasks service started on %s, instance=%s", addr, instanceID)
	if err := http.ListenAndServe(addr, wrappedMux); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.instanceID, map[string]string{
		"status":   "ok",
		"instance": s.instanceID,
	})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.instanceID, s.tasks)
}

func (s *Server) handleWhoami(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.instanceID, map[string]string{
		"instance": s.instanceID,
	})
}

func writeJSON(w http.ResponseWriter, status int, instanceID string, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Instance-ID", instanceID)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func requestLogger(instanceID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("instance=%s method=%s path=%s duration=%s", instanceID, r.Method, r.URL.Path, time.Since(started))
	})
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
