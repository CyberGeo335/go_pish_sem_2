package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/CyberGeo335/pz5-security/internal/config"
	"github.com/CyberGeo335/pz5-security/internal/httpapi"
	"github.com/CyberGeo335/pz5-security/internal/student"
)

func main() {
	cfg := config.New()

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	repo := student.NewRepo(db)

	getByIDStmt, err := repo.PrepareGetByID()
	if err != nil {
		log.Fatal(err)
	}
	defer getByIDStmt.Close()

	getByEmailStmt, err := repo.PrepareGetByEmail()
	if err != nil {
		log.Fatal(err)
	}
	defer getByEmailStmt.Close()

	handler := httpapi.NewHandler(getByIDStmt, getByEmailStmt)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/students", handler.GetStudentByID)
	mux.HandleFunc("/students/by-email", handler.GetStudentByEmail)

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	httpsServer := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	redirectServer := &http.Server{
		Addr:              cfg.HTTPRedirectAddr,
		Handler:           redirectToHTTPS("localhost" + cfg.Addr),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("HTTP redirect server started on http://localhost%s", cfg.HTTPRedirectAddr)
		if err := redirectServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Printf("HTTPS server started on https://localhost%s", cfg.Addr)
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	waitForShutdown(httpsServer, redirectServer)
}

func redirectToHTTPS(host string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := "https://" + host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}

func waitForShutdown(servers ...*http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, server := range servers {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}
}
