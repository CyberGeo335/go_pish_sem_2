package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/CyberGeo335/pz13-rabbit/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	taskCreatedEvent = "task.created"
	producerName     = "tasks"
	eventVersion     = "1"
)

type Publisher struct {
	channel   *amqp.Channel
	queueName string
}

func New(channel *amqp.Channel, queueName string) *Publisher {
	return &Publisher{channel: channel, queueName: queueName}
}

func (p *Publisher) PublishTaskCreated(ctx context.Context, taskID, requestID string) error {
	if p == nil || p.channel == nil {
		return errors.New("rabbit publisher is not configured")
	}

	_, err := p.channel.QueueDeclare(
		p.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	msg := events.TaskEvent{
		Event:     taskCreatedEvent,
		TaskID:    taskID,
		TS:        now.Format(time.RFC3339),
		RequestID: requestID,
		Producer:  producerName,
		Version:   eventVersion,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		"",          // default exchange
		p.queueName, // routing key is the queue name
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    now,
			MessageId:    taskID,
			Type:         taskCreatedEvent,
			AppId:        producerName,
			Body:         body,
		},
	)
}
