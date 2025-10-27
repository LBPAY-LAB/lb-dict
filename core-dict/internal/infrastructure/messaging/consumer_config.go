// Package messaging provides Apache Pulsar consumer configuration
package messaging

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// EntryEventConsumerConfig holds Pulsar consumer configuration for entry events
type EntryEventConsumerConfig struct {
	// PulsarURL is the Pulsar broker URL
	PulsarURL string

	// SubscriptionName is the consumer subscription name
	SubscriptionName string

	// SubscriptionType defines how messages are distributed among consumers
	// - Shared: Multiple consumers can share the same subscription (load balancing)
	// - Exclusive: Only one consumer can subscribe
	// - Failover: Multiple consumers, but only one active at a time
	// - KeyShared: Messages with same key go to same consumer
	SubscriptionType pulsar.SubscriptionType

	// NackRedeliveryDelay is the delay before redelivering a Nack'd message
	// Default: 60 seconds
	NackRedeliveryDelay time.Duration

	// MaxRedeliveryCount is the maximum number of times to redeliver a message
	// before sending it to the Dead Letter Queue (DLQ)
	// Default: 3
	MaxRedeliveryCount int

	// DLQTopic is the Dead Letter Queue topic for failed messages
	// Default: "dict.events.dlq"
	DLQTopic string

	// RetryTopic is the retry topic for temporary failures
	// Default: "dict.events.retry"
	RetryTopic string

	// ReceiveQueueSize is the consumer receive queue size
	// Default: 1000
	ReceiveQueueSize int

	// MaxPendingChunkedMessage limits the number of pending chunked messages
	// Default: 100
	MaxPendingChunkedMessage int
}

// DefaultEntryEventConsumerConfig returns default Pulsar consumer configuration for entry events
func DefaultEntryEventConsumerConfig() *EntryEventConsumerConfig {
	return &EntryEventConsumerConfig{
		PulsarURL:                "pulsar://localhost:6650",
		SubscriptionName:         "core-dict-events",
		SubscriptionType:         pulsar.Shared,
		NackRedeliveryDelay:      60 * time.Second,
		MaxRedeliveryCount:       3,
		DLQTopic:                 "dict.events.dlq",
		RetryTopic:               "dict.events.retry",
		ReceiveQueueSize:         1000,
		MaxPendingChunkedMessage: 100,
	}
}

// Validate validates the consumer configuration
func (c *EntryEventConsumerConfig) Validate() error {
	if c.PulsarURL == "" {
		return fmt.Errorf("PulsarURL cannot be empty")
	}
	if c.SubscriptionName == "" {
		return fmt.Errorf("SubscriptionName cannot be empty")
	}
	if c.NackRedeliveryDelay < 0 {
		return fmt.Errorf("NackRedeliveryDelay cannot be negative")
	}
	if c.MaxRedeliveryCount < 0 {
		return fmt.Errorf("MaxRedeliveryCount cannot be negative")
	}
	return nil
}

// WithPulsarURL sets the Pulsar broker URL
func (c *EntryEventConsumerConfig) WithPulsarURL(url string) *EntryEventConsumerConfig {
	c.PulsarURL = url
	return c
}

// WithSubscriptionName sets the subscription name
func (c *EntryEventConsumerConfig) WithSubscriptionName(name string) *EntryEventConsumerConfig {
	c.SubscriptionName = name
	return c
}

// WithSubscriptionType sets the subscription type
func (c *EntryEventConsumerConfig) WithSubscriptionType(subType pulsar.SubscriptionType) *EntryEventConsumerConfig {
	c.SubscriptionType = subType
	return c
}

// WithNackRedeliveryDelay sets the NACK redelivery delay
func (c *EntryEventConsumerConfig) WithNackRedeliveryDelay(delay time.Duration) *EntryEventConsumerConfig {
	c.NackRedeliveryDelay = delay
	return c
}

// WithMaxRedeliveryCount sets the max redelivery count before DLQ
func (c *EntryEventConsumerConfig) WithMaxRedeliveryCount(count int) *EntryEventConsumerConfig {
	c.MaxRedeliveryCount = count
	return c
}

// WithDLQTopic sets the Dead Letter Queue topic
func (c *EntryEventConsumerConfig) WithDLQTopic(topic string) *EntryEventConsumerConfig {
	c.DLQTopic = topic
	return c
}
