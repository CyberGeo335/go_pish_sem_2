package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/CyberGeo335/pz13-rabbit/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp.Connection
	queueName  string
	prefetch   int
	logger     *log.Logger
}

func New(connection *amqp.Connection, queueName string, prefetch int, logger *log.Logger) *Consumer {
	return &Consumer{
		connection: connection,
		queueName:  queueName,
		prefetch:   prefetch,
		logger:     logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	ch, err := c.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return err
	}

	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return err
	}

	messages, err := ch.Consume(
		c.queueName,
		"",    // generated consumer name
		false, // autoAck disabled: manual ack is used
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return err
	}

	c.logger.Printf("worker started: queue=%s prefetch=%d", c.queueName, c.prefetch)

	for {
		select {
		case <-ctx.Done():
			c.logger.Println("worker stopping")
			return nil
		case delivery, ok := <-messages:
			if !ok {
				c.logger.Println("deliveries channel closed")
				return nil
			}
			c.handleDelivery(delivery)
		}
	}
}

func (c *Consumer) handleDelivery(delivery amqp.Delivery) {
	var event events.TaskEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.logger.Printf("bad message: error=%v body=%q", err, string(delivery.Body))
		if nackErr := delivery.Nack(false, false); nackErr != nil {
			c.logger.Printf("nack error: %v", nackErr)
		}
		return
	}

	c.logger.Printf(
		"received event=%s task_id=%s ts=%s request_id=%s producer=%s version=%s",
		event.Event,
		event.TaskID,
		event.TS,
		event.RequestID,
		event.Producer,
		event.Version,
	)

	if err := delivery.Ack(false); err != nil {
		c.logger.Printf("ack error: %v", err)
	}
}
