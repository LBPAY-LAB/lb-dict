package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
)

// Producer wraps Pulsar producer with application-specific logic
type Producer struct {
	client   pulsar.Client
	producer pulsar.Producer
	logger   *logrus.Logger
	topic    string
}

// ProducerConfig holds configuration for Pulsar producer
type ProducerConfig struct {
	URL            string
	Topic          string
	ProducerName   string
	MaxReconnect   uint
	ConnectTimeout time.Duration
}

// NewProducer creates a new Pulsar producer
func NewProducer(config ProducerConfig, logger *logrus.Logger) (*Producer, error) {
	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.URL,
		ConnectionTimeout:       config.ConnectTimeout,
		MaxConnectionsPerBroker: 10,
		Logger:                  nil, // Use custom logger
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	// Create producer
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   config.Topic,
		Name:                    config.ProducerName,
		CompressionType:         pulsar.ZSTD,
		BatchingMaxPublishDelay: 100 * time.Millisecond,
		BatchingMaxMessages:     100,
		SendTimeout:             30 * time.Second,
		DisableBatching:         false,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	logger.Infof("Pulsar producer created: topic=%s, name=%s", config.Topic, config.ProducerName)

	return &Producer{
		client:   client,
		producer: producer,
		logger:   logger,
		topic:    config.Topic,
	}, nil
}

// PublishEvent publishes an event to Pulsar
func (p *Producer) PublishEvent(ctx context.Context, event interface{}, key string) error {
	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create message
	msg := &pulsar.ProducerMessage{
		Payload: payload,
		Key:     key,
		Properties: map[string]string{
			"event_time": time.Now().UTC().Format(time.RFC3339),
			"producer":   "conn-dict",
		},
	}

	// Send message asynchronously
	p.producer.SendAsync(ctx, msg, func(msgID pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
		if err != nil {
			p.logger.Errorf("Failed to publish event: key=%s, error=%v", key, err)
		} else {
			p.logger.Debugf("Event published: key=%s, msgID=%v", key, msgID)
		}
	})

	return nil
}

// PublishEventSync publishes an event synchronously
func (p *Producer) PublishEventSync(ctx context.Context, event interface{}, key string) (pulsar.MessageID, error) {
	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create message
	msg := &pulsar.ProducerMessage{
		Payload: payload,
		Key:     key,
		Properties: map[string]string{
			"event_time": time.Now().UTC().Format(time.RFC3339),
			"producer":   "conn-dict",
		},
	}

	// Send message synchronously
	msgID, err := p.producer.Send(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Debugf("Event published (sync): key=%s, msgID=%v", key, msgID)
	return msgID, nil
}

// Close closes the producer and client
func (p *Producer) Close() {
	p.producer.Close()
	p.client.Close()
	p.logger.Info("Pulsar producer closed")
}