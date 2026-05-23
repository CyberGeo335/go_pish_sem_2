package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CyberGeo335/pz13-rabbit/internal/amqpclient"
	"github.com/CyberGeo335/pz13-rabbit/services/tasks/internal/config"
	"github.com/CyberGeo335/pz13-rabbit/services/tasks/internal/httpapi"
	"github.com/CyberGeo335/pz13-rabbit/services/tasks/internal/publisher"
	"github.com/CyberGeo335/pz13-rabbit/services/tasks/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "tasks: ", log.LstdFlags|log.Lmicroseconds)
	cfg := config.Load()

	var taskPublisher *publisher.Publisher
	conn, err := amqpclient.Connect(cfg.RabbitURL)
	if err != nil {
		logger.Printf("rabbit connect failed, service will run in best-effort mode without publisher: %v", err)
	} else {
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			logger.Printf("rabbit channel failed, service will run in best-effort mode without publisher: %v", err)
		} else {
			defer ch.Close()
			taskPublisher = publisher.New(ch, cfg.QueueName)
			logger.Printf("rabbit connected: queue=%s", cfg.QueueName)
		}
	}

	taskService := service.NewTaskService()
	api := httpapi.NewServer(taskService, taskPublisher, logger)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Printf("HTTP server started: addr=%s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Printf("HTTP shutdown error: %v", err)
	}
	logger.Println("service stopped")
}
