package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/CyberGeo335/pz14-task-queue/internal/rabbitmq"
	"github.com/CyberGeo335/pz14-task-queue/services/worker/internal/consumer"
	"github.com/CyberGeo335/pz14-task-queue/services/worker/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, ch, err := rabbitmq.Dial(rabbitmq.URL())
	if err != nil {
		log.Fatalf("rabbitmq dial failed: %v", err)
	}
	defer func() { _ = ch.Close() }()
	defer func() { _ = conn.Close() }()

	if err := rabbitmq.DeclareQueues(ch); err != nil {
		log.Fatalf("queue declaration failed: %v", err)
	}

	processed := store.NewProcessedStore()
	worker := consumer.New(ch, processed)

	if err := worker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("worker stopped with error: %v", err)
	}

	log.Println("worker stopped")
}
