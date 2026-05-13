package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr          string
	RedisAddr         string
	RedisPassword     string
	CacheTTL          time.Duration
	CacheTTLJitter    time.Duration
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
}

func New() Config {
	return Config{
		HTTPAddr:          getenv("HTTP_ADDR", ":8082"),
		RedisAddr:         getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getenv("REDIS_PASSWORD", ""),
		CacheTTL:          getDuration("CACHE_TTL", 120*time.Second),
		CacheTTLJitter:    getDuration("CACHE_TTL_JITTER", 30*time.Second),
		RedisDialTimeout:  getDuration("REDIS_DIAL_TIMEOUT", 2*time.Second),
		RedisReadTimeout:  getDuration("REDIS_READ_TIMEOUT", 2*time.Second),
		RedisWriteTimeout: getDuration("REDIS_WRITE_TIMEOUT", 2*time.Second),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
