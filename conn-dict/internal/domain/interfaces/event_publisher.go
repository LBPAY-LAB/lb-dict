package interfaces

import (
	"context"

	"github.com/lbpay-lab/conn-dict/internal/domain/events"
)

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// Publish publishes a domain event to the message broker
	Publish(ctx context.Context, event events.DomainEvent) error

	// PublishBatch publishes multiple domain events in a batch
	PublishBatch(ctx context.Context, events []events.DomainEvent) error
}
