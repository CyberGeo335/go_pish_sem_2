package main

import (
	"log"
	"net/http"
	"os"

	"github.com/CyberGeo335/pz4-monitoring/internal/httpapi"
	"github.com/CyberGeo335/pz4-monitoring/internal/student"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	repo := student.NewRepo()
	handler := httpapi.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/students/", handler.GetStudentByID)
	mux.Handle("/metrics", promhttp.Handler())

	rootHandler := httpapi.MetricsMiddleware(mux)

	addr := ":" + port
	log.Printf("server started on %s", addr)

	if err := http.ListenAndServe(addr, rootHandler); err != nil {
		log.Fatal(err)
	}
}
