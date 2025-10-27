// Package messaging provides configuration for Pulsar producers
package messaging

import (
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// ProducerConfig holds centralized configuration for all Pulsar producers
type ProducerConfig struct {
	// PulsarURL is the Pulsar broker URL (e.g., "pulsar://localhost:6650")
	PulsarURL string

	// CompressionType defines the compression algorithm
	// LZ4 provides ~60% size reduction with minimal CPU overhead
	CompressionType pulsar.CompressionType

	// BatchingMaxMessages is the max number of messages to batch
	// Higher values = better throughput, slightly higher latency
	BatchingMaxMessages uint

	// BatchingMaxPublishDelay is the max time to wait before sending a batch
	// Lower values = lower latency, slightly lower throughput
	BatchingMaxPublishDelay time.Duration

	// MaxPendingMessages is the max number of messages in the internal queue
	// If queue is full, Send() will block
	MaxPendingMessages int

	// SendTimeout is the max time to wait for a Send() operation
	SendTimeout time.Duration

	// ConnectionTimeout is the max time to wait for initial connection
	ConnectionTimeout time.Duration

	// OperationTimeout is the max time for any Pulsar operation
	OperationTimeout time.Duration
}

// DefaultEntryProducerConfig returns production-ready defaults optimized for low latency (<2s)
// These values balance throughput and latency for DICT operations
func DefaultEntryProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		// Default local Pulsar (override in production via env var)
		PulsarURL: "pulsar://localhost:6650",

		// LZ4 compression: fast, 60% size reduction
		CompressionType: pulsar.LZ4,

		// Batch up to 100 messages OR 10ms (whichever comes first)
		// This achieves ~95% batching efficiency while keeping latency <10ms
		BatchingMaxMessages:     100,
		BatchingMaxPublishDelay: 10 * time.Millisecond,

		// Allow 1000 messages in queue (prevents backpressure under load)
		MaxPendingMessages: 1000,

		// Send timeout: 30s (enough for retries + network delays)
		SendTimeout: 30 * time.Second,

		// Connection timeouts
		ConnectionTimeout: 5 * time.Second,
		OperationTimeout:  30 * time.Second,
	}
}

// HighThroughputConfig returns config optimized for maximum throughput
// Use this when latency is less critical (e.g., batch processing)
func HighThroughputConfig() *ProducerConfig {
	config := DefaultEntryProducerConfig()
	config.BatchingMaxMessages = 1000
	config.BatchingMaxPublishDelay = 100 * time.Millisecond
	config.MaxPendingMessages = 10000
	return config
}

// LowLatencyConfig returns config optimized for minimum latency
// Use this for critical real-time operations (e.g., immediate deletions)
func LowLatencyConfig() *ProducerConfig {
	config := DefaultEntryProducerConfig()
	config.BatchingMaxMessages = 10
	config.BatchingMaxPublishDelay = 1 * time.Millisecond
	config.SendTimeout = 5 * time.Second
	return config
}

// WithURL sets a custom Pulsar URL (builder pattern)
func (c *ProducerConfig) WithURL(url string) *ProducerConfig {
	c.PulsarURL = url
	return c
}

// WithCompression sets a custom compression type
func (c *ProducerConfig) WithCompression(compression pulsar.CompressionType) *ProducerConfig {
	c.CompressionType = compression
	return c
}

// WithBatching sets custom batching parameters
func (c *ProducerConfig) WithBatching(maxMessages uint, maxDelay time.Duration) *ProducerConfig {
	c.BatchingMaxMessages = maxMessages
	c.BatchingMaxPublishDelay = maxDelay
	return c
}

// WithTimeouts sets custom timeouts
func (c *ProducerConfig) WithTimeouts(send, connection, operation time.Duration) *ProducerConfig {
	c.SendTimeout = send
	c.ConnectionTimeout = connection
	c.OperationTimeout = operation
	return c
}

// Validate checks if the config is valid
func (c *ProducerConfig) Validate() error {
	if c.PulsarURL == "" {
		return ErrInvalidConfig{Field: "PulsarURL", Reason: "cannot be empty"}
	}
	if c.BatchingMaxMessages == 0 {
		return ErrInvalidConfig{Field: "BatchingMaxMessages", Reason: "must be > 0"}
	}
	if c.BatchingMaxPublishDelay <= 0 {
		return ErrInvalidConfig{Field: "BatchingMaxPublishDelay", Reason: "must be > 0"}
	}
	if c.MaxPendingMessages <= 0 {
		return ErrInvalidConfig{Field: "MaxPendingMessages", Reason: "must be > 0"}
	}
	if c.SendTimeout <= 0 {
		return ErrInvalidConfig{Field: "SendTimeout", Reason: "must be > 0"}
	}
	return nil
}

// ErrInvalidConfig represents a configuration validation error
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid producer config: " + e.Field + " - " + e.Reason
}

// ========================================================================
// TOPIC CONSTANTS
// ========================================================================

// Input topics (Core DICT → Connect)
const (
	TopicEntryCreated         = "persistent://public/default/dict.entries.created"
	TopicEntryUpdated         = "persistent://public/default/dict.entries.updated"
	TopicEntryDeletedImmediate = "persistent://public/default/dict.entries.deleted.immediate"
)

// Output topics (Connect → Core DICT)
const (
	TopicEntryStatusChanged = "persistent://public/default/dict.entries.status.changed"
	TopicClaimCreated       = "persistent://public/default/dict.claims.created"
	TopicClaimCompleted     = "persistent://public/default/dict.claims.completed"
	TopicInfractionReported = "persistent://public/default/dict.infractions.reported"
	TopicInfractionResolved = "persistent://public/default/dict.infractions.resolved"
)

// ========================================================================
// PERFORMANCE NOTES
// ========================================================================
//
// LATENCY TARGET: <2s end-to-end (Core → Connect → Bridge → Bacen)
//
// Breakdown:
//   - Core → Pulsar: <10ms (with batching)
//   - Pulsar → Connect: <50ms (consumer poll interval)
//   - Connect → Bridge: <100ms (gRPC call)
//   - Bridge → Bacen: <1500ms (SOAP + mTLS + Bacen processing)
//   - Total: ~1660ms (comfortable margin under 2s SLA)
//
// OPTIMIZATION STRATEGIES:
//   1. Batching: 100 msgs or 10ms (reduces per-message overhead)
//   2. Compression: LZ4 (60% size reduction, minimal CPU)
//   3. Partition key: EntryID (ensures FIFO ordering per entry)
//   4. Async sends: Non-blocking (use SendAsync for fire-and-forget)
//   5. Connection pooling: Reuse producer instances (don't recreate)
//
// MONITORING METRICS:
//   - pulsar_producer_send_latency_ms (p50, p95, p99)
//   - pulsar_producer_messages_sent_total
//   - pulsar_producer_errors_total
//   - pulsar_producer_batch_size (avg, max)
//
// ========================================================================
