package config

import (
	"os"
	"strconv"
)

const (
	defaultRabbitURL = "amqp://guest:guest@localhost:5672/"
	defaultQueueName = "task_events"
	defaultPrefetch  = 1
)

type Config struct {
	RabbitURL string
	QueueName string
	Prefetch  int
}

func Load() Config {
	return Config{
		RabbitURL: getenv("RABBIT_URL", defaultRabbitURL),
		QueueName: getenv("QUEUE_NAME", defaultQueueName),
		Prefetch:  getenvInt("WORKER_PREFETCH", defaultPrefetch),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
