package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TaskJobsQueue = "task_jobs"
	TaskJobsDLQ   = "task_jobs_dlq"
)

// URL returns the RabbitMQ connection URL from the environment or a local default.
func URL() string {
	if value := os.Getenv("RABBIT_URL"); value != "" {
		return value
	}
	return "amqp://guest:guest@localhost:5672/"
}

// Dial opens a RabbitMQ connection and channel.
func Dial(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("open RabbitMQ channel: %w", err)
	}

	return conn, ch, nil
}

// DeclareQueues creates the main task queue and the dead-letter queue.
func DeclareQueues(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(
		TaskJobsDLQ,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	); err != nil {
		return fmt.Errorf("declare %s: %w", TaskJobsDLQ, err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": TaskJobsDLQ,
	}

	if _, err := ch.QueueDeclare(
		TaskJobsQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		args,
	); err != nil {
		return fmt.Errorf("declare %s: %w", TaskJobsQueue, err)
	}

	return nil
}

// PublishJSON publishes a durable JSON message to the selected queue via the default exchange.
func PublishJSON(ctx context.Context, ch *amqp.Channel, queue string, payload any, messageID string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return ch.PublishWithContext(
		ctx,
		"",    // default exchange
		queue, // routing key equals queue name
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			MessageId:    messageID,
			Body:         body,
		},
	)
}
