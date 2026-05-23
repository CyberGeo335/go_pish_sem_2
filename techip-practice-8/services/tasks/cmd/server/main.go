package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/CyberGeo335/techip-practice-8/services/tasks/internal/task"
)

func main() {
	addr := ":8080"
	if value := os.Getenv("HTTP_ADDR"); value != "" {
		addr = value
	}

	store := task.NewStore()
	handler := task.NewHandler(store)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("tasks service started", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
