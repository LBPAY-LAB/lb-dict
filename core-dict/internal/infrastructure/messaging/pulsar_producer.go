// Package messaging provides Apache Pulsar event producer functionality
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// PulsarProducerConfig holds Pulsar producer configuration
type PulsarProducerConfig struct {
	BrokerURL          string
	Topic              string
	ProducerName       string
	CompressionType    pulsar.CompressionType
	BatchingMaxMessages uint
	BatchingMaxDelay   time.Duration
	SendTimeout        time.Duration
	MaxReconnectToBroker *uint
}

// DefaultProducerConfig returns default Pulsar producer configuration
func DefaultProducerConfig() *PulsarProducerConfig {
	maxReconnect := uint(3)
	return &PulsarProducerConfig{
		BrokerURL:            "pulsar://localhost:6650",
		Topic:                "persistent://core-dict/events/domain-events",
		ProducerName:         "core-dict-producer",
		CompressionType:      pulsar.LZ4,
		BatchingMaxMessages:  100,
		BatchingMaxDelay:     10 * time.Millisecond,
		SendTimeout:          30 * time.Second,
		MaxReconnectToBroker: &maxReconnect,
	}
}

// EventProducer wraps Pulsar producer for sending domain events
type EventProducer struct {
	client   pulsar.Client
	producer pulsar.Producer
	config   *PulsarProducerConfig
}

// NewEventProducer creates a new Pulsar event producer
func NewEventProducer(config *PulsarProducerConfig) (*EventProducer, error) {
	if config == nil {
		config = DefaultProducerConfig()
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

	// Create producer
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   config.Topic,
		Name:                    config.ProducerName,
		CompressionType:         config.CompressionType,
		BatchingMaxMessages:     config.BatchingMaxMessages,
		BatchingMaxPublishDelay: config.BatchingMaxDelay,
		SendTimeout:             config.SendTimeout,
		MaxReconnectToBroker:    config.MaxReconnectToBroker,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Pulsar producer: %w", err)
	}

	return &EventProducer{
		client:   client,
		producer: producer,
		config:   config,
	}, nil
}

// DomainEvent represents a domain event to be published
type DomainEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Timestamp     time.Time              `json:"timestamp"`
	Version       int                    `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]string      `json:"metadata"`
}

// PublishEvent publishes a domain event to Pulsar
func (ep *EventProducer) PublishEvent(ctx context.Context, event *DomainEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create message properties
	properties := map[string]string{
		"event_id":       event.EventID,
		"event_type":     event.EventType,
		"aggregate_id":   event.AggregateID,
		"aggregate_type": event.AggregateType,
		"version":        fmt.Sprintf("%d", event.Version),
	}

	// Add metadata
	for k, v := range event.Metadata {
		properties["meta_"+k] = v
	}

	msg := &pulsar.ProducerMessage{
		Payload:    payload,
		Key:        event.AggregateID,
		Properties: properties,
		EventTime:  event.Timestamp,
	}

	// Send message
	messageID, err := ep.producer.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	// Log message ID (in production, use proper logger)
	_ = messageID

	return nil
}

// PublishEventAsync publishes an event asynchronously
func (ep *EventProducer) PublishEventAsync(ctx context.Context, event *DomainEvent, callback func(messageID pulsar.MessageID, err error)) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	payload, err := json.Marshal(event)
	if err != nil {
		if callback != nil {
			callback(nil, fmt.Errorf("failed to marshal event: %w", err))
		}
		return
	}

	properties := map[string]string{
		"event_id":       event.EventID,
		"event_type":     event.EventType,
		"aggregate_id":   event.AggregateID,
		"aggregate_type": event.AggregateType,
		"version":        fmt.Sprintf("%d", event.Version),
	}

	for k, v := range event.Metadata {
		properties["meta_"+k] = v
	}

	msg := &pulsar.ProducerMessage{
		Payload:    payload,
		Key:        event.AggregateID,
		Properties: properties,
		EventTime:  event.Timestamp,
	}

	ep.producer.SendAsync(ctx, msg, func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
		if callback != nil {
			callback(id, err)
		}
	})
}

// PublishBatch publishes multiple events in a batch
func (ep *EventProducer) PublishBatch(ctx context.Context, events []*DomainEvent) error {
	for _, event := range events {
		if err := ep.PublishEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventID, err)
		}
	}
	return nil
}

// Flush flushes all pending messages
func (ep *EventProducer) Flush() error {
	return ep.producer.Flush()
}

// Close closes the producer and client
func (ep *EventProducer) Close() {
	if ep.producer != nil {
		ep.producer.Close()
	}
	if ep.client != nil {
		ep.client.Close()
	}
}

// KeyEventProducer is specialized for DICT key events
type KeyEventProducer struct {
	producer *EventProducer
}

// NewKeyEventProducer creates a producer for DICT key events
func NewKeyEventProducer(brokerURL string) (*KeyEventProducer, error) {
	config := DefaultProducerConfig()
	config.BrokerURL = brokerURL
	config.Topic = "persistent://dict/events/key-events"
	config.ProducerName = "core-dict-key-producer"

	producer, err := NewEventProducer(config)
	if err != nil {
		return nil, err
	}

	return &KeyEventProducer{producer: producer}, nil
}

// PublishKeyCreated publishes a key created event
func (kp *KeyEventProducer) PublishKeyCreated(ctx context.Context, key, keyType, accountISPB string) error {
	event := &DomainEvent{
		EventID:       fmt.Sprintf("key_created_%s_%d", key, time.Now().UnixNano()),
		EventType:     "KeyCreated",
		AggregateID:   key,
		AggregateType: "DictEntry",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"key":          key,
			"key_type":     keyType,
			"account_ispb": accountISPB,
		},
		Metadata: map[string]string{
			"source": "core-dict",
		},
	}
	return kp.producer.PublishEvent(ctx, event)
}

// PublishKeyUpdated publishes a key updated event
func (kp *KeyEventProducer) PublishKeyUpdated(ctx context.Context, key string, changes map[string]interface{}) error {
	event := &DomainEvent{
		EventID:       fmt.Sprintf("key_updated_%s_%d", key, time.Now().UnixNano()),
		EventType:     "KeyUpdated",
		AggregateID:   key,
		AggregateType: "DictEntry",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"key":     key,
			"changes": changes,
		},
		Metadata: map[string]string{
			"source": "core-dict",
		},
	}
	return kp.producer.PublishEvent(ctx, event)
}

// PublishKeyDeleted publishes a key deleted event
func (kp *KeyEventProducer) PublishKeyDeleted(ctx context.Context, key, reason string) error {
	event := &DomainEvent{
		EventID:       fmt.Sprintf("key_deleted_%s_%d", key, time.Now().UnixNano()),
		EventType:     "KeyDeleted",
		AggregateID:   key,
		AggregateType: "DictEntry",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"key":    key,
			"reason": reason,
		},
		Metadata: map[string]string{
			"source": "core-dict",
		},
	}
	return kp.producer.PublishEvent(ctx, event)
}

// Close closes the producer
func (kp *KeyEventProducer) Close() {
	kp.producer.Close()
}

// ClaimEventProducer is specialized for claim events
type ClaimEventProducer struct {
	producer *EventProducer
}

// NewClaimEventProducer creates a producer for claim events
func NewClaimEventProducer(brokerURL string) (*ClaimEventProducer, error) {
	config := DefaultProducerConfig()
	config.BrokerURL = brokerURL
	config.Topic = "persistent://dict/events/claim-events"
	config.ProducerName = "core-dict-claim-producer"

	producer, err := NewEventProducer(config)
	if err != nil {
		return nil, err
	}

	return &ClaimEventProducer{producer: producer}, nil
}

// PublishClaimCreated publishes a claim created event
func (cp *ClaimEventProducer) PublishClaimCreated(ctx context.Context, claimID, key, claimType string) error {
	event := &DomainEvent{
		EventID:       fmt.Sprintf("claim_created_%s_%d", claimID, time.Now().UnixNano()),
		EventType:     "ClaimCreated",
		AggregateID:   claimID,
		AggregateType: "Claim",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"claim_id":   claimID,
			"key":        key,
			"claim_type": claimType,
		},
		Metadata: map[string]string{
			"source": "core-dict",
		},
	}
	return cp.producer.PublishEvent(ctx, event)
}

// PublishClaimConfirmed publishes a claim confirmed event
func (cp *ClaimEventProducer) PublishClaimConfirmed(ctx context.Context, claimID string) error {
	event := &DomainEvent{
		EventID:       fmt.Sprintf("claim_confirmed_%s_%d", claimID, time.Now().UnixNano()),
		EventType:     "ClaimConfirmed",
		AggregateID:   claimID,
		AggregateType: "Claim",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"claim_id": claimID,
		},
		Metadata: map[string]string{
			"source": "core-dict",
		},
	}
	return cp.producer.PublishEvent(ctx, event)
}

// PublishClaimCancelled publishes a claim cancelled event
func (cp *ClaimEventProducer) PublishClaimCancelled(ctx context.Context, claimID, reason string) error {
	event := &DomainEvent{
		EventID:       fmt.Sprintf("claim_cancelled_%s_%d", claimID, time.Now().UnixNano()),
		EventType:     "ClaimCancelled",
		AggregateID:   claimID,
		AggregateType: "Claim",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"claim_id": claimID,
			"reason":   reason,
		},
		Metadata: map[string]string{
			"source": "core-dict",
		},
	}
	return cp.producer.PublishEvent(ctx, event)
}

// Close closes the producer
func (cp *ClaimEventProducer) Close() {
	cp.producer.Close()
}
