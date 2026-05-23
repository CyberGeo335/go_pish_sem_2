package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CyberGeo335/pz13-rabbit/internal/amqpclient"
	"github.com/CyberGeo335/pz13-rabbit/services/worker/internal/config"
	"github.com/CyberGeo335/pz13-rabbit/services/worker/internal/consumer"
)

func main() {
	logger := log.New(os.Stdout, "worker: ", log.LstdFlags|log.Lmicroseconds)
	cfg := config.Load()

	conn, err := amqpclient.Connect(cfg.RabbitURL)
	if err != nil {
		logger.Fatalf("rabbit connect error: %v", err)
	}
	defer conn.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := consumer.New(conn, cfg.QueueName, cfg.Prefetch, logger)
	if err := worker.Run(ctx); err != nil {
		logger.Fatalf("worker error: %v", err)
	}
}
