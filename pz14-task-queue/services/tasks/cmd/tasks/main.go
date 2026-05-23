package main

import (
	"log"
	"net/http"
	"os"

	"github.com/CyberGeo335/pz14-task-queue/internal/rabbitmq"
	httpapi "github.com/CyberGeo335/pz14-task-queue/services/tasks/internal/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	rabbitURL := rabbitmq.URL()
	conn, ch, err := rabbitmq.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq dial failed: %v", err)
	}
	defer func() { _ = ch.Close() }()
	defer func() { _ = conn.Close() }()

	if err := rabbitmq.DeclareQueues(ch); err != nil {
		log.Fatalf("queue declaration failed: %v", err)
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8082"
	}

	handler := httpapi.NewHandler(ch)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", httpapi.Health)
	mux.HandleFunc("/v1/jobs/process-task", handler.ProcessTask)

	log.Printf("tasks service listening on %s", addr)
	log.Printf("POST /v1/jobs/process-task publishes messages to %s", rabbitmq.TaskJobsQueue)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("http server stopped: %v", err)
	}
}
