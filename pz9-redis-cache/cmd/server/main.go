package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/CyberGeo335/pz9-redis-cache/internal/cache"
	"github.com/CyberGeo335/pz9-redis-cache/internal/config"
	"github.com/CyberGeo335/pz9-redis-cache/internal/httpapi"
	"github.com/CyberGeo335/pz9-redis-cache/internal/service"
	"github.com/CyberGeo335/pz9-redis-cache/internal/task"
)

func main() {
	cfg := config.New()

	repo := task.NewRepo()
	redisClient := cache.NewRedisClient(cfg)
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Println("redis close error:", err)
		}
	}()

	if err := cache.Ping(context.Background(), redisClient); err != nil {
		log.Println("warning: redis is unavailable at startup:", err)
	} else {
		log.Println("redis connected:", cfg.RedisAddr)
	}

	taskService := service.NewTaskService(repo, redisClient, cfg)
	handler := httpapi.NewHandler(taskService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/v1/tasks", handler.Tasks)
	mux.HandleFunc("/v1/tasks/", handler.TaskByID)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("server started on", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
