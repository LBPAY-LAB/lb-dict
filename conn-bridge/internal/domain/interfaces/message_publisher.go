package interfaces

import (
	"context"
)

// Message represents a message to be published
type Message struct {
	Topic         string
	Key           string
	Payload       []byte
	Headers       map[string]string
	CorrelationID string
}

// MessagePublisher defines the interface for publishing messages to a message broker
type MessagePublisher interface {
	// Publish publishes a message to the specified topic
	Publish(ctx context.Context, message *Message) error

	// PublishBatch publishes multiple messages in a batch
	PublishBatch(ctx context.Context, messages []*Message) error

	// Close closes the publisher connection
	Close() error

	// HealthCheck checks if the publisher is healthy
	HealthCheck(ctx context.Context) error
}
