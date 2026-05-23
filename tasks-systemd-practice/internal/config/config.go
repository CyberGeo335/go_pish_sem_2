package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost         = "0.0.0.0"
	defaultPort         = "8082"
	defaultLogLevel     = "info"
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

// Config contains runtime settings loaded from environment variables.
type Config struct {
	Host        string
	Port        string
	AuthBaseURL string
	RedisAddr   string
	LogLevel    string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// FromEnv loads application configuration from environment variables.
func FromEnv() Config {
	return Config{
		Host:         getEnv("TASKS_HOST", defaultHost),
		Port:         normalizePort(getEnv("TASKS_PORT", defaultPort)),
		AuthBaseURL:  getEnv("AUTH_BASE_URL", ""),
		RedisAddr:    getEnv("REDIS_ADDR", ""),
		LogLevel:     strings.ToLower(getEnv("LOG_LEVEL", defaultLogLevel)),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT_SECONDS", defaultReadTimeout),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT_SECONDS", defaultWriteTimeout),
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT_SECONDS", defaultIdleTimeout),
	}
}

// Addr returns the address used by net/http server.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func normalizePort(port string) string {
	return strings.TrimPrefix(strings.TrimSpace(port), ":")
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}
