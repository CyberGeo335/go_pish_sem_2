package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/CyberGeo335/pz6-web-security/internal/httpapi"
	"github.com/CyberGeo335/pz6-web-security/internal/store"
)

func main() {
	port := envOrDefault("PORT", "8080")
	cookieSecure := boolEnv("COOKIE_SECURE", false)

	st := store.New()

	handler, err := httpapi.NewHandler(st, cookieSecure)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/logout", handler.Logout)
	mux.HandleFunc("/profile", handler.Profile)
	mux.HandleFunc("/hello", handler.Hello)
	mux.HandleFunc("/comments", handler.Comments)

	addr := ":" + port
	log.Printf("server started on http://localhost:%s", port)
	log.Printf("open http://localhost:%s/login", port)
	log.Printf("cookie Secure attribute: %v", cookieSecure)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func boolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
