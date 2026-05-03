package main

import (
	"log"
	"net/http"

	"github.com/CyberGeo335/pz3-logging/internal/httpapi"
	"github.com/CyberGeo335/pz3-logging/internal/student"
	applogger "github.com/CyberGeo335/pz3-logging/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	logger, err := applogger.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	repo := student.NewRepo()
	handler := httpapi.NewHandler(repo, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/students/", handler.GetStudentByID)

	rootHandler := httpapi.LoggingMiddleware(logger, mux)

	addr := ":8080"
	logger.Info("server is starting",
		zap.String("addr", addr),
	)

	if err := http.ListenAndServe(addr, rootHandler); err != nil {
		logger.Fatal("server failed",
			zap.Error(err),
		)
	}
}
