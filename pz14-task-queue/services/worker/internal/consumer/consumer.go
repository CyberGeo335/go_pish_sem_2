package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/CyberGeo335/pz14-task-queue/internal/jobs"
	"github.com/CyberGeo335/pz14-task-queue/internal/rabbitmq"
	"github.com/CyberGeo335/pz14-task-queue/services/worker/internal/store"
	amqp "github.com/rabbitmq/amqp091-go"
)

const MaxAttempts = 3

type Consumer struct {
	channel     *amqp.Channel
	processed   *store.ProcessedStore
	maxAttempts int
}

func New(channel *amqp.Channel, processed *store.ProcessedStore) *Consumer {
	return &Consumer{
		channel:     channel,
		processed:   processed,
		maxAttempts: MaxAttempts,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	if err := c.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	deliveries, err := c.channel.Consume(
		rabbitmq.TaskJobsQueue,
		"pz14-worker",
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume %s: %w", rabbitmq.TaskJobsQueue, err)
	}

	log.Printf("worker is consuming queue=%s max_attempts=%d", rabbitmq.TaskJobsQueue, c.maxAttempts)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel was closed")
			}
			c.handleDelivery(ctx, delivery)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	var job jobs.TaskJob
	if err := json.Unmarshal(d.Body, &job); err != nil {
		log.Printf("invalid json; message will be dead-lettered: %v body=%s", err, string(d.Body))
		_ = d.Nack(false, false)
		return
	}

	if err := validateJob(job); err != nil {
		log.Printf("invalid job; message will be dead-lettered: %v body=%s", err, string(d.Body))
		_ = d.Nack(false, false)
		return
	}

	log.Printf("received job=%s task_id=%s message_id=%s attempt=%d", job.Job, job.TaskID, job.MessageID, job.Attempt)

	if c.processed.Exists(job.MessageID) {
		log.Printf("duplicate message_id=%s detected; ack without processing", job.MessageID)
		_ = d.Ack(false)
		return
	}

	if err := processTask(job); err != nil {
		log.Printf("processing failed task_id=%s message_id=%s attempt=%d error=%v", job.TaskID, job.MessageID, job.Attempt, err)
		c.retryOrSendToDLQ(ctx, d, job)
		return
	}

	c.processed.MarkDone(job.MessageID)
	log.Printf("processed successfully task_id=%s message_id=%s; ack", job.TaskID, job.MessageID)
	_ = d.Ack(false)
}

func (c *Consumer) retryOrSendToDLQ(ctx context.Context, d amqp.Delivery, job jobs.TaskJob) {
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if job.Attempt >= c.maxAttempts {
		log.Printf("max attempts reached task_id=%s message_id=%s attempt=%d; publish to %s", job.TaskID, job.MessageID, job.Attempt, rabbitmq.TaskJobsDLQ)
		if err := rabbitmq.PublishJSON(publishCtx, c.channel, rabbitmq.TaskJobsDLQ, job, job.MessageID); err != nil {
			log.Printf("dlq publish failed, fallback to broker dead-lettering: %v", err)
			_ = d.Nack(false, false)
			return
		}
		_ = d.Ack(false)
		return
	}

	job.Attempt++
	log.Printf("retry task_id=%s message_id=%s next_attempt=%d", job.TaskID, job.MessageID, job.Attempt)
	if err := rabbitmq.PublishJSON(publishCtx, c.channel, rabbitmq.TaskJobsQueue, job, job.MessageID); err != nil {
		log.Printf("retry publish failed; original message will be requeued: %v", err)
		_ = d.Nack(false, true)
		return
	}

	_ = d.Ack(false)
}

func validateJob(job jobs.TaskJob) error {
	if strings.TrimSpace(job.Job) != "process_task" {
		return fmt.Errorf("unsupported job %q", job.Job)
	}
	if strings.TrimSpace(job.TaskID) == "" {
		return fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(job.MessageID) == "" {
		return fmt.Errorf("message_id is required")
	}
	if job.Attempt < 1 {
		return fmt.Errorf("attempt must be greater than zero")
	}
	return nil
}

func processTask(job jobs.TaskJob) error {
	log.Printf("simulate heavy work task_id=%s", job.TaskID)
	time.Sleep(2 * time.Second)

	if job.TaskID == "t_fail" {
		return fmt.Errorf("simulated processing error")
	}

	return nil
}
