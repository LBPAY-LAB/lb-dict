package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/domain/events"
)

const (
	// Topic name for DICT responses from Bacen
	TopicDictResOut = "rsfn-dict-res-out"

	// Producer name
	ProducerName = "rsfn-bridge-producer"

	// Max retry attempts for failed publishes
	MaxRetryAttempts = 3

	// Retry delay
	RetryDelay = 100 * time.Millisecond
)

var (
	// Prometheus metrics
	messagesPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsar_messages_published_total",
			Help: "Total number of messages published to Pulsar",
		},
		[]string{"event_type"},
	)

	publishErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pulsar_publish_errors_total",
			Help: "Total number of publish errors",
		},
	)

	publishDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "pulsar_publish_duration_seconds",
			Help:    "Duration of publish operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)

// DictPublisher handles publishing DICT events to Pulsar
type DictPublisher struct {
	client   pulsar.Client
	producer pulsar.Producer
	config   *DictPublisherConfig
	logger   *logrus.Logger
	mu       sync.RWMutex
	closed   bool
}

// DictPublisherConfig holds configuration for DictPublisher
type DictPublisherConfig struct {
	BrokerURL         string
	Topic             string
	ProducerName      string
	BatchingEnabled   bool
	MaxMessages       uint
	BatchingMaxDelay  time.Duration
	CompressionType   pulsar.CompressionType
	OperationTimeout  time.Duration
	ConnectionTimeout time.Duration
}

// NewDictPublisher creates a new DictPublisher
func NewDictPublisher(config *DictPublisherConfig) (*DictPublisher, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.BrokerURL == "" {
		return nil, fmt.Errorf("broker URL is required")
	}

	// Set defaults
	if config.Topic == "" {
		config.Topic = TopicDictResOut
	}
	if config.ProducerName == "" {
		config.ProducerName = ProducerName
	}
	if config.MaxMessages == 0 {
		config.MaxMessages = 100
	}
	if config.BatchingMaxDelay == 0 {
		config.BatchingMaxDelay = 10 * time.Millisecond
	}
	if config.CompressionType == 0 {
		config.CompressionType = pulsar.LZ4
	}
	if config.OperationTimeout == 0 {
		config.OperationTimeout = 30 * time.Second
	}
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 10 * time.Second
	}

	// Create Pulsar client
	clientOptions := pulsar.ClientOptions{
		URL:               config.BrokerURL,
		OperationTimeout:  config.OperationTimeout,
		ConnectionTimeout: config.ConnectionTimeout,
	}

	client, err := pulsar.NewClient(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	// Create producer
	producerOptions := pulsar.ProducerOptions{
		Topic:                   config.Topic,
		Name:                    config.ProducerName,
		CompressionType:         config.CompressionType,
		BatchingMaxPublishDelay: config.BatchingMaxDelay,
		BatchingMaxMessages:     config.MaxMessages,
		DisableBatching:         !config.BatchingEnabled,
	}

	producer, err := client.CreateProducer(producerOptions)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	dp := &DictPublisher{
		client:   client,
		producer: producer,
		config:   config,
		logger:   logger,
		closed:   false,
	}

	logger.WithFields(logrus.Fields{
		"broker_url":    config.BrokerURL,
		"topic":         config.Topic,
		"producer_name": config.ProducerName,
		"batching":      config.BatchingEnabled,
		"compression":   config.CompressionType,
	}).Info("DictPublisher initialized successfully")

	return dp, nil
}

// PublishEntryCreated publishes an entry created event
func (dp *DictPublisher) PublishEntryCreated(ctx context.Context, entry *entities.DictEntry, traceID string) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	event := events.NewEntryCreatedEvent(entry, traceID)
	return dp.publishEvent(ctx, event, string(events.EventTypeEntryCreated))
}

// PublishEntryUpdated publishes an entry updated event
func (dp *DictPublisher) PublishEntryUpdated(ctx context.Context, entry *entities.DictEntry, traceID string) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	event := events.NewEntryUpdatedEvent(entry, nil, traceID)
	return dp.publishEvent(ctx, event, string(events.EventTypeEntryUpdated))
}

// PublishEntryDeleted publishes an entry deleted event
func (dp *DictPublisher) PublishEntryDeleted(ctx context.Context, keyID string, traceID string) error {
	if keyID == "" {
		return fmt.Errorf("keyID cannot be empty")
	}

	event := events.NewEntryDeletedEvent(keyID, "", "", traceID)
	return dp.publishEvent(ctx, event, string(events.EventTypeEntryDeleted))
}

// PublishError publishes an error event
func (dp *DictPublisher) PublishError(ctx context.Context, err error, context string, traceID string) error {
	if err == nil {
		return fmt.Errorf("error cannot be nil")
	}

	event := events.NewErrorEvent("ERROR", err.Error(), context, traceID)
	return dp.publishEvent(ctx, event, string(events.EventTypeError))
}

// publishEvent is the internal method that handles the actual publishing
func (dp *DictPublisher) publishEvent(ctx context.Context, event interface{}, eventType string) error {
	dp.mu.RLock()
	if dp.closed {
		dp.mu.RUnlock()
		return fmt.Errorf("publisher is closed")
	}
	dp.mu.RUnlock()

	// Start timing
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		publishDuration.Observe(duration)
	}()

	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		dp.logger.WithError(err).Error("Failed to marshal event")
		publishErrorsTotal.Inc()
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Extract metadata from event
	var timestamp, traceID string
	switch e := event.(type) {
	case *events.EntryCreatedEvent:
		timestamp = e.Timestamp
		traceID = e.TraceID
	case *events.EntryUpdatedEvent:
		timestamp = e.Timestamp
		traceID = e.TraceID
	case *events.EntryDeletedEvent:
		timestamp = e.Timestamp
		traceID = e.TraceID
	case *events.ErrorEvent:
		timestamp = e.Timestamp
		traceID = e.TraceID
	}

	// Build Pulsar message
	msg := &pulsar.ProducerMessage{
		Payload: payload,
		Properties: map[string]string{
			"event_type": eventType,
			"source":     "rsfn-bridge",
			"timestamp":  timestamp,
			"trace_id":   traceID,
		},
	}

	// Publish with retry logic (async, non-blocking)
	go dp.publishWithRetry(ctx, msg, eventType)

	return nil
}

// publishWithRetry attempts to publish a message with retry logic
func (dp *DictPublisher) publishWithRetry(ctx context.Context, msg *pulsar.ProducerMessage, eventType string) {
	var lastErr error

	for attempt := 1; attempt <= MaxRetryAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			dp.logger.WithFields(logrus.Fields{
				"event_type": eventType,
				"attempt":    attempt,
			}).Warn("Publish cancelled by context")
			return
		default:
		}

		// Check if publisher is closed
		dp.mu.RLock()
		if dp.closed {
			dp.mu.RUnlock()
			return
		}
		dp.mu.RUnlock()

		// Attempt to send
		_, err := dp.producer.Send(ctx, msg)
		if err == nil {
			// Success
			messagesPublishedTotal.WithLabelValues(eventType).Inc()
			dp.logger.WithFields(logrus.Fields{
				"event_type": eventType,
				"trace_id":   msg.Properties["trace_id"],
				"attempt":    attempt,
			}).Debug("Message published successfully")
			return
		}

		// Log error
		lastErr = err
		dp.logger.WithFields(logrus.Fields{
			"event_type": eventType,
			"attempt":    attempt,
			"max_retry":  MaxRetryAttempts,
			"error":      err,
		}).Warn("Failed to publish message, will retry")

		// Wait before retry (except on last attempt)
		if attempt < MaxRetryAttempts {
			time.Sleep(RetryDelay * time.Duration(attempt))
		}
	}

	// All retries failed
	publishErrorsTotal.Inc()
	dp.logger.WithFields(logrus.Fields{
		"event_type": eventType,
		"trace_id":   msg.Properties["trace_id"],
		"error":      lastErr,
	}).Error("Failed to publish message after all retry attempts")
}

// Close closes the publisher gracefully
func (dp *DictPublisher) Close() error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if dp.closed {
		return nil
	}

	dp.closed = true

	dp.logger.Info("Closing DictPublisher...")

	// Flush pending messages
	if dp.producer != nil {
		if err := dp.producer.Flush(); err != nil {
			dp.logger.WithError(err).Warn("Failed to flush producer")
		}
		dp.producer.Close()
	}

	// Close client
	if dp.client != nil {
		dp.client.Close()
	}

	dp.logger.Info("DictPublisher closed successfully")
	return nil
}

// DefaultDictPublisherConfig returns a default configuration
func DefaultDictPublisherConfig(brokerURL string) *DictPublisherConfig {
	return &DictPublisherConfig{
		BrokerURL:         brokerURL,
		Topic:             TopicDictResOut,
		ProducerName:      ProducerName,
		BatchingEnabled:   true,
		MaxMessages:       100,
		BatchingMaxDelay:  10 * time.Millisecond,
		CompressionType:   pulsar.LZ4,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 10 * time.Second,
	}
}
