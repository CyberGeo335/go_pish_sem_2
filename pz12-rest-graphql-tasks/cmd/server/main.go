package main

import (
	"log"
	"net/http"
	"os"
	"time"

	graphqlhandler "github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/graphql"
	"github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/rest"
	"github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/task"
)

func main() {
	addr := ":8082"
	if fromEnv := os.Getenv("ADDR"); fromEnv != "" {
		addr = fromEnv
	}

	repo := task.NewRepository()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
	})

	rest.NewHandler(repo).Register(mux)
	graphqlhandler.NewHandler(repo).Register(mux)

	server := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("PZ12 server started on http://localhost%s", addr)
	log.Printf("REST:    http://localhost%s/v1/tasks", addr)
	log.Printf("GraphQL: http://localhost%s/query", addr)
	log.Printf("UI:      http://localhost%s/graphql", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
