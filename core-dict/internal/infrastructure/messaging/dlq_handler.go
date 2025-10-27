// Package messaging provides Dead Letter Queue handler for failed messages
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// DLQMessage represents a message that ended up in the Dead Letter Queue
type DLQMessage struct {
	OriginalTopic    string            `json:"original_topic"`
	MessageID        string            `json:"message_id"`
	MessageKey       string            `json:"message_key"`
	PublishTime      time.Time         `json:"publish_time"`
	DeliveryCount    uint32            `json:"delivery_count"`
	Payload          string            `json:"payload"`
	Properties       map[string]string `json:"properties"`
	FailureReason    string            `json:"failure_reason,omitempty"`
	ReceivedAt       time.Time         `json:"received_at"`
}

// DLQHandler handles messages that failed processing and ended up in DLQ
type DLQHandler struct {
	consumer pulsar.Consumer
	client   pulsar.Client
	config   *DLQConfig
}

// DLQConfig holds Dead Letter Queue handler configuration
type DLQConfig struct {
	PulsarURL        string
	DLQTopic         string
	SubscriptionName string
	AlertThreshold   int           // Alert after N failed messages
	AlertInterval    time.Duration // Alert once every interval
}

// DefaultDLQConfig returns default DLQ handler configuration
func DefaultDLQConfig() *DLQConfig {
	return &DLQConfig{
		PulsarURL:        "pulsar://localhost:6650",
		DLQTopic:         "dict.events.dlq",
		SubscriptionName: "core-dict-dlq-handler",
		AlertThreshold:   10,
		AlertInterval:    5 * time.Minute,
	}
}

// NewDLQHandler creates a new Dead Letter Queue handler
func NewDLQHandler(config *DLQConfig) (*DLQHandler, error) {
	if config == nil {
		config = DefaultDLQConfig()
	}

	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.PulsarURL,
		OperationTimeout:        30 * time.Second,
		ConnectionTimeout:       30 * time.Second,
		MaxConnectionsPerBroker: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client for DLQ handler: %w", err)
	}

	// Create DLQ consumer
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       config.DLQTopic,
		SubscriptionName:            config.SubscriptionName,
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create DLQ consumer: %w", err)
	}

	return &DLQHandler{
		consumer: consumer,
		client:   client,
		config:   config,
	}, nil
}

// Start starts the DLQ handler to monitor and process failed messages
func (h *DLQHandler) Start(ctx context.Context) error {
	log.Printf("[DLQHandler] Starting DLQ handler for topic: %s\n", h.config.DLQTopic)

	failedCount := 0
	lastAlert := time.Now()

	for {
		select {
		case <-ctx.Done():
			log.Println("[DLQHandler] Context cancelled, stopping DLQ handler...")
			return ctx.Err()
		default:
			msg, err := h.consumer.Receive(ctx)
			if err != nil {
				log.Printf("[DLQHandler] Error receiving DLQ message: %v\n", err)
				continue
			}

			// Process DLQ message
			if err := h.processDLQMessage(ctx, msg); err != nil {
				log.Printf("[DLQHandler] Error processing DLQ message: %v\n", err)
				// Even if processing fails, ACK the message to avoid infinite loop
				h.consumer.Ack(msg)
			} else {
				h.consumer.Ack(msg)
			}

			// Track failed messages and alert if threshold exceeded
			failedCount++
			if failedCount >= h.config.AlertThreshold && time.Since(lastAlert) > h.config.AlertInterval {
				h.alertOperationsTeam(failedCount)
				failedCount = 0
				lastAlert = time.Now()
			}
		}
	}
}

// processDLQMessage processes a single message from the DLQ
func (h *DLQHandler) processDLQMessage(ctx context.Context, msg pulsar.Message) error {
	// Extract message details
	dlqMsg := DLQMessage{
		OriginalTopic: extractOriginalTopic(msg),
		MessageID:     msg.ID().String(),
		MessageKey:    msg.Key(),
		PublishTime:   msg.PublishTime(),
		Payload:       string(msg.Payload()),
		Properties:    msg.Properties(),
		ReceivedAt:    time.Now(),
	}

	// Extract delivery count if available
	if deliveryCountStr, ok := msg.Properties()["RECONSUMETIMES"]; ok {
		// Pulsar stores redelivery count in properties
		var count uint32
		fmt.Sscanf(deliveryCountStr, "%d", &count)
		dlqMsg.DeliveryCount = count
	}

	// Extract failure reason if available
	if reason, ok := msg.Properties()["failure_reason"]; ok {
		dlqMsg.FailureReason = reason
	}

	// Log DLQ message details in structured format
	msgJSON, _ := json.MarshalIndent(dlqMsg, "", "  ")
	log.Printf("[DLQHandler] ===============================================\n")
	log.Printf("[DLQHandler] DEAD LETTER QUEUE MESSAGE\n")
	log.Printf("[DLQHandler] ===============================================\n")
	log.Printf("[DLQHandler] %s\n", string(msgJSON))
	log.Printf("[DLQHandler] ===============================================\n")

	// Store to DLQ storage (database, file, etc.) for manual review
	if err := h.storeDLQMessage(ctx, &dlqMsg); err != nil {
		log.Printf("[DLQHandler] Failed to store DLQ message: %v\n", err)
		// Don't return error - we still want to ACK to avoid reprocessing
	}

	// Check if message is retriable after manual intervention
	if h.isRetriable(dlqMsg) {
		log.Printf("[DLQHandler] Message is potentially retriable - flagged for manual review\n")
	}

	return nil
}

// storeDLQMessage stores DLQ message for later analysis and manual intervention
func (h *DLQHandler) storeDLQMessage(ctx context.Context, msg *DLQMessage) error {
	// TODO: Implement persistent storage
	// Options:
	// 1. Store in PostgreSQL table: dlq_messages
	// 2. Store in file system: /var/log/dict/dlq/
	// 3. Send to external monitoring system (Datadog, New Relic)
	// 4. Store in Redis for recent DLQ messages

	// For now, just log
	log.Printf("[DLQHandler] DLQ message stored (TODO: implement persistent storage)\n")
	return nil
}

// extractOriginalTopic extracts the original topic from DLQ message
func extractOriginalTopic(msg pulsar.Message) string {
	// Try to get from message properties first
	if originalTopic, ok := msg.Properties()["REAL_TOPIC"]; ok {
		return originalTopic
	}

	// Otherwise extract from topic name
	// DLQ topic format: "persistent://tenant/namespace/dict.events.dlq"
	// Original topic stored in properties or need to parse
	return msg.Topic()
}

// isRetriable determines if a DLQ message can be retried after manual intervention
func (h *DLQHandler) isRetriable(msg DLQMessage) bool {
	// Messages are retriable if:
	// 1. Failure was due to temporary infrastructure issue
	// 2. Database was down but now recovered
	// 3. External service was unavailable

	// Non-retriable:
	// 1. Invalid message format (proto unmarshal error)
	// 2. Business logic validation error
	// 3. Corrupted payload

	// Check failure reason
	retriableReasons := []string{
		"connection_timeout",
		"database_unavailable",
		"temporary_failure",
	}

	for _, reason := range retriableReasons {
		if msg.FailureReason == reason {
			return true
		}
	}

	return false
}

// alertOperationsTeam sends alert to operations team when DLQ threshold exceeded
func (h *DLQHandler) alertOperationsTeam(failedCount int) {
	log.Printf("[DLQHandler] ⚠️  ALERT: %d messages failed and sent to DLQ\n", failedCount)
	log.Printf("[DLQHandler] ⚠️  ACTION REQUIRED: Manual review needed for DLQ messages\n")

	// TODO: Send actual alerts via:
	// 1. Email notification
	// 2. Slack/Teams webhook
	// 3. PagerDuty incident
	// 4. Prometheus AlertManager

	// Example alert message:
	alertMsg := fmt.Sprintf(
		"🚨 DICT Core - DLQ Alert\n\n"+
			"Failed Messages: %d\n"+
			"DLQ Topic: %s\n"+
			"Time: %s\n\n"+
			"Action: Review DLQ messages and retry/discard as needed.",
		failedCount,
		h.config.DLQTopic,
		time.Now().Format(time.RFC3339),
	)

	log.Printf("[DLQHandler] Alert message: %s\n", alertMsg)
}

// Close closes the DLQ handler and releases resources
func (h *DLQHandler) Close() error {
	log.Println("[DLQHandler] Closing DLQ handler...")
	h.consumer.Close()
	h.client.Close()
	return nil
}

// GetDLQStats returns statistics about DLQ messages
func (h *DLQHandler) GetDLQStats(ctx context.Context) (*DLQStats, error) {
	// TODO: Implement DLQ statistics
	// Query stored DLQ messages and aggregate stats
	return &DLQStats{
		TotalMessages:     0,
		RetriableMessages: 0,
		ByTopic:           make(map[string]int),
		ByFailureReason:   make(map[string]int),
	}, nil
}

// DLQStats represents statistics about DLQ messages
type DLQStats struct {
	TotalMessages     int            `json:"total_messages"`
	RetriableMessages int            `json:"retriable_messages"`
	ByTopic           map[string]int `json:"by_topic"`
	ByFailureReason   map[string]int `json:"by_failure_reason"`
	OldestMessage     *time.Time     `json:"oldest_message,omitempty"`
	NewestMessage     *time.Time     `json:"newest_message,omitempty"`
}
