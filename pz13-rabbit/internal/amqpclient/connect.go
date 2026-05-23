package amqpclient

import amqp "github.com/rabbitmq/amqp091-go"

// Connect opens a RabbitMQ AMQP connection.
func Connect(url string) (*amqp.Connection, error) {
	return amqp.Dial(url)
}
