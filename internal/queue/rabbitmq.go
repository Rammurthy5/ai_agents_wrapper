package queue

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// Message represents a queued task
type Message struct {
	Prompt string `json:"prompt"`
	TaskID string `json:"task_id"`
}

// RabbitMQ manages queue connections
type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   amqp091.Queue
}

// NewRabbitMQ initializes a RabbitMQ connection
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		"ai_requests", // Queue name
		true,          // Durable
		false,         // Auto-delete
		false,         // Exclusive
		false,         // No-wait
		nil,           // Args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	return &RabbitMQ{conn: conn, channel: ch, queue: q}, nil
}

// Publish sends a message to the queue
func (r *RabbitMQ) Publish(msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	err = r.channel.Publish(
		"",           // Exchange
		r.queue.Name, // Routing key (queue name)
		false,        // Mandatory
		false,        // Immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish error: %v", err)
	}
	return nil
}

// Consume starts consuming messages from the queue
func (r *RabbitMQ) Consume() (<-chan amqp091.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name, // Queue
		"",           // Consumer tag
		true,         // Auto-ack (set to false for manual ack in production)
		false,        // Exclusive
		false,        // No-local
		false,        // No-wait
		nil,          // Args
	)
	if err != nil {
		return nil, fmt.Errorf("consume error: %v", err)
	}
	return msgs, nil
}

// Close shuts down the RabbitMQ connection
func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
