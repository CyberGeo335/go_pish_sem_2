package config

import "os"

type Config struct {
	Addr             string
	HTTPRedirectAddr string
	CertFile         string
	KeyFile          string
	DSN              string
}

func New() Config {
	return Config{
		Addr:             getenv("ADDR", ":8443"),
		HTTPRedirectAddr: getenv("HTTP_REDIRECT_ADDR", ":8080"),
		CertFile:         getenv("CERT_FILE", "certs/server.crt"),
		KeyFile:          getenv("KEY_FILE", "certs/server.key"),
		DSN:              getenv("DSN", "postgres://postgres:postgres@localhost:5432/study_security?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
