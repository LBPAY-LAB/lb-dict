package pulsar

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
)

// Publisher implements the MessagePublisher interface using Apache Pulsar
type Publisher struct {
	client    pulsar.Client
	producers map[string]pulsar.Producer
	config    *Config
}

// Config holds the configuration for the Pulsar publisher
type Config struct {
	BrokerURL          string
	OperationTimeout   time.Duration
	ConnectionTimeout  time.Duration
	MaxConnectionsPerBroker int
}

// NewPublisher creates a new Pulsar publisher
func NewPublisher(config *Config) (interfaces.MessagePublisher, error) {
	if config.BrokerURL == "" {
		return nil, fmt.Errorf("broker URL is required")
	}

	clientOptions := pulsar.ClientOptions{
		URL:                     config.BrokerURL,
		OperationTimeout:        config.OperationTimeout,
		ConnectionTimeout:       config.ConnectionTimeout,
		MaxConnectionsPerBroker: config.MaxConnectionsPerBroker,
	}

	client, err := pulsar.NewClient(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	return &Publisher{
		client:    client,
		producers: make(map[string]pulsar.Producer),
		config:    config,
	}, nil
}

// Publish publishes a message to the specified topic
func (p *Publisher) Publish(ctx context.Context, message *interfaces.Message) error {
	producer, err := p.getOrCreateProducer(message.Topic)
	if err != nil {
		return fmt.Errorf("failed to get producer for topic %s: %w", message.Topic, err)
	}

	// Build Pulsar message
	msg := &pulsar.ProducerMessage{
		Payload: message.Payload,
		Key:     message.Key,
		Properties: message.Headers,
	}

	// Add correlation ID to properties
	if msg.Properties == nil {
		msg.Properties = make(map[string]string)
	}
	msg.Properties["correlation_id"] = message.CorrelationID

	// Send message
	_, err = producer.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s: %w", message.Topic, err)
	}

	return nil
}

// PublishBatch publishes multiple messages in a batch
func (p *Publisher) PublishBatch(ctx context.Context, messages []*interfaces.Message) error {
	// Group messages by topic
	messagesByTopic := make(map[string][]*interfaces.Message)
	for _, msg := range messages {
		messagesByTopic[msg.Topic] = append(messagesByTopic[msg.Topic], msg)
	}

	// Publish each topic's messages
	for topic, topicMessages := range messagesByTopic {
		producer, err := p.getOrCreateProducer(topic)
		if err != nil {
			return fmt.Errorf("failed to get producer for topic %s: %w", topic, err)
		}

		for _, msg := range topicMessages {
			pulsarMsg := &pulsar.ProducerMessage{
				Payload: msg.Payload,
				Key:     msg.Key,
				Properties: msg.Headers,
			}

			if pulsarMsg.Properties == nil {
				pulsarMsg.Properties = make(map[string]string)
			}
			pulsarMsg.Properties["correlation_id"] = msg.CorrelationID

			// Use SendAsync for batch operations
			producer.SendAsync(ctx, pulsarMsg, func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
				if err != nil {
					// TODO: Implement proper error handling
					fmt.Printf("Failed to send message: %v\n", err)
				}
			})
		}

		// Flush to ensure all messages are sent
		if err := producer.Flush(); err != nil {
			return fmt.Errorf("failed to flush messages for topic %s: %w", topic, err)
		}
	}

	return nil
}

// Close closes the publisher connection
func (p *Publisher) Close() error {
	// Close all producers
	for _, producer := range p.producers {
		producer.Close()
	}

	// Close client
	p.client.Close()
	return nil
}

// HealthCheck checks if the publisher is healthy
func (p *Publisher) HealthCheck(ctx context.Context) error {
	// Try to create a test producer
	testTopic := "health-check"
	_, err := p.getOrCreateProducer(testTopic)
	return err
}

// getOrCreateProducer gets an existing producer or creates a new one for the topic
func (p *Publisher) getOrCreateProducer(topic string) (pulsar.Producer, error) {
	if producer, exists := p.producers[topic]; exists {
		return producer, nil
	}

	producerOptions := pulsar.ProducerOptions{
		Topic: topic,
		Name:  fmt.Sprintf("conn-bridge-%s", topic),
	}

	producer, err := p.client.CreateProducer(producerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	p.producers[topic] = producer
	return producer, nil
}

// DefaultConfig returns a default Pulsar configuration
func DefaultConfig() *Config {
	return &Config{
		BrokerURL:               "pulsar://localhost:6650",
		OperationTimeout:        30 * time.Second,
		ConnectionTimeout:       10 * time.Second,
		MaxConnectionsPerBroker: 1,
	}
}
