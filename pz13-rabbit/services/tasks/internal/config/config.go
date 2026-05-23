package config

import "os"

const (
	defaultRabbitURL = "amqp://guest:guest@localhost:5672/"
	defaultQueueName = "task_events"
	defaultTasksAddr = ":8082"
)

type Config struct {
	RabbitURL string
	QueueName string
	HTTPAddr  string
}

func Load() Config {
	return Config{
		RabbitURL: getenv("RABBIT_URL", defaultRabbitURL),
		QueueName: getenv("QUEUE_NAME", defaultQueueName),
		HTTPAddr:  getenv("TASKS_ADDR", defaultTasksAddr),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
