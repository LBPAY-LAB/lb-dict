package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// AlertPublisher publishes rate limit alert events to Pulsar
type AlertPublisher struct {
	client   pulsar.Client
	producer pulsar.Producer
}

// AlertPublisherConfig contains configuration for the alert publisher
type AlertPublisherConfig struct {
	PulsarURL string
	Topic     string
}

// NewAlertPublisher creates a new AlertPublisher
func NewAlertPublisher(config AlertPublisherConfig) (*AlertPublisher, error) {
	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               config.PulsarURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pulsar client: %w", err)
	}

	// Create producer for rate limit alerts topic
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: config.Topic,
		Name:  "rate-limit-alert-producer",
		CompressionType: pulsar.LZ4,
		BatchingMaxMessages: 100,
		BatchingMaxPublishDelay: 10 * time.Millisecond,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create pulsar producer: %w", err)
	}

	return &AlertPublisher{
		client:   client,
		producer: producer,
	}, nil
}

// GetProducer returns the underlying Pulsar producer
// This is used by PublishAlertEventActivity
func (p *AlertPublisher) GetProducer() pulsar.Producer {
	return p.producer
}

// Close closes the publisher and releases resources
func (p *AlertPublisher) Close() error {
	if p.producer != nil {
		p.producer.Close()
	}
	if p.client != nil {
		p.client.Close()
	}
	return nil
}

// Publish publishes a message to the alerts topic
func (p *AlertPublisher) Publish(ctx context.Context, payload []byte, key string, properties map[string]string) error {
	_, err := p.producer.Send(ctx, &pulsar.ProducerMessage{
		Payload:    payload,
		Key:        key,
		Properties: properties,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}
