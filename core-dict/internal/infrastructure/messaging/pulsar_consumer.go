// Package messaging provides Apache Pulsar event consumer functionality
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// PulsarConsumerConfig holds Pulsar consumer configuration
type PulsarConsumerConfig struct {
	BrokerURL        string
	Topic            string
	SubscriptionName string
	ConsumerName     string
	SubscriptionType pulsar.SubscriptionType
	ReceiverQueueSize int
	NackRedeliveryDelay time.Duration
	MaxReconnectToBroker *uint
}

// DefaultConsumerConfig returns default Pulsar consumer configuration
func DefaultConsumerConfig() *PulsarConsumerConfig {
	maxReconnect := uint(3)
	return &PulsarConsumerConfig{
		BrokerURL:           "pulsar://localhost:6650",
		Topic:               "persistent://lb-conn/dict/rsfn-dict-res-in",
		SubscriptionName:    "core-dict-sub",
		ConsumerName:        "core-dict-consumer",
		SubscriptionType:    pulsar.Shared,
		ReceiverQueueSize:   1000,
		NackRedeliveryDelay: 60 * time.Second,
		MaxReconnectToBroker: &maxReconnect,
	}
}

// EventConsumer wraps Pulsar consumer for receiving events
type EventConsumer struct {
	client   pulsar.Client
	consumer pulsar.Consumer
	config   *PulsarConsumerConfig
	handlers map[string]EventHandler
}

// EventHandler processes incoming events
type EventHandler func(ctx context.Context, event *DomainEvent) error

// NewEventConsumer creates a new Pulsar event consumer
func NewEventConsumer(config *PulsarConsumerConfig) (*EventConsumer, error) {
	if config == nil {
		config = DefaultConsumerConfig()
	}

	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.BrokerURL,
		OperationTimeout:        30 * time.Second,
		ConnectionTimeout:       30 * time.Second,
		MaxConnectionsPerBroker: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	// Create consumer
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       config.Topic,
		SubscriptionName:            config.SubscriptionName,
		Name:                        config.ConsumerName,
		Type:                        config.SubscriptionType,
		ReceiverQueueSize:           config.ReceiverQueueSize,
		NackRedeliveryDelay:         config.NackRedeliveryDelay,
		MaxReconnectToBroker:        config.MaxReconnectToBroker,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Pulsar consumer: %w", err)
	}

	return &EventConsumer{
		client:   client,
		consumer: consumer,
		config:   config,
		handlers: make(map[string]EventHandler),
	}, nil
}

// RegisterHandler registers an event handler for a specific event type
func (ec *EventConsumer) RegisterHandler(eventType string, handler EventHandler) {
	ec.handlers[eventType] = handler
}

// Start starts consuming messages
func (ec *EventConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := ec.consumer.Receive(ctx)
			if err != nil {
				// Check if context was cancelled
				if ctx.Err() != nil {
					return ctx.Err()
				}
				// Log error and continue
				fmt.Printf("error receiving message: %v\n", err)
				continue
			}

			// Process message
			if err := ec.processMessage(ctx, msg); err != nil {
				fmt.Printf("error processing message: %v\n", err)
				// Negative acknowledge (will be redelivered)
				ec.consumer.Nack(msg)
			} else {
				// Acknowledge
				ec.consumer.Ack(msg)
			}
		}
	}
}

// processMessage processes a single message
func (ec *EventConsumer) processMessage(ctx context.Context, msg pulsar.Message) error {
	// Unmarshal event
	var event DomainEvent
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Get handler for event type
	handler, ok := ec.handlers[event.EventType]
	if !ok {
		// No handler registered, log and ack
		fmt.Printf("no handler registered for event type: %s\n", event.EventType)
		return nil
	}

	// Call handler
	if err := handler(ctx, &event); err != nil {
		return fmt.Errorf("handler failed for event type %s: %w", event.EventType, err)
	}

	return nil
}

// Close closes the consumer and client
func (ec *EventConsumer) Close() {
	if ec.consumer != nil {
		ec.consumer.Close()
	}
	if ec.client != nil {
		ec.client.Close()
	}
}

// ResponseEventConsumer handles response events from Connect/Bridge
type ResponseEventConsumer struct {
	consumer *EventConsumer
}

// NewResponseEventConsumer creates a consumer for response events
func NewResponseEventConsumer(brokerURL string) (*ResponseEventConsumer, error) {
	config := DefaultConsumerConfig()
	config.BrokerURL = brokerURL
	config.Topic = "persistent://lb-conn/dict/rsfn-dict-res-in"
	config.SubscriptionName = "core-dict-response-sub"
	config.ConsumerName = "core-dict-response-consumer"

	consumer, err := NewEventConsumer(config)
	if err != nil {
		return nil, err
	}

	return &ResponseEventConsumer{consumer: consumer}, nil
}

// OnKeyCreatedResponse handles KeyCreated response from RSFN
func (rc *ResponseEventConsumer) OnKeyCreatedResponse(handler func(ctx context.Context, key string, success bool, errorMsg string) error) {
	rc.consumer.RegisterHandler("KeyCreatedResponse", func(ctx context.Context, event *DomainEvent) error {
		key, _ := event.Data["key"].(string)
		success, _ := event.Data["success"].(bool)
		errorMsg, _ := event.Data["error"].(string)
		return handler(ctx, key, success, errorMsg)
	})
}

// OnClaimConfirmedResponse handles ClaimConfirmed response from RSFN
func (rc *ResponseEventConsumer) OnClaimConfirmedResponse(handler func(ctx context.Context, claimID string, success bool) error) {
	rc.consumer.RegisterHandler("ClaimConfirmedResponse", func(ctx context.Context, event *DomainEvent) error {
		claimID, _ := event.Data["claim_id"].(string)
		success, _ := event.Data["success"].(bool)
		return handler(ctx, claimID, success)
	})
}

// OnVSYNCResponse handles VSYNC response from RSFN
func (rc *ResponseEventConsumer) OnVSYNCResponse(handler func(ctx context.Context, syncID string, entriesCount int) error) {
	rc.consumer.RegisterHandler("VSYNCResponse", func(ctx context.Context, event *DomainEvent) error {
		syncID, _ := event.Data["sync_id"].(string)
		entriesCount, _ := event.Data["entries_count"].(int)
		return handler(ctx, syncID, entriesCount)
	})
}

// Start starts consuming response events
func (rc *ResponseEventConsumer) Start(ctx context.Context) error {
	return rc.consumer.Start(ctx)
}

// Close closes the consumer
func (rc *ResponseEventConsumer) Close() {
	rc.consumer.Close()
}

// MultiTopicConsumer consumes from multiple topics
type MultiTopicConsumer struct {
	client    pulsar.Client
	consumer  pulsar.Consumer
	handlers  map[string]EventHandler
}

// NewMultiTopicConsumer creates a consumer for multiple topics
func NewMultiTopicConsumer(brokerURL string, topics []string, subscriptionName string) (*MultiTopicConsumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     brokerURL,
		OperationTimeout:        30 * time.Second,
		ConnectionTimeout:       30 * time.Second,
		MaxConnectionsPerBroker: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topics:                      topics,
		SubscriptionName:            subscriptionName,
		Type:                        pulsar.Shared,
		ReceiverQueueSize:           1000,
		NackRedeliveryDelay:         60 * time.Second,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create multi-topic consumer: %w", err)
	}

	return &MultiTopicConsumer{
		client:   client,
		consumer: consumer,
		handlers: make(map[string]EventHandler),
	}, nil
}

// RegisterHandler registers an event handler for a specific event type
func (mc *MultiTopicConsumer) RegisterHandler(eventType string, handler EventHandler) {
	mc.handlers[eventType] = handler
}

// Start starts consuming messages
func (mc *MultiTopicConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := mc.consumer.Receive(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				fmt.Printf("error receiving message: %v\n", err)
				continue
			}

			if err := mc.processMessage(ctx, msg); err != nil {
				fmt.Printf("error processing message: %v\n", err)
				mc.consumer.Nack(msg)
			} else {
				mc.consumer.Ack(msg)
			}
		}
	}
}

// processMessage processes a single message
func (mc *MultiTopicConsumer) processMessage(ctx context.Context, msg pulsar.Message) error {
	var event DomainEvent
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	handler, ok := mc.handlers[event.EventType]
	if !ok {
		fmt.Printf("no handler registered for event type: %s\n", event.EventType)
		return nil
	}

	if err := handler(ctx, &event); err != nil {
		return fmt.Errorf("handler failed for event type %s: %w", event.EventType, err)
	}

	return nil
}

// Close closes the consumer and client
func (mc *MultiTopicConsumer) Close() {
	if mc.consumer != nil {
		mc.consumer.Close()
	}
	if mc.client != nil {
		mc.client.Close()
	}
}

// DeadLetterPolicy configures dead letter queue for failed messages
type DeadLetterPolicy struct {
	MaxRedeliveryCount uint32
	DeadLetterTopic    string
}

// NewConsumerWithDLQ creates a consumer with dead letter queue support
func NewConsumerWithDLQ(config *PulsarConsumerConfig, dlqPolicy *DeadLetterPolicy) (*EventConsumer, error) {
	if config == nil {
		config = DefaultConsumerConfig()
	}

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.BrokerURL,
		OperationTimeout:        30 * time.Second,
		ConnectionTimeout:       30 * time.Second,
		MaxConnectionsPerBroker: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	dlq := pulsar.DLQPolicy{
		MaxDeliveries:   dlqPolicy.MaxRedeliveryCount,
		DeadLetterTopic: dlqPolicy.DeadLetterTopic,
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       config.Topic,
		SubscriptionName:            config.SubscriptionName,
		Name:                        config.ConsumerName,
		Type:                        config.SubscriptionType,
		ReceiverQueueSize:           config.ReceiverQueueSize,
		NackRedeliveryDelay:         config.NackRedeliveryDelay,
		DLQ:                         &dlq,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Pulsar consumer with DLQ: %w", err)
	}

	return &EventConsumer{
		client:   client,
		consumer: consumer,
		config:   config,
		handlers: make(map[string]EventHandler),
	}, nil
}
