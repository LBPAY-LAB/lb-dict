package pulsar

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/lbpay-lab/conn-dict/internal/domain/events"
	"github.com/lbpay-lab/conn-dict/internal/domain/interfaces"
	"github.com/sirupsen/logrus"
)

// EventPublisher implements the EventPublisher interface using Apache Pulsar
type EventPublisher struct {
	client   pulsar.Client
	producer pulsar.Producer
	topic    string
	logger   *logrus.Logger
}

// NewEventPublisher creates a new Pulsar event publisher
func NewEventPublisher(
	pulsarURL string,
	topic string,
	logger *logrus.Logger,
) (*EventPublisher, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
		Name:  "conn-dict-event-publisher",
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Pulsar producer: %w", err)
	}

	return &EventPublisher{
		client:   client,
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

// Publish publishes a domain event to Pulsar
func (p *EventPublisher) Publish(ctx context.Context, event events.DomainEvent) error {
	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Pulsar message
	msg := &pulsar.ProducerMessage{
		Payload: payload,
		Key:     event.AggregateID(),
		Properties: map[string]string{
			"event_id":   event.EventID(),
			"event_type": event.EventType(),
		},
	}

	// Send message
	messageID, err := p.producer.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID(),
		"event_type": event.EventType(),
		"message_id": messageID,
	}).Debug("Event published successfully")

	return nil
}

// PublishBatch publishes multiple domain events in a batch
func (p *EventPublisher) PublishBatch(ctx context.Context, domainEvents []events.DomainEvent) error {
	for _, event := range domainEvents {
		if err := p.Publish(ctx, event); err != nil {
			p.logger.WithError(err).WithField("event_type", event.EventType()).Error("Failed to publish event in batch")
			// Continue with other events
			continue
		}
	}
	return nil
}

// Close closes the Pulsar producer and client
func (p *EventPublisher) Close() {
	p.producer.Close()
	p.client.Close()
}

// Ensure EventPublisher implements EventPublisher interface
var _ interfaces.EventPublisher = (*EventPublisher)(nil)
